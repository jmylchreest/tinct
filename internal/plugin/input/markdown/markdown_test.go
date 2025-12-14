package markdown

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input"
)

func TestPlugin_Name(t *testing.T) {
	p := New()
	if got := p.Name(); got != "markdown" {
		t.Errorf("Name() = %v, want markdown", got)
	}
}

func TestPlugin_Description(t *testing.T) {
	p := New()
	if got := p.Description(); got == "" {
		t.Error("Description() returned empty string")
	}
}

func TestPlugin_Version(t *testing.T) {
	p := New()
	if got := p.Version(); got == "" {
		t.Error("Version() returned empty string")
	}
}

func TestPlugin_Validate(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "missing path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path provided",
			path:    "/some/path.md",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			p.path = tt.path
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlugin_Generate(t *testing.T) {
	// Create a temp directory for test files
	tmpDir := t.TempDir()

	// Create a test theme file
	themeContent := `---
name: Test Theme
version: "1.0"
generator: tinct
generated: 2024-01-15T10:30:00Z
colors:
  - hex: "#1a1b26"
    weight: 0.25
    role: background
  - hex: "#c0caf5"
    weight: 0.15
    role: foreground
  - hex: "#7aa2f7"
    weight: 0.10
    role: accent_1
---

# Test Theme

Test markdown content.
`

	themePath := filepath.Join(tmpDir, "test-theme.md")
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("Failed to write test theme file: %v", err)
	}

	p := New()
	p.path = themePath
	p.extractWallpaper = false

	opts := input.GenerateOptions{
		Verbose: false,
	}

	palette, err := p.Generate(context.Background(), opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if palette == nil {
		t.Fatal("Generate() returned nil palette")
	}

	// Check that colors were extracted
	if len(palette.Colors) != 3 {
		t.Errorf("Generate() extracted %d colors, want 3", len(palette.Colors))
	}
}

func TestPlugin_Generate_WithWallpaper(t *testing.T) {
	tmpDir := t.TempDir()

	// Base64 encoded minimal PNG (1x1 pixel)
	base64PNG := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADklEQVR4nGNgYGBgYGAAABAAA0AAAE2lAAAAAElFTkSuQmCC"

	themeContent := `---
name: Wallpaper Theme
version: "1.0"
generator: tinct
generated: 2024-01-15T10:30:00Z
wallpaper:
  embedded: true
  format: png
  width: 1
  height: 1
  data: ` + base64PNG + `
colors:
  - hex: "#ffffff"
    weight: 1.0
---

# Wallpaper Theme
`

	themePath := filepath.Join(tmpDir, "wallpaper-theme.md")
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("Failed to write test theme file: %v", err)
	}

	p := New()
	p.path = themePath
	p.extractWallpaper = true
	p.wallpaperDir = tmpDir

	opts := input.GenerateOptions{
		Verbose: false,
	}

	palette, err := p.Generate(context.Background(), opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if palette == nil {
		t.Fatal("Generate() returned nil palette")
	}

	// Check that wallpaper was extracted
	wallpaperPath := p.WallpaperPath()
	if wallpaperPath == "" {
		t.Error("WallpaperPath() returned empty string")
	}

	// Verify wallpaper file exists
	if _, err := os.Stat(wallpaperPath); os.IsNotExist(err) {
		t.Errorf("Wallpaper file not created at: %s", wallpaperPath)
	}
}

func TestPlugin_Generate_ExternalWallpaper(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an external wallpaper file
	wallpaperPath := filepath.Join(tmpDir, "external-wallpaper.png")
	// Minimal PNG content
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde,
		0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, 0x54,
		0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00, 0x00,
		0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
		0xae, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(wallpaperPath, pngData, 0644); err != nil {
		t.Fatalf("Failed to write wallpaper file: %v", err)
	}

	themeContent := `---
name: External Wallpaper Theme
version: "1.0"
generator: tinct
generated: 2024-01-15T10:30:00Z
wallpaper:
  embedded: false
  path: external-wallpaper.png
colors:
  - hex: "#000000"
    weight: 1.0
---

# External Wallpaper Theme
`

	themePath := filepath.Join(tmpDir, "external-theme.md")
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("Failed to write test theme file: %v", err)
	}

	p := New()
	p.path = themePath
	p.extractWallpaper = false

	opts := input.GenerateOptions{
		Verbose: false,
	}

	palette, err := p.Generate(context.Background(), opts)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if palette == nil {
		t.Fatal("Generate() returned nil palette")
	}

	// Check that wallpaper path is set to external file
	gotPath := p.WallpaperPath()
	expectedPath := wallpaperPath
	if gotPath != expectedPath {
		t.Errorf("WallpaperPath() = %v, want %v", gotPath, expectedPath)
	}
}

func TestPlugin_Generate_InvalidTheme(t *testing.T) {
	tmpDir := t.TempDir()

	// Invalid theme file (no front matter)
	themeContent := `# Just a markdown file

No YAML front matter here.
`

	themePath := filepath.Join(tmpDir, "invalid-theme.md")
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("Failed to write test theme file: %v", err)
	}

	p := New()
	p.path = themePath

	opts := input.GenerateOptions{
		Verbose: false,
	}

	_, err := p.Generate(context.Background(), opts)
	if err == nil {
		t.Error("Generate() expected error for invalid theme file")
	}
}

func TestPlugin_Generate_NoColors(t *testing.T) {
	tmpDir := t.TempDir()

	// Theme with no colors
	themeContent := `---
name: Empty Theme
version: "1.0"
generator: tinct
generated: 2024-01-15T10:30:00Z
colors: []
---

# Empty Theme
`

	themePath := filepath.Join(tmpDir, "empty-theme.md")
	if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
		t.Fatalf("Failed to write test theme file: %v", err)
	}

	p := New()
	p.path = themePath

	opts := input.GenerateOptions{
		Verbose: false,
	}

	_, err := p.Generate(context.Background(), opts)
	if err == nil {
		t.Error("Generate() expected error for theme with no colors")
	}
}

func TestParseHex(t *testing.T) {
	tests := []struct {
		input   string
		wantR   uint8
		wantG   uint8
		wantB   uint8
		wantErr bool
	}{
		{"#1a1b26", 0x1a, 0x1b, 0x26, false},
		{"1a1b26", 0x1a, 0x1b, 0x26, false},
		{"#fff", 0xff, 0xff, 0xff, false},
		{"fff", 0xff, 0xff, 0xff, false},
		{"#000000", 0x00, 0x00, 0x00, false},
		{"#FFFFFF", 0xff, 0xff, 0xff, false},
		{"invalid", 0, 0, 0, true},
		{"#12345", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseHex(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHex(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.R != tt.wantR || got.G != tt.wantG || got.B != tt.wantB {
				t.Errorf("parseHex(%q) = (%d, %d, %d), want (%d, %d, %d)",
					tt.input, got.R, got.G, got.B, tt.wantR, tt.wantG, tt.wantB)
			}
		})
	}
}

func TestPlugin_GetFlagHelp(t *testing.T) {
	p := New()
	help := p.GetFlagHelp()

	if len(help) == 0 {
		t.Error("GetFlagHelp() returned empty slice")
	}

	// Check for required flags
	hasPath := false
	for _, h := range help {
		if h.Name == "markdown.path" {
			hasPath = true
			if !h.Required {
				t.Error("markdown.path should be required")
			}
		}
	}

	if !hasPath {
		t.Error("GetFlagHelp() missing markdown.path flag")
	}
}
