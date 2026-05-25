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
		ExpectedFiles:      []string{"tinct-palette.conf", "tinct.conf"},
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

	// Palette file must declare every key @tinct-* user option.
	paletteContent := string(files["tinct-palette.conf"])
	for _, required := range []string{
		`set -gq @tinct-background `,
		`set -gq @tinct-foreground `,
		`set -gq @tinct-accent1 `,
		`set -gq @tinct-danger `,
	} {
		if !strings.Contains(paletteContent, required) {
			t.Errorf("palette missing required declaration: %s", required)
		}
	}

	// Main config must source the palette and apply opinionated styles
	// using #{@tinct-*} references (no inline hex).
	mainContent := string(files["tinct.conf"])
	for _, required := range []string{
		"source-file ~/.config/tmux/tinct-palette.conf",
		"#{@tinct-surface}",
		"#{@tinct-accent1}",
		"set -g status-style",
		"set -g pane-active-border-style",
	} {
		if !strings.Contains(mainContent, required) {
			t.Errorf("main config missing required string: %s", required)
		}
	}
}
