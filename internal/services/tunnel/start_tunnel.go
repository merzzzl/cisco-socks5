package tunnel

import (
	"fmt"
	"warp-server/pkg/log"
)

func (s *Service) StartTunnel() error {
	pid, ok, err := s.sshTunnelRepository.GetPID()
	if err != nil {
		return fmt.Errorf("cannot get PID %s", err)
	}
	if !ok {
		log.Info().Msg("Main", "Starting tunnel...")
		pid, err = s.sshTunnelRepository.StartTunnel(s.privateKeyPath)
		if err != nil {
			return fmt.Errorf("cannot start tunnel %s", err)
		}
		log.Info().Msg("Main", "Starting tunnel success!")
	}

	ok, err = s.sshTunnelRepository.CheckHealthTCP()
	if err != nil {
		return fmt.Errorf("cannot check health %s", err)
	}
	if !ok {
		log.Info().Msg("Main", "trying to stop ssh tunnel...")
		err = s.sshTunnelRepository.StopTunnel(pid)
		if err != nil {
			return fmt.Errorf("trying to restart ssh tunnel, cannot stop tunnel %s", err)
		}
		return fmt.Errorf("ssh tunnel health check failed")
	}

	return nil
}
