package tunnel

import "fmt"

func (s *Service) SetupSSHKey() error {

	if s.keyInstalled {
		return nil
	}
	publicKey, exist, err := s.sshKeysRepository.GetKey(s.publicKeyPath)
	if err != nil {
		return fmt.Errorf("cannot get public key %s", err)
	}
	if !exist {
		publicKey, err = s.sshKeysRepository.GenerateKey(s.privateKeyPath, s.publicKeyPath)
		if err != nil {
			return fmt.Errorf("cannot generate public key %s", err)
		}
	}
	ok, err := s.sshKeysRepository.KeyInstalled(publicKey)
	if err != nil {
		return fmt.Errorf("cannot check if key is installed: %s", err)
	}
	if !ok {
		err = s.sshKeysRepository.EnsureAuthorizedKeysSetup(publicKey)
		if err != nil {
			return fmt.Errorf("cannot set up authorized_keys: %s", err)
		}
	}
	s.keyInstalled = true

	return nil
}
