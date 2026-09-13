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
| 2026-09-05 | Interactive TUI (Network Selector) | BubbleTea | ✅ Confirmed |
| 2026-09-05 | TUI Redesign & Polishing | Minimalist layout with numbered entries, signal bars, security icons | ✅ Confirmed |
| 2026-09-05 | SSID Encoding | UTF-8 decoding with Latin-1 fallback | ✅ Confirmed |
| 2026-09-05 | Scan Retry Logic | Exponential backoff (1s, 2s, 4s) with friendly error messages | ✅ Confirmed |
| 2026-09-05 | Auto‑Connect Daemon | Pure Go + Cron/OpenRC/Runit (no systemd) | 🔄 Planned |
| 2026-09-13 | Permission vs. Busy Error Handling | Classify `Operation not permitted` separately; bound `connect` with a timeout | ✅ Confirmed |
| 2026-09-13 | Feature: Interface Auto-Detection | sysfs (`wireless`/`phy80211`) first, `iw dev` fallback | ✅ Confirmed |

---

## 📝 Decision Entries

### Backend Strategy & Feature Implementation (v1.0 Core)

**Date:** 2026-09-05  
**Status:** Confirmed

*(Full content preserved from previous version – see earlier commits)*

---

### Interactive TUI Selection

**Date:** 2026-09-05  
**Status:** Confirmed

*(Full content preserved from previous version – see earlier commits)*

---

### TUI Redesign & Polishing

**Date:** 2026-09-05  
**Status:** Confirmed

---

#### Context / Background

The first version of the TUI was functional but cluttered: long descriptions, redundant information, and poor alignment. After testing, it was clear that a cleaner, more compact layout would improve usability. The goal was to match the style of successful terminal apps like `ytunes` – numbered entries, compact signal bars, and intuitive keyboard shortcuts.

The constraints:
- Must display SSID, signal strength, security type, and connected status in a single line.
- Should use fixed‑width columns to avoid misalignment.
- Must handle SSIDs with extended characters (cedillas, accents, etc.) correctly.
- Should show a friendly status message when the interface is busy (e.g., scan retry).

---

#### Options Considered

| Aspect | Option A: Minimalist Redesign (fixed columns) | Option B: Keep previous layout with formatting tweaks |
|--------|-----------------------------------------------|------------------------------------------------------|
| **Advantages** | • Clean, professional look <br> • Fixed widths for perfect alignment <br> • Numbered entries for quick reference <br> • Easier to scan visually | • Minimal code changes <br> • Retains existing state management |
| **Disadvantages** | • Requires custom list delegate <br> • Slightly more code | • Still cluttered <br> • Alignment issues remain |
| **Difficulty** | Medium | Easy |
| **Fit** | ✅ Excellent – matches desired polish | ⚠️ Not enough improvement |

---

#### Decision & Rationale

**Chosen Option:** Minimalist Redesign with fixed columns

**Reasoning:**

The fixed‑column approach gives a clean, professional terminal interface that matches the user's expectations. It improves scanability, reduces visual noise, and handles long SSIDs gracefully by truncation. The numbered entries allow quick identification without relying on mouse or complex navigation.

**Trade‑offs accepted:**
- Custom delegate adds a few lines of code, but it's contained in one file.
- Truncation may hide parts of very long SSIDs, but the detail view (`d` key) shows the full name.

---

#### Implementation Notes

- Implemented `customDelegate` with fixed column widths: number (3 chars), SSID (28 chars), signal bars (6 chars), dBm (8 chars), security icon+type.
- Added numbering (`1.`, `2.`, ...) and connected indicator (`✔`) in the title line.
- Removed the separate description line – all info is now in the title.
- Shortened signal bars to exactly 6 chars (`██████`, `█████ `, etc.) for consistency.
- Added color coding for signal bars (green → yellow → orange → red).
- SSID truncation uses ellipsis (`…`) for names longer than 25 chars.

---

#### References

- [BubbleTea custom delegate example](https://github.com/charmbracelet/bubbletea/tree/master/examples/list-custom-delegate)
- [Lipgloss styling guide](https://github.com/charmbracelet/lipgloss)

---

### SSID Encoding & UTF-8 Handling

**Date:** 2026-09-05  
**Status:** Confirmed

---

#### Context / Background

Wi‑Fi SSIDs are transmitted as raw byte sequences, not necessarily UTF‑8. Common encodings include Latin‑1 (ISO‑8859‑1) for accented characters (e.g., `ç`, `é`, `á`). The `iw` tool often prints SSIDs with `\xHH` escape sequences for non‑ASCII bytes, which Go's default string conversion mishandles, resulting in garbled output like `C\xc3\xa7a` instead of `Cça`.

#### Options Considered

| Aspect | Option A: Direct `string([]byte)` conversion | Option B: UTF‑8 validation + Latin‑1 fallback | Option C: Always treat as Latin‑1 |
|--------|----------------------------------------------|-----------------------------------------------|-----------------------------------|
| **Advantages** | • Simplest code <br> • Works for pure ASCII | • Correctly displays most real‑world SSIDs <br> • Handles UTF‑8 and Latin‑1 | • Always works for Latin‑1 |
| **Disadvantages** | • Breaks on non‑UTF8 bytes | • Slightly more complex | • Breaks on pure UTF‑8 SSIDs |
| **Difficulty** | Easy | Medium | Easy |
| **Fit** | ❌ Fails on common SSIDs | ✅ Best of both worlds | ⚠️ May break valid UTF‑8 |

---

#### Decision & Rationale

**Chosen Option:** UTF‑8 validation + Latin‑1 fallback

**Reasoning:**

This approach validates the byte sequence as UTF‑8; if it passes, we display it directly. If it fails, we interpret the bytes as Latin‑1 (which maps 1‑to‑1 to Unicode code points 0‑255). This covers the vast majority of SSIDs found in the wild, including those with cedillas and accents, while still supporting pure UTF‑8 networks.

**Trade‑offs accepted:**
- Added complexity in the decoding function, but it's self‑contained.
- Some rare encodings (e.g., GBK) will still fail, but they are extremely uncommon for SSIDs.

---

#### Implementation Notes

- Added `decodeSSID()` function that checks UTF‑8 validity and falls back to Latin‑1.
- Added `unescapeSSID()` to convert `\xHH` sequences from `iw` output into raw bytes.
- Integrated these into the `parseIwScan()` flow.

---

#### References

- [Go strings and UTF‑8](https://go.dev/blog/strings)
- [Latin‑1 code chart](https://en.wikipedia.org/wiki/ISO/IEC_8859-1)

---

### Scan Retry Logic & Error Messaging

**Date:** 2026-09-05  
**Status:** Confirmed

---

#### Context / Background

The `iw scan` command frequently fails with `Device or resource busy` when the Wi‑Fi interface is in use (e.g., already scanning or connected). The raw error message is cryptic and unhelpful for end users.

#### Options Considered

| Aspect | Option A: No retry | Option B: Simple retry with fixed backoff | Option C: Exponential backoff + friendly message |
|--------|---------------------|-------------------------------------------|-------------------------------------------------|
| **Advantages** | • Simplest | • Improves success rate | • Best success rate <br> • User‑friendly |
| **Disadvantages** | • Frequent failures | • May still fail if busy for longer | • Slightly longer wait time |
| **Difficulty** | Easy | Easy | Medium |
| **Fit** | ❌ Poor UX | ⚠️ Better but not optimal | ✅ Best UX |

---

#### Decision & Rationale

**Chosen Option:** Exponential backoff (1s, 2s, 4s) + friendly message

**Reasoning:**

Retrying with increasing delays gives the interface time to become free. The friendly message `"Interface busy, retrying..."` informs the user without alarming them. The total wait of up to 7 seconds is acceptable for a network scan.

**Trade‑offs accepted:**
- Slightly longer wait on busy interfaces, but improves overall reliability.
- Additional code for backoff, but it's contained in the scanner.

---

#### Implementation Notes

- Implemented backoff slice `[]time.Duration{1, 2, 4} * time.Second`.
- In the TUI, the scan error is caught and transformed into a friendly status message.
- If all retries fail, show `"Interface is busy; please wait and try again"`.

---

#### References

- [Go time package](https://pkg.go.dev/time)

---

### Permission vs. Busy Error Handling

**Date:** 2026-09-13  
**Status:** Confirmed

---

#### Context / Background

The Scan Retry Logic decision (above) treated `Operation not permitted` exactly like `Device or resource busy`: both triggered the exponential backoff retry, and after all retries were exhausted, both produced the same generic `"interface is busy; please wait and try again"` message. In practice, `Operation not permitted` is returned by `iw` when the caller lacks `CAP_NET_ADMIN` (i.e., not running as root) — a permanent condition that retrying can never fix. A user running `gonet scan` without `sudo` would wait ~7 seconds through three pointless retries and then be told the interface was "busy," when the real problem was missing privileges. Separately, `connect.Connect()` never captured `wpa_cli`'s stderr and had no timeout around the interactive session, so if `wpa_cli` couldn't reach the `wpa_supplicant` control socket (the same privilege issue, or `wpa_supplicant` not running), it could hang indefinitely instead of failing with a useful message.

#### Decision & Rationale

**Chosen Option:** Classify `Operation not permitted` as a distinct, non-retryable error with an actionable message; only retry on genuine `Device or resource busy`. For `connect`, capture stderr and bound the whole `wpa_cli` exchange with a 10s timeout so a stuck control-socket connection fails loudly instead of hanging.

**Reasoning:**

> Retrying a permission error wastes time and produces a misleading diagnosis. Splitting the two cases lets each fail in a way that tells the user what to actually do: "requires root privileges (try: sudo)" for permissions, vs. the existing busy-retry-then-message flow for genuine transient contention. For `connect`, wiring up `cmd.Stderr` and adding a timeout turns a silent, indefinite hang into a clear, bounded failure.

**Trade‑offs accepted:**
- `connect` now spawns a goroutine to read `wpa_cli` stdout so it can race against a timeout; slightly more code, but contained to one function.

---

#### Implementation Notes

- `internal/scan/scan.go`: check `Operation not permitted` first and return immediately; only `Device or resource busy` continues the backoff loop.
- `internal/connect/connect.go`: `cmd.Stderr` now captured; stdout reading moved to a goroutine racing a 10s `setupTimeout`; both `wpa_cli`-reported `FAIL` and `Could not connect to wpa_supplicant` are detected explicitly.
- `cmd/gonet/tui.go`: fixed an unrelated dead-code duplicate in the `scanResultMsg` handler, and fixed the 5-second auto-refresh timer not rescheduling itself after the first tick (found while investigating the same bug report).

---

#### References

- Reported by the user via a CLI repro: `gonet scan` (no `sudo`) → `"Scan failed: interface is busy; please wait and try again"` after ~11s.

---

### Feature: Interface Auto-Detection

**Date:** 2026-09-13  
**Status:** Confirmed

---

#### Context / Background

Every command hardcoded `wlan0` as the default interface. On systems where the wireless card isn't named `wlan0` (e.g. `wlp2s0`, `wlp3s0` under systemd's predictable naming, or a second wireless adapter), users had to remember to pass `-iface` every time, and the tool's own error messages (like the busy/permission fix above) still referenced whatever name was guessed, which made a wrong default confusing to debug.

#### Options Considered

| Aspect | Option A: sysfs check only | Option B: `iw dev` parsing only | Option C: sysfs first, `iw dev` fallback |
|--------|------------------------------|----------------------------------|-------------------------------------------|
| **Advantages** | • No subprocess, instant <br> • Works even if `iw` isn't installed yet | • Single source of truth (same tool used elsewhere) | • Fast common case, no subprocess needed <br> • Still works if sysfs is unavailable or ambiguous |
| **Disadvantages** | • Relies on Linux-specific `/sys` layout | • Always shells out, slower | • Slightly more code (two code paths) |
| **Difficulty** | Easy | Easy | Easy-Medium |
| **Fit** | ✅ Good, but no fallback if layout differs | ⚠️ Works, but pays subprocess cost every run | ✅ Best of both |

#### Decision & Rationale

**Chosen Option:** sysfs check first (`/sys/class/net/*/wireless` or `phy80211` symlink), falling back to parsing `iw dev` — exactly as already specified in `ROADMAP.md`.

**Reasoning:**

> sysfs is the fastest and most reliable signal the kernel gives us for "is this a wireless NIC," with zero subprocess overhead. Falling back to `iw dev` covers any edge case where sysfs is unreadable or laid out unexpectedly. If neither finds anything, we don't hard-fail — we fall back to the historical `wlan0` default with a warning, so existing scripts/muscle memory that never pass `-iface` don't break outright; `-iface` still overrides auto-detection unconditionally.

**Trade-offs accepted:**
- Two detection code paths instead of one, but each is small and independently testable via the existing `Commander` mock (for `iw dev`) and a swappable `sysNetPath` var (for sysfs).

---

#### Implementation Notes

- New package `internal/iface` (feature-first, mirrors `scan`/`connect`/`status`).
- `Detect(commander exec.Commander) (string, error)` is the public entry point.
- `cmd/gonet/main.go`: `-iface` default changed from `"wlan0"` to `""`; when empty after parsing, `iface.Detect` runs and the result (or `wlan0` + a stderr warning on failure) becomes `globalIface` before command dispatch.
- Detection runs after the `-version`/`-help` early exits, so those paths stay fast and don't touch the filesystem or spawn `iw`.

---

#### References

- Confirmed working against this machine's real `/sys/class/net` (found `wlan0` via the `wireless` subdirectory, alongside `docker0`, `enp3s0`, `lo`).

---

## 📝 Review / Update Log (Single, Unified)

| Date | Update | Author |
|------|--------|--------|
| 2026-09-05 | Initial diary created with backend strategy and four core feature decisions | deltaog-117 |
| 2026-09-05 | Added decision entries for interactive TUI (BubbleTea) and daemon backlog | deltaog-117 |
| 2026-09-05 | Added TUI redesign, SSID encoding, and retry logic entries | deltaog-117 |
| 2026-09-05 | Consolidated review logs into a single table | deltaog-117 |
| 2026-09-13 | Added permission-vs-busy error handling entry (scan/connect fix) | deltaog-117 |
| 2026-09-13 | Added Interface Auto-Detection decision entry (v1.1.0) | deltaog-117 |

---
