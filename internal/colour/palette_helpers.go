// Package colour provides palette helper utilities for output plugins.
package colour

import (
	"fmt"
	"strings"
)

// ColorFormat represents different color string formats.
type ColorFormat int

const (
	FormatHex         ColorFormat = iota // #RRGGBB
	FormatHexAlpha                       // #RRGGBBAA
	FormatRGB                            // rgb(r,g,b)
	FormatRGBA                           // rgba(r,g,b,a)
	FormatHexNoHash                      // RRGGBB (for Hyprland)
	FormatRGBDecimal                     // "r,g,b" (for Hyprland)
	FormatRGBADecimal                    // "r,g,b,a" decimal format
)

// ColorValue provides multiple format accessors for a single color.
// This is the primary type templates and plugins will interact with.
type ColorValue struct {
	role  Role
	rgba  RGBA
	index int
}

// NewColorValue creates a ColorValue from RGBA with optional metadata.
func NewColorValue(rgba RGBA, role Role, index int) ColorValue {
	return ColorValue{
		role:  role,
		rgba:  rgba,
		index: index,
	}
}

// WithAlpha returns a copy of the ColorValue with custom alpha (0.0-1.0).
// This is useful for creating transparent variants in templates.
func (cv ColorValue) WithAlpha(alpha float64) ColorValue {
	newCV := cv
	if alpha < 0 {
		alpha = 0
	} else if alpha > 1 {
		alpha = 1
	}
	newCV.rgba.A = uint8(alpha * 255.0)
	return newCV
}

// Format returns the color in the specified format.
func (cv ColorValue) Format(format ColorFormat) string {
	switch format {
	case FormatHex:
		return cv.rgba.Hex()
	case FormatHexAlpha:
		return cv.rgba.HexAlpha()
	case FormatRGB:
		return cv.rgba.CSSRgb()
	case FormatRGBA:
		return cv.rgba.CSSRgba()
	case FormatHexNoHash:
		return strings.TrimPrefix(cv.rgba.Hex(), "#")
	case FormatRGBDecimal:
		return fmt.Sprintf("%d,%d,%d", cv.rgba.R, cv.rgba.G, cv.rgba.B)
	case FormatRGBADecimal:
		return fmt.Sprintf("%d,%d,%d,%.2f", cv.rgba.R, cv.rgba.G, cv.rgba.B, cv.rgba.AlphaFloat())
	default:
		return cv.rgba.Hex()
	}
}

// Convenience accessors for common formats.
func (cv ColorValue) Hex() string        { return cv.Format(FormatHex) }
func (cv ColorValue) HexAlpha() string   { return cv.Format(FormatHexAlpha) }
func (cv ColorValue) RGB() string        { return cv.Format(FormatRGB) }
func (cv ColorValue) RGBA() string       { return cv.Format(FormatRGBA) }
func (cv ColorValue) HexNoHash() string  { return cv.Format(FormatHexNoHash) }
func (cv ColorValue) RGBDecimal() string { return cv.Format(FormatRGBDecimal) }

// Metadata accessors.
func (cv ColorValue) Role() Role { return cv.role }
func (cv ColorValue) Index() int { return cv.index }

// Component accessors (for advanced template use).
func (cv ColorValue) R() uint8            { return cv.rgba.R }
func (cv ColorValue) G() uint8            { return cv.rgba.G }
func (cv ColorValue) B() uint8            { return cv.rgba.B }
func (cv ColorValue) A() uint8            { return cv.rgba.A }
func (cv ColorValue) AlphaFloat() float64 { return cv.rgba.AlphaFloat() }

// PaletteHelper provides DRY utilities for output plugins to access palette colors.
// This eliminates code duplication across all output plugins.
type PaletteHelper struct {
	palette *CategorisedPalette
	colors  map[Role]ColorValue
	indexed []ColorValue
}

// NewPaletteHelper creates a helper for the given categorised palette.
// This should be called once per plugin Generate() invocation.
func NewPaletteHelper(palette *CategorisedPalette) *PaletteHelper {
	ph := &PaletteHelper{
		palette: palette,
		colors:  make(map[Role]ColorValue),
		indexed: make([]ColorValue, 0, len(palette.AllColours)),
	}

	// Build color map by role.
	for role, cc := range palette.Colours {
		ph.colors[role] = ColorValue{
			role:  role,
			rgba:  cc.RGBA,
			index: cc.Index,
		}
	}

	// Build indexed array (deterministic order).
	for _, cc := range palette.AllColours {
		ph.indexed = append(ph.indexed, ColorValue{
			role:  cc.Role,
			rgba:  cc.RGBA,
			index: cc.Index,
		})
	}

	return ph
}

// Get returns color by role. Panics if role doesn't exist - use Has() first to check.
func (ph *PaletteHelper) Get(role Role) ColorValue {
	if cv, ok := ph.colors[role]; ok {
		return cv
	}
	panic(fmt.Sprintf("color role %q not found in palette", role))
}

// GetSafe returns color by role with a boolean indicating if it exists.
func (ph *PaletteHelper) GetSafe(role Role) (ColorValue, bool) {
	cv, ok := ph.colors[role]
	return cv, ok
}

// Has checks if a role exists in the palette.
func (ph *PaletteHelper) Has(role Role) bool {
	_, ok := ph.colors[role]
	return ok
}

// GetWithFallback returns color by role or parses a fallback hex string if missing.
func (ph *PaletteHelper) GetWithFallback(role Role, fallbackHex string) ColorValue {
	if cv, ok := ph.colors[role]; ok {
		return cv
	}
	// Parse fallback hex and create ColorValue.
	rgb := parseHex(fallbackHex)
	return ColorValue{
		role:  role,
		rgba:  RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255},
		index: -1,
	}
}

// GetByIndex returns color by index in the AllColours array.
// Returns false if index is out of bounds.
func (ph *PaletteHelper) GetByIndex(index int) (ColorValue, bool) {
	if index < 0 || index >= len(ph.indexed) {
		return ColorValue{}, false
	}
	return ph.indexed[index], true
}

// AllRoles returns all roles in deterministic order (core → accents → semantic → surface → variants).
func (ph *PaletteHelper) AllRoles() []Role {
	// Define priority order for consistency across all plugins.
	order := []Role{
		// Core colors.
		RoleBackground, RoleBackgroundMuted,
		RoleForeground, RoleForegroundMuted,

		// Accents.
		RoleAccent1, RoleAccent1Muted,
		RoleAccent2, RoleAccent2Muted,
		RoleAccent3, RoleAccent3Muted,
		RoleAccent4, RoleAccent4Muted,

		// Semantic.
		RoleDanger, RoleWarning, RoleSuccess, RoleInfo, RoleNotification,

		// Surface & Container (Priority 1).
		RoleSurface, RoleOnSurface, RoleOutline, RoleBorder,

		// Surface & Border Variants (Priority 2).
		RoleSurfaceVariant, RoleOnSurfaceVariant,
		RoleBorderMuted, RoleOutlineVariant,

		// On-colors for Accents (Priority 2).
		RoleOnAccent1, RoleOnAccent2, RoleOnAccent3, RoleOnAccent4,

		// On-colors for Semantic (Priority 2).
		RoleOnDanger, RoleOnWarning, RoleOnSuccess, RoleOnInfo,

		// Inverse Colors (Priority 3).
		RoleInverseSurface, RoleInverseOnSurface, RoleInversePrimary,

		// Scrim & Shadow (Priority 3).
		RoleScrim, RoleShadow,

		// Container Elevation Variants (Priority 3).
		RoleSurfaceContainerLowest, RoleSurfaceContainerLow, RoleSurfaceContainer,
		RoleSurfaceContainerHigh, RoleSurfaceContainerHighest,
	}

	var result []Role
	for _, role := range order {
		if ph.Has(role) {
			result = append(result, role)
		}
	}
	return result
}

// AllColors returns all colors indexed by palette order (AllColours array).
func (ph *PaletteHelper) AllColors() []ColorValue {
	return ph.indexed
}

// Count returns total number of colors in the palette.
func (ph *PaletteHelper) Count() int {
	return len(ph.indexed)
}

// ThemeType returns the detected or specified theme type.
func (ph *PaletteHelper) ThemeType() ThemeType {
	return ph.palette.ThemeType
}

// ThemeTypeString returns the theme type as a string ("dark" or "light").
func (ph *PaletteHelper) ThemeTypeString() string {
	return ph.palette.ThemeType.String()
}

// Palette returns the underlying CategorisedPalette.
// This is primarily used for backward compatibility with external plugins.
func (ph *PaletteHelper) Palette() *CategorisedPalette {
	return ph.palette
}

// parseHex parses a hex color string (#RRGGBB or RRGGBB) to RGB.
// Returns black if parsing fails.
func parseHex(hex string) RGB {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return RGB{R: 0, G: 0, B: 0}
	}

	var r, g, b uint8
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil {
		return RGB{R: 0, G: 0, B: 0}
	}

	return RGB{R: r, G: g, B: b}
}
