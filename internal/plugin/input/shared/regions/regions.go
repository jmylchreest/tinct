// Package regions provides utilities for extracting colors from specific.
// regions of 2D images for ambient lighting applications.
//
// This package supports extracting colors from image edges and corners in.
// various configurations (4, 8, 12, or 16 regions) to control LED lights.
// positioned around a monitor.
package regions

import (
	"fmt"
	"image"
	"image/color"

	"github.com/jmylchreest/tinct/internal/colour"
)

// Configuration represents the number of regions to extract.
type Configuration int

const (
	// Config4Regions extracts colors from 4 corners only.
	Config4Regions Configuration = 4

	// Config8Regions extracts colors from 4 corners + 4 mid-edges (default).
	Config8Regions Configuration = 8

	// Config12Regions extracts colors from 4 corners + 8 edge positions.
	Config12Regions Configuration = 12

	// Config16Regions extracts colors from 4 corners + 12 edge positions.
	Config16Regions Configuration = 16
)

// Position represents a specific sampling position around the image.
type Position struct {
	// Role is the semantic role for this position.
	Role colour.Role

	// Name is a human-readable name for this position.
	Name string

	// Rect is the sampling rectangle for this position.
	Rect image.Rectangle
}

// Sampler extracts colors from specific regions of an image.
type Sampler struct {
	// SamplePercent is the percentage of edge to sample (1-50).
	// Lower values = smaller sample regions, higher precision.
	// Higher values = larger sample regions, better averaging.
	// Default: 10.
	SamplePercent int

	// Method determines how to extract color from each region.
	// "average" = average all pixels (default)
	// "dominant" = most frequent color.
	Method string
}

// NewSampler creates a new region sampler with default settings.
func NewSampler() *Sampler {
	return &Sampler{
		SamplePercent: 10,
		Method:        "average",
	}
}

// Extract samples colors from the specified regions of an image.
// Returns a palette with colors, weights (based on region size), and role hints mapped to positions.
func (s *Sampler) Extract(img image.Image, config Configuration) (*colour.Palette, error) {
	// Validate configuration.
	if !isValidConfiguration(config) {
		return nil, fmt.Errorf("invalid configuration: %d (valid: 4, 8, 12, 16)", config)
	}

	// Define sampling positions for this configuration.
	positions := s.definePositions(img.Bounds(), config)

	// Extract color from each position.
	colors := make([]color.Color, len(positions))
	weights := make([]float64, len(positions))
	roleHints := make(map[colour.Role]int)

	totalPixels := 0
	pixelCounts := make([]int, len(positions))

	for i, pos := range positions {
		colors[i] = s.extractColorFromRegion(img, pos.Rect)
		roleHints[pos.Role] = i

		// Calculate pixel count for this region (for weight calculation).
		pixelCount := pos.Rect.Dx() * pos.Rect.Dy()
		pixelCounts[i] = pixelCount
		totalPixels += pixelCount
	}

	// Calculate weights based on region sizes.
	// Larger regions have more influence.
	for i := range weights {
		weights[i] = float64(pixelCounts[i]) / float64(totalPixels)
	}

	// Create palette with colors, weights, and role hints.
	palette := colour.NewPaletteWithWeights(colors, weights)
	palette.RoleHints = roleHints

	return palette, nil
}

// definePositions creates the sampling positions for the given configuration.
func (s *Sampler) definePositions(bounds image.Rectangle, config Configuration) []Position {
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate sample dimensions based on percentage.
	sampleWidth := width * s.SamplePercent / 100
	sampleHeight := height * s.SamplePercent / 100

	// Ensure minimum sample size.
	if sampleWidth < 10 {
		sampleWidth = 10
	}
	if sampleHeight < 10 {
		sampleHeight = 10
	}

	// Ensure samples don't exceed half the dimension.
	maxWidth := width / 2
	maxHeight := height / 2
	if sampleWidth > maxWidth {
		sampleWidth = maxWidth
	}
	if sampleHeight > maxHeight {
		sampleHeight = maxHeight
	}

	switch config {
	case Config4Regions:
		return s.define4Regions(bounds, sampleWidth, sampleHeight)
	case Config8Regions:
		return s.define8Regions(bounds, sampleWidth, sampleHeight)
	case Config12Regions:
		return s.define12Regions(bounds, sampleWidth, sampleHeight)
	case Config16Regions:
		return s.define16Regions(bounds, sampleWidth, sampleHeight)
	default:
		// Should never reach here due to validation.
		return nil
	}
}

// makePosition is a helper function to create a Position with role, name, and rectangle.
// This reduces duplication in the position definition methods.
func makePosition(role colour.Role, name string, rect image.Rectangle) Position {
	return Position{Role: role, Name: name, Rect: rect}
}

// cornerRect is a helper function to create a rectangle at one of the four corners.
// This encapsulates the corner positioning logic used across all configuration methods.
// Parameters: x0, y0 = bounds origin, w, h = dimensions, sw, sh = sample size.
func cornerRect(x0, y0, w, h, sw, sh int, corner string) image.Rectangle {
	switch corner {
	case "topLeft":
		return image.Rect(x0, y0, x0+sw, y0+sh)
	case "topRight":
		return image.Rect(x0+w-sw, y0, x0+w, y0+sh)
	case "bottomRight":
		return image.Rect(x0+w-sw, y0+h-sh, x0+w, y0+h)
	case "bottomLeft":
		return image.Rect(x0, y0+h-sh, x0+sw, y0+h)
	default:
		return image.Rectangle{}
	}
}

// edgeRect is a helper function to create a rectangle at a position along an edge.
// The fraction parameter (0.0-1.0) determines the position along the edge.
// For example, fraction=0.5 places the rectangle at the centre of the edge.
// This allows flexible positioning for 8, 12, and 16 region configurations.
// Parameters: x0, y0 = bounds origin, w, h = dimensions, sw, sh = sample size,.
//
//	edge = "top"|"bottom"|"left"|"right", fraction = position (0.0-1.0)
func edgeRect(x0, y0, w, h, sw, sh int, edge string, fraction float64) image.Rectangle {
	xPos := x0 + int(float64(w)*fraction) - sw/2
	yPos := y0 + int(float64(h)*fraction) - sh/2

	switch edge {
	case "top":
		return image.Rect(xPos, y0, xPos+sw, y0+sh)
	case "bottom":
		return image.Rect(xPos, y0+h-sh, xPos+sw, y0+h)
	case "left":
		return image.Rect(x0, yPos, x0+sw, yPos+sh)
	case "right":
		return image.Rect(x0+w-sw, yPos, x0+w, yPos+sh)
	default:
		return image.Rectangle{}
	}
}

// define4Regions creates 4 corner positions.
func (s *Sampler) define4Regions(bounds image.Rectangle, sw, sh int) []Position {
	w, h := bounds.Dx(), bounds.Dy()
	x0, y0 := bounds.Min.X, bounds.Min.Y

	return []Position{
		makePosition(colour.RolePositionTopLeft, "Top-Left", cornerRect(x0, y0, w, h, sw, sh, "topLeft")),
		makePosition(colour.RolePositionTopRight, "Top-Right", cornerRect(x0, y0, w, h, sw, sh, "topRight")),
		makePosition(colour.RolePositionBottomRight, "Bottom-Right", cornerRect(x0, y0, w, h, sw, sh, "bottomRight")),
		makePosition(colour.RolePositionBottomLeft, "Bottom-Left", cornerRect(x0, y0, w, h, sw, sh, "bottomLeft")),
	}
}

// define8Regions creates 4 corners + 4 mid-edge positions.
func (s *Sampler) define8Regions(bounds image.Rectangle, sw, sh int) []Position {
	w, h := bounds.Dx(), bounds.Dy()
	x0, y0 := bounds.Min.X, bounds.Min.Y

	return []Position{
		makePosition(colour.RolePositionTopLeft, "Top-Left", cornerRect(x0, y0, w, h, sw, sh, "topLeft")),
		makePosition(colour.RolePositionTop, "Top", edgeRect(x0, y0, w, h, sw, sh, "top", 0.5)),
		makePosition(colour.RolePositionTopRight, "Top-Right", cornerRect(x0, y0, w, h, sw, sh, "topRight")),
		makePosition(colour.RolePositionRight, "Right", edgeRect(x0, y0, w, h, sw, sh, "right", 0.5)),
		makePosition(colour.RolePositionBottomRight, "Bottom-Right", cornerRect(x0, y0, w, h, sw, sh, "bottomRight")),
		makePosition(colour.RolePositionBottom, "Bottom", edgeRect(x0, y0, w, h, sw, sh, "bottom", 0.5)),
		makePosition(colour.RolePositionBottomLeft, "Bottom-Left", cornerRect(x0, y0, w, h, sw, sh, "bottomLeft")),
		makePosition(colour.RolePositionLeft, "Left", edgeRect(x0, y0, w, h, sw, sh, "left", 0.5)),
	}
}

// define12Regions creates 4 corners + 8 edge positions (2 per edge).
func (s *Sampler) define12Regions(bounds image.Rectangle, sw, sh int) []Position {
	w, h := bounds.Dx(), bounds.Dy()
	x0, y0 := bounds.Min.X, bounds.Min.Y

	// For 12 regions, use slightly smaller samples to avoid overlap.
	sw = sw * 8 / 10
	sh = sh * 8 / 10
	if sw < 10 {
		sw = 10
	}
	if sh < 10 {
		sh = 10
	}

	return []Position{
		// Top edge: 4 positions.
		makePosition(colour.RolePositionTopLeft, "Top-Left", cornerRect(x0, y0, w, h, sw, sh, "topLeft")),
		makePosition(colour.RolePositionTopLeftInner, "Top-Left-Inner", edgeRect(x0, y0, w, h, sw, sh, "top", 1.0/3)),
		makePosition(colour.RolePositionTopRightInner, "Top-Right-Inner", edgeRect(x0, y0, w, h, sw, sh, "top", 2.0/3)),
		makePosition(colour.RolePositionTopRight, "Top-Right", cornerRect(x0, y0, w, h, sw, sh, "topRight")),

		// Right edge: 2 positions.
		makePosition(colour.RolePositionRightTop, "Right-Top", edgeRect(x0, y0, w, h, sw, sh, "right", 1.0/3)),
		makePosition(colour.RolePositionRightBottom, "Right-Bottom", edgeRect(x0, y0, w, h, sw, sh, "right", 2.0/3)),

		// Bottom edge: 4 positions.
		makePosition(colour.RolePositionBottomRight, "Bottom-Right", cornerRect(x0, y0, w, h, sw, sh, "bottomRight")),
		makePosition(colour.RolePositionBottomRightInner, "Bottom-Right-Inner", edgeRect(x0, y0, w, h, sw, sh, "bottom", 2.0/3)),
		makePosition(colour.RolePositionBottomLeftInner, "Bottom-Left-Inner", edgeRect(x0, y0, w, h, sw, sh, "bottom", 1.0/3)),
		makePosition(colour.RolePositionBottomLeft, "Bottom-Left", cornerRect(x0, y0, w, h, sw, sh, "bottomLeft")),

		// Left edge: 2 positions.
		makePosition(colour.RolePositionLeftBottom, "Left-Bottom", edgeRect(x0, y0, w, h, sw, sh, "left", 2.0/3)),
		makePosition(colour.RolePositionLeftTop, "Left-Top", edgeRect(x0, y0, w, h, sw, sh, "left", 1.0/3)),
	}
}

// define16Regions creates 4 corners + 12 edge positions (3 per edge).
func (s *Sampler) define16Regions(bounds image.Rectangle, sw, sh int) []Position {
	w, h := bounds.Dx(), bounds.Dy()
	x0, y0 := bounds.Min.X, bounds.Min.Y

	// For 16 regions, use smaller samples to avoid overlap.
	sw = sw * 6 / 10
	sh = sh * 6 / 10
	if sw < 10 {
		sw = 10
	}
	if sh < 10 {
		sh = 10
	}

	return []Position{
		// Top edge: 5 positions.
		makePosition(colour.RolePositionTopLeft, "Top-Left", cornerRect(x0, y0, w, h, sw, sh, "topLeft")),
		makePosition(colour.RolePositionTopLeftCenter, "Top-Left-Center", edgeRect(x0, y0, w, h, sw, sh, "top", 0.25)),
		makePosition(colour.RolePositionTopCenter, "Top-Center", edgeRect(x0, y0, w, h, sw, sh, "top", 0.5)),
		makePosition(colour.RolePositionTopRightCenter, "Top-Right-Center", edgeRect(x0, y0, w, h, sw, sh, "top", 0.75)),
		makePosition(colour.RolePositionTopRight, "Top-Right", cornerRect(x0, y0, w, h, sw, sh, "topRight")),

		// Right edge: 3 positions.
		makePosition(colour.RolePositionRightTopOuter, "Right-Top-Outer", edgeRect(x0, y0, w, h, sw, sh, "right", 0.25)),
		makePosition(colour.RolePositionRightTop, "Right-Top", edgeRect(x0, y0, w, h, sw, sh, "right", 0.5)),
		makePosition(colour.RolePositionRightBottomOuter, "Right-Bottom-Outer", edgeRect(x0, y0, w, h, sw, sh, "right", 0.75)),

		// Bottom edge: 5 positions.
		makePosition(colour.RolePositionBottomRight, "Bottom-Right", cornerRect(x0, y0, w, h, sw, sh, "bottomRight")),
		makePosition(colour.RolePositionBottomRightCenter, "Bottom-Right-Center", edgeRect(x0, y0, w, h, sw, sh, "bottom", 0.75)),
		makePosition(colour.RolePositionBottomCenter, "Bottom-Center", edgeRect(x0, y0, w, h, sw, sh, "bottom", 0.5)),
		makePosition(colour.RolePositionBottomLeftCenter, "Bottom-Left-Center", edgeRect(x0, y0, w, h, sw, sh, "bottom", 0.25)),
		makePosition(colour.RolePositionBottomLeft, "Bottom-Left", cornerRect(x0, y0, w, h, sw, sh, "bottomLeft")),

		// Left edge: 3 positions.
		makePosition(colour.RolePositionLeftBottomOuter, "Left-Bottom-Outer", edgeRect(x0, y0, w, h, sw, sh, "left", 0.75)),
		makePosition(colour.RolePositionLeftBottom, "Left-Bottom", edgeRect(x0, y0, w, h, sw, sh, "left", 0.5)),
		makePosition(colour.RolePositionLeftTopOuter, "Left-Top-Outer", edgeRect(x0, y0, w, h, sw, sh, "left", 0.25)),
	}
}

// extractColorFromRegion extracts a single color from the specified region.
func (s *Sampler) extractColorFromRegion(img image.Image, rect image.Rectangle) color.Color {
	switch s.Method {
	case "dominant":
		return s.extractDominantColor(img, rect)
	default:
		return s.extractAverageColor(img, rect)
	}
}

// extractAverageColor calculates the average color of all pixels in a region.
func (s *Sampler) extractAverageColor(img image.Image, rect image.Rectangle) color.Color {
	var totalR, totalG, totalB uint64
	var count uint64

	// Sample every pixel in the region.
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// RGBA() returns values in range [0, 65535], convert to [0, 255].
			totalR += uint64(r >> 8)
			totalG += uint64(g >> 8)
			totalB += uint64(b >> 8)
			count++
		}
	}

	if count == 0 {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}

	// Calculate averages with bounds checking.
	avgR := uint8(min(totalR/count, 255)) // #nosec G115 - min() ensures value is <= 255
	avgG := uint8(min(totalG/count, 255)) // #nosec G115 - min() ensures value is <= 255
	avgB := uint8(min(totalB/count, 255)) // #nosec G115 - min() ensures value is <= 255

	return color.RGBA{R: avgR, G: avgG, B: avgB, A: 255}
}

// extractDominantColor finds the most frequent color in a region.
// Colors are quantized to reduce the number of unique colors.
func (s *Sampler) extractDominantColor(img image.Image, rect image.Rectangle) color.Color {
	// Map to count color frequencies (quantized to reduce unique colors).
	colorCounts := make(map[uint32]int)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Quantize to 5-bit (32 values) per channel to reduce unique colors.
			// Safe conversion: (r >> 8) is already in 0-255 range, & 0xF8 keeps it there
			r8 := uint8(min((r>>8)&0xF8, 255)) // #nosec G115 - value masked to 8 bits and bounded by min
			g8 := uint8(min((g>>8)&0xF8, 255)) // #nosec G115 - value masked to 8 bits and bounded by min
			b8 := uint8(min((b>>8)&0xF8, 255)) // #nosec G115 - value masked to 8 bits and bounded by min

			// Pack into single uint32 for map key.
			packed := uint32(r8)<<16 | uint32(g8)<<8 | uint32(b8)
			colorCounts[packed]++
		}
	}

	// Find most frequent color.
	var maxCount int
	var dominantColor uint32
	for c, count := range colorCounts {
		if count > maxCount {
			maxCount = count
			dominantColor = c
		}
	}

	// Unpack the color with safe conversion.
	// dominantColor is already masked to 24-bit RGB, so shifts are safe
	r := uint8((dominantColor >> 16) & 0xFF) // #nosec G115 - value explicitly masked to 8 bits
	g := uint8((dominantColor >> 8) & 0xFF)  // #nosec G115 - value explicitly masked to 8 bits
	b := uint8(dominantColor & 0xFF)         // #nosec G115 - value explicitly masked to 8 bits

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// isValidConfiguration checks if the given configuration is supported.
func isValidConfiguration(config Configuration) bool {
	switch config {
	case Config4Regions, Config8Regions, Config12Regions, Config16Regions:
		return true
	default:
		return false
	}
}

// ValidConfigurations returns a list of valid configuration values.
func ValidConfigurations() []Configuration {
	return []Configuration{
		Config4Regions,
		Config8Regions,
		Config12Regions,
		Config16Regions,
	}
}

// ConfigurationFromInt converts an integer to a Configuration.
// Returns an error if the value is not a valid configuration.
func ConfigurationFromInt(n int) (Configuration, error) {
	config := Configuration(n)
	if !isValidConfiguration(config) {
		return 0, fmt.Errorf("invalid configuration: %d (valid: 4, 8, 12, 16)", n)
	}
	return config, nil
}
