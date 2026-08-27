package main

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

// newTestPalette creates a PaletteData with every role the spicetify
// template references.
func newTestPalette(themeType string) tinctplugin.PaletteData {
	colours := map[string]tinctplugin.CategorisedColour{
		"background":           {Hex: "#1e1e2e", RGB: tinctplugin.RGBColour{R: 30, G: 30, B: 46}},
		"backgroundMuted":      {Hex: "#181825", RGB: tinctplugin.RGBColour{R: 24, G: 24, B: 37}},
		"foreground":           {Hex: "#cdd6f4", RGB: tinctplugin.RGBColour{R: 205, G: 214, B: 244}},
		"foregroundMuted":      {Hex: "#a6adc8", RGB: tinctplugin.RGBColour{R: 166, G: 173, B: 200}},
		"surface":              {Hex: "#313244", RGB: tinctplugin.RGBColour{R: 49, G: 50, B: 68}},
		"surfaceContainer":     {Hex: "#45475a", RGB: tinctplugin.RGBColour{R: 69, G: 71, B: 90}},
		"surfaceContainerHigh": {Hex: "#585b70", RGB: tinctplugin.RGBColour{R: 88, G: 91, B: 112}},
		"accent1":              {Hex: "#89b4fa", RGB: tinctplugin.RGBColour{R: 137, G: 180, B: 250}},
		"accent2":              {Hex: "#b4befe", RGB: tinctplugin.RGBColour{R: 180, G: 190, B: 254}},
		"danger":               {Hex: "#f38ba8", RGB: tinctplugin.RGBColour{R: 243, G: 139, B: 168}},
		"info":                 {Hex: "#89dceb", RGB: tinctplugin.RGBColour{R: 137, G: 220, B: 235}},
	}

	allColours := make([]tinctplugin.CategorisedColour, 0, len(colours))
	i := 0
	for role, cc := range colours {
		cc.Role = role
		cc.Index = i
		colours[role] = cc
		allColours = append(allColours, cc)
		i++
	}

	return tinctplugin.PaletteData{
		Colours:    colours,
		AllColours: allColours,
		ThemeType:  themeType,
		Verbose:    false,
	}
}

// newMinimalTestPalette omits the optional surfaceContainer / surfaceContainerHigh
// roles to exercise the `has`-guarded fallback branches in the template.
func newMinimalTestPalette() tinctplugin.PaletteData {
	colours := map[string]tinctplugin.CategorisedColour{
		"background":      {Hex: "#1e1e2e", RGB: tinctplugin.RGBColour{R: 30, G: 30, B: 46}},
		"backgroundMuted": {Hex: "#181825", RGB: tinctplugin.RGBColour{R: 24, G: 24, B: 37}},
		"foreground":      {Hex: "#cdd6f4", RGB: tinctplugin.RGBColour{R: 205, G: 214, B: 244}},
		"foregroundMuted": {Hex: "#a6adc8", RGB: tinctplugin.RGBColour{R: 166, G: 173, B: 200}},
		"surface":         {Hex: "#313244", RGB: tinctplugin.RGBColour{R: 49, G: 50, B: 68}},
		"accent1":         {Hex: "#89b4fa", RGB: tinctplugin.RGBColour{R: 137, G: 180, B: 250}},
		"accent2":         {Hex: "#b4befe", RGB: tinctplugin.RGBColour{R: 180, G: 190, B: 254}},
		"danger":          {Hex: "#f38ba8", RGB: tinctplugin.RGBColour{R: 243, G: 139, B: 168}},
		"info":            {Hex: "#89dceb", RGB: tinctplugin.RGBColour{R: 137, G: 220, B: 235}},
	}

	allColours := make([]tinctplugin.CategorisedColour, 0, len(colours))
	i := 0
	for role, cc := range colours {
		cc.Role = role
		cc.Index = i
		colours[role] = cc
		allColours = append(allColours, cc)
		i++
	}

	return tinctplugin.PaletteData{
		Colours:    colours,
		AllColours: allColours,
		ThemeType:  "dark",
	}
}

func TestPlugin_GetMetadata(t *testing.T) {
	p := &Plugin{version: "1.0.0", commit: "abc123", date: "2024-01-01"}

	meta := p.GetMetadata()

	if meta.Name != "spicetify" {
		t.Errorf("expected name 'spicetify', got %q", meta.Name)
	}
	if meta.Type != "output" {
		t.Errorf("expected type 'output', got %q", meta.Type)
	}
	if meta.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", meta.Version)
	}
	if meta.PluginProtocol != "go-plugin" {
		t.Errorf("expected plugin_protocol 'go-plugin', got %q", meta.PluginProtocol)
	}
	if meta.ProtocolVersion != tinctplugin.ProtocolVersion {
		t.Errorf("expected protocol_version %q, got %q", tinctplugin.ProtocolVersion, meta.ProtocolVersion)
	}

	if meta.Metadata == nil {
		t.Fatal("expected Metadata block to be populated")
	}
	if len(meta.Metadata.RequiredBinaries) != 1 || meta.Metadata.RequiredBinaries[0] != "spicetify" {
		t.Errorf("expected required_binaries=[spicetify], got %v", meta.Metadata.RequiredBinaries)
	}
	if meta.Metadata.DefaultOutputDir != "~/.config/spicetify/Themes/tinct" {
		t.Errorf("unexpected default_output_dir: %q", meta.Metadata.DefaultOutputDir)
	}
	if len(meta.Metadata.GeneratedFiles) != 1 || meta.Metadata.GeneratedFiles[0] != "color.ini" {
		t.Errorf("expected generated_files=[color.ini], got %v", meta.Metadata.GeneratedFiles)
	}
	if meta.Metadata.Pattern != "single-file" {
		t.Errorf("expected pattern 'single-file', got %q", meta.Metadata.Pattern)
	}
	if meta.Metadata.Reload == nil {
		t.Fatal("expected reload metadata")
	}
	if meta.Metadata.Reload.Method != "none" {
		t.Errorf("expected reload.method 'none', got %q", meta.Metadata.Reload.Method)
	}
	if !meta.Metadata.Reload.UserActionRequired {
		t.Error("expected reload.user_action_required=true")
	}
}

func TestPlugin_GetFlagHelp(t *testing.T) {
	p := &Plugin{}
	flags := p.GetFlagHelp()

	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Name != "output-dir" {
		t.Errorf("expected flag 'output-dir', got %q", flags[0].Name)
	}
}

// Detection moved onto the declarative hooks.Spec: the shared runner
// evaluates RequiredBinaries before PreExecute is reached, so PreExecute
// itself must no longer gate. Two gates would be able to disagree.
func TestPlugin_HooksRequireSpicetify(t *testing.T) {
	spec := (&Plugin{}).Hooks()

	if !slices.Contains(spec.RequiredBinaries, "spicetify") {
		t.Errorf("RequiredBinaries = %v, want it to contain spicetify", spec.RequiredBinaries)
	}

	// Without spicetify the theme files cannot be applied at all, so the
	// binary is required rather than optional.
	if slices.Contains(spec.OptionalBinaries, "spicetify") {
		t.Error("spicetify is listed as optional; without it the theme cannot be applied")
	}
}

func TestPlugin_PreExecuteNeverSkips(t *testing.T) {
	t.Setenv("PATH", "/nonexistent")

	skip, reason, err := (&Plugin{}).PreExecute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skip {
		t.Errorf("PreExecute skipped (%q); the RequiredBinaries hook owns that decision", reason)
	}
}

// The manual-apply guidance moved to Hooks().Instructions, which the host
// prints in verbose mode after files are written.
func TestPlugin_InstructionsExplainManualApply(t *testing.T) {
	got := (&Plugin{}).Hooks().Instructions

	for _, want := range []string{"NOT applied", "Close Spotify", "spicetify apply"} {
		if !strings.Contains(got, want) {
			t.Errorf("Instructions missing %q; got:\n%s", want, got)
		}
	}
}

// PostExecute is a no-op: the guidance lives in Hooks().Instructions and
// `spicetify apply` is deliberately never invoked for the user.
func TestPlugin_PostExecuteIsNoop(t *testing.T) {
	if err := (&Plugin{}).PostExecute(context.Background(), []string{"/tmp/color.ini"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlugin_PostExecute_NoFilesIsNoop(t *testing.T) {
	p := &Plugin{}
	if err := p.PostExecute(context.Background(), nil); err != nil {
		t.Errorf("unexpected error with empty files slice: %v", err)
	}
}

func TestPlugin_Generate_DefaultOutputDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	p := &Plugin{version: "1.0.0"}
	files, err := p.Generate(context.Background(), newTestPalette("dark"))
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected exactly 1 file, got %d", len(files))
	}

	expected := filepath.Join(tmpHome, ".config", "spicetify", "Themes", "tinct", "color.ini")
	content, ok := files[expected]
	if !ok {
		t.Fatalf("expected file at %q, got: %v", expected, fileKeys(files))
	}

	body := string(content)
	if !strings.Contains(body, "[base]") {
		t.Error("expected [base] section in output")
	}
	if !strings.Contains(body, "text                = cdd6f4") {
		t.Errorf("expected text= cdd6f4 (no hash), output:\n%s", body)
	}
	if !strings.Contains(body, "main                = 1e1e2e") {
		t.Error("expected main= 1e1e2e")
	}
	if !strings.Contains(body, "button              = 89b4fa") {
		t.Error("expected button= 89b4fa (accent1)")
	}
	if !strings.Contains(body, "button-active       = b4befe") {
		t.Error("expected button-active= b4befe (accent2)")
	}
	if !strings.Contains(body, "notification-error  = f38ba8") {
		t.Error("expected notification-error= f38ba8 (danger)")
	}
	if !strings.Contains(body, "selected-row        = 585b70") {
		t.Error("expected selected-row to use surfaceContainerHigh")
	}
	if !strings.Contains(body, "tab-active          = 45475a") {
		t.Error("expected tab-active to use surfaceContainer")
	}

	// Values must be 6 hex chars with no leading #.
	for _, line := range strings.Split(body, "\n") {
		if !strings.Contains(line, "=") {
			continue
		}
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "[") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		val := strings.TrimSpace(parts[1])
		if strings.HasPrefix(val, "#") {
			t.Errorf("color value %q must not start with '#'", val)
		}
		if len(val) != 6 {
			t.Errorf("color value %q must be 6 hex chars, got %d", val, len(val))
		}
	}
}

func TestPlugin_Generate_CustomOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	custom := filepath.Join(tmpDir, "my", "themes", "tinct")

	p := &Plugin{outputDir: custom}
	files, err := p.Generate(context.Background(), newTestPalette("dark"))
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expected := filepath.Join(custom, "color.ini")
	if _, ok := files[expected]; !ok {
		t.Fatalf("expected file at %q, got: %v", expected, fileKeys(files))
	}

	info, err := os.Stat(custom)
	if err != nil {
		t.Fatalf("output dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("output path is not a directory")
	}
}

func TestPlugin_Generate_FallbackWhenSurfaceRolesMissing(t *testing.T) {
	tmpDir := t.TempDir()
	custom := filepath.Join(tmpDir, "themes", "tinct")
	p := &Plugin{outputDir: custom}

	files, err := p.Generate(context.Background(), newMinimalTestPalette())
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content := string(files[filepath.Join(custom, "color.ini")])

	// Both selected-row and tab-active should fall back to `surface` (313244).
	if !strings.Contains(content, "selected-row        = 313244") {
		t.Errorf("expected selected-row to fall back to surface, output:\n%s", content)
	}
	if !strings.Contains(content, "tab-active          = 313244") {
		t.Errorf("expected tab-active to fall back to surface, output:\n%s", content)
	}
}

func fileKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
