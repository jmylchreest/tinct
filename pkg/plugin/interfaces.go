// Package plugin provides the public API for tinct plugins.
package plugin

import (
	"context"
	"image/color"
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
