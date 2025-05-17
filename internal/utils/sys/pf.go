package sys

import (
	"fmt"
	"strings"
)

func DisablePF() error {
	if body, err := Command("pfctl -d"); err != nil {
		if strings.Contains(string(body), "pf not enabled") {
			return nil
		}

		return fmt.Errorf("disable pfctl error: %v\n %s", err, body)
	}

	return nil
}
