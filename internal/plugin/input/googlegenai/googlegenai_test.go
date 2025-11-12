package googlegenai

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/spf13/cobra"
)

// TestNew tests creating a new plugin with defaults.
func TestNew(t *testing.T) {
	plugin := New()

	if plugin == nil {
		t.Fatal("New() returned nil")
	}

	if plugin.Name() != "google-genai" {
		t.Errorf("Expected name 'google-genai', got '%s'", plugin.Name())
	}

	if plugin.model != defaultModel {
		t.Errorf("Expected default model '%s', got '%s'", defaultModel, plugin.model)
	}

	if plugin.aspectRatio != "16:9" {
		t.Errorf("Expected default aspect ratio '16:9', got '%s'", plugin.aspectRatio)
	}

	if plugin.imageSize != "2K" {
		t.Errorf("Expected default image size '2K', got '%s'", plugin.imageSize)
	}

	if plugin.backend != defaultBackend {
		t.Errorf("Expected default backend '%s', got '%s'", defaultBackend, plugin.backend)
	}

	if plugin.colours != 32 {
		t.Errorf("Expected default colours 32, got %d", plugin.colours)
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

	if !plugin.cacheEnabled {
		t.Error("Expected cacheEnabled to be true by default")
	}

	if plugin.cacheOverwrite {
		t.Error("Expected cacheOverwrite to be false by default")
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
	plugin := New()
	cmd := &cobra.Command{
		Use: "test",
	}

	plugin.RegisterFlags(cmd)

	// Check that flags were registered
	flags := []string{
		"prompt",
		"model",
		"aspect-ratio",
		"image-size",
		"negative-prompt",
		"genai-backend",
		"count",
		"extract-ambience",
		"regions",
		"sample-percent",
		"sample-method",
		"seed-mode",
		"seed-value",
		"cache",
		"cache-dir",
		"cache-filename",
		"cache-overwrite",
		"list-models",
		"no-extended-prompt",
		"no-negative-prompt",
	}

	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag '%s' not registered", flagName)
		}
	}
}

// TestValidateNoPrompt tests validation when prompt is missing.
func TestValidateNoPrompt(t *testing.T) {
	plugin := New()
	plugin.listModels = false

	err := plugin.Validate()
	if err == nil {
		t.Error("Expected error when prompt is empty")
	}
	if err.Error() != "prompt is required" {
		t.Errorf("Expected 'prompt is required' error, got '%v'", err)
	}
}

// TestValidateWithPrompt tests validation with a valid prompt.
func TestValidateWithPrompt(t *testing.T) {
	plugin := New()
	plugin.prompt = "a beautiful sunset"

	err := plugin.Validate()
	if err != nil {
		t.Errorf("Validation should pass with prompt, got error: %v", err)
	}
}

// TestValidateListModels tests validation when listing models.
func TestValidateListModels(t *testing.T) {
	plugin := New()
	plugin.listModels = true
	plugin.prompt = "" // Empty prompt should be OK when listing models

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
	for _, fh := range flagHelp {
		if fh.Name == "prompt" {
			foundPrompt = true
			if fh.Type != "string" {
				t.Errorf("Expected prompt type 'string', got '%s'", fh.Type)
			}
		}
		if fh.Name == "model" {
			foundModel = true
		}
	}

	if !foundPrompt {
		t.Error("GetFlagHelp should include 'prompt' flag")
	}
	if !foundModel {
		t.Error("GetFlagHelp should include 'model' flag")
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
		plugin := New()
		plugin.prompt = "test"
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

	plugin := New()
	plugin.prompt = "test"
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
