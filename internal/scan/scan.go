package scan

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deltaog-117/gonet/internal/shared/exec"
	"github.com/deltaog-117/gonet/internal/shared/models"
)

var (
	signalRegex = regexp.MustCompile(`signal:\s*(-?\d+(?:\.\d+)?)\s*dBm`)
	bssidRegex  = regexp.MustCompile(`BSS\s+([0-9a-fA-F:]{17})`)
)

type Scanner struct {
	commander exec.Commander
}

func New(commander exec.Commander) *Scanner {
	return &Scanner{commander: commander}
}

func (s *Scanner) Scan(iface string) ([]models.Network, error) {
	var stdout, stderr string
	var err error
	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt < len(backoff); attempt++ {
		stdout, stderr, err = s.commander.Run("iw", "dev", iface, "scan")
		if err == nil {
			break
		}
		if strings.Contains(stderr, "Device or resource busy") ||
			strings.Contains(stderr, "Operation not permitted") {
			time.Sleep(backoff[attempt])
			continue
		}
		return nil, fmt.Errorf("iw scan failed: %w (stderr: %s)", err, stderr)
	}
	if err != nil {
		return nil, fmt.Errorf("interface is busy; please wait and try again")
	}

	return parseIwScan(stdout)
}

func ScanInterface(iface string) ([]models.Network, error) {
	return New(&exec.RealCommander{}).Scan(iface)
}

func parseIwScan(output string) ([]models.Network, error) {
	var networks []models.Network
	var current *models.Network
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "BSS ") {
			if current != nil {
				networks = append(networks, *current)
			}
			current = &models.Network{
				SecurityType: models.SecurityUnknown,
				Signal:       -100,
				Channel:      0,
				SSID:         "",
				BSSID:        "",
			}
			if matches := bssidRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
				current.BSSID = matches[1]
			}
			continue
		}
		if current == nil {
			continue
		}

		// SSID parsing with escape sequence handling
		if strings.HasPrefix(trimmed, "SSID:") {
			raw := strings.TrimPrefix(trimmed, "SSID:")
			raw = strings.TrimSpace(raw)
			if raw == "" {
				current.SSID = "<hidden>"
			} else {
				// Unescape any \xHH sequences
				decoded := unescapeSSID(raw)
				current.SSID = decodeSSID(decoded)
			}
			continue
		}

		if strings.Contains(trimmed, "signal:") {
			if matches := signalRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
				if f, err := strconv.ParseFloat(matches[1], 64); err == nil {
					current.Signal = int(f)
				}
			}
			continue
		}

		if strings.Contains(trimmed, "channel") && !strings.Contains(trimmed, "signal:") {
			parts := strings.Fields(trimmed)
			for i, part := range parts {
				if part == "channel" && i+1 < len(parts) {
					if ch, err := strconv.Atoi(parts[i+1]); err == nil {
						current.Channel = ch
					}
					break
				}
			}
		}

		if strings.Contains(trimmed, "RSN:") || strings.Contains(trimmed, "WPA:") {
			current.SecurityType = models.SecurityWPA2
		} else if strings.Contains(trimmed, "WEP") {
			current.SecurityType = models.SecurityWEP
		} else if strings.Contains(trimmed, "Encryption:") && strings.Contains(trimmed, "off") {
			current.SecurityType = models.SecurityOpen
		}
	}
	if current != nil {
		networks = append(networks, *current)
	}
	if len(networks) == 0 {
		return nil, fmt.Errorf("no networks found in scan output")
	}
	return networks, nil
}

// unescapeSSID converts \xHH sequences to actual bytes.
func unescapeSSID(s string) []byte {
	var result []byte
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+3 < len(s) && s[i+1] == 'x' {
			// \xHH
			hexStr := s[i+2 : i+4]
			if b, err := hex.DecodeString(hexStr); err == nil && len(b) == 1 {
				result = append(result, b[0])
				i += 3
				continue
			}
		}
		result = append(result, s[i])
	}
	return result
}

// decodeSSID converts bytes to UTF-8 string, falling back to Latin-1 if needed.
func decodeSSID(b []byte) string {
	// Check if valid UTF-8
	valid := true
	for i := 0; i < len(b); i++ {
		if b[i] >= 0x80 {
			seqLen := utf8Len(b[i])
			if seqLen == 0 || i+seqLen > len(b) {
				valid = false
				break
			}
			for j := 1; j < seqLen; j++ {
				if b[i+j]&0xC0 != 0x80 {
					valid = false
					break
				}
			}
			i += seqLen - 1
		}
	}
	if valid {
		return string(b)
	}
	// Fallback: treat as Latin-1 (ISO-8859-1)
	runes := make([]rune, len(b))
	for i, ch := range b {
		runes[i] = rune(ch) // Latin-1 maps directly to Unicode code points 0-255
	}
	return string(runes)
}

func utf8Len(b byte) int {
	if b < 0x80 {
		return 1
	}
	if b&0xE0 == 0xC0 {
		return 2
	}
	if b&0xF0 == 0xE0 {
		return 3
	}
	if b&0xF8 == 0xF0 {
		return 4
	}
	return 0
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
