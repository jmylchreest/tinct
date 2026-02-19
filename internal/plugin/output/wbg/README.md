# wbg Output Plugin

Generate wallpaper configuration for [wbg](https://codeberg.org/dnkl/wbg), a super simple wallpaper application for Wayland compositors implementing the layer-shell protocol.

## Overview

The wbg plugin generates a wallpaper configuration reference file and automatically applies the wallpaper when an image source is provided. wbg is the simplest wallpaper tool available -- it takes a single image path and displays it scaled-to-fit on all monitors.

## Features

- Automatic wallpaper application via `wbg` process management
- Kill-and-restart pattern for seamless wallpaper changes
- Compositor integration examples (Hyprland, Sway)
- Graceful handling when wbg is not installed

## Generated Files

- `tinct-wbg.conf` - Wallpaper configuration reference with wbg commands

## Default Output Location

```
~/.config/wbg/tinct-wbg.conf
```

## Usage

### With Image Input (Automatic Wallpaper)

When using the `image` input plugin, the wbg plugin automatically:
1. Generates config with wallpaper path
2. Kills any existing wbg process
3. Starts a new wbg instance with the wallpaper

```bash
# Extract colours and set wallpaper
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o wbg,hyprland,waybar
```

### Without Image Input

When using non-image sources, a config is generated with usage instructions:

```bash
# Generate from remote theme
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  -o wbg,hyprland
```

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `--wbg.output-dir` | `~/.config/wbg` | Output directory |

## Pre-Execute Check

The plugin performs these checks:

1. Warns if `wbg` binary is not found on PATH (non-fatal)
2. Creates config directory if it doesn't exist

## Behavior Matrix

| Input Plugin | wbg Installed | Result |
|--------------|---------------|--------|
| `image` | Yes | Config generated + wallpaper applied |
| `image` | No | Config generated, wallpaper skipped |
| `remote-json` | Any | Config generated with usage instructions |

## How Wallpaper Application Works

When a wallpaper source is provided and wbg is installed:

1. **Kill existing**: Runs `killall wbg` to stop any running instance
2. **Start new**: Launches `wbg /path/to/wallpaper` as a background process
3. **Detach**: Releases the process so it continues after tinct exits

This ensures seamless wallpaper changes without manual intervention.

## Integration Examples

### Hyprland

```conf
# ~/.config/hypr/hyprland.conf
exec-once = wbg ~/Pictures/wallpaper.jpg
```

### Sway

```conf
# ~/.config/sway/config
exec wbg ~/Pictures/wallpaper.jpg
```

### Script-Based Wallpaper Rotation

```bash
#!/bin/bash
# Rotate wallpapers every 5 minutes
while true; do
    img=$(find ~/Pictures/Wallpapers -type f | shuf -n 1)
    tinct generate -i image -p "$img" -o wbg,hyprland,waybar
    sleep 300
done
```

## Supported Image Formats

wbg supports the following formats (depending on compile-time options):

- PNG
- JPEG
- JPEG XL
- WebP
- SVG

## Troubleshooting

### Wallpaper Not Applied

Check if wbg is installed:

```bash
which wbg
```

### Multiple wbg Instances

The plugin automatically kills existing wbg processes before starting a new one. If you see multiple instances, check for other scripts or autostart entries that launch wbg.

### Verbose Mode

See detailed process:

```bash
tinct generate -i image -p wallpaper.jpg -o wbg --verbose
```

## Resources

- [wbg Repository](https://codeberg.org/dnkl/wbg)
- [wbg Releases](https://codeberg.org/dnkl/wbg/releases)

## See Also

- [awww Plugin](../awww/README.md) - Animated wallpaper daemon with transitions
- [Hyprpaper Plugin](../hyprpaper/README.md) - Wallpaper daemon for Hyprland
- [Image Input Plugin](../../input/image/README.md) - Extract colours from images
