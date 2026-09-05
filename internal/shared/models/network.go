package models

import "fmt"

// SecurityType represents the wireless security protocol.
type SecurityType int

const (
	SecurityUnknown SecurityType = iota
	SecurityOpen
	SecurityWEP
	SecurityWPA2
	SecurityWPA3
	SecurityWPA2Enterprise
	SecurityWPA3Enterprise
)

// String implements fmt.Stringer for SecurityType.
func (s SecurityType) String() string {
	switch s {
	case SecurityOpen:
		return "Open"
	case SecurityWEP:
		return "WEP"
	case SecurityWPA2:
		return "WPA2"
	case SecurityWPA3:
		return "WPA3"
	case SecurityWPA2Enterprise:
		return "WPA2-Enterprise"
	case SecurityWPA3Enterprise:
		return "WPA3-Enterprise"
	default:
		return "Unknown"
	}
}

// ConnectionState represents the current state of the Wi-Fi interface.
type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateScanning
	StateConnecting
	StateConnected
	StateFailed
)

// String implements fmt.Stringer for ConnectionState.
func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateScanning:
		return "Scanning"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// Network represents a discovered Wi-Fi network.
type Network struct {
	SSID         string       // Network name (human-readable)
	BSSID        string       // MAC address of the access point
	Signal       int          // Signal strength in dBm (e.g., -50)
	Channel      int          // Wi-Fi channel (e.g., 6, 36)
	SecurityType SecurityType // Security protocol
}

// IsValidSignal checks if the signal strength is within a reasonable range.
func (n *Network) IsValidSignal() bool {
	return n.Signal >= -100 && n.Signal <= 0
}

// String returns a human-readable representation of the network.
func (n *Network) String() string {
	return fmt.Sprintf("%s (%s) [%d dBm] - %s", n.SSID, n.BSSID, n.Signal, n.SecurityType)
}
