// Package common provides shared utilities for output plugins.
package common

import (
	"bytes"
	"image/color"
	"strings"
	"testing"
	"text/template"

	"github.com/jmylchreest/tinct/internal/colour"
)

// createTestPalette creates a test palette with standard colors for testing.
func createTestPalette() *colour.CategorisedPalette {
	colors := []color.Color{
		color.RGBA{R: 26, G: 27, B: 38, A: 255},    // Dark background
		color.RGBA{R: 192, G: 202, B: 245, A: 255}, // Light foreground
		color.RGBA{R: 122, G: 162, B: 247, A: 255}, // Blue accent1
		color.RGBA{R: 187, G: 154, B: 247, A: 255}, // Purple accent2
		color.RGBA{R: 125, G: 207, B: 255, A: 255}, // Cyan accent3
		color.RGBA{R: 158, G: 206, B: 106, A: 255}, // Green accent4/success
		color.RGBA{R: 247, G: 118, B: 142, A: 255}, // Red/danger
		color.RGBA{R: 224, G: 175, B: 104, A: 255}, // Orange/warning
	}

	palette := &colour.Palette{Colors: colors}
	config := colour.DefaultCategorisationConfig()
	config.ThemeType = colour.ThemeDark

	return colour.Categorise(palette, config)
}

// TestTemplateFuncs_Get tests the get function.
func TestTemplateFuncs_Get(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ get . "background" | hex }}`)
	if err != nil {
		t.Fatalf("Template parse error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		t.Fatalf("Template execute error: %v", err)
	}

	result := buf.String()
	if !strings.HasPrefix(result, "#") {
		t.Errorf("Expected hex color, got: %s", result)
	}
}

// TestTemplateFuncs_GetSafe tests the getSafe function.
func TestTemplateFuncs_GetSafe(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	t.Run("ValidRole", func(t *testing.T) {
		tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ getSafe . "background" | hex }}`)
		if err != nil {
			t.Fatalf("Template parse error: %v", err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, themeData); err != nil {
			t.Fatalf("Template execute error: %v", err)
		}

		result := buf.String()
		if !strings.HasPrefix(result, "#") {
			t.Errorf("Expected hex color, got: %s", result)
		}
	})

	t.Run("InvalidRole", func(t *testing.T) {
		tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ getSafe . "nonexistent" }}`)
		if err != nil {
			t.Fatalf("Template parse error: %v", err)
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, themeData)
		if err == nil {
			t.Error("Expected error for nonexistent role")
		}
	})
}

// TestTemplateFuncs_Has tests the has function.
func TestTemplateFuncs_Has(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tests := []struct {
		role     string
		expected bool
	}{
		{"background", true},
		{"foreground", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ has . "` + tt.role + `" }}`)
			if err != nil {
				t.Fatalf("Template parse error: %v", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, themeData); err != nil {
				t.Fatalf("Template execute error: %v", err)
			}

			result := buf.String()
			expectedStr := "false"
			if tt.expected {
				expectedStr = "true"
			}
			if result != expectedStr {
				t.Errorf("Expected %s, got: %s", expectedStr, result)
			}
		})
	}
}

// TestTemplateFuncs_GetByIndex tests the getByIndex function.
func TestTemplateFuncs_GetByIndex(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	t.Run("ValidIndex", func(t *testing.T) {
		tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ getByIndex . 0 | hex }}`)
		if err != nil {
			t.Fatalf("Template parse error: %v", err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, themeData); err != nil {
			t.Fatalf("Template execute error: %v", err)
		}

		result := buf.String()
		if !strings.HasPrefix(result, "#") {
			t.Errorf("Expected hex color, got: %s", result)
		}
	})

	t.Run("InvalidIndex", func(t *testing.T) {
		tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ getByIndex . 999 }}`)
		if err != nil {
			t.Fatalf("Template parse error: %v", err)
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, themeData)
		if err == nil {
			t.Error("Expected error for out of range index")
		}
	})
}

// TestTemplateFuncs_HexFormats tests hex color format functions.
func TestTemplateFuncs_HexFormats(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tests := []struct {
		name     string
		template string
		check    func(string) bool
	}{
		{
			name:     "Hex",
			template: `{{ get . "background" | hex }}`,
			check:    func(s string) bool { return strings.HasPrefix(s, "#") && len(s) == 7 },
		},
		{
			name:     "HexAlpha",
			template: `{{ get . "background" | hexAlpha }}`,
			check:    func(s string) bool { return strings.HasPrefix(s, "#") && len(s) == 9 },
		},
		{
			name:     "HexNoHash",
			template: `{{ get . "background" | hexNoHash }}`,
			check:    func(s string) bool { return !strings.HasPrefix(s, "#") && len(s) == 6 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(tt.template)
			if err != nil {
				t.Fatalf("Template parse error: %v", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, themeData); err != nil {
				t.Fatalf("Template execute error: %v", err)
			}

			result := buf.String()
			if !tt.check(result) {
				t.Errorf("Format check failed for %s: %s", tt.name, result)
			}
		})
	}
}

// TestTemplateFuncs_RGBFormats tests RGB color format functions.
func TestTemplateFuncs_RGBFormats(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tests := []struct {
		name     string
		template string
		check    func(string) bool
	}{
		{
			name:     "RGB",
			template: `{{ get . "background" | rgb }}`,
			check:    func(s string) bool { return strings.HasPrefix(s, "rgb(") },
		},
		{
			name:     "RGBA",
			template: `{{ get . "background" | rgba }}`,
			check:    func(s string) bool { return strings.HasPrefix(s, "rgba(") },
		},
		{
			name:     "RGBDecimal",
			template: `{{ get . "background" | rgbDecimal }}`,
			check:    func(s string) bool { return strings.Contains(s, ",") && !strings.Contains(s, "(") },
		},
		{
			name:     "RGBADecimal",
			template: `{{ get . "background" | rgbaDecimal }}`,
			check:    func(s string) bool { return strings.Contains(s, ",") && !strings.Contains(s, "(") },
		},
		{
			name:     "RGBSpaces",
			template: `{{ get . "background" | rgbSpaces }}`,
			check:    func(s string) bool { return strings.Contains(s, " ") && !strings.Contains(s, ",") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(tt.template)
			if err != nil {
				t.Fatalf("Template parse error: %v", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, themeData); err != nil {
				t.Fatalf("Template execute error: %v", err)
			}

			result := buf.String()
			if !tt.check(result) {
				t.Errorf("Format check failed for %s: %s", tt.name, result)
			}
		})
	}
}

// TestTemplateFuncs_WithAlpha tests the withAlpha function.
func TestTemplateFuncs_WithAlpha(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	// Use withAlpha with explicit arguments instead of piping
	tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ withAlpha (get . "background") 0.5 | hexAlpha }}`)
	if err != nil {
		t.Fatalf("Template parse error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		t.Fatalf("Template execute error: %v", err)
	}

	result := buf.String()
	// Alpha 0.5 should produce hex value around 7f or 80 in the alpha channel
	if !strings.HasPrefix(result, "#") || len(result) != 9 {
		t.Errorf("Expected 9-char hex with alpha, got: %s", result)
	}
}

// TestTemplateFuncs_Metadata tests color metadata functions.
func TestTemplateFuncs_Metadata(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "Role",
			template: `{{ get . "background" | role }}`,
		},
		{
			name:     "Index",
			template: `{{ get . "background" | index }}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(tt.template)
			if err != nil {
				t.Fatalf("Template parse error: %v", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, themeData); err != nil {
				t.Fatalf("Template execute error: %v", err)
			}

			result := buf.String()
			if result == "" {
				t.Error("Expected non-empty metadata")
			}
		})
	}
}

// TestTemplateFuncs_PaletteMetadata tests palette metadata functions.
func TestTemplateFuncs_PaletteMetadata(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tests := []struct {
		name     string
		template string
		check    func(string) bool
	}{
		{
			name:     "ThemeType",
			template: `{{ themeType . }}`,
			check:    func(s string) bool { return s == "dark" || s == "light" },
		},
		{
			name:     "Count",
			template: `{{ count . }}`,
			check:    func(s string) bool { return s != "0" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(tt.template)
			if err != nil {
				t.Fatalf("Template parse error: %v", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, themeData); err != nil {
				t.Fatalf("Template execute error: %v", err)
			}

			result := buf.String()
			if !tt.check(result) {
				t.Errorf("Check failed for %s: %s", tt.name, result)
			}
		})
	}
}

// TestTemplateFuncs_StringManipulation tests string manipulation functions.
func TestTemplateFuncs_StringManipulation(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tests := []struct {
		name     string
		template string
		checkLen int
	}{
		{
			name:     "TrimPrefix",
			template: `{{ get . "background" | hex | trimPrefix "#" }}`,
			checkLen: 6, // Should be 6 chars without #
		},
		{
			name:     "ToLower",
			template: `{{ "HELLO" | toLower }}`,
			checkLen: 5,
		},
		{
			name:     "ToUpper",
			template: `{{ "hello" | toUpper }}`,
			checkLen: 5,
		},
		{
			name:     "Replace",
			template: `{{ "a_b_c" | replace "_" "-" }}`,
			checkLen: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(tt.template)
			if err != nil {
				t.Fatalf("Template parse error: %v", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, themeData); err != nil {
				t.Fatalf("Template execute error: %v", err)
			}

			result := buf.String()
			if len(result) != tt.checkLen {
				t.Errorf("Expected length %d, got %d (result: %s)", tt.checkLen, len(result), result)
			}
		})
	}
}

// TestTemplateFuncs_ANSIColors tests ANSI color functions.
func TestTemplateFuncs_ANSIColors(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	t.Run("ValidANSIColor", func(t *testing.T) {
		tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ ansi . "red" | hex }}`)
		if err != nil {
			t.Fatalf("Template parse error: %v", err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, themeData); err != nil {
			t.Fatalf("Template execute error: %v", err)
		}

		result := buf.String()
		if !strings.HasPrefix(result, "#") {
			t.Errorf("Expected hex color, got: %s", result)
		}
	})

	t.Run("InvalidANSIColor", func(t *testing.T) {
		tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ ansiSafe . "invalidcolor" }}`)
		if err != nil {
			t.Fatalf("Template parse error: %v", err)
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, themeData)
		if err == nil {
			t.Error("Expected error for invalid ANSI color")
		}
	})
}

// TestTemplateFuncs_AllColorsIteration tests iterating over all colors.
func TestTemplateFuncs_AllColorsIteration(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ range allColors . }}{{ . | hex }}{{ end }}`)
	if err != nil {
		t.Fatalf("Template parse error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		t.Fatalf("Template execute error: %v", err)
	}

	result := buf.String()
	if result == "" {
		t.Error("Expected non-empty result from color iteration")
	}

	// Should contain multiple hex colors
	hexCount := strings.Count(result, "#")
	if hexCount < 2 {
		t.Errorf("Expected multiple hex colors, got %d", hexCount)
	}
}

// TestTemplateFuncs_AllRolesIteration tests iterating over all roles.
func TestTemplateFuncs_AllRolesIteration(t *testing.T) {
	palette := createTestPalette()
	themeData := colour.NewThemeData(palette, "", "")

	tmpl, err := template.New("test").Funcs(TemplateFuncs()).Parse(`{{ range allRoles . }}{{ . }}{{ end }}`)
	if err != nil {
		t.Fatalf("Template parse error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		t.Fatalf("Template execute error: %v", err)
	}

	result := buf.String()
	if result == "" {
		t.Error("Expected non-empty result from role iteration")
	}
}
