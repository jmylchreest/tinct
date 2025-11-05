// Package colour provides utility functions for color manipulation and analysis.
package colour

import (
	"image/color"
	"math"
)

// Luminance calculates the relative luminance of a colour according to WCAG 2.0.
// Returns a value between 0 (darkest) and 1 (lightest).
// https://www.w3.org/TR/WCAG20/#relativeluminancedef.
func Luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Convert from 16-bit to 8-bit.
	rf := float64(r>>8) / 255.0
	rg := float64(g>>8) / 255.0
	rb := float64(b>>8) / 255.0

	// Apply gamma correction.
	rf = gammaCorrect(rf)
	rg = gammaCorrect(rg)
	rb = gammaCorrect(rb)

	// Calculate luminance using WCAG formula.
	return 0.2126*rf + 0.7152*rg + 0.0722*rb
}

// gammaCorrect applies gamma correction to a colour component.
func gammaCorrect(v float64) float64 {
	if v <= 0.03928 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// ContrastRatio calculates the contrast ratio between two colours according to WCAG 2.0.
// Returns a value between 1 and 21, where 21 is maximum contrast (black vs white).
// Meets WCAG AA standard for normal text at 4.5:1, large text at 3:1.
// https://www.w3.org/TR/WCAG20/#contrast-ratiodef.
func ContrastRatio(c1, c2 color.Color) float64 {
	l1 := Luminance(c1)
	l2 := Luminance(c2)

	// Ensure l1 is the lighter colour.
	if l1 < l2 {
		l1, l2 = l2, l1
	}

	return (l1 + 0.05) / (l2 + 0.05)
}

// HueDistance calculates the angular distance between two hues on the color wheel.
// Returns a value between 0 and 180 degrees (shortest path around the wheel).
// Used to determine if colors are analogous (<30°) or complementary (~180°).
func HueDistance(h1, h2 float64) float64 {
	diff := math.Abs(h1 - h2)
	if diff > 180 {
		diff = 360 - diff // Handle wraparound
	}
	return diff
}

// IsAnalogous checks if two colors are analogous (similar hue).
// Analogous colors are within 30° on the color wheel, creating visual harmony.
// Based on design theory for cohesive color schemes.
func IsAnalogous(h1, h2 float64) bool {
	return HueDistance(h1, h2) <= 30
}

// AdjustSaturation adjusts the saturation of a color by a given factor.
// factor < 1.0 reduces saturation (creates muted colors).
// factor > 1.0 increases saturation (creates more vibrant colors).
// factor = 1.0 leaves saturation unchanged.
func AdjustSaturation(h, s, l, factor float64) RGB {
	newS := math.Max(0.0, math.Min(1.0, s*factor))
	return HSLToRGB(h, newS, l)
}

// AdjustLuminance adjusts the luminance of a color by a delta value.
// delta > 0 makes the color lighter.
// delta < 0 makes the color darker.
// Result is clamped to [0.0, 1.0].
func AdjustLuminance(h, s, l, delta float64) RGB {
	newL := math.Max(0.0, math.Min(1.0, l+delta))
	return HSLToRGB(h, s, newL)
}

// rgbToHSL converts RGB to HSL colour space.
// Returns hue (0-360), saturation (0-1), lightness (0-1).
func rgbToHSL(rgb RGB) (h, s, l float64) {
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	maxVal := math.Max(r, math.Max(g, b))
	minVal := math.Min(r, math.Min(g, b))
	delta := maxVal - minVal

	// Lightness.
	l = (maxVal + minVal) / 2.0

	// Saturation.
	if delta == 0 {
		s = 0
		h = 0
		return
	}

	if l < 0.5 {
		s = delta / (maxVal + minVal)
	} else {
		s = delta / (2.0 - maxVal - minVal)
	}

	// Hue.
	switch maxVal {
	case r:
		h = (g - b) / delta
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/delta + 2
	case b:
		h = (r-g)/delta + 4
	}

	h *= 60
	return
}

// HSLToRGB converts HSL to RGB colour space.
// h is hue (0-360), s is saturation (0-1), l is luminance (0-1).
func HSLToRGB(h, s, l float64) RGB {
	if s == 0 {
		// Achromatic (grey).
		v := uint8(l * 255)
		return RGB{R: v, G: v, B: v}
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	r := hueToRGB(p, q, h+120)
	g := hueToRGB(p, q, h)
	b := hueToRGB(p, q, h-120)

	return RGB{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
	}
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	// Normalize t to 0-360 range.
	for t < 0 {
		t += 360
	}
	for t >= 360 {
		t -= 360
	}

	if t < 60 {
		return p + (q-p)*t/60
	}
	if t < 180 {
		return q
	}
	if t < 240 {
		return p + (q-p)*(240-t)/60
	}
	return p
}

// sortByLuminance sorts colours by luminance based on theme type.
// Dark theme: ascending (dark to light).
// Light theme: descending (light to dark).
func sortByLuminance(colours []CategorisedColour, themeType ThemeType) {
	n := len(colours)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			var shouldSwap bool
			if themeType == ThemeDark {
				// Ascending: dark to light.
				shouldSwap = colours[j].Luminance > colours[j+1].Luminance
			} else {
				// Descending: light to dark.
				shouldSwap = colours[j].Luminance < colours[j+1].Luminance
			}
			if shouldSwap {
				colours[j], colours[j+1] = colours[j+1], colours[j]
			}
		}
	}
}

// RGBToColor converts an RGB value to a color.Color (RGBA).
func RGBToColor(rgb RGB) color.Color {
	return color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

// RGBToRGBA converts an RGB struct to an RGBA struct with full opacity.
func RGBToRGBA(rgb RGB) RGBA {
	return RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
}

// adjustLuminanceForContrast iteratively adjusts luminance until minimum contrast is achieved.
// Used by foreground, accent, and semantic color generation to ensure WCAG compliance.
// stepSize defaults to 0.05 if set to 0.
func adjustLuminanceForContrast(h, s, targetLum float64, bgColor color.Color, minContrast float64, theme ThemeType, maxAttempts int) (float64, RGB) {
	stepSize := 0.05 // Default step size

	rgb := HSLToRGB(h, s, targetLum)
	testColor := RGBToColor(rgb)
	contrast := ContrastRatio(testColor, bgColor)

	attempts := 0
	for contrast < minContrast && attempts < maxAttempts {
		if theme == ThemeDark {
			targetLum = math.Min(0.99, targetLum+stepSize)
		} else {
			targetLum = math.Max(0.01, targetLum-stepSize)
		}
		rgb = HSLToRGB(h, s, targetLum)
		testColor = RGBToColor(rgb)
		contrast = ContrastRatio(testColor, bgColor)
		attempts++
	}

	return targetLum, rgb
}
