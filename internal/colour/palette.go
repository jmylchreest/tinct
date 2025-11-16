// Package color provides color extraction and palette generation functionality.
package colour

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strings"

	pkgcolour "github.com/jmylchreest/tinct/pkg/colour"
)

// Re-export conversion functions from pkg/colour.
var (
	ToRGB  = pkgcolour.ToRGB
	ToRGBA = pkgcolour.ToRGBA
)

// Palette represents a collection of colors extracted from an image.
type Palette struct {
	Colors    []color.Color
	Weights   []float64    // Optional: relative frequency/volume of each color (0.0-1.0)
	RoleHints map[Role]int // Optional: explicit role assignments (role -> color index)
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
func NewPaletteWithRoleHints(colors []color.Color, roleHints map[Role]int) *Palette {
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

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Palette with %d colors:\n", len(p.Colors)))
	for i, c := range p.Colors {
		rgb := ToRGB(c)
		result.WriteString(fmt.Sprintf("  %2d: %s (%s)\n", i+1, rgb.Hex(), rgb.String()))
	}
	return result.String()
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
