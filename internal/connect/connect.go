package connect

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/deltaog-117/gonet/internal/shared/exec"
)

type Connector struct {
	commander exec.Commander
}

func New(commander exec.Commander) *Connector {
	return &Connector{commander: commander}
}

func (c *Connector) Connect(iface, ssid, psk string) error {
	cmd := c.commander.Command("wpa_cli", "-i", iface)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wpa_cli: %w", err)
	}

	commands := []string{
		"add_network",
		fmt.Sprintf("set_network 0 ssid \"%s\"", ssid),
		fmt.Sprintf("set_network 0 psk \"%s\"", psk),
		"select_network 0",
		"enable_network 0",
	}
	for _, cmdStr := range commands {
		if _, err := io.WriteString(stdin, cmdStr+"\n"); err != nil {
			return fmt.Errorf("failed to send command %q: %w", cmdStr, err)
		}
	}
	stdin.Close()

	// wpa_cli keeps retrying forever if it can't reach the wpa_supplicant
	// control socket (e.g. missing permissions), so the read loop runs in
	// its own goroutine and is bounded by setupTimeout below instead of
	// blocking Connect indefinitely.
	type readResult struct {
		failLine string
		err      error
	}
	done := make(chan readResult, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Could not connect to wpa_supplicant") {
				done <- readResult{err: fmt.Errorf("could not reach wpa_supplicant on %q; is it running, and do you have permission to access it? (try: sudo)", iface)}
				return
			}
			if strings.Contains(line, "FAIL") {
				done <- readResult{failLine: line}
				return
			}
		}
		done <- readResult{err: scanner.Err()}
	}()

	const setupTimeout = 10 * time.Second
	select {
	case res := <-done:
		if res.failLine != "" {
			return fmt.Errorf("wpa_cli command failed: %s", res.failLine)
		}
		if res.err != nil {
			return fmt.Errorf("error reading wpa_cli output: %w", res.err)
		}
	case <-time.After(setupTimeout):
		cmd.Process.Kill()
		return fmt.Errorf("wpa_cli did not respond within %v; is wpa_supplicant running, and do you have permission to access it? (try: sudo)", setupTimeout)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wpa_cli exited with error: %w (stderr: %s)", err, strings.TrimSpace(stderr.String()))
	}

	return c.waitForConnection(iface)
}

func (c *Connector) waitForConnection(iface string) error {
	const timeout = 30 * time.Second
	const interval = 200 * time.Millisecond
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		out, _, err := c.commander.Run("wpa_cli", "-i", iface, "status")
		if err != nil {
			time.Sleep(interval)
			continue
		}
		if strings.Contains(out, "wpa_state=COMPLETED") {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("connection timed out after %v", timeout)
}

func ConnectInterface(iface, ssid, psk string) error {
	connector := New(&exec.RealCommander{})
	return connector.Connect(iface, ssid, psk)
}
