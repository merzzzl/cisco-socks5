package tunnel

import (
	"cisco-socks5/pkg/log"
	"fmt"
)

func (s *Service) StartTunnel() error {
	_, ok, err := s.sshTunnelRepository.GetPID()
	if err != nil {
		return fmt.Errorf("cannot get PID %s", err)
	}
	if !ok {
		log.Info().Msg("Main", "Starting tunnel...")
		_, err = s.sshTunnelRepository.StartTunnel(s.privateKeyPath)
		if err != nil {
			return fmt.Errorf("cannot start tunnel %s", err)
		}
		log.Info().Msg("Main", "Starting tunnel success!")
	}

	return nil
}
