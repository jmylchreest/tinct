// Package colour provides palette helper utilities for output plugins.
package colour

import (
	pkgcolour "github.com/jmylchreest/tinct/pkg/colour"
)

// Re-export public types from pkg/colour for backward compatibility.
type (
	ColorFormat   = pkgcolour.ColorFormat
	ColorValue    = pkgcolour.ColorValue
	PaletteHelper = pkgcolour.PaletteHelper
)

// Re-export ColorFormat constants.
const (
	FormatHex         = pkgcolour.FormatHex
	FormatHexAlpha    = pkgcolour.FormatHexAlpha
	FormatRGB         = pkgcolour.FormatRGB
	FormatRGBA        = pkgcolour.FormatRGBA
	FormatHexNoHash   = pkgcolour.FormatHexNoHash
	FormatRGBDecimal  = pkgcolour.FormatRGBDecimal
	FormatRGBADecimal = pkgcolour.FormatRGBADecimal
)

// NewColorValue creates a ColorValue from RGBA with optional metadata.
var NewColorValue = pkgcolour.NewColorValue

// NewPaletteHelper creates a helper for the given categorised palette.
// This is the internal factory that creates a PaletteHelper from a CategorisedPalette.
func NewPaletteHelper(palette *CategorisedPalette) *PaletteHelper {
	// Build color map and indexed array from palette.
	colors := make(map[Role]ColorValue, len(palette.Colours))
	indexed := make([]ColorValue, 0, len(palette.AllColours))

	for role, cc := range palette.Colours {
		colors[role] = NewColorValue(cc.RGBA, role, cc.Index)
	}

	for _, cc := range palette.AllColours {
		indexed = append(indexed, NewColorValue(cc.RGBA, cc.Role, cc.Index))
	}

	return pkgcolour.NewPaletteHelper(palette.ThemeType, colors, indexed)
}
