package themeformat

import (
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) { //nolint:gocognit // comprehensive parsing test
	tests := []struct {
		name    string
		input   string
		want    *Theme
		wantErr bool
	}{
		{
			name: "basic theme with colors",
			input: `---
name: Test Theme
version: "1.0"
generator: tinct v0.0.1
generated: 2024-01-15T10:30:00Z
colors:
  - hex: "#1a1b26"
    weight: 0.25
    role: background
  - hex: "#c0caf5"
    weight: 0.15
    role: foreground
---

# Test Theme

This is the markdown body.
`,
			want: &Theme{
				Name:      "Test Theme",
				Version:   "1.0",
				Generator: "tinct v0.0.1",
				Colors: []Color{
					{Hex: "#1a1b26", Weight: 0.25, Role: "background"},
					{Hex: "#c0caf5", Weight: 0.15, Role: "foreground"},
				},
			},
			wantErr: false,
		},
		{
			name: "theme with external wallpaper",
			input: `---
name: External Wallpaper Theme
version: "1.0"
generator: tinct
generated: 2024-01-15T10:30:00Z
wallpaper:
  embedded: false
  path: ./wallpapers/background.png
colors:
  - hex: "#ffffff"
    weight: 1.0
---
`,
			want: &Theme{
				Name:      "External Wallpaper Theme",
				Version:   "1.0",
				Generator: "tinct",
				Wallpaper: &Wallpaper{
					Embedded: false,
					Path:     "./wallpapers/background.png",
				},
				Colors: []Color{
					{Hex: "#ffffff", Weight: 1.0},
				},
			},
			wantErr: false,
		},
		{
			name: "theme with embedded wallpaper",
			input: `---
name: Embedded Theme
version: "1.0"
generator: tinct
generated: 2024-01-15T10:30:00Z
wallpaper:
  embedded: true
  format: png
  width: 100
  height: 100
  data: iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==
colors:
  - hex: "#000000"
    weight: 1.0
---
`,
			want: &Theme{
				Name:      "Embedded Theme",
				Version:   "1.0",
				Generator: "tinct",
				Wallpaper: &Wallpaper{
					Embedded: true,
					Format:   "png",
					Width:    100,
					Height:   100,
					Data:     "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
				},
				Colors: []Color{
					{Hex: "#000000", Weight: 1.0},
				},
			},
			wantErr: false,
		},
		{
			name:    "missing front matter delimiter",
			input:   "name: No Delimiters\n",
			wantErr: true,
		},
		{
			name: "unclosed front matter",
			input: `---
name: Unclosed
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Check basic fields
			if got.Name != tt.want.Name {
				t.Errorf("Parse() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Version != tt.want.Version {
				t.Errorf("Parse() Version = %v, want %v", got.Version, tt.want.Version)
			}
			if got.Generator != tt.want.Generator {
				t.Errorf("Parse() Generator = %v, want %v", got.Generator, tt.want.Generator)
			}

			// Check colors
			if len(got.Colors) != len(tt.want.Colors) {
				t.Errorf("Parse() Colors length = %v, want %v", len(got.Colors), len(tt.want.Colors))
			}
			for i := range tt.want.Colors {
				if i >= len(got.Colors) {
					break
				}
				if got.Colors[i].Hex != tt.want.Colors[i].Hex {
					t.Errorf("Parse() Colors[%d].Hex = %v, want %v", i, got.Colors[i].Hex, tt.want.Colors[i].Hex)
				}
				if got.Colors[i].Role != tt.want.Colors[i].Role {
					t.Errorf("Parse() Colors[%d].Role = %v, want %v", i, got.Colors[i].Role, tt.want.Colors[i].Role)
				}
			}

			// Check wallpaper
			if tt.want.Wallpaper != nil {
				if got.Wallpaper == nil {
					t.Errorf("Parse() Wallpaper = nil, want non-nil")
				} else {
					if got.Wallpaper.Embedded != tt.want.Wallpaper.Embedded {
						t.Errorf("Parse() Wallpaper.Embedded = %v, want %v", got.Wallpaper.Embedded, tt.want.Wallpaper.Embedded)
					}
					if got.Wallpaper.Path != tt.want.Wallpaper.Path {
						t.Errorf("Parse() Wallpaper.Path = %v, want %v", got.Wallpaper.Path, tt.want.Wallpaper.Path)
					}
					if got.Wallpaper.Format != tt.want.Wallpaper.Format {
						t.Errorf("Parse() Wallpaper.Format = %v, want %v", got.Wallpaper.Format, tt.want.Wallpaper.Format)
					}
				}
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	theme := &Theme{
		Name:      "Generated Theme",
		Version:   "1.0",
		Generator: "tinct-test",
		Generated: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Colors: []Color{
			{Hex: "#1a1b26", Weight: 0.25, Role: "background"},
			{Hex: "#c0caf5", Weight: 0.15, Role: "foreground"},
		},
	}

	content, err := Generate(theme, false, "")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that required elements are present
	if !strings.Contains(content, "---") {
		t.Error("Generate() missing front matter delimiters")
	}
	if !strings.Contains(content, "name: Generated Theme") {
		t.Error("Generate() missing theme name")
	}
	if !strings.Contains(content, "# Generated Theme") {
		t.Error("Generate() missing markdown title")
	}
	if !strings.Contains(content, "## Color Palette") {
		t.Error("Generate() missing color palette section")
	}
	if !strings.Contains(content, "#1a1b26") {
		t.Error("Generate() missing color hex value")
	}
}

func TestGenerateWithThumbnail(t *testing.T) {
	theme := &Theme{
		Name:      "Thumbnail Theme",
		Version:   "1.0",
		Generator: "tinct-test",
		Generated: time.Now(),
		Colors: []Color{
			{Hex: "#ffffff", Weight: 1.0},
		},
	}

	// Test thumbnail data (1x1 pixel PNG in base64)
	thumbnailData := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

	content, err := Generate(theme, true, thumbnailData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that thumbnail section is present
	if !strings.Contains(content, "## Preview") {
		t.Error("Generate() missing preview section with thumbnail")
	}
	if !strings.Contains(content, "data:image/jpeg;base64,") {
		t.Error("Generate() missing thumbnail data URL")
	}
}

func TestExternalWallpaper(t *testing.T) {
	wp := ExternalWallpaper("/path/to/wallpaper.png")

	if wp.Embedded {
		t.Error("ExternalWallpaper() Embedded = true, want false")
	}
	if wp.Path != "/path/to/wallpaper.png" {
		t.Errorf("ExternalWallpaper() Path = %v, want %v", wp.Path, "/path/to/wallpaper.png")
	}
}

func TestEncodeWallpaper(t *testing.T) {
	// Minimal valid PNG (1x1 pixel, red)
	pngData := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 dimensions
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, // bit depth, color type, CRC
		0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, 0x54, // IDAT chunk
		0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00, 0x00,
		0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xfe, 0xd4, // compressed data
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, // IEND chunk
		0xae, 0x42, 0x60, 0x82,
	}

	wp, err := EncodeWallpaper(pngData, "png")
	if err != nil {
		t.Fatalf("EncodeWallpaper() error = %v", err)
	}

	if !wp.Embedded {
		t.Error("EncodeWallpaper() Embedded = false, want true")
	}
	if wp.Format != "png" {
		t.Errorf("EncodeWallpaper() Format = %v, want png", wp.Format)
	}
	if wp.Width != 1 {
		t.Errorf("EncodeWallpaper() Width = %v, want 1", wp.Width)
	}
	if wp.Height != 1 {
		t.Errorf("EncodeWallpaper() Height = %v, want 1", wp.Height)
	}
	if wp.Data == "" {
		t.Error("EncodeWallpaper() Data is empty")
	}
}

func TestDecodeWallpaper_Embedded(t *testing.T) {
	// Base64-encoded minimal PNG (1x1 red pixel)
	base64Data := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAIAAACQd1PeAAAADklEQVR4nGNgYGBgYGAAABAAA0AAAE2lAAAAAElFTkSuQmCC"

	theme := &Theme{
		Name: "Test",
		Wallpaper: &Wallpaper{
			Embedded: true,
			Format:   "png",
			Data:     base64Data,
		},
	}

	data, format, err := DecodeWallpaper(theme, "")
	if err != nil {
		t.Fatalf("DecodeWallpaper() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("DecodeWallpaper() returned empty data")
	}
	if format != "png" {
		t.Errorf("DecodeWallpaper() format = %v, want png", format)
	}

	// Verify it starts with PNG signature
	if len(data) < 8 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Error("DecodeWallpaper() data is not valid PNG")
	}
}

func TestDecodeWallpaper_NoWallpaper(t *testing.T) {
	theme := &Theme{
		Name:      "Test",
		Wallpaper: nil,
	}

	_, _, err := DecodeWallpaper(theme, "")
	if err == nil {
		t.Error("DecodeWallpaper() expected error for nil wallpaper")
	}
}

func TestDecodeWallpaper_NoData(t *testing.T) {
	theme := &Theme{
		Name: "Test",
		Wallpaper: &Wallpaper{
			Embedded: false,
			// No path, URL, or data
		},
	}

	_, _, err := DecodeWallpaper(theme, "")
	if err == nil {
		t.Error("DecodeWallpaper() expected error for wallpaper with no data")
	}
}

func TestDetectFormatFromContentType(t *testing.T) {
	tests := []struct {
		contentType string
		want        string
	}{
		{"image/png", "png"},
		{"image/jpeg", "jpg"},
		{"image/webp", "webp"},
		{"image/gif", "gif"},
		{"text/html", ""},
		{"", ""},
		{"image/png; charset=utf-8", "png"},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			got := detectFormatFromContentType(tt.contentType)
			if got != tt.want {
				t.Errorf("detectFormatFromContentType(%q) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}
