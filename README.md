# gonet

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://go.dev/dl/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

**A distro‑agnostic Wi‑Fi manager for Linux – scan, view, connect, and disconnect from wireless networks directly from your terminal.**

---

## 💡 About

Every Linux distribution handles Wi‑Fi differently. `gonet` cuts through the complexity by orchestrating standard system tools (`iw`, `wpa_cli`, `dhcpcd`) – no NetworkManager, no systemd, no bloat.

Whether you're running **Alpine, Void, Arch, Ubuntu, or any other Linux**, `gonet` gives you a consistent, reliable way to manage your wireless connections.

**Who is this for?**
- 🐧 Linux users who want a simple, dependency‑free Wi‑Fi tool
- 🖥️ Terminal enthusiasts who prefer keyboard‑driven interfaces
- 🔧 Minimalists who avoid NetworkManager and systemd
- 🧠 Anyone who wants to understand how Wi‑Fi actually works

---

## ✨ Features

- 📡 **Scan** – Discover all Wi‑Fi networks in range with signal strength and security type
- 🔍 **Status** – View current connection (SSID, signal, IP, MAC, gateway)
- 🔗 **Connect** – Connect to WPA2‑PSK networks (Open, WEP, WPA3 coming soon)
- 🔌 **Disconnect** – Disconnect cleanly from the current network
- 🖥️ **TUI** – Interactive terminal interface with:
  - Auto‑refreshing network list (every 5 seconds)
  - Signal strength bars with color coding (green → yellow → orange → red)
  - Security icons (🔓 Open, 🔐 WPA2/WPA3, 🔒 Enterprise)
  - Connected indicator (`✔`)
  - Hidden network support
  - Network details view
  - Sort by signal strength
- 🔐 **Secure** – Passwords are sent via `stdin` (never appear in `ps aux`)
- 🌍 **Distro‑agnostic** – Works on Alpine, Void, Arch, Ubuntu, Fedora, and more
- 📦 **Single binary** – No dependencies, no runtime, just one file

---

## 📋 Requirements

- **Linux** (kernel 3.0+)
- **`iw`** – Wireless configuration tool
- **`wpa_supplicant`** – Wi‑Fi authentication daemon
- **DHCP client** – `dhcpcd` (recommended) or `udhcpc` (BusyBox fallback)

Install them on your distribution:

```bash
# Debian / Ubuntu / Pop!_OS
sudo apt install iw wpasupplicant dhcpcd5

# Arch / Manjaro
sudo pacman -S iw wpa_supplicant dhcpcd

# Fedora / RHEL
sudo dnf install iw wpa_supplicant dhcpcd

# Alpine
sudo apk add iw wpa_supplicant dhcpcd

# Void
sudo xbps-install iw wpa_supplicant dhcpcd
```

---

## 📦 Installation

### Option 1: Download Binary (Recommended)

Download the latest release from [GitHub Releases](https://github.com/deltaog-117/gonet/releases):

```bash
# Example: download and install v1.0.0 for Linux amd64
wget https://github.com/deltaog-117/gonet/releases/download/v1.0.0/gonet-linux-amd64
chmod +x gonet-linux-amd64
sudo mv gonet-linux-amd64 /usr/local/bin/gonet
```

### Option 2: Build from Source

```bash
git clone https://github.com/deltaog-117/gonet.git
cd gonet
go build ./cmd/gonet/
sudo cp gonet /usr/local/bin/
```

---

## 🚀 Usage

### CLI Commands

| Command | Description |
|---------|-------------|
| `gonet scan -iface wlan0` | Scan for available networks |
| `gonet status -iface wlan0` | Show current connection status |
| `gonet connect "SSID" "PSK" -iface wlan0` | Connect to a WPA2‑PSK network |
| `gonet disconnect -iface wlan0` | Disconnect from current network |
| `gonet tui -iface wlan0` | Launch interactive TUI |
| `gonet --version` | Print version and exit |
| `gonet --help` | Show help |

**Note:** Most commands require `sudo` (Wi‑Fi operations need elevated privileges).

```bash
# Scan for networks
sudo gonet scan -iface wlan0

# Check current status
sudo gonet status -iface wlan0

# Connect to a network
sudo gonet connect "MyHomeWiFi" "MySecretPassword" -iface wlan0

# Disconnect
sudo gonet disconnect -iface wlan0

# Launch the TUI
sudo gonet tui -iface wlan0
```

---

### 🖥️ TUI Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑` / `↓` / `k` / `j` | Navigate network list |
| `Enter` | Select network / Connect |
| `r` | Refresh scan |
| `d` | Show network details |
| `n` | Connect to a hidden network |
| `s` | Toggle sorting (signal / name) |
| `q` / `Ctrl+C` | Quit |
| `Esc` | Cancel / Go back |

---

### 🎨 Signal Bars Legend

| dBm Range | Quality | Color |
|-----------|---------|-------|
| `-30` to `-49` | Excellent | 🟢 Green |
| `-50` to `-59` | Good | 🟢 Green |
| `-60` to `-69` | Fair | 🟡 Yellow |
| `-70` to `-79` | Poor | 🟠 Orange |
| `-80` and below | Weak | 🔴 Red |

---

## ⚙️ Configuration

*Configuration support is planned for v1.1.0.*  
Future features include:
- Default interface and DHCP client settings in `~/.gonet/config.toml`
- TPM‑backed encrypted storage for auto‑connect passwords

---

## 📁 Project Structure

```
gonet/
├── cmd/gonet/           # CLI and TUI entrypoints
│   ├── main.go          # CLI subcommands
│   └── tui.go           # BubbleTea TUI implementation
├── internal/            # Feature modules (vertical slices)
│   ├── appstate/        # Global application state
│   ├── connect/         # Connect to networks
│   ├── disconnect/      # Disconnect from networks
│   ├── scan/            # Scan for networks
│   ├── shared/          # Read‑only infrastructure
│   │   ├── exec/        # Testable os/exec wrapper
│   │   ├── logger/      # Structured logging (slog)
│   │   └── models/      # Data models (Network, SecurityType)
│   └── status/          # View current connection status
├── tests/               # Unit and integration tests
├── DIARY.md             # Architectural decision log
├── ROADMAP.md           # Project roadmap
└── README.md            # This file
```

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a specific test
go test -v ./internal/scan/
```

**Note:** Integration tests are skipped by default (they require real hardware). To run them, use `go test -tags=integration ./...`.

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a feature branch (`git checkout -b feature/amazing`).
3. Commit your changes (`git commit -m 'Add amazing feature'`).
4. Push to the branch (`git push origin feature/amazing`).
5. Open a Pull Request.

See [DIARY.md](DIARY.md) for architectural decisions and [ROADMAP.md](ROADMAP.md) for planned features.

---

## 📄 License

This project is licensed under the **GNU Affero General Public License v3.0** – see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgements

- [BubbleTea](https://github.com/charmbracelet/bubbletea) – The amazing TUI framework
- [iw](https://wireless.wiki.kernel.org/en/users/documentation/iw) – Linux wireless configuration
- [wpa_supplicant](https://w1.fi/wpa_supplicant/) – Wi‑Fi authentication
- The open‑source community for their incredible tools and inspiration

---

## 💬 Questions / Support

- Open an [issue](https://github.com/deltaog-117/gonet/issues)
- Star ⭐ the project on GitHub to show your support

---

## 📜 Changelog

See the [CHANGELOG.md](CHANGELOG.md) file for version history.
