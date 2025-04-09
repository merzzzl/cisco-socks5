package fw

func (s *Service) Disable() error {
	return s.firewallRepository.Disable()
}
