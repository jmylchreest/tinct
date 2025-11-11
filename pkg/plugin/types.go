// Package plugin provides the public API for tinct plugins.
package plugin

// InputOptions holds options for input plugin generation.
type InputOptions struct {
	Verbose         bool           `json:"verbose"`
	DryRun          bool           `json:"dry_run"`
	ColourOverrides []string       `json:"colour_overrides,omitempty"`
	PluginArgs      map[string]any `json:"plugin_args,omitempty"`
}

// PaletteData is the palette data sent to output plugins.
type PaletteData struct {
	Colours    map[string]CategorisedColour `json:"colours"`
	AllColours []CategorisedColour          `json:"all_colours"`
	ThemeType  string                       `json:"theme_type"`
	PluginArgs map[string]any               `json:"plugin_args,omitempty"`
	DryRun     bool                         `json:"dry_run"`
}

// CategorisedColour represents a color with metadata for RPC transfer.
type CategorisedColour struct {
	RGB        RGBColour `json:"rgb"`
	Hex        string    `json:"hex"`
	Role       string    `json:"role,omitempty"`
	Luminance  float64   `json:"luminance,omitempty"`
	IsLight    bool      `json:"is_light,omitempty"`
	Hue        float64   `json:"hue,omitempty"`
	Saturation float64   `json:"saturation,omitempty"`
	Index      int       `json:"index,omitempty"`
}

// RGBColour represents an RGB color.
type RGBColour struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}
