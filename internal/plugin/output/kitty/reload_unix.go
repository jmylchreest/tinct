//go:build unix

package kitty

import (
	"fmt"
	"syscall"

	"github.com/mitchellh/go-ps"
)

// reloadAllKittyInstances sends SIGUSR1 to all kitty processes to reload config.
// This works on Unix-like systems (Linux, macOS, BSD).
func (p *Plugin) reloadAllKittyInstances() error {
	// Find kitty process PIDs using native Go.
	pids, err := findProcessByName("kitty")
	if err != nil {
		return fmt.Errorf("failed to find kitty processes: %w", err)
	}

	if len(pids) == 0 {
		return fmt.Errorf("no running kitty instances found")
	}

	// Send SIGUSR1 to all kitty instances to reload config.
	for _, pid := range pids {
		if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
			return fmt.Errorf("failed to send reload signal to kitty (PID %d): %w", pid, err)
		}
	}

	return nil
}

// findProcessByName finds all processes with the given executable name.
func findProcessByName(name string) ([]int, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %w", err)
	}

	var pids []int
	for _, p := range processes {
		if p.Executable() == name {
			pids = append(pids, p.Pid())
		}
	}

	return pids, nil
}
