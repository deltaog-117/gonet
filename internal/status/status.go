package status

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/deltaog-117/gonet/internal/shared/exec"
)

// Status holds the current connection information.
type Status struct {
	SSID    string
	Signal  int
	IP      string
	MAC     string
	Gateway string
}

// Manager handles retrieving connection status.
type Manager struct {
	commander exec.Commander
}

// New creates a new Status Manager.
func New(commander exec.Commander) *Manager {
	return &Manager{commander: commander}
}

// GetStatus retrieves the current connection status.
func (m *Manager) GetStatus(iface string) (*Status, error) {
	status := &Status{}

	// Get Wi-Fi details from iw link
	stdout, stderr, err := m.commander.Run("iw", "dev", iface, "link")
	if err != nil {
		return nil, fmt.Errorf("failed to get link status: %w (stderr: %s)", err, stderr)
	}

	// Parse SSID and signal
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "SSID:") {
			status.SSID = strings.TrimSpace(strings.TrimPrefix(trimmed, "SSID:"))
		}
		if strings.HasPrefix(trimmed, "signal:") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				signalStr := strings.TrimSuffix(parts[1], "dBm")
				if sig, err := strconv.Atoi(signalStr); err == nil {
					status.Signal = sig
				}
			}
		}
	}

	// Get IP address using Go's net package
	ifaceObj, err := net.InterfaceByName(iface)
	if err != nil {
		return status, nil // Return partial status if no IP
	}
	addrs, err := ifaceObj.Addrs()
	if err == nil {
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ipNet.IP.To4() != nil {
					status.IP = ipNet.IP.String()
					break
				}
			}
		}
	}

	// Get MAC address
	status.MAC = ifaceObj.HardwareAddr.String()

	// Get gateway via parsing ip route
	stdout, stderr, err = m.commander.Run("ip", "route", "show", "default")
	if err == nil {
		// Parse default route
		parts := strings.Fields(stdout)
		for i, part := range parts {
			if part == "via" && i+1 < len(parts) {
				status.Gateway = parts[i+1]
				break
			}
		}
	}

	return status, nil
}

// GetStatusInterface is a convenience function using the real Commander.
func GetStatusInterface(iface string) (*Status, error) {
	manager := New(&exec.RealCommander{})
	return manager.GetStatus(iface)
}
