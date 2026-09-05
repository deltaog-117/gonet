# Roadmap: Wi-Fi Manager (Go v1.0)

This document outlines the future direction of the network management tool.  
Items are organized by **priority**, not by timeline.

---

## ✅ Completed (Milestones Achieved)

- ✅ **Architecture & Planning**
  - Language selected: **Go** (single binary, distro-agnostic)
  - Backend strategy: **Subprocess orchestration** (`iw`, `wpa_cli`, `dhcpcd`)
  - Feature decisions confirmed: Scan (A), View (A), Connect (A), Disconnect (A)
  - Project structure: Feature-first (vertical slices) with `cmd/` and `internal/`

---

## 🔥 High Priority (Critical – Must Have for v1.0)

These items are the **core functionality** – without them, the app doesn't work.

- **Data Models & Types**
  - Define `Network` struct (SSID, BSSID, Signal, SecurityType, Channel)
  - Define `SecurityType` using `iota` (Open, WPA2, WPA3, WEP, Unknown)
  - Define `ConnectionState` (Disconnected, Scanning, Connecting, Connected, Failed)
  - Implement validation methods (`IsValidSSID`, `IsValidPSK`)

- **Scanner Module (`internal/scanner`)**
  - Implement `exec.Command("iw", "dev", iface, "scan")`
  - Parse `iw` output using `bufio.Scanner` with a state machine
  - Extract: SSID (hex decode), BSSID, signal (dBm), security flags (RSN/WPA)
  - Return a slice of `Network` structs
  - Handle timeouts (scan should not hang forever)

- **Status / View Module (`internal/status`)**
  - Get current SSID + signal via `iw dev wlan0 link`
  - Get IP address + MAC via Go's `net.Interfaces()` and `net.InterfaceAddrs()`
  - Get default gateway via parsing `ip route show default` (or using `net` fallback)
  - Combine into a single `ConnectionStatus` struct

- **Connector Module (`internal/connector`)**
  - Implement `Connect(ssid, psk string) error`
  - Use `wpa_cli add_network` to get a network ID
  - Set SSID and PSK via `wpa_cli set_network`
  - Call `wpa_cli select_network` and `wpa_cli enable_network`
  - Poll `wpa_cli status` every 200ms until `wpa_state=COMPLETED` (with timeout, e.g., 30s)
  - Spawn `dhcpcd wlan0` (or fallback to `udhcpc -i wlan0`)
  - Return connection success/failure

- **Disconnector Module (`internal/connector`)**
  - Implement `Disconnect(iface string) error`
  - Call `wpa_cli disconnect`
  - Optionally call `wpa_cli remove_network 0` to forget the network
  - Optionally call `ip link set wlan0 down` (make configurable)

- **CLI / TUI Entry Point (`cmd/myapp/main.go`)**
  - Parse command-line flags (interface name, scan, connect, disconnect, status)
  - For v1.0, support both CLI subcommands (`scan`, `status`, `connect`, `disconnect`) AND a simple interactive TUI (using BubbleTea or just a loop)

- **Error Handling & Logging**
  - All `exec.Command` calls must handle `stderr` and return meaningful errors
  - Use `log/slog` for structured logging (JSON or text) – aligns with your T2 observability pillar
  - Never `panic` – always return `error` up the stack

---

## 🟡 Medium Priority (Important – Polish & Reliability)

These make the tool **robust, user-friendly, and production-ready**.

- **Interface Auto-Detection**
  - Instead of hardcoding `wlan0`, scan `/sys/class/net/` for wireless interfaces (check for `wireless` subdirectory or `phy80211` symlink)
  - Fallback to `iw dev` to list available interfaces

- **DHCP Fallback Strategy**
  - If `dhcpcd` not found, try `udhcpc -i wlan0` (common on Alpine/BusyBox)
  - If neither found, print clear installation instructions

- **Connection State Machine (Concurrency-Safe)**
  - Use `sync.RWMutex` to protect a global state struct shared between scanner goroutine and UI
  - Ensure UI doesn't freeze while scanning (run scan in a goroutine)

- **Signal Strength Interpretation**
  - Convert dBm to a human-readable percentage or bars (e.g., > -50 = Excellent, > -60 = Good, etc.)

- **Security Type Detection**
  - Improve parser to distinguish between WPA2-Personal, WPA2-Enterprise, WPA3, and Open
  - Handle mixed-mode networks (WPA/WPA2)

- **Configuration File Support**
  - Support reading default interface and preferred DHCP client from a config file (e.g., `~/.netman.toml`)
  - Support storing known networks (for auto-connect later)

- **Test Coverage**
  - Write unit tests for parsers using mock `exec` outputs (store sample `iw scan` output in `testdata/`)
  - Write integration tests (optional, if you have a test VM)

- **Observability (Structured Logging)**
  - Log every `exec.Command` call with its args, exit code, and duration
  - Log connection attempts with SSID (but redact PSK)

---

## 🟢 Low Priority (Nice‑to‑Have – Stretch Goals)

These are **quality-of-life improvements** that aren't essential but elevate the app.

- **TUI Polishing (BubbleTea)**
  - Interactive network list with signal bars and security icons
  - Connect/Disconnect from a menu
  - Auto-refresh every 5 seconds

- **Multiple Interface Support**
  - Allow the user to switch between `wlan0`, `wlan1`, etc.

- **Friendly Error Messages**
  - Replace `exec.ExitError` with human-readable explanations: "Network unreachable", "Wrong password", "Interface not found"

- **Hidden Network Support**
  - Allow connecting to hidden SSIDs (user must manually enter SSID)

- **Connection History**
  - Store successfully connected networks in a simple JSON file
  - Auto-connect to the last known network on startup

- **DNS Verification**
  - After DHCP succeeds, ping `1.1.1.1` or resolve a domain to verify internet connectivity

- **Feature Flags for Rollback** (from your T2 manifesto)
  - Implement a simple feature flag mechanism to enable/disable experimental features (e.g., the D-Bus backend) without recompiling

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
  - Store known networks with PSK in an encrypted keyring (or plaintext with user warning)  
  - Auto-connect to the strongest known network on boot

- **Graphical UI (Web or Qt)**  
  - Embed a web UI (using `Fyne` or `Wails`) or expose an HTTP API for a frontend  
  - **Why:** Broader audience, but scope creep for v1.0

- **Hotspot / AP Mode**  
  - Allow the tool to create a Wi-Fi hotspot (using `hostapd`)  
  - **Why:** Makes it a full network management suite

---

## 📋 Summary of v1.0 Scope (Hard Cut)

| Module | Status | Deliverable |
|--------|--------|-------------|
| Scanner | ✅ High | Scan and list networks |
| Status | ✅ High | Show current connection + IP |
| Connector | ✅ High | Connect to WPA2-PSK networks |
| Disconnector | ✅ High | Disconnect cleanly |
| CLI/TUI | ✅ High | Basic subcommands + interactive TUI |
| Error Handling | ✅ High | No panics, clear errors |
| DHCP Fallback | 🟡 Medium | Support `udhcpc` |
| Config File | 🟡 Medium | `~/.netman.toml` |
| Tests | 🟡 Medium | Unit tests for parsers |
| BSD Support | 🔭 Long | Not in v1.0 |

---

## 🎯 Next Actions (Immediate)

1. **Scaffold the project** – Create the directory structure (`cmd/`, `internal/`, `pkg/`, `testdata/`).
2. **Define data models** – Create `internal/models/network.go`.
3. **Write the scanner** – Implement `internal/scanner/scan.go` with a sample `iw scan` parser.
4. **Write the status** – Implement `internal/status/status.go` using `net` + `iw link`.
5. **Write the connector** – Implement `internal/connector/connect.go` with `wpa_cli` + `dhcpcd`.
6. **Write the entrypoint** – `cmd/netman/main.go` with flag parsing.
