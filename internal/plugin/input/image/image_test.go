package image

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/input"
)

// TestNew tests creating a new plugin with defaults.
func TestNew(t *testing.T) {
	plugin := New()

	if plugin == nil {
		t.Fatal("New() returned nil")
	}

	if plugin.Name() != "image" {
		t.Errorf("Expected name 'image', got '%s'", plugin.Name())
	}

	if plugin.colours != 16 {
		t.Errorf("Expected default colours 16, got %d", plugin.colours)
	}

	if plugin.extractAmbience {
		t.Error("Expected extractAmbience to be false by default")
	}

	if plugin.regions != 8 {
		t.Errorf("Expected default regions 8, got %d", plugin.regions)
	}

	if plugin.samplePercent != 10 {
		t.Errorf("Expected default samplePercent 10, got %d", plugin.samplePercent)
	}

	if plugin.sampleMethod != "average" {
		t.Errorf("Expected default sampleMethod 'average', got '%s'", plugin.sampleMethod)
	}

	if plugin.seedMode != "content" {
		t.Errorf("Expected default seedMode 'content', got '%s'", plugin.seedMode)
	}
}

// TestNewWithEnv tests creating plugin with environment variables.
func TestNewWithEnv(t *testing.T) {
	// Set environment variables.
	os.Setenv("TINCT_IMAGE_CACHE", "true")
	os.Setenv("TINCT_IMAGE_CACHE_DIR", "/tmp/test-cache")
	os.Setenv("TINCT_IMAGE_CACHE_FILENAME", "test.jpg")
	os.Setenv("TINCT_IMAGE_CACHE_OVERWRITE", "true")
	defer func() {
		os.Unsetenv("TINCT_IMAGE_CACHE")
		os.Unsetenv("TINCT_IMAGE_CACHE_DIR")
		os.Unsetenv("TINCT_IMAGE_CACHE_FILENAME")
		os.Unsetenv("TINCT_IMAGE_CACHE_OVERWRITE")
	}()

	plugin := New()

	if !plugin.cacheEnabled {
		t.Error("Expected cacheEnabled to be true from env")
	}

	if plugin.cacheDir != "/tmp/test-cache" {
		t.Errorf("Expected cacheDir '/tmp/test-cache', got '%s'", plugin.cacheDir)
	}

	if plugin.cacheFilename != "test.jpg" {
		t.Errorf("Expected cacheFilename 'test.jpg', got '%s'", plugin.cacheFilename)
	}

	if !plugin.cacheOverwrite {
		t.Error("Expected cacheOverwrite to be true from env")
	}
}

// TestName tests the Name method.
func TestName(t *testing.T) {
	plugin := New()
	if name := plugin.Name(); name != "image" {
		t.Errorf("Expected name 'image', got '%s'", name)
	}
}

// TestDescription tests the Description method.
func TestDescription(t *testing.T) {
	plugin := New()
	desc := plugin.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
}

// TestVersion tests the Version method.
func TestVersion(t *testing.T) {
	plugin := New()
	version := plugin.Version()
	if version == "" {
		t.Error("Version should not be empty")
	}
}

// TestRegisterFlags tests flag registration.
func TestRegisterFlags(t *testing.T) {
	plugin := New()
	cmd := &cobra.Command{
		Use: "test",
	}

	plugin.RegisterFlags(cmd)

	// Check that flags were registered (with image. prefix).
	flags := []string{
		"image.path",
		"image.colours",
		"image.extractAmbience",
		"image.regions",
		"image.sample-size",
		"image.sample-method",
		"image.seed-mode",
		"image.seed-value",
		"image.cache",
		"image.cache-dir",
		"image.cache-filename",
		"image.cache-overwrite",
	}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag '%s' not registered", flagName)
		}
	}
}

// TestValidateNoPath tests validation when path is missing.
func TestValidateNoPath(t *testing.T) {
	plugin := New()

	err := plugin.Validate()
	if err == nil {
		t.Error("Expected error when path is empty")
	}
}

// TestValidateWithPath tests that validation checks if path exists.
func TestValidateWithPath(t *testing.T) {
	// Note: We can't easily test successful validation without creating
	// a valid image file, so we just test that the path is checked.
	plugin := New()
	plugin.path = "/tmp/nonexistent.jpg"

	err := plugin.Validate()
	// Should fail because file doesn't exist or isn't a valid image.
	if err == nil {
		t.Error("Expected validation to fail for nonexistent file")
	}
}

// TestGetFlagHelp tests GetFlagHelp method.
func TestGetFlagHelp(t *testing.T) {
	plugin := New()
	flagHelp := plugin.GetFlagHelp()

	if len(flagHelp) == 0 {
		t.Error("GetFlagHelp should return flag information")
	}

	// Check for some key flags (with image. prefix).
	foundPath := false
	foundColours := false
	for _, fh := range flagHelp {
		if fh.Name == "image.path" {
			foundPath = true
			if fh.Type != "string" {
				t.Errorf("Expected path type 'string', got '%s'", fh.Type)
			}
		}
		if fh.Name == "image.colours" {
			foundColours = true
		}
	}

	if !foundPath {
		t.Error("GetFlagHelp should include 'image.path' flag")
	}
	if !foundColours {
		t.Error("GetFlagHelp should include 'image.colours' flag")
	}
}

// TestWallpaperPath tests the WallpaperPath method.
func TestWallpaperPath(t *testing.T) {
	plugin := New()

	// Initially empty.
	if path := plugin.WallpaperPath(); path != "" {
		t.Errorf("Expected empty wallpaper path initially, got '%s'", path)
	}

	// Set a path.
	testPath := "/tmp/test-image.jpg"
	plugin.loadedImagePath = testPath

	if path := plugin.WallpaperPath(); path != testPath {
		t.Errorf("Expected wallpaper path '%s', got '%s'", testPath, path)
	}
}

// TestGenerateRequiresPath tests that Generate requires a path.
func TestGenerateRequiresPath(t *testing.T) {
	plugin := New()
	// No path set.

	ctx := context.Background()
	opts := input.GenerateOptions{
		Verbose: false,
		DryRun:  false,
	}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error when path is not set")
	}
}

// TestGenerateNonExistentFile tests generating from non-existent file.
func TestGenerateNonExistentFile(t *testing.T) {
	plugin := New()
	plugin.path = "/nonexistent/file.jpg"

	ctx := context.Background()
	opts := input.GenerateOptions{
		Verbose: false,
		DryRun:  false,
	}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestSeedModeConfiguration tests seed mode configuration.
func TestSeedModeConfiguration(t *testing.T) {
	validModes := []string{"content", "filepath", "manual", "random"}

	for _, mode := range validModes {
		plugin := New()
		plugin.seedMode = mode

		if plugin.seedMode != mode {
			t.Errorf("Expected seedMode '%s', got '%s'", mode, plugin.seedMode)
		}
	}
}

// TestRegionsConfiguration tests regions count configuration.
func TestRegionsConfiguration(t *testing.T) {
	validRegions := []int{4, 8, 12, 16, 0}

	for _, regions := range validRegions {
		plugin := New()
		plugin.regions = regions
		plugin.extractAmbience = true

		if plugin.regions != regions {
			t.Errorf("Expected regions %d, got %d", regions, plugin.regions)
		}
	}
}

// TestSampleMethodConfiguration tests sample method configuration.
func TestSampleMethodConfiguration(t *testing.T) {
	validMethods := []string{"average", "dominant"}

	for _, method := range validMethods {
		plugin := New()
		plugin.sampleMethod = method

		if plugin.sampleMethod != method {
			t.Errorf("Expected sampleMethod '%s', got '%s'", method, plugin.sampleMethod)
		}
	}
}

// TestCachingConfiguration tests caching settings.
func TestCachingConfiguration(t *testing.T) {
	plugin := New()
	plugin.cacheEnabled = true
	plugin.cacheDir = "/tmp/cache"
	plugin.cacheFilename = "test.jpg"
	plugin.cacheOverwrite = true

	if !plugin.cacheEnabled {
		t.Error("Expected cacheEnabled to be true")
	}
	if plugin.cacheDir != "/tmp/cache" {
		t.Errorf("Expected cacheDir '/tmp/cache', got '%s'", plugin.cacheDir)
	}
	if plugin.cacheFilename != "test.jpg" {
		t.Errorf("Expected cacheFilename 'test.jpg', got '%s'", plugin.cacheFilename)
	}
	if !plugin.cacheOverwrite {
		t.Error("Expected cacheOverwrite to be true")
	}
}

// TestManualSeedValue tests manual seed configuration.
func TestManualSeedValue(t *testing.T) {
	plugin := New()
	plugin.seedMode = "manual"
	plugin.seedValue = 12345

	if plugin.seedMode != "manual" {
		t.Errorf("Expected seedMode 'manual', got '%s'", plugin.seedMode)
	}
	if plugin.seedValue != 12345 {
		t.Errorf("Expected seedValue 12345, got %d", plugin.seedValue)
	}
}

// TestExtractAmbienceConfiguration tests ambient extraction settings.
func TestExtractAmbienceConfiguration(t *testing.T) {
	plugin := New()
	plugin.extractAmbience = true
	plugin.regions = 12
	plugin.samplePercent = 15
	plugin.sampleMethod = "dominant"

	if !plugin.extractAmbience {
		t.Error("Expected extractAmbience to be true")
	}
	if plugin.regions != 12 {
		t.Errorf("Expected regions 12, got %d", plugin.regions)
	}
	if plugin.samplePercent != 15 {
		t.Errorf("Expected samplePercent 15, got %d", plugin.samplePercent)
	}
	if plugin.sampleMethod != "dominant" {
		t.Errorf("Expected sampleMethod 'dominant', got '%s'", plugin.sampleMethod)
	}
}

// TestGenerateWithValidImage tests successfully generating a palette from a valid image.
func TestGenerateWithValidImage(t *testing.T) {
	// Create a temporary directory.
	tempDir, err := os.MkdirTemp("", "tinct-image-tests-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test image with a few distinct colors.
	imagePath := filepath.Join(tempDir, "test.png")
	createTestImage(t, imagePath)

	plugin := New()
	plugin.path = imagePath

	ctx := context.Background()
	opts := input.GenerateOptions{
		Verbose: false,
		DryRun:  false,
		Backend: "kmeans",
	}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if palette == nil {
		t.Fatal("Generate() returned nil palette")
	}

	// Should have extracted some colors (default is 8).
	if len(palette.Colors) == 0 {
		t.Error("Generate() returned empty palette")
	}

	// Verify all colors are valid (non-nil).
	for i, c := range palette.Colors {
		if c == nil {
			t.Errorf("Color at index %d is nil", i)
		}
	}
}

// TestGenerateWithInvalidBackend tests error handling for invalid backend.
func TestGenerateWithInvalidBackend(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tinct-image-tests-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	imagePath := filepath.Join(tempDir, "test.png")
	createTestImage(t, imagePath)

	plugin := New()
	plugin.path = imagePath

	ctx := context.Background()
	opts := input.GenerateOptions{
		Backend: "invalid-backend",
	}

	_, err = plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error for invalid backend")
	}
}

// createTestImage creates a simple PNG image for testing with distinct colors.
func createTestImage(t *testing.T, path string) {
	t.Helper()

	// Create a 100x100 image with 4 quadrants of different colors.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Top-left: dark blue
	for y := range 50 {
		for x := range 50 {
			img.Set(x, y, color.RGBA{R: 26, G: 27, B: 38, A: 255})
		}
	}

	// Top-right: light blue
	for y := range 50 {
		for x := 50; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 122, G: 162, B: 247, A: 255})
		}
	}

	// Bottom-left: light gray
	for y := 50; y < 100; y++ {
		for x := range 50 {
			img.Set(x, y, color.RGBA{R: 192, G: 202, B: 245, A: 255})
		}
	}

	// Bottom-right: purple
	for y := 50; y < 100; y++ {
		for x := 50; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 187, G: 154, B: 247, A: 255})
		}
	}

	// Write PNG file.
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create image file: %v", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}
}
