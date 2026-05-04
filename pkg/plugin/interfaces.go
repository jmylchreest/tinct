// Package plugin provides the public API for tinct plugins.
package plugin

import (
	"context"
	"image/color"

	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
)

// InputPlugin is the interface that input plugins must implement for go-plugin RPC.
type InputPlugin interface {
	// Generate creates a palette from plugin-specific inputs.
	Generate(ctx context.Context, opts InputOptions) ([]color.Color, error)

	// GetMetadata returns plugin metadata.
	GetMetadata() PluginInfo

	// WallpaperPath returns the canonical path to a wallpaper image, if available.
	// This path is resolved to be usable from any working directory:
	// - Relative paths are converted to absolute paths
	// - Tilde-prefixed paths (~/...) are preserved for portability
	// - URLs are returned as-is
	// Returns empty string if no wallpaper is available.
	WallpaperPath() string

	// WallpaperRawPath returns the literal path as provided by the user.
	// This is the unmodified input before any path canonicalization.
	// Returns empty string if no wallpaper is available.
	WallpaperRawPath() string

	// GetFlagHelp returns help information for plugin flags.
	GetFlagHelp() []FlagHelp
}

// Validator is an optional interface input or output plugins may
// implement to fail-fast before any expensive work runs.
//
// The host calls Validate(args) with the plugin's persistent arguments
// (set via `tinct plugins config <name> args=…`) and aborts the run if
// it returns a non-nil error. Use it for cheap, deterministic checks
// against the supplied args — missing required keys, malformed values,
// unsupported combinations — not for network calls or filesystem-state
// probing, which belong in PreExecute or Generate.
//
// Per-invocation args supplied via `--plugin-args` are NOT included in
// the map passed to Validate, because they're applied at Generate time
// after Validate has run. Plugins that need to validate per-run args
// should do so at the top of Generate.
//
// Plugins that don't implement Validator are treated as always-valid;
// older plugins built against pre-0.3.0 SDKs continue to work unchanged.
type Validator interface {
	Validate(args map[string]any) error
}

// RoleHinter is an optional interface input plugins may implement to declare
// explicit role assignments alongside the colours returned by Generate.
//
// Each entry maps a role name (matching the canonical Tinct role strings —
// "background", "foreground", "danger", "success", "warning", "info",
// "notification", "accent1".."accent4", "*muted" variants) to the index of
// the colour in the slice returned by Generate.
//
// RoleHints is queried by the host immediately after a successful Generate
// call. Implementations may return nil or an empty map to indicate no hints.
type RoleHinter interface {
	RoleHints() map[string]int
}

// ThemeHinter is an optional interface input plugins may implement to advise
// the host categoriser about the intended theme variant.
//
// Returns "dark", "light", "auto", or "" (no hint). The host treats the value
// as advisory — the categoriser may still override based on colour analysis.
type ThemeHinter interface {
	ThemeHint() string
}

// TemplateLister is an optional interface output plugins may implement
// to expose their bundled templates so `tinct plugins templates list`
// and `tinct plugins templates dump <plugin>` work for external
// plugins. Without it those commands skip the plugin (the legacy
// embed.FS introspection only worked for in-tree plugins).
//
// Templates returns a map from template name → contents. Names should
// be relative paths within the plugin's template tree (e.g.
// "theme.conf.tmpl", "templates/0.53/theme.conf.tmpl") so the host can
// reproduce the same structure under
// ~/.config/tinct/templates/<plugin>/ when the user dumps for editing.
//
// Implementations typically read directly from their embed.FS:
//
//	//go:embed templates/*.tmpl
//	var templatesFS embed.FS
//
//	func (p *Plugin) Templates() map[string][]byte {
//	    out := map[string][]byte{}
//	    fs.WalkDir(templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
//	        if err != nil || d.IsDir() { return err }
//	        b, _ := templatesFS.ReadFile(path)
//	        out[path] = b
//	        return nil
//	    })
//	    return out
//	}
type TemplateLister interface {
	Templates() map[string][]byte
}

// HooksProvider is an optional interface output plugins may implement to
// declare routine pre/post-execute behaviour (binary checks, dir
// auto-creation, reload commands, instructions text). The host runs the
// returned spec around the plugin's optional imperative PreExecute /
// PostExecute methods, so authors can move static behaviour off the
// imperative path and let the shared hooks runner handle it.
//
// Function fields on hooks.Spec (ReloadFn, InstructionsFn, Wallpaper)
// cannot be transmitted across the RPC boundary and are silently
// dropped in the marshalled payload — keep using PostExecute when you
// need dynamic argument templating or live state probing.
type HooksProvider interface {
	Hooks() hooks.Spec
}

// OutputPlugin is the interface that output plugins must implement for go-plugin RPC.
type OutputPlugin interface {
	// Generate creates output file(s) from the given theme data.
	Generate(ctx context.Context, palette PaletteData) (map[string][]byte, error)

	// PreExecute runs before Generate() for validation checks.
	PreExecute(ctx context.Context) (skip bool, reason string, err error)

	// PostExecute runs after successful Generate() and file writing.
	PostExecute(ctx context.Context, writtenFiles []string) error

	// GetMetadata returns plugin metadata.
	GetMetadata() PluginInfo

	// GetFlagHelp returns help information for plugin flags.
	GetFlagHelp() []FlagHelp
}
