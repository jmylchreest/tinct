package telemetry

import (
	"github.com/jmylchreest/tinct/internal/version"
)

// Status string constants used as the value of plugin.status /
// plugins.status across telemetry events. Centralised here so the same
// vocabulary appears on every event, and the spelling can't drift.
const (
	StatusOK      = "ok"
	StatusFailed  = "failed"
	StatusSkipped = "skipped"
)

// Event represents a telemetry event to be sent to statsfactory.
type Event struct {
	// Name is the event name (e.g. "generate", "plugin_used").
	Name string

	// Props contains dimensions for this event using dot-notation grouping.
	// Values should be strings, numbers, or booleans.
	Props map[string]any
}

// NewEvent creates a new event with the given name.
func NewEvent(name string) Event {
	return Event{
		Name:  name,
		Props: make(map[string]any),
	}
}

// Set adds a property to the event. Returns the event for chaining.
func (e Event) Set(key string, value any) Event {
	e.Props[key] = value
	return e
}

// GenerateEventParams contains the parameters for a generate telemetry event.
type GenerateEventParams struct {
	// InputPlugin is the name of the input plugin used (e.g. "image", "google-genai").
	InputPlugin string

	// OutputPlugins is the list of output plugin names used.
	OutputPlugins []string

	// ThemeType is the detected theme type (e.g. "dark", "light").
	ThemeType string

	// SeedMode is the seed mode used for colour generation.
	SeedMode string

	// Backend is the colour backend used for generation.
	Backend string

	// ExtractAmbience indicates whether ambience extraction was enabled.
	ExtractAmbience bool

	// DryRun indicates whether this was a dry-run invocation.
	DryRun bool

	// DualTheme indicates whether dual-theme (light+dark) generation was enabled.
	DualTheme bool
}

// NewGenerateEvent creates a telemetry event for the "generate" command.
//
// Dimensions use dot-notation grouping for statsfactory:
//   - app.version         — application version
//   - input.plugin        — input plugin name
//   - input.ai            — whether the input is AI-powered
//   - output.plugins      — list of output plugin names ([]string)
//   - generate.theme_type — detected theme type
//   - generate.seed_mode  — seed mode
//   - generate.backend    — colour backend
//   - generate.ambience   — ambience extraction enabled
//   - generate.dry_run    — dry-run mode
//   - generate.dual_theme — dual-theme generation enabled
func NewGenerateEvent(params GenerateEventParams) Event {
	e := NewEvent("generate")

	// App-level dimensions.
	e.Props["app.version"] = version.Version

	// Input dimensions.
	if params.InputPlugin != "" {
		e.Props["input.plugin"] = params.InputPlugin
	}
	e.Props["input.ai"] = isAIInput(params.InputPlugin)

	// Output dimensions.
	if len(params.OutputPlugins) > 0 {
		e.Props["output.plugins"] = params.OutputPlugins
	}

	// Generate dimensions.
	if params.ThemeType != "" {
		e.Props["generate.theme_type"] = params.ThemeType
	}
	if params.SeedMode != "" {
		e.Props["generate.seed_mode"] = params.SeedMode
	}
	if params.Backend != "" {
		e.Props["generate.backend"] = params.Backend
	}
	e.Props["generate.ambience"] = params.ExtractAmbience
	e.Props["generate.dry_run"] = params.DryRun
	e.Props["generate.dual_theme"] = params.DualTheme

	return e
}

// NewPluginUsedEvent creates a telemetry event for a single output plugin execution.
//
// Dimensions use dot-notation grouping for statsfactory:
//   - plugin.name     — plugin name
//   - plugin.version  — plugin version (omitted if empty)
//   - plugin.external — whether the plugin is external (contrib)
//   - plugin.status   — execution status: "ok", "failed", or "skipped"
func NewPluginUsedEvent(pluginName, pluginVersion string, isExternal bool, status string) Event {
	e := NewEvent("plugin_used")
	e.Props["plugin.name"] = pluginName
	e.Props["plugin.external"] = isExternal
	e.Props["plugin.status"] = status
	if pluginVersion != "" {
		e.Props["plugin.version"] = pluginVersion
	}
	return e
}

// PluginCommandEventParams describes one invocation of a `tinct plugins
// <subcommand>` (or sub-subcommand) handler. It powers the
// "plugins_command" telemetry event, which we use to learn which
// subcommands actually have a real audience — driving keep / simplify /
// remove decisions for the curated-index machinery.
type PluginCommandEventParams struct {
	// Subcommand is the top-level subcommand name as the user typed it
	// (e.g. "list", "add", "install", "sync", "repo", "templates").
	Subcommand string

	// Action is the leaf-level verb when the top-level subcommand has
	// further sub-subcommands. For "plugins repo add foo" this is "add";
	// for plain top-level subcommands it stays empty.
	Action string

	// Status is "ok", "failed", or "skipped" (or empty for handlers that
	// don't have a meaningful binary outcome).
	Status string

	// Items counts the number of entities the command operated on
	// (repos listed, plugins synced, search results returned, …).
	Items int

	// ReposConfigured is the number of plugin repositories the user has
	// configured at the time of the call. Recorded for every subcommand
	// so we can correlate "is this user a curated-repo user?" with their
	// other behaviour.
	ReposConfigured int
}

// NewPluginCommandEvent creates a "plugins_command" telemetry event.
//
// Dimensions use dot-notation grouping for statsfactory:
//   - app.version          — application version
//   - plugins.subcommand   — top-level plugins subcommand
//   - plugins.action       — leaf verb for nested subcommands (omitted if empty)
//   - plugins.status       — "ok" | "failed" | "skipped" (omitted if empty)
//   - plugins.items        — count of items the command operated on
//   - plugins.repos_configured — count of configured plugin repositories
func NewPluginCommandEvent(p PluginCommandEventParams) Event {
	e := NewEvent("plugins_command")
	e.Props["app.version"] = version.Version
	e.Props["plugins.subcommand"] = p.Subcommand
	if p.Action != "" {
		e.Props["plugins.action"] = p.Action
	}
	if p.Status != "" {
		e.Props["plugins.status"] = p.Status
	}
	e.Props["plugins.items"] = p.Items
	e.Props["plugins.repos_configured"] = p.ReposConfigured
	return e
}

// isAIInput returns true if the input plugin is AI-powered.
func isAIInput(inputPlugin string) bool {
	switch inputPlugin {
	case "google-genai", "openrouter":
		return true
	default:
		return false
	}
}
