package tunnel

import "context"

func (s *Service) StopTunnel(ctx context.Context, pid int) error {
	err := s.sshTunnelRepository.StopTunnel(pid)
	if err != nil {
		return err
	}
	return nil
}
