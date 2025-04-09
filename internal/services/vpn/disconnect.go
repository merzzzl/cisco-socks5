package vpn

func (s *Service) Disconnect() error {
	return s.repositoryVPN.Disconnect()
}
