package walker

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

// TestWalkerPlugin runs all standard plugin tests using shared utilities.
func TestWalkerPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "walker",
		ExpectedFiles:      []string{"tinct-colours.css", "tinct/style.css"},
		ExpectedBinaryName: "walker",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestWalkerPlugin_ContentValidation tests walker-specific content requirements.
func TestWalkerPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	styleCSS := string(files["tinct/style.css"])
	colorsFile := string(files["tinct-colours.css"])

	// Check for required CSS content in style.
	requiredStyleStrings := []string{
		".box-wrapper",
		".input",
		".item-box",
		"@import '../tinct-colours.css'",
	}

	for _, required := range requiredStyleStrings {
		if !strings.Contains(styleCSS, required) {
			t.Errorf("Style CSS missing required string: %s", required)
		}
	}

	// Check colors file contains all 5 GTK color definitions.
	requiredColorStrings := []string{
		"@define-color window_bg_color",
		"@define-color theme_fg_color",
		"@define-color accent_bg_color",
		"@define-color error_bg_color",
		"@define-color error_fg_color",
	}

	for _, required := range requiredColorStrings {
		if !strings.Contains(colorsFile, required) {
			t.Errorf("Colors file missing required string: %s", required)
		}
	}

	// Verify we have exactly 5 color definitions.
	colorCount := strings.Count(colorsFile, "@define-color")
	if colorCount != 5 {
		t.Errorf("Expected 5 color definitions, got %d", colorCount)
	}
}

// TestWalkerPlugin_GenerateWithLightTheme tests light theme generation.
func TestWalkerPlugin_GenerateWithLightTheme(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeLight)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	colorsFile := string(files["tinct-colours.css"])

	// Check that colors file is generated.
	if colorsFile == "" {
		t.Error("Colors file should not be empty for light theme")
	}

	// Verify it contains @define-color directives.
	if !strings.Contains(colorsFile, "@define-color") {
		t.Error("Colors file should contain @define-color directives")
	}
}

// TestWalkerPlugin_CustomOutputDir tests custom output directory handling.
func TestWalkerPlugin_CustomOutputDir(t *testing.T) {
	plugin := New()
	plugin.outputDir = "/custom/path"

	dir := plugin.DefaultOutputDir()
	if dir != "/custom/path" {
		t.Errorf("DefaultOutputDir() = %s, want /custom/path", dir)
	}
}

// TestWalkerPlugin_GetEmbeddedTemplates tests embedded template access.
func TestWalkerPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	expectedTemplates := []string{"tinct-colours.css.tmpl", "tinct.css.tmpl"}
	foundTemplates := make(map[string]bool)

	for _, entry := range entries {
		foundTemplates[entry.Name()] = true
	}

	for _, expected := range expectedTemplates {
		if !foundTemplates[expected] {
			t.Errorf("Template file %s not found in embedded filesystem", expected)
		}
	}
}

// TestWalkerPlugin_PrepareThemeData tests theme data preparation.
func TestWalkerPlugin_PrepareThemeData(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	data := colour.NewThemeData(palette, "", "")

	// Check that PaletteHelper is created properly.
	if data == nil {
		t.Fatal("NewThemeData should return non-nil ThemeData")
	}

	// Check that required roles exist.
	requiredRoles := []colour.Role{
		colour.RoleBackground,
		colour.RoleForeground,
		colour.RoleAccent1,
		colour.RoleDanger,
		colour.RoleOnDanger,
	}

	for _, role := range requiredRoles {
		if !data.Has(role) {
			t.Errorf("ThemeData missing required role: %s", role)
		}
	}

	// Check that theme type matches.
	if data.ThemeTypeString() != "dark" {
		t.Errorf("ThemeTypeString() = %s, want dark", data.ThemeTypeString())
	}
}

// TestWalkerPlugin_FilePathStructure tests that files are generated with correct paths.
func TestWalkerPlugin_FilePathStructure(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that tinct-colours.css is at root level.
	if _, ok := files["tinct-colours.css"]; !ok {
		t.Error("Expected tinct-colours.css at root level")
	}

	// Check that style.css is in tinct/ subdirectory.
	if _, ok := files["tinct/style.css"]; !ok {
		t.Error("Expected tinct/style.css in subdirectory")
	}

	// Ensure we don't have the old path.
	if _, ok := files["tinct.css"]; ok {
		t.Error("Should not have tinct.css at root level (should be tinct/style.css)")
	}
}
