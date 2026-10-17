package swaync

import (
	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
	"strings"
	"testing"
)

func TestSwayNC(t *testing.T) {
	plugintesting.RunAllTests(t, New(), plugintesting.TestConfig{ExpectedName: "swaync", ExpectedFiles: []string{"tinct.css"}, ExpectedBinaryName: "swaync-client"})
	for _, mode := range []colour.ThemeType{colour.ThemeDark, colour.ThemeLight} {
		data := colour.NewThemeData(plugintesting.CreateTestPalette(mode), "", "", "")
		files, err := New().Generate(data)
		if err != nil {
			t.Fatal(err)
		}
		css := string(files["tinct.css"])
		for _, expected := range []string{":root", "--noti-bg:", "--text-color:", ".notification.critical"} {
			if !strings.Contains(css, expected) {
				t.Errorf("missing %s", expected)
			}
		}
		if strings.Contains(css, "<no value>") {
			t.Fatal("unresolved template value")
		}
		if len(files) != 1 {
			t.Fatal("must only emit additive palette")
		}
	}
	spec := New().Hooks()
	if strings.Join(spec.Reload.Args, " ") != "swaync-client --reload-css --skip-wait" {
		t.Fatal("reload must not wait for daemon startup")
	}
}
