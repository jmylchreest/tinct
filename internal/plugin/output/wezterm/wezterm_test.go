package wezterm

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

func TestWeztermPlugin(t *testing.T) {
	plugin := New()
	plugintesting.RunAllTests(t, plugin, plugintesting.TestConfig{
		ExpectedName:       "wezterm",
		ExpectedFiles:      []string{"tinct.toml"},
		ExpectedBinaryName: "wezterm",
	})
}

func TestWeztermPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.toml"])
	for _, required := range []string{
		"[colors]", "foreground", "background", "ansi = [",
		"[colors.tab_bar.active_tab]",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("missing required string: %s", required)
		}
	}
}
