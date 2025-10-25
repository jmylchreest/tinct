package fuzzel

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestFuzzelPlugin runs all standard plugin tests using shared utilities.
func TestFuzzelPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:       "fuzzel",
		ExpectedFiles:      []string{"tinct.ini"},
		ExpectedBinaryName: "fuzzel",
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestFuzzelPlugin_RGBAConversion tests fuzzel-specific RGBA conversion.
func TestFuzzelPlugin_RGBAConversion(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		alpha    string
		expected string
	}{
		{
			name:     "Standard hex with hash",
			hex:      "#1a1b26",
			alpha:    "ff",
			expected: "1a1b26ff",
		},
		{
			name:     "Hex without hash",
			hex:      "c0caf5",
			alpha:    "ee",
			expected: "c0caf5ee",
		},
		{
			name:     "Transparent background",
			hex:      "#1a1b26",
			alpha:    "ee",
			expected: "1a1b26ee",
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

// TestFuzzelPlugin_ContentValidation tests fuzzel-specific content requirements.
func TestFuzzelPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	files, err := plugin.Generate(palette)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.ini"])

	// Fuzzel-specific validations
	requiredStrings := []string{
		"[colors]",
		"background=",
		"text=",
		"prompt=",
		"placeholder=",
		"input=",
		"match=",
		"selection=",
		"selection-text=",
		"selection-match=",
		"counter=",
		"border=",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(content, required) {
			t.Errorf("Generated content missing required string: %s", required)
		}
	}

	// Check that theme type is present
	if !strings.Contains(content, "Detected theme: dark") {
		t.Error("Generated content missing theme type")
	}
}

// TestFuzzelPlugin_ThemeDataMethods tests all RGBA helper methods.
func TestFuzzelPlugin_ThemeDataMethods(t *testing.T) {
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
		name   string
		method func(string) string
		alpha  string
	}{
		{"BackgroundRGBA", data.BackgroundRGBA, "ee"},
		{"BackgroundMutedRGBA", data.BackgroundMutedRGBA, "ff"},
		{"ForegroundRGBA", data.ForegroundRGBA, "ff"},
		{"ForegroundMutedRGBA", data.ForegroundMutedRGBA, "cc"},
		{"Accent1RGBA", data.Accent1RGBA, "ff"},
		{"Accent2RGBA", data.Accent2RGBA, "ff"},
		{"Accent3RGBA", data.Accent3RGBA, "ff"},
		{"Accent4RGBA", data.Accent4RGBA, "ff"},
		{"DangerRGBA", data.DangerRGBA, "ff"},
		{"WarningRGBA", data.WarningRGBA, "ff"},
		{"SuccessRGBA", data.SuccessRGBA, "ff"},
		{"InfoRGBA", data.InfoRGBA, "ff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(tt.alpha)
			// Should be 8 characters (RRGGBBAA)
			if len(result) != 8 {
				t.Errorf("%s(%s) = %s (length %d), want length 8", tt.name, tt.alpha, result, len(result))
			}
			// Should end with the alpha value
			if !strings.HasSuffix(result, tt.alpha) {
				t.Errorf("%s(%s) = %s, should end with %s", tt.name, tt.alpha, result, tt.alpha)
			}
		})
	}
}

// TestFuzzelPlugin_GetEmbeddedTemplates tests embedded template access.
func TestFuzzelPlugin_GetEmbeddedTemplates(t *testing.T) {
	fs := GetEmbeddedTemplates()

	entries, err := fs.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Name() == "tinct.ini.tmpl" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Template file tinct.ini.tmpl not found in embedded filesystem")
	}
}

// TestFuzzelPlugin_HexToRGBAEdgeCases tests edge cases in hex conversion.
func TestFuzzelPlugin_HexToRGBAEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		alpha    string
		expected string
	}{
		{
			name:     "Already 6 chars no hash",
			hex:      "ffffff",
			alpha:    "ff",
			expected: "ffffffff",
		},
		{
			name:     "Lowercase hex",
			hex:      "#abcdef",
			alpha:    "80",
			expected: "abcdef80",
		},
		{
			name:     "Short hex (invalid but handled)",
			hex:      "fff",
			alpha:    "ff",
			expected: "fffff",
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
