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

// Preview returns an ANSI-coloured preview string for a colour.
// Width specifies how many characters wide the colour block should be.
// Uses background colour with spaces for a solid block.
func Preview(c RGB, width int) string {
	if width <= 0 {
		width = defaultWidth
	}

	// Build ANSI background colour escape sequence.
	bgColour := fmt.Sprintf("%s%d;%d;%d%s", ansiBgPrefix, c.R, c.G, c.B, ansiSuffix)

	// Create solid colour block using spaces with background colour.
	block := strings.Repeat(" ", width)

	return bgColour + block + ansiReset
}

// FormatColourWithPreview formats a colour with its preview and hex code.
func FormatColourWithPreview(rgb RGB, width int) string {
	preview := Preview(rgb, width)
	return fmt.Sprintf("%s %s", preview, rgb.Hex())
}
