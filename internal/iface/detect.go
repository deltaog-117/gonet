// Package iface detects which network interface on the system is wireless,
// so callers no longer need to hardcode a name like "wlan0".
package iface

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deltaog-117/gonet/internal/shared/exec"
)

// sysNetPath is the sysfs directory listing network interfaces.
// It's a var (not a const) so tests can point it at a fake sysfs tree.
var sysNetPath = "/sys/class/net"

// Detect returns the name of the first wireless interface found on the
// system. It checks sysfs first (no subprocess needed), and falls back to
// parsing `iw dev` if sysfs doesn't turn up a match.
func Detect(commander exec.Commander) (string, error) {
	if name, ok := detectFromSysfs(); ok {
		return name, nil
	}
	if name, ok := detectFromIwDev(commander); ok {
		return name, nil
	}
	return "", fmt.Errorf("no wireless interface found; specify one with -iface")
}

// detectFromSysfs looks for a network interface directory that has either a
// "wireless" subdirectory or a "phy80211" symlink, both of which the kernel
// only creates for wireless devices.
func detectFromSysfs() (string, bool) {
	entries, err := os.ReadDir(sysNetPath)
	if err != nil {
		return "", false
	}
	for _, entry := range entries {
		if isWireless(entry.Name()) {
			return entry.Name(), true
		}
	}
	return "", false
}

func isWireless(name string) bool {
	ifaceDir := filepath.Join(sysNetPath, name)
	if _, err := os.Stat(filepath.Join(ifaceDir, "wireless")); err == nil {
		return true
	}
	if _, err := os.Lstat(filepath.Join(ifaceDir, "phy80211")); err == nil {
		return true
	}
	return false
}

// detectFromIwDev shells out to `iw dev` and returns the first interface
// name it lists. Used when sysfs is unavailable (e.g. non-Linux, or a
// restricted environment).
func detectFromIwDev(commander exec.Commander) (string, bool) {
	stdout, _, err := commander.Run("iw", "dev")
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(stdout, "\n") {
		trimmed := strings.TrimSpace(line)
		if name, found := strings.CutPrefix(trimmed, "Interface "); found {
			return name, true
		}
	}
	return "", false
}
