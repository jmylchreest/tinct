// Package color provides color extraction and palette generation functionality.
package colour

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strings"
)

// ColourRole represents the semantic role of a colour in a theme.
type ColourRole string

const (
	// Core roles.
	RoleBackground      ColourRole = "background"
	RoleBackgroundMuted ColourRole = "backgroundMuted"
	RoleForeground      ColourRole = "foreground"
	RoleForegroundMuted ColourRole = "foregroundMuted"

	// Accent roles.
	RoleAccent1      ColourRole = "accent1"
	RoleAccent1Muted ColourRole = "accent1Muted"
	RoleAccent2      ColourRole = "accent2"
	RoleAccent2Muted ColourRole = "accent2Muted"
	RoleAccent3      ColourRole = "accent3"
	RoleAccent3Muted ColourRole = "accent3Muted"
	RoleAccent4      ColourRole = "accent4"
	RoleAccent4Muted ColourRole = "accent4Muted"

	// Semantic roles.
	RoleDanger       ColourRole = "danger"
	RoleWarning      ColourRole = "warning"
	RoleSuccess      ColourRole = "success"
	RoleInfo         ColourRole = "info"
	RoleNotification ColourRole = "notification"

	// Surface and container roles (Priority 1 - Material Design 3).
	RoleSurface   ColourRole = "surface"   // Base surface for cards, sheets, dialogs
	RoleOnSurface ColourRole = "onSurface" // Text/icons on surface
	RoleOutline   ColourRole = "outline"   // Borders, dividers, outlines
	RoleBorder    ColourRole = "border"    // Primary border color

	// Surface and border variants (Priority 2).
	RoleSurfaceVariant   ColourRole = "surfaceVariant"   // Alternate surface color
	RoleOnSurfaceVariant ColourRole = "onSurfaceVariant" // Text on surface variant
	RoleBorderMuted      ColourRole = "borderMuted"      // Inactive/muted borders
	RoleOutlineVariant   ColourRole = "outlineVariant"   // Secondary outline

	// On-colors for accents (Priority 2).
	RoleOnAccent1 ColourRole = "onAccent1" // Text on accent1 background
	RoleOnAccent2 ColourRole = "onAccent2" // Text on accent2 background
	RoleOnAccent3 ColourRole = "onAccent3" // Text on accent3 background
	RoleOnAccent4 ColourRole = "onAccent4" // Text on accent4 background

	// On-colors for semantic roles (Priority 2).
	RoleOnDanger  ColourRole = "onDanger"  // Text on danger background
	RoleOnWarning ColourRole = "onWarning" // Text on warning background
	RoleOnSuccess ColourRole = "onSuccess" // Text on success background
	RoleOnInfo    ColourRole = "onInfo"    // Text on info background

	// Inverse colors for overlays (Priority 3).
	RoleInverseSurface   ColourRole = "inverseSurface"   // Inverse surface (tooltip backgrounds)
	RoleInverseOnSurface ColourRole = "inverseOnSurface" // Text on inverse surface
	RoleInversePrimary   ColourRole = "inversePrimary"   // Inverse accent color

	// Scrim and shadow with alpha (Priority 3).
	RoleScrim  ColourRole = "scrim"  // Modal backdrop overlay (with alpha)
	RoleShadow ColourRole = "shadow" // Elevation shadows (with alpha)

	// Surface container elevation variants (Priority 3 - Material Design 3).
	RoleSurfaceContainerLowest  ColourRole = "surfaceContainerLowest"  // Lowest elevation
	RoleSurfaceContainerLow     ColourRole = "surfaceContainerLow"     // Low elevation
	RoleSurfaceContainer        ColourRole = "surfaceContainer"        // Default container
	RoleSurfaceContainerHigh    ColourRole = "surfaceContainerHigh"    // High elevation
	RoleSurfaceContainerHighest ColourRole = "surfaceContainerHighest" // Highest elevation

	// Positional roles for ambient lighting.
	// Core 8 positions (corners + mid-edges).
	RolePositionTopLeft     ColourRole = "positionTopLeft"
	RolePositionTop         ColourRole = "positionTop"
	RolePositionTopRight    ColourRole = "positionTopRight"
	RolePositionRight       ColourRole = "positionRight"
	RolePositionBottomRight ColourRole = "positionBottomRight"
	RolePositionBottom      ColourRole = "positionBottom"
	RolePositionBottomLeft  ColourRole = "positionBottomLeft"
	RolePositionLeft        ColourRole = "positionLeft"

	// Extended positions for 12+ region configurations.
	RolePositionTopLeftInner     ColourRole = "positionTopLeftInner"
	RolePositionTopCenter        ColourRole = "positionTopCenter"
	RolePositionTopRightInner    ColourRole = "positionTopRightInner"
	RolePositionRightTop         ColourRole = "positionRightTop"
	RolePositionRightBottom      ColourRole = "positionRightBottom"
	RolePositionBottomRightInner ColourRole = "positionBottomRightInner"
	RolePositionBottomCenter     ColourRole = "positionBottomCenter"
	RolePositionBottomLeftInner  ColourRole = "positionBottomLeftInner"
	RolePositionLeftBottom       ColourRole = "positionLeftBottom"
	RolePositionLeftTop          ColourRole = "positionLeftTop"

	// Ultra-extended positions for 16+ region configurations.
	RolePositionTopLeftOuter      ColourRole = "positionTopLeftOuter"
	RolePositionTopLeftCenter     ColourRole = "positionTopLeftCenter"
	RolePositionTopRightCenter    ColourRole = "positionTopRightCenter"
	RolePositionTopRightOuter     ColourRole = "positionTopRightOuter"
	RolePositionRightTopOuter     ColourRole = "positionRightTopOuter"
	RolePositionRightBottomOuter  ColourRole = "positionRightBottomOuter"
	RolePositionBottomRightOuter  ColourRole = "positionBottomRightOuter"
	RolePositionBottomRightCenter ColourRole = "positionBottomRightCenter"
	RolePositionBottomLeftCenter  ColourRole = "positionBottomLeftCenter"
	RolePositionBottomLeftOuter   ColourRole = "positionBottomLeftOuter"
	RolePositionLeftBottomOuter   ColourRole = "positionLeftBottomOuter"
	RolePositionLeftTopOuter      ColourRole = "positionLeftTopOuter"
)

// CategorisedColour represents a colour with its assigned role and metadata.
type CategorisedColour struct {
	Colour      color.Color `json:"-"`
	Role        ColourRole  `json:"role"`
	Hex         string      `json:"hex"`  // #RRGGBB format (backwards compatible)
	RGB         RGB         `json:"rgb"`  // RGB without alpha (backwards compatible)
	RGBA        RGBA        `json:"rgba"` // RGBA with alpha channel (defaults to 255/opaque)
	Luminance   float64     `json:"luminance"`
	IsLight     bool        `json:"is_light"`
	Hue         float64     `json:"hue,omitempty"`          // HSL hue (0-360)
	Saturation  float64     `json:"saturation,omitempty"`   // HSL saturation (0-1)
	Index       int         `json:"index,omitempty"`        // Index in AllColours array (sorted by luminance)
	IsGenerated bool        `json:"is_generated,omitempty"` // True if colour was generated/enhanced, not extracted
	Weight      float64     `json:"weight,omitempty"`       // Original weight from palette (0.0-1.0, 0 if generated)
}

// ThemeType represents whether a theme is light-on-dark or dark-on-light.
type ThemeType int

const (
	// ThemeAuto automatically detects the best theme type based on dominant color luminance.
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
//
// This is the main orchestrator that coordinates all categorization modules:.
// - background.go: Selects background color (theme-aware)
// - foreground.go: Selects foreground color (highest contrast for text)
// - accents.go: Selects and sorts accent colors (analogous to background)
// - muted.go: Creates muted variants (50% saturation reduction)
// - semantic.go: Assigns semantic colors (danger, warning, success, etc.)
//
// Role hints always override automatic categorization.
func Categorise(palette *Palette, config CategorisationConfig) *CategorisedPalette {
	if palette == nil || len(palette.Colors) == 0 {
		return NewCategorisedPalette(ThemeAuto)
	}

	// Create categorised colours with metadata.
	extracted := make([]CategorisedColour, len(palette.Colors))

	// Use equal weights if not provided.
	weights := palette.Weights
	if weights == nil || len(weights) != len(palette.Colors) {
		weights = make([]float64, len(palette.Colors))
		equalWeight := 1.0 / float64(len(palette.Colors))
		for i := range weights {
			weights[i] = equalWeight
		}
	}

	for i, c := range palette.Colors {
		lum := Luminance(c)
		rgb := ToRGB(c)
		rgba := ToRGBA(c)
		h, s, _ := rgbToHSL(rgb)

		extracted[i] = CategorisedColour{
			Colour:      c,
			Hex:         rgb.Hex(),
			RGB:         rgb,
			RGBA:        rgba,
			Luminance:   lum,
			IsLight:     lum > 0.5,
			Hue:         h,
			Saturation:  s,
			IsGenerated: false,
			Weight:      weights[i],
		}
	}

	// Store all extracted colors for later inclusion.
	allExtracted := make([]CategorisedColour, len(extracted))
	copy(allExtracted, extracted)

	// BACKGROUND SELECTION (background.go).
	// Determines theme type and selects background color.
	themeType := config.ThemeType
	var bg CategorisedColour
	bgIdx := -1

	// Apply explicit role hints from input plugins (if provided).
	hintsApplied := make(map[ColourRole]bool)
	if palette.RoleHints != nil {
		for role, originalIndex := range palette.RoleHints {
			if originalIndex >= 0 && originalIndex < len(allExtracted) {
				hintedColor := allExtracted[originalIndex]
				hintedColor.Role = role

				// Store hinted background separately for later use.
				if role == RoleBackground {
					bg = hintedColor
					bgIdx = originalIndex
					hintsApplied[RoleBackground] = true
				}
			}
		}
	}

	// If background not hinted, select it.
	if !hintsApplied[RoleBackground] {
		bg, themeType = selectBackground(extracted, themeType)
		// Find background index in extracted array.
		for i, cc := range extracted {
			if cc.Hex == bg.Hex {
				bgIdx = i
				break
			}
		}
	}

	// Sort extracted colours by luminance for consistent ordering.
	sortByLuminance(extracted, themeType)

	result := NewCategorisedPalette(themeType)
	result.Set(RoleBackground, bg)

	// Apply all other role hints.
	if palette.RoleHints != nil {
		for role, originalIndex := range palette.RoleHints {
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

	// Create background-muted variant (use hint if provided).
	if !hintsApplied[RoleBackgroundMuted] {
		bgMuted := createMutedVariant(bg, config.MutedLuminanceAdjust, themeType, true)
		bgMuted.Role = RoleBackgroundMuted
		bgMuted.IsGenerated = true
		result.Set(RoleBackgroundMuted, bgMuted)
	}

	// FOREGROUND SELECTION (foreground.go).
	var fg CategorisedColour
	fgIdx := -1
	if !hintsApplied[RoleForeground] {
		fgIdx = selectForeground(extracted, bg, config, bgIdx)
		if fgIdx >= 0 {
			fg = extracted[fgIdx]
			fg.Role = RoleForeground
			result.Set(RoleForeground, fg)
		} else {
			// No suitable foreground found in extracted colors (monochromatic palette).
			// Generate synthetic foreground with guaranteed contrast.
			fg = generateSyntheticForeground(bg, themeType, config)
			fg.Role = RoleForeground
			result.Set(RoleForeground, fg)
		}
	} else {
		// Use the hinted foreground.
		fg, _ = result.Get(RoleForeground)
		// Find the index in extracted for later exclusion.
		for i, cc := range extracted {
			if cc.Hex == fg.Hex {
				fgIdx = i
				break
			}
		}
	}

	// Create foreground-muted variant (use hint if provided).
	// Check if foreground exists (either extracted or generated).
	if _, hasFg := result.Get(RoleForeground); hasFg && !hintsApplied[RoleForegroundMuted] {
		fgMuted := createMutedVariant(fg, config.MutedLuminanceAdjust, themeType, false)
		fgMuted.Role = RoleForegroundMuted
		fgMuted.IsGenerated = true
		result.Set(RoleForegroundMuted, fgMuted)
	}

	// ACCENT SELECTION (accents.go).
	// Collect remaining colours for accents (excluding background, foreground, and hinted roles).
	accents := make([]CategorisedColour, 0)
	usedIndices := make(map[int]bool)

	// Mark background index as used.
	if bgIdx >= 0 {
		usedIndices[bgIdx] = true
	}
	// Mark foreground index as used.
	if fgIdx >= 0 {
		usedIndices[fgIdx] = true
	}
	// Mark any hinted role indices as used.
	if palette.RoleHints != nil {
		for _, index := range palette.RoleHints {
			usedIndices[index] = true
		}
	}

	for i, cc := range extracted {
		if !usedIndices[i] {
			accents = append(accents, cc)
		}
	}

	// Sort accents intelligently for better visual progression.
	sortAccentsForTheme(accents, bg, fg, themeType)

	// Check if accents lack sufficient diversity (monochromatic palette).
	// If so, generate synthetic accents with guaranteed contrast progression.
	// Also generate if we have fewer than 4 accents.
	if len(accents) < 4 || areAccentsTooSimilar(accents, bg) {
		accents = generateSyntheticAccents(bg, themeType, 4)
	}

	// Track which accent colors are used for semantic roles.
	usedForSemantic := make(map[string]bool) // Track by hex value

	// Assign accent roles (up to 4) and their muted variants.
	accentRoles := []struct {
		primary ColourRole
		muted   ColourRole
	}{
		{RoleAccent1, RoleAccent1Muted},
		{RoleAccent2, RoleAccent2Muted},
		{RoleAccent3, RoleAccent3Muted},
		{RoleAccent4, RoleAccent4Muted},
	}
	accentIndex := 0
	for _, roles := range accentRoles {
		// Skip if this accent role was explicitly hinted.
		if hintsApplied[roles.primary] {
			continue
		}

		if accentIndex < len(accents) {
			accent := accents[accentIndex]
			accent.Role = roles.primary
			result.Set(roles.primary, accent)

			// Create muted variant for this accent (skip if hinted).
			if !hintsApplied[roles.muted] {
				accentMuted := createMutedVariant(accent, config.MutedLuminanceAdjust, themeType, false)
				accentMuted.Role = roles.muted
				accentMuted.IsGenerated = true
				result.Set(roles.muted, accentMuted)
			}

			accentIndex++
		}
	}

	// SEMANTIC COLOR ASSIGNMENT (semantic.go).
	assignSemanticRolesWithHints(result, accents, usedForSemantic, hintsApplied)

	// SURFACE & CONTAINER COLOR GENERATION (surface.go).
	generateSurfaceColors(result, bg, fg, themeType, hintsApplied)

	// Collect all remaining colors that weren't assigned to any role.
	additionalColors := make([]CategorisedColour, 0)
	for _, cc := range allExtracted {
		// Check if this color was assigned to any role.
		alreadyAssigned := false
		for _, assigned := range result.Colours {
			if assigned.Hex == cc.Hex {
				alreadyAssigned = true
				break
			}
		}
		if !alreadyAssigned {
			// This color wasn't assigned to any semantic role - preserve it.
			additionalColors = append(additionalColors, cc)
		}
	}

	// Build final AllColours array with consistent indices.
	result.AllColours = buildSortedAllColours(result, themeType, additionalColors)

	return result
}

// buildSortedAllColours creates the final sorted array of all colours.
// All colours are sorted by luminance (dark→light for dark themes, light→dark for light themes).
// Indices are assigned sequentially based on sort order (0, 1, 2, ...).
// The index field is purely positional metadata with no semantic meaning.
func buildSortedAllColours(palette *CategorisedPalette, themeType ThemeType, additionalColors []CategorisedColour) []CategorisedColour {
	// Collect all colours from the palette.
	allColours := make([]CategorisedColour, 0, len(palette.Colours)+len(additionalColors))

	for _, cc := range palette.Colours {
		allColours = append(allColours, cc)
	}

	// Add any extra colors that weren't assigned to semantic roles.
	allColours = append(allColours, additionalColors...)

	// Sort all colours by luminance (theme-aware).
	sortByLuminance(allColours, themeType)

	// Assign sequential indices based on sorted position.
	for i := range allColours {
		allColours[i].Index = i
	}

	return allColours
}

// ToJSON converts the categorised palette to JSON format.
func (cp *CategorisedPalette) ToJSON() ([]byte, error) {
	return json.MarshalIndent(cp, "", "  ")
}

// String returns a human-readable string representation of the categorised palette.
func (cp *CategorisedPalette) String() string {
	return cp.StringWithPreview(false)
}

// StringWithPreview returns a string representation with optional table display.
// When showPreview is true, displays a detailed table with all colors and properties.
// Color blocks are always shown in the table.
func (cp *CategorisedPalette) StringWithPreview(showPreview bool) string {
	var result string

	// Show theme type and contrast ratio on one line.
	themeInfo := fmt.Sprintf("Theme Type: %s", cp.ThemeType.String())
	if bg, bgOk := cp.Get(RoleBackground); bgOk {
		if fg, fgOk := cp.Get(RoleForeground); fgOk {
			contrast := ContrastRatio(fg.Colour, bg.Colour)
			themeInfo += fmt.Sprintf(" | Contrast: %.2f:1", contrast)
		}
	}
	result += themeInfo + "\n"

	// Only show table if preview is enabled.
	if !showPreview {
		return result
	}

	result += "\n"

	// Tabular format showing all colours with proper alignment.
	result += "All Colours (sorted by luminance):\n"

	// Build table data.
	rows := make([][]string, 0, len(cp.AllColours))

	// Header row (added empty first column for marker).
	header := []string{"", "Preview", "Role", "Index", "Hex", "Luminance", "Saturation", "Weight", "Source"}

	// Data rows.
	for _, cc := range cp.AllColours {
		roleName := string(cc.Role)
		if roleName == "" {
			roleName = "-"
		}
		indexStr := fmt.Sprintf("colour%d", cc.Index)
		source := "extracted"
		if cc.IsGenerated {
			source = "generated"
		}

		weightStr := "-"
		if cc.Weight > 0 {
			weightStr = fmt.Sprintf("%.1f%%", cc.Weight*100)
		}

		// Always show color preview blocks.
		preview := ColourPreview(cc.RGB, 8)

		// Mark the background role with ">".
		marker := " "
		if cc.Role == RoleBackground {
			marker = ">"
		}

		rows = append(rows, []string{
			marker,
			preview,
			roleName,
			indexStr,
			cc.Hex,
			fmt.Sprintf("%.2f", cc.Luminance),
			fmt.Sprintf("%.2f", cc.Saturation),
			weightStr,
			source,
		})
	}

	// Calculate column widths.
	colWidths := make([]int, len(header))
	for i, h := range header {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			// For preview column (index 1), count visible width (8 for the color block).
			cellLen := len(cell)
			if i == 1 {
				cellLen = 8 // Color preview is always 8 visible characters
			}
			if cellLen > colWidths[i] {
				colWidths[i] = cellLen
			}
		}
	}

	// Format header.
	headerParts := make([]string, 0, len(header))
	for i, h := range header {
		headerParts = append(headerParts, padRight(h, colWidths[i]))
	}
	result += "  " + strings.Join(headerParts, "  ") + "\n"

	// Format separator.
	sepParts := make([]string, 0, len(colWidths))
	for _, w := range colWidths {
		sepParts = append(sepParts, strings.Repeat("-", w))
	}
	result += "  " + strings.Join(sepParts, "  ") + "\n"

	// Format data rows.
	for _, row := range rows {
		var rowParts []string
		for i, cell := range row {
			if i == 1 {
				// Preview column (index 1): don't pad (ANSI codes mess up padding).
				rowParts = append(rowParts, cell+strings.Repeat(" ", colWidths[i]-8))
			} else {
				rowParts = append(rowParts, padRight(cell, colWidths[i]))
			}
		}
		result += "  " + strings.Join(rowParts, "  ") + "\n"
	}

	return result
}

// padRight pads a string with spaces on the right to reach the desired width.
// If the string is already longer than or equal to the width, it is returned unchanged.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
