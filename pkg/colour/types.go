// Package colour provides public types and utilities for working with color palettes.
// This package is designed for use by external plugins that need to access and format
// categorized color data from tinct.
package colour

import (
	"fmt"
	"image/color"
)

// ThemeType represents whether a theme is light or dark.
type ThemeType int

const (
	// ThemeDark represents a dark theme.
	ThemeDark ThemeType = iota
	// ThemeLight represents a light theme.
	ThemeLight
)

// String returns the string representation of the theme type.
func (t ThemeType) String() string {
	switch t {
	case ThemeDark:
		return "dark"
	case ThemeLight:
		return "light"
	default:
		return "unknown"
	}
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
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
	}
}

// ToRGBA converts a color.Color to RGBA.
func ToRGBA(c color.Color) RGBA {
	r, g, b, a := c.RGBA()
	// RGBA returns values in the range [0, 65535], convert to [0, 255].
	return RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}
