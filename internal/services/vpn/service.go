package vpn

import "cisco-socks5/internal/dto"

type RepositoryVPN interface {
	Connect() error
	CurrentState() (dto.VPNState, dto.VPNNotice, error)
	Disconnect() error
}

type Service struct {
	repositoryVPN RepositoryVPN
}

func NewService(repositoryVPN RepositoryVPN) *Service {
	return &Service{
		repositoryVPN: repositoryVPN,
	}
}
