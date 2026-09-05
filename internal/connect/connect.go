package connect

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/deltaog-117/gonet/internal/shared/exec"
)

type Connector struct {
	commander exec.Commander
}

func New(commander exec.Commander) *Connector {
	return &Connector{commander: commander}
}

func (c *Connector) Connect(iface, ssid, psk string) error {
	cmd := c.commander.Command("wpa_cli", "-i", iface)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wpa_cli: %w", err)
	}

	commands := []string{
		"add_network",
		fmt.Sprintf("set_network 0 ssid \"%s\"", ssid),
		fmt.Sprintf("set_network 0 psk \"%s\"", psk),
		"select_network 0",
		"enable_network 0",
	}
	for _, cmdStr := range commands {
		if _, err := io.WriteString(stdin, cmdStr+"\n"); err != nil {
			return fmt.Errorf("failed to send command %q: %w", cmdStr, err)
		}
	}
	stdin.Close()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "FAIL") {
			return fmt.Errorf("wpa_cli command failed: %s", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading wpa_cli output: %w", err)
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wpa_cli exited with error: %w", err)
	}

	return c.waitForConnection(iface)
}

func (c *Connector) waitForConnection(iface string) error {
	const timeout = 30 * time.Second
	const interval = 200 * time.Millisecond
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		out, _, err := c.commander.Run("wpa_cli", "-i", iface, "status")
		if err != nil {
			time.Sleep(interval)
			continue
		}
		if strings.Contains(out, "wpa_state=COMPLETED") {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("connection timed out after %v", timeout)
}

func ConnectInterface(iface, ssid, psk string) error {
	connector := New(&exec.RealCommander{})
	return connector.Connect(iface, ssid, psk)
}
