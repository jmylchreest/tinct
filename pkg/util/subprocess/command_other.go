//go:build !unix

package subprocess

import "os/exec"

// Other platforms retain os/exec's direct-child cancellation. WaitDelay still
// bounds pipe draining; process-tree termination is only supported on Unix.
func configureCancellation(_ *exec.Cmd) {}
