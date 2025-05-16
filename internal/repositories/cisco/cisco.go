package cisco

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"cisco-socks5/internal/dto"
)

const (
	VPNPath = "/opt/cisco/secureclient/bin/vpn"
)

type Repository struct {
	ciscoHost     string
	ciscoUsername string
	ciscoPassword string
}

func NewRepository(ciscoHost, ciscoUsername, ciscoPassword string) *Repository {
	return &Repository{ciscoHost, ciscoUsername, ciscoPassword}
}

func (r *Repository) Connect() error {
	var outBuf bytes.Buffer

	cmd := exec.Command(
		VPNPath,
		"-s",
		"connect",
		r.ciscoHost,
	)
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s\n%s\ny\n", r.ciscoUsername, r.ciscoPassword))
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("vpn connection error: %v\n", err)
	}
	output := outBuf.String()
	currentState := getLastState(string(output))
	if currentState != dto.VPNStateConnected {
		return fmt.Errorf("VPN connection not established: %s", string(output))
	}

	return nil
}

func (r *Repository) CurrentState() (dto.VPNState, dto.VPNNotice, error) {
	cmd := exec.Command(
		VPNPath,
		"-s",
		"state",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return dto.VPNStateUnknown, dto.VPNNoticeUnknown, fmt.Errorf("vpn connection error: %v\n", err)
	}
	currentState := getLastState(string(output))
	currentNotice := getLastNotice(string(output))
	return currentState, currentNotice, nil
}

func (r *Repository) Disconnect() error {
	cmd := exec.Command(
		VPNPath,
		"-s",
		"disconnect",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("vpn disconnection error: %v\n %s", err, output)
	}

	return nil
}

func getLastState(output string) dto.VPNState {
	_, states := parseOutput(output)
	if len(states) > 0 {
		return getState(states[len(states)-1])
	}
	return dto.VPNStateUnknown
}

func getLastNotice(output string) dto.VPNNotice {
	notices, _ := parseOutput(output)
	if len(notices) > 0 {
		return getNotice(notices[len(notices)-1])
	}
	return dto.VPNNoticeUnknown
}

func parseOutput(output string) ([]string, []string) {
	var notices []string
	var states []string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, ">> notice: ") {
			notice := strings.TrimPrefix(line, ">> notice: ")
			notices = append(notices, notice)
		}

		if strings.HasPrefix(line, ">> state: ") {
			state := strings.TrimPrefix(line, ">> state: ")
			states = append(states, state)
		}
	}
	return notices, states
}

func getState(state string) dto.VPNState {
	switch state {
	case "Подключено", "Connected":
		return dto.VPNStateConnected
	case "Отключено", "Disconnected":
		return dto.VPNStateDisconnected
	default:
		return dto.VPNStateUnknown
	}
}

func getNotice(notice string) dto.VPNNotice {
	switch notice {
	case "Готово к подключению.":
		return dto.VPNNoticeReadyForConnect
	default:
		return dto.VPNNoticeUnknown
	}
}
