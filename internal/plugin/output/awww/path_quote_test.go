package awww

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
	"github.com/jmylchreest/tinct/pkg/util/subprocess"
)

func TestWallpaperPathShellQuoting(t *testing.T) {
	path := "/tmp/it's $(printf surprise);wallpaper.png"
	data := colour.NewThemeData(plugintesting.CreateTestPalette(colour.ThemeDark), path, "", "")
	data.WallpaperPath = path
	files, err := New().Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	script := string(files["tinct-awww.sh"])
	if !strings.Contains(script, "TINCT_AWWW_WALLPAPER="+subprocess.ShellQuote(path)) {
		t.Fatalf("wallpaper path is not shell quoted: %s", script)
	}
}

// Execute the generated script against a fake client so shell parsing itself,
// rather than matching the quoting helper, proves every value stays literal.
func TestGeneratedScriptPreservesArguments(t *testing.T) {
	dir := t.TempDir()
	executable := filepath.Join(dir, "awww")
	stub := "#!/bin/sh\nif [ \"$1\" = query ]; then exit 0; fi\nprintf '%s\\0' \"$@\"\n"
	if err := os.WriteFile(executable, []byte(stub), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	value := "it's $(printf injected); $HOME `printf bad`\nnext"
	p := New()
	p.transitionType, p.transitionStep, p.transitionFPS = value, value, value
	p.resizeMode, p.fillColor = value, value
	path := "/tmp/" + value + ".png"
	data := colour.NewThemeData(plugintesting.CreateTestPalette(colour.ThemeDark), path, "", "")
	data.WallpaperPath = path
	files, err := p.Generate(data)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := subprocess.CommandContext(ctx, "sh")
	cmd.Stdin = strings.NewReader(string(files["tinct-awww.sh"]))
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")
	want := []string{"img", "--transition-type", value, "--transition-step", value, "--transition-fps", value, "--resize", value, "--fill-color", value, path}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("arguments = %q; want %q", got, want)
	}
}
