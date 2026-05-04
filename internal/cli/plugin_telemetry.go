package cli

import (
	"context"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/telemetry"
)

// pluginTelemetry holds a single lazy-initialised telemetry.Client that's
// reused across every `tinct plugins <subcommand>` handler in this
// process. Each handler enqueues events via Send; a cobra.OnFinalize
// hook (registered in init below) flushes the client once at the end of
// Execute so events actually reach statsfactory before the CLI tears
// down. Without that flush the statsfactory background worker is
// killed mid-drain and queued events are silently dropped — that was
// the bug behind "I see generate events but no plugins_command events"
// up to commit 11c5686.
//
// The previous fix flushed inline in every handler, which worked but
// paid the 2-second timeout per Send and re-initialised the client on
// every call. Centralising into a singleton gives one init + one flush
// per CLI invocation regardless of how many events fire.
var pluginTelemetry struct {
	once   sync.Once
	client *telemetry.Client
}

// telemetryClient returns the shared telemetry client, initialising it
// on first use. Returns nil when telemetry is disabled (missing
// credentials, opt-out, etc.) so callers can skip work without an
// IsEnabled() check.
func telemetryClient() *telemetry.Client {
	pluginTelemetry.once.Do(func() {
		c := telemetry.New(telemetry.WithVerbose(globalVerbose()))
		if c.IsEnabled() {
			pluginTelemetry.client = c
		}
	})
	return pluginTelemetry.client
}

func init() {
	cobra.OnFinalize(func() {
		c := pluginTelemetry.client
		if c == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		c.Flush(ctx)
	})
}

// sendPluginCommandTelemetry emits a single "plugins_command" event for
// the just-completed `tinct plugins <subcommand>` invocation.
// Best-effort and silent — failures (telemetry disabled, repo config
// unreadable, network down) are swallowed so a CLI run never breaks
// because of telemetry. ReposConfigured is filled in automatically when
// the caller leaves it at zero so every event carries the correlation
// dimension regardless of subcommand. The actual send-to-wire flush
// happens once in cobra.OnFinalize.
func sendPluginCommandTelemetry(p telemetry.PluginCommandEventParams) {
	c := telemetryClient()
	if c == nil {
		return
	}

	if p.ReposConfigured == 0 {
		p.ReposConfigured = countConfiguredRepos()
	}

	c.Send(telemetry.NewPluginCommandEvent(p))
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
