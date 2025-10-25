// Package util provides shared utility functions used across the application.
package util

import (
	"strings"

	"github.com/jmylchreest/tinct/internal/colour"
)

// StripHash removes the # prefix from a hex colour string.
// This is useful for formats that don't expect the hash prefix.
func StripHash(hex string) string {
	return strings.TrimPrefix(hex, "#")
}

// GetColour retrieves a colour by role from a categorised palette with a fallback.
// This is a common helper used by output plugins to safely get colours with defaults.
func GetColour(palette *colour.CategorisedPalette, role colour.ColourRole, fallback string) string {
	if c, ok := palette.Get(role); ok {
		return c.Hex
	}
	return fallback
}
