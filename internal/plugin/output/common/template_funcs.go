// Package common provides shared utilities for output plugins.
package common

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/jmylchreest/tinct/internal/colour"
)

// TemplateFuncs returns standard template functions for all output plugins.
// These functions provide consistent color access and formatting across all templates.
func TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		// Color role access.
		"get":        getRoleFunc,
		"getSafe":    getSafeRoleFunc,
		"has":        hasRoleFunc,
		"getByIndex": getByIndexFunc,

		// ANSI color matching.
		"ansi":     ansiFunc,
		"ansiSafe": ansiSafeFunc,

		// Format conversion.
		"hex":         hexFunc,
		"hexAlpha":    hexAlphaFunc,
		"hexNoHash":   hexNoHashFunc,
		"rgb":         rgbFunc,
		"rgba":        rgbaFunc,
		"rgbDecimal":  rgbDecimalFunc,
		"rgbaDecimal": rgbaDecimalFunc,
		"rgbSpaces":   rgbSpacesFunc,

		// Alpha manipulation.
		"withAlpha": withAlphaFunc,

		// Color metadata.
		"role":  roleFunc,
		"index": indexFunc,

		// Palette metadata.
		"themeType": themeTypeFunc,
		"allRoles":  allRolesFunc,
		"allColors": allColorsFunc,
		"count":     countFunc,

		// String manipulation (custom wrappers for pipe-friendly argument order).
		"trimPrefix": trimPrefixFunc,
		"trimSuffix": trimSuffixFunc,
		"replace":    replaceFunc,
		"toLower":    strings.ToLower,
		"toUpper":    strings.ToUpper,
	}
}

// getRoleFunc returns a color by role name (string).
// Panics if role doesn't exist - use getSafe or has to check first.
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func getRoleFunc(data any, roleName string) colour.ColorValue {
	ph := extractPaletteHelper(data)
	return ph.Get(colour.Role(roleName))
}

// getSafeRoleFunc returns a color by role name with existence check.
// Returns error if role doesn't exist (Go template convention).
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func getSafeRoleFunc(data any, roleName string) (colour.ColorValue, error) {
	ph := extractPaletteHelper(data)
	cv, ok := ph.GetSafe(colour.Role(roleName))
	if !ok {
		return colour.ColorValue{}, fmt.Errorf("role %q not found", roleName)
	}
	return cv, nil
}

// hasRoleFunc checks if a role exists in the palette.
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func hasRoleFunc(data any, roleName string) bool {
	ph := extractPaletteHelper(data)
	return ph.Has(colour.Role(roleName))
}

// getByIndexFunc returns a color by index in the AllColors array.
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func getByIndexFunc(data any, index int) (colour.ColorValue, error) {
	ph := extractPaletteHelper(data)
	cv, ok := ph.GetByIndex(index)
	if !ok {
		return colour.ColorValue{}, fmt.Errorf("index %d out of range (palette has %d colors)", index, ph.Count())
	}
	return cv, nil
}

// extractPaletteHelper extracts PaletteHelper from either *ThemeData or *PaletteHelper.
func extractPaletteHelper(data any) *colour.PaletteHelper {
	switch v := data.(type) {
	case *colour.ThemeData:
		return v.PaletteHelper
	case *colour.PaletteHelper:
		return v
	default:
		panic(fmt.Sprintf("expected *colour.ThemeData or *colour.PaletteHelper, got %T", data))
	}
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

// rgbaDecimalFunc returns color in "r,g,b,a" decimal format with alpha (for Hyprland/Hyprlock).
func rgbaDecimalFunc(cv colour.ColorValue) string {
	return cv.Format(colour.FormatRGBADecimal)
}

// rgbSpacesFunc returns color in "r g b" space-separated format (for Zellij KDL).
func rgbSpacesFunc(cv colour.ColorValue) string {
	return fmt.Sprintf("%d %d %d", cv.R(), cv.G(), cv.B())
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

// trimPrefixFunc removes a prefix from a string (pipe-friendly argument order).
// Unlike strings.TrimPrefix, this takes prefix first so it works in pipes:
//
//	{{ value | trimPrefix "#" }}
func trimPrefixFunc(prefix, s string) string {
	return strings.TrimPrefix(s, prefix)
}

// trimSuffixFunc removes a suffix from a string (pipe-friendly argument order).
// Unlike strings.TrimSuffix, this takes suffix first so it works in pipes:
//
//	{{ value | trimSuffix ".txt" }}
func trimSuffixFunc(suffix, s string) string {
	return strings.TrimSuffix(s, suffix)
}

// replaceFunc replaces all occurrences of old with new (pipe-friendly argument order).
// Unlike strings.ReplaceAll, this takes old and new first so it works in pipes:
//
//	{{ value | replace "_" "-" }}
func replaceFunc(old, new, s string) string {
	return strings.ReplaceAll(s, old, new)
}

// themeTypeFunc returns the theme type string ("dark" or "light").
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func themeTypeFunc(data any) string {
	ph := extractPaletteHelper(data)
	return ph.ThemeTypeString()
}

// allRolesFunc returns all color roles in deterministic order.
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func allRolesFunc(data any) []colour.Role {
	ph := extractPaletteHelper(data)
	return ph.AllRoles()
}

// allColorsFunc returns all colors in index order.
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func allColorsFunc(data any) []colour.ColorValue {
	ph := extractPaletteHelper(data)
	return ph.AllColors()
}

// countFunc returns the total number of colors in the palette.
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func countFunc(data any) int {
	ph := extractPaletteHelper(data)
	return ph.Count()
}

// ansiFunc finds the closest color to a given ANSI color name.
// Panics if color name is not found - use ansiSafe to check first.
// Supported names: black, red, green, yellow, blue, magenta, cyan, white,.
// brightblack, brightred, etc., and aliases like color0-color15, gray, purple, etc.
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func ansiFunc(data any, colorName string) colour.ColorValue {
	ph := extractPaletteHelper(data)
	cv, ok := ph.FindClosestANSIColor(colorName)
	if !ok {
		panic(fmt.Sprintf("ANSI color name %q not recognized", colorName))
	}
	return cv
}

// ansiSafeFunc finds the closest color to a given ANSI color name with error handling.
// Returns error if color name is not recognized.
// Accepts both *ThemeData and *PaletteHelper for backward compatibility.
func ansiSafeFunc(data any, colorName string) (colour.ColorValue, error) {
	ph := extractPaletteHelper(data)
	cv, ok := ph.FindClosestANSIColor(colorName)
	if !ok {
		return colour.ColorValue{}, fmt.Errorf("ANSI color name %q not recognized", colorName)
	}
	return cv, nil
}
