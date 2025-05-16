package tunnel

func (s *Service) StopTunnel(pid int) error {
	err := s.sshTunnelRepository.StopTunnel(pid)
	if err != nil {
		return err
	}
	return nil
}
