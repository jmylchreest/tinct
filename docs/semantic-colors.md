# Semantic Color Enhancement

## Overview

Tinct intelligently generates and enhances semantic colors to ensure they are vibrant, distinct, and accessible. Semantic colors (danger, warning, success, info, notification) are automatically assigned from your palette or generated as fallbacks when missing, with adjustments to ensure they effectively communicate meaning.

## Why Semantic Color Enhancement?

When extracting colors from images, especially monochrome or muted palettes, the resulting colors may not be suitable for semantic purposes:

- **Low Saturation**: Muddy, greyish colors that don't convey urgency or meaning
- **Poor Contrast**: Colors that blend into the background
- **Missing Colors**: Some hue ranges may not be present in the image
- **Inconsistent Lightness**: Colors that are too dark or too light for the theme

Tinct's semantic color enhancement solves these issues automatically.

## Semantic Color Roles

| Role | Hue | Color | Purpose |
|------|-----|-------|---------|
| `danger` | 0° | Red | Errors, destructive actions, critical alerts |
| `warning` | 45° | Orange | Warnings, caution, attention needed |
| `success` | 120° | Green | Success messages, confirmations, positive actions |
| `info` | 210° | Blue | Information, help, neutral notifications |
| `notification` | 285° | Purple | Notifications, badges, highlights |

## Enhancement Features

### 1. Saturation Boosting

**Minimum Saturation**: 60% (0.6 in HSL)

All semantic colors are boosted to at least 60% saturation to ensure they are vibrant and eye-catching.

```
Before: #a06464 (sat: 0.24) → After: #c81c1c (sat: 0.75)
```

### 2. Theme-Aware Lightness

Semantic colors are adjusted based on the detected theme to ensure optimal visibility:

| Theme Type | Target Lightness | Reason |
|------------|-----------------|--------|
| Dark | 55-65% | Lighter colors stand out on dark backgrounds |
| Light | 35-50% | Darker colors stand out on light backgrounds |

```
Dark theme:  #c81c1c (dark red) → #e54c4c (lighter red)
Light theme: #e54c4c (light red) → #c81c1c (darker red)
```

### 3. Contrast Assurance

**Minimum Contrast**: 3:1 with background

All semantic colors are verified to have at least a 3:1 contrast ratio with the background. If contrast is insufficient, lightness is automatically adjusted.

### 4. Fallback Generation

If your palette doesn't contain colors in a semantic hue range, Tinct automatically generates appropriate fallback colors:

```
Monochrome Palette (only greys):
  Input:  #1e1e28, #c8c8d2, #64646e
  Output: 
    danger:       #e54c4c (generated red)
    warning:      #e5bf4c (generated orange)
    success:      #4ce54c (generated green)
    info:         #4c99e5 (generated blue)
    notification: #bf4ce5 (generated purple)
```

### 5. Hue Standardization

When enhancing existing colors, hues are normalized to standard semantic values:

```
Extracted:  hue=15° (reddish-orange)
Enhanced:   hue=0°  (pure red for danger)

Extracted:  hue=110° (yellow-green)
Enhanced:   hue=120° (pure green for success)
```

## Configuration

### Default Configuration

```go
config := colour.DefaultCategorisationConfig()
// EnhanceSemanticColors: true
// SemanticBoostAmount: 0.3 (30% saturation boost)
```

### Custom Configuration

```go
config := colour.CategorisationConfig{
    ThemeType:             colour.ThemeDark,
    MinContrastRatio:      4.5,
    EnhanceSemanticColors: true,
    SemanticBoostAmount:   0.3,
}

categorised := colour.Categorise(palette, config)
```

### Disable Enhancement

If you want to use extracted colors as-is without enhancement:

```go
config := colour.DefaultCategorisationConfig()
config.EnhanceSemanticColors = false
```

## Examples

### Example 1: Monochrome Palette

**Input**: Dark background (#1e1e28), light foreground (#c8c8d2), grey accent (#64646e)

**Output** (Dark Theme):
```
danger:       #e54c4c (sat: 0.75, contrast: 4.30:1)
warning:      #e5bf4c (sat: 0.75, contrast: 9.35:1)
success:      #4ce54c (sat: 0.75, contrast: 9.92:1)
info:         #4c99e5 (sat: 0.75, contrast: 5.50:1)
notification: #bf4ce5 (sat: 0.75, contrast: 4.23:1)
```

All semantic colors are generated as fallbacks with optimal saturation and contrast.

### Example 2: Muted Colorful Palette

**Input**: Light background (#f0f0f5), with desaturated reds, greens, blues

**Before Enhancement**:
```
dull red:    #a06464 (sat: 0.24, lum: 0.24)
dull orange: #8c825a (sat: 0.22, lum: 0.31)
dull green:  #648264 (sat: 0.20, lum: 0.24)
```

**After Enhancement** (Light Theme):
```
danger:  #c81c1c (sat: 0.75, lum: 0.13, contrast: 5.08:1)
warning: #9c7a16 (sat: 0.75, lum: 0.21, contrast: 3.55:1)
success: #169c16 (sat: 0.75, lum: 0.24, contrast: 3.19:1)
```

Saturation boosted to 75%, lightness adjusted for light theme contrast.

### Example 3: Vibrant Palette

**Input**: Already saturated colors (#e63c3c, #3ce63c, #3c8ce6)

**Enhancement**: Fine-tuning only
```
Before: #e63c3c (sat: 0.77, lum: 0.26)
After:  #e64343 (sat: 0.77, lum: 0.27) - minimal adjustment
```

Already vibrant colors receive only minor adjustments for optimal contrast.

## CLI Usage

### Extract with Semantic Colors

```bash
# Auto-detect theme and generate semantic colors
tinct extract wallpaper.jpg --categorise --preview

# Force dark theme
tinct extract wallpaper.jpg --categorise --theme dark --preview

# Force light theme
tinct extract wallpaper.jpg --categorise --theme light --preview
```

### JSON Output

```bash
tinct extract wallpaper.jpg --categorise --format json
```

Output includes semantic colors with full metadata:

```json
{
  "colours": {
    "danger": {
      "role": "danger",
      "hex": "#e54c4c",
      "rgb": {"r": 229, "g": 76, "b": 76},
      "luminance": 0.22,
      "is_light": false,
      "hue": 0,
      "saturation": 0.75
    },
    "warning": {
      "role": "warning",
      "hex": "#e5bf4c",
      "rgb": {"r": 229, "g": 191, "b": 76},
      "luminance": 0.54,
      "is_light": true,
      "hue": 45,
      "saturation": 0.75
    }
  },
  "theme_type": "dark"
}
```

## Technical Details

### HSL Color Space

Semantic colour enhancement uses the HSL (Hue, Saturation, Lightness) colour space for adjustments:

- **Hue (H)**: 0-360° color wheel position
- **Saturation (S)**: 0-1, intensity of color (0 = grey, 1 = pure color)
- **Lightness (L)**: 0-1, brightness (0 = black, 0.5 = pure color, 1 = white)

### WCAG Luminance

Contrast checking uses WCAG 2.0 relative luminance, which accounts for human perception:

```
Luminance = 0.2126*R + 0.7152*G + 0.0722*B
```

Note: WCAG luminance differs from HSL lightness due to gamma correction and perceptual weighting.

### Enhancement Algorithm

1. **Extract colors** from image using k-means clustering
2. **Detect theme** based on average luminance
3. **Assign semantic roles** by matching colors to hue ranges
4. **Enhance saturation** to minimum threshold (60%)
5. **Adjust lightness** based on theme (darker for light, lighter for dark)
6. **Check contrast** with background (minimum 3:1)
7. **Generate fallbacks** for missing semantic colors

### Fallback Generation

When a semantic color is missing from the palette:

1. Use standard hue from `SemanticHues` map
2. Set saturation to 75% (vibrant)
3. Set lightness based on theme:
   - Dark theme: 60% (lighter)
   - Light theme: 45% (darker)
4. Iteratively adjust lightness for 3:1 contrast with background
5. Convert HSL back to RGB

## Best Practices

### For Theme Designers

1. **Trust the Enhancement**: Let Tinct adjust colors for optimal visibility
2. **Preview Before Applying**: Use `--preview` to see colors in terminal
3. **Check Contrast**: Verify semantic colors work with your background
4. **Test Both Themes**: Try both `--theme dark` and `--theme light`

### For Plugin Developers

```go
// Get enhanced semantic colors
categorised := colour.Categorise(palette, config)

danger, _ := categorised.Get(colour.RoleDanger)
warning, _ := categorised.Get(colour.RoleWarning)
success, _ := categorised.Get(colour.RoleSuccess)

// Use in your config
fmt.Printf("error_color: %s\n", danger.Hex)
fmt.Printf("warning_color: %s\n", warning.Hex)
fmt.Printf("success_color: %s\n", success.Hex)
```

### For Application Integration

Semantic colors are guaranteed to:
- ✅ Have at least 60% saturation (vibrant)
- ✅ Have 3:1 contrast with background (visible)
- ✅ Use standard hues (recognizable)
- ✅ Be theme-appropriate (light/dark)
- ✅ Always be available (fallback generation)

## Troubleshooting

### Colors Look Too Saturated

Reduce the saturation boost amount:

```go
config.SemanticBoostAmount = 0.1 // 10% instead of 30%
```

Or disable enhancement entirely:

```go
config.EnhanceSemanticColors = false
```

### Colors Don't Match My Palette

This is expected! Semantic colors are intentionally adjusted for visibility and meaning. If you need colors closer to the original palette:

```go
// Use accent colors instead
accent1, _ := categorised.Get(colour.RoleAccent1)
accent2, _ := categorised.Get(colour.RoleAccent2)
```

Accent colors maintain saturation from the original palette.

### Poor Contrast in Specific Applications

Increase minimum contrast ratio:

```go
config.MinContrastRatio = 7.0  // AAA level
config.RequireAAA = true
```

### Wrong Theme Detected

Force the theme type explicitly:

```bash
tinct extract image.jpg --categorise --theme dark
```

Or in code:

```go
config.ThemeType = colour.ThemeDark
```

## Related Documentation

- [Color Categorisation](./categorisation.md) - Core categorisation system
- [Accessibility](./accessibility.md) - WCAG contrast standards
- [Plugin Development](./plugin-development.md) - Using semantic colors in plugins

## Demo

Run the semantic colors demo to see enhancement in action:

```bash
cd tinct
go run examples/semantic_colors_demo.go
```

This demonstrates:
- Monochrome palette → full semantic colors
- Muted palette → saturation boosting
- Vibrant palette → fine-tuning
- Theme-aware adjustments
- Contrast checking

---

**Made with 🎨 and Go 1.25+**