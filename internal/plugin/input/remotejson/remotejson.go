// Package remotejson provides an input plugin for fetching colour palettes from remote JSON sources with JSONPath queries.
package remotejson

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	httputil "github.com/jmylchreest/tinct/internal/util/http"
	"github.com/spf13/cobra"
)

// Plugin implements the input.Plugin interface for remote JSON palette fetching with JSONPath queries.
type Plugin struct {
	url     string
	query   string // JSONPath query to extract colors (optional)
	timeout time.Duration
	mapping map[string]string // Maps source color names to tinct roles
}

// New creates a new remote-json input plugin.
func New() *Plugin {
	return &Plugin{
		timeout: 10 * time.Second,
		mapping: make(map[string]string),
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "remote-json"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Fetch colour palette from remote JSON source with optional JSONPath queries"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.url, "remote-json.url", "", "URL to fetch JSON palette from (required)")
	cmd.Flags().StringVar(&p.query, "remote-json.query", "", "JSONPath query to extract colors (e.g. '$.colors', optional)")
	cmd.Flags().DurationVar(&p.timeout, "remote-json.timeout", 10*time.Second, "HTTP timeout")
	cmd.Flags().StringToStringVar(&p.mapping, "remote-json.map", map[string]string{}, "Map colors to roles (e.g. base=background,text=foreground)")
}

// Validate checks if the plugin has all required inputs configured.
func (p *Plugin) Validate() error {
	if p.url == "" {
		return fmt.Errorf("--remote-json.url is required")
	}

	// Basic URL validation
	if !strings.HasPrefix(p.url, "http://") && !strings.HasPrefix(p.url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	return nil
}

// Generate fetches and parses a remote JSON colour palette.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	if opts.Verbose {
		fmt.Printf("→ Fetching JSON palette from: %s\n", p.url)
	}

	// Fetch content
	content, err := p.fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch palette: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("   Size: %d bytes\n", len(content))
	}

	// Parse JSON
	colors, err := p.parseJSON(content, opts.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("   Extracted %d colors\n", len(colors))
	}

	// Convert to palette
	palette, err := p.buildPalette(colors, opts.Verbose)
	if err != nil {
		return nil, err
	}

	return palette, nil
}

// fetch retrieves content from the remote URL.
func (p *Plugin) fetch(ctx context.Context) ([]byte, error) {
	return httputil.Fetch(ctx, p.url, httputil.FetchOptions{
		Timeout: p.timeout,
	})
}

// parseJSON parses JSON content and extracts colors.
func (p *Plugin) parseJSON(content []byte, verbose bool) (map[string]string, error) {
	var data any
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	// If query is provided, navigate to that path
	if p.query != "" {
		if verbose {
			fmt.Printf("   Applying query: %s\n", p.query)
		}
		var err error
		data, err = p.applyQuery(data, p.query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
	}

	// Extract colors from the data
	colors := make(map[string]string)
	p.extractColors(data, "", colors)

	return colors, nil
}

// applyQuery applies a simplified JSONPath-like query to the data.
// Supports: $.path.to.field or just path.to.field
func (p *Plugin) applyQuery(data any, query string) (any, error) {
	// Remove leading $. if present
	query = strings.TrimPrefix(query, "$.")
	query = strings.TrimPrefix(query, "$")

	if query == "" {
		return data, nil
	}

	// Split path into segments
	segments := strings.Split(query, ".")

	current := data
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]any:
			val, ok := v[segment]
			if !ok {
				return nil, fmt.Errorf("path not found: %s", segment)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot navigate into %T at segment '%s'", current, segment)
		}
	}

	return current, nil
}

// extractColors recursively extracts color values from JSON data.
func (p *Plugin) extractColors(data any, prefix string, colors map[string]string) {
	switch v := data.(type) {
	case map[string]any:
		// Check if this object looks like a color object (has hex/rgb properties)
		if hexVal, hasHex := v["hex"].(string); hasHex && isColor(hexVal) {
			// This is a color object (e.g., Catppuccin format: {"hex": "#1e1e2e", "rgb": {...}})
			// Use the prefix as the key
			if prefix != "" {
				key := prefix
				if idx := strings.LastIndex(prefix, "."); idx >= 0 {
					key = prefix[idx+1:]
				}
				colors[key] = hexVal
			}
			return
		}

		// Not a color object, traverse its properties
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			// Check if this is a color value (string that looks like a hex color)
			if strVal, ok := value.(string); ok {
				if isColor(strVal) {
					colors[key] = strVal
				} else {
					// Might be nested, recurse
					p.extractColors(value, fullKey, colors)
				}
			} else {
				// Recurse into nested structures
				p.extractColors(value, fullKey, colors)
			}
		}
	case []any:
		for i, item := range v {
			p.extractColors(item, fmt.Sprintf("%s[%d]", prefix, i), colors)
		}
	case string:
		if isColor(v) && prefix != "" {
			// Use the last segment of the path as the key
			key := prefix
			if idx := strings.LastIndex(prefix, "."); idx >= 0 {
				key = prefix[idx+1:]
			}
			colors[key] = v
		}
	}
}

// isColor checks if a string looks like a color value.
func isColor(s string) bool {
	s = strings.TrimSpace(s)

	// Hex color: #RGB or #RRGGBB
	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		return len(hex) == 3 || len(hex) == 6
	}

	// RGB/RGBA format: rgb(r, g, b) or rgba(r, g, b, a)
	if strings.HasPrefix(s, "rgb(") || strings.HasPrefix(s, "rgba(") {
		return true
	}

	// HSL/HSLA format: hsl(h, s, l) or hsla(h, s, l, a)
	if strings.HasPrefix(s, "hsl(") || strings.HasPrefix(s, "hsla(") {
		return true
	}

	return false
}

// buildPalette converts extracted colors to a Palette.
func (p *Plugin) buildPalette(colors map[string]string, verbose bool) (*colour.Palette, error) {
	if len(colors) == 0 {
		return nil, fmt.Errorf("no colors extracted")
	}

	var paletteColors []colour.RGB
	var roleHints map[colour.ColourRole]int

	// First, add ALL colors to the palette
	colorNameToIndex := make(map[string]int)
	for name, hex := range colors {
		rgb, err := parseHex(hex)
		if err != nil {
			if verbose {
				fmt.Printf("   Skipping invalid color '%s': %v\n", name, err)
			}
			continue
		}
		colorNameToIndex[name] = len(paletteColors)
		paletteColors = append(paletteColors, rgb)
	}

	// Then, if mapping is provided, create role hints for the mapped colors
	if len(p.mapping) > 0 {
		if verbose {
			fmt.Printf("→ Applying color mappings:\n")
		}

		roleHints = make(map[colour.ColourRole]int)

		for sourceKey, targetRole := range p.mapping {
			if index, ok := colorNameToIndex[sourceKey]; ok {
				// Parse the target role
				role, err := parseColourRole(targetRole)
				if err != nil {
					return nil, fmt.Errorf("invalid role '%s': %w", targetRole, err)
				}

				roleHints[role] = index

				if verbose {
					hex := colors[sourceKey]
					fmt.Printf("   %s (%s) → %s\n", sourceKey, hex, targetRole)
				}
			} else {
				if verbose {
					fmt.Printf("   Warning: color '%s' not found in source\n", sourceKey)
				}
			}
		}
	}

	if len(paletteColors) == 0 {
		return nil, fmt.Errorf("no valid colors extracted")
	}

	// Convert RGB to color.Color
	colorColors := make([]color.Color, len(paletteColors))
	for i, rgb := range paletteColors {
		colorColors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
	}

	// Create palette with role hints if mapping was used
	if len(roleHints) > 0 {
		return colour.NewPaletteWithRoleHints(colorColors, roleHints), nil
	}

	return colour.NewPalette(colorColors), nil
}

// parseHex parses a hex color string into an RGB struct.
// Supports formats: #RRGGBB, RRGGBB, #RGB, RGB
func parseHex(hex string) (colour.RGB, error) {
	hex = strings.TrimSpace(hex)
	hex = strings.TrimPrefix(hex, "#")

	// Expand shorthand format (RGB -> RRGGBB)
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

// parseColourRole parses a role name string into a ColourRole constant.
func parseColourRole(name string) (colour.ColourRole, error) {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, "-", "")

	roleMap := map[string]colour.ColourRole{
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
