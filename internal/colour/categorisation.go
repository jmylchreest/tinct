// Package color provides color extraction and palette generation functionality.
package colour

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strings"
)

// Role represents the semantic role of a colour in a theme.
type Role string

const (
	// Core roles.
	RoleBackground      Role = "background"
	RoleBackgroundMuted Role = "backgroundMuted"
	RoleForeground      Role = "foreground"
	RoleForegroundMuted Role = "foregroundMuted"

	// Accent roles.
	RoleAccent1      Role = "accent1"
	RoleAccent1Muted Role = "accent1Muted"
	RoleAccent2      Role = "accent2"
	RoleAccent2Muted Role = "accent2Muted"
	RoleAccent3      Role = "accent3"
	RoleAccent3Muted Role = "accent3Muted"
	RoleAccent4      Role = "accent4"
	RoleAccent4Muted Role = "accent4Muted"

	// Semantic roles.
	RoleDanger       Role = "danger"
	RoleWarning      Role = "warning"
	RoleSuccess      Role = "success"
	RoleInfo         Role = "info"
	RoleNotification Role = "notification"

	// Surface and container roles (Priority 1 - Material Design 3).
	RoleSurface   Role = "surface"   // Base surface for cards, sheets, dialogs
	RoleOnSurface Role = "onSurface" // Text/icons on surface
	RoleOutline   Role = "outline"   // Borders, dividers, outlines
	RoleBorder    Role = "border"    // Primary border color

	// Surface and border variants (Priority 2).
	RoleSurfaceVariant   Role = "surfaceVariant"   // Alternate surface color
	RoleOnSurfaceVariant Role = "onSurfaceVariant" // Text on surface variant
	RoleBorderMuted      Role = "borderMuted"      // Inactive/muted borders
	RoleOutlineVariant   Role = "outlineVariant"   // Secondary outline

	// On-colors for accents (Priority 2).
	RoleOnAccent1 Role = "onAccent1" // Text on accent1 background
	RoleOnAccent2 Role = "onAccent2" // Text on accent2 background
	RoleOnAccent3 Role = "onAccent3" // Text on accent3 background
	RoleOnAccent4 Role = "onAccent4" // Text on accent4 background

	// On-colors for semantic roles (Priority 2).
	RoleOnDanger  Role = "onDanger"  // Text on danger background
	RoleOnWarning Role = "onWarning" // Text on warning background
	RoleOnSuccess Role = "onSuccess" // Text on success background
	RoleOnInfo    Role = "onInfo"    // Text on info background

	// Inverse colors for overlays (Priority 3).
	RoleInverseSurface   Role = "inverseSurface"   // Inverse surface (tooltip backgrounds)
	RoleInverseOnSurface Role = "inverseOnSurface" // Text on inverse surface
	RoleInversePrimary   Role = "inversePrimary"   // Inverse accent color

	// Scrim and shadow with alpha (Priority 3).
	RoleScrim  Role = "scrim"  // Modal backdrop overlay (with alpha)
	RoleShadow Role = "shadow" // Elevation shadows (with alpha)

	// Surface container elevation variants (Priority 3 - Material Design 3).
	RoleSurfaceContainerLowest  Role = "surfaceContainerLowest"  // Lowest elevation
	RoleSurfaceContainerLow     Role = "surfaceContainerLow"     // Low elevation
	RoleSurfaceContainer        Role = "surfaceContainer"        // Default container
	RoleSurfaceContainerHigh    Role = "surfaceContainerHigh"    // High elevation
	RoleSurfaceContainerHighest Role = "surfaceContainerHighest" // Highest elevation

	// Positional roles for ambient lighting.
	// Core 8 positions (corners + mid-edges).
	RolePositionTopLeft     Role = "positionTopLeft"
	RolePositionTop         Role = "positionTop"
	RolePositionTopRight    Role = "positionTopRight"
	RolePositionRight       Role = "positionRight"
	RolePositionBottomRight Role = "positionBottomRight"
	RolePositionBottom      Role = "positionBottom"
	RolePositionBottomLeft  Role = "positionBottomLeft"
	RolePositionLeft        Role = "positionLeft"

	// Extended positions for 12+ region configurations.
	RolePositionTopLeftInner     Role = "positionTopLeftInner"
	RolePositionTopCenter        Role = "positionTopCenter"
	RolePositionTopRightInner    Role = "positionTopRightInner"
	RolePositionRightTop         Role = "positionRightTop"
	RolePositionRightBottom      Role = "positionRightBottom"
	RolePositionBottomRightInner Role = "positionBottomRightInner"
	RolePositionBottomCenter     Role = "positionBottomCenter"
	RolePositionBottomLeftInner  Role = "positionBottomLeftInner"
	RolePositionLeftBottom       Role = "positionLeftBottom"
	RolePositionLeftTop          Role = "positionLeftTop"

	// Ultra-extended positions for 16+ region configurations.
	RolePositionTopLeftOuter      Role = "positionTopLeftOuter"
	RolePositionTopLeftCenter     Role = "positionTopLeftCenter"
	RolePositionTopRightCenter    Role = "positionTopRightCenter"
	RolePositionTopRightOuter     Role = "positionTopRightOuter"
	RolePositionRightTopOuter     Role = "positionRightTopOuter"
	RolePositionRightBottomOuter  Role = "positionRightBottomOuter"
	RolePositionBottomRightOuter  Role = "positionBottomRightOuter"
	RolePositionBottomRightCenter Role = "positionBottomRightCenter"
	RolePositionBottomLeftCenter  Role = "positionBottomLeftCenter"
	RolePositionBottomLeftOuter   Role = "positionBottomLeftOuter"
	RolePositionLeftBottomOuter   Role = "positionLeftBottomOuter"
	RolePositionLeftTopOuter      Role = "positionLeftTopOuter"
)

// CategorisedColour represents a colour with its assigned role and metadata.
type CategorisedColour struct {
	Colour      color.Color `json:"-"`
	Role        Role        `json:"role"`
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
	Colours    map[Role]CategorisedColour `json:"colours"`
	ThemeType  ThemeType                  `json:"theme_type"`
	AllColours []CategorisedColour        `json:"all_colours,omitempty"`
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
