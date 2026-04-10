package telemetry

import (
	"github.com/jmylchreest/tinct/internal/version"
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

// isAIInput returns true if the input plugin is AI-powered.
func isAIInput(inputPlugin string) bool {
	switch inputPlugin {
	case "google-genai", "openrouter":
		return true
	default:
		return false
	}
}
