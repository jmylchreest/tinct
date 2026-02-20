package telemetry

import (
	"runtime"
	"strings"

	"github.com/jmylchreest/tinct/internal/version"
)

// Event represents a telemetry event to be sent to Aptabase.
type Event struct {
	// Name is the event name (e.g. "generate", "extract").
	Name string

	// Props contains arbitrary key-value properties for this event.
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

// GenerateEvent builds a telemetry event for the generate command.
// All parameters are optional — pass zero values for fields you don't have.
type GenerateEventParams struct {
	// InputPlugin is the name of the input plugin used (e.g. "image", "google-genai").
	InputPlugin string

	// OutputPlugins is the list of output plugin names that were executed.
	OutputPlugins []string

	// ThemeType is the detected/selected theme (e.g. "dark", "light", "auto").
	ThemeType string

	// SeedMode is the seed mode used (e.g. "content", "random", "manual").
	SeedMode string

	// Backend is the colour extraction backend (e.g. "kmeans").
	Backend string

	// ExtractAmbience indicates whether ambient edge extraction was enabled.
	ExtractAmbience bool

	// DryRun indicates whether this was a dry-run invocation.
	DryRun bool

	// DualTheme indicates whether both primary and alternate themes were generated.
	DualTheme bool
}

// NewGenerateEvent creates a comprehensive telemetry event for a generate invocation.
func NewGenerateEvent(params GenerateEventParams) Event {
	e := NewEvent("generate")

	// System info (supplementary to Aptabase's systemProps).
	e.Props["arch"] = runtime.GOARCH
	e.Props["version"] = version.Version

	// Input configuration.
	if params.InputPlugin != "" {
		e.Props["input_plugin"] = params.InputPlugin
	}

	// Output configuration.
	if len(params.OutputPlugins) > 0 {
		e.Props["output_plugins"] = strings.Join(params.OutputPlugins, ",")
	}

	// Theme and colour settings.
	if params.ThemeType != "" {
		e.Props["theme_type"] = params.ThemeType
	}
	if params.SeedMode != "" {
		e.Props["seed_mode"] = params.SeedMode
	}
	if params.Backend != "" {
		e.Props["backend"] = params.Backend
	}

	// Feature flags.
	e.Props["extract_ambience"] = params.ExtractAmbience
	e.Props["dry_run"] = params.DryRun
	e.Props["dual_theme"] = params.DualTheme

	// Detect AI input usage.
	e.Props["ai_input"] = isAIInput(params.InputPlugin)

	return e
}

// NewPluginUsedEvent creates a telemetry event for an individual output plugin usage.
// Sending one event per plugin allows Aptabase to count and rank plugin popularity
// directly, rather than requiring string splitting on a comma-joined list.
//
// Status should be one of:
//   - "ok"      — plugin generated and wrote files successfully
//   - "failed"  — plugin was attempted but generation or write failed
//   - "skipped" — plugin was not attempted (validation failed, pre-hook skip, etc.)
func NewPluginUsedEvent(pluginName string, pluginVersion string, isExternal bool, status string) Event {
	e := NewEvent("plugin_used")
	e.Props["plugin_name"] = pluginName
	e.Props["is_external"] = isExternal
	e.Props["status"] = status
	if pluginVersion != "" {
		e.Props["plugin_version"] = pluginVersion
	}
	return e
}

// isAIInput returns true if the input plugin is an AI image generation plugin.
func isAIInput(inputPlugin string) bool {
	switch inputPlugin {
	case "google-genai", "openrouter":
		return true
	default:
		return false
	}
}
