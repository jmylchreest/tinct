package executor

import (
	"context"
	"errors"
	"io"
	"os/exec"
)

// ProcessRunner defines an interface for running external processes.
// This abstraction allows for dependency injection and easier testing.
type ProcessRunner interface {
	// Run executes a command with the given context, arguments, stdin, and returns stdout/stderr.
	Run(ctx context.Context, path string, args []string, stdin io.Reader) (stdout, stderr []byte, err error)
}

// RealProcessRunner implements ProcessRunner using actual os/exec commands.
type RealProcessRunner struct{}

// Run executes a real external process.
func (r *RealProcessRunner) Run(ctx context.Context, path string, args []string, stdin io.Reader) (stdout, stderr []byte, err error) {
	cmd := exec.CommandContext(ctx, path, args...) // #nosec G204 -- path is the plugin binary path from the lock file; exec.CommandContext does not invoke a shell so args cannot cause injection
	cmd.Stdin = stdin

	stdout, err = cmd.Output()
	if err != nil {
		// Output() returns stderr in the error if it's an ExitError
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return stdout, exitErr.Stderr, err
		}
		return stdout, nil, err
	}

	return stdout, nil, nil
}

// NewRealProcessRunner creates a new real process runner.
func NewRealProcessRunner() *RealProcessRunner {
	return &RealProcessRunner{}
}
