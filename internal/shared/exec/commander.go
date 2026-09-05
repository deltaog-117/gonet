package exec

import (
	"bytes"
	"os/exec"
)

// Commander defines the interface for executing system commands.
// This abstraction allows us to swap the real exec with a mock for testing.
type Commander interface {
	// Run executes a command and returns its stdout, stderr, and any error.
	Run(name string, args ...string) (stdout string, stderr string, err error)
}

// RealCommander is the production implementation that uses os/exec.
type RealCommander struct{}

// Run executes the command using os/exec.Command.
func (r *RealCommander) Run(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// MockCommander is a test implementation that returns predefined responses.
type MockCommander struct {
	// RunFunc allows callers to define custom behavior for each test.
	RunFunc func(name string, args ...string) (stdout string, stderr string, err error)
}

// Run executes the mock function if set; otherwise returns empty strings.
func (m *MockCommander) Run(name string, args ...string) (string, string, error) {
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	return "", "", nil
}
