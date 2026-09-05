package disconnect

import (
	"fmt"

	"github.com/deltaog-117/gonet/internal/shared/exec"
)

// Disconnector handles Wi-Fi disconnections.
type Disconnector struct {
	commander exec.Commander
}

// New creates a new Disconnector.
func New(commander exec.Commander) *Disconnector {
	return &Disconnector{commander: commander}
}

// Disconnect disconnects from the current network.
func (d *Disconnector) Disconnect(iface string) error {
	_, stderr, err := d.commander.Run("wpa_cli", "-i", iface, "disconnect")
	if err != nil {
		return fmt.Errorf("failed to disconnect: %w (stderr: %s)", err, stderr)
	}
	return nil
}

// DisconnectInterface is a convenience function using the real Commander.
func DisconnectInterface(iface string) error {
	disconnector := New(&exec.RealCommander{})
	return disconnector.Disconnect(iface)
}
