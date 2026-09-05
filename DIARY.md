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

## 📝 Review / Update Log (Single, Unified)

| Date | Update | Author |
|------|--------|--------|
| 2026-09-05 | Initial diary created with backend strategy and four core feature decisions | deltaog-117 |
| 2026-09-05 | Added decision entries for interactive TUI (BubbleTea) and daemon backlog | deltaog-117 |
| 2026-09-05 | Added TUI redesign, SSID encoding, and retry logic entries | deltaog-117 |
| 2026-09-05 | Consolidated review logs into a single table | deltaog-117 |

---
