// Package colour provides semantic color assignment logic.
package colour

import (
	"math"
)

// Semantic color constants.
const (
	MinSemanticSaturation = 0.6  // Minimum saturation for semantic colors
	MinSemanticLightness  = 0.35 // Minimum lightness for semantic colors
	MaxSemanticLightness  = 0.65 // Maximum lightness for semantic colors
)

// SemanticHues defines the standard hue values for semantic colors.
// Based on universal color psychology and industry standards.
var SemanticHues = map[ColourRole]float64{
	RoleDanger:       0,   // Red - danger, errors, destructive actions
	RoleWarning:      45,  // Orange - warnings, caution
	RoleSuccess:      120, // Green - success, confirmation, positive actions
	RoleInfo:         210, // Blue - information, neutral notifications
	RoleNotification: 285, // Purple - badges, notifications, highlights
}

// assignSemanticRolesWithHints assigns semantic roles (danger, warning, success, etc.) based on hue.
// Skips roles that were explicitly provided via role hints.
//
// Design Theory:.
// - Semantic colors signal user information (success, error, warning)
// - Green = success (positive connotation, universal)
// - Red = danger/error (warning, universal)
// - Orange/Yellow = warning (caution)
// - Blue = info (neutral information)
// - Purple = notification (badges, highlights)
// - Must have good contrast with background for visibility.
// - Enhanced saturation for visual distinctiveness.
func assignSemanticRolesWithHints(palette *CategorisedPalette, accents []CategorisedColour, usedForSemantic map[string]bool, hintsApplied map[ColourRole]bool) {
	// Map hue ranges to semantic roles.
	// Red: 0-30, 330-360 (danger).
	// Orange/Yellow: 30-60 (warning).
	// Green: 90-150 (success).
	// Blue: 180-240 (info).
	// Cyan/Purple: 240-330 (notification).

	var danger, warning, success, info, notification *CategorisedColour

	for i := range accents {
		cc := &accents[i]
		h := cc.Hue

		// Skip low saturation colours (too grey).
		if cc.Saturation < 0.3 {
			continue
		}

		if (h >= 0 && h < 30) || h >= 330 {
			// Red - danger.
			if danger == nil || cc.Saturation > danger.Saturation {
				danger = cc
			}
		} else if h >= 30 && h < 60 {
			// Orange/Yellow - warning.
			if warning == nil || cc.Saturation > warning.Saturation {
				warning = cc
			}
		} else if h >= 90 && h < 150 {
			// Green - success.
			if success == nil || cc.Saturation > success.Saturation {
				success = cc
			}
		} else if h >= 180 && h < 240 {
			// Blue - info.
			if info == nil || cc.Saturation > info.Saturation {
				info = cc
			}
		} else if h >= 240 && h < 330 {
			// Cyan/Purple - notification.
			if notification == nil || cc.Saturation > notification.Saturation {
				notification = cc
			}
		}
	}

	// Get background for theme-aware adjustments.
	bg, hasBg := palette.Get(RoleBackground)
	themeType := palette.ThemeType

	// Set semantic roles with enhancement (skip if role was explicitly hinted).
	if !hintsApplied[RoleDanger] {
		if danger != nil {
			enhanced := enhanceSemanticColour(*danger, RoleDanger, themeType, hasBg, bg)
			palette.Set(RoleDanger, enhanced)
			usedForSemantic[danger.Hex] = true
		} else {
			// Generate fallback danger color if none found.
			fallback := generateFallbackSemanticColour(RoleDanger, themeType, hasBg, bg)
			palette.Set(RoleDanger, fallback)
		}
	}

	if !hintsApplied[RoleWarning] {
		if warning != nil {
			enhanced := enhanceSemanticColour(*warning, RoleWarning, themeType, hasBg, bg)
			palette.Set(RoleWarning, enhanced)
			usedForSemantic[warning.Hex] = true
		} else {
			fallback := generateFallbackSemanticColour(RoleWarning, themeType, hasBg, bg)
			palette.Set(RoleWarning, fallback)
		}
	}

	if !hintsApplied[RoleSuccess] {
		if success != nil {
			enhanced := enhanceSemanticColour(*success, RoleSuccess, themeType, hasBg, bg)
			palette.Set(RoleSuccess, enhanced)
			usedForSemantic[success.Hex] = true
		} else {
			fallback := generateFallbackSemanticColour(RoleSuccess, themeType, hasBg, bg)
			palette.Set(RoleSuccess, fallback)
		}
	}

	if !hintsApplied[RoleInfo] {
		if info != nil {
			enhanced := enhanceSemanticColour(*info, RoleInfo, themeType, hasBg, bg)
			palette.Set(RoleInfo, enhanced)
			usedForSemantic[info.Hex] = true
		} else {
			fallback := generateFallbackSemanticColour(RoleInfo, themeType, hasBg, bg)
			palette.Set(RoleInfo, fallback)
		}
	}

	if !hintsApplied[RoleNotification] {
		if notification != nil {
			enhanced := enhanceSemanticColour(*notification, RoleNotification, themeType, hasBg, bg)
			palette.Set(RoleNotification, enhanced)
			usedForSemantic[notification.Hex] = true
		} else {
			fallback := generateFallbackSemanticColour(RoleNotification, themeType, hasBg, bg)
			palette.Set(RoleNotification, fallback)
		}
	}
}

// assignSemanticRoles assigns semantic roles (danger, warning, success, etc.) based on hue.
// This is a wrapper that calls assignSemanticRolesWithHints with no hints applied.
func assignSemanticRoles(palette *CategorisedPalette, accents []CategorisedColour, usedForSemantic map[string]bool) {
	assignSemanticRolesWithHints(palette, accents, usedForSemantic, make(map[ColourRole]bool))
}

// enhanceSemanticColour boosts saturation and adjusts lightness for better visibility.
func enhanceSemanticColour(cc CategorisedColour, role ColourRole, themeType ThemeType, hasBg bool, bg CategorisedColour) CategorisedColour {
	h, s, l := rgbToHSL(cc.RGB)

	// Boost saturation to minimum threshold.
	if s < MinSemanticSaturation {
		s = MinSemanticSaturation
	}

	// Adjust lightness based on theme.
	targetLightness := 0.5 // Default middle ground
	if themeType == ThemeDark {
		// Dark theme: make colors lighter for visibility.
		targetLightness = 0.60
	} else {
		// Light theme: make colors darker for visibility.
		targetLightness = 0.45
	}

	// Ensure within bounds.
	if l < MinSemanticLightness {
		l = targetLightness
	} else if l > MaxSemanticLightness {
		l = targetLightness
	} else {
		// Blend towards target.
		l = (l + targetLightness) / 2.0
	}

	// Ensure good contrast with background if available.
	if hasBg {
		newRGB := HSLToRGB(h, s, l)
		testColor := RGBToColor(newRGB)
		contrast := ContrastRatio(testColor, bg.Colour)

		// If contrast is too low, adjust lightness.
		if contrast < 3.0 {
			if themeType == ThemeDark {
				l = math.Min(0.75, l+0.15)
			} else {
				l = math.Max(0.30, l-0.15)
			}
		}
	}

	newRGB := HSLToRGB(h, s, l)
	newColor := RGBToColor(newRGB)
	return CategorisedColour{
		Colour:      newColor,
		Role:        role,
		Hex:         newRGB.Hex(),
		RGB:         newRGB,
		RGBA:        RGBToRGBA(newRGB),
		Luminance:   Luminance(newColor),
		IsLight:     l > 0.5,
		Hue:         h,
		Saturation:  s,
		IsGenerated: true, // Enhanced colour
	}
}

// generateFallbackSemanticColour creates a semantic color when none exists in the palette.
func generateFallbackSemanticColour(role ColourRole, themeType ThemeType, hasBg bool, bg CategorisedColour) CategorisedColour {
	// Get standard hue for this role.
	hue, exists := SemanticHues[role]
	if !exists {
		hue = 0 // Fallback to red
	}

	// Set saturation and lightness based on theme.
	saturation := 0.75 // Vibrant
	lightness := 0.5

	if themeType == ThemeDark {
		lightness = 0.60 // Lighter for dark backgrounds
	} else {
		lightness = 0.45 // Darker for light backgrounds
	}

	// Ensure good contrast with background if available.
	var newRGB RGB
	if hasBg {
		// Note: Using larger luminance steps (0.1 vs 0.05) for semantic colors.
		lightness, newRGB = adjustLuminanceForContrast(hue, saturation, lightness, bg.Colour, 3.0, themeType, 5)
	} else {
		newRGB = HSLToRGB(hue, saturation, lightness)
	}

	newColor := RGBToColor(newRGB)
	return CategorisedColour{
		Colour:      newColor,
		Role:        role,
		Hex:         newRGB.Hex(),
		RGB:         newRGB,
		RGBA:        RGBToRGBA(newRGB),
		Luminance:   Luminance(newColor),
		IsLight:     lightness > 0.5,
		Hue:         hue,
		Saturation:  saturation,
		IsGenerated: true, // Generated fallback colour
	}
}
