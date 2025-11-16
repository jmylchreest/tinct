// Package colour provides public types and utilities for working with color palettes.
package colour

import "image/color"

// CategorisedColour represents a colour with its assigned role and metadata.
type CategorisedColour struct {
	Colour      color.Color `json:"-"`
	Role        Role        `json:"role"`
	Hex         string      `json:"hex"`  // #RRGGBB format (backwards compatible)
	RGB         RGB         `json:"rgb"`  // RGB without alpha (backwards compatible)
	RGBA        RGBA        `json:"rgba"` // RGBA with alpha channel (defaults to 255/opaque)
	Luminance   float64     `json:"luminance"`
	IsLight     bool        `json:"is_light"`
	Hue         float64     `json:"hue,omitempty"`          // HSL hue (0-360)
	Saturation  float64     `json:"saturation,omitempty"`   // HSL saturation (0-1)
	Index       int         `json:"index,omitempty"`        // Index in AllColours array (sorted by luminance)
	IsGenerated bool        `json:"is_generated,omitempty"` // True if colour was generated/enhanced, not extracted
	Weight      float64     `json:"weight,omitempty"`       // Original weight from palette (0.0-1.0, 0 if generated)
}

// CategorisedPalette represents a palette with categorised colours.
type CategorisedPalette struct {
	Colours    map[Role]CategorisedColour `json:"colours"`
	ThemeType  ThemeType                  `json:"theme_type"`
	AllColours []CategorisedColour        `json:"all_colours,omitempty"`
	ANSI       *ANSIColors                `json:"ansi,omitempty"` // ANSI terminal colors
}

// ANSIColors represents the 16-color ANSI palette for terminal emulators.
type ANSIColors struct {
	// Standard colors (0-7).
	Black   RGBA `json:"black"`
	Red     RGBA `json:"red"`
	Green   RGBA `json:"green"`
	Yellow  RGBA `json:"yellow"`
	Blue    RGBA `json:"blue"`
	Magenta RGBA `json:"magenta"`
	Cyan    RGBA `json:"cyan"`
	White   RGBA `json:"white"`

	// Bright colors (8-15).
	BrightBlack   RGBA `json:"bright_black"`
	BrightRed     RGBA `json:"bright_red"`
	BrightGreen   RGBA `json:"bright_green"`
	BrightYellow  RGBA `json:"bright_yellow"`
	BrightBlue    RGBA `json:"bright_blue"`
	BrightMagenta RGBA `json:"bright_magenta"`
	BrightCyan    RGBA `json:"bright_cyan"`
	BrightWhite   RGBA `json:"bright_white"`
}

// ThemeData wraps a categorised palette with additional metadata for plugins.
// This is the primary data structure passed to templates and external plugins.
type ThemeData struct {
	*PaletteHelper
	WallpaperPath string
	ThemeName     string
}

// NewThemeData creates a ThemeData wrapper from a CategorisedPalette.
// This is called by tinct when preparing data for plugins.
func NewThemeData(palette *CategorisedPalette, wallpaperPath, themeName string) *ThemeData {
	// Build color map and indexed array from palette.
	colors := make(map[Role]ColorValue, len(palette.Colours))
	indexed := make([]ColorValue, 0, len(palette.AllColours))

	for role, cc := range palette.Colours {
		colors[role] = ColorValue{
			role:  role,
			rgba:  cc.RGBA,
			index: cc.Index,
		}
	}

	for _, cc := range palette.AllColours {
		indexed = append(indexed, ColorValue{
			role:  cc.Role,
			rgba:  cc.RGBA,
			index: cc.Index,
		})
	}

	helper := NewPaletteHelper(palette.ThemeType, colors, indexed)

	return &ThemeData{
		PaletteHelper: helper,
		WallpaperPath: wallpaperPath,
		ThemeName:     themeName,
	}
}
