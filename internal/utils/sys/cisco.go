package sys

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const (
	ciscoPath = "/opt/cisco/secureclient/bin/vpn"
)

const (
	ciscoStateConnected        = "Connected"
	ciscoStateDisconnected     = "Disconnected"
	ciscoNoticeReadyForConnect = "ReadyForConnect"
	ciscoUnknown               = "Unknown"
)

func CiscoConnect(profile, user, password string) error {
	var outBuf bytes.Buffer

	cmd := exec.Command(
		ciscoPath,
		"-s",
		"connect",
		profile,
	)

	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s\n%s\ny\n", user, password))

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("vpn connection error: %v\n", err)
	}

	output := outBuf.String()

	currentState := getLastCiscoState(string(output))
	if currentState != ciscoStateConnected {
		return fmt.Errorf("VPN connection not established: %s", string(output))
	}

	return nil
}

func CiscoCurrentState() (bool, bool, error) {
	output, err := Command("%s -s state", ciscoPath)
	if err != nil {
		return false, false, fmt.Errorf("vpn connection error: %v\n", err)
	}

	currentState := getLastCiscoState(string(output))
	currentNotice := getLastCiscoNotice(string(output))

	return currentState == ciscoStateConnected, currentNotice == ciscoNoticeReadyForConnect, nil
}

func CiscoDisconnect() error {
	output, err := Command("%s -s disconnect", ciscoPath)
	if err != nil {
		return fmt.Errorf("vpn disconnection error: %v\n %s", err, output)
	}

	return nil
}

func getLastCiscoState(output string) string {
	_, states := parseCiscoOutput(output)
	if len(states) > 0 {
		return getCiscoState(states[len(states)-1])
	}

	return ciscoUnknown
}

func getLastCiscoNotice(output string) string {
	notices, _ := parseCiscoOutput(output)
	if len(notices) > 0 {
		return getCiscoNotice(notices[len(notices)-1])
	}

	return ciscoUnknown
}

func parseCiscoOutput(output string) ([]string, []string) {
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

func getCiscoState(state string) string {
	switch state {
	case "Подключено", "Connected":
		return ciscoStateConnected
	case "Отключено", "Disconnected":
		return ciscoStateDisconnected
	default:
		return ciscoUnknown
	}
}

func getCiscoNotice(notice string) string {
	switch notice {
	case "Готово к подключению.":
		return ciscoNoticeReadyForConnect
	default:
		return ciscoUnknown
	}
}

func DisablePF() error {
	_, _ = Command("pfctl -d")

	return nil
}
