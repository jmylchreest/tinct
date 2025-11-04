// Package color provides color extraction and palette generation functionality.
package colour

import (
	"encoding/json"
	"fmt"
	"image/color"

	"github.com/jmylchreest/tinct/internal/security"
)

// Palette represents a collection of colors extracted from an image.
type Palette struct {
	Colors    []color.Color
	Weights   []float64          // Optional: relative frequency/volume of each color (0.0-1.0)
	RoleHints map[ColourRole]int // Optional: explicit role assignments (role -> color index)
}

// NewPalette creates a new Palette with the given colors.
func NewPalette(colors []color.Color) *Palette {
	return &Palette{
		Colors:    colors,
		Weights:   nil, // No weights by default
		RoleHints: nil, // No role hints by default
	}
}

// NewPaletteWithWeights creates a new Palette with colors and their relative weights.
// Weights represent the frequency/volume of each color (e.g., from k-means cluster sizes).
// Weights should be normalised (sum to 1.0) but will be normalised if not.
func NewPaletteWithWeights(colors []color.Color, weights []float64) *Palette {
	if len(weights) != len(colors) {
		// Fallback to equal weights if mismatch.
		return NewPalette(colors)
	}

	// Normalise weights to sum to 1.0.
	sum := 0.0
	for _, w := range weights {
		sum += w
	}

	normalized := make([]float64, len(weights))
	if sum > 0 {
		for i, w := range weights {
			normalized[i] = w / sum
		}
	} else {
		// Equal weights if all zero.
		equalWeight := 1.0 / float64(len(weights))
		for i := range normalized {
			normalized[i] = equalWeight
		}
	}

	return &Palette{
		Colors:    colors,
		Weights:   normalized,
		RoleHints: nil, // No role hints by default
	}
}

// NewPaletteWithRoleHints creates a new Palette with colors and explicit role assignments.
// RoleHints maps semantic roles to color indices, allowing input plugins to override.
// automatic categorization for specific roles.
func NewPaletteWithRoleHints(colors []color.Color, roleHints map[ColourRole]int) *Palette {
	return &Palette{
		Colors:    colors,
		Weights:   nil,
		RoleHints: roleHints,
	}
}

// Len returns the number of colors in the palette.
func (p *Palette) Len() int {
	return len(p.Colors)
}

// RGB represents a color in RGB format (without alpha).
type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// String returns the RGB color as a string in the format "rgb(r, g, b)".
func (rgb RGB) String() string {
	return fmt.Sprintf("rgb(%d, %d, %d)", rgb.R, rgb.G, rgb.B)
}

// Hex returns the RGB color as a hex string (e.g., "#1a2b3c").
func (rgb RGB) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
}

// RGBA represents a color with alpha channel.
type RGBA struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
	A uint8 `json:"a"` // 0-255 (0 = transparent, 255 = opaque)
}

// AlphaFloat returns the alpha channel as a float in the range 0.0-1.0.
func (rgba RGBA) AlphaFloat() float64 {
	return float64(rgba.A) / 255.0
}

// String returns the RGBA color as a string in the format "rgba(r, g, b, a)".
func (rgba RGBA) String() string {
	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", rgba.R, rgba.G, rgba.B, rgba.AlphaFloat())
}

// Hex returns the RGB color as a hex string without alpha (e.g., "#1a2b3c").
// This maintains backwards compatibility with RGB.Hex().
func (rgba RGBA) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", rgba.R, rgba.G, rgba.B)
}

// HexAlpha returns the RGBA color as a hex string with alpha (e.g., "#1a2b3cff").
// This format is used by some applications like Dunst.
func (rgba RGBA) HexAlpha() string {
	return fmt.Sprintf("#%02x%02x%02x%02x", rgba.R, rgba.G, rgba.B, rgba.A)
}

// CSSRgb returns the color in CSS rgb() format (e.g., "rgb(26, 43, 60)").
func (rgba RGBA) CSSRgb() string {
	return fmt.Sprintf("rgb(%d, %d, %d)", rgba.R, rgba.G, rgba.B)
}

// CSSRgba returns the color in CSS rgba() format (e.g., "rgba(26, 43, 60, 1.00)").
// This format is used by CSS and applications like SwayOSD.
func (rgba RGBA) CSSRgba() string {
	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", rgba.R, rgba.G, rgba.B, rgba.AlphaFloat())
}

// ToRGB converts RGBA to RGB by dropping the alpha channel.
func (rgba RGBA) ToRGB() RGB {
	return RGB{R: rgba.R, G: rgba.G, B: rgba.B}
}

// ToRGB converts a color.Color to RGB.
func ToRGB(c color.Color) RGB {
	r, g, b, _ := c.RGBA()
	// RGBA returns values in the range [0, 65535], convert to [0, 255].
	return RGB{
		R: security.SafeUint8FromUint32(r >> 8),
		G: security.SafeUint8FromUint32(g >> 8),
		B: security.SafeUint8FromUint32(b >> 8),
	}
}

// ToRGBA converts a color.Color to RGBA.
// If the color doesn't have an alpha channel, it defaults to 255 (fully opaque).
func ToRGBA(c color.Color) RGBA {
	r, g, b, a := c.RGBA()
	// RGBA returns values in the range [0, 65535], convert to [0, 255].
	return RGBA{
		R: security.SafeUint8FromUint32(r >> 8),
		G: security.SafeUint8FromUint32(g >> 8),
		B: security.SafeUint8FromUint32(b >> 8),
		A: security.SafeUint8FromUint32(a >> 8),
	}
}

// ToHex converts the palette colors to hex strings.
// Returns a slice of hex color codes (e.g., ["#1a2b3c", "#4d5e6f"]).
func (p *Palette) ToHex() []string {
	hexColors := make([]string, len(p.Colors))
	for i, c := range p.Colors {
		rgb := ToRGB(c)
		hexColors[i] = rgb.Hex()
	}
	return hexColors
}

// ToRGBSlice converts the palette colors to RGB structs.
func (p *Palette) ToRGBSlice() []RGB {
	rgbColors := make([]RGB, len(p.Colors))
	for i, c := range p.Colors {
		rgbColors[i] = ToRGB(c)
	}
	return rgbColors
}

// ColorJSON represents a color in JSON output format.
type ColorJSON struct {
	Hex string `json:"hex"`
	RGB RGB    `json:"rgb"`
}

// PaletteJSON represents the palette in JSON format.
type PaletteJSON struct {
	Count  int         `json:"count"`
	Colors []ColorJSON `json:"colors"`
}

// ToJSON converts the palette to JSON format.
func (p *Palette) ToJSON() ([]byte, error) {
	colors := make([]ColorJSON, len(p.Colors))
	for i, c := range p.Colors {
		rgb := ToRGB(c)
		colors[i] = ColorJSON{
			Hex: rgb.Hex(),
			RGB: rgb,
		}
	}

	paletteJSON := PaletteJSON{
		Count:  len(p.Colors),
		Colors: colors,
	}

	return json.MarshalIndent(paletteJSON, "", "  ")
}

// String returns a human-readable string representation of the palette.
func (p *Palette) String() string {
	if len(p.Colors) == 0 {
		return "Empty palette"
	}

	result := fmt.Sprintf("Palette with %d colors:\n", len(p.Colors))
	for i, c := range p.Colors {
		rgb := ToRGB(c)
		result += fmt.Sprintf("  %2d: %s (%s)\n", i+1, rgb.Hex(), rgb.String())
	}
	return result
}

// Get returns the color at the specified index.
// Returns an error if the index is out of bounds.
func (p *Palette) Get(index int) (color.Color, error) {
	if index < 0 || index >= len(p.Colors) {
		return nil, fmt.Errorf("index out of bounds: %d (palette has %d colors)", index, len(p.Colors))
	}
	return p.Colors[index], nil
}

// All returns an iterator over all colors in the palette using Go 1.25 range over functions.
func (p *Palette) All() func(func(int, color.Color) bool) {
	return func(yield func(int, color.Color) bool) {
		for i, c := range p.Colors {
			if !yield(i, c) {
				return
			}
		}
	}
}
