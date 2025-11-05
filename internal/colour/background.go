// Package colour provides background color selection logic.
package colour

// selectBackground selects the background color based on theme type.
//
// Design Theory:.
// - Background is the BASE color of the theme, used for the largest surface area.
// - For ThemeAuto: Theme is determined by the most dominant color's luminance.
//   - If dominant color is light (luminance >= 0.5): light theme.
//   - If dominant color is dark (luminance < 0.5): dark theme.
//
// - For ThemeDark/ThemeLight: Theme is user-specified.
//   - ThemeDark: Select most dominant DARK color (luminance < 0.5) by weight.
//   - ThemeLight: Select most dominant LIGHT color (luminance >= 0.5) by weight.
//
// - Everything else in the palette is proportionate to this background color.
func selectBackground(extracted []CategorisedColour, themeType ThemeType) (CategorisedColour, ThemeType) {
	if len(extracted) == 0 {
		// Fallback to a default background.
		return CategorisedColour{}, themeType
	}

	// For ThemeAuto: determine theme based on most dominant color.
	if themeType == ThemeAuto {
		maxWeight := 0.0
		maxWeightIdx := 0
		for i, color := range extracted {
			if color.Weight > maxWeight {
				maxWeight = color.Weight
				maxWeightIdx = i
			}
		}

		// Theme is determined by dominant color's luminance.
		const luminanceThreshold = 0.5
		if extracted[maxWeightIdx].Luminance >= luminanceThreshold {
			themeType = ThemeLight
		} else {
			themeType = ThemeDark
		}

		// Background is the most dominant color.
		bg := extracted[maxWeightIdx]
		bg.Role = RoleBackground
		return bg, themeType
	}

	// For explicit theme type: select most dominant color of appropriate luminance.
	maxWeight := 0.0
	maxWeightIdx := -1

	if themeType == ThemeDark {
		// Dark theme: find most dominant DARK color (luminance < 0.5).
		for i, color := range extracted {
			if color.Luminance < 0.5 && color.Weight > maxWeight {
				maxWeight = color.Weight
				maxWeightIdx = i
			}
		}
		// Fallback: if no dark colors, use darkest color.
		if maxWeightIdx != -1 {
			return extracted[maxWeightIdx], themeType
		}

		minLuminance := 1.0
		for i, color := range extracted {
			if color.Luminance < minLuminance {
				minLuminance = color.Luminance
				maxWeightIdx = i
			}
		}
		return extracted[maxWeightIdx], themeType
	}

	// Light theme: find most dominant LIGHT color (luminance >= 0.5).
	for i, color := range extracted {
		if color.Luminance >= 0.5 && color.Weight > maxWeight {
			maxWeight = color.Weight
			maxWeightIdx = i
		}
	}
	// Fallback: if no light colors, use lightest color.
	if maxWeightIdx != -1 {
		return extracted[maxWeightIdx], themeType
	}

	maxLuminance := 0.0
	for i, color := range extracted {
		if color.Luminance > maxLuminance {
			maxLuminance = color.Luminance
			maxWeightIdx = i
		}
	}

	if maxWeightIdx == -1 {
		// Ultimate fallback: use first color.
		maxWeightIdx = 0
	}

	bg := extracted[maxWeightIdx]
	bg.Role = RoleBackground
	return bg, themeType
}
