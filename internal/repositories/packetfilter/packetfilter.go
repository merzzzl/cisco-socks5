package packetfilter

import (
	"fmt"
	"os/exec"
	"strings"
)

type Repository struct {
	sudoPassword string
}

func NewRepository(sudoPassword string) *Repository {
	return &Repository{sudoPassword: sudoPassword}
}

func (r *Repository) Disable() error {
	cmd := exec.Command(
		"sudo",
		"-S",
		"pfctl",
		"-d",
	)
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s\n", r.sudoPassword))
	body, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(body), "pf not enabled") {
			return nil
		}
		return fmt.Errorf("disable pfctl error: %v\n %s", err, body)
	}
	return nil
}
