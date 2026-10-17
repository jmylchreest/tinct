// Package subprocess manages short-lived commands owned by a caller's context.
// It must not be used to launch persistent daemons or RPC plugin sessions.
package subprocess

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
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

// Probe runs a short discovery command with a five-second deadline.
func Probe(ctx context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	data, err := CommandContext(ctx, name, args...).Output()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("probe %s: %w", name, ctx.Err())
	}
	if err != nil {
		return nil, fmt.Errorf("probe %s: %w", name, err)
	}
	return data, nil
}

// ValidPath reports whether path can be passed as a literal process argument.
// It is not a sandbox containment check or an escaping function for shell code.
func ValidPath(path string) bool { return path != "" && !strings.ContainsRune(path, 0) }

// ShellQuote quotes a literal for a POSIX shell, including embedded apostrophes.
func ShellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'" }
