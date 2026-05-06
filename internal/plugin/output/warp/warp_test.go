package warp

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

// TestWarpPlugin runs all standard plugin tests using shared utilities.
func TestWarpPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:  "warp",
		ExpectedFiles: []string{"tinct.yaml"},
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestWarpPlugin_ContentValidation tests warp-specific content requirements.
func TestWarpPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	themeFile := string(files["tinct.yaml"])

	// Required top-level keys per Warp's theme schema.
	required := []string{
		"name: Tinct",
		"accent:",
		"cursor:",
		"background:",
		"foreground:",
		"details:",
		"terminal_colors:",
		"normal:",
		"bright:",
	}
	for _, r := range required {
		if !strings.Contains(themeFile, r) {
			t.Errorf("theme missing required field: %q", r)
		}
	}

	// All 16 ANSI slots must appear under terminal_colors.
	ansiNames := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
	for _, name := range ansiNames {
		// each name appears once under normal: and once under bright:
		if c := strings.Count(themeFile, name+":"); c < 2 {
			t.Errorf("ANSI colour %q should appear under both normal and bright (found %d)", name, c)
		}
	}
}

// TestWarpPlugin_DetailsLighter tests dark themes get details: lighter.
func TestWarpPlugin_DetailsLighter(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, _ := plugin.Generate(themeData)
	if !strings.Contains(string(files["tinct.yaml"]), "details: lighter") {
		t.Errorf("dark theme should set details: lighter; got:\n%s", files["tinct.yaml"])
	}
}

// TestWarpPlugin_DetailsDarker tests light themes get details: darker.
func TestWarpPlugin_DetailsDarker(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeLight)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, _ := plugin.Generate(themeData)
	if !strings.Contains(string(files["tinct.yaml"]), "details: darker") {
		t.Errorf("light theme should set details: darker; got:\n%s", files["tinct.yaml"])
	}
}

// TestWarpPlugin_HexColours tests every colour value uses #RRGGBB format
// quoted as a YAML string (single quotes) so YAML doesn't interpret the
// hash as a comment.
func TestWarpPlugin_HexColours(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	themeFile := string(files["tinct.yaml"])

	// Lines containing colour values should be quoted; otherwise YAML
	// would treat the leading `#` as a comment marker and lose the value.
	for _, line := range strings.Split(themeFile, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, "#") {
			continue
		}
		// Comment-only lines start with '#'; skip those.
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.Contains(trimmed, "'#") {
			t.Errorf("colour value not single-quoted (YAML would parse # as comment): %s", line)
		}
	}
}

// TestWarpPlugin_CustomOutputDir tests custom output directory handling.
func TestWarpPlugin_CustomOutputDir(t *testing.T) {
	plugin := New()
	plugin.outputDir = "/custom/path"

	if dir := plugin.DefaultOutputDir(); dir != "/custom/path" {
		t.Errorf("DefaultOutputDir() = %s, want /custom/path", dir)
	}
}
