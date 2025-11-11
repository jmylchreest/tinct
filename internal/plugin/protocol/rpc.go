// Package protocol provides conversion utilities between internal and external plugin types.
package protocol

import (
	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/pkg/plugin"
)

// ConvertCategorisedPalette converts a colour.CategorisedPalette to PaletteData for RPC.
func ConvertCategorisedPalette(palette *colour.CategorisedPalette, pluginArgs map[string]any, dryRun bool) plugin.PaletteData {
	data := plugin.PaletteData{
		Colours:    make(map[string]plugin.CategorisedColour),
		AllColours: make([]plugin.CategorisedColour, len(palette.AllColours)),
		ThemeType:  palette.ThemeType.String(),
		PluginArgs: pluginArgs,
		DryRun:     dryRun,
	}

	// Convert colours map.
	for role, cc := range palette.Colours {
		data.Colours[string(role)] = plugin.CategorisedColour{
			RGB: plugin.RGBColour{
				R: cc.RGB.R,
				G: cc.RGB.G,
				B: cc.RGB.B,
			},
			Hex:        cc.Hex,
			Role:       string(cc.Role),
			Luminance:  cc.Luminance,
			IsLight:    cc.IsLight,
			Hue:        cc.Hue,
			Saturation: cc.Saturation,
			Index:      cc.Index,
		}
	}

	// Convert all colours slice.
	for i, cc := range palette.AllColours {
		data.AllColours[i] = plugin.CategorisedColour{
			RGB: plugin.RGBColour{
				R: cc.RGB.R,
				G: cc.RGB.G,
				B: cc.RGB.B,
			},
			Hex:        cc.Hex,
			Role:       string(cc.Role),
			Luminance:  cc.Luminance,
			IsLight:    cc.IsLight,
			Hue:        cc.Hue,
			Saturation: cc.Saturation,
			Index:      cc.Index,
		}
	}

	return data
}
