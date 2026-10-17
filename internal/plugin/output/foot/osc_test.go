package foot

import (
	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
	"strings"
	"testing"
)

func TestLivePalette(t *testing.T) {
	p := New()
	p.tty = "/dev/tty"
	files, err := p.Generate(colour.NewThemeData(plugintesting.CreateTestPalette(colour.ThemeDark), "", "", ""))
	if err != nil {
		t.Fatal(err)
	}
	if p.Hooks().ReloadFn == nil || p.Hooks().Reload != nil {
		t.Fatal("expected OSC callback, never signals")
	}
	if strings.Count(string(p.osc), "\x1b]") != 21 {
		t.Fatalf("wrong OSC count: %q", p.osc)
	}
	if !strings.Contains(string(p.osc), "\x1b]4;15;#") || !strings.Contains(string(p.osc), "\x1b]19;#") {
		t.Fatal("missing palette/selection sequences")
	}
	for _, bad := range []string{"ffffff\x1b]52;evil", "#ffffff", "", "ffffffff"} {
		data := strings.Replace(string(files["tinct-colors.ini"]), "foreground=", "foreground="+bad+"\nignored=", 1)
		if _, err := paletteOSC([]byte(data)); err == nil {
			t.Errorf("accepted invalid colour %q", bad)
		}
	}
	if _, err := p.Generate(nil); err == nil || len(p.osc) != 0 {
		t.Fatal("failed generation must clear stale payload")
	}
}

func TestTTYValidation(t *testing.T) {
	for _, target := range []string{"", "/dev/tty", "/dev/pts/123"} {
		if err := (&Plugin{tty: target}).Validate(); err != nil {
			t.Errorf("%q: %v", target, err)
		}
	}
	for _, target := range []string{"/tmp/tty", "/dev/pts/../null", "/dev/pts/1\n"} {
		if err := (&Plugin{tty: target}).Validate(); err == nil {
			t.Errorf("accepted %q", target)
		}
	}
}
