// Package protocol defines the plugin protocol version and compatibility checking.
package protocol

import (
	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/pkg/plugin"
)

// Type aliases to public plugin API for backward compatibility.
// Internal code can continue using protocol.* while external plugins use plugin.* directly.

type (
	InputPlugin           = plugin.InputPlugin
	OutputPlugin          = plugin.OutputPlugin
	InputOptions          = plugin.InputOptions
	PaletteData           = plugin.PaletteData
	CategorisedColour     = plugin.CategorisedColour
	RGBColour             = plugin.RGBColour
	InputPluginRPC        = plugin.InputPluginRPC
	OutputPluginRPC       = plugin.OutputPluginRPC
	InputPluginRPCClient  = plugin.InputPluginRPCClient
	OutputPluginRPCClient = plugin.OutputPluginRPCClient
)

// ConvertCategorisedPalette converts a colour.CategorisedPalette to PaletteData for RPC.
func ConvertCategorisedPalette(palette *colour.CategorisedPalette, pluginArgs map[string]any, dryRun bool) PaletteData {
	data := PaletteData{
		Colours:    make(map[string]CategorisedColour),
		AllColours: make([]CategorisedColour, len(palette.AllColours)),
		ThemeType:  palette.ThemeType.String(),
		PluginArgs: pluginArgs,
		DryRun:     dryRun,
	}

	// Convert colours map.
	for role, cc := range palette.Colours {
		data.Colours[string(role)] = CategorisedColour{
			RGB: RGBColour{
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
		data.AllColours[i] = CategorisedColour{
			RGB: RGBColour{
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
