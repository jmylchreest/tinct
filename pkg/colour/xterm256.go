// Package colour provides public types and utilities for working with color palettes.
package colour

import (
	"fmt"
	"math"
)

// Xterm256CubeValues maps xterm-256 color cube axis indices (0-5) to RGB values.
// The 6x6x6 color cube occupies indices 16-231, where:
//
//	index = 16 + 36*r + 6*g + b  (r,g,b in 0-5)
var Xterm256CubeValues = [6]uint8{0, 95, 135, 175, 215, 255}

// NearestXterm256 finds the nearest xterm-256 color index for an RGB color.
// It considers both the 6x6x6 color cube (indices 16-231) and the
// grayscale ramp (indices 232-255). System colors 0-15 are intentionally
// skipped for terminal-independent behavior.
//
// The xterm-256 color space is composed of:
//   - Indices 0-15:    System colors (terminal-dependent, skipped)
//   - Indices 16-231:  6x6x6 color cube with axis values [0, 95, 135, 175, 215, 255]
//   - Indices 232-255: 24-step grayscale ramp from rgb(8,8,8) to rgb(238,238,238)
func NearestXterm256(r, g, b uint8) int {
	// Try the 6x6x6 color cube (indices 16-231).
	cubeR := nearestCubeIndex(r)
	cubeG := nearestCubeIndex(g)
	cubeB := nearestCubeIndex(b)
	cubeIdx := 16 + 36*cubeR + 6*cubeG + cubeB
	cubeDist := xterm256SqDist(r, g, b,
		Xterm256CubeValues[cubeR],
		Xterm256CubeValues[cubeG],
		Xterm256CubeValues[cubeB],
	)

	// Try the grayscale ramp (indices 232-255).
	// Grayscale values: 8, 18, 28, ..., 238 (24 shades, step of 10).
	// Use luminance-weighted average to find the best grayscale match.
	gray := int(math.Round(float64(r)*0.2126 + float64(g)*0.7152 + float64(b)*0.0722))
	grayIdx := max(232+int(math.Round(float64(gray-8)/10.0)), 232)
	if grayIdx > 255 {
		grayIdx = 255
	}
	grayVal := uint8(8 + (grayIdx-232)*10) // #nosec G115 -- grayIdx is clamped to [232,255] above, so result is [8,238], always fits in uint8
	grayDist := xterm256SqDist(r, g, b, grayVal, grayVal, grayVal)

	if grayDist < cubeDist {
		return grayIdx
	}
	return cubeIdx
}

// NearestXterm256String returns the nearest xterm-256 color as a string
// in the format "colorN" (e.g., "color196" for pure red).
// This is the format used by applications like Midnight Commander.
func NearestXterm256String(r, g, b uint8) string {
	return fmt.Sprintf("color%d", NearestXterm256(r, g, b))
}

// Xterm256ToRGB converts an xterm-256 color index to its RGB values.
// Returns (0,0,0) for system colors (0-15) since those are terminal-dependent.
func Xterm256ToRGB(index int) (r, g, b uint8) {
	switch {
	case index < 16:
		// System colors - terminal dependent, return black as fallback.
		return 0, 0, 0
	case index < 232:
		// 6x6x6 color cube.
		index -= 16
		b = Xterm256CubeValues[index%6]
		g = Xterm256CubeValues[(index/6)%6]
		r = Xterm256CubeValues[(index/36)%6]
		return r, g, b
	default:
		// Grayscale ramp (232-255).
		v := uint8(8 + (index-232)*10) // #nosec G115 -- index is in [232,255] in this branch, so result is [8,238], always fits in uint8
		return v, v, v
	}
}

// nearestCubeIndex finds the nearest index in the 6x6x6 cube for a single channel value.
func nearestCubeIndex(val uint8) int {
	best := 0
	bestDist := int(math.MaxInt32)
	for i, v := range Xterm256CubeValues {
		d := int(val) - int(v)
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist = d
			best = i
		}
	}
	return best
}

// xterm256SqDist computes the squared Euclidean distance between two RGB colors.
func xterm256SqDist(r1, g1, b1, r2, g2, b2 uint8) int {
	dr := int(r1) - int(r2)
	dg := int(g1) - int(g2)
	db := int(b1) - int(b2)
	return dr*dr + dg*dg + db*db
}
