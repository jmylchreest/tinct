package tmux

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

func TestTmuxPlugin(t *testing.T) {
	plugin := New()
	plugintesting.RunAllTests(t, plugin, plugintesting.TestConfig{
		ExpectedName:       "tmux",
		ExpectedFiles:      []string{"tinct.conf"},
		ExpectedBinaryName: "tmux",
	})
}

func TestTmuxPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.conf"])
	for _, required := range []string{
		"@thm_bg", "@thm_fg", "@thm_accent1",
		"set -g status-style", "set -g pane-active-border-style",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("missing required string: %s", required)
		}
	}
}
