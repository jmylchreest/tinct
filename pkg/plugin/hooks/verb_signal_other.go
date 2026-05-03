//go:build !unix

package hooks

import (
	"fmt"
	"os"
)

// runSignalVerb is a stub for non-Unix platforms. Signal-based reload
// (SIGUSR1, SIGUSR2, etc.) is fundamentally a Unix concept; on Windows
// applications use different reload mechanisms (named events, IPC).
// Plugins targeting Windows should use VerbExec (e.g. invoking a CLI
// reload command) or implement ReloadFn themselves.
func runSignalVerb(args []string, verbose bool) {
	if verbose {
		name := "<unknown>"
		if len(args) > 0 {
			name = args[0]
		}
		fmt.Fprintf(os.Stderr, "   Warning: signal-based reload not supported on this platform (target: %s)\n", name)
	}
}
