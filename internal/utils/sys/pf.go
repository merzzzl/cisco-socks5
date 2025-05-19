package sys

import (
	"fmt"
	"strings"
)

func DisablePF() error {
	if out, err := Command("pfctl -d"); err != nil {
		if strings.ContainsAny(out, "pf not enabled") {
			return nil
		}

		return fmt.Errorf("disable pfctl error: %w", err)
	}

	return nil
}

func CheckPF() (bool, error) {
	output, err := Command("pfctl -s info")
	if err != nil {
		return false, fmt.Errorf("check pfctl error: %w", err)
	}

	if strings.ContainsAny(output, "Status: Enabled") {
		return true, nil
	}

	return false, nil
}
