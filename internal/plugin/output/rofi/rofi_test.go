package rofi

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

func TestRofiPlugin(t *testing.T) {
	plugin := New()
	plugintesting.RunAllTests(t, plugin, plugintesting.TestConfig{
		ExpectedName:       "rofi",
		ExpectedFiles:      []string{"tinct-colours.rasi", "tinct.rasi"},
		ExpectedBinaryName: "rofi",
	})
}

func TestRofiPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Palette file must declare named colours but include no widget rules.
	coloursContent := string(files["tinct-colours.rasi"])
	for _, required := range []string{
		"* {",
		"background:",
		"foreground:",
		"accent1:",
		"danger:",
	} {
		if !strings.Contains(coloursContent, required) {
			t.Errorf("palette missing required declaration: %s", required)
		}
	}
	for _, forbidden := range []string{"window {", "mainbox {", "element selected {"} {
		if strings.Contains(coloursContent, forbidden) {
			t.Errorf("palette should not contain widget rule: %s", forbidden)
		}
	}

	// Main theme must import the palette and use @<name> references.
	mainContent := string(files["tinct.rasi"])
	for _, required := range []string{
		`@import "tinct-colours.rasi"`,
		"window {",
		"mainbox {",
		"listview {",
		"element selected {",
		"@accent1",
		"@surface",
	} {
		if !strings.Contains(mainContent, required) {
			t.Errorf("main theme missing required string: %s", required)
		}
	}
}
