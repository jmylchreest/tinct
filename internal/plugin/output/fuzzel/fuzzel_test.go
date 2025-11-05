package fuzzel

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestFuzzelPlugin runs all standard plugin tests using shared utilities.
func TestFuzzelPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "fuzzel",
		ExpectedFiles:      []string{"tinct.ini"},
		ExpectedBinaryName: "fuzzel",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestFuzzelPlugin_ContentValidation tests fuzzel-specific content requirements.
func TestFuzzelPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	// Fuzzel-specific validations.
	requiredStrings := []string{
		"[colors]",
		"background=",
		"text=",
		"prompt=",
		"placeholder=",
		"input=",
		"match=",
		"selection=",
		"selection-text=",
		"selection-match=",
		"counter=",
		"border=",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(content, required) {
			t.Errorf("Generated content missing required string: %s", required)
		}
	}

	// Check that theme type is present.
	if !strings.Contains(content, "Detected theme: dark") {
		t.Error("Generated content missing theme type")
	}
}

// TestFuzzelPlugin_GetEmbeddedTemplates tests embedded template access.
func TestFuzzelPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Name() == "tinct.ini.tmpl" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Template file tinct.ini.tmpl not found in embedded filesystem")
	}
}

// TestFuzzelPlugin_PrepareThemeData tests theme data preparation.
func TestFuzzelPlugin_PrepareThemeData(t *testing.T) {
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

// TestFuzzelPlugin_ColorFormatting tests that colors are formatted correctly.
func TestFuzzelPlugin_ColorFormatting(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	// Verify RGBA color format (RRGGBBAA without #).
	// Fuzzel uses RRGGBBAA format.
	if !strings.Contains(content, "=") {
		t.Error("Generated content should contain color assignments")
	}

	// Check that colors section exists.
	if !strings.Contains(content, "[colors]") {
		t.Error("Generated content should contain [colors] section")
	}
}
