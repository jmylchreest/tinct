// Package colour provides muted color variant generation logic.
package colour

import (
	"math"
)

// createMutedVariant creates a muted variant of a colour for inactive/disabled states.
//
// Design Theory (Industry Standards):.
// - Muted colors are used for INACTIVE or DISABLED UI elements.
// - Created by reducing saturation by ~50% (to 50% of original)
// - Luminance adjusted by ±10-15% to maintain perceptual consistency.
// - Creates visual hierarchy: active vs inactive states.
// - Common in design systems (Material Design, Atlassian, etc.)
//
// Parameters:.
// - cc: The base color to create a muted variant from.
// - adjustment: Luminance adjustment amount (typically 0.15 = 15%)
// - themeType: Dark or light theme affects luminance adjustment direction.
// - isBackground: Background vs foreground affects adjustment direction.
func createMutedVariant(cc CategorisedColour, adjustment float64, themeType ThemeType, isBackground bool) CategorisedColour {
	h, s, l := rgbToHSL(cc.RGB)

	// Luminance adjustment based on theme and role.
	newLum := l
	if isBackground {
		if themeType == ThemeDark {
			// Dark background: make slightly lighter.
			newLum = math.Min(1.0, l+adjustment)
		} else {
			// Light background: make slightly darker.
			newLum = math.Max(0.0, l-adjustment)
		}
	} else {
		// Foreground/accent: reduce contrast slightly.
		if themeType == ThemeDark {
			// Light text/accent: make slightly darker.
			newLum = math.Max(0.0, l-adjustment)
		} else {
			// Dark text/accent: make slightly lighter.
			newLum = math.Min(1.0, l+adjustment)
		}
	}

	// Saturation reduction: ~50% (reduce to half of original).
	// This is the key change from the previous 80% (0.8 factor).
	// Industry standard for muted/inactive states.
	newSat := s * 0.5

	// Convert back to RGB.
	newRGB := HSLToRGB(h, newSat, newLum)
	newColor := RGBToColor(newRGB)

	// Calculate actual relative luminance (WCAG standard) from the RGB color.
	actualLuminance := Luminance(newColor)

	return CategorisedColour{
		Colour:     newColor,
		Hex:        newRGB.Hex(),
		RGB:        newRGB,
		RGBA:       RGBToRGBA(newRGB),
		Luminance:  actualLuminance, // Use calculated relative luminance, not HSL lightness
		IsLight:    actualLuminance > 0.5,
		Hue:        h,
		Saturation: newSat,
	}
}
