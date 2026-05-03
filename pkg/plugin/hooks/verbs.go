package hooks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// runReloadVerb dispatches a ReloadSpec to its verb implementation. An
// empty Verb is treated as VerbExec, so plugins can write
// {Args: ["dunstctl", "reload"]} without spelling out the verb.
func runReloadVerb(ctx context.Context, r ReloadSpec, verbose bool) {
	switch r.Verb {
	case VerbExec, "":
		runExecVerb(ctx, r.Args, verbose)
	case VerbSignal:
		runSignalVerb(r.Args, verbose)
	default:
		if verbose {
			fmt.Fprintf(os.Stderr, "   Warning: unknown reload verb %q\n", r.Verb)
		}
	}
}

func runExecVerb(ctx context.Context, args []string, verbose bool) {
	if len(args) == 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "   Warning: exec verb requires at least one argument\n")
		}
		return
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec // verb args declared by the plugin, not user input
	if err := cmd.Run(); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "   Warning: %s failed: %v\n", args[0], err)
	}
}
