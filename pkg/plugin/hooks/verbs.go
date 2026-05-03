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
//
// pluginName is forwarded to the verb runners so verbose-mode messages
// carry the same "   <plugin>: ..." prefix as the rest of the runner.
func runReloadVerb(ctx context.Context, r ReloadSpec, pluginName string, verbose bool) {
	switch r.Verb {
	case VerbExec, "":
		runExecVerb(ctx, r.Args, pluginName, verbose)
	case VerbSignal:
		runSignalVerb(r.Args, pluginName, verbose)
	default:
		if verbose {
			fmt.Fprintf(os.Stderr, "%swarning: unknown reload verb %q\n", linePrefix(pluginName), r.Verb)
		}
	}
}

func runExecVerb(ctx context.Context, args []string, pluginName string, verbose bool) {
	if len(args) == 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, "%swarning: exec verb requires at least one argument\n", linePrefix(pluginName))
		}
		return
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec // verb args declared by the plugin, not user input
	if err := cmd.Run(); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "%swarning: %s failed: %v\n", linePrefix(pluginName), args[0], err)
	}
}
