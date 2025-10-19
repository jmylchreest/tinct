// Package main demonstrates semantic color enhancement in Tinct.
// This example shows how Tinct generates vibrant, distinct semantic colors
// even from monochrome or muted palettes.
package main

import (
	"fmt"
	"image/color"

	"github.com/jmylchreest/tinct/internal/colour"
)

func main() {
	fmt.Println("=== Tinct Semantic Color Enhancement Demo ===")
	fmt.Println()

	// Demo 1: Monochrome palette (greys only)
	fmt.Println("1. Monochrome Palette (Dark Theme)")
	fmt.Println("   Input: Only grey colors")
	fmt.Println("   Expected: Generate all semantic colors with proper hues")
	fmt.Println()

	monochromeColors := []color.Color{
		color.RGBA{R: 30, G: 30, B: 40, A: 255},    // Dark background
		color.RGBA{R: 200, G: 200, B: 210, A: 255}, // Light foreground
		color.RGBA{R: 100, G: 100, B: 110, A: 255}, // Mid grey
		color.RGBA{R: 60, G: 60, B: 70, A: 255},    // Dark grey
	}

	monoPalette := &colour.Palette{Colors: monochromeColors}
	monoConfig := colour.DefaultCategorisationConfig()
	monoConfig.ThemeType = colour.ThemeDark

	monoCategorised := colour.Categorise(monoPalette, monoConfig)
	fmt.Println(monoCategorised.StringWithPreview(true))
	fmt.Println()

	// Demo 2: Muted colorful palette (low saturation)
	fmt.Println()
	fmt.Println("2. Muted Colorful Palette (Light Theme)")
	fmt.Println("   Input: Desaturated, muddy colors")
	fmt.Println("   Expected: Boost saturation and adjust lightness")
	fmt.Println()

	mutedColors := []color.Color{
		color.RGBA{R: 240, G: 240, B: 245, A: 255}, // Light background
		color.RGBA{R: 30, G: 30, B: 40, A: 255},    // Dark foreground
		color.RGBA{R: 160, G: 100, B: 100, A: 255}, // Dull red
		color.RGBA{R: 140, G: 130, B: 90, A: 255},  // Dull orange
		color.RGBA{R: 100, G: 140, B: 100, A: 255}, // Dull green
		color.RGBA{R: 90, G: 110, B: 160, A: 255},  // Dull blue
		color.RGBA{R: 130, G: 100, B: 150, A: 255}, // Dull purple
	}

	mutedPalette := &colour.Palette{Colors: mutedColors}
	mutedConfig := colour.DefaultCategorisationConfig()
	mutedConfig.ThemeType = colour.ThemeLight

	mutedCategorised := colour.Categorise(mutedPalette, mutedConfig)
	fmt.Println(mutedCategorised.StringWithPreview(true))
	fmt.Println()

	// Demo 3: Vibrant palette (high saturation already)
	fmt.Println()
	fmt.Println("3. Vibrant Palette (Dark Theme)")
	fmt.Println("   Input: Already saturated colors")
	fmt.Println("   Expected: Fine-tune for optimal contrast")
	fmt.Println()

	vibrantColors := []color.Color{
		color.RGBA{R: 20, G: 20, B: 30, A: 255},    // Dark background
		color.RGBA{R: 230, G: 230, B: 240, A: 255}, // Light foreground
		color.RGBA{R: 230, G: 60, B: 60, A: 255},   // Vibrant red
		color.RGBA{R: 230, G: 180, B: 50, A: 255},  // Vibrant orange
		color.RGBA{R: 60, G: 230, B: 60, A: 255},   // Vibrant green
		color.RGBA{R: 60, G: 140, B: 230, A: 255},  // Vibrant blue
		color.RGBA{R: 180, G: 60, B: 230, A: 255},  // Vibrant purple
	}

	vibrantPalette := &colour.Palette{Colors: vibrantColors}
	vibrantConfig := colour.DefaultCategorisationConfig()
	vibrantConfig.ThemeType = colour.ThemeDark

	vibrantCategorised := colour.Categorise(vibrantPalette, vibrantConfig)
	fmt.Println(vibrantCategorised.StringWithPreview(true))
	fmt.Println()

	// Demo 4: Show semantic color properties
	fmt.Println()
	fmt.Println("4. Semantic Color Properties")
	fmt.Println("   Demonstrating hue, saturation, and contrast requirements")
	fmt.Println()

	semanticRoles := []colour.ColourRole{
		colour.RoleDanger,
		colour.RoleWarning,
		colour.RoleSuccess,
		colour.RoleInfo,
		colour.RoleNotification,
	}

	fmt.Println("From Monochrome Palette (generated):")
	bg, _ := monoCategorised.Get(colour.RoleBackground)
	for _, role := range semanticRoles {
		if cc, ok := monoCategorised.Get(role); ok {
			contrast := colour.ContrastRatio(cc.Colour, bg.Colour)
			fmt.Printf("  %-15s: hue=%3.0f°, sat=%.2f, lum=%.2f, contrast=%.2f:1, %s\n",
				role, cc.Hue, cc.Saturation, cc.Luminance, contrast, cc.Hex)
		}
	}

	fmt.Println("\nFrom Muted Palette (enhanced):")
	bgLight, _ := mutedCategorised.Get(colour.RoleBackground)
	for _, role := range semanticRoles {
		if cc, ok := mutedCategorised.Get(role); ok {
			contrast := colour.ContrastRatio(cc.Colour, bgLight.Colour)
			fmt.Printf("  %-15s: hue=%3.0f°, sat=%.2f, lum=%.2f, contrast=%.2f:1, %s\n",
				role, cc.Hue, cc.Saturation, cc.Luminance, contrast, cc.Hex)
		}
	}

	// Demo 5: Semantic hue standards
	fmt.Println()
	fmt.Println("5. Semantic Color Hue Standards")
	fmt.Println("   Standard hues used for fallback generation:")
	fmt.Println()

	for role, hue := range colour.SemanticHues {
		colorName := getColorName(hue)
		fmt.Printf("  %-15s: %3.0f° (%s)\n", role, hue, colorName)
	}

	fmt.Println("\n=== Key Features ===")
	fmt.Println("✓ Minimum saturation: 60% for all semantic colors")
	fmt.Println("✓ Theme-aware lightness adjustment (darker for light themes, lighter for dark themes)")
	fmt.Println("✓ Automatic fallback generation when colors missing from palette")
	fmt.Println("✓ Contrast checking: minimum 3:1 ratio with background")
	fmt.Println("✓ Distinct hues: minimum 30° separation between semantic colors")
	fmt.Println()
}

func getColorName(hue float64) string {
	switch {
	case hue >= 350 || hue < 30:
		return "Red"
	case hue >= 30 && hue < 60:
		return "Orange"
	case hue >= 60 && hue < 150:
		return "Yellow-Green"
	case hue >= 90 && hue < 150:
		return "Green"
	case hue >= 150 && hue < 210:
		return "Cyan"
	case hue >= 210 && hue < 270:
		return "Blue"
	case hue >= 270 && hue < 330:
		return "Purple"
	default:
		return "Magenta"
	}
}
