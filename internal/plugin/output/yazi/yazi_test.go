package yazi

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

func TestYaziPlugin(t *testing.T) {
	plugin := New()
	plugintesting.RunAllTests(t, plugin, plugintesting.TestConfig{
		ExpectedName:         "yazi",
		ExpectedFiles:        []string{"flavor.toml"},
		ExpectedBinaryName:   "yazi",
		ExpectedDirSubstring: "yazi/flavors",
	})
}

func TestYaziPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["flavor.toml"])
	for _, required := range []string{
		"[mgr]", "[status]", "[filetype]", "border_style", "mode_normal",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("missing required string: %s", required)
		}
	}
}
