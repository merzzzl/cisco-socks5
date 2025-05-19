package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/merzzzl/cisco-socks5/internal/utils/log"
	"github.com/merzzzl/cisco-socks5/internal/utils/sys"
	socks5 "github.com/things-go/go-socks5"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	status        *ServiceState
	proxyCloser   func() error
	ciscoUser     string
	ciscoPassword string
	ciscoProfile  string
}

type ServiceState struct {
	CiscoConnected       bool
	CiscoReadyForConnect bool
	PFDisabled           bool
	ProxyStarted         bool
}

func New(ciscoUser, ciscoPassword, ciscoProfile string) *Service {
	return &Service{
		status:        &ServiceState{},
		ciscoUser:     ciscoUser,
		ciscoPassword: ciscoPassword,
		ciscoProfile:  ciscoProfile,
	}
}

func (s *Service) GetState() ServiceState {
	return *s.status
}

func (s *Service) Start(ctx context.Context) error {
	ctx, closer := context.WithCancel(ctx)
	defer closer()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := s.StartCisco(ctx); err != nil {
			return fmt.Errorf("failed to start cisco: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		if err := s.DisablePF(ctx); err != nil {
			return fmt.Errorf("failed to disable pf: %w", err)
		}

		return nil
	})

	g.Go(func() error {
		if err := s.ProxyServer(ctx); err != nil {
			return fmt.Errorf("failed to start proxy: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		log.Error().Msgf("SYS", "failed to start service: %v", err)

		return err
	}

	return nil
}

func (s *Service) DisablePF(ctx context.Context) error {
	ctx, closer := context.WithCancel(ctx)

	defer func() {
		closer()

		s.status.PFDisabled = false
	}()

	for ctx.Err() == nil {
		if err := sys.DisablePF(); err != nil {
			return fmt.Errorf("failed to disable pf: %w", err)
		}

		s.status.PFDisabled = true

		time.Sleep(5 * time.Second)
	}

	return nil
}

func (s *Service) StartCisco(ctx context.Context) error {
	ctx, closer := context.WithCancel(ctx)

	defer func() {
		closer()

		s.status.CiscoConnected = false
		s.status.CiscoReadyForConnect = false
	}()

	if err := sys.CiscoConnect(s.ciscoProfile, s.ciscoUser, s.ciscoPassword); err != nil {
		return fmt.Errorf("failed to connect to cisco: %w", err)
	}

	for ctx.Err() == nil {
		connected, readyForConnect, err := sys.CiscoCurrentState()
		if err != nil {
			return fmt.Errorf("failed to get cisco state: %w", err)
		}

		s.status.CiscoConnected = connected
		s.status.CiscoReadyForConnect = readyForConnect

		if connected && readyForConnect {
			time.Sleep(5 * time.Second)
		}

		if !connected {
			if err := sys.CiscoConnect(s.ciscoProfile, s.ciscoUser, s.ciscoPassword); err != nil {
				return fmt.Errorf("failed to connect to cisco: %w", err)
			}
		}
	}

	return nil
}

type proxyLogger struct{}

func (p *proxyLogger) Errorf(format string, args ...interface{}) {
	log.Error().Msgf("SOC", format, args...)
}

func (s *Service) ProxyServer(ctx context.Context) error {
	ctx, closer := context.WithCancel(ctx)

	defer func() {
		closer()

		s.status.ProxyStarted = false
	}()

	server := socks5.NewServer(socks5.WithConnectMiddleware(func(ctx context.Context, writer io.Writer, request *socks5.Request) error {
		log.Info().Msgf("SOC", "new connection from %s", request.DestAddr.Address())

		return nil
	}), socks5.WithLogger(&proxyLogger{}))

	lc := net.ListenConfig{}

	list, err := lc.Listen(ctx, "tcp", ":8000")
	if err != nil {
		return fmt.Errorf("failed to listen on port 8000: %w", err)
	}

	go func() {
		<-ctx.Done()

		_ = list.Close()
	}()

	s.status.ProxyStarted = true

	if err := server.Serve(list); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	return nil
}
