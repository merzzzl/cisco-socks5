package vpn

import "warp-server/internal/dto"

func (s *Service) GetState() (dto.VPNState, dto.VPNNotice, error) {
	return s.repositoryVPN.CurrentState()
}
