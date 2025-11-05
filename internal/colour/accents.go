// Package colour provides accent color selection logic.
package colour

import (
	"math"
)

// Minimum contrast requirements for accents.
const (
	MinAccentBgContrast     = 3.0  // Minimum contrast between accent and background (WCAG AA for large text)
	MinAccentAccentContrast = 1.5  // Minimum contrast between consecutive accents for visual distinction
	MaxAccentSimilarity     = 0.05 // Maximum luminance difference to consider accents "identical" (5%)
)

// sortAccentsForTheme sorts accent colors to create optimal visual progression.
//
// Design Theory (Based on Industry Standards):.
// 1. ANALOGOUS HARMONY: Accents should be similar in hue to BACKGROUND (not foreground)
//   - Creates calm, cohesive designs.
//   - Analogous colors are within 30° on the color wheel.
//   - 60-30-10 rule: 60% background, 30% secondary, 10% accent.
//
// 2. CONTRAST PROGRESSION:
//   - Dark themes: accent1 = LIGHTEST (highest contrast) → accent4 = DARKEST
//   - Light themes: accent1 = DARKEST (highest contrast) → accent4 = LIGHTEST
//   - Material Design: use desaturated colors in dark themes to avoid visual vibration.
//   - More saturated colors "vibrate" against dark backgrounds, causing eye strain.
//
// 3. ACCESSIBILITY:
//   - All accents must have at least 3:1 contrast with background (WCAG AA)
//   - Accents should be distinguishable from each other (visual hierarchy)
//
// 4. SATURATION:
//   - Prefer more saturated colors for accents (visual distinctiveness)
//   - But not too saturated in dark themes (visual vibration)
func sortAccentsForTheme(accents []CategorisedColour, bg, fg CategorisedColour, theme ThemeType) {
	if len(accents) == 0 {
		return
	}

	// Calculate a score for each accent based on multiple factors.
	type accentScore struct {
		index int
		score float64
	}

	scores := make([]accentScore, len(accents))

	for i, accent := range accents {
		score := 0.0

		// Factor 1: Similarity to BACKGROUND hue (0-1, higher is better).
		// FIXED: Changed from foreground to background for analogous harmony.
		hueDiff := HueDistance(accent.Hue, bg.Hue)
		hueSimilarity := 1.0 - (hueDiff / 180.0)
		score += hueSimilarity * 3.0 // Weight: 3.0

		// Factor 2: Saturation similarity to background (prefer colorful accents).
		satDiff := math.Abs(accent.Saturation - bg.Saturation)
		satSimilarity := 1.0 - satDiff
		score += satSimilarity * 2.0 // Weight: 2.0

		// Factor 3: Contrast with background (must be readable).
		bgContrast := ContrastRatio(accent.Colour, bg.Colour)
		if bgContrast >= 4.5 {
			score += 2.0 // Good contrast bonus
		} else if bgContrast >= 3.0 {
			score += 1.0 // Acceptable contrast
		}

		// Factor 4: Contrast with foreground (accents should be distinguishable).
		fgContrast := ContrastRatio(accent.Colour, fg.Colour)
		if fgContrast >= 3.0 {
			score += 1.5 // Good distinction bonus
		} else if fgContrast >= 2.0 {
			score += 0.75
		}

		// Factor 5: Saturation (prefer more saturated colors for accents).
		if accent.Saturation >= 0.4 {
			score += accent.Saturation * 1.5 // Weight: 1.5
		}

		scores[i] = accentScore{index: i, score: score}
	}

	// Sort by score (highest first).
	n := len(scores)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if scores[j].score < scores[j+1].score {
				scores[j], scores[j+1] = scores[j+1], scores[j]
			}
		}
	}

	// Reorder accents based on scores.
	reordered := make([]CategorisedColour, len(accents))
	for i, s := range scores {
		reordered[i] = accents[s.index]
	}

	// FIXED: Correct contrast progression based on theme type.
	// Now sort the top candidates by luminance for visual progression.
	// Dark themes: accent1 = LIGHTEST (highest contrast) → accent4 = DARKEST.
	// Light themes: accent1 = DARKEST (highest contrast) → accent4 = LIGHTEST.
	topCount := min(12, len(reordered)) // Consider top 12 candidates
	if topCount == 0 {
		return
	}

	if theme == ThemeDark {
		// Dark theme: Sort LIGHTEST → DARKEST (descending luminance).
		// accent1 will be the lightest/highest contrast.
		for i := 0; i < topCount-1; i++ {
			for j := 0; j < topCount-i-1; j++ {
				if reordered[j].Luminance < reordered[j+1].Luminance {
					reordered[j], reordered[j+1] = reordered[j+1], reordered[j]
				}
			}
		}
	} else {
		// Light theme: Sort DARKEST → LIGHTEST (ascending luminance).
		// accent1 will be the darkest/highest contrast.
		for i := 0; i < topCount-1; i++ {
			for j := 0; j < topCount-i-1; j++ {
				if reordered[j].Luminance > reordered[j+1].Luminance {
					reordered[j], reordered[j+1] = reordered[j+1], reordered[j]
				}
			}
		}
	}

	// Copy back to original slice.
	copy(accents, reordered)
}

// areAccentsTooSimilar checks if accents lack sufficient diversity.
// Returns true if accents are too similar to each other or to the background.
func areAccentsTooSimilar(accents []CategorisedColour, bg CategorisedColour) bool {
	if len(accents) == 0 {
		return true
	}

	// Check if all accents are nearly identical to background.
	identicalToBg := 0
	for _, accent := range accents {
		lumDiff := math.Abs(accent.Luminance - bg.Luminance)
		contrast := ContrastRatio(accent.Colour, bg.Colour)

		if lumDiff < MaxAccentSimilarity || contrast < MinAccentBgContrast {
			identicalToBg++
		}
	}

	// If more than 75% of accents are too similar to background, palette is monochromatic.
	if float64(identicalToBg)/float64(len(accents)) > 0.75 {
		return true
	}

	// Check if accents are too similar to each other.
	if len(accents) >= 2 {
		identicalToEachOther := 0
		for i := 0; i < len(accents)-1; i++ {
			lumDiff := math.Abs(accents[i].Luminance - accents[i+1].Luminance)
			if lumDiff < MaxAccentSimilarity {
				identicalToEachOther++
			}
		}

		// If more than 75% of accent pairs are identical, generate new ones.
		if float64(identicalToEachOther)/float64(len(accents)-1) > 0.75 {
			return true
		}
	}

	return false
}

// generateSyntheticAccents creates accent colors when extracted palette lacks diversity.
//
// Design Theory for Monochromatic Palettes:
// - When extracted colors are too similar, generate accents with guaranteed contrast.
// - Introduces hue diversity using complementary/triadic color theory.
// - Accent1: Highest contrast with background, analogous hue (±30°)
// - Accent2-4: Complementary and triadic hues for visual interest.
// - Each accent must have minimum 3:1 contrast with background.
// - Adjusts saturation based on theme (desaturated for dark themes).
func generateSyntheticAccents(bg CategorisedColour, theme ThemeType, count int) []CategorisedColour {
	if count <= 0 {
		return nil
	}

	h, s, l := rgbToHSL(bg.RGB)
	accents := make([]CategorisedColour, 0, count)

	// Define hue offsets for accent diversity.
	// Using color theory: analogous (±30°), complementary (180°), triadic (120°, 240°).
	hueOffsets := []float64{
		30,  // accent1: analogous (warm/cool shift)
		180, // accent2: complementary (maximum contrast)
		120, // accent3: triadic (harmonious diversity)
		240, // accent4: triadic (harmonious diversity)
	}

	// Target saturation: slightly higher than background for accent visibility.
	// But reduce in dark themes to avoid visual vibration.
	targetSat := s
	if targetSat < 0.5 {
		targetSat = 0.6 // Boost saturation for muted backgrounds
	}
	if theme == ThemeDark && targetSat > 0.7 {
		targetSat = 0.7 // Cap saturation in dark themes to reduce eye strain
	}

	// Determine base luminance for accents (need good contrast with background).
	var baseLum float64
	if theme == ThemeDark {
		// Dark theme: accents should be lighter than background.
		baseLum = math.Min(0.75, l+0.40)
	} else {
		// Light theme: accents should be darker than background.
		baseLum = math.Max(0.25, l-0.40)
	}

	// Generate accents with varied hues.
	for i := 0; i < count; i++ {
		// Calculate hue with offset for diversity.
		offset := hueOffsets[i%len(hueOffsets)]
		newHue := math.Mod(h+offset, 360.0)

		// Vary luminance slightly across accents for visual progression.
		// accent1 = highest contrast, accent4 = closer to background.
		lumAdjust := float64(i) * 0.05
		var accentLum float64
		if theme == ThemeDark {
			accentLum = baseLum - lumAdjust // Get progressively darker
		} else {
			accentLum = baseLum + lumAdjust // Get progressively lighter
		}
		accentLum = math.Max(0.15, math.Min(0.85, accentLum))

		// Adjust saturation slightly for each accent.
		accentSat := targetSat * (0.95 + 0.1*float64(i)/float64(count))
		accentSat = math.Max(0.4, math.Min(1.0, accentSat))

		// Generate RGB and ensure minimum contrast with background.
		var accentRGB RGB
		accentLum, accentRGB = adjustLuminanceForContrast(newHue, accentSat, accentLum, bg.Colour, MinAccentBgContrast, theme, 10)

		accents = append(accents, CategorisedColour{
			Colour:      RGBToColor(accentRGB),
			Hex:         accentRGB.Hex(),
			RGB:         accentRGB,
			RGBA:        RGBToRGBA(accentRGB),
			Luminance:   accentLum,
			IsLight:     accentLum > 0.5,
			Hue:         newHue,
			Saturation:  accentSat,
			IsGenerated: true,
			Weight:      0,
		})
	}

	return accents
}
