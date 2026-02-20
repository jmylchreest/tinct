// Package color provides color extraction and palette generation functionality.
package colour

import (
	"encoding/json"
	"fmt"
	"strings"

	pkgcolour "github.com/jmylchreest/tinct/pkg/colour"
)

// Re-export public types from pkg/colour for backward compatibility.
type (
	Role              = pkgcolour.Role
	ThemeType         = pkgcolour.ThemeType
	RGB               = pkgcolour.RGB
	RGBA              = pkgcolour.RGBA
	CategorisedColour = pkgcolour.CategorisedColour
	ANSIColors        = pkgcolour.ANSIColors
)

// Re-export Role constants from pkg/colour.
const (
	RoleBackground                = pkgcolour.RoleBackground
	RoleBackgroundMuted           = pkgcolour.RoleBackgroundMuted
	RoleForeground                = pkgcolour.RoleForeground
	RoleForegroundMuted           = pkgcolour.RoleForegroundMuted
	RoleAccent1                   = pkgcolour.RoleAccent1
	RoleAccent1Muted              = pkgcolour.RoleAccent1Muted
	RoleAccent2                   = pkgcolour.RoleAccent2
	RoleAccent2Muted              = pkgcolour.RoleAccent2Muted
	RoleAccent3                   = pkgcolour.RoleAccent3
	RoleAccent3Muted              = pkgcolour.RoleAccent3Muted
	RoleAccent4                   = pkgcolour.RoleAccent4
	RoleAccent4Muted              = pkgcolour.RoleAccent4Muted
	RoleDanger                    = pkgcolour.RoleDanger
	RoleWarning                   = pkgcolour.RoleWarning
	RoleSuccess                   = pkgcolour.RoleSuccess
	RoleInfo                      = pkgcolour.RoleInfo
	RoleNotification              = pkgcolour.RoleNotification
	RoleSurface                   = pkgcolour.RoleSurface
	RoleOnSurface                 = pkgcolour.RoleOnSurface
	RoleOutline                   = pkgcolour.RoleOutline
	RoleBorder                    = pkgcolour.RoleBorder
	RoleSurfaceVariant            = pkgcolour.RoleSurfaceVariant
	RoleOnSurfaceVariant          = pkgcolour.RoleOnSurfaceVariant
	RoleBorderMuted               = pkgcolour.RoleBorderMuted
	RoleOutlineVariant            = pkgcolour.RoleOutlineVariant
	RoleOnAccent1                 = pkgcolour.RoleOnAccent1
	RoleOnAccent2                 = pkgcolour.RoleOnAccent2
	RoleOnAccent3                 = pkgcolour.RoleOnAccent3
	RoleOnAccent4                 = pkgcolour.RoleOnAccent4
	RoleOnDanger                  = pkgcolour.RoleOnDanger
	RoleOnWarning                 = pkgcolour.RoleOnWarning
	RoleOnSuccess                 = pkgcolour.RoleOnSuccess
	RoleOnInfo                    = pkgcolour.RoleOnInfo
	RoleInverseSurface            = pkgcolour.RoleInverseSurface
	RoleInverseOnSurface          = pkgcolour.RoleInverseOnSurface
	RoleInversePrimary            = pkgcolour.RoleInversePrimary
	RoleScrim                     = pkgcolour.RoleScrim
	RoleShadow                    = pkgcolour.RoleShadow
	RoleSurfaceContainerLowest    = pkgcolour.RoleSurfaceContainerLowest
	RoleSurfaceContainerLow       = pkgcolour.RoleSurfaceContainerLow
	RoleSurfaceContainer          = pkgcolour.RoleSurfaceContainer
	RoleSurfaceContainerHigh      = pkgcolour.RoleSurfaceContainerHigh
	RoleSurfaceContainerHighest   = pkgcolour.RoleSurfaceContainerHighest
	RolePositionTopLeft           = pkgcolour.RolePositionTopLeft
	RolePositionTop               = pkgcolour.RolePositionTop
	RolePositionTopRight          = pkgcolour.RolePositionTopRight
	RolePositionRight             = pkgcolour.RolePositionRight
	RolePositionBottomRight       = pkgcolour.RolePositionBottomRight
	RolePositionBottom            = pkgcolour.RolePositionBottom
	RolePositionBottomLeft        = pkgcolour.RolePositionBottomLeft
	RolePositionLeft              = pkgcolour.RolePositionLeft
	RolePositionTopLeftInner      = pkgcolour.RolePositionTopLeftInner
	RolePositionTopCenter         = pkgcolour.RolePositionTopCenter
	RolePositionTopRightInner     = pkgcolour.RolePositionTopRightInner
	RolePositionRightTop          = pkgcolour.RolePositionRightTop
	RolePositionRightBottom       = pkgcolour.RolePositionRightBottom
	RolePositionBottomRightInner  = pkgcolour.RolePositionBottomRightInner
	RolePositionBottomCenter      = pkgcolour.RolePositionBottomCenter
	RolePositionBottomLeftInner   = pkgcolour.RolePositionBottomLeftInner
	RolePositionLeftBottom        = pkgcolour.RolePositionLeftBottom
	RolePositionLeftTop           = pkgcolour.RolePositionLeftTop
	RolePositionTopLeftOuter      = pkgcolour.RolePositionTopLeftOuter
	RolePositionTopLeftCenter     = pkgcolour.RolePositionTopLeftCenter
	RolePositionTopRightCenter    = pkgcolour.RolePositionTopRightCenter
	RolePositionTopRightOuter     = pkgcolour.RolePositionTopRightOuter
	RolePositionRightTopOuter     = pkgcolour.RolePositionRightTopOuter
	RolePositionRightBottomOuter  = pkgcolour.RolePositionRightBottomOuter
	RolePositionBottomRightOuter  = pkgcolour.RolePositionBottomRightOuter
	RolePositionBottomRightCenter = pkgcolour.RolePositionBottomRightCenter
	RolePositionBottomLeftCenter  = pkgcolour.RolePositionBottomLeftCenter
	RolePositionBottomLeftOuter   = pkgcolour.RolePositionBottomLeftOuter
	RolePositionLeftBottomOuter   = pkgcolour.RolePositionLeftBottomOuter
	RolePositionLeftTopOuter      = pkgcolour.RolePositionLeftTopOuter
)

// Re-export ThemeType constants from pkg/colour.
const (
	ThemeDark  = pkgcolour.ThemeDark
	ThemeLight = pkgcolour.ThemeLight
)

// ThemeAuto is internal-only for automatic theme detection.
const ThemeAuto ThemeType = 0

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
	Colours    map[Role]CategorisedColour `json:"colours"`
	ThemeType  ThemeType                  `json:"theme_type"`
	AllColours []CategorisedColour        `json:"all_colours,omitempty"`
	ANSI       *ANSIColors                `json:"ansi,omitempty"` // ANSI terminal colors
}

// NewCategorisedPalette creates a new categorised palette.
func NewCategorisedPalette(themeType ThemeType) *CategorisedPalette {
	return &CategorisedPalette{
		Colours:   make(map[Role]CategorisedColour),
		ThemeType: themeType,
	}
}

// Get returns a colour by role, if it exists.
func (cp *CategorisedPalette) Get(role Role) (CategorisedColour, bool) {
	c, ok := cp.Colours[role]
	return c, ok
}

// Set sets a colour for a given role.
func (cp *CategorisedPalette) Set(role Role, colour CategorisedColour) {
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

	// Step 1: Create categorised colours with metadata.
	extracted := createCategorisedColours(palette)
	allExtracted := make([]CategorisedColour, len(extracted))
	copy(allExtracted, extracted)

	// Step 2: Select background and apply hints.
	hintsApplied := make(map[Role]bool)
	bg, bgIdx, themeType := selectBackgroundWithHints(extracted, allExtracted,
		palette.RoleHints, config.ThemeType, hintsApplied)

	// Sort extracted colours by luminance for consistent ordering.
	sortByLuminance(extracted, themeType)

	result := NewCategorisedPalette(themeType)
	result.Set(RoleBackground, bg)

	// Step 3: Apply other role hints.
	applyRoleHints(result, extracted, allExtracted, palette.RoleHints, hintsApplied)

	// Step 4: Select foreground.
	fg, fgIdx := selectForegroundWithHints(result, extracted, bg, bgIdx, config,
		themeType, hintsApplied)

	// Step 5: Create muted variants for background and foreground.
	addMutedVariants(result, bg, fg, themeType, config, palette.RoleHints)

	// Step 6: Collect and sort accent colours.
	accents := collectAccentColours(extracted, palette.RoleHints, bgIdx, fgIdx)
	sortAccentsForTheme(accents, bg, fg, themeType)

	// Generate synthetic accents if needed.
	if needsSyntheticAccents(accents, bg) {
		accents = generateSyntheticAccents(bg, themeType, 4)
	}

	// Step 7: Assign accent roles and their muted variants.
	assignAccentRoles(result, accents, themeType, config, palette.RoleHints)

	// Step 8: Assign semantic roles.
	usedForSemantic := make(map[string]bool)
	assignSemanticRolesWithHints(result, accents, usedForSemantic, hintsApplied)

	// Step 9: Generate surface and container colors.
	generateSurfaceColors(result, bg, fg, themeType, hintsApplied)

	// Step 10: Collect unassigned colors.
	additionalColors := collectUnassignedColors(allExtracted, result)

	// Step 11: Build final AllColours array.
	result.AllColours = buildSortedAllColours(result, themeType, additionalColors)

	return result
}

// collectUnassignedColors collects colors that weren't assigned to any role.
func collectUnassignedColors(allExtracted []CategorisedColour, result *CategorisedPalette) []CategorisedColour {
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
			additionalColors = append(additionalColors, cc)
		}
	}
	return additionalColors
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
func (cp *CategorisedPalette) StringWithPreview(showPreview bool) string { //nolint:gocognit // formatting logic with many branches
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
		preview := Preview(cc.RGB, 8)

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
