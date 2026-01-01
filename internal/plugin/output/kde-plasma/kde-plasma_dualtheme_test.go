package kdeplasma

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	pkgcolour "github.com/jmylchreest/tinct/pkg/colour"
)

func TestGenerateDualTheme(t *testing.T) {
	// Create a simple dark palette
	darkPalette := &colour.CategorisedPalette{
		Colours: map[pkgcolour.Role]pkgcolour.CategorisedColour{
			pkgcolour.RoleBackground: {Hex: "#1e1e2e", RGBA: pkgcolour.RGBA{R: 30, G: 30, B: 46, A: 255}},
			pkgcolour.RoleForeground: {Hex: "#cdd6f4", RGBA: pkgcolour.RGBA{R: 205, G: 214, B: 244, A: 255}},
			pkgcolour.RoleAccent1:    {Hex: "#89b4fa", RGBA: pkgcolour.RGBA{R: 137, G: 180, B: 250, A: 255}},
		},
		ThemeType: pkgcolour.ThemeDark,
	}

	// Create a simple light palette
	lightPalette := &colour.CategorisedPalette{
		Colours: map[pkgcolour.Role]pkgcolour.CategorisedColour{
			pkgcolour.RoleBackground: {Hex: "#eff1f5", RGBA: pkgcolour.RGBA{R: 239, G: 241, B: 245, A: 255}},
			pkgcolour.RoleForeground: {Hex: "#4c4f69", RGBA: pkgcolour.RGBA{R: 76, G: 79, B: 105, A: 255}},
			pkgcolour.RoleAccent1:    {Hex: "#1e66f5", RGBA: pkgcolour.RGBA{R: 30, G: 102, B: 245, A: 255}},
		},
		ThemeType: pkgcolour.ThemeLight,
	}

	// Create theme data
	darkTheme := colour.NewThemeData(darkPalette, "", "", "")
	lightTheme := colour.NewThemeData(lightPalette, "", "", "")

	// Create plugin
	plugin := New()

	// Test GenerateDualTheme
	files, err := plugin.GenerateDualTheme(darkTheme, lightTheme)
	if err != nil {
		t.Fatalf("GenerateDualTheme() error = %v", err)
	}

	// Verify all 4 variant files were generated (2 variants x 2 themes)
	if len(files) != 4 {
		t.Errorf("Expected 4 files (2 variants per theme), got %d", len(files))
	}

	// Verify TinctDark1.colors exists (check variant 1)
	darkContent, hasDark := files["TinctDark1.colors"]
	if !hasDark {
		t.Error("Missing TinctDark1.colors file")
	} else {
		// Verify it contains dark theme colors
		content := string(darkContent)
		if !strings.Contains(content, "Name=Tinct Dark 1") {
			t.Errorf("TinctDark1.colors missing 'Name=Tinct Dark 1'")
		}
		if !strings.Contains(content, "[General]") {
			t.Error("TinctDark1.colors missing [General] section")
		}
		if !strings.Contains(content, "ColorScheme=TinctDark1") {
			t.Error("TinctDark1.colors missing 'ColorScheme=TinctDark1'")
		}
		// Check for dark background color (30,30,46) - using rgbDecimal format
		if !strings.Contains(content, "30,30,46") {
			t.Error("TinctDark1.colors missing dark background color 30,30,46")
		}
	}

	// Verify TinctLight1.colors exists
	lightContent, hasLight := files["TinctLight1.colors"]
	if !hasLight {
		t.Error("Missing TinctLight1.colors file")
	} else {
		// Verify it contains light theme colors
		content := string(lightContent)
		if !strings.Contains(content, "Name=Tinct Light 1") {
			t.Errorf("TinctLight1.colors missing 'Name=Tinct Light 1'")
		}
		if !strings.Contains(content, "[General]") {
			t.Error("TinctLight1.colors missing [General] section")
		}
		if !strings.Contains(content, "ColorScheme=TinctLight1") {
			t.Error("TinctLight1.colors missing 'ColorScheme=TinctLight1'")
		}
		// Check for light background color (239,241,245) - using rgbDecimal format
		if !strings.Contains(content, "239,241,245") {
			t.Error("TinctLight1.colors missing light background color 239,241,245")
		}
	}

	// Verify variant 2 files also exist
	if _, ok := files["TinctDark2.colors"]; !ok {
		t.Error("Missing TinctDark2.colors file")
	}
	if _, ok := files["TinctLight2.colors"]; !ok {
		t.Error("Missing TinctLight2.colors file")
	}
}

func TestGenerateDualThemeWithNilPrimary(t *testing.T) {
	lightPalette := &colour.CategorisedPalette{
		Colours:   map[pkgcolour.Role]pkgcolour.CategorisedColour{},
		ThemeType: pkgcolour.ThemeLight,
	}
	lightTheme := colour.NewThemeData(lightPalette, "", "", "")

	plugin := New()
	_, err := plugin.GenerateDualTheme(nil, lightTheme)
	if err == nil {
		t.Error("Expected error with nil primary theme, got nil")
	}
}

func TestGenerateDualThemeWithNilAlternate(t *testing.T) {
	darkPalette := &colour.CategorisedPalette{
		Colours:   map[pkgcolour.Role]pkgcolour.CategorisedColour{},
		ThemeType: pkgcolour.ThemeDark,
	}
	darkTheme := colour.NewThemeData(darkPalette, "", "", "")

	plugin := New()
	_, err := plugin.GenerateDualTheme(darkTheme, nil)
	if err == nil {
		t.Error("Expected error with nil alternate theme, got nil")
	}
}
