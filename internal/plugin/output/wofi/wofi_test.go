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

	styleCSS := string(files["style.css"])
	colorsFile := string(files["tinct-colors"])

	// Check for required CSS content in style
	requiredStyleStrings := []string{
		"window",
		"#input",
		"#entry",
	}

	for _, required := range requiredStyleStrings {
		if !strings.Contains(styleCSS, required) {
			t.Errorf("Style CSS missing required string: %s", required)
		}
	}

	// Check colors file contains color definitions
	if len(colorsFile) == 0 {
		t.Error("Colors file should not be empty")
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

	colorsFile := string(files["tinct-colors"])

	// Check that colors file is generated
	if len(colorsFile) == 0 {
		t.Error("Colors file should not be empty for light theme")
	}
}

// TestWofiPlugin_CustomOutputDir tests custom output directory handling.
func TestWofiPlugin_CustomOutputDir(t *testing.T) {
	plugin := New()
	plugin.outputDir = "/custom/path"

	dir := plugin.DefaultOutputDir()
	if dir != "/custom/path" {
		t.Errorf("DefaultOutputDir() = %s, want /custom/path", dir)
	}
}

// TestWofiPlugin_GetEmbeddedTemplates tests embedded template access.
func TestWofiPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Name() == "style.css.tmpl" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Template file style.css.tmpl not found in embedded filesystem")
	}
}

// TestWofiPlugin_PrepareThemeData tests theme data preparation.
func TestWofiPlugin_PrepareThemeData(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	data := plugin.prepareThemeData(palette)

	// Check that PaletteHelper is created properly
	if data == nil {
		t.Fatal("prepareThemeData should return non-nil PaletteHelper")
	}

	// Check that required roles exist
	requiredRoles := []colour.ColourRole{
		colour.RoleBackground,
		colour.RoleForeground,
	}

	for _, role := range requiredRoles {
		if !data.Has(role) {
			t.Errorf("PaletteHelper missing required role: %s", role)
		}
	}

	// Check that theme type matches
	if data.ThemeTypeString() != "dark" {
		t.Errorf("ThemeTypeString() = %s, want dark", data.ThemeTypeString())
	}
}
