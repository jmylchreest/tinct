// Package colour provides foreground color selection logic.
package colour

// selectForeground selects the foreground color based on background and config.
//
// Design Theory (WCAG Accessibility Standards):.
// - Foreground is used for TEXT on the background.
// - MUST have minimum 4.5:1 contrast ratio with background (WCAG AA for normal text)
// - MUST have minimum 3:1 contrast ratio with background (WCAG AA for large text)
// - AAA standard requires 7:1 contrast for normal text.
// - Selects the color with HIGHEST contrast against background.
// - Hue is NOT considered - only contrast matters for readability.
//
// Returns the index of the selected foreground color, or -1 if none found.
func selectForeground(extracted []CategorisedColour, bg CategorisedColour, config CategorisationConfig, bgIdx int) int {
	fgIdx := -1
	maxContrast := 0.0
	minContrast := config.MinContrastRatio
	if config.RequireAAA {
		minContrast = 7.0 // WCAG AAA standard
	}

	// Find color with highest contrast that meets minimum threshold.
	for i, cc := range extracted {
		if i == bgIdx {
			continue // Skip background itself
		}
		contrast := ContrastRatio(cc.Colour, bg.Colour)
		if contrast > maxContrast && contrast >= minContrast {
			maxContrast = contrast
			fgIdx = i
		}
	}

	// Fallback: use color with highest contrast even if below threshold.
	// This ensures we always have a foreground color, even if palette is limited.
	if fgIdx < 0 {
		maxContrast = 0.0
		for i, cc := range extracted {
			if i == bgIdx {
				continue
			}
			contrast := ContrastRatio(cc.Colour, bg.Colour)
			if contrast > maxContrast {
				maxContrast = contrast
				fgIdx = i
			}
		}
	}

	return fgIdx
}

// generateSyntheticForeground creates a foreground color when none can be extracted.
// This is used for monochromatic images where all colors are too similar.
//
// Design Theory:.
// - Foreground must have high contrast with background (minimum 4.5:1 for text)
// - For monochromatic palettes, generate the "inverse" of background.
// - Maintains same hue as background (visual cohesion)
// - Adjusts luminance to opposite end of spectrum (light vs dark)
// - Ensures WCAG AA compliance for text readability.
func generateSyntheticForeground(bg CategorisedColour, theme ThemeType, config CategorisationConfig) CategorisedColour {
	h, s, _ := rgbToHSL(bg.RGB)

	// Determine target luminance (opposite end of spectrum).
	var targetLum float64
	if theme == ThemeDark {
		// Dark theme: foreground should be very light.
		targetLum = 0.90
	} else {
		// Light theme: foreground should be very dark.
		targetLum = 0.10
	}

	// Reduce saturation slightly for text (less visual vibration).
	targetSat := s * 0.7

	// Adjust luminance iteratively until we hit minimum contrast.
	minContrast := config.MinContrastRatio
	if config.RequireAAA {
		minContrast = 7.0
	}

	var fgRGB RGB
	targetLum, fgRGB = adjustLuminanceForContrast(h, targetSat, targetLum, bg.Colour, minContrast, theme, 20)

	return CategorisedColour{
		Colour:      RGBToColor(fgRGB),
		Hex:         fgRGB.Hex(),
		RGB:         fgRGB,
		RGBA:        RGBToRGBA(fgRGB),
		Luminance:   targetLum,
		IsLight:     targetLum > 0.5,
		Hue:         h,
		Saturation:  targetSat,
		IsGenerated: true,
		Weight:      0,
	}
}
