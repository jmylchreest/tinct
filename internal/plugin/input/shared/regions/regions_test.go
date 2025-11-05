// Package regions provides utilities for extracting colors from specific
// regions of 2D images for ambient lighting applications.
package regions

import (
	"image"
	"image/color"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
)

// createTestImage creates a simple test image with distinct colors in each quadrant.
func createTestImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill quadrants with different colors
	// Top-Left: Red
	for y := 0; y < height/2; y++ {
		for x := 0; x < width/2; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Top-Right: Green
	for y := 0; y < height/2; y++ {
		for x := width / 2; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	// Bottom-Left: Blue
	for y := height / 2; y < height; y++ {
		for x := 0; x < width/2; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 0, B: 255, A: 255})
		}
	}

	// Bottom-Right: Yellow
	for y := height / 2; y < height; y++ {
		for x := width / 2; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 0, A: 255})
		}
	}

	return img
}

func TestNewSampler(t *testing.T) {
	sampler := NewSampler()

	if sampler.SamplePercent != 10 {
		t.Errorf("Expected default SamplePercent to be 10, got %d", sampler.SamplePercent)
	}

	if sampler.Method != "average" {
		t.Errorf("Expected default Method to be 'average', got %s", sampler.Method)
	}
}

func TestConfigurationFromInt(t *testing.T) {
	tests := []struct {
		input     int
		expected  Configuration
		shouldErr bool
	}{
		{4, Config4Regions, false},
		{8, Config8Regions, false},
		{12, Config12Regions, false},
		{16, Config16Regions, false},
		{6, 0, true},
		{0, 0, true},
		{20, 0, true},
	}

	for _, tt := range tests {
		config, err := ConfigurationFromInt(tt.input)

		if tt.shouldErr {
			if err == nil {
				t.Errorf("ConfigurationFromInt(%d) should have returned an error", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("ConfigurationFromInt(%d) returned unexpected error: %v", tt.input, err)
			}
			if config != tt.expected {
				t.Errorf("ConfigurationFromInt(%d) = %v, want %v", tt.input, config, tt.expected)
			}
		}
	}
}

func TestValidConfigurations(t *testing.T) {
	configs := ValidConfigurations()

	expected := []Configuration{Config4Regions, Config8Regions, Config12Regions, Config16Regions}

	if len(configs) != len(expected) {
		t.Errorf("ValidConfigurations() returned %d configs, want %d", len(configs), len(expected))
	}

	for i, config := range configs {
		if config != expected[i] {
			t.Errorf("ValidConfigurations()[%d] = %v, want %v", i, config, expected[i])
		}
	}
}

func TestExtract4Regions(t *testing.T) {
	img := createTestImage(400, 400)
	sampler := NewSampler()

	palette, err := sampler.Extract(img, Config4Regions)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	// Should extract 4 colors
	if len(palette.Colors) != 4 {
		t.Errorf("Expected 4 colors, got %d", len(palette.Colors))
	}

	// Should have 4 role hints
	if len(palette.RoleHints) != 4 {
		t.Errorf("Expected 4 role hints, got %d", len(palette.RoleHints))
	}

	// Check that expected roles are present
	expectedRoles := []colour.Role{
		colour.RolePositionTopLeft,
		colour.RolePositionTopRight,
		colour.RolePositionBottomRight,
		colour.RolePositionBottomLeft,
	}

	for _, role := range expectedRoles {
		if _, exists := palette.RoleHints[role]; !exists {
			t.Errorf("Missing role hint for %s", role)
		}
	}

	// Verify corner colors match expected quadrant colors
	// Top-Left should be red-ish
	tlIndex := palette.RoleHints[colour.RolePositionTopLeft]
	tlColor := palette.Colors[tlIndex]
	r, g, b, _ := tlColor.RGBA()
	if r>>8 < 200 || g>>8 > 50 || b>>8 > 50 {
		t.Errorf("Top-Left color should be red-ish, got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}

	// Top-Right should be green-ish
	trIndex := palette.RoleHints[colour.RolePositionTopRight]
	trColor := palette.Colors[trIndex]
	r, g, b, _ = trColor.RGBA()
	if r>>8 > 50 || g>>8 < 200 || b>>8 > 50 {
		t.Errorf("Top-Right color should be green-ish, got RGB(%d, %d, %d)", r>>8, g>>8, b>>8)
	}
}

func TestExtract8Regions(t *testing.T) {
	img := createTestImage(800, 600)
	sampler := NewSampler()

	palette, err := sampler.Extract(img, Config8Regions)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	// Should extract 8 colors
	if len(palette.Colors) != 8 {
		t.Errorf("Expected 8 colors, got %d", len(palette.Colors))
	}

	// Should have 8 role hints
	if len(palette.RoleHints) != 8 {
		t.Errorf("Expected 8 role hints, got %d", len(palette.RoleHints))
	}

	// Check that expected roles are present
	expectedRoles := []colour.Role{
		colour.RolePositionTopLeft,
		colour.RolePositionTop,
		colour.RolePositionTopRight,
		colour.RolePositionRight,
		colour.RolePositionBottomRight,
		colour.RolePositionBottom,
		colour.RolePositionBottomLeft,
		colour.RolePositionLeft,
	}

	for _, role := range expectedRoles {
		if _, exists := palette.RoleHints[role]; !exists {
			t.Errorf("Missing role hint for %s", role)
		}
	}
}

func TestExtract12Regions(t *testing.T) {
	img := createTestImage(1200, 800)
	sampler := NewSampler()

	palette, err := sampler.Extract(img, Config12Regions)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	if len(palette.Colors) != 12 {
		t.Errorf("Expected 12 colors, got %d", len(palette.Colors))
	}

	if len(palette.RoleHints) != 12 {
		t.Errorf("Expected 12 role hints, got %d", len(palette.RoleHints))
	}
}

func TestExtract16Regions(t *testing.T) {
	img := createTestImage(1600, 1200)
	sampler := NewSampler()

	palette, err := sampler.Extract(img, Config16Regions)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	if len(palette.Colors) != 16 {
		t.Errorf("Expected 16 colors, got %d", len(palette.Colors))
	}

	if len(palette.RoleHints) != 16 {
		t.Errorf("Expected 16 role hints, got %d", len(palette.RoleHints))
	}
}

func TestExtractInvalidConfiguration(t *testing.T) {
	img := createTestImage(400, 400)
	sampler := NewSampler()

	_, err := sampler.Extract(img, Configuration(7))
	if err == nil {
		t.Error("Extract() with invalid configuration should return error")
	}
}

func TestCustomSamplePercent(t *testing.T) {
	img := createTestImage(400, 400)

	sampler := &Sampler{
		SamplePercent: 20,
		Method:        "average",
	}

	palette, err := sampler.Extract(img, Config4Regions)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	if len(palette.Colors) != 4 {
		t.Errorf("Expected 4 colors, got %d", len(palette.Colors))
	}
}

func TestDominantMethod(t *testing.T) {
	img := createTestImage(400, 400)

	sampler := &Sampler{
		SamplePercent: 10,
		Method:        "dominant",
	}

	palette, err := sampler.Extract(img, Config4Regions)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	if len(palette.Colors) != 4 {
		t.Errorf("Expected 4 colors, got %d", len(palette.Colors))
	}

	// Dominant method should still extract reasonable colors
	for i, c := range palette.Colors {
		r, g, b, _ := c.RGBA()
		// Just verify we got valid RGB values
		if r == 0 && g == 0 && b == 0 {
			t.Errorf("Color %d is black, which is unexpected for this test image", i)
		}
	}
}

func TestSmallImage(t *testing.T) {
	// Test with a very small image
	img := createTestImage(50, 50)
	sampler := NewSampler()

	palette, err := sampler.Extract(img, Config8Regions)
	if err != nil {
		t.Fatalf("Extract() with small image returned error: %v", err)
	}

	// Should still extract 8 colors even with small image
	if len(palette.Colors) != 8 {
		t.Errorf("Expected 8 colors from small image, got %d", len(palette.Colors))
	}
}

func TestLargeImage(t *testing.T) {
	// Test with a larger image
	img := createTestImage(3840, 2160) // 4K resolution
	sampler := NewSampler()

	palette, err := sampler.Extract(img, Config16Regions)
	if err != nil {
		t.Fatalf("Extract() with large image returned error: %v", err)
	}

	if len(palette.Colors) != 16 {
		t.Errorf("Expected 16 colors from large image, got %d", len(palette.Colors))
	}
}

func TestAverageColorMethod(t *testing.T) {
	// Create image with solid red
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := range 100 {
		for x := range 100 {
			img.Set(x, y, color.RGBA{R: 200, G: 50, B: 75, A: 255})
		}
	}

	sampler := &Sampler{
		SamplePercent: 20,
		Method:        "average",
	}

	rect := image.Rect(0, 0, 100, 100)
	c := sampler.extractAverageColor(img, rect)

	r, g, b, _ := c.RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

	// Should be very close to the original color
	if r8 < 195 || r8 > 205 {
		t.Errorf("Red component should be ~200, got %d", r8)
	}
	if g8 < 45 || g8 > 55 {
		t.Errorf("Green component should be ~50, got %d", g8)
	}
	if b8 < 70 || b8 > 80 {
		t.Errorf("Blue component should be ~75, got %d", b8)
	}
}

func TestDominantColorMethod(t *testing.T) {
	// Create image with mostly red and a bit of blue
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := range 100 {
		for x := range 100 {
			if x < 90 {
				img.Set(x, y, color.RGBA{R: 200, G: 0, B: 0, A: 255})
			} else {
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 200, A: 255})
			}
		}
	}

	sampler := &Sampler{
		SamplePercent: 20,
		Method:        "dominant",
	}

	rect := image.Rect(0, 0, 100, 100)
	c := sampler.extractDominantColor(img, rect)

	r, g, b, _ := c.RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

	// Should be red (the dominant color)
	if r8 < 150 {
		t.Errorf("Red should be dominant, got R=%d", r8)
	}
	if g8 > 50 || b8 > 50 {
		t.Errorf("Green and Blue should be low for red, got G=%d, B=%d", g8, b8)
	}
}

func BenchmarkExtract4Regions(b *testing.B) {
	img := createTestImage(1920, 1080)
	sampler := NewSampler()

	for b.Loop() {
		_, _ = sampler.Extract(img, Config4Regions)
	}
}

func BenchmarkExtract8Regions(b *testing.B) {
	img := createTestImage(1920, 1080)
	sampler := NewSampler()

	for b.Loop() {
		_, _ = sampler.Extract(img, Config8Regions)
	}
}

func BenchmarkExtract12Regions(b *testing.B) {
	img := createTestImage(1920, 1080)
	sampler := NewSampler()

	for b.Loop() {
		_, _ = sampler.Extract(img, Config12Regions)
	}
}

func BenchmarkExtract16Regions(b *testing.B) {
	img := createTestImage(1920, 1080)
	sampler := NewSampler()

	for b.Loop() {
		_, _ = sampler.Extract(img, Config16Regions)
	}
}

func BenchmarkAverageMethod(b *testing.B) {
	img := createTestImage(1920, 1080)
	sampler := &Sampler{
		SamplePercent: 10,
		Method:        "average",
	}

	for b.Loop() {
		_, _ = sampler.Extract(img, Config8Regions)
	}
}

func BenchmarkDominantMethod(b *testing.B) {
	img := createTestImage(1920, 1080)
	sampler := &Sampler{
		SamplePercent: 10,
		Method:        "dominant",
	}

	for b.Loop() {
		_, _ = sampler.Extract(img, Config8Regions)
	}
}

func TestWeightsCalculation(t *testing.T) {
	img := createTestImage(800, 600)
	sampler := NewSampler()

	palette, err := sampler.Extract(img, Config4Regions)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	// Should have weights
	if palette.Weights == nil {
		t.Fatal("Expected weights to be set, got nil")
	}

	if len(palette.Weights) != len(palette.Colors) {
		t.Errorf("Expected %d weights, got %d", len(palette.Colors), len(palette.Weights))
	}

	// Weights should sum to 1.0 (or very close due to floating point)
	sum := 0.0
	for _, w := range palette.Weights {
		sum += w
		// Each weight should be positive
		if w <= 0 {
			t.Errorf("Expected positive weight, got %f", w)
		}
	}

	if sum < 0.999 || sum > 1.001 {
		t.Errorf("Expected weights to sum to ~1.0, got %f", sum)
	}
}

func TestWeightsReflectRegionSizes(t *testing.T) {
	img := createTestImage(1000, 1000)

	// Test with different sample percentages
	sampler1 := &Sampler{
		SamplePercent: 10,
		Method:        "average",
	}
	sampler2 := &Sampler{
		SamplePercent: 20,
		Method:        "average",
	}

	palette1, err := sampler1.Extract(img, Config4Regions)
	if err != nil {
		t.Fatalf("Extract() with 10%% sample returned error: %v", err)
	}

	palette2, err := sampler2.Extract(img, Config4Regions)
	if err != nil {
		t.Fatalf("Extract() with 20%% sample returned error: %v", err)
	}

	// Both should have weights
	if palette1.Weights == nil || palette2.Weights == nil {
		t.Fatal("Expected weights to be set")
	}

	// For 4 regions with equal-sized corners, weights should be approximately equal
	for i, w := range palette1.Weights {
		expected := 0.25 // Each of 4 corners should have ~25% weight
		if w < expected*0.9 || w > expected*1.1 {
			t.Errorf("Weight[%d] = %f, expected ~%f", i, w, expected)
		}
	}
}

func TestWeightsWithDifferentConfigurations(t *testing.T) {
	img := createTestImage(800, 600)
	sampler := NewSampler()

	configs := []struct {
		config Configuration
		count  int
	}{
		{Config4Regions, 4},
		{Config8Regions, 8},
		{Config12Regions, 12},
		{Config16Regions, 16},
	}

	for _, tc := range configs {
		palette, err := sampler.Extract(img, tc.config)
		if err != nil {
			t.Fatalf("Extract(%d) returned error: %v", tc.count, err)
		}

		if len(palette.Weights) != tc.count {
			t.Errorf("Config %d: expected %d weights, got %d", tc.count, tc.count, len(palette.Weights))
		}

		// Verify weights sum to 1.0
		sum := 0.0
		for _, w := range palette.Weights {
			sum += w
		}

		if sum < 0.999 || sum > 1.001 {
			t.Errorf("Config %d: weights sum to %f, expected ~1.0", tc.count, sum)
		}
	}
}
