package fw

type FirewallRepository interface {
	Disable() error
}

type Service struct {
	firewallRepository FirewallRepository
}

func NewService(firewallRepository FirewallRepository) *Service {
	return &Service{
		firewallRepository: firewallRepository,
	}
}
