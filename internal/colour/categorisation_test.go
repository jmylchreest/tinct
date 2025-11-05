// Package colour provides colour extraction and palette generation functionality.
package colour

import (
	"image/color"
	"math"
	"testing"
)

func TestSemanticColourEnhancement(t *testing.T) {
	tests := []struct {
		name            string
		inputColor      color.Color
		role            Role
		themeType       ThemeType
		wantMinSat      float64
		wantMinContrast float64
	}{
		{
			name:            "Danger color in dark theme",
			inputColor:      color.RGBA{R: 180, G: 50, B: 50, A: 255}, // Dull red
			role:            RoleDanger,
			themeType:       ThemeDark,
			wantMinSat:      MinSemanticSaturation,
			wantMinContrast: 3.0,
		},
		{
			name:            "Success color in light theme",
			inputColor:      color.RGBA{R: 100, G: 150, B: 100, A: 255}, // Dull green
			role:            RoleSuccess,
			themeType:       ThemeLight,
			wantMinSat:      MinSemanticSaturation,
			wantMinContrast: 3.0,
		},
		{
			name:            "Warning color in dark theme",
			inputColor:      color.RGBA{R: 150, G: 130, B: 80, A: 255}, // Dull orange
			role:            RoleWarning,
			themeType:       ThemeDark,
			wantMinSat:      MinSemanticSaturation,
			wantMinContrast: 3.0,
		},
		{
			name:            "Info color in light theme",
			inputColor:      color.RGBA{R: 100, G: 120, B: 180, A: 255}, // Dull blue
			role:            RoleInfo,
			themeType:       ThemeLight,
			wantMinSat:      MinSemanticSaturation,
			wantMinContrast: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a background color based on theme.
			var bgColor color.Color
			if tt.themeType == ThemeDark {
				bgColor = color.RGBA{R: 20, G: 20, B: 30, A: 255} // Dark background
			} else {
				bgColor = color.RGBA{R: 240, G: 240, B: 245, A: 255} // Light background
			}

			bgCategorised := CategorisedColour{
				Colour:    bgColor,
				RGB:       ToRGB(bgColor),
				Luminance: Luminance(bgColor),
			}

			// Create input color.
			inputRGB := ToRGB(tt.inputColor)
			h, s, _ := rgbToHSL(inputRGB)
			inputCategorised := CategorisedColour{
				Colour:     tt.inputColor,
				RGB:        inputRGB,
				Luminance:  Luminance(tt.inputColor),
				Hue:        h,
				Saturation: s,
			}

			// Enhance the color.
			enhanced := enhanceSemanticColour(inputCategorised, tt.role, tt.themeType, true, bgCategorised)

			// Check saturation is boosted.
			if enhanced.Saturation < tt.wantMinSat {
				t.Errorf("Saturation too low: got %.2f, want >= %.2f", enhanced.Saturation, tt.wantMinSat)
			}

			// Check contrast with background.
			contrast := ContrastRatio(enhanced.Colour, bgColor)
			if contrast < tt.wantMinContrast {
				t.Errorf("Contrast too low: got %.2f:1, want >= %.2f:1", contrast, tt.wantMinContrast)
			}

			// Check lightness is within reasonable bounds for semantic colors.
			// Note: WCAG luminance can be lower than HSL lightness due to gamma correction.
			if enhanced.Luminance < 0.10 || enhanced.Luminance > 0.85 {
				t.Errorf("Luminance out of reasonable bounds: got %.2f", enhanced.Luminance)
			}

			t.Logf("Enhanced %s: sat=%.2f, lum=%.2f, contrast=%.2f:1, hex=%s",
				tt.role, enhanced.Saturation, enhanced.Luminance, contrast, enhanced.Hex)
		})
	}
}

func TestGenerateFallbackSemanticColours(t *testing.T) {
	tests := []struct {
		name            string
		role            Role
		themeType       ThemeType
		wantHue         float64
		wantMinSat      float64
		wantMinContrast float64
	}{
		{
			name:            "Danger fallback dark theme",
			role:            RoleDanger,
			themeType:       ThemeDark,
			wantHue:         0,
			wantMinSat:      0.7,
			wantMinContrast: 3.0,
		},
		{
			name:            "Warning fallback light theme",
			role:            RoleWarning,
			themeType:       ThemeLight,
			wantHue:         45,
			wantMinSat:      0.7,
			wantMinContrast: 3.0,
		},
		{
			name:            "Success fallback dark theme",
			role:            RoleSuccess,
			themeType:       ThemeDark,
			wantHue:         120,
			wantMinSat:      0.7,
			wantMinContrast: 3.0,
		},
		{
			name:            "Info fallback light theme",
			role:            RoleInfo,
			themeType:       ThemeLight,
			wantHue:         210,
			wantMinSat:      0.7,
			wantMinContrast: 3.0,
		},
		{
			name:            "Notification fallback dark theme",
			role:            RoleNotification,
			themeType:       ThemeDark,
			wantHue:         285,
			wantMinSat:      0.7,
			wantMinContrast: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a background color based on theme.
			var bgColor color.Color
			if tt.themeType == ThemeDark {
				bgColor = color.RGBA{R: 20, G: 20, B: 30, A: 255}
			} else {
				bgColor = color.RGBA{R: 240, G: 240, B: 245, A: 255}
			}

			bgCategorised := CategorisedColour{
				Colour:    bgColor,
				RGB:       ToRGB(bgColor),
				Luminance: Luminance(bgColor),
			}

			// Generate fallback color.
			fallback := generateFallbackSemanticColour(tt.role, tt.themeType, true, bgCategorised)

			// Check hue is correct.
			hueDiff := math.Abs(fallback.Hue - tt.wantHue)
			if hueDiff > 5.0 { // Allow 5 degree tolerance
				t.Errorf("Hue mismatch: got %.1f, want %.1f (±5)", fallback.Hue, tt.wantHue)
			}

			// Check saturation.
			if fallback.Saturation < tt.wantMinSat {
				t.Errorf("Saturation too low: got %.2f, want >= %.2f", fallback.Saturation, tt.wantMinSat)
			}

			// Check contrast.
			contrast := ContrastRatio(fallback.Colour, bgColor)
			if contrast < tt.wantMinContrast {
				t.Errorf("Contrast too low: got %.2f:1, want >= %.2f:1", contrast, tt.wantMinContrast)
			}

			t.Logf("Fallback %s: hue=%.1f, sat=%.2f, lum=%.2f, contrast=%.2f:1, hex=%s",
				tt.role, fallback.Hue, fallback.Saturation, fallback.Luminance, contrast, fallback.Hex)
		})
	}
}

func TestCategoriseWithSemanticFallbacks(t *testing.T) {
	tests := []struct {
		name      string
		colors    []color.Color
		themeType ThemeType
		wantRoles []Role
	}{
		{
			name: "Monochrome palette generates all semantic colors",
			colors: []color.Color{
				color.RGBA{R: 30, G: 30, B: 40, A: 255},    // Dark background
				color.RGBA{R: 200, G: 200, B: 210, A: 255}, // Light foreground
				color.RGBA{R: 100, G: 100, B: 110, A: 255}, // Mid grey
			},
			themeType: ThemeDark,
			wantRoles: []Role{
				RoleBackground,
				RoleForeground,
				RoleDanger,
				RoleWarning,
				RoleSuccess,
				RoleInfo,
				RoleNotification,
			},
		},
		{
			name: "Colorful palette enhances semantic colors",
			colors: []color.Color{
				color.RGBA{R: 240, G: 240, B: 245, A: 255}, // Light background
				color.RGBA{R: 30, G: 30, B: 40, A: 255},    // Dark foreground
				color.RGBA{R: 180, G: 80, B: 80, A: 255},   // Dull red
				color.RGBA{R: 80, G: 150, B: 80, A: 255},   // Dull green
				color.RGBA{R: 80, G: 100, B: 180, A: 255},  // Dull blue
				color.RGBA{R: 150, G: 120, B: 80, A: 255},  // Dull orange
			},
			themeType: ThemeLight,
			wantRoles: []Role{
				RoleBackground,
				RoleForeground,
				RoleDanger,
				RoleWarning,
				RoleSuccess,
				RoleInfo,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create palette.
			palette := &Palette{Colors: tt.colors}

			// Categorise with default config.
			config := DefaultCategorisationConfig()
			config.ThemeType = tt.themeType

			categorised := Categorise(palette, config)

			// Check that all expected roles are present.
			for _, role := range tt.wantRoles {
				if _, ok := categorised.Get(role); !ok {
					t.Errorf("Missing role: %s", role)
				} else {
					cc, _ := categorised.Get(role)
					t.Logf("Role %s: hex=%s, sat=%.2f, lum=%.2f",
						role, cc.Hex, cc.Saturation, cc.Luminance)
				}
			}

			// Verify semantic colors have good saturation.
			semanticRoles := []Role{
				RoleDanger, RoleWarning, RoleSuccess, RoleInfo, RoleNotification,
			}

			for _, role := range semanticRoles {
				if cc, ok := categorised.Get(role); ok {
					if cc.Saturation < 0.5 {
						t.Errorf("Semantic color %s has low saturation: %.2f", role, cc.Saturation)
					}

					// Check contrast with background.
					if bg, bgOk := categorised.Get(RoleBackground); bgOk {
						contrast := ContrastRatio(cc.Colour, bg.Colour)
						if contrast < 2.5 {
							t.Errorf("Semantic color %s has poor contrast with background: %.2f:1", role, contrast)
						}
					}
				}
			}
		})
	}
}

func TestSemanticColourDistinctness(t *testing.T) {
	// Create a monochrome palette.
	colors := []color.Color{
		color.RGBA{R: 20, G: 20, B: 30, A: 255},
		color.RGBA{R: 220, G: 220, B: 230, A: 255},
	}

	palette := &Palette{Colors: colors}
	config := DefaultCategorisationConfig()
	config.ThemeType = ThemeDark

	categorised := Categorise(palette, config)

	// Get all semantic colors.
	semanticRoles := []Role{
		RoleDanger, RoleWarning, RoleSuccess, RoleInfo, RoleNotification,
	}

	semanticColors := make(map[Role]CategorisedColour)
	for _, role := range semanticRoles {
		if cc, ok := categorised.Get(role); ok {
			semanticColors[role] = cc
		}
	}

	// Check that semantic colors are distinct from each other.
	roles := make([]Role, 0, len(semanticColors))
	for role := range semanticColors {
		roles = append(roles, role)
	}

	for i := 0; i < len(roles); i++ {
		for j := i + 1; j < len(roles); j++ {
			role1 := roles[i]
			role2 := roles[j]
			cc1 := semanticColors[role1]
			cc2 := semanticColors[role2]

			// Check hue difference.
			hueDiff := math.Abs(cc1.Hue - cc2.Hue)
			// Handle wrap-around (e.g., 350° vs 10°).
			if hueDiff > 180 {
				hueDiff = 360 - hueDiff
			}

			if hueDiff < 30 {
				t.Errorf("Colors %s and %s are too similar: hue diff = %.1f°",
					role1, role2, hueDiff)
			}

			t.Logf("%s (%.1f°, %s) vs %s (%.1f°, %s): hue diff = %.1f°",
				role1, cc1.Hue, cc1.Hex, role2, cc2.Hue, cc2.Hex, hueDiff)
		}
	}
}

func TestSemanticHueRanges(t *testing.T) {
	// Test that semantic hues are correctly defined.
	expectedRanges := map[Role]struct {
		minHue float64
		maxHue float64
	}{
		RoleDanger:       {minHue: 350, maxHue: 20},  // Red (wraps around)
		RoleWarning:      {minHue: 30, maxHue: 60},   // Orange
		RoleSuccess:      {minHue: 100, maxHue: 140}, // Green
		RoleInfo:         {minHue: 190, maxHue: 230}, // Blue
		RoleNotification: {minHue: 270, maxHue: 300}, // Purple
	}

	for role, hue := range SemanticHues {
		expected, ok := expectedRanges[role]
		if !ok {
			t.Errorf("Unexpected semantic role: %s", role)
			continue
		}

		// Check if hue is within expected range.
		inRange := false
		if expected.minHue > expected.maxHue {
			// Wraps around 360.
			inRange = hue >= expected.minHue || hue <= expected.maxHue
		} else {
			inRange = hue >= expected.minHue && hue <= expected.maxHue
		}

		if !inRange {
			t.Errorf("Role %s has hue %.1f° outside expected range [%.1f°, %.1f°]",
				role, hue, expected.minHue, expected.maxHue)
		}

		t.Logf("Role %s: hue = %.1f° (expected range: %.1f° - %.1f°)",
			role, hue, expected.minHue, expected.maxHue)
	}
}

func TestThemeAwareLightness(t *testing.T) {
	colors := []color.Color{
		color.RGBA{R: 20, G: 20, B: 30, A: 255},
		color.RGBA{R: 220, G: 220, B: 230, A: 255},
	}

	tests := []struct {
		name             string
		themeType        ThemeType
		wantMinLuminance float64
		wantMaxLuminance float64
	}{
		{
			name:             "Dark theme semantic colors",
			themeType:        ThemeDark,
			wantMinLuminance: 0.15, // WCAG luminance is lower due to gamma correction
			wantMaxLuminance: 0.80,
		},
		{
			name:             "Light theme semantic colors",
			themeType:        ThemeLight,
			wantMinLuminance: 0.08, // WCAG luminance for dark colors on light bg
			wantMaxLuminance: 0.65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			palette := &Palette{Colors: colors}
			config := DefaultCategorisationConfig()
			config.ThemeType = tt.themeType

			categorised := Categorise(palette, config)

			semanticRoles := []Role{
				RoleDanger, RoleWarning, RoleSuccess, RoleInfo, RoleNotification,
			}

			for _, role := range semanticRoles {
				if cc, ok := categorised.Get(role); ok {
					if cc.Luminance < tt.wantMinLuminance || cc.Luminance > tt.wantMaxLuminance {
						t.Errorf("Semantic color %s luminance %.2f outside range [%.2f, %.2f]",
							role, cc.Luminance, tt.wantMinLuminance, tt.wantMaxLuminance)
					}
					t.Logf("Role %s: luminance = %.2f, hex = %s", role, cc.Luminance, cc.Hex)
				}
			}
		})
	}
}
