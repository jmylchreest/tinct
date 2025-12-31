---
sidebar_position: 4
---

# Theme detection

Tinct automatically detects whether a colour palette suggests a dark or light theme. This page explains the detection algorithm and how to override it.

## Automatic detection

Theme detection is based on the luminance of the background colour:

1. Extract the dominant background colour from the palette
2. Calculate its relative luminance (per WCAG definition)
3. Compare against a threshold (~0.18)
4. Classify as dark or light

### Luminance calculation

Relative luminance is calculated using the sRGB formula:

```
L = 0.2126 * R + 0.7152 * G + 0.0722 * B
```

Where R, G, B are linearised from sRGB values.

### Threshold

- **L < 0.18**: Dark theme
- **L >= 0.18**: Light theme

This threshold is tuned to match human perception of "dark" vs "light" backgrounds.

## Override detection

Force a specific theme type:

```bash
# Force dark theme
tinct generate -i image -p ~/wallpaper.jpg -o all -t dark

# Force light theme
tinct generate -i image -p ~/wallpaper.jpg -o all -t light

# Auto-detect (default)
tinct generate -i image -p ~/wallpaper.jpg -o all -t auto
```

## Theme-specific adjustments

Depending on the theme type, tinct adjusts colour generation:

### Dark themes

- Background is the darkest extracted colour
- Foreground is a light, contrasting colour
- Surface colours are slightly lighter than background
- Accents may be slightly brightened for visibility

### Light themes

- Background is the lightest extracted colour
- Foreground is a dark, contrasting colour
- Surface colours are slightly darker than background
- Accents may be slightly darkened for visibility

## Semantic colour generation

Semantic colours (danger, warning, success, info) are generated to match the theme:

| Theme | Semantic colour adjustment |
|-------|---------------------------|
| Dark | Slightly desaturated, good contrast against dark surfaces |
| Light | Full saturation, good contrast against light surfaces |

## On-colour generation

"On-colour" variants are calculated for readability:

- **Dark background**: Light text (usually white or near-white)
- **Light background**: Dark text (usually near-black)

The specific shade is chosen to meet WCAG AA contrast requirements (4.5:1 minimum).

## Per-plugin behaviour

Some plugins behave differently based on theme type:

### KDE Plasma

Generates either `TinctDark.colors` or `TinctLight.colors`.

### GNOME Shell

Both themes use the same output, but CSS variables differ.

### Terminals

ANSI colour mappings are adjusted for the theme:

- Dark theme: color0 is background, color7 is foreground
- Light theme: color0 is foreground, color7 is background

## Checking detected theme

Use the `--preview` flag to see the detected theme:

```bash
tinct generate -i image -p ~/wallpaper.jpg -o all --preview
```

The output will show:

```
Theme type: dark
```

## Input plugin theme hints

Some input plugins provide theme hints:

| Plugin | Detection method |
|--------|------------------|
| `image` | Luminance of dominant colour |
| `remote-json` | May include theme metadata |
| `file` | Based on specified background |
| `google-genai` | Prompt-based inference |

## Next steps

- [Colour roles](/docs/concepts/color-roles) - Role-based colour system
- [Material Design 3](/docs/concepts/material-design) - Design system compliance
- [Generate command](/docs/commands/generate) - Full command options
