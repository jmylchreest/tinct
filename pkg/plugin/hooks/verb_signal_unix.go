//go:build unix

package hooks

import (
	"fmt"
	"os"
	"syscall"

	ps "github.com/mitchellh/go-ps"
)

// runSignalVerb sends a Unix signal to all running processes whose
// executable basename matches args[0]. args[1] is the signal name
// (SIGUSR1, SIGUSR2, SIGHUP, SIGTERM). Used by reload-on-signal apps
// like kitty (SIGUSR1) and ghostty (SIGUSR2). On non-Unix platforms a
// stub in verb_signal_other.go warns and returns.
func runSignalVerb(args []string, pluginName string, verbose bool) {
	prefix := linePrefix(pluginName)
	if len(args) < 2 {
		if verbose {
			fmt.Fprintf(os.Stderr, "%swarning: signal verb requires [process_name, signal]\n", prefix)
		}
		return
	}

	procName, sigName := args[0], args[1]
	sig := parseSignal(sigName)
	if sig == nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "%swarning: unknown signal: %s\n", prefix, sigName)
		}
		return
	}

	procs, err := ps.Processes()
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "%swarning: failed to list processes: %v\n", prefix, err)
		}
		return
	}

	sent := 0
	for _, p := range procs {
		if p.Executable() != procName {
			continue
		}
		proc, ferr := os.FindProcess(p.Pid())
		if ferr != nil {
			continue
		}
		if sigErr := proc.Signal(sig); sigErr == nil {
			sent++
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "%ssent %s to %d %s instance(s)\n", prefix, sigName, sent, procName)
	}
}

func parseSignal(name string) os.Signal {
	switch name {
	case "SIGUSR1":
		return syscall.SIGUSR1
	case "SIGUSR2":
		return syscall.SIGUSR2
	case "SIGHUP":
		return syscall.SIGHUP
	case "SIGTERM":
		return syscall.SIGTERM
	case "SIGINT":
		return syscall.SIGINT
	}
	return nil
}
