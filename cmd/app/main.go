package main

import (
	"cisco-socks5/api"
	"cisco-socks5/internal/config"
	"cisco-socks5/internal/controllers"
	repoVPN "cisco-socks5/internal/repositories/cisco"
	"cisco-socks5/internal/repositories/packetfilter"
	"cisco-socks5/internal/repositories/sshtunnel"
	"cisco-socks5/internal/services/fw"
	"cisco-socks5/internal/services/tunnel"
	"cisco-socks5/internal/services/vpn"
	"cisco-socks5/internal/ui"
	cl "cisco-socks5/pkg/controlloop"
	"cisco-socks5/pkg/log"
	"context"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jroimartin/gocui"
)

func main() {

	currentUser, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Cannot get current user: %v", err))
	}
	if currentUser.Gid == "0" || currentUser.Uid == "0" {
		panic(fmt.Sprintf("Don't run from root"))
	}

	l := &ui.LogWriter{Logs: make(chan string, 100)}
	log.SetOutput(l)
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("Error loading config: %s ", err))
	}

	newResource := cl.NewResource("newResource")
	mc := &api.MainConfig{
		Resource: *newResource,
		Spec:     api.MainConfigSpec{},
	}

	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	privateKeyPath := filepath.Join(sshDir, "id_rsa_cisco")
	publicKeyPath := filepath.Join(sshDir, "id_rsa_cisco.pub")
	sshTunnelRepo := sshtunnel.NewRepository(currentUser.Username)

	vpnService := vpn.NewService(repoVPN.NewRepository(cfg.CiscoProfile, cfg.CiscoUsername, cfg.CiscoPassword))
	fwService := fw.NewService(packetfilter.NewRepository(cfg.LocalPassword))
	tunnelService := tunnel.NewService(publicKeyPath, privateKeyPath, sshTunnelRepo)

	conditionChannel := make(chan []cl.Condition, 100)

	mainController := controllers.NewMainReconcile(
		conditionChannel,
		vpnService,
		fwService,
		tunnelService,
	)

	mainLoop, _ := cl.New(mainController, cl.WithLogger(log.NewLogger()))
	mainLoop.Storage.Add(mc)
	mainLoop.Run()
	log.Info().Msg("Main", "Start run main loop")

	ctxExit, cancel := context.WithCancel(context.Background())

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		panic(fmt.Errorf("Error loading gui: %s ", err))
	}
	go func() {
		defer cancel()

		err := ui.CreateTUI(cancel, g, l, conditionChannel)
		if err != nil {
			fmt.Println("CreateTUI error:", err)
		}
	}()

	go func() {
		defer cancel()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
	}()
	log.Info().Msg("Main", "cisco-socks5 started!")
	<-ctxExit.Done()

	log.Info().Msg("Main", "Stopping cisco-socks5...")
	mainLoop.Stop()
	log.Info().Msg("Main", "cisco-socks5 stopped")
	time.Sleep(time.Second * 2)
	g.Close()

	fmt.Println("Application stopped")
}
