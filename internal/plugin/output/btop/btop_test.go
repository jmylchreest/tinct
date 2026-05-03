package btop

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
)

func TestBtopPlugin(t *testing.T) {
	plugin := New()
	plugintesting.RunAllTests(t, plugin, plugintesting.TestConfig{
		ExpectedName:       "btop",
		ExpectedFiles:      []string{"tinct.theme"},
		ExpectedBinaryName: "btop",
	})
}

func TestBtopPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(colour.NewThemeData(palette, "", "", ""))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.theme"])
	for _, required := range []string{
		"theme[main_bg]=", "theme[main_fg]=", "theme[title]=",
		"theme[cpu_start]=", "theme[mem_box]=",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("missing required key: %s", required)
		}
	}
}
