package main

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

// newTestPalette creates a comprehensive PaletteData with all roles
// required by the OpenCode theme template.
func newTestPalette(themeType string) tinctplugin.PaletteData {
	colours := map[string]tinctplugin.CategorisedColour{
		// Core
		"background":      {Hex: "#1e1e2e", RGB: tinctplugin.RGBColour{R: 30, G: 30, B: 46}},
		"backgroundMuted": {Hex: "#181825", RGB: tinctplugin.RGBColour{R: 24, G: 24, B: 37}},
		"foreground":      {Hex: "#cdd6f4", RGB: tinctplugin.RGBColour{R: 205, G: 214, B: 244}},
		"foregroundMuted": {Hex: "#a6adc8", RGB: tinctplugin.RGBColour{R: 166, G: 173, B: 200}},

		// Surface / container hierarchy
		"surface":              {Hex: "#313244", RGB: tinctplugin.RGBColour{R: 49, G: 50, B: 68}},
		"surfaceContainer":     {Hex: "#45475a", RGB: tinctplugin.RGBColour{R: 69, G: 71, B: 90}},
		"surfaceContainerLow":  {Hex: "#313244", RGB: tinctplugin.RGBColour{R: 49, G: 50, B: 68}},
		"surfaceContainerHigh": {Hex: "#585b70", RGB: tinctplugin.RGBColour{R: 88, G: 91, B: 112}},
		"onSurface":            {Hex: "#cdd6f4", RGB: tinctplugin.RGBColour{R: 205, G: 214, B: 244}},
		"onSurfaceVariant":     {Hex: "#a6adc8", RGB: tinctplugin.RGBColour{R: 166, G: 173, B: 200}},

		// Borders / outlines
		"outline":        {Hex: "#6c7086", RGB: tinctplugin.RGBColour{R: 108, G: 112, B: 134}},
		"outlineVariant": {Hex: "#585b70", RGB: tinctplugin.RGBColour{R: 88, G: 91, B: 112}},
		"borderMuted":    {Hex: "#45475a", RGB: tinctplugin.RGBColour{R: 69, G: 71, B: 90}},

		// Accents
		"accent1": {Hex: "#89b4fa", RGB: tinctplugin.RGBColour{R: 137, G: 180, B: 250}},
		"accent2": {Hex: "#b4befe", RGB: tinctplugin.RGBColour{R: 180, G: 190, B: 254}},
		"accent3": {Hex: "#94e2d5", RGB: tinctplugin.RGBColour{R: 148, G: 226, B: 213}},
		"accent4": {Hex: "#cba6f7", RGB: tinctplugin.RGBColour{R: 203, G: 166, B: 247}},

		// Semantic
		"danger":  {Hex: "#f38ba8", RGB: tinctplugin.RGBColour{R: 243, G: 139, B: 168}},
		"warning": {Hex: "#f9e2af", RGB: tinctplugin.RGBColour{R: 249, G: 226, B: 175}},
		"success": {Hex: "#a6e3a1", RGB: tinctplugin.RGBColour{R: 166, G: 227, B: 161}},
		"info":    {Hex: "#89dceb", RGB: tinctplugin.RGBColour{R: 137, G: 220, B: 235}},
	}

	// Build AllColours list.
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

func TestPlugin_GetMetadata(t *testing.T) {
	p := &Plugin{
		version: "1.0.0",
		commit:  "abc123",
		date:    "2024-01-01",
	}

	meta := p.GetMetadata()

	if meta.Name != "opencode" {
		t.Errorf("expected name 'opencode', got %q", meta.Name)
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
}

func TestPlugin_GetFlagHelp(t *testing.T) {
	p := &Plugin{}
	flags := p.GetFlagHelp()

	if flags == nil {
		t.Fatal("expected non-nil slice")
	}
	if len(flags) != 1 {
		t.Fatalf("expected 1 flag, got %d", len(flags))
	}
	if flags[0].Name != "output-dir" {
		t.Errorf("expected flag name 'output-dir', got %q", flags[0].Name)
	}
}

func TestPlugin_PreExecute_NoConfig(t *testing.T) {
	// Point XDG_CONFIG_HOME to a temp dir with no opencode config.
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Also ensure ~/.config/opencode doesn't influence.
	oldHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	p := &Plugin{}
	ctx := context.Background()

	skip, reason, err := p.PreExecute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !skip {
		t.Error("expected skip=true when no OpenCode config exists")
	}
	if reason == "" {
		t.Error("expected non-empty skip reason")
	}
}

func TestPlugin_PreExecute_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake opencode config directory.
	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	p := &Plugin{}
	ctx := context.Background()

	skip, _, err := p.PreExecute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skip {
		t.Error("expected skip=false when OpenCode config exists")
	}
}

func TestPlugin_PostExecute(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()

	err := p.PostExecute(ctx, []string{"/tmp/test.json"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlugin_Generate_DarkTheme(t *testing.T) {
	tmpDir := t.TempDir()

	// Create opencode config directory so detection works.
	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Isolate from real user config.
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	p := &Plugin{version: "1.0.0", commit: "test", date: "2024-01-01"}
	palette := newTestPalette("dark")

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	// Verify file path.
	expectedPath := filepath.Join(tmpDir, "opencode", "themes", "tinct.json")
	content, ok := files[expectedPath]
	if !ok {
		t.Fatalf("expected file at %q, got files: %v", expectedPath, fileKeys(files))
	}

	// Verify it's valid JSON.
	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("generated file is not valid JSON: %v\nContent:\n%s", err, string(content))
	}

	// Verify schema.
	schema, ok := parsed["$schema"].(string)
	if !ok || schema != "https://opencode.ai/theme.json" {
		t.Errorf("expected schema 'https://opencode.ai/theme.json', got %v", parsed["$schema"])
	}

	// Verify defs section exists and has expected keys.
	defs, ok := parsed["defs"].(map[string]any)
	if !ok {
		t.Fatal("expected defs section in theme JSON")
	}

	expectedDefs := []string{
		"background", "backgroundMuted", "foreground", "foregroundMuted",
		"surface", "surfaceContainer", "surfaceContainerLow", "surfaceContainerHigh",
		"onSurface", "onSurfaceVariant", "outline", "outlineVariant",
		"accent1", "accent2", "accent3", "accent4",
		"danger", "warning", "success", "info", "borderMuted",
	}
	for _, key := range expectedDefs {
		val, exists := defs[key]
		if !exists {
			t.Errorf("defs missing key %q", key)
			continue
		}
		hexStr, ok := val.(string)
		if !ok || !strings.HasPrefix(hexStr, "#") {
			t.Errorf("defs[%q] should be a hex color string, got %v", key, val)
		}
	}

	// Verify theme section exists.
	theme, ok := parsed["theme"].(map[string]any)
	if !ok {
		t.Fatal("expected theme section in theme JSON")
	}

	// Verify required OpenCode theme tokens.
	requiredTokens := []string{
		"primary", "secondary", "accent", "text", "textMuted", "background",
	}
	for _, key := range requiredTokens {
		if _, exists := theme[key]; !exists {
			t.Errorf("theme missing required token %q", key)
		}
	}

	// Verify theme references point to defs.
	refChecks := map[string]string{
		"primary":   "accent1",
		"secondary": "accent2",
		"accent":    "accent3",
		"error":     "danger",
		"text":      "foreground",
		"textMuted": "foregroundMuted",
		"border":    "outline",
	}
	for token, expectedRef := range refChecks {
		val, ok := theme[token].(string)
		if !ok {
			t.Errorf("theme[%q] should be a string, got %T", token, theme[token])
			continue
		}
		if val != expectedRef {
			t.Errorf("theme[%q] = %q, expected %q", token, val, expectedRef)
		}
	}
}

func TestPlugin_Generate_LightTheme(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	p := &Plugin{version: "1.0.0", commit: "test", date: "2024-01-01"}
	palette := newTestPalette("light")

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "opencode", "themes", "tinct.json")
	content, ok := files[expectedPath]
	if !ok {
		t.Fatalf("expected file at %q", expectedPath)
	}

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("generated file is not valid JSON: %v", err)
	}

	// Verify the defs contain hex colors (same validation regardless of theme type).
	defs, ok := parsed["defs"].(map[string]any)
	if !ok {
		t.Fatal("expected defs section")
	}
	if _, ok := defs["background"].(string); !ok {
		t.Error("defs[background] should be a string")
	}
}

func TestPlugin_Generate_DiffColors(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	p := &Plugin{version: "1.0.0", commit: "test", date: "2024-01-01"}
	palette := newTestPalette("dark")

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "opencode", "themes", "tinct.json")
	content := files[expectedPath]

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}

	theme := parsed["theme"].(map[string]any)

	// Diff bg colors should be hex strings with alpha (inline-rendered, not refs).
	diffAddedBg, ok := theme["diffAddedBg"].(string)
	if !ok {
		t.Fatal("diffAddedBg should be a string")
	}
	if !strings.HasPrefix(diffAddedBg, "#") {
		t.Errorf("diffAddedBg should be a hex color, got %q", diffAddedBg)
	}
	// hexAlpha produces 9-char strings: #RRGGBBAA
	if len(diffAddedBg) != 9 {
		t.Errorf("diffAddedBg should be #RRGGBBAA (9 chars), got %q (%d chars)", diffAddedBg, len(diffAddedBg))
	}

	diffRemovedBg, ok := theme["diffRemovedBg"].(string)
	if !ok {
		t.Fatal("diffRemovedBg should be a string")
	}
	if len(diffRemovedBg) != 9 {
		t.Errorf("diffRemovedBg should be #RRGGBBAA (9 chars), got %q", diffRemovedBg)
	}

	// Diff reference tokens should point to defs.
	if theme["diffAdded"] != "success" {
		t.Errorf("diffAdded should reference 'success', got %v", theme["diffAdded"])
	}
	if theme["diffRemoved"] != "danger" {
		t.Errorf("diffRemoved should reference 'danger', got %v", theme["diffRemoved"])
	}
}

func TestPlugin_Generate_SyntaxHighlighting(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	p := &Plugin{version: "1.0.0", commit: "test", date: "2024-01-01"}
	palette := newTestPalette("dark")

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "opencode", "themes", "tinct.json")
	content := files[expectedPath]

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}

	theme := parsed["theme"].(map[string]any)

	syntaxChecks := map[string]string{
		"syntaxComment":     "foregroundMuted",
		"syntaxKeyword":     "accent4",
		"syntaxFunction":    "accent1",
		"syntaxString":      "success",
		"syntaxNumber":      "accent3",
		"syntaxType":        "accent2",
		"syntaxVariable":    "foreground",
		"syntaxOperator":    "onSurface",
		"syntaxPunctuation": "onSurfaceVariant",
	}

	for token, expectedRef := range syntaxChecks {
		val, ok := theme[token].(string)
		if !ok {
			t.Errorf("theme[%q] should be a string", token)
			continue
		}
		if val != expectedRef {
			t.Errorf("theme[%q] = %q, expected %q", token, val, expectedRef)
		}
	}
}

func TestPlugin_Generate_MarkdownTokens(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	p := &Plugin{version: "1.0.0", commit: "test", date: "2024-01-01"}
	palette := newTestPalette("dark")

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "opencode", "themes", "tinct.json")
	content := files[expectedPath]

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}

	theme := parsed["theme"].(map[string]any)

	markdownChecks := map[string]string{
		"markdownText":    "foreground",
		"markdownHeading": "accent1",
		"markdownLink":    "accent2",
		"markdownCode":    "success",
		"markdownEmph":    "accent4",
		"markdownStrong":  "accent1",
	}

	for token, expectedRef := range markdownChecks {
		val, ok := theme[token].(string)
		if !ok {
			t.Errorf("theme[%q] should be a string", token)
			continue
		}
		if val != expectedRef {
			t.Errorf("theme[%q] = %q, expected %q", token, val, expectedRef)
		}
	}
}

func TestPlugin_Generate_CustomOutputDir(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "custom", "themes")

	p := &Plugin{
		version:   "1.0.0",
		commit:    "test",
		date:      "2024-01-01",
		outputDir: outputDir,
	}
	palette := newTestPalette("dark")

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedPath := filepath.Join(outputDir, "tinct.json")
	if _, ok := files[expectedPath]; !ok {
		t.Fatalf("expected file at %q, got files: %v", expectedPath, fileKeys(files))
	}
}

func TestPlugin_Generate_DualTheme(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	// Build the primary (dark) palette.
	palette := newTestPalette("dark")

	// Build an alternate (light) palette.
	altColours := make(map[string]tinctplugin.CategorisedColour, len(palette.Colours))
	maps.Copy(altColours, palette.Colours)
	// Override a few to differentiate.
	altColours["background"] = tinctplugin.CategorisedColour{
		Hex: "#eff1f5", RGB: tinctplugin.RGBColour{R: 239, G: 241, B: 245}, Role: "background",
	}
	altColours["foreground"] = tinctplugin.CategorisedColour{
		Hex: "#4c4f69", RGB: tinctplugin.RGBColour{R: 76, G: 79, B: 105}, Role: "foreground",
	}

	altAllColours := make([]tinctplugin.CategorisedColour, 0, len(altColours))
	for _, cc := range altColours {
		altAllColours = append(altAllColours, cc)
	}

	palette.AlternateTheme = &tinctplugin.AlternateThemeData{
		Colours:    altColours,
		AllColours: altAllColours,
		ThemeType:  "light",
	}

	p := &Plugin{version: "1.0.0", commit: "test", date: "2024-01-01"}
	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should produce 2 files: tinct.json (dark primary) + tinct-light.json (alternate).
	if len(files) != 2 {
		t.Fatalf("expected 2 files for dual-theme, got %d: %v", len(files), fileKeys(files))
	}

	primaryPath := filepath.Join(tmpDir, "opencode", "themes", "tinct.json")
	altPath := filepath.Join(tmpDir, "opencode", "themes", "tinct-light.json")

	if _, ok := files[primaryPath]; !ok {
		t.Errorf("missing primary theme at %q", primaryPath)
	}
	if _, ok := files[altPath]; !ok {
		t.Errorf("missing alternate theme at %q", altPath)
	}

	// Verify alternate theme has different background color.
	altContent := files[altPath]
	var altParsed map[string]any
	if err := json.Unmarshal(altContent, &altParsed); err != nil {
		t.Fatalf("alternate theme is not valid JSON: %v", err)
	}

	altDefs, ok := altParsed["defs"].(map[string]any)
	if !ok {
		t.Fatal("alternate theme missing defs section")
	}
	if altDefs["background"] != "#eff1f5" {
		t.Errorf("alternate background should be #eff1f5, got %v", altDefs["background"])
	}
}

func TestPlugin_Generate_ValidJSON(t *testing.T) {
	// Verify the output is well-formed JSON against the OpenCode schema structure.
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	p := &Plugin{version: "1.0.0", commit: "test", date: "2024-01-01"}
	palette := newTestPalette("dark")

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "opencode", "themes", "tinct.json")
	content := files[expectedPath]

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Top-level structure: $schema, defs, theme
	topKeys := []string{"$schema", "defs", "theme"}
	for _, key := range topKeys {
		if _, ok := parsed[key]; !ok {
			t.Errorf("missing top-level key %q", key)
		}
	}

	// No unexpected top-level keys.
	for key := range parsed {
		switch key {
		case "$schema", "defs", "theme":
			// expected
		default:
			t.Errorf("unexpected top-level key %q", key)
		}
	}
}

func TestPlugin_Generate_AllDefsAreHexColors(t *testing.T) {
	tmpDir := t.TempDir()

	configDir := filepath.Join(tmpDir, "opencode")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	p := &Plugin{version: "1.0.0", commit: "test", date: "2024-01-01"}
	palette := newTestPalette("dark")

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "opencode", "themes", "tinct.json")
	content := files[expectedPath]

	var parsed map[string]any
	if err := json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}

	defs := parsed["defs"].(map[string]any)
	for key, val := range defs {
		hexStr, ok := val.(string)
		if !ok {
			t.Errorf("defs[%q] is not a string: %T", key, val)
			continue
		}
		if !strings.HasPrefix(hexStr, "#") || len(hexStr) != 7 {
			t.Errorf("defs[%q] = %q is not a valid #RRGGBB hex color", key, hexStr)
		}
	}
}

func TestPlugin_Generate_CreatesThemesDir(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "deeply", "nested", "themes")

	p := &Plugin{
		version:   "1.0.0",
		commit:    "test",
		date:      "2024-01-01",
		outputDir: outputDir,
	}
	palette := newTestPalette("dark")

	ctx := context.Background()
	_, err := p.Generate(ctx, palette)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify the directory was created.
	info, err := os.Stat(outputDir)
	if err != nil {
		t.Fatalf("output directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("output path is not a directory")
	}
}

// fileKeys returns the keys of a map for diagnostic output.
func fileKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
