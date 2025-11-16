package ghostty

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestGhosttyPlugin runs all standard plugin tests using shared utilities.
func TestGhosttyPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "ghostty",
		ExpectedFiles:      []string{"tinct.conf"},
		ExpectedBinaryName: "ghostty",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestGhosttyPlugin_ContentValidation tests ghostty-specific content requirements.
func TestGhosttyPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	themeFile := string(files["tinct.conf"])

	// Check for required theme configuration options.
	requiredOptions := []string{
		"background =",
		"foreground =",
		"cursor-color =",
		"cursor-text =",
		"selection-background =",
		"selection-foreground =",
	}

	for _, required := range requiredOptions {
		if !strings.Contains(themeFile, required) {
			t.Errorf("Theme file missing required option: %s", required)
		}
	}

	// Check for all 16 palette entries (0-15).
	for i := 0; i < 16; i++ {
		paletteEntry := "palette = " + string(rune('0'+i))
		if i >= 10 {
			paletteEntry = "palette = 1" + string(rune('0'+(i-10)))
		}
		if !strings.Contains(themeFile, paletteEntry) {
			t.Errorf("Theme file missing palette entry: %d", i)
		}
	}

	// Verify we have exactly 16 palette entries.
	paletteCount := strings.Count(themeFile, "palette =")
	if paletteCount != 16 {
		t.Errorf("Expected 16 palette entries, got %d", paletteCount)
	}

	// Count unique hex color values to verify we have all expected colors.
	// The template uses:
	// - 3 unique semantic colors (background, foreground, accent1)
	// - 16 ANSI palette colors
	// Total: 19 unique colors, though some semantic colors appear multiple times.
	hexColorPattern := "#"
	hexColorCount := strings.Count(themeFile, hexColorPattern)

	// We expect at least 22 hex colors in the output:
	// - background (2 times: background, cursor-text)
	// - foreground (2 times: foreground, selection-foreground)
	// - accent1 (2 times: cursor-color, selection-background)
	// - 16 ANSI palette colors
	// Total: 22 hex color values
	if hexColorCount < 22 {
		t.Errorf("Expected at least 22 hex colors in output, got %d", hexColorCount)
	}
}

// TestGhosttyPlugin_GenerateWithLightTheme tests light theme generation.
func TestGhosttyPlugin_GenerateWithLightTheme(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeLight)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	themeFile := string(files["tinct.conf"])

	// Check that theme file is generated.
	if themeFile == "" {
		t.Error("Theme file should not be empty for light theme")
	}

	// Verify it contains basic configuration.
	if !strings.Contains(themeFile, "background =") {
		t.Error("Theme file should contain background color")
	}
}

// TestGhosttyPlugin_CustomOutputDir tests custom output directory handling.
func TestGhosttyPlugin_CustomOutputDir(t *testing.T) {
	plugin := New()
	plugin.outputDir = "/custom/path"

	dir := plugin.DefaultOutputDir()
	if dir != "/custom/path" {
		t.Errorf("DefaultOutputDir() = %s, want /custom/path", dir)
	}
}

// TestGhosttyPlugin_GetEmbeddedTemplates tests embedded template access.
func TestGhosttyPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Name() == "tinct.tmpl" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Template file tinct.tmpl not found in embedded filesystem")
	}
}

// TestGhosttyPlugin_PrepareThemeData tests theme data preparation.
func TestGhosttyPlugin_PrepareThemeData(t *testing.T) {
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

// TestGhosttyPlugin_ReloadFlag tests reload flag configuration.
func TestGhosttyPlugin_ReloadFlag(t *testing.T) {
	plugin := New()

	// Default should be true.
	if !plugin.reload {
		t.Error("Default reload flag should be true")
	}

	// Test disabling reload.
	plugin.reload = false
	if plugin.reload {
		t.Error("Reload flag should be false after setting")
	}
}

// TestGhosttyPlugin_HexColors tests that colors are in hex format.
func TestGhosttyPlugin_HexColors(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	themeFile := string(files["tinct.conf"])

	// Check that colors are in hex format (#RRGGBB or #RRGGBBAA).
	lines := strings.Split(themeFile, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check lines with color assignments.
		if strings.Contains(line, " = ") {
			parts := strings.SplitN(line, " = ", 2)
			if len(parts) == 2 {
				value := strings.TrimSpace(parts[1])
				// For palette entries like "palette = 0=#000000", extract the color after the =.
				if strings.HasPrefix(parts[0], "palette") && strings.Contains(value, "=") {
					colorParts := strings.SplitN(value, "=", 2)
					if len(colorParts) == 2 {
						value = colorParts[1]
					}
				}
				// Verify hex format.
				if value != "" && !strings.HasPrefix(value, "#") {
					t.Errorf("Color value should start with #, got line: %s", line)
				}
			}
		}
	}
}
