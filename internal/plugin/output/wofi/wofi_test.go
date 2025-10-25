package wofi

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestWofiPlugin runs all standard plugin tests using shared utilities.
func TestWofiPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "wofi",
		ExpectedFiles:      []string{"tinct-colors", "style.css"},
		ExpectedBinaryName: "wofi",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestWofiPlugin_ContentValidation tests wofi-specific content requirements.
func TestWofiPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(palette)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check colors file
	colors := string(files["tinct-colors"])
	if !strings.Contains(colors, "#") {
		t.Error("Colors file should contain hex colors")
	}

	// Check style file
	style := string(files["style.css"])
	if !strings.Contains(style, "window") {
		t.Error("Style file should contain window selector")
	}
	if !strings.Contains(style, "--wofi-color") {
		t.Error("Style file should contain --wofi-color variables")
	}
}

// TestWofiPlugin_GenerateColors tests color file generation.
func TestWofiPlugin_GenerateColors(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	content, err := plugin.generateColors(palette)
	if err != nil {
		t.Fatalf("generateColors() error = %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "#") {
		t.Error("Colors should contain hex colors")
	}
}

// TestWofiPlugin_GenerateStyle tests style file generation.
func TestWofiPlugin_GenerateStyle(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	content, err := plugin.generateStyle(palette)
	if err != nil {
		t.Fatalf("generateStyle() error = %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "#window") {
		t.Error("Style should contain #window selector")
	}
	if !strings.Contains(contentStr, "--wofi-color") {
		t.Error("Style should contain --wofi-color variables")
	}
}

// TestWofiPlugin_PrepareThemeData tests theme data preparation.
func TestWofiPlugin_PrepareThemeData(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	data := plugin.prepareThemeData(palette)

	if data.SourceTheme == "" {
		t.Error("SourceTheme should not be empty")
	}
	if data.Background == "" {
		t.Error("Background should not be empty")
	}
	if data.SourceTheme != "dark" {
		t.Errorf("SourceTheme = %s, want dark", data.SourceTheme)
	}
}

// TestWofiPlugin_GenerateWithLightTheme tests light theme generation.
func TestWofiPlugin_GenerateWithLightTheme(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeLight)
	plugin := New()

	files, err := plugin.Generate(palette)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) == 0 {
		t.Fatal("Generate() returned no files")
	}

	data := plugin.prepareThemeData(palette)
	if data.SourceTheme != "light" {
		t.Errorf("SourceTheme = %s, want light", data.SourceTheme)
	}
}

// TestWofiPlugin_GetEmbeddedTemplates tests embedded template access.
func TestWofiPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	foundColors := false
	foundStyle := false
	for _, entry := range entries {
		if entry.Name() == "tinct-colors" {
			foundColors = true
		}
		if entry.Name() == "style.css.tmpl" {
			foundStyle = true
		}
	}

	if !foundColors {
		t.Error("Template file tinct-colors not found")
	}
	if !foundStyle {
		t.Error("Template file style.css.tmpl not found")
	}
}
