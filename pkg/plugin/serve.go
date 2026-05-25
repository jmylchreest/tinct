// Package plugin provides the public API for tinct plugins.
package plugin

import (
	"os"
	"time"

	"github.com/hashicorp/go-plugin"
)

const parentDeathPollInterval = time.Second

// Serve wraps plugin.Serve with a safety-net that exits the plugin if
// the host dies without sending Quit (crash, SIGKILL). go-plugin's
// main loop only unblocks on Quit, so without this an orphaned plugin
// holds the binary open and blocks future `tinct plugins install`
// with ETXTBSY.
func Serve(opts *plugin.ServeConfig) {
	go watchParentDeath(os.Getppid(), parentDeathPollInterval)
	plugin.Serve(opts)
}

// watchParentDeath exits when the parent PID changes — the kernel
// reparents us to PID 1 once the original host dies.
func watchParentDeath(initial int, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if os.Getppid() != initial {
			// Process is exiting; the deferred ticker.Stop is moot.
			os.Exit(0) //nolint:gocritic // exitAfterDefer is fine here; we're terminating
		}
	}
}
