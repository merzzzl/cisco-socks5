package vpn

import "cisco-socks5/internal/dto"

func (s *Service) GetState() (dto.VPNState, dto.VPNNotice, error) {
	return s.repositoryVPN.CurrentState()
}
