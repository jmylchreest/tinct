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
