// Package color provides colour extraction and palette generation functionality.
package colour

import (
	"fmt"
	"strings"
)

// ANSI escape codes for terminal colours.
const (
	ansiReset    = "\033[0m"
	ansiFgPrefix = "\033[38;2;"
	ansiBgPrefix = "\033[48;2;"
	ansiSuffix   = "m"
	defaultWidth = 8
)

// ColourPreview returns an ANSI-coloured preview string for a colour.
// Width specifies how many characters wide the colour block should be.
// Uses background colour with spaces for a solid block.
func ColourPreview(c RGB, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	// Build ANSI background colour escape sequence.
	bgColour := fmt.Sprintf("%s%d;%d;%d%s", ansiBgPrefix, c.R, c.G, c.B, ansiSuffix)

	// Create solid colour block using spaces with background colour.
	block := strings.Repeat(" ", width)

	return bgColour + block + ansiReset
}

// ColourPreviewWithText returns a colour preview with text overlay.
// The text colour is chosen to have good contrast with the background.
func ColourPreviewWithText(c RGB, text string, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	// Determine foreground colour for good contrast.
	lum := luminanceFromRGB(c)
	var fgR, fgG, fgB uint8
	if lum > 0.5 {
		// Light background, use dark text.
		fgR, fgG, fgB = 0, 0, 0
	} else {
		// Dark background, use light text.
		fgR, fgG, fgB = 255, 255, 255
	}

	// Build ANSI escape sequences.
	bgColour := fmt.Sprintf("%s%d;%d;%d%s", ansiBgPrefix, c.R, c.G, c.B, ansiSuffix)
	fgColour := fmt.Sprintf("%s%d;%d;%d%s", ansiFgPrefix, fgR, fgG, fgB, ansiSuffix)

	// Pad or truncate text to fit width.
	displayText := text
	if len(text) > width {
		displayText = text[:width]
	} else if len(text) < width {
		padding := (width - len(text)) / 2
		displayText = strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
	}

	return bgColour + fgColour + displayText + ansiReset
}

// luminanceFromRGB calculates approximate luminance from RGB (simplified version).
func luminanceFromRGB(rgb RGB) float64 {
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	// Simplified gamma correction.
	if r <= 0.03928 {
		r = r / 12.92
	} else {
		r = ((r + 0.055) / 1.055)
		r = r * r * r * r * r * r * r // Approximate pow(r, 2.4)
	}

	if g <= 0.03928 {
		g = g / 12.92
	} else {
		g = ((g + 0.055) / 1.055)
		g = g * g * g * g * g * g * g
	}

	if b <= 0.03928 {
		b = b / 12.92
	} else {
		b = ((b + 0.055) / 1.055)
		b = b * b * b * b * b * b * b
	}

	return 0.2126*r + 0.7152*g + 0.0722*b
}

// FormatColourWithPreview formats a colour with its preview and hex code.
func FormatColourWithPreview(rgb RGB, width int) string {
	preview := ColourPreview(rgb, width)
	return fmt.Sprintf("%s %s", preview, rgb.Hex())
}

// FormatColourWithLabel formats a colour with a label and preview.
func FormatColourWithLabel(rgb RGB, label string, width int) string {
	preview := ColourPreview(rgb, width)
	return fmt.Sprintf("%s  %-20s %s", preview, label, rgb.Hex())
}

// SupportsANSIColours checks if the terminal likely supports ANSI colour codes.
// This is a simple heuristic and may not be 100% accurate.
func SupportsANSIColours() bool {
	// Simple check - just return true for now.
	// In a more sophisticated version, we'd check TERM, COLORTERM env vars.
	// and whether output is a TTY
	return true
}

// DisableColourOutput can be used to disable colour output.
var DisableColourOutput = false

// ColourString returns a coloured string if colour output is enabled, plain text otherwise.
func ColourString(rgb RGB, text string) string {
	if DisableColourOutput || !SupportsANSIColours() {
		return text
	}

	fgColour := fmt.Sprintf("%s%d;%d;%d%s", ansiFgPrefix, rgb.R, rgb.G, rgb.B, ansiSuffix)
	return fgColour + text + ansiReset
}
