package colour

import "image/color"

// createCategorisedColours creates categorised colours with metadata from raw palette.
func createCategorisedColours(palette *Palette) []CategorisedColour {
	extracted := make([]CategorisedColour, len(palette.Colors))
	weights := normalizeWeights(palette.Weights, len(palette.Colors))

	for i, c := range palette.Colors {
		extracted[i] = createCategorisedColour(c, weights[i])
	}

	return extracted
}

// normalizeWeights creates equal weights if none provided.
func normalizeWeights(weights []float64, count int) []float64 {
	if weights == nil || len(weights) != count {
		result := make([]float64, count)
		equalWeight := 1.0 / float64(count)
		for i := range result {
			result[i] = equalWeight
		}
		return result
	}
	return weights
}

// createCategorisedColour creates a single categorised colour with metadata.
func createCategorisedColour(c color.Color, weight float64) CategorisedColour {
	lum := Luminance(c)
	rgb := ToRGB(c)
	rgba := ToRGBA(c)
	h, s, _ := rgbToHSL(rgb)

	return CategorisedColour{
		Colour:      c,
		Hex:         rgb.Hex(),
		RGB:         rgb,
		RGBA:        rgba,
		Luminance:   lum,
		IsLight:     lum > 0.5,
		Hue:         h,
		Saturation:  s,
		IsGenerated: false,
		Weight:      weight,
	}
}

// selectBackgroundWithHints selects background color, applying hints if available.
func selectBackgroundWithHints(extracted, allExtracted []CategorisedColour,
	hints map[Role]int, themeType ThemeType, hintsApplied map[Role]bool) (CategorisedColour, int, ThemeType) {

	// Check for hinted background.
	if bgIdx, hasHint := hints[RoleBackground]; hasHint {
		if bgIdx >= 0 && bgIdx < len(allExtracted) {
			bg := allExtracted[bgIdx]
			bg.Role = RoleBackground
			hintsApplied[RoleBackground] = true
			return bg, bgIdx, themeType
		}
	}

	// No hint, select background normally.
	bg, finalTheme := selectBackground(extracted, themeType)

	// Find index.
	bgIdx := findColourIndex(extracted, bg.Hex)

	return bg, bgIdx, finalTheme
}

// applyRoleHints applies all non-background role hints to the palette.
func applyRoleHints(result *CategorisedPalette, extracted, allExtracted []CategorisedColour,
	hints map[Role]int, hintsApplied map[Role]bool) {

	if hints == nil {
		return
	}

	for role, originalIndex := range hints {
		if role == RoleBackground {
			continue // Already handled
		}

		if originalIndex >= 0 && originalIndex < len(allExtracted) {
			hintedColor := allExtracted[originalIndex]
			// Find in sorted array.
			for _, cc := range extracted {
				if cc.Hex == hintedColor.Hex {
					cc.Role = role
					result.Set(role, cc)
					hintsApplied[role] = true
					break
				}
			}
		}
	}
}

// selectForegroundWithHints selects foreground color, applying hints if available.
func selectForegroundWithHints(result *CategorisedPalette, extracted []CategorisedColour,
	bg CategorisedColour, bgIdx int, config CategorisationConfig,
	themeType ThemeType, hintsApplied map[Role]bool) (fg CategorisedColour, fgIdx int) {

	// Check if already hinted.
	if hintsApplied[RoleForeground] {
		fg, _ = result.Get(RoleForeground)
		fgIdx = findColourIndex(extracted, fg.Hex)
		return fg, fgIdx
	}

	// Select foreground.
	fgIdx = selectForeground(extracted, bg, config, bgIdx)
	if fgIdx >= 0 {
		fg = extracted[fgIdx]
		fg.Role = RoleForeground
		result.Set(RoleForeground, fg)
		return fg, fgIdx
	}

	// Generate synthetic foreground.
	fg = generateSyntheticForeground(bg, themeType, config)
	fg.Role = RoleForeground
	result.Set(RoleForeground, fg)
	return fg, -1
}

// addMutedVariants creates and adds muted variants for background and foreground.
func addMutedVariants(result *CategorisedPalette, bg, fg CategorisedColour,
	themeType ThemeType, config CategorisationConfig, hints map[Role]int) {

	// Background muted.
	if _, hasHint := hints[RoleBackgroundMuted]; !hasHint {
		bgMuted := createMutedVariant(bg, config.MutedLuminanceAdjust, themeType, true)
		bgMuted.Role = RoleBackgroundMuted
		bgMuted.IsGenerated = true
		result.Set(RoleBackgroundMuted, bgMuted)
	}

	// Foreground muted (if foreground exists).
	if _, hasFg := result.Get(RoleForeground); hasFg {
		if _, hasHint := hints[RoleForegroundMuted]; !hasHint {
			fgMuted := createMutedVariant(fg, config.MutedLuminanceAdjust, themeType, false)
			fgMuted.Role = RoleForegroundMuted
			fgMuted.IsGenerated = true
			result.Set(RoleForegroundMuted, fgMuted)
		}
	}
}

// collectAccentColours collects colors for accents, excluding used indices.
func collectAccentColours(extracted []CategorisedColour, hints map[Role]int, bgIdx, fgIdx int) []CategorisedColour {
	usedIndices := buildUsedIndicesSet(hints, bgIdx, fgIdx)

	accents := make([]CategorisedColour, 0)
	for i, cc := range extracted {
		if !usedIndices[i] {
			accents = append(accents, cc)
		}
	}

	return accents
}

// buildUsedIndicesSet builds a set of indices that are already used.
func buildUsedIndicesSet(hints map[Role]int, bgIdx, fgIdx int) map[int]bool {
	used := make(map[int]bool)

	// Mark background and foreground indices as used.
	if bgIdx >= 0 {
		used[bgIdx] = true
	}
	if fgIdx >= 0 {
		used[fgIdx] = true
	}

	// Mark any hinted role indices as used.
	for _, index := range hints {
		used[index] = true
	}

	return used
}

// assignAccentRoles assigns accent colors to primary and muted roles.
func assignAccentRoles(result *CategorisedPalette, accents []CategorisedColour,
	themeType ThemeType, config CategorisationConfig, hints map[Role]int) {

	accentRoles := []struct {
		primary Role
		muted   Role
	}{
		{RoleAccent1, RoleAccent1Muted},
		{RoleAccent2, RoleAccent2Muted},
		{RoleAccent3, RoleAccent3Muted},
		{RoleAccent4, RoleAccent4Muted},
	}

	accentIndex := 0
	for _, roles := range accentRoles {
		if accentIndex >= len(accents) {
			break
		}

		assignAccentPair(result, accents, &accentIndex, roles, themeType, config, hints)
	}
}

// assignAccentPair assigns a single accent pair (primary + muted).
func assignAccentPair(result *CategorisedPalette, accents []CategorisedColour,
	accentIndex *int, roles struct{ primary, muted Role },
	themeType ThemeType, config CategorisationConfig, hints map[Role]int) {

	// Check for hint.
	if _, hasHint := hints[roles.primary]; hasHint {
		return // Already applied
	}

	if *accentIndex >= len(accents) {
		return
	}

	accent := accents[*accentIndex]
	accent.Role = roles.primary
	result.Set(roles.primary, accent)

	// Create muted variant if not hinted.
	if _, hasHint := hints[roles.muted]; !hasHint {
		muted := createMutedVariant(accent, config.MutedLuminanceAdjust, themeType, false)
		muted.Role = roles.muted
		muted.IsGenerated = true
		result.Set(roles.muted, muted)
	}

	*accentIndex++
}

// findColourIndex finds the index of a color by hex value.
func findColourIndex(colours []CategorisedColour, hex string) int {
	for i, cc := range colours {
		if cc.Hex == hex {
			return i
		}
	}
	return -1
}

// needsSyntheticAccents determines if synthetic accents should be generated.
func needsSyntheticAccents(accents []CategorisedColour, bg CategorisedColour) bool {
	return len(accents) < 4 || areAccentsTooSimilar(accents, bg)
}
