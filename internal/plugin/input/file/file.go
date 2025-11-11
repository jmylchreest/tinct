// Package file provides an input plugin for loading colour palettes from files or manual specifications.
package file

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
)

// Plugin implements the input.Plugin interface for file-based palette loading.
type Plugin struct {
	path            string
	colourOverrides []string
}

// New creates a new file input plugin.
func New() *Plugin {
	return &Plugin{
		colourOverrides: []string{},
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "file"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Load palette from file or build from colour specifications"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.path, "file.path", "", "Path to palette file (JSON or text, optional)")
	cmd.Flags().StringArrayVar(&p.colourOverrides, "colour", []string{}, "Colour override (role=hex, repeatable)")
}

// Validate checks if the plugin has all required inputs configured.
func (p *Plugin) Validate() error {
	// Either path or colour overrides must be provided.
	if p.path == "" && len(p.colourOverrides) == 0 {
		return fmt.Errorf("must provide either --file.path or --colour specifications")
	}

	// Validate colour overrides format.
	for _, override := range p.colourOverrides {
		if !strings.Contains(override, "=") {
			return fmt.Errorf("invalid colour format '%s': expected 'role=hex'", override)
		}
	}

	return nil
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		{Name: "file.path", Type: "string", Default: "", Description: "Path to palette file (JSON or text, optional)", Required: false},
		{Name: "colour", Type: "stringArray", Default: "[]", Description: "Colour override (role=hex, repeatable)", Required: false},
	}
}

// Generate creates a raw colour palette from file and/or manual specifications.
// Returns only the colors - categorization happens separately.
// If role-based colors are provided (role=hex), they are stored as metadata for later categorization.
func (p *Plugin) Generate(_ context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	var colors []color.Color
	roleHints := make(map[colour.Role]int) // Map roles to color indices

	// Merge colour overrides from options with plugin's own overrides.
	allOverrides := append([]string{}, p.colourOverrides...)
	allOverrides = append(allOverrides, opts.ColourOverrides...)

	// Load from file if provided.
	if p.path != "" {
		loadedColors, hints, err := p.loadFromFile(p.path)
		if err != nil {
			return nil, fmt.Errorf("failed to load palette file: %w", err)
		}
		colors = loadedColors
		roleHints = hints
	}

	// Apply colour overrides (from both plugin flags and options).
	if len(allOverrides) > 0 {
		overrideColors, overrideHints, err := p.parseOverrides(allOverrides)
		if err != nil {
			return nil, fmt.Errorf("failed to apply colour overrides: %w", err)
		}

		// Merge override colors and hints.
		for role, idx := range overrideHints {
			// Check if we need to add this color or replace an existing one.
			if existingIdx, exists := roleHints[role]; exists {
				// Replace existing color.
				colors[existingIdx] = overrideColors[idx]
			} else {
				// Add new color.
				roleHints[role] = len(colors)
				colors = append(colors, overrideColors[idx])
			}
		}
	}

	// If no colors provided at all, return error.
	if len(colors) == 0 {
		return nil, fmt.Errorf("no colors provided")
	}

	// Create palette with role hints if any were provided.
	var palette *colour.Palette
	if len(roleHints) > 0 {
		palette = colour.NewPaletteWithRoleHints(colors, roleHints)
	} else {
		palette = colour.NewPalette(colors)
	}

	return palette, nil
}

// loadFromFile loads colors from a JSON or text file.
// Returns colors and optional role hints.
func (p *Plugin) loadFromFile(path string) ([]color.Color, map[colour.Role]int, error) {
	data, err := os.ReadFile(path) // #nosec G304 - User-specified input file, intended to be read
	if err != nil {
		return nil, nil, err
	}

	// Try JSON first (categorised palette format).
	var categorised colour.CategorisedPalette
	if err := json.Unmarshal(data, &categorised); err == nil {
		// Extract colors and role hints from categorised palette.
		colors := make([]color.Color, 0)
		roleHints := make(map[colour.Role]int)

		for role, catColor := range categorised.Colours {
			roleHints[role] = len(colors)
			colors = append(colors, catColor.Colour)
		}

		// Also add any colors from AllColours that aren't in roles.
		for _, catColor := range categorised.AllColours {
			colors = append(colors, catColor.Colour)
		}

		return colors, roleHints, nil
	}

	// Try simple text format: hex colors or role=hex.
	return p.parseTextFormat(string(data))
}

// parseTextFormat parses a simple text format palette file.
// Format: hex colors (one per line) or role=hex (one per line), # for comments.
func (p *Plugin) parseTextFormat(content string) ([]color.Color, map[colour.Role]int, error) {
	colors := make([]color.Color, 0)
	roleHints := make(map[colour.Role]int)

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check if it's role=hex format or just hex.
		if !strings.Contains(line, "=") {
			// Just a hex color.
			hex := line
			rgb, err := parseHex(hex)
			if err != nil {
				return nil, nil, fmt.Errorf("line %d: invalid hex colour '%s': %w", lineNum+1, hex, err)
			}

			colors = append(colors, rgbToColor(rgb))
			continue
		}

		// Parse role=hex or colourN=hex.
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("line %d: invalid format, expected 'role=hex' or just 'hex'", lineNum+1)
		}

		roleName := strings.TrimSpace(parts[0])
		hex := strings.TrimSpace(parts[1])

		// Check if it's an indexed color (colourN or colorN).
		lowerRole := strings.ToLower(roleName)
		if strings.HasPrefix(lowerRole, "colour") || strings.HasPrefix(lowerRole, "color") {
			// Indexed color - just add the hex without role hint.
			rgb, err := parseHex(hex)
			if err != nil {
				return nil, nil, fmt.Errorf("line %d: invalid hex colour '%s': %w", lineNum+1, hex, err)
			}
			colors = append(colors, rgbToColor(rgb))
			continue
		}

		// Role-based color.
		role, err := parseColourRole(roleName)
		if err != nil {
			return nil, nil, fmt.Errorf("line %d: %w", lineNum+1, err)
		}

		rgb, err := parseHex(hex)
		if err != nil {
			return nil, nil, fmt.Errorf("line %d: invalid hex colour '%s': %w", lineNum+1, hex, err)
		}

		roleHints[role] = len(colors)
		colors = append(colors, rgbToColor(rgb))
	}

	return colors, roleHints, nil
}

// parseOverrides parses colour overrides from command line or options.
// Returns colors and role hints.
func (p *Plugin) parseOverrides(overrides []string) ([]color.Color, map[colour.Role]int, error) {
	colors := make([]color.Color, 0)
	roleHints := make(map[colour.Role]int)

	for _, override := range overrides {
		parts := strings.SplitN(override, "=", 2)
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid override format '%s': expected 'role=hex'", override)
		}

		roleName := strings.TrimSpace(parts[0])
		hex := strings.TrimSpace(parts[1])

		// Parse role name.
		role, err := parseColourRole(roleName)
		if err != nil {
			return nil, nil, err
		}

		// Parse hex colour.
		rgb, err := parseHex(hex)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid hex colour '%s': %w", hex, err)
		}

		roleHints[role] = len(colors)
		colors = append(colors, rgbToColor(rgb))
	}

	return colors, roleHints, nil
}

// parseColourRole parses a role name string into a Role constant.
// Accepts both British English (colour) and American English (color) spelling.
func parseColourRole(name string) (colour.Role, error) {
	// Normalise the name - convert to lowercase and replace underscores/hyphens.
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")

	roleMap := map[string]colour.Role{
		// Core semantic roles.
		"background":      colour.RoleBackground,
		"backgroundmuted": colour.RoleBackgroundMuted,
		"foreground":      colour.RoleForeground,
		"foregroundmuted": colour.RoleForegroundMuted,
		"accent1":         colour.RoleAccent1,
		"accent1muted":    colour.RoleAccent1Muted,
		"accent2":         colour.RoleAccent2,
		"accent2muted":    colour.RoleAccent2Muted,
		"accent3":         colour.RoleAccent3,
		"accent3muted":    colour.RoleAccent3Muted,
		"accent4":         colour.RoleAccent4,
		"accent4muted":    colour.RoleAccent4Muted,
		"danger":          colour.RoleDanger,
		"warning":         colour.RoleWarning,
		"success":         colour.RoleSuccess,
		"info":            colour.RoleInfo,
		"notification":    colour.RoleNotification,

		// Position hints (edge/corner regions for ambient lighting).
		"positiontopleft":     colour.RolePositionTopLeft,
		"positiontop":         colour.RolePositionTop,
		"positiontopright":    colour.RolePositionTopRight,
		"positionright":       colour.RolePositionRight,
		"positionbottomright": colour.RolePositionBottomRight,
		"positionbottom":      colour.RolePositionBottom,
		"positionbottomleft":  colour.RolePositionBottomLeft,
		"positionleft":        colour.RolePositionLeft,

		// Position hints (12-region grid).
		"positiontopleftinner":     colour.RolePositionTopLeftInner,
		"positiontopcenter":        colour.RolePositionTopCenter,
		"positiontoprightinner":    colour.RolePositionTopRightInner,
		"positionrighttop":         colour.RolePositionRightTop,
		"positionrightbottom":      colour.RolePositionRightBottom,
		"positionbottomrightinner": colour.RolePositionBottomRightInner,
		"positionbottomcenter":     colour.RolePositionBottomCenter,
		"positionbottomleftinner":  colour.RolePositionBottomLeftInner,
		"positionleftbottom":       colour.RolePositionLeftBottom,
		"positionlefttop":          colour.RolePositionLeftTop,

		// Position hints (16-region grid).
		"positiontopleftouter":      colour.RolePositionTopLeftOuter,
		"positiontopleftcenter":     colour.RolePositionTopLeftCenter,
		"positiontoprightcenter":    colour.RolePositionTopRightCenter,
		"positiontoprightouter":     colour.RolePositionTopRightOuter,
		"positionrighttopoputer":    colour.RolePositionRightTopOuter,
		"positionrightbottomouter":  colour.RolePositionRightBottomOuter,
		"positionbottomrightouter":  colour.RolePositionBottomRightOuter,
		"positionbottomrightcenter": colour.RolePositionBottomRightCenter,
		"positionbottomleftcenter":  colour.RolePositionBottomLeftCenter,
		"positionbottomleftouter":   colour.RolePositionBottomLeftOuter,
		"positionleftbottomouter":   colour.RolePositionLeftBottomOuter,
		"positionlefttopoputer":     colour.RolePositionLeftTopOuter,
	}

	role, ok := roleMap[name]
	if !ok {
		return "", fmt.Errorf("unknown colour role '%s'", name)
	}

	return role, nil
}

// parseHex parses a hex colour string into an RGB struct.
// Supports formats: #RRGGBB, RRGGBB, #RGB, RGB.
func parseHex(hex string) (colour.RGB, error) {
	// Remove # prefix if present.
	hex = strings.TrimPrefix(hex, "#")

	// Expand shorthand format (RGB -> RRGGBB).
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	// Validate length.
	if len(hex) != 6 {
		return colour.RGB{}, fmt.Errorf("invalid hex colour length: expected 6 characters, got %d", len(hex))
	}

	// Parse hex values.
	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return colour.RGB{}, fmt.Errorf("invalid red component: %w", err)
	}

	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return colour.RGB{}, fmt.Errorf("invalid green component: %w", err)
	}

	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return colour.RGB{}, fmt.Errorf("invalid blue component: %w", err)
	}

	return colour.RGB{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}, nil
}

// rgbToColor converts an RGB struct to a color.Color interface.
func rgbToColor(rgb colour.RGB) color.Color {
	return color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}
