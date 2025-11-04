package kitty

import (
	"image/color"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
)

func TestKittyPluginWithPaletteHelper(t *testing.T) {
	// Create a simple test palette
	colors := []color.Color{
		color.RGBA{R: 30, G: 30, B: 46, A: 255},    // dark bg
		color.RGBA{R: 205, G: 214, B: 244, A: 255}, // light fg
		color.RGBA{R: 137, G: 180, B: 250, A: 255}, // blue accent
		color.RGBA{R: 180, G: 190, B: 254, A: 255}, // lighter blue
		color.RGBA{R: 243, G: 139, B: 168, A: 255}, // pink
		color.RGBA{R: 166, G: 227, B: 161, A: 255}, // green
	}

	palette := colour.NewPalette(colors)
	config := colour.DefaultCategorisationConfig()
	categorised := colour.Categorise(palette, config)

	// Create plugin and generate
	plugin := New()
	themeData := colour.NewThemeData(categorised, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that tinct.conf was created
	content, ok := files["tinct.conf"]
	if !ok {
		t.Fatal("tinct.conf not generated")
	}

	contentStr := string(content)

	// Verify it contains expected colors in hex format
	if !strings.Contains(contentStr, "#1e1e2e") { // background
		t.Error("Missing background color in output")
	}

	if !strings.Contains(contentStr, "#cdd6f4") { // foreground
		t.Error("Missing foreground color in output")
	}

	// Verify template functions worked (should have "themeType" output)
	if !strings.Contains(contentStr, "Detected theme:") {
		t.Error("Missing theme type in output")
	}

	// Verify no template syntax errors (no {{ or }})
	if strings.Contains(contentStr, "{{") || strings.Contains(contentStr, "}}") {
		t.Error("Template syntax not fully processed")
	}

	// Check for new surface colors if they exist
	if strings.Contains(contentStr, "surface") {
		t.Log("Surface colors are being generated (expected with new system)")
	}

	t.Logf("Generated config length: %d bytes", len(content))
}
