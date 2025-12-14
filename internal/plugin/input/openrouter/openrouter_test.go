package openrouter

import (
	"context"
	"os"
	"regexp"
	"strings"
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

	if plugin.Name() != "openrouter" {
		t.Errorf("Expected name 'openrouter', got '%s'", plugin.Name())
	}

	// Check plugin-specific default
	if !plugin.preferFree {
		t.Error("Expected preferFree to be true by default")
	}
}

// TestName tests the Name method.
func TestName(t *testing.T) {
	plugin := New()
	if name := plugin.Name(); name != "openrouter" {
		t.Errorf("Expected name 'openrouter', got '%s'", name)
	}
}

// TestDescription tests the Description method.
func TestDescription(t *testing.T) {
	plugin := New()
	desc := plugin.Description()
	if desc == "" {
		t.Error("Description should not be empty")
	}
	if desc != "Generate images with OpenRouter.ai models and extract colours" {
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
		"openrouter.prefer-free",
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
	foundPreferFree := false
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
		if fh.Name == "openrouter.prefer-free" {
			foundPreferFree = true
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
	if !foundPreferFree {
		t.Error("GetFlagHelp should include 'openrouter.prefer-free' flag")
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

	// Set a path
	testPath := "/tmp/test-image.jpg"
	plugin.loadedImagePath = testPath

	if path := plugin.WallpaperPath(); path != testPath {
		t.Errorf("Expected wallpaper path '%s', got '%s'", testPath, path)
	}
}

// TestGenerateRequiresAPIKey tests that Generate requires API key.
func TestGenerateRequiresAPIKey(t *testing.T) {
	// Skip if API key is set (to avoid actual API calls)
	if os.Getenv("OPENROUTER_API_KEY") != "" {
		t.Skip("Skipping test because OPENROUTER_API_KEY is set")
	}

	resetFlags()
	plugin := New()
	aiflags.Prompt = "test"
	aiflags.Model = "test-model" // Not "auto" to skip model fetch

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
	aiflags.Model = "test-model" // Not "auto" to skip model fetch

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
	// With negative prompt enabled
	negPrompt := buildNegativePrompt(false)
	if negPrompt == "" {
		t.Error("Expected non-empty negative prompt when enabled")
	}
	if negPrompt != defaultNegativePrompt {
		t.Errorf("Expected default negative prompt, got '%s'", negPrompt)
	}

	// With negative prompt disabled
	disabledPrompt := buildNegativePrompt(true)
	if disabledPrompt != "" {
		t.Errorf("Expected empty negative prompt when disabled, got '%s'", disabledPrompt)
	}
}

// TestParsePrice tests the parsePrice function.
func TestParsePrice(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"", 0},
		{"0", 0},
		{"0.001", 0.001},
		{"0.00001", 0.00001},
		{"1.5", 1.5},
	}

	for _, tt := range tests {
		result := parsePrice(tt.input)
		if result != tt.expected {
			t.Errorf("parsePrice(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

// TestGetPricingInfo tests the getPricingInfo function.
func TestGetPricingInfo(t *testing.T) {
	tests := []struct {
		name         string
		pricing      Pricing
		expectedType PricingType
	}{
		{
			name:         "free model",
			pricing:      Pricing{},
			expectedType: PricingFree,
		},
		{
			name:         "per-image pricing",
			pricing:      Pricing{Image: "0.02"},
			expectedType: PricingPerImage,
		},
		{
			name:         "per-request pricing",
			pricing:      Pricing{Request: "0.01"},
			expectedType: PricingPerRequest,
		},
		{
			name:         "per-token pricing",
			pricing:      Pricing{Prompt: "0.000001", Completion: "0.000002"},
			expectedType: PricingPerToken,
		},
		{
			name:         "image takes precedence",
			pricing:      Pricing{Image: "0.02", Prompt: "0.001", Completion: "0.002"},
			expectedType: PricingPerImage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := Model{Pricing: tt.pricing}
			info := getPricingInfo(model)
			if info.Type != tt.expectedType {
				t.Errorf("Expected pricing type %v, got %v", tt.expectedType, info.Type)
			}
		})
	}
}

// TestIsModelFree tests the isModelFree function.
func TestIsModelFree(t *testing.T) {
	freeModel := Model{Pricing: Pricing{}}
	if !isModelFree(freeModel) {
		t.Error("Expected model with empty pricing to be free")
	}

	paidModel := Model{Pricing: Pricing{Prompt: "0.000001"}}
	if isModelFree(paidModel) {
		t.Error("Expected model with token pricing to not be free")
	}
}

// TestGetModelCost tests the getModelCost function with real model IDs.
func TestGetModelCost(t *testing.T) {
	// Helper for approximate float comparison
	approxEqual := func(a, b float64) bool {
		const epsilon = 0.0001
		diff := a - b
		if diff < 0 {
			diff = -diff
		}
		return diff < epsilon
	}

	// Gemini 2.5 Flash - hybrid pricing with high output tokens
	// Actual observed: $0.0387
	geminiFlash := Model{
		ID: "google/gemini-2.5-flash-image",
		Pricing: Pricing{
			Image:      "0.001238",
			Prompt:     "0.0000003",
			Completion: "0.0000025",
		},
	}
	cost := getModelCost(geminiFlash)
	// Expected: 0.001238 + (3000 * 0.0000003) + (14500 * 0.0000025) = ~0.0384
	expectedGeminiFlash := 0.001238 + (3000 * 0.0000003) + (14500 * 0.0000025)
	if !approxEqual(cost, expectedGeminiFlash) {
		t.Errorf("Gemini Flash: expected ~%v, got %v", expectedGeminiFlash, cost)
	}

	// OpenAI GPT-5 Image - token-based with low usage
	// Actual observed: $0.022
	gpt5Image := Model{
		ID: "openai/gpt-5-image",
		Pricing: Pricing{
			Image:      "0.00001", // negligible
			Prompt:     "0.00001",
			Completion: "0.00001",
		},
	}
	cost = getModelCost(gpt5Image)
	// Expected: (1500 * 0.00001) + (700 * 0.00001) = 0.015 + 0.007 = 0.022
	expectedGPT5 := (1500 * 0.00001) + (700 * 0.00001)
	if !approxEqual(cost, expectedGPT5) {
		t.Errorf("GPT-5 Image: expected ~%v, got %v", expectedGPT5, cost)
	}

	// Model with pure request cost (no token costs, unknown provider)
	modelWithRequest := Model{
		ID: "unknown/model",
		Pricing: Pricing{
			Request: "0.01",
		},
	}
	cost = getModelCost(modelWithRequest)
	if cost != 0.01 {
		t.Errorf("Expected pure request cost 0.01, got %v", cost)
	}

	// Free model
	freeModel := Model{
		ID: "some/free-model",
		Pricing: Pricing{},
	}
	cost = getModelCost(freeModel)
	if cost != 0 {
		t.Errorf("Expected cost 0 for free model, got %v", cost)
	}
}

// TestFormatPricing tests the formatPricing function.
func TestFormatPricing(t *testing.T) {
	tests := []struct {
		name     string
		pricing  Pricing
		expected string
	}{
		{
			name:     "free",
			pricing:  Pricing{},
			expected: "free",
		},
		{
			name:     "per-image",
			pricing:  Pricing{Image: "0.02"},
			expected: "$0.0200/image",
		},
		{
			name:     "per-request",
			pricing:  Pricing{Request: "0.01"},
			expected: "$0.0100/request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := Model{Pricing: tt.pricing}
			result := formatPricing(model)
			if result != tt.expected {
				t.Errorf("formatPricing() = %q, expected %q", result, tt.expected)
			}
		})
	}

	// Test token-based pricing separately (more complex format)
	tokenModel := Model{Pricing: Pricing{Prompt: "0.000001", Completion: "0.000002"}}
	result := formatPricing(tokenModel)
	if !strings.Contains(result, "/M input") || !strings.Contains(result, "/M output") {
		t.Errorf("Expected token pricing format, got %q", result)
	}
}

// TestStripMarkdownLinks tests the stripMarkdownLinks function.
func TestStripMarkdownLinks(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			input:    "Check out [GPT-5](https://openrouter.ai/openai/gpt-5) for more",
			expected: "Check out GPT-5 for more",
		},
		{
			input:    "[Link1](url1) and [Link2](url2)",
			expected: "Link1 and Link2",
		},
		{
			input:    "No links here",
			expected: "No links here",
		},
		{
			input:    "Incomplete [link without url",
			expected: "Incomplete [link without url",
		},
	}

	for _, tt := range tests {
		result := stripMarkdownLinks(tt.input)
		if result != tt.expected {
			t.Errorf("stripMarkdownLinks(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// TestFilterImageModels tests the filterImageModels function.
func TestFilterImageModels(t *testing.T) {
	models := []Model{
		{
			ID: "text-only",
			Architecture: Architecture{
				OutputModalities: []string{"text"},
			},
		},
		{
			ID: "image-model",
			Architecture: Architecture{
				OutputModalities: []string{"image", "text"},
			},
		},
		{
			ID: "another-text",
			Architecture: Architecture{
				OutputModalities: []string{"text"},
			},
		},
		{
			ID: "image-only",
			Architecture: Architecture{
				OutputModalities: []string{"image"},
			},
		},
	}

	filtered := filterImageModels(models)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 image models, got %d", len(filtered))
	}

	// Check that the right models were selected
	ids := make(map[string]bool)
	for _, m := range filtered {
		ids[m.ID] = true
	}
	if !ids["image-model"] || !ids["image-only"] {
		t.Error("Expected image-model and image-only to be in filtered results")
	}
	if ids["text-only"] || ids["another-text"] {
		t.Error("Text-only models should not be in filtered results")
	}
}

// TestGetAPIKey tests the getAPIKey function.
func TestGetAPIKey(t *testing.T) {
	// Unset the environment variable for testing
	original := os.Getenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENROUTER_API_KEY")
	defer func() {
		if original != "" {
			os.Setenv("OPENROUTER_API_KEY", original)
		}
	}()

	_, err := getAPIKey()
	if err == nil {
		t.Error("Expected error when OPENROUTER_API_KEY is not set")
	}

	// Set a test key
	os.Setenv("OPENROUTER_API_KEY", "test-key")
	key, err := getAPIKey()
	if err != nil {
		t.Errorf("Expected no error with API key set, got: %v", err)
	}
	if key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", key)
	}
}

// TestDefaultModelValue tests that the default ai.model value is "auto".
func TestDefaultModelValue(t *testing.T) {
	resetFlags()
	if aiflags.Model != "auto" {
		t.Errorf("Expected default model to be 'auto', got '%s'", aiflags.Model)
	}
}

// TestWrapText tests the wrapText function.
func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		indent   string
		expected string
	}{
		{
			name:     "short text no wrap",
			text:     "Hello world",
			width:    80,
			indent:   "  ",
			expected: "Hello world",
		},
		{
			name:     "wrap at word boundary",
			text:     "This is a longer text that should wrap to the next line",
			width:    30,
			indent:   "  ",
			expected: "This is a longer text that\n  should wrap to the next line",
		},
		{
			name:     "multiple wraps",
			text:     "One two three four five six seven eight nine ten eleven twelve",
			width:    20,
			indent:   "    ",
			expected: "One two three four\n    five six seven\n    eight nine ten\n    eleven twelve",
		},
		{
			name:     "empty text",
			text:     "",
			width:    80,
			indent:   "  ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.width, tt.indent)
			if result != tt.expected {
				t.Errorf("wrapText(%q, %d, %q) =\n%q\nexpected:\n%q", tt.text, tt.width, tt.indent, result, tt.expected)
			}
		})
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
