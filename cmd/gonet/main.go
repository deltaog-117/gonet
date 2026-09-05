package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/deltaog-117/gonet/internal/connect"
	"github.com/deltaog-117/gonet/internal/disconnect"
	"github.com/deltaog-117/gonet/internal/scan"
	"github.com/deltaog-117/gonet/internal/status"
)

var globalIface string

func main() {
	flag.StringVar(&globalIface, "iface", "wlan0", "wireless interface to use")
	flag.Parse()

	if len(flag.Args()) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := flag.Args()[0]
	switch cmd {
	case "scan":
		scanCmd()
	case "status":
		statusCmd()
	case "connect":
		if len(flag.Args()) < 3 {
			fmt.Println("Usage: connect <SSID> <PSK>")
			os.Exit(1)
		}
		ssid := flag.Args()[1]
		psk := flag.Args()[2]
		connectCmd(ssid, psk)
	case "disconnect":
		disconnectCmd()
	case "tui":
		tuiCmd()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`gonet - Wi-Fi manager for Linux

Usage:
  gonet [flags] <command> [arguments]

Commands:
  scan                    Scan for available networks
  status                  Show current connection status
  connect <SSID> <PSK>    Connect to a WPA2-PSK network
  disconnect              Disconnect from current network
  tui                     Start interactive TUI

Flags:
  -iface string   wireless interface (default "wlan0")
  -h, -help       show this help
`)
}

func scanCmd() {
	networks, err := scan.ScanInterface(globalIface)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
		os.Exit(1)
	}
	if len(networks) == 0 {
		fmt.Println("No networks found.")
		return
	}
	fmt.Printf("%-30s %-18s %-8s %s\n", "SSID", "BSSID", "Signal", "Security")
	fmt.Println("---------------------------------------------------------------")
	for _, net := range networks {
		fmt.Printf("%-30s %-18s %-8d %s\n", net.SSID, net.BSSID, net.Signal, net.SecurityType)
	}
}

func statusCmd() {
	s, err := status.GetStatusInterface(globalIface)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Status failed: %v\n", err)
		os.Exit(1)
	}
	if s.SSID == "" {
		fmt.Println("Not connected to any network.")
		return
	}
	fmt.Printf("SSID:    %s\n", s.SSID)
	fmt.Printf("Signal:  %d dBm\n", s.Signal)
	fmt.Printf("IP:      %s\n", s.IP)
	fmt.Printf("MAC:     %s\n", s.MAC)
	fmt.Printf("Gateway: %s\n", s.Gateway)
}

func connectCmd(ssid, psk string) {
	fmt.Printf("Connecting to %s...\n", ssid)
	err := connect.ConnectInterface(globalIface, ssid, psk)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Association successful. Obtaining IP via DHCP...")
	// Try dhcpcd first
	cmd := exec.Command("dhcpcd", globalIface)
	if err := cmd.Run(); err != nil {
		// Fallback to udhcpc
		cmd = exec.Command("udhcpc", "-i", globalIface)
		if err2 := cmd.Run(); err2 != nil {
			fmt.Fprintf(os.Stderr, "Warning: DHCP failed: %v (tried dhcpcd and udhcpc)\n", err)
			// Still connected, but no IP
		} else {
			fmt.Println("IP obtained via udhcpc.")
		}
	} else {
		fmt.Println("IP obtained via dhcpcd.")
	}
	fmt.Println("Connected successfully!")
}

func disconnectCmd() {
	err := disconnect.DisconnectInterface(globalIface)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Disconnect failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Disconnected.")
}

func tuiCmd() {
	m := NewModel(globalIface)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI failed: %v\n", err)
		os.Exit(1)
	}
}
