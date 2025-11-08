//go:build unix

package waybar

import (
	"context"
	"fmt"
	"syscall"

	"github.com/jmylchreest/tinct/internal/plugin/output"
)

// PostExecute reloads waybar configuration if requested.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, _ output.ExecutionContext, _ []string) error {
	if !p.reloadConfig {
		return nil
	}

	// Find waybar process PIDs using native Go.
	pids, err := findProcessByName("waybar")
	if err != nil {
		return fmt.Errorf("failed to find waybar processes: %w", err)
	}

	if len(pids) == 0 {
		return fmt.Errorf("no running waybar instances found to reload")
	}

	// Send SIGUSR2 to all waybar instances to reload config using native syscall.
	for _, pid := range pids {
		if err := syscall.Kill(pid, syscall.SIGUSR2); err != nil {
			return fmt.Errorf("failed to send reload signal to waybar (PID %d): %w", pid, err)
		}
	}

	return nil
}
