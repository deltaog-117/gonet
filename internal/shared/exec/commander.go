package exec

import (
	"bytes"
	"os/exec"
)

// Commander defines the interface for executing system commands.
type Commander interface {
	Run(name string, args ...string) (stdout string, stderr string, err error)
	Command(name string, args ...string) *exec.Cmd // For interactive usage
}

// RealCommander is the production implementation.
type RealCommander struct{}

func (r *RealCommander) Run(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (r *RealCommander) Command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// MockCommander for testing.
type MockCommander struct {
	RunFunc func(name string, args ...string) (stdout string, stderr string, err error)
}

func (m *MockCommander) Run(name string, args ...string) (string, string, error) {
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return "", "", nil
}

func (m *MockCommander) Command(name string, args ...string) *exec.Cmd {
	// Return a dummy command that does nothing
	return exec.Command("echo")
}
