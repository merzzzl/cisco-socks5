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
	tick := time.NewTicker(5 * time.Second)

	for t := range tick.C {
		restarted, err := s.IsRestarted()
		if err != nil {
			log.Error().Err(err).Msg("SYS", "failed to check cisco")

			continue
		}

		if restarted {
			log.Info().Msgf("SYS", "cisco restarted at %s", t.Format(time.RFC3339))

			if s.proxyCloser != nil {
				if err := s.proxyCloser(); err != nil {
					log.Error().Err(err).Msg("SYS", "failed to close proxy")
				}

				s.proxyCloser = nil
			}
		}

		if s.proxyCloser == nil {
			if s.proxyCloser, err = s.StartProxy(ctx); err != nil {
				log.Error().Err(err).Msg("SYS", "failed to start proxy")

				continue
			}

			log.Info().Msg("SYS", "proxy started")
		}
	}

	if err := sys.CiscoDisconnect(); err != nil {
		log.Error().Err(err).Msg("SYS", "failed to disconnect cisco")
	}

	if s.proxyCloser != nil {
		if err := s.proxyCloser(); err != nil {
			log.Error().Err(err).Msg("SYS", "failed to close proxy")
		}
	}

	return nil
}

func (s *Service) IsRestarted() (bool, error) {
	connected, readyForConnect, err := sys.CiscoCurrentState()
	if err != nil {
		return false, fmt.Errorf("failed to get cisco state: %w", err)
	}

	s.status.CiscoConnected = connected
	s.status.CiscoReadyForConnect = readyForConnect

	if !connected {
		if err := sys.CiscoConnect(s.ciscoProfile, s.ciscoUser, s.ciscoPassword); err != nil {
			return true, fmt.Errorf("failed to connect to cisco: %w", err)
		}

		s.status.CiscoConnected = true
		s.status.CiscoReadyForConnect = false
		s.status.PFDisabled = false

		if err := sys.DisablePF(); err != nil {
			return false, fmt.Errorf("failed to disable pf: %w", err)
		}

		s.status.PFDisabled = true
	}

	return false, nil
}

func (s *Service) StartProxy(ctx context.Context) (func() error, error) {
	server := socks5.NewServer(socks5.WithConnectMiddleware(func(ctx context.Context, writer io.Writer, request *socks5.Request) error {
		log.Info().Msgf("SOC", "new connection from %s", request.DestAddr.Address())

		return nil
	}))

	lc := net.ListenConfig{}

	list, err := lc.Listen(ctx, "tcp", ":8000")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port 8000: %w", err)
	}

	go func() {
		if err := server.Serve(list); err != nil {
			s.status.ProxyStarted = false
		}
	}()

	s.status.ProxyStarted = true

	return func() error {
		s.status.ProxyStarted = false

		return list.Close()
	}, nil
}
