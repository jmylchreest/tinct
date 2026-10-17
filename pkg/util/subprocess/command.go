// Package subprocess manages short-lived commands owned by a caller's context.
// It must not be used to launch persistent daemons or RPC plugin sessions.
package subprocess

import (
	"context"
	"os/exec"
	"time"
)

// CommandContext creates a command with bounded pipe draining after cancellation
// or exit. On Unix, cancellation also kills descendants in the command's process
// group. Descendants that deliberately detach are not owned by that group; pipe
// draining is still bounded. Successful commands may leave background jobs alive.
func CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- callers supply an executable and argument vector; no shell is added.
	cmd.WaitDelay = time.Second
	configureCancellation(cmd)
	return cmd
}
