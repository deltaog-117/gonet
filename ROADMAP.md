# Roadmap: gonet – Wi-Fi Manager (Go v1.0)

This document outlines the future direction of the network management tool.  
Items are organized by **priority**, not by timeline.

---

## ✅ Completed (Milestones Achieved)

### v1.0.0 – Initial Stable Release (2026-09-05) 🎉

- ✅ **Architecture & Planning**
  - Language selected: **Go** (single binary, distro-agnostic)
  - Backend strategy: **Subprocess orchestration** (`iw`, `wpa_cli`, `dhcpcd`)
  - Project structure: Feature-first (vertical slices) with `cmd/` and `internal/`

- ✅ **Core Features**
  - Data models: `Network`, `SecurityType`, `ConnectionState`
  - Scanner: `iw dev wlan0 scan` with SSID, BSSID, signal, security parsing
  - Status: `iw link` + Go `net` stdlib (IP, MAC, gateway)
  - Connector: `wpa_cli` sequential commands + `dhcpcd` / `udhcpc` fallback
  - Disconnector: `wpa_cli disconnect`
  - CLI: `scan`, `status`, `connect`, `disconnect` subcommands

- ✅ **TUI (BubbleTea)**
  - Interactive network selector with auto-refresh (every 5 seconds)
  - Signal bars with color coding (green → yellow → orange → red)
  - Security icons (🔓 Open, 🔐 WPA2/WPA3, 🔒 Enterprise)
  - Connected indicator (`✔`)
  - Hidden network support (`n` key)
  - Network details view (`d` key)
  - Sorting by signal strength (`s` key)
  - Password prompt with masked input

- ✅ **Reliability & UX**
  - UTF-8 / Latin-1 SSID decoding (supports cedillas, accents, etc.)
  - Exponential backoff retry (1s, 2s, 4s) for busy interfaces
  - Secure password handling (stdin piping – no password in `ps aux`)
  - Version flag (`--version`, `-V`)
  - Help flag (`--help`, `-h`)
  - Friendly error messages (no raw `exec` errors)

---

## 🔥 High Priority (v1.1.0 – Next Release)

These items are the **next logical steps** for a more polished and automated experience.

- **Interface Auto-Detection**
  - Instead of hardcoding `wlan0`, scan `/sys/class/net/` for wireless interfaces
  - Check for `wireless` subdirectory or `phy80211` symlink
  - Fallback to `iw dev` to list available interfaces

- **Configuration File Support**
  - Location: `~/.gonet/config.toml`
  - Store default interface, preferred DHCP client, and known networks
  - Enable auto-connect functionality

- **Auto-Connect on Boot (Daemon Mode)**
  - Subcommand: `gonet daemon` that runs in the background
  - Reads trusted networks from config file (SSID + PSK)
  - Periodically scans and auto-connects to known networks
  - **No systemd required** – works with Cron/OpenRC/Runit/rc.local
  - Supports boot-time startup via `@reboot` cron or init scripts

---

## 🟡 Medium Priority (Important – Polish & Reliability)

These make the tool **robust, user-friendly, and production-ready**.

- **TPM 2.0 Secure Storage for Auto-Connect**
  - Encrypt stored PSKs using the system's TPM 2.0 chip
  - Prevent offline decryption of config file if stolen
  - Use `github.com/google/go-tpm` to seal/unseal encryption keys
  - Bind to PCRs: key tied to system's boot state
  - Fallback: `swtpm` or system UUID + root‑only salt (TBD)

- **Connection State Machine (Concurrency-Safe)**
  - Use `sync.RWMutex` to protect global state shared between scanner goroutine and UI
  - Ensure UI doesn't freeze while scanning (run scan in a goroutine)

- **Observability (Structured Logging)**
  - Log every `exec.Command` call with args, exit code, and duration
  - Log connection attempts with SSID (but redact PSK)
  - Use `log/slog` with JSON format for production

- **Test Coverage**
  - Unit tests for parsers using mock `exec` outputs (sample `iw scan` output in `testdata/`)
  - Integration tests (optional, if you have a test VM)

---

## 🟢 Low Priority (Nice‑to‑Have – Stretch Goals)

These are **quality-of-life improvements** that aren't essential but elevate the app.

- **Multiple Interface Support**
  - Allow the user to switch between `wlan0`, `wlan1`, etc. in the TUI

- **Friendly Error Messages**
  - Replace `exec.ExitError` with human-readable explanations: "Network unreachable", "Wrong password", "Interface not found"

- **Connection History**
  - Store successfully connected networks in a simple JSON file
  - Auto-connect to the last known network on startup (if no auto-connect config exists)

- **DNS Verification**
  - After DHCP succeeds, ping `1.1.1.1` or resolve a domain to verify internet connectivity

- **Feature Flags for Rollback** (from your T2 manifesto)
  - Enable/disable experimental features (e.g., the D-Bus backend) without recompiling

- **CLI Tab Completion**
  - Bash/Zsh autocomplete for `gonet` subcommands

---

## 🔭 Long‑Term Vision (Ambitious / Exploratory)

These are **experimental ideas** that may or may not materialize. They're for v2.0 or beyond.

- **Pure Go D-Bus Backend**  
  - Replace `wpa_cli` subprocess calls with `godbus` to talk directly to `wpa_supplicant`  
  - **Why:** Event-driven, no text parsing, cleaner architecture

- **Pure Go Netlink Scanner**  
  - Replace `iw` with `vishvananda/netlink` and implement `nl80211` scanning natively  
  - **Why:** Zero external dependencies, kernel-level performance  
  - **Risk:** High implementation effort, poorly documented APIs

- **BSD Support (FreeBSD / OpenBSD)**  
  - Add build tags (`//go:build freebsd`) to call `ifconfig wlan0 scan` and adjust parsers  
  - **Why:** Makes the tool truly cross-platform (as originally planned)  
  - **Timeline:** Post-v1.0, once Linux is stable

- **Network Profiles / Saved Networks**  
  - Store known networks with PSK in an encrypted keyring (TPM-backed)
  - Auto-connect to the strongest known network on boot

- **Graphical UI (Web or Qt)**  
  - Embed a web UI (using `Fyne` or `Wails`) or expose an HTTP API for a frontend  
  - **Why:** Broader audience, but scope creep for v1.0

- **Hotspot / AP Mode**  
  - Allow the tool to create a Wi-Fi hotspot (using `hostapd`)  
  - **Why:** Makes it a full network management suite
