package helix

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

func TestHelixPlugin(t *testing.T) {
	plugin := New()
	plugintesting.RunAllTests(t, plugin, plugintesting.TestConfig{
		ExpectedName:       "helix",
		ExpectedFiles:      []string{"tinct-palette.toml", "tinct.toml"},
		ExpectedBinaryName: "hx",
	})
}

func TestHelixPlugin_PaletteContent(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	paletteContent := string(files["tinct-palette.toml"])
	// The palette file MUST be a [palette] block. Helix only honours
	// named colours that appear under this header.
	for _, required := range []string{
		"[palette]",
		"background",
		"foreground",
		"accent1",
		"danger",
		"warning",
		"success",
	} {
		if !strings.Contains(paletteContent, required) {
			t.Errorf("palette missing required string: %s", required)
		}
	}

	// `inherits =` is the load-bearing line — without it the main file
	// has no colour names to reference and Helix renders everything
	// against the built-in default palette.
	themeContent := string(files["tinct.toml"])
	if !strings.Contains(themeContent, `inherits = "tinct-palette"`) {
		t.Errorf("theme file missing inherits directive")
	}
	for _, required := range []string{
		`"ui.background"`,
		`"ui.text"`,
		`"ui.cursor"`,
		`"ui.statusline"`,
		`"diagnostic.error"`,
		`"keyword"`,
		`"function"`,
		`"string"`,
		`"comment"`,
	} {
		if !strings.Contains(themeContent, required) {
			t.Errorf("theme missing required key: %s", required)
		}
	}
}

func TestHelixPlugin_LightTheme(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeLight)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}
}
