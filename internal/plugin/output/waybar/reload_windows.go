//go:build windows

package waybar

import (
	"context"
	"fmt"

	"github.com/jmylchreest/tinct/internal/plugin/output"
)

// PostExecute is a no-op on Windows since waybar doesn't run on this platform.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, _ output.ExecutionContext, _ []string) error {
	return fmt.Errorf("waybar reload is not supported on Windows")
}
