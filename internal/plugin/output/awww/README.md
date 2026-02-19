# awww Output Plugin

Generate wallpaper configuration for [awww](https://codeberg.org/LGFae/awww) (An Answer to your Wayland Wallpaper Woes), an efficient animated wallpaper daemon for Wayland compositors implementing the wlr-layer-shell protocol.

## Overview

The awww plugin generates a wallpaper configuration reference file and automatically applies the wallpaper (with configurable transitions) when an image source is provided. awww is particularly well-suited for animated GIF wallpapers and supports smooth transition effects when switching images.

## Features

- Automatic wallpaper application via `awww img` command
- Configurable transition effects (simple, wipe, grow, center, outer, etc.)
- Transition speed and FPS control
- Animated GIF wallpaper support
- Per-output wallpaper targeting
- Colour-matched clear screen using theme background colour
- Resize mode support (crop, fit, no)
- Fill colour support for fit mode
- Graceful handling when awww-daemon is not running

## Generated Files

- `tinct-awww.conf` - Wallpaper configuration reference with awww commands

## Default Output Location

```
~/.config/awww/tinct-awww.conf
```

## Usage

### With Image Input (Automatic Wallpaper)

When using the `image` input plugin, awww automatically:
1. Generates config with wallpaper path and transition settings
2. Applies wallpaper immediately with transition (if awww-daemon is running)

```bash
# Extract colours and set wallpaper with default transition
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o awww,hyprland,waybar

# With custom transition effect
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o awww \
  --awww.transition-type center \
  --awww.transition-step 90 \
  --awww.transition-fps 60
```

### Animated GIF Wallpapers

awww excels at displaying animated GIF wallpapers:

```bash
# Set an animated wallpaper
tinct generate -i image -p ~/Pictures/animated-bg.gif -o awww,hyprland

# AI-generated palette with animated wallpaper
tinct generate -i google-genai --ai.prompt "cyberpunk cityscape" -o awww
```

### Without Image Input

When using non-image sources, the config is generated with helpful commands:

```bash
# Generate from remote theme
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  -o awww,hyprland

# Config generated with clear screen command using theme colours
```

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `--awww.output-dir` | `~/.config/awww` | Output directory |
| `--awww.transition-type` | `simple` | Transition effect type |
| `--awww.transition-step` | `2` | Transition step (1-255, smaller = smoother) |
| `--awww.transition-fps` | `30` | Transition FPS (1-65535) |
| `--awww.resize` | | Resize mode (crop, fit, no) |
| `--awww.fill-color` | | Fill colour for fit mode (rrggbb) |
| `--awww.outputs` | | Comma-separated output names |

## Transition Types

| Type | Description |
|------|-------------|
| `simple` | Fade into the new image (default) |
| `left` | Slide from the left |
| `right` | Slide from the right |
| `top` | Slide from the top |
| `bottom` | Slide from the bottom |
| `wipe` | Wipe transition (supports angle) |
| `grow` | Grow from a point |
| `center` | Expand from center to edges |
| `outer` | Contract from edges to center |
| `any` | Random point with center or outer |
| `random` | Random transition effect |

## Pre-Execute Check

The plugin performs these checks:

1. Warns if `awww` binary is not found on PATH (non-fatal)
2. Creates config directory if it doesn't exist

The plugin will still generate configuration files even if awww is not running - wallpaper application is simply skipped.

## Behavior Matrix

| Input Plugin | awww-daemon Running | Result |
|--------------|---------------------|--------|
| `image` | Yes | Config generated + wallpaper applied with transition |
| `image` | No | Config generated, wallpaper skipped (warning in verbose) |
| `remote-json` | Yes | Config generated with clear screen command |
| `remote-json` | No | Config generated with clear screen command |

## AI-Generated Animated Backgrounds

awww's animated GIF support makes it an ideal target for AI-generated animated wallpapers. While current AI image generation APIs produce static images, the concept of AI-generated animated backgrounds is an area of active development:

### Current Workflow

```bash
# Generate a static AI wallpaper and apply with awww
tinct generate -i google-genai --ai.prompt "serene forest at sunset" -o awww

# The AI generates a static image, extracts colours, and awww displays it
```

### Future Possibilities

As AI video/animation generation APIs mature (Google Veo, Runway, Luma, etc.), a dedicated input plugin could:
1. Generate short animated clips from text prompts
2. Convert them to GIF/animated WebP format
3. Pass the animated file to awww for display
4. Extract colours from representative frames for consistent theming

This would enable fully AI-generated animated desktop wallpapers with matching colour themes across all applications.

## Example Workflows

### Daily Wallpaper Theme Update

```bash
#!/bin/bash
# Update theme from wallpaper with smooth transition

WALLPAPER="$HOME/.wallpaper"

tinct generate -i image -p "$WALLPAPER" \
  -o awww,hyprland,hyprlock,waybar \
  --awww.transition-type grow \
  --awww.transition-fps 60

# awww automatically applies wallpaper with grow transition
```

### Wallpaper Slideshow

```bash
#!/bin/bash
# Cycle through wallpapers with matching themes
for img in ~/Pictures/Wallpapers/*.{jpg,png,gif}; do
    tinct generate -i image -p "$img" \
      -o awww,hyprland,waybar \
      --awww.transition-type random
    sleep 300
done
```

### Clear Screen with Theme Colour

```bash
# Set a solid colour background matching your theme
tinct generate -i remote-json \
  --remote-json.url "https://..." -o awww

# Then manually run the clear command from the generated config
```

## Troubleshooting

### Wallpaper Not Applied

Check if awww-daemon is running:

```bash
# Start the daemon
awww-daemon &

# Check if it's running
awww query

# Then apply your theme
tinct generate -i image -p wallpaper.jpg -o awww
```

### High CPU During GIF Caching

awww caches GIF frames on first display. For large GIFs, consider resizing before sending to awww:

```bash
# Resize GIF with gifsicle
gifsicle --resize-fit 1920x1080 large-animation.gif > optimised.gif
tinct generate -i image -p optimised.gif -o awww
```

### Verbose Mode

See detailed wallpaper application process:

```bash
tinct generate -i image -p wallpaper.jpg -o awww --verbose
```

## Resources

- [awww Repository](https://codeberg.org/LGFae/awww)
- [awww Documentation](https://codeberg.org/LGFae/awww/src/branch/main/README.md)

## See Also

- [Hyprpaper Plugin](../hyprpaper/README.md) - Alternative wallpaper daemon for Hyprland
- [wbg Plugin](../wbg/README.md) - Simple static wallpaper application
- [Image Input Plugin](../../input/image/README.md) - Extract colours from images
