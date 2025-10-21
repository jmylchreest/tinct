// Package tailwind provides a Tailwind CSS / shadcn/ui output plugin.
package tailwind

import (
	"image/color"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
)

func TestTailwindPlugin_Name(t *testing.T) {
	plugin := New()
	if plugin.Name() != "tailwind" {
		t.Errorf("Name() = %s, want tailwind", plugin.Name())
	}
}

func TestTailwindPlugin_Description(t *testing.T) {
	plugin := New()
	desc := plugin.Description()
	if desc == "" {
		t.Error("Description() should not be empty")
	}
}

func TestTailwindPlugin_DefaultOutputDir(t *testing.T) {
	plugin := New()
	dir := plugin.DefaultOutputDir()
	if dir == "" {
		t.Error("DefaultOutputDir() should not be empty")
	}
}

func TestTailwindPlugin_Validate(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "Valid CSS format",
			format:  "css",
			wantErr: false,
		},
		{
			name:    "Valid config format",
			format:  "config",
			wantErr: false,
		},
		{
			name:    "Invalid format",
			format:  "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewWithFormat(tt.format)
			err := plugin.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTailwindPlugin_GenerateCSS(t *testing.T) {
	// Create a test palette
	palette := createTestPalette(colour.ThemeDark)

	plugin := New() // Default is CSS format
	files, err := plugin.Generate(palette)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	var outputStr string
	for _, content := range files {
		outputStr = string(content)
		break
	}

	// Check for essential CSS structure
	requiredStrings := []string{
		"@layer base",
		":root",
		".dark",
		"--background:",
		"--foreground:",
		"--primary:",
		"--secondary:",
		"--muted:",
		"--accent:",
		"--destructive:",
		"--border:",
		"--input:",
		"--ring:",
		"--card:",
		"--popover:",
		"--radius:",
		"--warning:",
		"--success:",
		"--info:",
	}

	for _, req := range requiredStrings {
		if !strings.Contains(outputStr, req) {
			t.Errorf("Generated CSS missing required string: %s", req)
		}
	}
}

func TestTailwindPlugin_GenerateConfig(t *testing.T) {
	// Create a test palette
	palette := createTestPalette(colour.ThemeDark)

	plugin := NewWithFormat("config")
	files, err := plugin.Generate(palette)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	var outputStr string
	for _, content := range files {
		outputStr = string(content)
		break
	}

	// Check for essential config structure
	requiredStrings := []string{
		"module.exports",
		"theme:",
		"extend:",
		"colors:",
		"destructive:",
		"warning:",
		"success:",
		"info:",
		"accent:",
	}

	for _, req := range requiredStrings {
		if !strings.Contains(outputStr, req) {
			t.Errorf("Generated config missing required string: %s", req)
		}
	}
}

func TestTailwindPlugin_GenerateNilPalette(t *testing.T) {
	plugin := New()
	_, err := plugin.Generate(nil)
	if err == nil {
		t.Error("Generate() with nil palette should return error")
	}
}

func TestToHSL(t *testing.T) {
	tests := []struct {
		name string
		cc   colour.CategorisedColour
		want string
	}{
		{
			name: "Red colour",
			cc: colour.CategorisedColour{
				Hue:        0,
				Saturation: 0.75,
				Luminance:  0.22,
			},
			want: "0.0 75.0% 22.0%",
		},
		{
			name: "Blue colour",
			cc: colour.CategorisedColour{
				Hue:        210,
				Saturation: 0.847,
				Luminance:  0.049,
			},
			want: "210.0 84.7% 4.9%",
		},
		{
			name: "Green colour",
			cc: colour.CategorisedColour{
				Hue:        120,
				Saturation: 1.0,
				Luminance:  0.5,
			},
			want: "120.0 100.0% 50.0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toHSL(tt.cc)
			if got != tt.want {
				t.Errorf("toHSL() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestBuildCSSTheme(t *testing.T) {
	// Create test colours
	bg := colour.CategorisedColour{
		Colour:     color.RGBA{R: 30, G: 30, B: 40, A: 255},
		Hue:        240,
		Saturation: 0.14,
		Luminance:  0.01,
	}

	fg := colour.CategorisedColour{
		Colour:     color.RGBA{R: 230, G: 230, B: 240, A: 255},
		Hue:        240,
		Saturation: 0.17,
		Luminance:  0.87,
	}

	primary := colour.CategorisedColour{
		Colour:     color.RGBA{R: 229, G: 76, B: 76, A: 255},
		Hue:        0,
		Saturation: 0.75,
		Luminance:  0.22,
	}

	theme := buildCSSTheme(bg, fg, bg, fg, primary, primary, primary, primary,
		true, true, true, true, true, true, true, true, true)

	// Check that all required fields are populated
	if theme.Background == "" {
		t.Error("Background should not be empty")
	}
	if theme.Foreground == "" {
		t.Error("Foreground should not be empty")
	}
	if theme.Primary == "" {
		t.Error("Primary should not be empty")
	}
	if theme.Card == "" {
		t.Error("Card should not be empty")
	}
}

func TestPrepareCSSData(t *testing.T) {
	tests := []struct {
		name      string
		palette   *colour.CategorisedPalette
		wantTheme string
	}{
		{
			name:      "Dark theme palette",
			palette:   createTestPalette(colour.ThemeDark),
			wantTheme: "dark",
		},
		{
			name:      "Light theme palette",
			palette:   createTestPalette(colour.ThemeLight),
			wantTheme: "light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := prepareCSSData(tt.palette)

			if data.ThemeType != tt.wantTheme {
				t.Errorf("ThemeType = %s, want %s", data.ThemeType, tt.wantTheme)
			}

			// Check that custom colours are populated
			if !data.HasCustom {
				t.Error("HasCustom should be true")
			}

			if len(data.Custom) == 0 {
				t.Error("Custom colours should not be empty")
			}

			// Check for semantic colours
			hasWarning := false
			hasSuccess := false
			hasInfo := false
			for _, cc := range data.Custom {
				switch cc.Name {
				case "warning":
					hasWarning = true
				case "success":
					hasSuccess = true
				case "info":
					hasInfo = true
				}
			}

			if !hasWarning {
				t.Error("Custom colours should include warning")
			}
			if !hasSuccess {
				t.Error("Custom colours should include success")
			}
			if !hasInfo {
				t.Error("Custom colours should include info")
			}
		})
	}
}

func TestPrepareConfigData(t *testing.T) {
	palette := createTestPalette(colour.ThemeDark)
	data := prepareConfigData(palette)

	// Check that semantic colours are present
	if _, ok := data.Colors["destructive"]; !ok {
		t.Error("Config should include destructive colour")
	}
	if _, ok := data.Colors["warning"]; !ok {
		t.Error("Config should include warning colour")
	}
	if _, ok := data.Colors["success"]; !ok {
		t.Error("Config should include success colour")
	}
	if _, ok := data.Colors["info"]; !ok {
		t.Error("Config should include info colour")
	}

	// Check that accent colours are present
	accents, ok := data.Colors["accent"]
	if !ok {
		t.Error("Config should include accent colours")
	}

	// Check that accents is a map
	if accentsMap, ok := accents.(map[string]string); ok {
		if len(accentsMap) == 0 {
			t.Error("Accent colours map should not be empty")
		}
	} else {
		t.Error("Accent colours should be a map")
	}
}

// Helper function to create a test palette
func createTestPalette(themeType colour.ThemeType) *colour.CategorisedPalette {
	colors := []color.Color{
		color.RGBA{R: 30, G: 30, B: 40, A: 255},    // Dark background
		color.RGBA{R: 230, G: 230, B: 240, A: 255}, // Light foreground
		color.RGBA{R: 229, G: 76, B: 76, A: 255},   // Red
		color.RGBA{R: 229, G: 191, B: 76, A: 255},  // Orange
		color.RGBA{R: 76, G: 229, B: 76, A: 255},   // Green
		color.RGBA{R: 76, G: 153, B: 229, A: 255},  // Blue
		color.RGBA{R: 191, G: 76, B: 229, A: 255},  // Purple
	}

	palette := &colour.Palette{Colors: colors}
	config := colour.DefaultCategorisationConfig()
	config.ThemeType = themeType

	return colour.Categorise(palette, config)
}
