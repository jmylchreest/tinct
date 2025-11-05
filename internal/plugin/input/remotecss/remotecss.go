// Package remotecss provides an input plugin for fetching colour palettes from remote CSS sources.
package remotecss

import (
	"context"
	"fmt"
	"image/color"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	httputil "github.com/jmylchreest/tinct/internal/util/http"
)

// Plugin implements the input.Plugin interface for remote CSS palette fetching.
type Plugin struct {
	url     string
	timeout time.Duration
	mapping map[string]string // Maps source color names to tinct roles
}

// New creates a new remote-css input plugin.
func New() *Plugin {
	return &Plugin{
		timeout: 10 * time.Second,
		mapping: make(map[string]string),
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "remote-css"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Fetch colour palette from remote CSS source (extracts CSS variables and color values)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.url, "remote-css.url", "", "URL to fetch CSS palette from (required)")
	cmd.Flags().DurationVar(&p.timeout, "remote-css.timeout", 10*time.Second, "HTTP timeout")
	cmd.Flags().StringToStringVar(&p.mapping, "remote-css.map", map[string]string{}, "Map colors to roles (e.g. primary=background,secondary=foreground)")
}

// Validate checks if the plugin has all required inputs configured.
func (p *Plugin) Validate() error {
	if p.url == "" {
		return fmt.Errorf("--remote-css.url is required")
	}

	// Basic URL validation.
	if !strings.HasPrefix(p.url, "http://") && !strings.HasPrefix(p.url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	return nil
}

// Generate fetches and parses a remote CSS colour palette.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	if opts.Verbose {
		fmt.Printf("→ Fetching CSS palette from: %s\n", p.url)
	}

	// Fetch content.
	content, err := p.fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch palette: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("   Size: %d bytes\n", len(content))
	}

	// Parse CSS.
	colors, err := p.parseCSS(string(content), opts.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSS: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("   Extracted %d colors\n", len(colors))
	}

	// Convert to palette.
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

// parseCSS extracts color values from CSS content using regex patterns.
// Supports: CSS custom properties, color properties, hex, rgb, hsl, oklch, oklab.
func (p *Plugin) parseCSS(content string, verbose bool) (map[string]string, error) {
	colors := make(map[string]string)

	// Extract CSS custom properties (--variable-name: value).
	cssVarRegex := regexp.MustCompile(`--([a-zA-Z0-9_-]+)\s*:\s*([^;]+);`)
	matches := cssVarRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) == 3 {
			varName := match[1]
			value := strings.TrimSpace(match[2])

			if hexColor := extractColor(value); hexColor != "" {
				colors[varName] = hexColor
			}
		}
	}

	// Extract color properties (color: value, background-color: value, etc.).
	colorPropRegex := regexp.MustCompile(`(?:color|background-color|background|border-color|fill|stroke)\s*:\s*([^;]+);`)
	propMatches := colorPropRegex.FindAllStringSubmatch(content, -1)
	for _, match := range propMatches {
		if len(match) == 2 {
			value := strings.TrimSpace(match[1])

			if hexColor := extractColor(value); hexColor != "" {
				// Only add if not already in colors (avoid duplicates).
				found := false
				for _, existing := range colors {
					if existing == hexColor {
						found = true
						break
					}
				}
				if !found {
					colors[fmt.Sprintf("color-%s", hexColor[1:])] = hexColor
				}
			}
		}
	}

	if len(colors) == 0 {
		return nil, fmt.Errorf("no colors found in CSS")
	}

	return colors, nil
}

// extractColor extracts and converts a color value to hex format.
// Supports: hex, rgb, rgba, hsl, hsla, oklch, oklab.
func extractColor(value string) string {
	value = strings.TrimSpace(value)

	// Hex color.
	if hexColor := extractHexColor(value); hexColor != "" {
		return hexColor
	}

	// RGB/RGBA.
	if rgbColor := convertRGBToHex(value); rgbColor != "" {
		return rgbColor
	}

	// HSL/HSLA.
	if hslColor := convertHSLToHex(value); hslColor != "" {
		return hslColor
	}

	// OKLCH.
	if oklchColor := convertOKLCHToHex(value); oklchColor != "" {
		return oklchColor
	}

	// OKLAB.
	if oklabColor := convertOKLABToHex(value); oklabColor != "" {
		return oklabColor
	}

	return ""
}

// extractHexColor extracts hex color from a value string.
func extractHexColor(value string) string {
	hexRegex := regexp.MustCompile(`#([0-9a-fA-F]{6}|[0-9a-fA-F]{3})\b`)
	if match := hexRegex.FindString(value); match != "" {
		return match
	}
	return ""
}

// convertRGBToHex extracts rgb/rgba color and converts to hex.
func convertRGBToHex(value string) string {
	rgbRegex := regexp.MustCompile(`rgba?\s*\(\s*([0-9.]+)\s*,?\s*([0-9.]+)\s*,?\s*([0-9.]+)`)
	matches := rgbRegex.FindStringSubmatch(value)
	if len(matches) == 4 {
		// Regex guarantees these are valid floats, errors ignored
		r, _ := strconv.ParseFloat(matches[1], 64) //nolint:errcheck
		g, _ := strconv.ParseFloat(matches[2], 64) //nolint:errcheck
		b, _ := strconv.ParseFloat(matches[3], 64) //nolint:errcheck
		return fmt.Sprintf("#%02x%02x%02x",
			clamp(int(r), 0, 255),
			clamp(int(g), 0, 255),
			clamp(int(b), 0, 255))
	}
	return ""
}

// convertHSLToHex extracts hsl/hsla color and converts to hex.
func convertHSLToHex(value string) string {
	hslRegex := regexp.MustCompile(`hsla?\s*\(\s*([0-9.]+)\s*,?\s*([0-9.]+)%?\s*,?\s*([0-9.]+)%?`)
	matches := hslRegex.FindStringSubmatch(value)
	if len(matches) == 4 {
		// Regex guarantees these are valid floats, errors ignored
		h, _ := strconv.ParseFloat(matches[1], 64) //nolint:errcheck
		s, _ := strconv.ParseFloat(matches[2], 64) //nolint:errcheck
		l, _ := strconv.ParseFloat(matches[3], 64) //nolint:errcheck

		// Handle percentage values.
		if s > 1 {
			s /= 100.0
		}
		if l > 1 {
			l /= 100.0
		}

		rgb := hslToRGB(h, s, l)
		return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
	}
	return ""
}

// convertOKLCHToHex extracts oklch color and converts to hex.
// Format: oklch(L C H) where L is 0-1, C is 0-0.4, H is 0-360.
func convertOKLCHToHex(value string) string {
	oklchRegex := regexp.MustCompile(`oklch\s*\(\s*([0-9.]+)\s+([0-9.]+)\s+([0-9.]+)`)
	matches := oklchRegex.FindStringSubmatch(value)
	if len(matches) == 4 {
		// Regex guarantees these are valid floats, errors ignored
		l, _ := strconv.ParseFloat(matches[1], 64) //nolint:errcheck
		c, _ := strconv.ParseFloat(matches[2], 64) //nolint:errcheck
		h, _ := strconv.ParseFloat(matches[3], 64) //nolint:errcheck

		rgb := oklchToRGB(l, c, h)
		return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
	}
	return ""
}

// convertOKLABToHex extracts oklab color and converts to hex.
// Format: oklab(L a b) where L is 0-1, a and b are typically -0.4 to 0.4.
func convertOKLABToHex(value string) string {
	oklabRegex := regexp.MustCompile(`oklab\s*\(\s*([0-9.-]+)\s+([0-9.-]+)\s+([0-9.-]+)`)
	matches := oklabRegex.FindStringSubmatch(value)
	if len(matches) == 4 {
		// Regex guarantees these are valid floats, errors ignored
		l, _ := strconv.ParseFloat(matches[1], 64) //nolint:errcheck
		a, _ := strconv.ParseFloat(matches[2], 64) //nolint:errcheck
		b, _ := strconv.ParseFloat(matches[3], 64) //nolint:errcheck

		rgb := oklabToRGB(l, a, b)
		return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
	}
	return ""
}

// clamp restricts a value to a given range.
func clamp(val, minVal, maxVal int) int {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

// hslToRGB converts HSL to RGB.
func hslToRGB(h, s, l float64) colour.RGB {
	h /= 360.0

	var r, g, b float64

	if s == 0 {
		r = l
		g = l
		b = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return colour.RGB{
		R: uint8(clamp(int(r*255), 0, 255)), // #nosec G115 -- clamped to 0-255
		G: uint8(clamp(int(g*255), 0, 255)), // #nosec G115 -- clamped to 0-255
		B: uint8(clamp(int(b*255), 0, 255)), // #nosec G115 -- clamped to 0-255
	}
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// oklchToRGB converts OKLCH to RGB.
// OKLCH: Lightness (0-1), Chroma (0-0.4), Hue (0-360).
func oklchToRGB(l, c, h float64) colour.RGB {
	// Convert OKLCH to OKLAB.
	hRad := h * math.Pi / 180.0
	a := c * math.Cos(hRad)
	b := c * math.Sin(hRad)

	// Convert OKLAB to RGB.
	return oklabToRGB(l, a, b)
}

// oklabToRGB converts OKLAB to RGB.
// OKLAB: Lightness (0-1), a (-0.4 to 0.4), b (-0.4 to 0.4).
// Reference: https://bottosson.github.io/posts/oklab/.
func oklabToRGB(l, a, b float64) colour.RGB {
	// OKLAB to linear RGB (D65 illuminant).
	lVal := l + 0.3963377774*a + 0.2158037573*b
	mVal := l - 0.1055613458*a - 0.0638541728*b
	sVal := l - 0.0894841775*a - 1.2914855480*b

	lVal = lVal * lVal * lVal
	mVal = mVal * mVal * mVal
	sVal = sVal * sVal * sVal

	r := +4.0767416621*lVal - 3.3077115913*mVal + 0.2309699292*sVal
	g := -1.2684380046*lVal + 2.6097574011*mVal - 0.3413193965*sVal
	bVal := -0.0041960863*lVal - 0.7034186147*mVal + 1.7076147010*sVal

	// Convert linear RGB to sRGB (gamma correction).
	r = linearToSRGB(r)
	g = linearToSRGB(g)
	bVal = linearToSRGB(bVal)

	return colour.RGB{
		R: uint8(clamp(int(r*255+0.5), 0, 255)),    // #nosec G115 -- clamped to 0-255
		G: uint8(clamp(int(g*255+0.5), 0, 255)),    // #nosec G115 -- clamped to 0-255
		B: uint8(clamp(int(bVal*255+0.5), 0, 255)), // #nosec G115 -- clamped to 0-255
	}
}

// linearToSRGB converts linear RGB to sRGB (gamma correction).
func linearToSRGB(c float64) float64 {
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

// buildPalette converts extracted colors to a Palette.
func (p *Plugin) buildPalette(colors map[string]string, verbose bool) (*colour.Palette, error) {
	if len(colors) == 0 {
		return nil, fmt.Errorf("no colors extracted")
	}

	var paletteColors []colour.RGB
	var roleHints map[colour.ColourRole]int

	// First, add ALL colors to the palette.
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

	// Then, if mapping is provided, create role hints for the mapped colors.
	if len(p.mapping) > 0 {
		if verbose {
			fmt.Printf("→ Applying color mappings:\n")
		}

		roleHints = make(map[colour.ColourRole]int)

		for sourceKey, targetRole := range p.mapping {
			if index, ok := colorNameToIndex[sourceKey]; ok {
				// Parse the target role.
				role, err := parseColourRole(targetRole)
				if err != nil {
					return nil, fmt.Errorf("invalid role '%s': %w", targetRole, err)
				}

				roleHints[role] = index

				if verbose {
					hex := colors[sourceKey]
					fmt.Printf("   %s (%s) → %s\n", sourceKey, hex, targetRole)
				}
			} else if verbose {
				fmt.Printf("   Warning: color '%s' not found in source\n", sourceKey)
			}
		}
	}

	if len(paletteColors) == 0 {
		return nil, fmt.Errorf("no valid colors extracted")
	}

	// Convert RGB to color.Color.
	colorColors := make([]color.Color, len(paletteColors))
	for i, rgb := range paletteColors {
		colorColors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
	}

	// Create palette with role hints if mapping was used.
	if len(roleHints) > 0 {
		return colour.NewPaletteWithRoleHints(colorColors, roleHints), nil
	}

	return colour.NewPalette(colorColors), nil
}

// parseHex parses a hex color string into an RGB struct.
// Supports formats: #RRGGBB, RRGGBB, #RGB, RGB.
func parseHex(hex string) (colour.RGB, error) {
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
