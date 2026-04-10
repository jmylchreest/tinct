// Package palettebuilder provides shared helpers for building colour palettes
// from map[string]string colour sources with optional role-mapping support.
// Used by remotecss and remotejson input plugins.
package palettebuilder

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/jmylchreest/tinct/internal/colour"
)

// BuildPalette converts a map of colour names → hex strings into a Palette.
// If mapping is non-empty, role hints are applied: each entry maps a source
// colour name to a colour.Role string. When verbose is true, missing source
// keys are reported as warnings.
func BuildPalette(colors, mapping map[string]string, verbose bool) (*colour.Palette, error) { //nolint:gocyclo
	if len(colors) == 0 {
		return nil, fmt.Errorf("no colors extracted")
	}

	paletteColors := make([]colour.RGB, 0, len(colors))
	var roleHints map[colour.Role]int

	// Add all colours to the palette, tracking name → index.
	colorNameToIndex := make(map[string]int)
	for name, hex := range colors {
		rgb, err := ParseHex(hex)
		if err != nil {
			if verbose {
				fmt.Printf("   Skipping invalid color '%s': %v\n", name, err)
			}
			continue
		}
		colorNameToIndex[name] = len(paletteColors)
		paletteColors = append(paletteColors, rgb)
	}

	// If no mapping, return a plain palette.
	if len(mapping) == 0 {
		colorColors := make([]color.Color, len(paletteColors))
		for i, rgb := range paletteColors {
			colorColors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
		}
		return colour.NewPalette(colorColors), nil
	}

	if verbose {
		fmt.Printf("→ Applying color mappings:\n")
	}

	roleHints = make(map[colour.Role]int)

	for sourceKey, targetRole := range mapping {
		index, ok := colorNameToIndex[sourceKey]
		if !ok {
			if verbose {
				fmt.Printf("   Warning: color '%s' not found in source\n", sourceKey)
			}
			continue
		}

		role, err := ParseColourRole(targetRole)
		if err != nil {
			return nil, fmt.Errorf("invalid role '%s': %w", targetRole, err)
		}

		roleHints[role] = index

		if verbose {
			hex := colors[sourceKey]
			fmt.Printf("   %s (%s) → %s\n", sourceKey, hex, targetRole)
		}
	}

	if len(paletteColors) == 0 {
		return nil, fmt.Errorf("no valid colors extracted")
	}

	colorColors := make([]color.Color, len(paletteColors))
	for i, rgb := range paletteColors {
		colorColors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
	}

	if len(roleHints) > 0 {
		return colour.NewPaletteWithRoleHints(colorColors, roleHints), nil
	}

	return colour.NewPalette(colorColors), nil
}

// ParseHex parses a hex colour string into an RGB struct.
// Supports formats: #RRGGBB, RRGGBB, #RGB, RGB.
func ParseHex(hex string) (colour.RGB, error) {
	hex = strings.TrimSpace(hex)
	hex = strings.TrimPrefix(hex, "#")

	// Expand shorthand format (RGB -> RRGGBB).
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}

	if len(hex) != 6 {
		return colour.RGB{}, fmt.Errorf("invalid hex color length: expected 6 characters, got %d", len(hex))
	}

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

// ParseColourRole parses a role name string into a Role constant.
// The name is normalised to lowercase with underscores and hyphens stripped.
func ParseColourRole(name string) (colour.Role, error) {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")

	roleMap := map[string]colour.Role{
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
	}

	role, ok := roleMap[name]
	if !ok {
		return "", fmt.Errorf("unknown colour role '%s'", name)
	}

	return role, nil
}
