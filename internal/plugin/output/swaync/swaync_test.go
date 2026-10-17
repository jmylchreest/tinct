package swaync

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
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
		palette := data.GetPaletteHelper()
		for _, expected := range []string{
			"--noti-bg: " + palette.Get("surface").RGBDecimal() + ";",
			"--text-color: " + palette.Get("foreground").RGBA() + ";",
			// The upstream border shorthand has three class selectors; this
			// override needs four to win even when imported after that rule.
			".notification-row .notification-background .notification.critical {\n    border-color: " + palette.Get("danger").RGBA() + ";",
		} {
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
	if len(spec.RequiredBinaries) != 1 || spec.RequiredBinaries[0] != "swaync-client" || !spec.AutoCreateDir {
		t.Fatal("hooks must check swaync-client and create the output directory")
	}
	if spec.Reload == nil {
		t.Fatal("reload hook is required")
	}
	if strings.Join(spec.Reload.Args, " ") != "swaync-client --reload-css --skip-wait" {
		t.Fatal("reload must not wait for daemon startup")
	}
}
