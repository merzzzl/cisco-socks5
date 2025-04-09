package vpn

func (s *Service) Connect() error {
	return s.repositoryVPN.Connect()
}
