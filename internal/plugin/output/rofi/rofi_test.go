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
		ExpectedFiles:      []string{"tinct.rasi"},
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

	content := string(files["tinct.rasi"])
	for _, required := range []string{
		"window {", "mainbox {", "listview {", "element selected {",
		"@accent1", "@background",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("missing required string: %s", required)
		}
	}
}
