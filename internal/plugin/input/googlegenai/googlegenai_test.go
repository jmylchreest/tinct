package googlegenai

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
	"github.com/spf13/cobra"
)

// resetFlags resets all shared flags to their defaults before each test.
func resetFlags() {
	aiflags.Reset()
	commonflags.Reset()
}

// TestNew tests creating a new plugin with defaults.
func TestNew(t *testing.T) {
	plugin := New()

	if plugin == nil {
		t.Fatal("New() returned nil")
	}

	if plugin.Name() != "google-genai" {
		t.Errorf("Expected name 'google-genai', got '%s'", plugin.Name())
	}

	// Check plugin-specific defaults
	if plugin.backend != defaultBackend {
		t.Errorf("Expected default backend '%s', got '%s'", defaultBackend, plugin.backend)
	}

	if plugin.imageSize != "2K" {
		t.Errorf("Expected default image size '2K', got '%s'", plugin.imageSize)
	}
}

// TestName tests the Name method.
func TestName(t *testing.T) {
	plugin := New()
	if name := plugin.Name(); name != "google-genai" {
		t.Errorf("Expected name 'google-genai', got '%s'", name)
	}
}

// TestDescription tests the Description method.
func TestDescription(t *testing.T) {
	plugin := New()
	desc := plugin.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if desc != "Generate images with Google Imagen and extract colours" {
		t.Errorf("Unexpected description: %s", desc)
	}
}

// TestVersion tests the Version method.
func TestVersion(t *testing.T) {
	plugin := New()
	version := plugin.Version()
	if version == "" {
		t.Error("Version should not be empty")
	}
	// Check it matches semver format (X.Y.Z)
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+$`, version)
	if !matched {
		t.Errorf("Version '%s' does not follow semver format (X.Y.Z)", version)
	}
}

// TestRegisterFlags tests flag registration.
func TestRegisterFlags(t *testing.T) {
	resetFlags()
	plugin := New()
	cmd := &cobra.Command{
		Use: "test",
	}

	plugin.RegisterFlags(cmd)

	// Check that shared AI flags were registered
	aiFlags := []string{
		"ai.prompt",
		"ai.model",
		"ai.list-models",
		"ai.no-extended-prompt",
		"ai.no-negative-prompt",
		"ai.negative-prompt",
	}

	for _, flagName := range aiFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Shared AI flag '%s' not registered", flagName)
		}
	}

	// Check that shared common flags were registered (unprefixed)
	commonFlags := []string{
		"count",
		"aspect-ratio",
		"cache",
		"cache-dir",
		"cache-filename",
		"cache-overwrite",
		"extract-ambience",
		"regions",
		"sample-percent",
		"sample-method",
		"seed-mode",
		"seed-value",
	}

	for _, flagName := range commonFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Shared common flag '%s' not registered", flagName)
		}
	}

	// Check that plugin-specific flags were registered
	pluginFlags := []string{
		"googlegenai.backend",
		"googlegenai.image-size",
	}

	for _, flagName := range pluginFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Plugin-specific flag '%s' not registered", flagName)
		}
	}
}

// TestValidateNoPrompt tests validation when prompt is missing.
func TestValidateNoPrompt(t *testing.T) {
	resetFlags()
	plugin := New()
	aiflags.ListModels = false
	aiflags.Prompt = ""

	err := plugin.Validate()
	if err == nil {
		t.Error("Expected error when prompt is empty")
	}
	if err.Error() != "--ai.prompt is required" {
		t.Errorf("Expected '--ai.prompt is required' error, got '%v'", err)
	}
}

// TestValidateWithPrompt tests validation with a valid prompt.
func TestValidateWithPrompt(t *testing.T) {
	resetFlags()
	plugin := New()
	aiflags.Prompt = "a beautiful sunset"

	err := plugin.Validate()
	if err != nil {
		t.Errorf("Validation should pass with prompt, got error: %v", err)
	}
}

// TestValidateListModels tests validation when listing models.
func TestValidateListModels(t *testing.T) {
	resetFlags()
	plugin := New()
	aiflags.ListModels = true
	aiflags.Prompt = "" // Empty prompt should be OK when listing models

	err := plugin.Validate()
	if err != nil {
		t.Errorf("Validation should pass when listing models, got error: %v", err)
	}
}

// TestGetFlagHelp tests GetFlagHelp method.
func TestGetFlagHelp(t *testing.T) {
	plugin := New()
	flagHelp := plugin.GetFlagHelp()

	if len(flagHelp) == 0 {
		t.Error("GetFlagHelp should return flag information")
	}

	// Check for some key flags
	foundPrompt := false
	foundModel := false
	foundBackend := false
	foundCount := false
	for _, fh := range flagHelp {
		if fh.Name == "ai.prompt" {
			foundPrompt = true
			if fh.Type != "string" {
				t.Errorf("Expected prompt type 'string', got '%s'", fh.Type)
			}
		}
		if fh.Name == "ai.model" {
			foundModel = true
		}
		if fh.Name == "googlegenai.backend" {
			foundBackend = true
		}
		if fh.Name == "count" {
			foundCount = true
		}
	}

	if !foundPrompt {
		t.Error("GetFlagHelp should include 'ai.prompt' flag")
	}
	if !foundModel {
		t.Error("GetFlagHelp should include 'ai.model' flag")
	}
	if !foundBackend {
		t.Error("GetFlagHelp should include 'googlegenai.backend' flag")
	}
	if !foundCount {
		t.Error("GetFlagHelp should include 'count' flag")
	}
}

// TestWallpaperPath tests the WallpaperPath method.
func TestWallpaperPath(t *testing.T) {
	plugin := New()

	// Initially empty
	if path := plugin.WallpaperPath(); path != "" {
		t.Errorf("Expected empty wallpaper path initially, got '%s'", path)
	}

	// Set a path (using exported field access through Generate would be needed in real scenario)
	testPath := "/tmp/test-image.jpg"
	plugin.loadedImagePath = testPath

	if path := plugin.WallpaperPath(); path != testPath {
		t.Errorf("Expected wallpaper path '%s', got '%s'", testPath, path)
	}
}

// TestValidateValidBackends tests validation with valid backends.
func TestValidateValidBackends(t *testing.T) {
	validBackends := []string{"gemini-api", "vertex-ai"}

	for _, backend := range validBackends {
		resetFlags()
		plugin := New()
		aiflags.Prompt = "test"
		plugin.backend = backend

		err := plugin.Validate()
		if err != nil {
			t.Errorf("Expected validation to pass for backend '%s', got error: %v", backend, err)
		}
	}
}

// TestGenerateRequiresAPIKey tests that Generate requires API key.
func TestGenerateRequiresAPIKey(t *testing.T) {
	// Skip if API key is set (to avoid actual API calls)
	if os.Getenv("GOOGLE_API_KEY") != "" {
		t.Skip("Skipping test because GOOGLE_API_KEY is set")
	}

	resetFlags()
	plugin := New()
	aiflags.Prompt = "test"
	plugin.backend = "gemini-api"

	ctx := context.Background()
	opts := input.GenerateOptions{
		Verbose: false,
		DryRun:  false,
	}

	_, err := plugin.Generate(ctx, opts)
	if err == nil {
		t.Error("Expected error when API key is not set")
	}
}

// TestGenerateDryRun tests dry-run mode.
func TestGenerateDryRun(t *testing.T) {
	resetFlags()
	plugin := New()
	aiflags.Prompt = "test prompt"

	ctx := context.Background()
	opts := input.GenerateOptions{
		Verbose: false,
		DryRun:  true,
	}

	palette, err := plugin.Generate(ctx, opts)
	if err != nil {
		t.Errorf("DryRun should not return error, got: %v", err)
	}
	if palette == nil {
		t.Error("DryRun should return a palette")
	}
	if len(palette.Colors) != 0 {
		t.Errorf("DryRun should return empty palette, got %d colors", len(palette.Colors))
	}
}

// TestEnhancePromptForWallpaper tests prompt enhancement.
func TestEnhancePromptForWallpaper(t *testing.T) {
	resetFlags()
	plugin := New()
	basePrompt := "a mountain landscape"

	// With enhancement enabled (default)
	aiflags.NoExtendedPrompt = false
	enhanced := plugin.enhancePromptForWallpaper(basePrompt)
	if enhanced == basePrompt {
		t.Error("Expected prompt to be enhanced by default")
	}
	if len(enhanced) <= len(basePrompt) {
		t.Error("Enhanced prompt should be longer than base prompt")
	}

	// With enhancement disabled
	aiflags.NoExtendedPrompt = true
	notEnhanced := plugin.enhancePromptForWallpaper(basePrompt)
	if notEnhanced != basePrompt {
		t.Errorf("Expected prompt to not be enhanced when disabled, got '%s'", notEnhanced)
	}
}

// TestBuildNegativePrompt tests negative prompt building.
func TestBuildNegativePrompt(t *testing.T) {
	// With negative prompt enabled (empty user prompt)
	negPrompt := buildNegativePrompt("", false)
	if negPrompt == "" {
		t.Error("Expected non-empty negative prompt when enabled")
	}
	if negPrompt != defaultNegativePrompt {
		t.Errorf("Expected default negative prompt, got '%s'", negPrompt)
	}

	// With user-provided negative prompt
	userPrompt := "ugly, bad quality"
	combined := buildNegativePrompt(userPrompt, false)
	if combined == "" {
		t.Error("Expected non-empty combined negative prompt")
	}
	if combined != userPrompt+", "+defaultNegativePrompt {
		t.Errorf("Expected combined prompt, got '%s'", combined)
	}

	// With negative prompt disabled
	disabledPrompt := buildNegativePrompt("", true)
	if disabledPrompt != "" {
		t.Errorf("Expected empty negative prompt when disabled, got '%s'", disabledPrompt)
	}
}

// TestDefaultModelValue tests that the default ai.model value is "auto".
func TestDefaultModelValue(t *testing.T) {
	resetFlags()
	if aiflags.Model != "auto" {
		t.Errorf("Expected default model to be 'auto', got '%s'", aiflags.Model)
	}
}

// TestCommonFlagsDefaults tests that common flags have correct defaults.
func TestCommonFlagsDefaults(t *testing.T) {
	resetFlags()

	if commonflags.Count != 32 {
		t.Errorf("Expected default count 32, got %d", commonflags.Count)
	}

	if commonflags.AspectRatio != "16:9" {
		t.Errorf("Expected default aspect ratio '16:9', got '%s'", commonflags.AspectRatio)
	}

	if !commonflags.CacheEnabled {
		t.Error("Expected cacheEnabled to be true by default")
	}

	if commonflags.CacheOverwrite {
		t.Error("Expected cacheOverwrite to be false by default")
	}

	if commonflags.ExtractAmbience {
		t.Error("Expected extractAmbience to be false by default")
	}

	if commonflags.Regions != 8 {
		t.Errorf("Expected default regions 8, got %d", commonflags.Regions)
	}

	if commonflags.SamplePercent != 10 {
		t.Errorf("Expected default samplePercent 10, got %d", commonflags.SamplePercent)
	}

	if commonflags.SampleMethod != "average" {
		t.Errorf("Expected default sampleMethod 'average', got '%s'", commonflags.SampleMethod)
	}

	if commonflags.SeedMode != "content" {
		t.Errorf("Expected default seedMode 'content', got '%s'", commonflags.SeedMode)
	}
}

// TestIsGeminiModel tests the isGeminiModel function.
func TestIsGeminiModel(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"gemini-2.5-flash-image", true},
		{"imagen-4.0-generate-001", false},
		{"imagen-4.0-fast-generate-001", false},
		{"imagen-3.0-generate-002", false},
	}

	for _, tt := range tests {
		result := isGeminiModel(tt.model)
		if result != tt.expected {
			t.Errorf("isGeminiModel(%q) = %v, expected %v", tt.model, result, tt.expected)
		}
	}
}

// TestIsImageGenerationModel tests the isImageGenerationModel function.
func TestIsImageGenerationModel(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"models/imagen-4.0-generate-001", true},
		{"models/imagen-3.0-generate-002", true},
		{"models/gemini-2.5-flash-image", true},
		{"models/gemini-pro", false},
		{"models/text-bison", false},
		{"imagen-4.0-fast-generate-001", true},
	}

	for _, tt := range tests {
		result := isImageGenerationModel(tt.name)
		if result != tt.expected {
			t.Errorf("isImageGenerationModel(%q) = %v, expected %v", tt.name, result, tt.expected)
		}
	}
}
