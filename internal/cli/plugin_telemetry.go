package cli

import (
	"context"
	"time"

	"github.com/jmylchreest/tinct/internal/telemetry"
)

// sendPluginCommandTelemetry emits a single "plugins_command" event for
// the just-completed `tinct plugins <subcommand>` invocation.
// Best-effort and silent — failures (telemetry disabled, repo config
// unreadable, network down) are swallowed so a CLI run never breaks
// because of telemetry. ReposConfigured is filled in automatically when
// the caller leaves it at zero so every event carries the correlation
// dimension regardless of subcommand.
//
// Verbose mode is picked up from the global --verbose flag so a `-v`
// invocation emits the "→ Sending telemetry..." trace next to the rest
// of the command output.
func sendPluginCommandTelemetry(p telemetry.PluginCommandEventParams) {
	client := telemetry.New(telemetry.WithVerbose(globalVerbose()))
	if !client.IsEnabled() {
		return
	}

	if p.ReposConfigured == 0 {
		p.ReposConfigured = countConfiguredRepos()
	}

	client.Send(telemetry.NewPluginCommandEvent(p))

	// Most plugin subcommands return to the shell immediately after
	// this call, so the statsfactory background worker would otherwise
	// be torn down before it drains the queue. Block on Flush with a
	// short timeout — telemetry is best-effort, but it's pointless to
	// emit events that never reach the wire.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client.Flush(ctx)
}

// globalVerbose returns the value of the persistent --verbose flag on
// the root command, or false if the lookup fails. The plugins
// subcommands all live under the root, so checking the persistent flag
// gives us a verbose reading regardless of which subcommand was
// invoked.
func globalVerbose() bool {
	if RootCmd == nil {
		return false
	}
	f := RootCmd.PersistentFlags().Lookup("verbose")
	if f == nil {
		return false
	}
	v, err := RootCmd.PersistentFlags().GetBool("verbose")
	if err != nil {
		return false
	}
	return v
}

// countConfiguredRepos returns the number of plugin repositories the
// user has configured. Best-effort: returns 0 on any error so it never
// surfaces a failure to the caller.
func countConfiguredRepos() int {
	mgr, err := getRepoManager()
	if err != nil {
		return 0
	}
	return len(mgr.ListRepositories())
}

// sendPluginSubcommandResult is the shorthand used at the end of every
// `tinct plugins <subcommand>` handler: pass the subcommand name, the
// optional leaf action (for nested subcommands like `repo <verb>`), the
// item count, and the handler's named-return err. Status is derived
// from err so handlers don't need an inline status variable.
func sendPluginSubcommandResult(subcommand, action string, items int, err error) {
	status := telemetry.StatusOK
	if err != nil {
		status = telemetry.StatusFailed
	}
	sendPluginCommandTelemetry(telemetry.PluginCommandEventParams{
		Subcommand: subcommand,
		Action:     action,
		Status:     status,
		Items:      items,
	})
}
