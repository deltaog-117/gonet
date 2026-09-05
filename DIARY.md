# 📘 DIARY.md

## Architectural Decision Log

*This document serves as a chronological record of architectural decisions, trade-offs, and reasoning throughout the project's lifecycle. Just as Git tracks code changes, this diary tracks the **why** behind the code.*

---

## 📋 Decision Index

| Date | Decision Area | Choice | Status |
|------|---------------|--------|--------|
| 2026-09-05 | Language & Backend Strategy | Go + Subprocess Orchestration | ✅ Confirmed |
| 2026-09-05 | Feature: Scanning | `iw dev wlan0 scan` + Text Parsing | ✅ Confirmed |
| 2026-09-05 | Feature: View/Status | `iw link` + Go `net` stdlib | ✅ Confirmed |
| 2026-09-05 | Feature: Connect | `wpa_cli` sequential commands + `dhcpcd` | ✅ Confirmed |
| 2026-09-05 | Feature: Disconnect | `wpa_cli disconnect` | ✅ Confirmed |
| 2026-09-06 | Interactive TUI (Network Selector) | BubbleTea | ✅ Confirmed |
| 2026-09-06 | Auto‑Connect Daemon | Pure Go + Cron/OpenRC/Runit (no systemd) | 🔄 Planned |

---

## 📝 Decision Entries

### Backend Strategy & Feature Implementation (v1.0 Core)

**Date:** 2026-09-05  
**Status:** Confirmed

---

#### Context / Background

I am building a network management tool (Find, View, Connect, Disconnect) for Linux (and later BSD) that must be **distro-agnostic** (no systemd dependency) and easy to distribute. The goal for v1.0 is a working CLI/TUI tool that orchestrates existing system binaries rather than reinventing kernel-level networking logic.

Key constraints:
- Must work on Alpine (musl), Void (runit), Devuan (OpenRC), Ubuntu (systemd), and Arch (systemd).
- Must be maintainable and debuggable.
- Should be fast enough for interactive use (scans under 3 seconds).
- No external libraries beyond the Go standard library for core functionality.

I previously evaluated Rust (type-safe but slow to compile, complex for subprocess parsing) and Python (fast to prototype but distribution is messy). I chose **Go** for its single-binary distribution (`CGO_ENABLED=0`), excellent `os/exec` package, fast compile times, and portfolio value.

This entry documents the specific implementation choices for each of the four core features.

---

#### Options Considered

**Feature 1: Find / Scan (Discovering available networks)**

| Aspect | Option A: `iw dev wlan0 scan` + Text Parsing | Option B: D-Bus via `godbus` (wpa_supplicant) | Option C: Pure Go Netlink (`vishvananda/netlink`) |
|--------|-----------------------------------------------|-----------------------------------------------|---------------------------------------------------|
| **Advantages** | • `iw` installed on 99% of distros <br> • Pure Go stdlib <br> • Easy to debug (raw output visible) <br> • No D-Bus permissions required | • Structured data (no parsing) <br> • Access to security flags as enums <br> • Event-driven (signals) | • Zero external dependencies <br> • Kernel-level speed <br> • Full control |
| **Disadvantages** | • Text parsing is brittle (format changes) <br> • Requires `iw` binary <br> • Parsing nested blocks is error-prone | • `godbus` steep learning curve <br> • Requires D-Bus & wpa_supplicant <br> • Heavier error handling | • Netlink Wi-Fi APIs barely documented <br> • Constructing nested attributes manually <br> • Debugging packets is a nightmare |
| **Difficulty** | **Easy-Medium** | Hard | Extreme |
| **Fit** | ✅ Excellent for v1.0 | ❌ Overkill | ❌ Unrealistic |

**Feature 2: View / Status (Current connection + IP address)**

| Aspect | Option A: Hybrid (`iw link` + Go `net` stdlib) | Option B: Pure Subprocess (parse everything) | Option C: D-Bus (NetworkManager + wpa_supplicant) |
|--------|-----------------------------------------------|-----------------------------------------------|---------------------------------------------------|
| **Advantages** | • `net` package is rock-solid <br> • No parsing `ip addr` <br> • Works on BSD too | • Consistent methodology (all subprocess) | • Structured data |
| **Disadvantages** | • Still depends on `iw` for Wi-Fi specifics | • `ip addr` output is ugly to parse <br> • `ip route` for gateway is annoying | • Depends on two D-Bus services <br> • Extremely heavy |
| **Difficulty** | **Easy** | Medium | Hard |
| **Fit** | ✅ **Golden path** | ⚠️ Avoid | ❌ Too complex |

**Feature 3: Connect (Authenticate and get IP)**

| Aspect | Option A: `wpa_cli` sequential + `dhcpcd` | Option B: D-Bus `wpa_supplicant` AddNetwork | Option C: Write `wpa_supplicant.conf` + restart |
|--------|-------------------------------------------|---------------------------------------------|------------------------------------------------|
| **Advantages** | • `wpa_cli` standard everywhere <br> • Easy to debug (run manually) <br> • `dhcpcd` available on all major distros | • Event-driven (no polling) <br> • Cleaner architecture | • Simple file IO <br> • Reliable |
| **Disadvantages** | • Requires polling for `COMPLETED` <br> • `dhcpcd` might be missing on minimal systems | • `godbus` complexity <br> • Harder to debug | • Requires root to write `/etc/` <br> • Restarting daemon kills all interfaces |
| **Difficulty** | **Medium** | Hard | Easy (but dangerous) |
| **Fit** | ✅ Recommended | ⚠️ Post-v1.0 | ❌ Avoid (disruptive) |

**Feature 4: Disconnect (De-authenticating)**

| Aspect | Option A: `wpa_cli disconnect` | Option B: D-Bus `wpa_supplicant` Disconnect | Option C: Netlink `NL80211_CMD_DEAUTHENTICATE` |
|--------|--------------------------------|---------------------------------------------|------------------------------------------------|
| **Advantages** | • One line <br> • Works everywhere | • Clean programmatic call | • Fast |
| **Disadvantages** | • None significant | • D-Bus overhead | • Requires knowing BSSID <br> • Dangerous if misused |
| **Difficulty** | **Easy** | Medium | Hard |
| **Fit** | ✅ **The way** | ⚠️ Unnecessary | ❌ Absolutely not |

---

#### Decision & Rationale

**Chosen Options:**  
- **Scan:** Option A (`iw` + text parsing)  
- **View:** Option A (Hybrid `iw link` + Go `net` stdlib)  
- **Connect:** Option A (`wpa_cli` sequential + `dhcpcd`)  
- **Disconnect:** Option A (`wpa_cli disconnect`)

**Reasoning:**

I chose the subprocess/hybrid strategy across all four features for the following reasons:

1. **Universality** – All chosen binaries (`iw`, `wpa_cli`, `dhcpcd`) are installed on every mainstream Linux distro (Debian, Ubuntu, Arch, Fedora, Alpine, Void) without exception. This guarantees distro-agnosticism without relying on systemd or any specific init system.

2. **Debuggability** – When something fails, I can run the same command manually in a terminal, see the exact error, and fix the parser. With D-Bus, errors are often hidden behind introspection quirks or permission issues that are hard to diagnose.

3. **Go's Strengths** – Go's `os/exec` package is a joy to use. Piping stdin, reading stdout/stderr, and handling exit codes is straightforward. The `net` package for IP retrieval is battle-tested and cross-platform.

4. **Developer Velocity** – Writing subprocess orchestration in Go takes a fraction of the time compared to writing D-Bus clients or Netlink parsers. I can have a working scanning loop in a few hours.

5. **Maintainability** – The code will be linear, synchronous, and easy to follow. Future contributors (or future me) won't need to understand D-Bus internals or kernel netlink attributes.

**Trade-offs accepted:**

| Trade-off | Why it's acceptable |
|-----------|---------------------|
| **Text parsing brittleness** | `iw` and `wpa_cli` output formats are stable across kernel versions. I'll write forgiving parsers (using `strings.Contains` and `strings.HasPrefix`). If formats change, I'll add fallback parsers. |
| **Dependency on `iw`/`wpa_cli`** | These are standard packages. If a system lacks them, I'll provide a clear error message and suggest installing `wireless-tools` or `wpa_supplicant`. |
| **Polling for `COMPLETED`** | I'll poll `wpa_cli status` every 200ms. The hardware scan/connect takes ~2 seconds, so 10 checks is trivial overhead. |
| **`dhcpcd` may be missing** | I'll fall back to `udhcpc -i wlan0` for Alpine/BusyBox systems. If neither is found, I'll exit with a helpful message. |
| **No structured data** | I'll manually parse SSID, BSSID, signal, and security flags. The parser will be isolated in a dedicated `internal/parser` package, making it easy to swap later. |

---

#### Implementation Notes

**Feature 1 (Scan):**
- Use `exec.Command("iw", "dev", wlanInterface, "scan")`.
- Use `bufio.NewScanner` to read stdout line by line.
- Use a small state machine to detect block boundaries (`BSS` lines).
- Extract SSID (hex to string), BSSID, signal (parse `signal:`), security (parse `RSN` and `WPA` lines).
- Return a slice of `Network` structs.

**Feature 2 (View):**
- Use `exec.Command("iw", "dev", wlanInterface, "link")` to get SSID and signal.
- Use Go's `net.Interfaces()` and `net.InterfaceAddrs()` for IP and MAC.
- Use `net.InterfaceByName("wlan0")` to get the MAC address.
- Fall back to parsing `ip route` if `net` doesn't provide the gateway.

**Feature 3 (Connect):**
- Use `exec.Command("wpa_cli", "-i", wlanInterface, "add_network")` to get a network ID.
- Use `exec.Command("wpa_cli", "-i", wlanInterface, "set_network", id, "ssid", "\""+ssid+"\"")`.
- Use `exec.Command("wpa_cli", "-i", wlanInterface, "set_network", id, "psk", "\""+psk+"\"")`.
- Use `exec.Command("wpa_cli", "-i", wlanInterface, "select_network", id)`.
- Poll `wpa_cli -i wlanInterface status` with a 200ms ticker until `wpa_state=COMPLETED`.
- Once connected, spawn `dhcpcd wlanInterface` (fallback to `udhcpc -i wlanInterface`).
- Store the network ID in the application state for disconnect.

**Feature 4 (Disconnect):**
- Use `exec.Command("wpa_cli", "-i", wlanInterface, "disconnect")`.
- Optionally, call `wpa_cli remove_network 0` to forget the network.
- Optionally, call `ip link set wlan0 down` to deactivate the interface.

---

#### References

- [Go os/exec package](https://pkg.go.dev/os/exec)
- [Linux wireless `iw` documentation](https://wireless.wiki.kernel.org/en/users/documentation/iw)
- [wpa_cli manual](https://linux.die.net/man/1/wpa_cli)
- [dhcpcd manual](https://linux.die.net/man/8/dhcpcd)
- [Go net package](https://pkg.go.dev/net)

---

### Interactive TUI Selection

**Date:** 2026-09-06  
**Status:** Confirmed

---

#### Context / Background

The CLI version of `gonet` works well for one‑off commands, but daily usage requires a more interactive experience: constantly updated list of available networks, arrow‑key navigation, and a simple way to select and connect. I wanted to add a full‑screen terminal interface that runs inside the same terminal window, without requiring a graphical environment.

The constraints:
- Must run in a standard terminal (no X11/Wayland required).
- Must handle periodic scanning without blocking the UI.
- Should be keyboard‑driven (arrows, Enter, Esc, Ctrl+C).
- Must integrate with the existing `scan`, `connect`, `disconnect` backends.

---

#### Options Considered

**Option A: BubbleTea (`charmbracelet/bubbletea`)**

| Aspect | Assessment |
|--------|------------|
| **Advantages** | • Pure Go, compiles to single binary <br> • Modern Elm‑architecture model <br> • Excellent built‑in components (`list`, `textinput`, `help`) <br> • Handles terminal resize and signals cleanly <br> • Active community and documentation |
| **Disadvantages** | • Requires learning the update/view pattern <br> • Slightly different from traditional imperative TUI |
| **Implementation Difficulty** | Medium |
| **Fit with Constraints** | ✅ Perfect. Provides all needed interactivity with minimal boilerplate. |

**Option B: `tview` / `cview`**

| Aspect | Assessment |
|--------|------------|
| **Advantages** | • Mature library with many primitives <br> • Supports mouse clicks <br> • Traditional widget‑based approach |
| **Disadvantages** | • More verbose for simple list scenarios <br> • Requires manual layout management <br> • Slightly heavier terminal interaction |
| **Implementation Difficulty** | Medium‑Hard |
| **Fit with Constraints** | ✅ Good, but BubbleTea is more modern and lighter. |

**Option C: `gocui` (Minimalist)**

| Aspect | Assessment |
|--------|------------|
| **Advantages** | • Extremely lightweight <br> • Total control over screen rendering |
| **Disadvantages** | • No built‑in list widget <br> • Must manually handle scrolling, highlighting, input <br> • High boilerplate |
| **Implementation Difficulty** | Hard |
| **Fit with Constraints** | ❌ Overkill for this use case. |

---

#### Decision & Rationale

**Chosen Option:** BubbleTea

**Reasoning:**

BubbleTea gives the best balance of developer experience, maintainability, and end‑user polish. Its built‑in `list` component handles all the heavy lifting (scrolling, pagination, filtering, key handling), and the `textinput` component provides a secure password prompt. The periodic auto‑refresh is elegantly handled by sending a tick command every 5 seconds.

I chose BubbleTea over the alternatives because it will produce a professional, responsive interface with minimal code, and its single‑binary nature aligns perfectly with `gonet`'s distribution philosophy.

**Trade‑offs accepted:**
- **Learning curve** – I’ll need to understand the `tea` model, but it's worth the investment.
- **No mouse support by default** – Keyboard is sufficient for a terminal tool.
- **Slightly larger binary** – The BubbleTea dependency adds a few MB, but it's still negligible.

---

#### Implementation Notes

- New command: `gonet tui`.
- Model contains: a `list.Model` for networks, a `textinput.Model` for password input, a `scanning` state flag, and a `connecting` flag.
- Periodic tick sends a `ScanMsg` that calls `scan.ScanInterface()` in a separate goroutine and sends the result back via a `tea.Cmd`.
- On selection, the model transitions to a password prompt; pressing Enter calls `connect.ConnectInterface()` and shows a status message.
- Press Esc or Ctrl+C to quit.

---

#### References

- [BubbleTea documentation](https://github.com/charmbracelet/bubbletea)
- [BubbleTea list component](https://github.com/charmbracelet/bubbles/list)
- [BubbleTea textinput component](https://github.com/charmbracelet/bubbles/textinput)

---

### Backlog: Systemd‑Free Auto‑Connect Daemon

**Date:** 2026-09-06  
**Status:** Planned (post‑v1.0)

---

#### Context / Background

After the interactive TUI, the next logical step is to make `gonet` automatically connect to known networks on boot or when a known network comes into range. This must be done **without systemd** to maintain distro‑agnosticism.

**Goal:** A subcommand `gonet daemon` that runs in the background, periodically scans, and auto‑connects to trusted networks defined in a config file.

**Proposed solution (pure Go + init‑script integration):**
1. The daemon itself is a pure Go program that runs continuously in a loop (e.g., scan every 10 seconds).
2. It reads a TOML config file (`~/.gonet/config.toml`) with a list of networks and their PSKs.
3. When it detects a known SSID, it triggers the same `connect.ConnectInterface()` flow.
4. For startup integration (no systemd):
   - Provide a `cron @reboot` entry that starts the daemon.
   - Provide an OpenRC init script (`/etc/init.d/gonet`).
   - Provide a Runit run script (`/etc/sv/gonet/run`).
   - Document manual `nohup` usage.

This approach guarantees the daemon works on **any** Linux distribution, regardless of init system.

---

#### Implementation Notes (Future)

- Add a `internal/daemon/` package with a `Run()` function that loops and scans.
- Config file parsing using `BurntSushi/toml` (pure Go).
- Use `go-homedir` to locate `~/.gonet/config.toml`.
- Logging to `/var/log/gonet.log` or syslog via `log/slog`.
- Graceful shutdown on SIGTERM/SIGINT.

---

#### References

- [TOML spec](https://toml.io/en/)
- [golang‑cron examples](https://pkg.go.dev/github.com/robfig/cron/v3) (for scheduling inside the daemon)
- OpenRC documentation

---

## 📝 Review / Update Log (Single, Unified)

| Date | Update | Author |
|------|--------|--------|
| 2026-09-05 | Initial diary created with backend strategy and four core feature decisions | deltaog-117 |
| 2026-09-06 | Added decision entry for interactive TUI (BubbleTea) | deltaog-117 |
| 2026-09-06 | Added backlog entry for systemd‑free auto‑connect daemon | deltaog-117 |
| 2026-09-06 | Consolidated multiple review logs into a single unified table | deltaog-117 |
