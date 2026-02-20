package colour

import (
	"fmt"
	"testing"
)

func TestNearestXterm256_KnownColors(t *testing.T) {
	tests := []struct {
		name     string
		r, g, b  uint8
		expected int
	}{
		// Cube exact matches.
		{"pure black", 0, 0, 0, 16},        // cube (0,0,0)
		{"pure white", 255, 255, 255, 231}, // cube (5,5,5)
		{"pure red", 255, 0, 0, 196},       // cube (5,0,0)
		{"pure green", 0, 255, 0, 46},      // cube (0,5,0)
		{"pure blue", 0, 0, 255, 21},       // cube (0,0,5)

		// Grayscale matches.
		{"mid gray", 128, 128, 128, 244}, // grayscale value 128, index 244

		// Cube color matches.
		{"orange-ish 224,175,104", 224, 175, 104, 179}, // cube (4,3,1) = 16+144+18+1=179
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NearestXterm256(tc.r, tc.g, tc.b)
			if got != tc.expected {
				t.Errorf("NearestXterm256(%d,%d,%d) = %d, want %d", tc.r, tc.g, tc.b, got, tc.expected)
			}
		})
	}
}

func TestNearestXterm256_ValidRange(t *testing.T) {
	// Test a wide spread of colors across the RGB space.
	for r := 0; r < 256; r += 51 {
		for g := 0; g < 256; g += 51 {
			for b := 0; b < 256; b += 51 {
				idx := NearestXterm256(uint8(r), uint8(g), uint8(b))
				if idx < 16 || idx > 255 {
					t.Errorf("NearestXterm256(%d,%d,%d) = %d, out of range [16,255]", r, g, b, idx)
				}
			}
		}
	}
}

func TestNearestXterm256_SkipsSystemColors(t *testing.T) {
	// Every possible RGB value should map to 16-255, never 0-15.
	for r := 0; r < 256; r += 17 {
		for g := 0; g < 256; g += 17 {
			for b := 0; b < 256; b += 17 {
				idx := NearestXterm256(uint8(r), uint8(g), uint8(b))
				if idx < 16 {
					t.Errorf("NearestXterm256(%d,%d,%d) = %d, should not use system colors 0-15", r, g, b, idx)
				}
			}
		}
	}
}

func TestNearestXterm256String(t *testing.T) {
	result := NearestXterm256String(255, 0, 0) // pure red
	expected := "color196"
	if result != expected {
		t.Errorf("NearestXterm256String(255,0,0) = %s, want %s", result, expected)
	}
}

func TestXterm256ToRGB_CubeRoundtrip(t *testing.T) {
	// All cube exact values should round-trip through NearestXterm256.
	for ri := range 6 {
		for gi := range 6 {
			for bi := range 6 {
				idx := 16 + 36*ri + 6*gi + bi
				r, g, b := Xterm256ToRGB(idx)
				// These exact cube RGB values should map back to the same index.
				gotIdx := NearestXterm256(r, g, b)
				if gotIdx != idx {
					t.Errorf("Round-trip failed: index %d -> RGB(%d,%d,%d) -> index %d",
						idx, r, g, b, gotIdx)
				}
			}
		}
	}
}

func TestXterm256ToRGB_Grayscale(t *testing.T) {
	// Grayscale indices 232-255 should produce rgb(v,v,v) where v = 8 + (idx-232)*10.
	for idx := 232; idx <= 255; idx++ {
		r, g, b := Xterm256ToRGB(idx)
		expectedVal := uint8(8 + (idx-232)*10)
		if r != expectedVal || g != expectedVal || b != expectedVal {
			t.Errorf("Xterm256ToRGB(%d) = (%d,%d,%d), want (%d,%d,%d)",
				idx, r, g, b, expectedVal, expectedVal, expectedVal)
		}
	}
}

func TestXterm256ToRGB_SystemColors(t *testing.T) {
	// System colors should return (0,0,0) as fallback.
	for idx := range 16 {
		r, g, b := Xterm256ToRGB(idx)
		if r != 0 || g != 0 || b != 0 {
			t.Errorf("Xterm256ToRGB(%d) = (%d,%d,%d), want (0,0,0) for system color",
				idx, r, g, b)
		}
	}
}

func TestNearestCubeIndex(t *testing.T) {
	tests := []struct {
		val      uint8
		expected int
	}{
		{0, 0},   // Exact match: cubeValues[0] = 0
		{95, 1},  // Exact match: cubeValues[1] = 95
		{135, 2}, // Exact match: cubeValues[2] = 135
		{175, 3}, // Exact match: cubeValues[3] = 175
		{215, 4}, // Exact match: cubeValues[4] = 215
		{255, 5}, // Exact match: cubeValues[5] = 255
		{47, 0},  // Midpoint between 0 and 95, closer to 0
		{48, 1},  // Just past midpoint, closer to 95
		{115, 1}, // Between 95 and 135: dist 20 each, tie -> first found = 1
		{154, 2}, // Between 135 and 175: 154-135=19 < 175-154=21, so 2
		{195, 3}, // Between 175 and 215: 195-175=20 = 215-195=20, tie -> first found = 3
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("val_%d", tc.val), func(t *testing.T) {
			got := nearestCubeIndex(tc.val)
			if got != tc.expected {
				t.Errorf("nearestCubeIndex(%d) = %d, want %d", tc.val, got, tc.expected)
			}
		})
	}
}
