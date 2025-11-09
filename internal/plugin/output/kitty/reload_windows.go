//go:build windows

package kitty

import (
	"fmt"
)

// reloadAllKittyInstances is not supported on Windows.
// Kitty on Windows doesn't support SIGUSR1 signals.
func (p *Plugin) reloadAllKittyInstances() error {
	return fmt.Errorf("automatic reload is not supported on Windows - please manually restart kitty instances")
}
