package colour

import (
	"math"
	"strings"
)

// ANSIColor represents a standard ANSI terminal color name and its typical RGB value.
type ANSIColor struct {
	Name     string
	R, G, B  uint8
	Aliases  []string
	IsBright bool
}

// Standard ANSI color palette (xterm-256 basic 16 colors).
// These are the typical/common values - actual terminals may vary slightly.
var ansiColors = []ANSIColor{
	// Normal colors (0-7).
	{Name: "black", R: 0, G: 0, B: 0, Aliases: []string{"color0"}},
	{Name: "red", R: 205, G: 49, B: 49, Aliases: []string{"color1"}},
	{Name: "green", R: 13, G: 188, B: 121, Aliases: []string{"color2"}},
	{Name: "yellow", R: 229, G: 229, B: 16, Aliases: []string{"color3"}},
	{Name: "blue", R: 36, G: 114, B: 200, Aliases: []string{"color4"}},
	{Name: "magenta", R: 188, G: 63, B: 188, Aliases: []string{"color5", "purple"}},
	{Name: "cyan", R: 17, G: 168, B: 205, Aliases: []string{"color6"}},
	{Name: "white", R: 229, G: 229, B: 229, Aliases: []string{"color7", "gray", "grey"}},

	// Bright colors (8-15).
	{Name: "brightblack", R: 102, G: 102, B: 102, Aliases: []string{"color8", "darkgray", "darkgrey", "brightBlack"}, IsBright: true},
	{Name: "brightred", R: 241, G: 76, B: 76, Aliases: []string{"color9", "brightRed"}, IsBright: true},
	{Name: "brightgreen", R: 35, G: 209, B: 139, Aliases: []string{"color10", "brightGreen"}, IsBright: true},
	{Name: "brightyellow", R: 245, G: 245, B: 67, Aliases: []string{"color11", "brightYellow"}, IsBright: true},
	{Name: "brightblue", R: 59, G: 142, B: 234, Aliases: []string{"color12", "brightBlue"}, IsBright: true},
	{Name: "brightmagenta", R: 214, G: 112, B: 214, Aliases: []string{"color13", "brightpurple", "brightPurple", "brightMagenta"}, IsBright: true},
	{Name: "brightcyan", R: 41, G: 184, B: 219, Aliases: []string{"color14", "brightCyan"}, IsBright: true},
	{Name: "brightwhite", R: 255, G: 255, B: 255, Aliases: []string{"color15", "brightWhite"}, IsBright: true},

	// Additional common color names.
	{Name: "orange", R: 255, G: 165, B: 0, Aliases: []string{}},
	{Name: "pink", R: 255, G: 192, B: 203, Aliases: []string{}},
	{Name: "brown", R: 165, G: 42, B: 42, Aliases: []string{}},
	{Name: "lime", R: 0, G: 255, B: 0, Aliases: []string{}},
	{Name: "navy", R: 0, G: 0, B: 128, Aliases: []string{"darkblue"}},
	{Name: "teal", R: 0, G: 128, B: 128, Aliases: []string{"darkcyan"}},
	{Name: "maroon", R: 128, G: 0, B: 0, Aliases: []string{"darkred"}},
	{Name: "olive", R: 128, G: 128, B: 0, Aliases: []string{"darkyellow"}},
	{Name: "violet", R: 238, G: 130, B: 238, Aliases: []string{}},
	{Name: "indigo", R: 75, G: 0, B: 130, Aliases: []string{}},
}

// FindClosestANSIColor finds the color in the palette that most closely matches.
// the given ANSI color name using perceptual color distance (CIEDE2000-like).
// Returns the ColorValue and true if found, or ColorValue{} and false if not found.
//
// Supported color names include:.
//   - Standard ANSI: black, red, green, yellow, blue, magenta, cyan, white
//   - Bright variants: brightblack, brightred, brightgreen, etc.
//   - Aliases: color0-color15, gray, grey, purple, darkgray, etc.
//   - Extended: orange, pink, brown, lime, navy, teal, maroon, olive, violet, indigo
//
// Name matching is case-insensitive.
func (ph *PaletteHelper) FindClosestANSIColor(colorName string) (ColorValue, bool) {
	// Normalize color name (lowercase, remove spaces/dashes).
	normalizedName := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(colorName, " ", ""), "-", ""))

	// Find the ANSI color definition.
	var targetColor *ANSIColor
	for i := range ansiColors {
		if ansiColors[i].Name == normalizedName {
			targetColor = &ansiColors[i]
			break
		}
		// Check aliases.
		for _, alias := range ansiColors[i].Aliases {
			if strings.ToLower(alias) == normalizedName {
				targetColor = &ansiColors[i]
				break
			}
		}
		if targetColor != nil {
			break
		}
	}

	if targetColor == nil {
		return ColorValue{}, false
	}

	// Find closest color in palette using perceptual distance.
	var closestColor ColorValue
	minDistance := math.MaxFloat64

	for _, cv := range ph.AllColors() {
		distance := colorDistance(targetColor.R, targetColor.G, targetColor.B, cv.R(), cv.G(), cv.B())
		if distance < minDistance {
			minDistance = distance
			closestColor = cv
		}
	}

	return closestColor, true
}

// colorDistance calculates perceptual color distance using a simplified
// CIEDE2000-like formula (weighted Euclidean distance in RGB space).
// This is faster than full CIEDE2000 and good enough for color matching.
func colorDistance(r1, g1, b1, r2, g2, b2 uint8) float64 {
	// Convert to float64 for calculation.
	dr := float64(r1) - float64(r2)
	dg := float64(g1) - float64(g2)
	db := float64(b1) - float64(b2)

	// Weighted Euclidean distance (emphasizes green more like human perception).
	// Weights based on human color perception sensitivity.
	return math.Sqrt(2*dr*dr + 4*dg*dg + 3*db*db)
}

// GetANSIColorNames returns all supported ANSI color names.
// Useful for documentation and autocomplete.
func GetANSIColorNames() []string {
	names := make([]string, 0, len(ansiColors)*2)
	for _, ac := range ansiColors {
		names = append(names, ac.Name)
		names = append(names, ac.Aliases...)
	}
	return names
}

// ANSIColorInfo returns information about an ANSI color name.
// Returns the canonical name, RGB values, and whether it's a bright variant.
func ANSIColorInfo(colorName string) (name string, r, g, b uint8, isBright bool, found bool) {
	normalizedName := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(colorName, " ", ""), "-", ""))

	for i := range ansiColors {
		if ansiColors[i].Name == normalizedName {
			return ansiColors[i].Name, ansiColors[i].R, ansiColors[i].G, ansiColors[i].B, ansiColors[i].IsBright, true
		}
		for _, alias := range ansiColors[i].Aliases {
			if strings.ToLower(alias) == normalizedName {
				return ansiColors[i].Name, ansiColors[i].R, ansiColors[i].G, ansiColors[i].B, ansiColors[i].IsBright, true
			}
		}
	}

	return "", 0, 0, 0, false, false
}
