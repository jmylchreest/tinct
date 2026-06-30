package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

var hexRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// m3Roles are the 16 Material-3 keys Noctalia expects per variant.
var m3Roles = []string{
	"mPrimary", "mOnPrimary", "mSecondary", "mOnSecondary",
	"mTertiary", "mOnTertiary", "mError", "mOnError",
	"mSurface", "mOnSurface", "mSurfaceVariant", "mOnSurfaceVariant",
	"mOutline", "mShadow", "mHover", "mOnHover",
}

func cc(hex string, r, g, b uint8, idx int) tinctplugin.CategorisedColour {
	return tinctplugin.CategorisedColour{RGB: tinctplugin.RGBColour{R: r, G: g, B: b}, Hex: hex, Index: idx}
}

// testPalette builds a PaletteData carrying every role the template reads,
// plus an indexed set so the ANSI matcher has candidates.
func testPalette(themeType string) tinctplugin.PaletteData {
	colours := map[string]tinctplugin.CategorisedColour{
		"accent1": cc("#b4befe", 180, 190, 254, 0), "onAccent1": cc("#11111b", 17, 17, 27, 1),
		"accent2": cc("#f5bde6", 245, 189, 230, 2), "onAccent2": cc("#11111b", 17, 17, 27, 3),
		"accent3": cc("#c6a0f6", 198, 160, 246, 4), "onAccent3": cc("#11111b", 17, 17, 27, 5),
		"accent4": cc("#8aadf4", 138, 173, 244, 6), "onAccent4": cc("#11111b", 17, 17, 27, 7),
		"danger": cc("#f38ba8", 243, 139, 168, 8), "onDanger": cc("#11111b", 17, 17, 27, 9),
		"surface": cc("#1e1e2e", 30, 30, 46, 10), "onSurface": cc("#cdd6f4", 205, 214, 244, 11),
		"surfaceVariant": cc("#313244", 49, 50, 68, 12), "onSurfaceVariant": cc("#a3b4eb", 163, 180, 235, 13),
		"outline": cc("#4c4f69", 76, 79, 105, 14), "shadow": cc("#11111b", 17, 17, 27, 15),
		"foreground": cc("#cdd6f4", 205, 214, 244, 16), "background": cc("#1e1e2e", 30, 30, 46, 17),
	}
	all := []tinctplugin.CategorisedColour{
		cc("#45475a", 69, 71, 90, 0), cc("#f38ba8", 243, 139, 168, 1),
		cc("#a6e3a1", 166, 227, 161, 2), cc("#f9e2af", 249, 226, 175, 3),
		cc("#89b4fa", 137, 180, 250, 4), cc("#f5c2e7", 245, 194, 231, 5),
		cc("#94e2d5", 148, 226, 213, 6), cc("#cdd6f4", 205, 214, 244, 7),
	}
	for role, c := range colours {
		c.Role = role
		all = append(all, c)
	}
	return tinctplugin.PaletteData{Colours: colours, AllColours: all, ThemeType: themeType}
}

func generate(t *testing.T, palette tinctplugin.PaletteData) map[string]any {
	t.Helper()
	dir := t.TempDir()
	p := &Plugin{version: "test", outputDir: dir}

	files, err := p.Generate(context.Background(), palette)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Generate() returned %d files, want 1", len(files))
	}

	// Generate returns path->content for the host to write; it does not
	// write the file itself (it only ensures the output dir exists).
	wantPath := filepath.Join(dir, "tinct.json")
	content, ok := files[wantPath]
	if !ok {
		t.Fatalf("Generate() did not return %s (got %v)", wantPath, keys(files))
	}
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("output dir not created: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("generated content is not valid JSON: %v", err)
	}
	return parsed
}

func assertVariant(t *testing.T, v any) {
	t.Helper()
	obj, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("variant is not an object: %T", v)
	}
	for _, role := range m3Roles {
		val, ok := obj[role].(string)
		if !ok {
			t.Errorf("variant missing role %q", role)
			continue
		}
		if !hexRe.MatchString(val) {
			t.Errorf("role %q = %q, want #rrggbb hex", role, val)
		}
	}
	term, ok := obj["terminal"].(map[string]any)
	if !ok {
		t.Fatalf("variant missing terminal block")
	}
	for _, k := range []string{"foreground", "background", "selectionFg", "selectionBg", "cursorText", "cursor"} {
		if s, ok := term[k].(string); !ok || !hexRe.MatchString(s) {
			t.Errorf("terminal.%s invalid: %v", k, term[k])
		}
	}
	for _, set := range []string{"normal", "bright"} {
		colours, ok := term[set].(map[string]any)
		if !ok {
			t.Fatalf("terminal.%s missing", set)
		}
		for _, name := range []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"} {
			if s, ok := colours[name].(string); !ok || !hexRe.MatchString(s) {
				t.Errorf("terminal.%s.%s invalid: %v", set, name, colours[name])
			}
		}
	}
}

func TestGenerateDualTheme(t *testing.T) {
	palette := testPalette("dark")
	palette.AlternateTheme = &tinctplugin.AlternateThemeData{
		Colours:    testPalette("light").Colours,
		AllColours: testPalette("light").AllColours,
		ThemeType:  "light",
	}

	parsed := generate(t, palette)
	if _, ok := parsed["dark"]; !ok {
		t.Fatal("dual-theme output missing 'dark' key")
	}
	if _, ok := parsed["light"]; !ok {
		t.Fatal("dual-theme output missing 'light' key")
	}
	assertVariant(t, parsed["dark"])
	assertVariant(t, parsed["light"])
}

func TestGenerateSingleDarkTheme(t *testing.T) {
	parsed := generate(t, testPalette("dark"))
	if _, ok := parsed["dark"]; !ok {
		t.Fatal("single-theme output missing required 'dark' key")
	}
	if _, ok := parsed["light"]; ok {
		t.Error("single dark theme should not emit 'light' key")
	}
	assertVariant(t, parsed["dark"])
}

func TestGenerateSingleLightFallsBackToDark(t *testing.T) {
	// A light-only palette must still produce the required 'dark' key.
	parsed := generate(t, testPalette("light"))
	if _, ok := parsed["dark"]; !ok {
		t.Fatal("light-only output must still populate required 'dark' key")
	}
	assertVariant(t, parsed["dark"])
}

func TestMetadata(t *testing.T) {
	p := &Plugin{version: "1.2.3"}
	m := p.GetMetadata()
	if m.Name != "noctalia" {
		t.Errorf("Name = %q, want noctalia", m.Name)
	}
	if m.Type != "output" {
		t.Errorf("Type = %q, want output", m.Type)
	}
	if m.Metadata == nil || m.Metadata.Reload == nil || m.Metadata.Reload.Method != "watch" {
		t.Errorf("expected reload method 'watch', got %+v", m.Metadata)
	}
	if !strings.Contains(m.Metadata.DefaultOutputDir, "noctalia") {
		t.Errorf("DefaultOutputDir = %q, want it to contain noctalia", m.Metadata.DefaultOutputDir)
	}
}

func keys(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
