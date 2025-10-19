// Package color provides colour extraction and palette generation functionality.
package colour

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
)

// Luminance calculates the relative luminance of a colour according to WCAG 2.0.
// Returns a value between 0 (darkest) and 1 (lightest).
// https://www.w3.org/TR/WCAG20/#relativeluminancedef
func Luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Convert from 16-bit to 8-bit
	rf := float64(r>>8) / 255.0
	rg := float64(g>>8) / 255.0
	rb := float64(b>>8) / 255.0

	// Apply gamma correction
	rf = gammaCorrect(rf)
	rg = gammaCorrect(rg)
	rb = gammaCorrect(rb)

	// Calculate luminance using WCAG formula
	return 0.2126*rf + 0.7152*rg + 0.0722*rb
}

// gammaCorrect applies gamma correction to a colour component.
func gammaCorrect(v float64) float64 {
	if v <= 0.03928 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// ContrastRatio calculates the contrast ratio between two colours according to WCAG 2.0.
// Returns a value between 1 and 21, where 21 is maximum contrast (black vs white).
// https://www.w3.org/TR/WCAG20/#contrast-ratiodef
func ContrastRatio(c1, c2 color.Color) float64 {
	l1 := Luminance(c1)
	l2 := Luminance(c2)

	// Ensure l1 is the lighter colour
	if l1 < l2 {
		l1, l2 = l2, l1
	}

	return (l1 + 0.05) / (l2 + 0.05)
}

// ColourRole represents the semantic role of a colour in a theme.
type ColourRole string

const (
	// Core roles
	RoleBackground      ColourRole = "background"
	RoleBackgroundMuted ColourRole = "background-muted"
	RoleForeground      ColourRole = "foreground"
	RoleForegroundMuted ColourRole = "foreground-muted"

	// Accent roles
	RoleAccent1 ColourRole = "accent-1"
	RoleAccent2 ColourRole = "accent-2"
	RoleAccent3 ColourRole = "accent-3"
	RoleAccent4 ColourRole = "accent-4"

	// Semantic roles
	RoleDanger       ColourRole = "danger"
	RoleWarning      ColourRole = "warning"
	RoleSuccess      ColourRole = "success"
	RoleInfo         ColourRole = "info"
	RoleNotification ColourRole = "notification"
)

// CategorisedColour represents a colour with its assigned role and metadata.
type CategorisedColour struct {
	Colour     color.Color `json:"-"`
	Role       ColourRole  `json:"role"`
	Hex        string      `json:"hex"`
	RGB        RGB         `json:"rgb"`
	Luminance  float64     `json:"luminance"`
	IsLight    bool        `json:"is_light"`
	Hue        float64     `json:"hue,omitempty"`        // HSL hue (0-360)
	Saturation float64     `json:"saturation,omitempty"` // HSL saturation (0-1)
}

// ThemeType represents whether a theme is light-on-dark or dark-on-light.
type ThemeType int

const (
	// ThemeAuto automatically detects the best theme type based on average luminance.
	ThemeAuto ThemeType = iota
	// ThemeDark is a dark theme (light text on dark background).
	ThemeDark
	// ThemeLight is a light theme (dark text on light background).
	ThemeLight
)

// String returns the string representation of a ThemeType.
func (t ThemeType) String() string {
	switch t {
	case ThemeDark:
		return "dark"
	case ThemeLight:
		return "light"
	case ThemeAuto:
		return "auto"
	default:
		return "unknown"
	}
}

// CategorisationConfig holds configuration for colour categorisation.
type CategorisationConfig struct {
	ThemeType             ThemeType
	MinContrastRatio      float64 // Minimum contrast between foreground and background
	RequireAAA            bool    // Require AAA contrast (7:1) instead of AA (4.5:1)
	MutedLuminanceAdjust  float64 // How much to adjust luminance for muted variants (0.0-1.0)
	EnhanceSemanticColors bool    // Boost saturation and adjust lightness for semantic colors
	SemanticBoostAmount   float64 // How much to boost semantic saturation (0.0-1.0)
}

// Semantic color constants
const (
	MinSemanticSaturation = 0.6  // Minimum saturation for semantic colors
	MinSemanticLightness  = 0.35 // Minimum lightness for semantic colors
	MaxSemanticLightness  = 0.65 // Maximum lightness for semantic colors
)

// SemanticHues defines the standard hue values for semantic colors
var SemanticHues = map[ColourRole]float64{
	RoleDanger:       0,   // Red
	RoleWarning:      45,  // Orange
	RoleSuccess:      120, // Green
	RoleInfo:         210, // Blue
	RoleNotification: 285, // Purple
}

// DefaultCategorisationConfig returns the default categorisation configuration.
func DefaultCategorisationConfig() CategorisationConfig {
	return CategorisationConfig{
		ThemeType:             ThemeAuto,
		MinContrastRatio:      4.5, // WCAG AA standard
		RequireAAA:            false,
		MutedLuminanceAdjust:  0.15, // 15% adjustment for muted variants
		EnhanceSemanticColors: true, // Enable semantic color enhancement by default
		SemanticBoostAmount:   0.3,  // 30% saturation boost
	}
}

// CategorisedPalette represents a palette with categorised colours.
type CategorisedPalette struct {
	Colours    map[ColourRole]CategorisedColour `json:"colours"`
	ThemeType  ThemeType                        `json:"theme_type"`
	AllColours []CategorisedColour              `json:"all_colours,omitempty"`
}

// NewCategorisedPalette creates a new categorised palette.
func NewCategorisedPalette(themeType ThemeType) *CategorisedPalette {
	return &CategorisedPalette{
		Colours:   make(map[ColourRole]CategorisedColour),
		ThemeType: themeType,
	}
}

// Get returns a colour by role, if it exists.
func (cp *CategorisedPalette) Get(role ColourRole) (CategorisedColour, bool) {
	c, ok := cp.Colours[role]
	return c, ok
}

// Set sets a colour for a given role.
func (cp *CategorisedPalette) Set(role ColourRole, colour CategorisedColour) {
	colour.Role = role
	cp.Colours[role] = colour
}

// Categorise assigns roles to colours in a palette based on luminance, contrast, and hue.
func Categorise(palette *Palette, config CategorisationConfig) *CategorisedPalette {
	if palette == nil || len(palette.Colors) == 0 {
		return NewCategorisedPalette(ThemeAuto)
	}

	// Create categorised colours with metadata
	categorised := make([]CategorisedColour, len(palette.Colors))
	totalLuminance := 0.0

	for i, c := range palette.Colors {
		lum := Luminance(c)
		rgb := ToRGB(c)
		h, s, _ := rgbToHSL(rgb)

		categorised[i] = CategorisedColour{
			Colour:     c,
			Hex:        rgb.Hex(),
			RGB:        rgb,
			Luminance:  lum,
			IsLight:    lum > 0.5,
			Hue:        h,
			Saturation: s,
		}
		totalLuminance += lum
	}

	// Determine theme type if auto
	themeType := config.ThemeType
	if themeType == ThemeAuto {
		avgLuminance := totalLuminance / float64(len(categorised))
		if avgLuminance > 0.5 {
			themeType = ThemeLight
		} else {
			themeType = ThemeDark
		}
	}

	result := NewCategorisedPalette(themeType)
	result.AllColours = categorised

	// Find background colour (darkest for dark theme, lightest for light theme)
	bgIdx := findBackgroundIndex(categorised, themeType)
	bg := categorised[bgIdx]
	bg.Role = RoleBackground
	result.Set(RoleBackground, bg)

	// Create background-muted variant
	bgMuted := createMutedVariant(bg, config.MutedLuminanceAdjust, themeType, true)
	bgMuted.Role = RoleBackgroundMuted
	result.Set(RoleBackgroundMuted, bgMuted)

	// Find foreground colour (highest contrast with background)
	fgIdx := findForegroundIndex(categorised, bg, config, bgIdx)
	if fgIdx >= 0 {
		fg := categorised[fgIdx]
		fg.Role = RoleForeground
		result.Set(RoleForeground, fg)

		// Create foreground-muted variant
		fgMuted := createMutedVariant(fg, config.MutedLuminanceAdjust, themeType, false)
		fgMuted.Role = RoleForegroundMuted
		result.Set(RoleForegroundMuted, fgMuted)
	}

	// Categorise remaining colours by saturation and hue
	accents := make([]CategorisedColour, 0)
	for i, cc := range categorised {
		if i == bgIdx || i == fgIdx {
			continue
		}
		accents = append(accents, cc)
	}

	// Sort accents by saturation (most vibrant first)
	sortBySaturation(accents)

	// Assign accent roles (up to 4)
	accentRoles := []ColourRole{RoleAccent1, RoleAccent2, RoleAccent3, RoleAccent4}
	for i, role := range accentRoles {
		if i < len(accents) {
			accent := accents[i]
			accent.Role = role
			result.Set(role, accent)
		}
	}

	// Assign semantic roles based on hue
	assignSemanticRoles(result, accents)

	return result
}

// findBackgroundIndex finds the best background colour index.
func findBackgroundIndex(colours []CategorisedColour, themeType ThemeType) int {
	idx := 0
	if themeType == ThemeDark {
		// Find darkest colour
		minLum := colours[0].Luminance
		for i, cc := range colours {
			if cc.Luminance < minLum {
				minLum = cc.Luminance
				idx = i
			}
		}
	} else {
		// Find lightest colour
		maxLum := colours[0].Luminance
		for i, cc := range colours {
			if cc.Luminance > maxLum {
				maxLum = cc.Luminance
				idx = i
			}
		}
	}
	return idx
}

// findForegroundIndex finds the best foreground colour index.
func findForegroundIndex(colours []CategorisedColour, bg CategorisedColour, config CategorisationConfig, bgIdx int) int {
	fgIdx := -1
	maxContrast := 0.0
	minContrast := config.MinContrastRatio
	if config.RequireAAA {
		minContrast = 7.0
	}

	for i, cc := range colours {
		if i == bgIdx {
			continue
		}
		contrast := ContrastRatio(cc.Colour, bg.Colour)
		if contrast > maxContrast && contrast >= minContrast {
			maxContrast = contrast
			fgIdx = i
		}
	}

	// Fallback: use colour with highest contrast even if below threshold
	if fgIdx < 0 {
		maxContrast = 0.0
		for i, cc := range colours {
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

// createMutedVariant creates a muted variant of a colour by adjusting luminance.
func createMutedVariant(cc CategorisedColour, adjustment float64, themeType ThemeType, isBackground bool) CategorisedColour {
	// Adjust luminance to create muted variant
	newLum := cc.Luminance

	if isBackground {
		if themeType == ThemeDark {
			// Dark background: make slightly lighter
			newLum = math.Min(1.0, cc.Luminance+adjustment)
		} else {
			// Light background: make slightly darker
			newLum = math.Max(0.0, cc.Luminance-adjustment)
		}
	} else {
		// Foreground: reduce contrast slightly
		if themeType == ThemeDark {
			// Light text: make slightly darker
			newLum = math.Max(0.0, cc.Luminance-adjustment)
		} else {
			// Dark text: make slightly lighter
			newLum = math.Min(1.0, cc.Luminance+adjustment)
		}
	}

	// Convert back to RGB (approximate)
	h, s, _ := rgbToHSL(cc.RGB)
	newRGB := hslToRGB(h, s*0.8, newLum) // Also reduce saturation for muted effect

	return CategorisedColour{
		Colour:     color.RGBA{R: newRGB.R, G: newRGB.G, B: newRGB.B, A: 255},
		Hex:        newRGB.Hex(),
		RGB:        newRGB,
		Luminance:  newLum,
		IsLight:    newLum > 0.5,
		Hue:        h,
		Saturation: s * 0.8,
	}
}

// assignSemanticRoles assigns semantic roles (danger, warning, success, etc.) based on hue.
func assignSemanticRoles(palette *CategorisedPalette, accents []CategorisedColour) {
	// Map hue ranges to semantic roles
	// Red: 0-30, 330-360 (danger)
	// Orange/Yellow: 30-60 (warning)
	// Green: 90-150 (success)
	// Blue: 180-240 (info)
	// Cyan/Purple: 240-330 (notification)

	var danger, warning, success, info, notification *CategorisedColour

	for i := range accents {
		cc := &accents[i]
		h := cc.Hue

		// Skip low saturation colours (too grey)
		if cc.Saturation < 0.3 {
			continue
		}

		if (h >= 0 && h < 30) || h >= 330 {
			// Red - danger
			if danger == nil || cc.Saturation > danger.Saturation {
				danger = cc
			}
		} else if h >= 30 && h < 60 {
			// Orange/Yellow - warning
			if warning == nil || cc.Saturation > warning.Saturation {
				warning = cc
			}
		} else if h >= 90 && h < 150 {
			// Green - success
			if success == nil || cc.Saturation > success.Saturation {
				success = cc
			}
		} else if h >= 180 && h < 240 {
			// Blue - info
			if info == nil || cc.Saturation > info.Saturation {
				info = cc
			}
		} else if h >= 240 && h < 330 {
			// Cyan/Purple - notification
			if notification == nil || cc.Saturation > notification.Saturation {
				notification = cc
			}
		}
	}

	// Get background for theme-aware adjustments
	bg, hasBg := palette.Get(RoleBackground)
	themeType := palette.ThemeType

	// Set semantic roles with enhancement
	if danger != nil {
		enhanced := enhanceSemanticColour(*danger, RoleDanger, themeType, hasBg, bg)
		palette.Set(RoleDanger, enhanced)
	} else {
		// Generate fallback danger color if none found
		fallback := generateFallbackSemanticColour(RoleDanger, themeType, hasBg, bg)
		palette.Set(RoleDanger, fallback)
	}

	if warning != nil {
		enhanced := enhanceSemanticColour(*warning, RoleWarning, themeType, hasBg, bg)
		palette.Set(RoleWarning, enhanced)
	} else {
		fallback := generateFallbackSemanticColour(RoleWarning, themeType, hasBg, bg)
		palette.Set(RoleWarning, fallback)
	}

	if success != nil {
		enhanced := enhanceSemanticColour(*success, RoleSuccess, themeType, hasBg, bg)
		palette.Set(RoleSuccess, enhanced)
	} else {
		fallback := generateFallbackSemanticColour(RoleSuccess, themeType, hasBg, bg)
		palette.Set(RoleSuccess, fallback)
	}

	if info != nil {
		enhanced := enhanceSemanticColour(*info, RoleInfo, themeType, hasBg, bg)
		palette.Set(RoleInfo, enhanced)
	} else {
		fallback := generateFallbackSemanticColour(RoleInfo, themeType, hasBg, bg)
		palette.Set(RoleInfo, fallback)
	}

	if notification != nil {
		enhanced := enhanceSemanticColour(*notification, RoleNotification, themeType, hasBg, bg)
		palette.Set(RoleNotification, enhanced)
	} else {
		fallback := generateFallbackSemanticColour(RoleNotification, themeType, hasBg, bg)
		palette.Set(RoleNotification, fallback)
	}
}

// enhanceSemanticColour boosts saturation and adjusts lightness for better visibility.
func enhanceSemanticColour(cc CategorisedColour, role ColourRole, themeType ThemeType, hasBg bool, bg CategorisedColour) CategorisedColour {
	h, s, l := rgbToHSL(cc.RGB)

	// Boost saturation to minimum threshold
	if s < MinSemanticSaturation {
		s = MinSemanticSaturation
	}

	// Adjust lightness based on theme
	targetLightness := 0.5 // Default middle ground
	if themeType == ThemeDark {
		// Dark theme: make colors lighter for visibility
		targetLightness = 0.60
	} else {
		// Light theme: make colors darker for visibility
		targetLightness = 0.45
	}

	// Ensure within bounds
	if l < MinSemanticLightness {
		l = targetLightness
	} else if l > MaxSemanticLightness {
		l = targetLightness
	} else {
		// Blend towards target
		l = (l + targetLightness) / 2.0
	}

	// Ensure good contrast with background if available
	if hasBg {
		newRGB := hslToRGB(h, s, l)
		testColor := color.RGBA{R: newRGB.R, G: newRGB.G, B: newRGB.B, A: 255}
		contrast := ContrastRatio(testColor, bg.Colour)

		// If contrast is too low, adjust lightness
		if contrast < 3.0 {
			if themeType == ThemeDark {
				l = math.Min(0.75, l+0.15)
			} else {
				l = math.Max(0.30, l-0.15)
			}
		}
	}

	newRGB := hslToRGB(h, s, l)
	return CategorisedColour{
		Colour:     color.RGBA{R: newRGB.R, G: newRGB.G, B: newRGB.B, A: 255},
		Role:       role,
		Hex:        newRGB.Hex(),
		RGB:        newRGB,
		Luminance:  Luminance(color.RGBA{R: newRGB.R, G: newRGB.G, B: newRGB.B, A: 255}),
		IsLight:    l > 0.5,
		Hue:        h,
		Saturation: s,
	}
}

// generateFallbackSemanticColour creates a semantic color when none exists in the palette.
func generateFallbackSemanticColour(role ColourRole, themeType ThemeType, hasBg bool, bg CategorisedColour) CategorisedColour {
	// Get standard hue for this role
	hue, exists := SemanticHues[role]
	if !exists {
		hue = 0 // Fallback to red
	}

	// Set saturation and lightness based on theme
	saturation := 0.75 // Vibrant
	lightness := 0.5

	if themeType == ThemeDark {
		lightness = 0.60 // Lighter for dark backgrounds
	} else {
		lightness = 0.45 // Darker for light backgrounds
	}

	// Ensure good contrast with background if available
	if hasBg {
		newRGB := hslToRGB(hue, saturation, lightness)
		testColor := color.RGBA{R: newRGB.R, G: newRGB.G, B: newRGB.B, A: 255}
		contrast := ContrastRatio(testColor, bg.Colour)

		// Adjust lightness for better contrast
		attempts := 0
		for contrast < 3.0 && attempts < 5 {
			if themeType == ThemeDark {
				lightness = math.Min(0.80, lightness+0.1)
			} else {
				lightness = math.Max(0.25, lightness-0.1)
			}
			newRGB = hslToRGB(hue, saturation, lightness)
			testColor = color.RGBA{R: newRGB.R, G: newRGB.G, B: newRGB.B, A: 255}
			contrast = ContrastRatio(testColor, bg.Colour)
			attempts++
		}
	}

	newRGB := hslToRGB(hue, saturation, lightness)
	return CategorisedColour{
		Colour:     color.RGBA{R: newRGB.R, G: newRGB.G, B: newRGB.B, A: 255},
		Role:       role,
		Hex:        newRGB.Hex(),
		RGB:        newRGB,
		Luminance:  Luminance(color.RGBA{R: newRGB.R, G: newRGB.G, B: newRGB.B, A: 255}),
		IsLight:    lightness > 0.5,
		Hue:        hue,
		Saturation: saturation,
	}
}

// sortBySaturation sorts colours by saturation (highest first).
func sortBySaturation(colours []CategorisedColour) {
	// Simple bubble sort - good enough for small arrays
	n := len(colours)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if colours[j].Saturation < colours[j+1].Saturation {
				colours[j], colours[j+1] = colours[j+1], colours[j]
			}
		}
	}
}

// rgbToHSL converts RGB to HSL colour space.
// Returns hue (0-360), saturation (0-1), lightness (0-1).
func rgbToHSL(rgb RGB) (h, s, l float64) {
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	// Lightness
	l = (max + min) / 2.0

	// Saturation
	if delta == 0 {
		s = 0
		h = 0
		return
	}

	if l < 0.5 {
		s = delta / (max + min)
	} else {
		s = delta / (2.0 - max - min)
	}

	// Hue
	switch max {
	case r:
		h = (g - b) / delta
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/delta + 2
	case b:
		h = (r-g)/delta + 4
	}

	h *= 60
	return
}

// hslToRGB converts HSL to RGB colour space.
func hslToRGB(h, s, l float64) RGB {
	if s == 0 {
		// Achromatic (grey)
		v := uint8(l * 255)
		return RGB{R: v, G: v, B: v}
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r := hueToRGB(p, q, h+120)
	g := hueToRGB(p, q, h)
	b := hueToRGB(p, q, h-120)

	return RGB{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
	}
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	// Normalize t to 0-360 range
	for t < 0 {
		t += 360
	}
	for t >= 360 {
		t -= 360
	}

	if t < 60 {
		return p + (q-p)*t/60
	}
	if t < 180 {
		return q
	}
	if t < 240 {
		return p + (q-p)*(240-t)/60
	}
	return p
}

// ToJSON converts the categorised palette to JSON format.
func (cp *CategorisedPalette) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cp, "", "  ")
}

// String returns a human-readable string representation of the categorised palette.
func (cp *CategorisedPalette) String() string {
	return cp.StringWithPreview(false)
}

// StringWithPreview returns a string representation with optional ANSI colour previews.
func (cp *CategorisedPalette) StringWithPreview(showPreview bool) string {
	var result string
	result += fmt.Sprintf("Theme Type: %s\n\n", cp.ThemeType.String())

	// Core colours
	result += "Core Colours:\n"
	coreRoles := []ColourRole{RoleBackground, RoleBackgroundMuted, RoleForeground, RoleForegroundMuted}
	for _, role := range coreRoles {
		if cc, ok := cp.Get(role); ok {
			if showPreview {
				preview := ColourPreview(cc.RGB, 8)
				result += fmt.Sprintf("  %s  %-20s %s (luminance: %.2f)\n", preview, role, cc.Hex, cc.Luminance)
			} else {
				result += fmt.Sprintf("  %-20s %s (luminance: %.2f)\n", role, cc.Hex, cc.Luminance)
			}
		}
	}

	// Show contrast ratio for foreground
	if bg, bgOk := cp.Get(RoleBackground); bgOk {
		if fg, fgOk := cp.Get(RoleForeground); fgOk {
			contrast := ContrastRatio(fg.Colour, bg.Colour)
			result += fmt.Sprintf("  Contrast ratio: %.2f:1\n", contrast)
		}
	}

	// Accent colours
	result += "\nAccent Colours:\n"
	accentRoles := []ColourRole{RoleAccent1, RoleAccent2, RoleAccent3, RoleAccent4}
	for _, role := range accentRoles {
		if cc, ok := cp.Get(role); ok {
			if showPreview {
				preview := ColourPreview(cc.RGB, 8)
				result += fmt.Sprintf("  %s  %-20s %s (sat: %.2f)\n", preview, role, cc.Hex, cc.Saturation)
			} else {
				result += fmt.Sprintf("  %-20s %s (saturation: %.2f)\n", role, cc.Hex, cc.Saturation)
			}
		}
	}

	// Semantic colours
	semanticRoles := []ColourRole{RoleDanger, RoleWarning, RoleSuccess, RoleInfo, RoleNotification}
	hasSemanticRoles := false
	for _, role := range semanticRoles {
		if _, ok := cp.Get(role); ok {
			hasSemanticRoles = true
			break
		}
	}

	if hasSemanticRoles {
		result += "\nSemantic Colours:\n"
		for _, role := range semanticRoles {
			if cc, ok := cp.Get(role); ok {
				if showPreview {
					preview := ColourPreview(cc.RGB, 8)
					result += fmt.Sprintf("  %s  %-20s %s\n", preview, role, cc.Hex)
				} else {
					result += fmt.Sprintf("  %-20s %s\n", role, cc.Hex)
				}
			}
		}
	}

	return result
}
