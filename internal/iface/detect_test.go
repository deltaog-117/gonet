package iface

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deltaog-117/gonet/internal/shared/exec"
)

func withFakeSysNet(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	orig := sysNetPath
	sysNetPath = dir
	t.Cleanup(func() { sysNetPath = orig })
	return dir
}

func TestDetect_SysfsWirelessSubdir(t *testing.T) {
	dir := withFakeSysNet(t)
	mustMkdirAll(t, filepath.Join(dir, "eth0"))
	mustMkdirAll(t, filepath.Join(dir, "wlan0", "wireless"))

	mock := &exec.MockCommander{RunFunc: func(name string, args ...string) (string, string, error) {
		t.Fatal("iw dev should not be called when sysfs finds a match")
		return "", "", nil
	}}

	got, err := Detect(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wlan0" {
		t.Errorf("got %q, want %q", got, "wlan0")
	}
}

func TestDetect_SysfsPhy80211Symlink(t *testing.T) {
	dir := withFakeSysNet(t)
	mustMkdirAll(t, filepath.Join(dir, "eth0"))
	wlanDir := filepath.Join(dir, "wlp2s0")
	mustMkdirAll(t, wlanDir)
	if err := os.Symlink("/sys/devices/fake/phy80211", filepath.Join(wlanDir, "phy80211")); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	got, err := Detect(&exec.MockCommander{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wlp2s0" {
		t.Errorf("got %q, want %q", got, "wlp2s0")
	}
}

func TestDetect_SkipsNonWirelessInterfaces(t *testing.T) {
	dir := withFakeSysNet(t)
	mustMkdirAll(t, filepath.Join(dir, "lo"))
	mustMkdirAll(t, filepath.Join(dir, "eth0"))

	mock := &exec.MockCommander{RunFunc: func(name string, args ...string) (string, string, error) {
		if name != "iw" {
			t.Fatalf("unexpected command: %s", name)
		}
		return "Interface wlan1\n\ttype managed", "", nil
	}}

	got, err := Detect(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wlan1" {
		t.Errorf("got %q, want %q (fallback to iw dev)", got, "wlan1")
	}
}

func TestDetect_FallsBackToIwDev(t *testing.T) {
	withFakeSysNet(t) // empty/unreadable sysfs dir

	mock := &exec.MockCommander{RunFunc: func(name string, args ...string) (string, string, error) {
		return "phy#0\nInterface wlan0\n\tifindex 3\n\ttype managed", "", nil
	}}

	got, err := Detect(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "wlan0" {
		t.Errorf("got %q, want %q", got, "wlan0")
	}
}

func TestDetect_NoneFound(t *testing.T) {
	withFakeSysNet(t)

	mock := &exec.MockCommander{RunFunc: func(name string, args ...string) (string, string, error) {
		return "", "", nil
	}}

	_, err := Detect(mock)
	if err == nil {
		t.Fatal("expected an error when no wireless interface is found")
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create %s: %v", path, err)
	}
}
