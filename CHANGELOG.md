# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Planned
- Interface auto‑detection
- Configuration file support (`~/.gonet/config.toml`)
- Auto‑connect daemon (no systemd)
- TPM 2.0 secure storage for passwords
- Full test coverage

---

## [1.0.0] - 2026-09-05

### Added

#### Core Features
- **CLI subcommands**: `scan`, `status`, `connect`, `disconnect`
- **Interactive TUI** built with BubbleTea:
  - Auto‑refreshing network list (every 5 seconds)
  - Signal strength bars with color coding (green → yellow → orange → red)
  - Security icons (🔓 Open, 🔐 WPA2/WPA3, 🔒 Enterprise)
  - Connected indicator (`✔`)
  - Hidden network support (`n` key)
  - Network details view (`d` key)
  - Sorting by signal strength (`s` key)
  - Password prompt with masked input
- **Scanner**: Parses `iw dev wlan0 scan` to extract SSID, BSSID, signal, channel, security type
- **Status**: Retrieves current SSID, signal, IP, MAC, gateway using `iw link` + Go `net` stdlib
- **Connector**: Securely connects to WPA2‑PSK networks using `wpa_cli` via stdin piping (password never appears in `ps aux`)
- **Disconnector**: Cleanly disconnects using `wpa_cli disconnect`

#### Reliability & UX
- Exponential backoff retry (1s, 2s, 4s) for busy interfaces
- UTF‑8 / Latin‑1 SSID decoding (supports cedillas, accents, etc.)
- Friendly error messages (no raw `exec` errors)
- Version flag (`--version`, `-V`)
- Help flag (`--help`, `-h`)
- DHCP fallback (`dhcpcd` → `udhcpc`)

#### Architecture
- Feature‑first vertical slices (scan, status, connect, disconnect)
- Read‑only `shared/` infrastructure (exec wrapper, logger, models)
- Single static binary (`CGO_ENABLED=0`), distro‑agnostic
- No systemd dependency – works with any init system

#### Documentation
- Comprehensive `README.md` with usage, installation, TUI shortcuts
- `DIARY.md` with architectural decision log
- `ROADMAP.md` with future plans

---

## [0.1.0] - 2026-09-05

### Added
- Initial project scaffold with feature‑first architecture
- Data models: `Network`, `SecurityType`, `ConnectionState`
- Basic CLI: `scan`, `status`, `connect`, `disconnect`
- Subprocess orchestration using `iw`, `wpa_cli`, `dhcpcd`
- `exec` abstraction for testability (Commander interface)
- Placeholder test files
