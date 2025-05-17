package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/merzzzl/cisco-socks5/internal/service"
	"github.com/merzzzl/cisco-socks5/internal/utils/log"
	"github.com/merzzzl/cisco-socks5/internal/utils/tui"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("APP", "failed on load config")
	}

	if cfg.debug {
		log.EnableDebug()
	}

	srv := service.New(cfg.CiscoUser, cfg.CiscoPassword, cfg.CiscoProfile)

	if !cfg.verbose {
		go func() {
			defer cancel()

			if err := tui.CreateTUI(srv, cfg.fun); err != nil {
				log.Error().Err(err).Msg("APP", "failed on create tui")
			}
		}()
	} else {
		go func() {
			defer cancel()

			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM)
			<-c

			if _, err := fmt.Print("\n"); err != nil {
				return
			}
		}()
	}

	if err := srv.Start(ctx); err != nil {
		log.Error().Err(err).Msg("APP", "failed on start service")
	}
}
