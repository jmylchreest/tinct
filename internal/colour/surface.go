// Package colour provides surface and container color generation.
package colour

import "image/color"

// newGeneratedColour constructs a CategorisedColour from computed HSL values.
// It is used by the generate* helpers to avoid repeating the same struct literal.
// For colours that also need Hue/Saturation fields, the caller sets them after.
func newGeneratedColour(role Role, newRGB RGB, newL float64) CategorisedColour {
	newColor := RGBToColor(newRGB)
	return CategorisedColour{
		Colour:      newColor,
		Role:        role,
		Hex:         newRGB.Hex(),
		RGB:         newRGB,
		RGBA:        RGBToRGBA(newRGB),
		Luminance:   Luminance(newColor),
		IsLight:     newL > 0.5,
		IsGenerated: true,
		Weight:      0,
	}
}

// generateSurfaceColors generates all surface, border, on-color, and container variants.
// These colors are essential for UI design following Material Design 3 principles.
func generateSurfaceColors(palette *CategorisedPalette, bg, fg CategorisedColour, theme ThemeType, hintsApplied map[Role]bool) {
	// Priority 1: Core surface colors.
	generatePriority1SurfaceColors(palette, bg, fg, theme, hintsApplied)

	// Priority 2: Surface variants, border variants, and on-colors.
	generatePriority2Colors(palette, bg, fg, theme, hintsApplied)

	// Priority 3: Inverse colors, scrim/shadow, container variants.
	generatePriority3Colors(palette, bg, fg, theme, hintsApplied)
}

// generatePriority1SurfaceColors generates essential surface colors.
func generatePriority1SurfaceColors(palette *CategorisedPalette, bg, fg CategorisedColour, theme ThemeType, hintsApplied map[Role]bool) {
	// Generate Surface (if not provided via hints).
	if !hintsApplied[RoleSurface] {
		surface := generateSurface(bg, theme)
		palette.Set(RoleSurface, surface)
	}

	// Generate OnSurface (if not provided via hints).
	if !hintsApplied[RoleOnSurface] {
		var surface CategorisedColour
		if s, ok := palette.Get(RoleSurface); ok {
			surface = s
		} else {
			surface = bg
		}
		onSurface := generateOnSurface(surface, fg, theme)
		palette.Set(RoleOnSurface, onSurface)
	}

	// Generate Outline (if not provided via hints).
	if !hintsApplied[RoleOutline] {
		var surface CategorisedColour
		if s, ok := palette.Get(RoleSurface); ok {
			surface = s
		} else {
			surface = bg
		}
		outline := generateOutline(surface, theme)
		palette.Set(RoleOutline, outline)
	}

	// Generate Border (if not provided via hints).
	if !hintsApplied[RoleBorder] {
		var surface CategorisedColour
		if s, ok := palette.Get(RoleSurface); ok {
			surface = s
		} else {
			surface = bg
		}
		border := generateBorder(surface, theme)
		palette.Set(RoleBorder, border)
	}
}

// luminanceOffsetParams describes a colour derivation that shifts the source
// luminance (theme-dependent) and scales its saturation. In dark themes the
// luminance is raised by lumDelta and capped at darkLumCap; in light themes
// it is lowered by lumDelta and floored at lightLumFloor.
type luminanceOffsetParams struct {
	role          Role
	lumDelta      float64
	darkLumCap    float64
	lightLumFloor float64
	satFactor     float64
	satFloor      float64
}

// generateLuminanceOffset returns a CategorisedColour derived from source by
// applying a theme-aware luminance shift and saturation scale per p. Used for
// surface variants where the relationship to the source colour is uniform but
// the magnitudes differ (surface vs outline vs border).
func generateLuminanceOffset(source CategorisedColour, theme ThemeType, p luminanceOffsetParams) CategorisedColour {
	h, s, l := rgbToHSL(source.RGB)

	var newL float64
	if theme == ThemeDark {
		newL = l + p.lumDelta
		if newL > p.darkLumCap {
			newL = p.darkLumCap
		}
	} else {
		newL = l - p.lumDelta
		if newL < p.lightLumFloor {
			newL = p.lightLumFloor
		}
	}

	newS := s * p.satFactor
	if newS < p.satFloor {
		newS = p.satFloor
	}

	newRGB := HSLToRGB(h, newS, newL)
	c := newGeneratedColour(p.role, newRGB, newL)
	c.Hue = h
	c.Saturation = newS
	return c
}

// generateSurface creates a surface color slightly different from background.
// Material Design 3 uses tonal elevation - surfaces are slightly lighter/darker than background.
func generateSurface(bg CategorisedColour, theme ThemeType) CategorisedColour {
	return generateLuminanceOffset(bg, theme, luminanceOffsetParams{
		role:          RoleSurface,
		lumDelta:      0.06,
		darkLumCap:    0.25,
		lightLumFloor: 0.75,
		satFactor:     0.9,
		satFloor:      0.05,
	})
}

// generateOnSurface creates a high-contrast text color for surface.
// Typically same as foreground, but can be adjusted if surface differs significantly.
func generateOnSurface(surface, fg CategorisedColour, theme ThemeType) CategorisedColour {
	// Check if foreground has adequate contrast with surface.
	fgColor := fg.Colour
	surfaceColor := surface.Colour
	contrast := ContrastRatio(fgColor, surfaceColor)

	// If foreground works well on surface, use it.
	if contrast >= 4.5 {
		return CategorisedColour{
			Colour:      fg.Colour,
			Role:        RoleOnSurface,
			Hex:         fg.Hex,
			RGB:         fg.RGB,
			RGBA:        fg.RGBA,
			Luminance:   fg.Luminance,
			IsLight:     fg.IsLight,
			Hue:         fg.Hue,
			Saturation:  fg.Saturation,
			IsGenerated: true,
			Weight:      0,
		}
	}

	// Otherwise, generate a color with better contrast.
	rgb := fg.RGB
	h, s, l := rgbToHSL(rgb)

	// Adjust luminance to ensure 7:1 contrast (AAA).
	minContrast := 7.0
	var newRGB RGB
	_, newRGB = adjustLuminanceForContrast(h, s, l, surfaceColor, minContrast, theme, 20)

	c := newGeneratedColour(RoleOnSurface, newRGB, Luminance(RGBToColor(newRGB)))
	c.Hue = h
	c.Saturation = s
	return c
}

// generateOutline creates a desaturated border color with moderate contrast.
// Used for dividers, borders, and outlines.
func generateOutline(surface CategorisedColour, theme ThemeType) CategorisedColour {
	return generateLuminanceOffset(surface, theme, luminanceOffsetParams{
		role:          RoleOutline,
		lumDelta:      0.18,
		darkLumCap:    0.35,
		lightLumFloor: 0.60,
		satFactor:     0.3,
		satFloor:      0.02,
	})
}

// generateBorder creates a border color similar to outline but slightly more prominent.
// Used for primary borders, focus indicators.
func generateBorder(surface CategorisedColour, theme ThemeType) CategorisedColour {
	return generateLuminanceOffset(surface, theme, luminanceOffsetParams{
		role:          RoleBorder,
		lumDelta:      0.28,
		darkLumCap:    0.45,
		lightLumFloor: 0.50,
		satFactor:     0.4,
		satFloor:      0.03,
	})
}

// generatePriority2Colors generates surface/border variants and on-colors.
func generatePriority2Colors(palette *CategorisedPalette, bg, fg CategorisedColour, theme ThemeType, hintsApplied map[Role]bool) {
	surface, hasSurface := palette.Get(RoleSurface)
	if !hasSurface {
		surface = bg
	}

	// Surface variant: between surface and background.
	if !hintsApplied[RoleSurfaceVariant] {
		surfaceVariant := generateSurfaceVariant(surface, bg, theme)
		palette.Set(RoleSurfaceVariant, surfaceVariant)
	}

	// OnSurface variant.
	if !hintsApplied[RoleOnSurfaceVariant] {
		surfaceVariant, hasVariant := palette.Get(RoleSurfaceVariant)
		if !hasVariant {
			surfaceVariant = surface
		}
		onSurfaceVariant := generateOnSurface(surfaceVariant, fg, theme)
		onSurfaceVariant.Role = RoleOnSurfaceVariant
		palette.Set(RoleOnSurfaceVariant, onSurfaceVariant)
	}

	// Border muted.
	if !hintsApplied[RoleBorderMuted] {
		borderMuted := generateBorderMuted(surface, theme)
		palette.Set(RoleBorderMuted, borderMuted)
	}

	// Outline variant.
	if !hintsApplied[RoleOutlineVariant] {
		outlineVariant := generateOutlineVariant(surface, theme)
		palette.Set(RoleOutlineVariant, outlineVariant)
	}

	// Generate on-colors for accents.
	generateOnColors(palette, theme, hintsApplied)
}

// generatePriority3Colors generates inverse colors, scrim/shadow, and container variants.
func generatePriority3Colors(palette *CategorisedPalette, bg, fg CategorisedColour, theme ThemeType, hintsApplied map[Role]bool) {
	// Inverse colors.
	if !hintsApplied[RoleInverseSurface] {
		inverseSurface := generateInverseSurface(bg, theme)
		palette.Set(RoleInverseSurface, inverseSurface)
	}

	if !hintsApplied[RoleInverseOnSurface] {
		inverseSurface, hasInverse := palette.Get(RoleInverseSurface)
		if !hasInverse {
			// Generate from background.
			inverseSurface = generateInverseSurface(bg, theme)
		}
		inverseOnSurface := generateInverseOnSurface(inverseSurface, theme)
		palette.Set(RoleInverseOnSurface, inverseOnSurface)
	}

	if !hintsApplied[RoleInversePrimary] {
		// Use accent1 or fallback to foreground.
		var primary CategorisedColour
		if a1, ok := palette.Get(RoleAccent1); ok {
			primary = a1
		} else {
			primary = fg
		}
		inversePrimary := generateInversePrimary(primary, bg, theme)
		palette.Set(RoleInversePrimary, inversePrimary)
	}

	// Scrim and shadow with alpha.
	if !hintsApplied[RoleScrim] {
		scrim := generateScrim(theme)
		palette.Set(RoleScrim, scrim)
	}

	if !hintsApplied[RoleShadow] {
		shadow := generateShadow(theme)
		palette.Set(RoleShadow, shadow)
	}

	// Container elevation variants.
	generateContainerVariants(palette, bg, theme, hintsApplied)
}

// generateSurfaceVariant creates an intermediate color between surface and background.
func generateSurfaceVariant(surface, bg CategorisedColour, _ ThemeType) CategorisedColour {
	surfaceRGB := surface.RGB
	bgRGB := bg.RGB

	// Blend 60% surface + 40% background.
	blendedR := uint8(float64(surfaceRGB.R)*0.6 + float64(bgRGB.R)*0.4)
	blendedG := uint8(float64(surfaceRGB.G)*0.6 + float64(bgRGB.G)*0.4)
	blendedB := uint8(float64(surfaceRGB.B)*0.6 + float64(bgRGB.B)*0.4)

	newRGB := RGB{R: blendedR, G: blendedG, B: blendedB}
	newColor := RGBToColor(newRGB)

	return CategorisedColour{
		Colour:      newColor,
		Role:        RoleSurfaceVariant,
		Hex:         newRGB.Hex(),
		RGB:         newRGB,
		RGBA:        RGBToRGBA(newRGB),
		Luminance:   Luminance(newColor),
		IsLight:     Luminance(newColor) > 0.5,
		IsGenerated: true,
		Weight:      0,
	}
}

// generateBorderMuted creates a more subtle border than primary border.
func generateBorderMuted(surface CategorisedColour, theme ThemeType) CategorisedColour {
	rgb := surface.RGB
	h, s, l := rgbToHSL(rgb)

	// More subtle than border: 15-20% luminance difference.
	var newL float64
	if theme == ThemeDark {
		newL = l + 0.18
	} else {
		newL = l - 0.18
	}

	// Very low saturation.
	newS := s * 0.25

	newRGB := HSLToRGB(h, newS, newL)
	return newGeneratedColour(RoleBorderMuted, newRGB, newL)
}

// generateOutlineVariant creates a secondary outline color.
func generateOutlineVariant(surface CategorisedColour, theme ThemeType) CategorisedColour {
	rgb := surface.RGB
	h, s, l := rgbToHSL(rgb)

	// Between outline and surface: 10-12% luminance difference.
	var newL float64
	if theme == ThemeDark {
		newL = l + 0.11
	} else {
		newL = l - 0.11
	}

	newS := s * 0.2
	newRGB := HSLToRGB(h, newS, newL)
	return newGeneratedColour(RoleOutlineVariant, newRGB, newL)
}

// generateOnColors generates high-contrast text colors for all accent and semantic colors.
func generateOnColors(palette *CategorisedPalette, _ ThemeType, hintsApplied map[Role]bool) {
	// On-colors for accents.
	generateOnColor(palette, RoleAccent1, RoleOnAccent1, hintsApplied)
	generateOnColor(palette, RoleAccent2, RoleOnAccent2, hintsApplied)
	generateOnColor(palette, RoleAccent3, RoleOnAccent3, hintsApplied)
	generateOnColor(palette, RoleAccent4, RoleOnAccent4, hintsApplied)

	// On-colors for semantic roles.
	generateOnColor(palette, RoleDanger, RoleOnDanger, hintsApplied)
	generateOnColor(palette, RoleWarning, RoleOnWarning, hintsApplied)
	generateOnColor(palette, RoleSuccess, RoleOnSuccess, hintsApplied)
	generateOnColor(palette, RoleInfo, RoleOnInfo, hintsApplied)
}

// generateOnColor generates a high-contrast "on" color for a given background role.
func generateOnColor(palette *CategorisedPalette, bgRole, onRole Role, hintsApplied map[Role]bool) {
	if hintsApplied[onRole] {
		return
	}

	bgColor, hasBg := palette.Get(bgRole)
	if !hasBg {
		return
	}

	// Test white vs black for better contrast.
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	whiteContrast := ContrastRatio(white, bgColor.Colour)
	blackContrast := ContrastRatio(black, bgColor.Colour)

	var onColor color.Color
	var onRGB RGB
	if whiteContrast > blackContrast {
		onColor = white
		onRGB = RGB{R: 255, G: 255, B: 255}
	} else {
		onColor = black
		onRGB = RGB{R: 0, G: 0, B: 0}
	}

	palette.Set(onRole, CategorisedColour{
		Colour:      onColor,
		Role:        onRole,
		Hex:         onRGB.Hex(),
		RGB:         onRGB,
		RGBA:        RGBToRGBA(onRGB),
		Luminance:   Luminance(onColor),
		IsLight:     Luminance(onColor) > 0.5,
		IsGenerated: true,
		Weight:      0,
	})
}

// fixedLuminanceParams describes a colour derivation that pins the resulting
// luminance to a theme-specific constant while preserving source hue.
// satFactor scales the source saturation (1.0 preserves it; <1 desaturates).
type fixedLuminanceParams struct {
	role      Role
	darkL     float64
	lightL    float64
	satFactor float64
}

// generateFixedLuminance returns a CategorisedColour with luminance pinned per
// theme. Used for inverse-surface variants where the hue carries over from the
// source but the luminance is independent of it.
func generateFixedLuminance(source CategorisedColour, theme ThemeType, p fixedLuminanceParams) CategorisedColour {
	h, s, _ := rgbToHSL(source.RGB)
	newL := p.darkL
	if theme != ThemeDark {
		newL = p.lightL
	}
	newRGB := HSLToRGB(h, s*p.satFactor, newL)
	return newGeneratedColour(p.role, newRGB, newL)
}

// generateInverseSurface creates an inverse surface color (opposite theme).
func generateInverseSurface(bg CategorisedColour, theme ThemeType) CategorisedColour {
	return generateFixedLuminance(bg, theme, fixedLuminanceParams{
		role:      RoleInverseSurface,
		darkL:     0.90,
		lightL:    0.15,
		satFactor: 1.0,
	})
}

// generateInverseOnSurface creates text color for inverse surface.
func generateInverseOnSurface(inverseSurface CategorisedColour, theme ThemeType) CategorisedColour {
	return generateFixedLuminance(inverseSurface, theme, fixedLuminanceParams{
		role:      RoleInverseOnSurface,
		darkL:     0.10,
		lightL:    0.95,
		satFactor: 0.1,
	})
}

// generateInversePrimary creates an inverse accent color.
func generateInversePrimary(primary, _ CategorisedColour, theme ThemeType) CategorisedColour {
	return generateFixedLuminance(primary, theme, fixedLuminanceParams{
		role:      RoleInversePrimary,
		darkL:     0.40,
		lightL:    0.70,
		satFactor: 1.0,
	})
}

// generateScrim creates a dark overlay color with alpha for modals.
func generateScrim(_ ThemeType) CategorisedColour {
	// Scrim is always dark, but alpha varies.
	rgb := RGB{R: 0, G: 0, B: 0}
	rgba := RGBA{R: 0, G: 0, B: 0, A: 82} // ~32% opacity

	newColor := RGBToColor(rgb)

	return CategorisedColour{
		Colour:      newColor,
		Role:        RoleScrim,
		Hex:         rgb.Hex(),
		RGB:         rgb,
		RGBA:        rgba,
		Luminance:   0.0,
		IsLight:     false,
		IsGenerated: true,
		Weight:      0,
	}
}

// generateShadow creates a shadow color with alpha for elevation.
func generateShadow(_ ThemeType) CategorisedColour {
	// Shadow is always black with low alpha.
	rgb := RGB{R: 0, G: 0, B: 0}
	rgba := RGBA{R: 0, G: 0, B: 0, A: 38} // ~15% opacity

	newColor := RGBToColor(rgb)

	return CategorisedColour{
		Colour:      newColor,
		Role:        RoleShadow,
		Hex:         rgb.Hex(),
		RGB:         rgb,
		RGBA:        rgba,
		Luminance:   0.0,
		IsLight:     false,
		IsGenerated: true,
		Weight:      0,
	}
}

// generateContainerVariants creates elevation-based container colors.
func generateContainerVariants(palette *CategorisedPalette, bg CategorisedColour, theme ThemeType, hintsApplied map[Role]bool) {
	surface, hasSurface := palette.Get(RoleSurface)
	if !hasSurface {
		surface = bg
	}

	rgb := surface.RGB
	h, s, l := rgbToHSL(rgb)

	// Generate 5 elevation levels with progressive luminance.
	containerRoles := []struct {
		role   Role
		offset float64 // Luminance offset from surface
	}{
		{RoleSurfaceContainerLowest, -0.04}, // Lowest elevation (closer to background)
		{RoleSurfaceContainerLow, -0.02},    // Low elevation
		{RoleSurfaceContainer, 0.0},         // Default (same as surface)
		{RoleSurfaceContainerHigh, 0.02},    // High elevation
		{RoleSurfaceContainerHighest, 0.04}, // Highest elevation
	}

	for _, container := range containerRoles {
		if hintsApplied[container.role] {
			continue
		}

		newL := l
		if theme == ThemeDark {
			newL += container.offset
		} else {
			newL -= container.offset
		}

		// Clamp luminance.
		if newL < 0.05 {
			newL = 0.05
		} else if newL > 0.95 {
			newL = 0.95
		}

		newRGB := HSLToRGB(h, s, newL)
		newColor := RGBToColor(newRGB)

		palette.Set(container.role, CategorisedColour{
			Colour:      newColor,
			Role:        container.role,
			Hex:         newRGB.Hex(),
			RGB:         newRGB,
			RGBA:        RGBToRGBA(newRGB),
			Luminance:   Luminance(newColor),
			IsLight:     newL > 0.5,
			IsGenerated: true,
			Weight:      0,
		})
	}
}
