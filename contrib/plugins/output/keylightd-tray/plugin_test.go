package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

func TestPlugin_GetMetadata(t *testing.T) {
	p := &Plugin{
		version: "1.0.0",
		commit:  "abc123",
		date:    "2024-01-01",
	}

	meta := p.GetMetadata()

	if meta.Name != "keylightd-tray" {
		t.Errorf("expected name 'keylightd-tray', got %q", meta.Name)
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
}

func TestPlugin_PreExecute(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()

	skip, reason, err := p.PreExecute(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if skip {
		t.Errorf("expected skip=false, got skip=true with reason: %s", reason)
	}
}

func TestPlugin_PostExecute(t *testing.T) {
	p := &Plugin{}
	ctx := context.Background()

	err := p.PostExecute(ctx, []string{"/tmp/test.css"})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlugin_GetFlagHelp(t *testing.T) {
	p := &Plugin{}

	flags := p.GetFlagHelp()

	// Currently no flags, but should return empty slice not nil
	if flags == nil {
		t.Error("expected non-nil slice")
	}
}

func TestPlugin_Generate(t *testing.T) {
	// Create a temporary directory for output
	tmpDir := t.TempDir()

	// Set XDG_CONFIG_HOME to our temp directory
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	p := &Plugin{
		version: "1.0.0",
		commit:  "test",
		date:    "2024-01-01",
	}

	palette := tinctplugin.PaletteData{
		Colours: map[string]tinctplugin.CategorisedColour{
			"background": {
				Hex: "#1e1e2e",
				RGB: tinctplugin.RGBColour{R: 30, G: 30, B: 46},
			},
			"foreground": {
				Hex: "#cdd6f4",
				RGB: tinctplugin.RGBColour{R: 205, G: 214, B: 244},
			},
			"surface": {
				Hex: "#313244",
				RGB: tinctplugin.RGBColour{R: 49, G: 50, B: 68},
			},
			"subtle": {
				Hex: "#a6adc8",
				RGB: tinctplugin.RGBColour{R: 166, G: 173, B: 200},
			},
			"accent1": {
				Hex: "#89b4fa",
				RGB: tinctplugin.RGBColour{R: 137, G: 180, B: 250},
			},
			"green": {
				Hex: "#a6e3a1",
				RGB: tinctplugin.RGBColour{R: 166, G: 227, B: 161},
			},
			"yellow": {
				Hex: "#f9e2af",
				RGB: tinctplugin.RGBColour{R: 249, G: 226, B: 175},
			},
			"red": {
				Hex: "#f38ba8",
				RGB: tinctplugin.RGBColour{R: 243, G: 139, B: 168},
			},
			"overlay": {
				Hex: "#45475a",
				RGB: tinctplugin.RGBColour{R: 69, G: 71, B: 90},
			},
		},
		ThemeType: "dark",
		Verbose:   false,
	}

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	// Check file path
	expectedPath := filepath.Join(tmpDir, "keylightd", "keylightd-tray", "tinct-custom.css")
	content, ok := files[expectedPath]
	if !ok {
		t.Fatalf("expected file at %q, got files: %v", expectedPath, files)
	}

	// Check CSS content
	css := string(content)

	// Verify CSS variables are present
	expectedVars := []string{
		"--bg-primary: #1e1e2e",
		"--bg-secondary: #313244",
		"--text-primary: #cdd6f4",
		"--text-secondary: #a6adc8",
		"--accent: #89b4fa",
		"--success: #a6e3a1",
		"--warning: #f9e2af",
		"--error: #f38ba8",
		"--surface: #313244",
		"--overlay: #45475a",
	}

	for _, v := range expectedVars {
		if !strings.Contains(css, v) {
			t.Errorf("CSS missing variable %q", v)
		}
	}

	// Verify :root selector
	if !strings.Contains(css, ":root {") {
		t.Error("CSS missing :root selector")
	}

	// Verify header comment
	if !strings.Contains(css, "Tinct-generated theme") {
		t.Error("CSS missing header comment")
	}
}

func TestPlugin_Generate_MinimalPalette(t *testing.T) {
	// Test with minimal palette (only background and foreground)
	tmpDir := t.TempDir()

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	p := &Plugin{
		version: "1.0.0",
		commit:  "test",
		date:    "2024-01-01",
	}

	palette := tinctplugin.PaletteData{
		Colours: map[string]tinctplugin.CategorisedColour{
			"background": {
				Hex: "#000000",
				RGB: tinctplugin.RGBColour{R: 0, G: 0, B: 0},
			},
			"foreground": {
				Hex: "#ffffff",
				RGB: tinctplugin.RGBColour{R: 255, G: 255, B: 255},
			},
		},
		ThemeType: "dark",
	}

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	// Get the content
	var content []byte
	for _, c := range files {
		content = c
		break
	}

	css := string(content)

	// Should still have background and foreground
	if !strings.Contains(css, "--bg-primary: #000000") {
		t.Error("CSS missing --bg-primary")
	}
	if !strings.Contains(css, "--text-primary: #ffffff") {
		t.Error("CSS missing --text-primary")
	}
}

func TestPlugin_Generate_FallbackColors(t *testing.T) {
	// Test fallback behavior when optional colors are missing
	tmpDir := t.TempDir()

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	p := &Plugin{
		version: "1.0.0",
		commit:  "test",
		date:    "2024-01-01",
	}

	// Palette with accent colors instead of named colors (green, yellow, red)
	palette := tinctplugin.PaletteData{
		Colours: map[string]tinctplugin.CategorisedColour{
			"background": {
				Hex: "#1e1e2e",
				RGB: tinctplugin.RGBColour{R: 30, G: 30, B: 46},
			},
			"foreground": {
				Hex: "#cdd6f4",
				RGB: tinctplugin.RGBColour{R: 205, G: 214, B: 244},
			},
			"accent2": {
				Hex: "#a6e3a1",
				RGB: tinctplugin.RGBColour{R: 166, G: 227, B: 161},
			},
			"accent3": {
				Hex: "#f9e2af",
				RGB: tinctplugin.RGBColour{R: 249, G: 226, B: 175},
			},
			"accent4": {
				Hex: "#f38ba8",
				RGB: tinctplugin.RGBColour{R: 243, G: 139, B: 168},
			},
		},
		ThemeType: "dark",
	}

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Get the content
	var content []byte
	for _, c := range files {
		content = c
		break
	}

	css := string(content)

	// Should use accent2/3/4 as fallbacks for success/warning/error
	if !strings.Contains(css, "--success: #a6e3a1") {
		t.Error("CSS missing --success fallback from accent2")
	}
	if !strings.Contains(css, "--warning: #f9e2af") {
		t.Error("CSS missing --warning fallback from accent3")
	}
	if !strings.Contains(css, "--error: #f38ba8") {
		t.Error("CSS missing --error fallback from accent4")
	}
}

func TestGenerateCSS(t *testing.T) {
	p := &Plugin{}

	palette := tinctplugin.PaletteData{
		Colours: map[string]tinctplugin.CategorisedColour{
			"background": {Hex: "#000000"},
			"foreground": {Hex: "#ffffff"},
			"accent1":    {Hex: "#ff0000"},
		},
		ThemeType: "dark",
	}

	css := p.generateCSS(palette)

	// Check structure
	if !strings.HasPrefix(css, "/*") {
		t.Error("CSS should start with comment")
	}
	if !strings.Contains(css, ":root {") {
		t.Error("CSS should contain :root selector")
	}
	if !strings.HasSuffix(strings.TrimSpace(css), "}") {
		t.Error("CSS should end with closing brace")
	}
}

func TestGetTinctCSSPath(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	path, err := getTinctCSSPath()
	if err != nil {
		t.Fatalf("getTinctCSSPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "keylightd", "keylightd-tray", "tinct-custom.css")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestGetTinctCSSPath_DefaultHome(t *testing.T) {
	// Test with XDG_CONFIG_HOME unset
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Unsetenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	path, err := getTinctCSSPath()
	if err != nil {
		t.Fatalf("getTinctCSSPath failed: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "keylightd", "keylightd-tray", "tinct-custom.css")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestGetCustomCSSPath(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	path, err := getCustomCSSPath()
	if err != nil {
		t.Fatalf("getCustomCSSPath failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "keylightd", "keylightd-tray", "custom.css")
	if path != expected {
		t.Errorf("expected %q, got %q", expected, path)
	}
}

func TestPlugin_PostExecute_NoCustomCSS(t *testing.T) {
	// Test PostExecute when custom.css doesn't exist
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	p := &Plugin{}
	ctx := context.Background()

	err := p.PostExecute(ctx, []string{filepath.Join(tmpDir, "keylightd", "keylightd-tray", "tinct-custom.css")})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlugin_PostExecute_WithImport(t *testing.T) {
	// Test PostExecute when custom.css contains the import
	tmpDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Create the directory and custom.css with import
	configDir := filepath.Join(tmpDir, "keylightd", "keylightd-tray")
	os.MkdirAll(configDir, 0755)
	customCSS := filepath.Join(configDir, "custom.css")
	os.WriteFile(customCSS, []byte("@import url(\"tinct-custom.css\");\n"), 0644)

	p := &Plugin{}
	ctx := context.Background()

	err := p.PostExecute(ctx, []string{filepath.Join(configDir, "tinct-custom.css")})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPlugin_Generate_ComponentColors(t *testing.T) {
	// Test that component-specific colors are generated
	tmpDir := t.TempDir()

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	p := &Plugin{
		version: "1.0.0",
		commit:  "test",
		date:    "2024-01-01",
	}

	palette := tinctplugin.PaletteData{
		Colours: map[string]tinctplugin.CategorisedColour{
			"background": {
				Hex: "#1e1e2e",
				RGB: tinctplugin.RGBColour{R: 30, G: 30, B: 46},
			},
			"foreground": {
				Hex: "#cdd6f4",
				RGB: tinctplugin.RGBColour{R: 205, G: 214, B: 244},
			},
			"surface": {
				Hex: "#313244",
				RGB: tinctplugin.RGBColour{R: 49, G: 50, B: 68},
			},
			"surfaceContainerLowest": {
				Hex: "#11111b",
				RGB: tinctplugin.RGBColour{R: 17, G: 17, B: 27},
			},
			"surfaceContainerLow": {
				Hex: "#181825",
				RGB: tinctplugin.RGBColour{R: 24, G: 24, B: 37},
			},
			"borderMuted": {
				Hex: "#45475a",
				RGB: tinctplugin.RGBColour{R: 69, G: 71, B: 90},
			},
		},
		ThemeType: "dark",
	}

	ctx := context.Background()
	files, err := p.Generate(ctx, palette)

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Get the content
	var content []byte
	for _, c := range files {
		content = c
		break
	}

	css := string(content)

	// Check component-specific variables
	expectedVars := []string{
		"--slider-track: #181825",
		"--input-bg: #11111b",
		"--input-border: #45475a",
		"--list-item-bg: #11111b",
	}

	for _, v := range expectedVars {
		if !strings.Contains(css, v) {
			t.Errorf("CSS missing component variable %q", v)
		}
	}
}
