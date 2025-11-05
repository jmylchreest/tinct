package main

import (
	"fmt"
	"strings"
)

// PaletteInput represents the JSON input from Tinct
type PaletteInput struct {
	Colours    map[string]CategorisedColour `json:"colours"`
	AllColours []CategorisedColour          `json:"all_colours"`
	ThemeType  int                          `json:"theme_type"`
	PluginArgs map[string]interface{}       `json:"plugin_args"`
	DryRun     bool                         `json:"dry_run"`
}

// CategorisedColour represents a color from the palette
type CategorisedColour struct {
	Hex   string `json:"hex"`
	Role  string `json:"role"`
	Index int    `json:"index"`
	RGB   RGB    `json:"rgb"`
}

// RGB represents RGB color values
type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// ThemeData is the data structure passed to templates
// This mirrors the internal tinct ThemeData structure
type ThemeData struct {
	Colors        map[string]*ColorValue
	AllColors     []*ColorValue
	WallpaperPath string
	ThemeName     string
	themeType     string
}

// ColorValue represents a color with multiple format accessors
type ColorValue struct {
	hex   string
	role  string
	index int
	rgb   RGB
}

// Format methods for ColorValue
func (c *ColorValue) Hex() string        { return c.hex }
func (c *ColorValue) HexNoHash() string  { return strings.TrimPrefix(c.hex, "#") }
func (c *ColorValue) RGB() string        { return fmt.Sprintf("rgb(%d,%d,%d)", c.rgb.R, c.rgb.G, c.rgb.B) }
func (c *ColorValue) RGBA() string       { return fmt.Sprintf("rgba(%d,%d,%d,1)", c.rgb.R, c.rgb.G, c.rgb.B) }
func (c *ColorValue) R() uint8           { return c.rgb.R }
func (c *ColorValue) G() uint8           { return c.rgb.G }
func (c *ColorValue) B() uint8           { return c.rgb.B }
func (c *ColorValue) Role() string       { return c.role }
func (c *ColorValue) Index() int         { return c.index }
func (c *ColorValue) RGBDecimal() string { return fmt.Sprintf("%d,%d,%d", c.rgb.R, c.rgb.G, c.rgb.B) }

// ThemeType returns the theme type as a string
func (td *ThemeData) ThemeType() string {
	return td.themeType
}

// ProcessingResult represents the result of processing a template
type ProcessingResult struct {
	TemplateName string
	OutputPath   string
	Success      bool
	Error        error
	BytesWritten int
	Content      string // For go-plugin mode - rendered template content
}

// convertToThemeData converts PaletteInput to ThemeData for templates
func convertToThemeData(input *PaletteInput) *ThemeData {
	themeData := &ThemeData{
		Colors:    make(map[string]*ColorValue),
		AllColors: make([]*ColorValue, 0, len(input.AllColours)),
	}

	// Convert theme type integer to string
	switch input.ThemeType {
	case 1:
		themeData.themeType = "dark"
	case 2:
		themeData.themeType = "light"
	default:
		themeData.themeType = "auto"
	}

	// Convert colors map
	for role, color := range input.Colours {
		themeData.Colors[role] = &ColorValue{
			hex:   color.Hex,
			role:  role,
			index: color.Index,
			rgb:   color.RGB,
		}
	}

	// Convert all colors array
	for _, color := range input.AllColours {
		themeData.AllColors = append(themeData.AllColors, &ColorValue{
			hex:   color.Hex,
			role:  color.Role,
			index: color.Index,
			rgb:   color.RGB,
		})
	}

	return themeData
}
