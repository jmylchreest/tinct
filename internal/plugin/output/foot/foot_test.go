package foot

import (
	"regexp"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/shared/testing"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
)

// TestFootPlugin runs the shared plugin test battery.
func TestFootPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "foot",
		ExpectedFiles:      []string{"tinct-colors.ini", "tinct.ini"},
		ExpectedBinaryName: "foot",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestFootPlugin_PaletteContent verifies the palette fragment carries
// every key foot expects in its [colors] section, using foot's
// hex-without-# format.
func TestFootPlugin_PaletteContent(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, ok := files["tinct-colors.ini"]
	if !ok {
		t.Fatalf("Generate() did not return tinct-colors.ini")
	}
	str := string(content)

	requiredKeys := []string{
		"[colors]",
		"foreground=",
		"background=",
		"selection-foreground=",
		"selection-background=",
		"urls=",
		"regular0=", "regular1=", "regular2=", "regular3=",
		"regular4=", "regular5=", "regular6=", "regular7=",
		"bright0=", "bright1=", "bright2=", "bright3=",
		"bright4=", "bright5=", "bright6=", "bright7=",
	}
	for _, key := range requiredKeys {
		if !strings.Contains(str, key) {
			t.Errorf("palette missing required key: %s", key)
		}
	}

	// foot rejects values with a leading #. Verify the template emits
	// bare 6-digit hex (i.e. no `=#` anywhere in the [colors] section).
	if strings.Contains(str, "=#") {
		t.Error("palette contains '=#' — foot expects hex without the # prefix")
	}

	// Every assignment in the [colors] section should be exactly 6 hex
	// chars. Sample by checking a couple of well-known keys with a regex.
	hexRE := regexp.MustCompile(`(?m)^(foreground|background|regular0|bright7)=[0-9a-fA-F]{6}\s*$`)
	for _, key := range []string{"foreground", "background", "regular0", "bright7"} {
		line := key + "="
		if !strings.Contains(str, line) {
			continue
		}
		if !hexRE.MatchString(str) {
			t.Errorf("expected `%s=<6 hex chars>`, didn't match in:\n%s", key, str)
			break
		}
	}
}

// TestFootPlugin_ExampleConfigContent verifies the example main config
// references the palette fragment via include= and lands at the top of
// the file (foot processes include= directives in order).
func TestFootPlugin_ExampleConfigContent(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content, ok := files["tinct.ini"]
	if !ok {
		t.Fatalf("Generate() did not return tinct.ini")
	}
	str := string(content)

	if !strings.Contains(str, "include=") {
		t.Error("example config missing include= directive")
	}
	if !strings.Contains(str, "tinct-colors.ini") {
		t.Error("example config does not reference tinct-colors.ini")
	}
	if !strings.Contains(str, "font=") {
		t.Error("example config missing reasonable font= default")
	}
	if !strings.Contains(str, "[scrollback]") {
		t.Error("example config missing reasonable [scrollback] default")
	}
}

// TestFootPlugin_Hooks verifies the declarative hooks spec wires up
// SIGUSR1 reload (foot's documented colour-reload signal) and an
// InstructionsFn for the user-facing include-line reminder.
func TestFootPlugin_Hooks(t *testing.T) {
	plugin := New()
	spec := plugin.Hooks()

	if len(spec.RequiredBinaries) != 1 || spec.RequiredBinaries[0] != "foot" {
		t.Errorf("RequiredBinaries = %v, want [foot]", spec.RequiredBinaries)
	}
	if !spec.AutoCreateDir {
		t.Error("AutoCreateDir should be true")
	}
	if spec.Reload == nil {
		t.Fatal("Reload spec should be set")
	}
	if spec.Reload.Verb != hooks.VerbSignal {
		t.Errorf("Reload.Verb = %v, want %v", spec.Reload.Verb, hooks.VerbSignal)
	}
	wantArgs := []string{"foot", "SIGUSR1"}
	if len(spec.Reload.Args) != len(wantArgs) {
		t.Fatalf("Reload.Args = %v, want %v", spec.Reload.Args, wantArgs)
	}
	for i, a := range wantArgs {
		if spec.Reload.Args[i] != a {
			t.Errorf("Reload.Args[%d] = %q, want %q", i, spec.Reload.Args[i], a)
		}
	}
	if spec.InstructionsFn == nil {
		t.Fatal("InstructionsFn should be set so users see the include= reminder")
	}
	msg := spec.InstructionsFn(hooks.Context{})
	if !strings.Contains(msg, "include=") {
		t.Errorf("InstructionsFn message should mention include=, got: %q", msg)
	}
}

// TestFootPlugin_CustomOutputDir confirms the --foot.output-dir override
// path is honoured.
func TestFootPlugin_CustomOutputDir(t *testing.T) {
	plugin := New()
	plugin.outputDir = "/custom/foot"

	if got := plugin.DefaultOutputDir(); got != "/custom/foot" {
		t.Errorf("DefaultOutputDir() = %s, want /custom/foot", got)
	}
}

// TestFootPlugin_GetEmbeddedTemplates verifies both templates are
// embedded so `tinct plugins templates dump` can extract them.
func TestFootPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	want := map[string]bool{
		"tinct-colors.ini.tmpl": false,
		"tinct.ini.tmpl":        false,
	}
	for _, e := range entries {
		if _, ok := want[e.Name()]; ok {
			want[e.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("embedded template %s missing", name)
		}
	}
}
