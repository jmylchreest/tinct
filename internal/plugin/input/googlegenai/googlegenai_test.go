package googlegenai

import (
	"context"
	"io"
	"os"
	"path/filepath"
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
		t.Fatal("DryRun should return a palette")
	}
	if len(palette.Colors) != 0 {
		t.Errorf("DryRun should return empty palette, got %d colors", len(palette.Colors))
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
		{"gemini-3-pro-image-preview", true},
		{"gemini-3.1-flash-image", true},
		{"imagen-4.0-generate-001", false},
		{"imagen-4.0-ultra-generate-001", false},
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

func TestGetExtensionForMIME(t *testing.T) {
	tests := map[string]string{
		"image/jpeg":    ".jpg",
		"image/jpg":     ".jpg",
		"image/png":     ".png",
		"image/webp":    ".webp",
		"image/gif":     ".gif",
		"application/x": ".png", // unknown -> png (requested output format)
	}
	for mime, want := range tests {
		if got := getExtensionForMIME(mime); got != want {
			t.Errorf("getExtensionForMIME(%q) = %q, want %q", mime, got, want)
		}
	}
}

func TestImageExtension(t *testing.T) {
	jpegBytes := []byte("\xff\xd8\xff\xe0\x00\x10JFIF")
	pngBytes := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")

	// Reported MIME type wins over the bytes.
	if mime, ext := imageExtension("image/jpeg", pngBytes); mime != "image/jpeg" || ext != ".jpg" {
		t.Errorf("imageExtension(reported jpeg) = (%q, %q), want (image/jpeg, .jpg)", mime, ext)
	}

	// Missing MIME type falls back to sniffing the bytes.
	if mime, ext := imageExtension("", jpegBytes); mime != "image/jpeg" || ext != ".jpg" {
		t.Errorf("imageExtension(sniff jpeg) = (%q, %q), want (image/jpeg, .jpg)", mime, ext)
	}
	if mime, ext := imageExtension("", pngBytes); mime != "image/png" || ext != ".png" {
		t.Errorf("imageExtension(sniff png) = (%q, %q), want (image/png, .png)", mime, ext)
	}
}

func TestFindCachedImage(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "genai-abc123")

	if got := findCachedImage(base); got != "" {
		t.Errorf("expected empty result for missing image, got %q", got)
	}

	want := base + ".jpg"
	if err := os.WriteFile(want, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := findCachedImage(base); got != want {
		t.Errorf("findCachedImage = %q, want %q", got, want)
	}
}

func TestGetImageBasePathStripsExtension(t *testing.T) {
	resetFlags()
	commonflags.CacheDir = t.TempDir()
	commonflags.CacheFilename = "mywallpaper.png"
	defer resetFlags()

	plugin := New()
	base, err := plugin.getImageBasePath("gemini-3.1-flash-image")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(base) != "" {
		t.Errorf("expected base path without extension, got %q", base)
	}
	if filepath.Base(base) != "mywallpaper" {
		t.Errorf("expected base name 'mywallpaper', got %q", filepath.Base(base))
	}
}

// TestDeprecatedModelsMatchCatalogue guards against the two lists drifting:
// every model carrying a Notes entry in the ListModels catalogue must also
// appear in deprecatedModels, so a user who pins that model on the command
// line gets the same warning that --ai.list-models shows.
func TestDeprecatedModelsMatchCatalogue(t *testing.T) {
	// The catalogue is built inside ListModels, so assert against the model
	// IDs we know carry deprecation notes.
	wantDeprecated := []string{
		"imagen-4.0-ultra-generate-001",
		"imagen-4.0-generate-001",
		"imagen-4.0-fast-generate-001",
		"imagen-3.0-generate-002",
	}

	for _, id := range wantDeprecated {
		if _, ok := deprecatedModels[id]; !ok {
			t.Errorf("model %q is deprecated in the catalogue but missing from deprecatedModels", id)
		}
	}

	if len(deprecatedModels) != len(wantDeprecated) {
		t.Errorf("deprecatedModels has %d entries, want %d", len(deprecatedModels), len(wantDeprecated))
	}

	// The default must never itself be deprecated.
	if _, ok := deprecatedModels[defaultModel]; ok {
		t.Errorf("defaultModel %q is on the deprecation list", defaultModel)
	}

	// Every replacement suggestion must name a live Gemini model.
	for id, advice := range deprecatedModels {
		if !strings.Contains(advice, "gemini-") {
			t.Errorf("advice for %q does not name a replacement model: %q", id, advice)
		}
	}
}

// TestWarnIfDeprecatedModelIsSilentForLiveModels ensures the warning does not
// fire for the models we actually recommend.
func TestWarnIfDeprecatedModelIsSilentForLiveModels(t *testing.T) {
	for _, id := range []string{defaultModel, "gemini-3-pro-image-preview", "gemini-2.5-flash-image"} {
		if _, ok := deprecatedModels[id]; ok {
			t.Errorf("live model %q must not be on the deprecation list", id)
		}
	}
}

// captureStderr runs fn with os.Stderr redirected and returns what it wrote.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return string(out)
}

// TestWarnIfDeprecatedModelOutput checks the warning users actually see: that
// it names the model, explains the status, and points at a live replacement.
func TestWarnIfDeprecatedModelOutput(t *testing.T) {
	tests := []struct {
		name        string
		model       string
		wantWarning bool
		wantSubstr  []string
	}{
		{
			name:        "retired 3.0 reports its real past shutdown date",
			model:       "imagen-3.0-generate-002",
			wantWarning: true,
			// The date this PR corrects: it was previously advertised as 2026-08-17.
			wantSubstr: []string{"imagen-3.0-generate-002", "retired", "2025-11-10", "gemini-3.1-flash-image"},
		},
		{
			name:        "deprecated 4.0 reports earliest-shutdown wording",
			model:       "imagen-4.0-generate-001",
			wantWarning: true,
			wantSubstr:  []string{"imagen-4.0-generate-001", "earliest shutdown", "2026-08-17", "gemini-3.1-flash-image"},
		},
		{
			name:        "ultra names the nearest quality tier as well",
			model:       "imagen-4.0-ultra-generate-001",
			wantWarning: true,
			wantSubstr:  []string{"gemini-3.1-flash-image", "gemini-3-pro-image-preview"},
		},
		{
			name:        "default model is silent",
			model:       defaultModel,
			wantWarning: false,
		},
		{
			name:        "unknown model is silent",
			model:       "some-future-model",
			wantWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := captureStderr(t, func() { warnIfDeprecatedModel(tt.model) })

			if !tt.wantWarning {
				if out != "" {
					t.Errorf("expected no warning for %q, got %q", tt.model, out)
				}
				return
			}

			if out == "" {
				t.Fatalf("expected a warning for %q, got none", tt.model)
			}
			for _, want := range tt.wantSubstr {
				if !strings.Contains(out, want) {
					t.Errorf("warning for %q missing %q\ngot: %s", tt.model, want, out)
				}
			}
		})
	}
}

// TestWarnIfDeprecatedModelNeverMentionsARetiredReplacement guards the failure
// mode where advice sends a user from one dead model to another: Google's own
// table lists imagen-4.0-generate-001 as the replacement for imagen-3.0, and
// that model is itself on the way out.
func TestWarnIfDeprecatedModelNeverMentionsARetiredReplacement(t *testing.T) {
	for model := range deprecatedModels {
		out := captureStderr(t, func() { warnIfDeprecatedModel(model) })
		for dead := range deprecatedModels {
			if dead == model {
				continue
			}
			if strings.Contains(out, dead) {
				t.Errorf("advice for %q points at deprecated model %q: %s", model, dead, out)
			}
		}
	}
}

// captureStdout runs fn with os.Stdout redirected and returns what it wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	return string(out)
}

// TestListModelsOutput covers the catalogue rendering, which is the only place
// a user can discover model status. It runs offline and needs no API key.
func TestListModelsOutput(t *testing.T) {
	out := captureStdout(t, ListModels)

	if out == "" {
		t.Fatal("ListModels produced no output")
	}

	// Every model the plugin accepts must be listed, or users cannot discover it.
	for _, id := range []string{
		"gemini-3-pro-image-preview",
		"gemini-3.1-flash-image",
		"gemini-2.5-flash-image",
		"imagen-4.0-ultra-generate-001",
		"imagen-4.0-generate-001",
		"imagen-4.0-fast-generate-001",
		"imagen-3.0-generate-002",
	} {
		if !strings.Contains(out, id) {
			t.Errorf("model %q missing from ListModels output", id)
		}
	}

	// The default must be identified as such.
	if !strings.Contains(out, defaultModel) {
		t.Errorf("default model %q not named in output", defaultModel)
	}
	if !strings.Contains(out, "Default Model:") {
		t.Error("output does not label the default model")
	}

	// Every deprecated model must carry a visible note. Without this the
	// catalogue can silently list a dead model as though it were live.
	for id := range deprecatedModels {
		idx := strings.Index(out, id)
		if idx < 0 {
			continue // reported above
		}
		rest := out[idx:]
		if end := strings.Index(rest, "ID: "); end > 0 {
			rest = rest[:end]
		}
		if !strings.Contains(rest, "Notes:") {
			t.Errorf("deprecated model %q has no Notes line in ListModels output", id)
		}
	}
}

// TestListModelsReportsCorrectedShutdownDates pins the specific dates this
// change fixes, so a future edit cannot quietly reintroduce the wrong one.
func TestListModelsReportsCorrectedShutdownDates(t *testing.T) {
	out := captureStdout(t, ListModels)

	// imagen-3.0 was previously advertised with the 4.0 family's date.
	idx := strings.Index(out, "imagen-3.0-generate-002")
	if idx < 0 {
		t.Fatal("imagen-3.0-generate-002 missing from output")
	}
	entry := out[idx:]
	if end := strings.Index(entry, "ID: "); end > 0 {
		entry = entry[:end]
	}
	if strings.Contains(entry, "2026-08-17") {
		t.Errorf("imagen-3.0-generate-002 still shows the 4.0 shutdown date:\n%s", entry)
	}
	if !strings.Contains(entry, "2025-11-10") {
		t.Errorf("imagen-3.0-generate-002 missing its real shutdown date:\n%s", entry)
	}

	// Google's dates are earliest-possible; the wording must not overstate them.
	if strings.Contains(out, "shuts down 2026-08-17") {
		t.Error("output states a firm shutdown date; Google publishes earliest-possible dates")
	}
}
