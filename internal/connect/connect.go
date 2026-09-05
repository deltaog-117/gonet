package connect

import (
	"fmt"
	"strings"
	"time"

	"github.com/deltaog-117/gonet/internal/shared/exec"
)

// Connector handles Wi-Fi connections.
type Connector struct {
	commander exec.Commander
}

// New creates a new Connector.
func New(commander exec.Commander) *Connector {
	return &Connector{commander: commander}
}

// Connect attempts to connect to a WPA2-PSK network.
func (c *Connector) Connect(iface, ssid, psk string) error {
	// Step 1: Add network
	out, _, err := c.commander.Run("wpa_cli", "-i", iface, "add_network")
	if err != nil {
		return fmt.Errorf("failed to add network: %w", err)
	}
	networkID := strings.TrimSpace(out)

	// Step 2: Set SSID
	_, _, err = c.commander.Run("wpa_cli", "-i", iface, "set_network", networkID, "ssid", fmt.Sprintf(`"%s"`, ssid))
	if err != nil {
		return fmt.Errorf("failed to set SSID: %w", err)
	}

	// Step 3: Set PSK
	_, _, err = c.commander.Run("wpa_cli", "-i", iface, "set_network", networkID, "psk", fmt.Sprintf(`"%s"`, psk))
	if err != nil {
		return fmt.Errorf("failed to set PSK: %w", err)
	}

	// Step 4: Select network
	_, _, err = c.commander.Run("wpa_cli", "-i", iface, "select_network", networkID)
	if err != nil {
		return fmt.Errorf("failed to select network: %w", err)
	}

	// Step 5: Enable network
	_, _, err = c.commander.Run("wpa_cli", "-i", iface, "enable_network", networkID)
	if err != nil {
		return fmt.Errorf("failed to enable network: %w", err)
	}

	// Step 6: Wait for connection
	return c.waitForConnection(iface)
}

// waitForConnection polls wpa_cli status until connected or timeout.
func (c *Connector) waitForConnection(iface string) error {
	const timeout = 30 * time.Second
	const interval = 200 * time.Millisecond
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		out, _, err := c.commander.Run("wpa_cli", "-i", iface, "status")
		if err != nil {
			// Continue polling; wpa_cli might be busy
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

// ConnectInterface is a convenience function using the real Commander.
func ConnectInterface(iface, ssid, psk string) error {
	connector := New(&exec.RealCommander{})
	return connector.Connect(iface, ssid, psk)
}
