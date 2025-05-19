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
