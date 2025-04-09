package sshkeys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Repository struct {
	sshDir string
	user   string
	host   string
}

func NewRepository(localUsername, host, sshDir string) *Repository {
	return &Repository{
		sshDir: sshDir,
		user:   localUsername,
		host:   host,
	}
}

func (r *Repository) GenerateKey(privateKeyPath, publicKeyPath string) ([]byte, error) {
	if err := os.MkdirAll(r.sshDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create .ssh directory %s", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key %s", err)
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if err := os.WriteFile(privateKeyPath, pem.EncodeToMemory(privateKeyPEM), 0600); err != nil {
		return nil, fmt.Errorf("failed to write private key %s", err)
	}

	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key %s", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(publicKey)
	if err := os.WriteFile(publicKeyPath, publicKeyBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write public key %s", err)
	}

	return publicKeyBytes, nil
}

func (r *Repository) GetKey(keyPath string) ([]byte, bool, error) {
	ok, err := fileExists(keyPath)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	publicKeyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read public key %s", err)
	}

	return publicKeyBytes, true, nil
}

func (r *Repository) KeyInstalled(publicKeyBytes []byte) (bool, error) {
	authorizedKeysPath := filepath.Join(r.sshDir, "authorized_keys")
	var existingContent []byte
	if _, err := os.Stat(authorizedKeysPath); err == nil {
		existingContent, err = os.ReadFile(authorizedKeysPath)
		if err != nil {
			return false, fmt.Errorf("failed to read authorized_keys %s", err)
		}
	}

	if strings.Contains(string(existingContent), string(publicKeyBytes)) {
		return true, nil
	}

	return false, nil
}

func (r *Repository) EnsureAuthorizedKeysSetup(publicKeyBytes []byte) error {
	authorizedKeysPath := filepath.Join(r.sshDir, "authorized_keys")
	f, err := os.OpenFile(authorizedKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open authorized_keys %s", err)
	}
	defer f.Close()

	if _, err := f.Write(publicKeyBytes); err != nil {
		return fmt.Errorf("failed to write to authorized_keys %s", err)
	}

	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
