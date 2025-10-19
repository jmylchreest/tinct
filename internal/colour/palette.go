// Package color provides color extraction and palette generation functionality.
package colour

import (
	"encoding/json"
	"fmt"
	"image/color"
)

// Palette represents a collection of colors extracted from an image.
type Palette struct {
	Colors []color.Color
}

// NewPalette creates a new Palette with the given colors.
func NewPalette(colors []color.Color) *Palette {
	return &Palette{
		Colors: colors,
	}
}

// Len returns the number of colors in the palette.
func (p *Palette) Len() int {
	return len(p.Colors)
}

// RGB represents a color in RGB format.
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

// ToRGB converts a color.Color to RGB.
func ToRGB(c color.Color) RGB {
	r, g, b, _ := c.RGBA()
	// RGBA returns values in the range [0, 65535], convert to [0, 255]
	return RGB{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
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
