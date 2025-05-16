package tunnel

type SSHKeysRepository interface {
	GetKey(keyPath string) ([]byte, bool, error)
	KeyInstalled(publicKeyBytes []byte) (bool, error)
	EnsureAuthorizedKeysSetup(publicKeyBytes []byte) error
	GenerateKey(privateKeyPath, publicKeyPath string) ([]byte, error)
}

type SSHTunnelRepository interface {
	StartTunnel(sshKeyPath string) (int, error)
	IsRunning(pid int) (bool, error)
	GetPID() (int, bool, error)
	StopTunnel(pid int) error
}

type Service struct {
	publicKeyPath       string
	privateKeyPath      string
	sshTunnelRepository SSHTunnelRepository

	keyInstalled bool
}

func NewService(
	publicKeyPath string,
	privateKeyPath string,
	sshTunnelRepository SSHTunnelRepository,
) *Service {
	return &Service{
		publicKeyPath:       publicKeyPath,
		privateKeyPath:      privateKeyPath,
		sshTunnelRepository: sshTunnelRepository,
	}
}
