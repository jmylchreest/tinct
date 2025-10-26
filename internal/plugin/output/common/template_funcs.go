// Package common provides shared utilities for output plugins.
package common

import (
	"fmt"
	"text/template"

	"github.com/jmylchreest/tinct/internal/colour"
)

// TemplateFuncs returns standard template functions for all output plugins.
// These functions provide consistent color access and formatting across all templates.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// Color role access
		"get":        getRoleFunc,
		"getSafe":    getSafeRoleFunc,
		"has":        hasRoleFunc,
		"getByIndex": getByIndexFunc,

		// Format conversion
		"hex":        hexFunc,
		"hexAlpha":   hexAlphaFunc,
		"hexNoHash":  hexNoHashFunc,
		"rgb":        rgbFunc,
		"rgba":       rgbaFunc,
		"rgbDecimal": rgbDecimalFunc,

		// Alpha manipulation
		"withAlpha": withAlphaFunc,

		// Color metadata
		"role":  roleFunc,
		"index": indexFunc,

		// Palette metadata
		"themeType": themeTypeFunc,
		"allRoles":  allRolesFunc,
		"allColors": allColorsFunc,
		"count":     countFunc,
	}
}

// getRoleFunc returns a color by role name (string).
// Panics if role doesn't exist - use getSafe or has to check first.
func getRoleFunc(ph *colour.PaletteHelper, roleName string) colour.ColorValue {
	return ph.Get(colour.ColourRole(roleName))
}

// getSafeRoleFunc returns a color by role name with existence check.
// Returns error if role doesn't exist (Go template convention).
func getSafeRoleFunc(ph *colour.PaletteHelper, roleName string) (colour.ColorValue, error) {
	cv, ok := ph.GetSafe(colour.ColourRole(roleName))
	if !ok {
		return colour.ColorValue{}, fmt.Errorf("role %q not found", roleName)
	}
	return cv, nil
}

// hasRoleFunc checks if a role exists in the palette.
func hasRoleFunc(ph *colour.PaletteHelper, roleName string) bool {
	return ph.Has(colour.ColourRole(roleName))
}

// getByIndexFunc returns a color by index in the AllColors array.
func getByIndexFunc(ph *colour.PaletteHelper, index int) (colour.ColorValue, error) {
	cv, ok := ph.GetByIndex(index)
	if !ok {
		return colour.ColorValue{}, fmt.Errorf("index %d out of range (palette has %d colors)", index, ph.Count())
	}
	return cv, nil
}

// hexFunc returns color in #RRGGBB format.
func hexFunc(cv colour.ColorValue) string {
	return cv.Hex()
}

// hexAlphaFunc returns color in #RRGGBBAA format.
func hexAlphaFunc(cv colour.ColorValue) string {
	return cv.HexAlpha()
}

// hexNoHashFunc returns color in RRGGBB format (no # prefix).
func hexNoHashFunc(cv colour.ColorValue) string {
	return cv.HexNoHash()
}

// rgbFunc returns color in CSS rgb(r,g,b) format.
func rgbFunc(cv colour.ColorValue) string {
	return cv.RGB()
}

// rgbaFunc returns color in CSS rgba(r,g,b,a) format.
func rgbaFunc(cv colour.ColorValue) string {
	return cv.RGBA()
}

// rgbDecimalFunc returns color in "r,g,b" decimal format (for Hyprland).
func rgbDecimalFunc(cv colour.ColorValue) string {
	return cv.RGBDecimal()
}

// withAlphaFunc returns a copy of the color with custom alpha (0.0-1.0).
func withAlphaFunc(cv colour.ColorValue, alpha float64) colour.ColorValue {
	return cv.WithAlpha(alpha)
}

// roleFunc returns the role name of a color.
func roleFunc(cv colour.ColorValue) string {
	return string(cv.Role())
}

// indexFunc returns the index of a color in the AllColors array.
func indexFunc(cv colour.ColorValue) int {
	return cv.Index()
}

// themeTypeFunc returns the theme type string ("dark" or "light").
func themeTypeFunc(ph *colour.PaletteHelper) string {
	return ph.ThemeTypeString()
}

// allRolesFunc returns all color roles in deterministic order.
func allRolesFunc(ph *colour.PaletteHelper) []colour.ColourRole {
	return ph.AllRoles()
}

// allColorsFunc returns all colors in index order.
func allColorsFunc(ph *colour.PaletteHelper) []colour.ColorValue {
	return ph.AllColors()
}

// countFunc returns the total number of colors in the palette.
func countFunc(ph *colour.PaletteHelper) int {
	return ph.Count()
}
