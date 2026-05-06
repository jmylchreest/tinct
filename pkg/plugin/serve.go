// Package plugin provides the public API for tinct plugins.
package plugin

import (
	"os"
	"time"

	"github.com/hashicorp/go-plugin"
)

// parentDeathPollInterval is how often the safety-net goroutine checks
// whether the host process is still alive. One second is well below
// any human-perceptible latency for plugin shutdown and trivial cost
// (one syscall per second per plugin process, only while running).
const parentDeathPollInterval = time.Second

// Serve is the recommended entry point for an external plugin's main()
// — a thin wrapper around hashicorp/go-plugin's Serve that adds a
// safety-net for the case where the host process dies without sending
// the usual Quit RPC (segfault, SIGKILL, panic without recover).
//
// Normally the host calls executor.Close → client.Kill, which sends a
// graceful Quit over the control connection; the plugin's main select
// unblocks and Serve returns. That path is unchanged.
//
// Without a backstop, a host that disconnects without saying goodbye
// leaves the plugin process alive indefinitely as a zombie reparented
// to PID 1, holding the binary open against ETXTBSY and slowly
// accumulating across CLI invocations (we hit this with the random
// plugin's `tinct plugins install` failing because nine stale
// instances held the binary). go-plugin's main loop only listens for
// the Quit RPC; control-connection EOF doesn't unblock it.
//
// The fix is a portable poll: we record os.Getppid() at startup and
// re-check it once a second. When the parent dies the kernel reparents
// us to PID 1 (or the user's init), so a changed ppid is the
// universal signal "host is gone, exit voluntarily." No signals, no
// syscalls beyond getppid, no platform-specific code.
//
// Plugins should drop-in replace plugin.Serve(...) with
// tinctplugin.Serve(...). The signature matches go-plugin's exactly so
// no other call-site changes are required.
func Serve(opts *plugin.ServeConfig) {
	go watchParentDeath(os.Getppid(), parentDeathPollInterval)
	plugin.Serve(opts)
}

// watchParentDeath polls the parent PID and exits the process when it
// changes — which happens iff the original parent has died and the
// kernel has reparented us. Detection latency is up to one poll
// interval; the trade-off vs an instant kernel signal (Linux's
// PR_SET_PDEATHSIG) is portability and zero cgo / no platform-specific
// build tags.
func watchParentDeath(initial int, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		if os.Getppid() != initial {
			os.Exit(0)
		}
	}
}
