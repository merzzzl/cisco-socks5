package tunnel

func (s *Service) GetTunnelPID() (int, bool, error) {
	return s.sshTunnelRepository.GetPID()
}
