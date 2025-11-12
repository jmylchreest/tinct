package zellij

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestZellijPlugin runs all standard plugin tests using shared utilities.
func TestZellijPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "zellij",
		ExpectedFiles:      []string{"tinct.kdl"},
		ExpectedBinaryName: "zellij",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestZellijPlugin_ContentValidation tests zellij-specific content requirements.
func TestZellijPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.kdl"])

	// Check for required KDL theme structure.
	requiredStrings := []string{
		"themes {",
		"tinct {",
		"text_unselected {",
		"base",
		"background",
		"emphasis_",
		"}",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(content, required) {
			t.Errorf("Generated content missing required string: %s", required)
		}
	}
}
