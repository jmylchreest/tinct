// Package markdown provides an input plugin for loading themes from markdown theme files.
package markdown

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/shared/themeformat"
	"github.com/jmylchreest/tinct/internal/version"
)

// Plugin implements the input.Plugin interface for markdown theme loading.
type Plugin struct {
	path             string
	extractWallpaper bool
	wallpaperDir     string
	wallpaperPath    string // Stores the path to extracted/referenced wallpaper
}

// New creates a new markdown input plugin.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "markdown"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Load theme from tinct markdown theme file"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.path, "markdown.path", "", "Path to markdown theme file (required)")
	cmd.Flags().BoolVar(&p.extractWallpaper, "markdown.extract-wallpaper", true, "Extract embedded wallpaper to file")
	cmd.Flags().StringVar(&p.wallpaperDir, "markdown.wallpaper-dir", "", "Directory to extract wallpaper to (default: same as theme file)")
}

// Validate checks if the plugin has all required inputs configured.
func (p *Plugin) Validate() error {
	if p.path == "" {
		return fmt.Errorf("--markdown.path is required")
	}
	return nil
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		{Name: "markdown.path", Type: "string", Default: "", Description: "Path to markdown theme file", Required: true},
		{Name: "markdown.extract-wallpaper", Type: "bool", Default: "true", Description: "Extract embedded wallpaper to file", Required: false},
		{Name: "markdown.wallpaper-dir", Type: "string", Default: "", Description: "Directory to extract wallpaper to (default: same as theme file)", Required: false},
	}
}

// WallpaperPath returns the path to the wallpaper (extracted or referenced).
// Implements the input.WallpaperProvider interface.
func (p *Plugin) WallpaperPath() string {
	return p.wallpaperPath
}

// Generate creates a raw colour palette from a markdown theme file.
func (p *Plugin) Generate(_ context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	// Parse the markdown theme file
	theme, err := themeformat.ParseFile(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme file: %w", err)
	}

	// Extract colors from theme
	colors := make([]color.Color, 0, len(theme.Colors))
	roleHints := make(map[colour.Role]int)

	for _, c := range theme.Colors {
		rgb, err := parseHex(c.Hex)
		if err != nil {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: skipping invalid color %s: %v\n", c.Hex, err)
			}
			continue
		}

		// If color has a role, record it
		if c.Role != "" {
			role := colour.Role(c.Role)
			roleHints[role] = len(colors)
		}

		colors = append(colors, rgbToColor(rgb))
	}

	// Handle wallpaper
	if theme.Wallpaper != nil {
		if err := p.handleWallpaper(theme, opts.Verbose); err != nil {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to handle wallpaper: %v\n", err)
			}
		}
	}

	if len(colors) == 0 {
		return nil, fmt.Errorf("no valid colors found in theme file")
	}

	// Create palette with role hints if any were provided
	var palette *colour.Palette
	if len(roleHints) > 0 {
		palette = colour.NewPaletteWithRoleHints(colors, roleHints)
	} else {
		palette = colour.NewPalette(colors)
	}

	return palette, nil
}

// handleWallpaper extracts or locates the wallpaper from the theme.
func (p *Plugin) handleWallpaper(theme *themeformat.Theme, verbose bool) error {
	if theme.Wallpaper == nil {
		return nil
	}

	wp := theme.Wallpaper

	// If wallpaper is external reference
	if !wp.Embedded {
		if wp.Path != "" {
			// Resolve relative path against theme file location
			if !filepath.IsAbs(wp.Path) {
				p.wallpaperPath = filepath.Join(filepath.Dir(p.path), wp.Path)
			} else {
				p.wallpaperPath = wp.Path
			}

			// Check if file exists
			if _, err := os.Stat(p.wallpaperPath); err != nil {
				return fmt.Errorf("wallpaper file not found: %s", p.wallpaperPath)
			}
			return nil
		}
		if wp.URL != "" {
			// URL references are not automatically downloaded
			// The downstream consumer can handle the URL if needed
			if verbose {
				fmt.Fprintf(os.Stderr, "Note: wallpaper is a URL reference: %s\n", wp.URL)
			}
			return nil
		}
		return nil
	}

	// Embedded wallpaper - extract if requested
	if !p.extractWallpaper {
		return nil
	}

	// Decode wallpaper data
	data, format, err := themeformat.DecodeWallpaper(theme, p.path)
	if err != nil {
		return fmt.Errorf("failed to decode wallpaper: %w", err)
	}

	// Determine output directory
	outputDir := p.wallpaperDir
	if outputDir == "" {
		outputDir = filepath.Dir(p.path)
	}

	// Create output directory if needed
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return fmt.Errorf("failed to create wallpaper directory: %w", err)
	}

	// Generate filename from theme name
	baseName := strings.TrimSuffix(filepath.Base(p.path), filepath.Ext(p.path))
	if format == "" {
		format = "png" // Default format
	}
	wallpaperFilename := fmt.Sprintf("%s-wallpaper.%s", baseName, format)
	p.wallpaperPath = filepath.Join(outputDir, wallpaperFilename)

	// Write wallpaper file
	if err := os.WriteFile(p.wallpaperPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write wallpaper: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Extracted wallpaper to: %s\n", p.wallpaperPath)
	}

	return nil
}

// parseHex parses a hex colour string into an RGB struct.
// Supports formats: #RRGGBB, RRGGBB, #RGB, RGB.
func parseHex(hex string) (colour.RGB, error) {
	// Remove # prefix if present
	hex = strings.TrimPrefix(hex, "#")

	// Expand shorthand format (RGB -> RRGGBB)
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	// Validate length
	if len(hex) != 6 {
		return colour.RGB{}, fmt.Errorf("invalid hex colour length: expected 6 characters, got %d", len(hex))
	}

	// Parse hex values
	var r, g, b uint64
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return colour.RGB{}, fmt.Errorf("invalid hex colour: %w", err)
	}

	// Bounds check (hex pairs can only be 0-255, but validate explicitly)
	if r > 255 || g > 255 || b > 255 {
		return colour.RGB{}, fmt.Errorf("invalid hex colour values: r=%d, g=%d, b=%d", r, g, b)
	}

	return colour.RGB{
		R: uint8(r), // #nosec G115 -- bounds checked above
		G: uint8(g), // #nosec G115 -- bounds checked above
		B: uint8(b), // #nosec G115 -- bounds checked above
	}, nil
}

// rgbToColor converts an RGB struct to a color.Color interface.
func rgbToColor(rgb colour.RGB) color.Color {
	return color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}
