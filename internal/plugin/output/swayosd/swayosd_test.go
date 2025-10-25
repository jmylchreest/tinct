package swayosd

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestSwayOSDPlugin runs all standard plugin tests using shared utilities.
func TestSwayOSDPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "swayosd",
		ExpectedFiles:      []string{"style.css"},
		ExpectedBinaryName: "swayosd",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestSwayOSDPlugin_ContentValidation tests swayosd-specific content requirements.
func TestSwayOSDPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(palette)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["style.css"])

	requiredStrings := []string{
		"/* SwayOSD colour theme",
		"window#osd",
		"background:",
		"color:",
		"rgb(",
		"rgba(",
		"progressbar",
		"trough",
		"progress",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(content, required) {
			t.Errorf("Generated content missing required string: %s", required)
		}
	}

	if !strings.Contains(content, "Detected theme: dark") {
		t.Error("Generated content missing theme type")
	}
}

// TestSwayOSDPlugin_GenerateWithLightTheme tests light theme generation.
func TestSwayOSDPlugin_GenerateWithLightTheme(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeLight)
	plugin := New()

	files, err := plugin.Generate(palette)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["style.css"])

	if !strings.Contains(content, "Detected theme: light") {
		t.Error("Generated content missing light theme type")
	}
}

// TestSwayOSDPlugin_CustomOutputDir tests custom output directory handling.
func TestSwayOSDPlugin_CustomOutputDir(t *testing.T) {
	plugin := New()
	plugin.outputDir = "/custom/path"

	dir := plugin.DefaultOutputDir()
	if dir != "/custom/path" {
		t.Errorf("DefaultOutputDir() = %s, want /custom/path", dir)
	}
}

// TestSwayOSDPlugin_HexToRGB tests the hexToRGB helper function.
func TestSwayOSDPlugin_HexToRGB(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		expected string
	}{
		{
			name:     "Standard hex with hash",
			hex:      "#1a1b26",
			expected: "rgb(26, 27, 38)",
		},
		{
			name:     "Hex without hash",
			hex:      "c0caf5",
			expected: "rgb(192, 202, 245)",
		},
		{
			name:     "Uppercase hex",
			hex:      "#FF6550",
			expected: "rgb(255, 101, 80)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hexToRGB(tt.hex)
			if result != tt.expected {
				t.Errorf("hexToRGB(%s) = %s, want %s", tt.hex, result, tt.expected)
			}
		})
	}
}

// TestSwayOSDPlugin_HexToRGBA tests the hexToRGBA helper function.
func TestSwayOSDPlugin_HexToRGBA(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		alpha    string
		expected string
	}{
		{
			name:     "Full opacity",
			hex:      "#1a1b26",
			alpha:    "1.0",
			expected: "rgba(26, 27, 38, 1.00)",
		},
		{
			name:     "Semi-transparent",
			hex:      "#c0caf5",
			alpha:    "0.85",
			expected: "rgba(192, 202, 245, 0.85)",
		},
		{
			name:     "Low opacity",
			hex:      "#7aa2f7",
			alpha:    "0.3",
			expected: "rgba(122, 162, 247, 0.30)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hexToRGBA(tt.hex, tt.alpha)
			if result != tt.expected {
				t.Errorf("hexToRGBA(%s, %s) = %s, want %s", tt.hex, tt.alpha, result, tt.expected)
			}
		})
	}
}

// TestSwayOSDPlugin_ThemeDataRGBMethods tests RGB helper methods.
func TestSwayOSDPlugin_ThemeDataRGBMethods(t *testing.T) {
	data := ThemeData{
		Background:      "#1a1b26",
		BackgroundMuted: "#16161e",
		Foreground:      "#c0caf5",
		ForegroundMuted: "#a9b1d6",
		Accent1:         "#7aa2f7",
		Accent2:         "#bb9af7",
		Accent3:         "#7dcfff",
		Accent4:         "#9ece6a",
		Danger:          "#f7768e",
		Warning:         "#e0af68",
		Success:         "#9ece6a",
		Info:            "#7aa2f7",
	}

	tests := []struct {
		name     string
		method   func() string
		contains string
	}{
		{"BackgroundRGB", data.BackgroundRGB, "rgb(26, 27, 38)"},
		{"ForegroundRGB", data.ForegroundRGB, "rgb(192, 202, 245)"},
		{"Accent1RGB", data.Accent1RGB, "rgb(122, 162, 247)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			if result != tt.contains {
				t.Errorf("%s() = %s, want %s", tt.name, result, tt.contains)
			}
		})
	}
}

// TestSwayOSDPlugin_ThemeDataRGBAMethods tests RGBA helper methods.
func TestSwayOSDPlugin_ThemeDataRGBAMethods(t *testing.T) {
	data := ThemeData{
		Background: "#1a1b26",
		Foreground: "#c0caf5",
		Accent1:    "#7aa2f7",
	}

	tests := []struct {
		name   string
		method func(string) string
		alpha  string
	}{
		{"BackgroundRGBA", data.BackgroundRGBA, "0.85"},
		{"ForegroundRGBA", data.ForegroundRGBA, "1.0"},
		{"Accent1RGBA", data.Accent1RGBA, "0.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.alpha)
			if !strings.Contains(result, "rgba") {
				t.Errorf("%s(%s) = %s, should contain 'rgba'", tt.name, tt.alpha, result)
			}
			if !strings.Contains(result, tt.alpha) {
				t.Errorf("%s(%s) = %s, should contain alpha %s", tt.name, tt.alpha, result, tt.alpha)
			}
		})
	}
}

// TestSwayOSDPlugin_PrepareThemeData tests theme data preparation.
func TestSwayOSDPlugin_PrepareThemeData(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	data := plugin.prepareThemeData(palette)

	if data.SourceTheme == "" {
		t.Error("SourceTheme should not be empty")
	}
	if data.Background == "" {
		t.Error("Background should not be empty")
	}
	if data.Foreground == "" {
		t.Error("Foreground should not be empty")
	}
	if data.SourceTheme != "dark" {
		t.Errorf("SourceTheme = %s, want dark", data.SourceTheme)
	}
}

// TestSwayOSDPlugin_GetEmbeddedTemplates tests embedded template access.
func TestSwayOSDPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Name() == "style.css.tmpl" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Template file style.css.tmpl not found in embedded filesystem")
	}
}
