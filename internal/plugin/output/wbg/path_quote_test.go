package wbg

import (
	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
	"github.com/jmylchreest/tinct/pkg/util/subprocess"
	"strings"
	"testing"
)

func TestWallpaperPathShellQuoting(t *testing.T) {
	path := "/tmp/it's $(printf surprise);wallpaper.png"
	data := colour.NewThemeData(plugintesting.CreateTestPalette(colour.ThemeDark), path, "", "")
	data.WallpaperPath = path
	files, err := New().Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	script := string(files["tinct-wbg.sh"])
	if !strings.Contains(script, "TINCT_WBG_WALLPAPER="+subprocess.ShellQuote(path)) {
		t.Fatalf("wallpaper path is not shell quoted: %s", script)
	}
}
