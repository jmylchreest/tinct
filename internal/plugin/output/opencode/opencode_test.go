package opencode

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

func TestOpencodePlugin(t *testing.T) {
	plugin := New()
	plugintesting.RunAllTests(t, plugin, plugintesting.TestConfig{
		ExpectedName:       "opencode",
		ExpectedFiles:      []string{"tinct.json"},
		ExpectedBinaryName: "opencode",
	})
}

// TestOpencodePlugin_ValidJSON ensures the rendered template parses as JSON
// (catches stray commas / unescaped quotes from the template).
func TestOpencodePlugin_ValidJSON(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(files["tinct.json"], &parsed); err != nil {
		t.Fatalf("rendered theme is not valid JSON: %v", err)
	}

	theme, ok := parsed["theme"].(map[string]any)
	if !ok {
		t.Fatalf("theme key missing or wrong type")
	}

	for _, required := range []string{
		"primary", "secondary", "background", "border",
		"diffAdded", "markdownHeading", "syntaxKeyword",
	} {
		if _, ok := theme[required]; !ok {
			t.Errorf("theme missing required token: %s", required)
		}
	}
}

func TestOpencodePlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.json"])
	for _, required := range []string{
		`"$schema"`, `"theme"`, `"dark"`, `"light"`,
	} {
		if !strings.Contains(content, required) {
			t.Errorf("missing required string: %s", required)
		}
	}
}
