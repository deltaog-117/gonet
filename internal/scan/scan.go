package scan

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/deltaog-117/gonet/internal/shared/exec"
	"github.com/deltaog-117/gonet/internal/shared/models"
)

// Scanner handles Wi-Fi network scanning.
type Scanner struct {
	commander exec.Commander
}

// New creates a new Scanner with the given Commander.
// For production, use exec.RealCommander{}.
func New(commander exec.Commander) *Scanner {
	return &Scanner{commander: commander}
}

// Scan executes a scan on the given wireless interface and returns a list of networks.
func (s *Scanner) Scan(iface string) ([]models.Network, error) {
	stdout, stderr, err := s.commander.Run("iw", "dev", iface, "scan")
	if err != nil {
		return nil, fmt.Errorf("iw scan failed: %w (stderr: %s)", err, stderr)
	}

	networks, err := parseIwScan(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse iw scan output: %w", err)
	}

	return networks, nil
}

// ScanInterface is a convenience function that uses the real Commander.
func ScanInterface(iface string) ([]models.Network, error) {
	scanner := New(&exec.RealCommander{})
	return scanner.Scan(iface)
}

// parseIwScan parses the output of `iw dev wlan0 scan` and returns a slice of Networks.
func parseIwScan(output string) ([]models.Network, error) {
	var networks []models.Network
	var current *models.Network
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Start of a new BSS entry
		if strings.HasPrefix(trimmed, "BSS ") {
			// Save the previous network if it exists
			if current != nil {
				networks = append(networks, *current)
			}
			current = &models.Network{
				SecurityType: models.SecurityUnknown,
				Signal:       0,
				Channel:      0,
			}
			// BSS format: "BSS 00:11:22:33:44:55 (on wlan0)"
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				current.BSSID = parts[1]
			}
			continue
		}

		// Skip if we're not inside a BSS section
		if current == nil {
			continue
		}

		// Parse SSID
		if strings.HasPrefix(trimmed, "SSID:") {
			// Some SSIDs are hex-encoded; handle both.
			ssidPart := strings.TrimPrefix(trimmed, "SSID:")
			ssidPart = strings.TrimSpace(ssidPart)
			// If it's a hex string (all hex chars and no spaces), decode it.
			if isHex(ssidPart) && len(ssidPart)%2 == 0 {
				if decoded, err := hex.DecodeString(ssidPart); err == nil {
					current.SSID = string(decoded)
				} else {
					current.SSID = ssidPart
				}
			} else {
				current.SSID = ssidPart
			}
			continue
		}

		// Parse Signal
		if strings.HasPrefix(trimmed, "signal:") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				signalStr := strings.TrimSuffix(parts[1], "dBm")
				if val, err := strconv.Atoi(signalStr); err == nil {
					current.Signal = val
				}
			}
			continue
		}

		// Parse Channel (looks like "channel 6" or "freq: 2412" etc)
		if strings.HasPrefix(trimmed, "channel") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				if ch, err := strconv.Atoi(parts[1]); err == nil {
					current.Channel = ch
				}
			}
		}
		// Alternative: freq line contains channel info, but we may skip.

		// Parse Security - look for RSN (WPA2/WPA3) and WPA lines
		if strings.Contains(trimmed, "RSN:") {
			// RSN usually means WPA2 or WPA3
			// We can differentiate by checking if "group cipher: CCMP" appears etc.
			// For simplicity, we'll mark as WPA2 for now; later we might refine.
			// But check if there's "WPA:" too.
			if strings.Contains(line, "WPA:") {
				// Could be WPA/WPA2 mixed, we'll use WPA2 as default.
				current.SecurityType = models.SecurityWPA2
			} else {
				// If RSN only, maybe WPA2 or WPA3, but we can't tell easily without deeper parsing.
				// We'll assume WPA2 for now.
				current.SecurityType = models.SecurityWPA2
			}
		} else if strings.Contains(trimmed, "WPA:") {
			current.SecurityType = models.SecurityWPA2
		} else if strings.Contains(trimmed, "WEP") {
			current.SecurityType = models.SecurityWEP
		} else if strings.Contains(trimmed, "Encryption:") && strings.Contains(trimmed, "off") {
			current.SecurityType = models.SecurityOpen
		}
		// If none of these match, it stays Unknown (which may be Open or hidden).
	}

	// Append the last network
	if current != nil {
		networks = append(networks, *current)
	}

	return networks, nil
}

// isHex checks if a string is a valid hex string (only 0-9A-Fa-f).
func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
