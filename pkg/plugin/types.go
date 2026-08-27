// Package plugin provides the public API for tinct plugins.
package plugin

import "github.com/jmylchreest/tinct/pkg/plugin/hooks"

// HookSpecPayload is the gob-safe subset of hooks.Spec used by the
// HooksProvider RPC. It mirrors hooks.Spec without the function fields
// (ReloadFn, InstructionsFn, Wallpaper) which net/rpc's gob codec
// cannot transmit. Helpers HookSpecFromSpec / HookSpecToSpec convert
// between the payload and the live hooks.Spec.
type HookSpecPayload struct {
	RequiredBinaries  []string
	OptionalBinaries  []string
	RequiredDirs      []string
	RequiredAny       []hooks.AnyOf
	AutoCreateDir     bool
	Reload            *hooks.ReloadSpec
	MakeExecutable    []string
	SupportsWallpaper bool
	Instructions      string
}

// HookSpecFromSpec converts a live hooks.Spec into a transmissible
// HookSpecPayload, dropping non-marshalable function fields.
func HookSpecFromSpec(s hooks.Spec) HookSpecPayload {
	return HookSpecPayload{
		RequiredBinaries:  s.RequiredBinaries,
		OptionalBinaries:  s.OptionalBinaries,
		RequiredDirs:      s.RequiredDirs,
		RequiredAny:       s.RequiredAny,
		AutoCreateDir:     s.AutoCreateDir,
		Reload:            s.Reload,
		MakeExecutable:    s.MakeExecutable,
		SupportsWallpaper: s.SupportsWallpaper,
		Instructions:      s.Instructions,
	}
}

// HookSpecToSpec rehydrates a HookSpecPayload into a hooks.Spec the
// shared hook runner can execute. Function fields stay nil — the runner
// already treats nil function fields as no-ops.
func HookSpecToSpec(p HookSpecPayload) hooks.Spec {
	return hooks.Spec{
		RequiredBinaries:  p.RequiredBinaries,
		OptionalBinaries:  p.OptionalBinaries,
		RequiredDirs:      p.RequiredDirs,
		RequiredAny:       p.RequiredAny,
		AutoCreateDir:     p.AutoCreateDir,
		Reload:            p.Reload,
		MakeExecutable:    p.MakeExecutable,
		SupportsWallpaper: p.SupportsWallpaper,
		Instructions:      p.Instructions,
	}
}

// ConfigureRequest carries the host's per-run configuration to a plugin
// before any other lifecycle call.
//
// Args, DryRun and Verbose also travel inside PaletteData on Generate;
// this delivers them earlier, so a plugin's PreExecute and Hooks can act
// on flags such as an overridden output directory instead of having to
// assume defaults. It exists because those two RPCs carry no arguments
// of their own and, historically, ran in a throwaway process that could
// not have learned anything anyway.
type ConfigureRequest struct {
	Args    map[string]any
	DryRun  bool
	Verbose bool
}

// Configurable is an optional interface plugins may implement to receive
// the host's per-run configuration before PreExecute / Hooks / Generate.
//
// Implementations should only record the values; the host treats a
// Configure error as non-fatal and continues the run. Plugins that do
// not implement it keep working unchanged — they simply see their args
// for the first time in Generate, as before.
type Configurable interface {
	Configure(req ConfigureRequest) error
}

// InputOptions holds options for input plugin generation.
type InputOptions struct {
	Verbose         bool           `json:"verbose"`
	DryRun          bool           `json:"dry_run"`
	Backend         string         `json:"backend,omitempty"`
	ColourOverrides []string       `json:"colour_overrides,omitempty"`
	PluginArgs      map[string]any `json:"plugin_args,omitempty"`
}

// PaletteData is the palette data sent to output plugins.
type PaletteData struct {
	Colours    map[string]CategorisedColour `json:"colours"`
	AllColours []CategorisedColour          `json:"all_colours"`
	ThemeType  string                       `json:"theme_type"`
	PluginArgs map[string]any               `json:"plugin_args,omitempty"`
	DryRun     bool                         `json:"dry_run"`
	Verbose    bool                         `json:"verbose"`

	// DualTheme support (optional, for plugins that support both light and dark)
	// If present, contains the alternate theme variant (light if primary is dark, vice versa)
	AlternateTheme *AlternateThemeData `json:"alternate_theme,omitempty"`
}

// AlternateThemeData contains the palette data for the alternate theme variant.
type AlternateThemeData struct {
	Colours    map[string]CategorisedColour `json:"colours"`
	AllColours []CategorisedColour          `json:"all_colours"`
	ThemeType  string                       `json:"theme_type"`
}

// CategorisedColour represents a color with metadata for RPC transfer.
type CategorisedColour struct {
	RGB        RGBColour `json:"rgb"`
	Hex        string    `json:"hex"`
	Role       string    `json:"role,omitempty"`
	Luminance  float64   `json:"luminance,omitempty"`
	IsLight    bool      `json:"is_light,omitempty"`
	Hue        float64   `json:"hue,omitempty"`
	Saturation float64   `json:"saturation,omitempty"`
	Index      int       `json:"index,omitempty"`
}

// RGBColour represents an RGB color.
type RGBColour struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}
