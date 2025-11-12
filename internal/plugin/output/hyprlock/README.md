# Hyprlock Output Plugin

Generate colour themes for [Hyprlock](https://github.com/hyprwm/hyprlock), the GPU-accelerated screen lock for Hyprland.

## Overview

The Hyprlock plugin generates a colour configuration file (`tinct-hyprlock.conf`) with Hyprland-style colour variables that can be sourced in your hyprlock configuration.

## Features

- Single configuration file generation
- Hyprland variable syntax (`$variable`)
- Both RGB and RGBA colour formats
- Semantic colour mapping
- Pre-defined colour variables for quick theming
- Wallpaper path variable (when using image input plugin)
- Usage examples in comments

## Wallpaper Support

When using the `image` input plugin, hyprlock receives the wallpaper path and includes it as a variable:

```bash
# Generate with wallpaper
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o hyprlock,hyprpaper
```

The generated config includes:

```conf
$tinct_wallpaper = /home/user/Pictures/wallpaper.jpg
```

Use it in your hyprlock configuration:

```conf
source = ~/.config/hypr/tinct-hyprlock.conf

background {
    path = $tinct_wallpaper
    colour = $tinct_background
    blur_passes = 3
}
```

**Note:** When using non-image input sources (remote-json, remote-css, file), the `$tinct_wallpaper` variable is not included.

## Generated Files

- `tinct-hyprlock.conf` - Complete colour variable definitions

## Default Output Location

```
~/.config/hypr/tinct-hyprlock.conf
```

## Configuration

### Enable the Plugin

In your `~/.config/tinct/config.toml`:

```toml
[[output.hyprlock]]
enabled = true
```

### Custom Output Directory

```toml
[[output.hyprlock]]
enabled = true
output_dir = "/custom/path"
```

### Command Line Options

```bash
# Specify output directory
tinct generate --hyprlock.output-dir ~/.config/hypr

# Enable hyprlock only
tinct generate --output hyprlock
```

## Usage

### Method 1: Source in hyprlock.conf

Edit `~/.config/hypr/hyprlock.conf`:

```conf
# Source tinct colours
source = ~/.config/hypr/tinct-hyprlock.conf

# Use the colour variables
background {
    monitor =
    colour = $tinct_background
    blur_passes = 3
}

input-field {
    monitor =
    size = 20%, 5%
    outline_thickness = 3
    
    inner_color = $tinct_background_rgba
    outer_color = $tinct_accent1_rgba
    check_color = $tinct_success_rgba
    fail_color = $tinct_danger_rgba
    
    font_color = $tinct_foreground
    placeholder_text = <i>Input password...</i>
    
    position = 0, -20
    halign = center
    valign = center
}

label {
    monitor =
    text = $TIME
    font_size = 90
    colour = $tinct_foreground
    
    position = -30, 0
    halign = right
    valign = top
}
```

### Method 2: Copy Variables

Alternatively, copy the generated colour variables directly into your `hyprlock.conf`.

## Colour Variables

The plugin generates two sets of colour variables:

### RGB Variables (for solid colours)

```conf
$tinct_background
$tinct_background_muted
$tinct_foreground
$tinct_foreground_muted
$tinct_accent1
$tinct_accent2
$tinct_accent3
$tinct_accent4
$tinct_danger
$tinct_warning
$tinct_success
$tinct_info
```

### RGBA Variables (with transparency)

```conf
$tinct_background_rgba       # 93% opacity
$tinct_background_muted_rgba # 80% opacity
$tinct_foreground_rgba       # 100% opacity
$tinct_foreground_muted_rgba # 87% opacity
$tinct_accent1_rgba          # 100% opacity
$tinct_accent2_rgba          # 100% opacity
$tinct_accent3_rgba          # 100% opacity
$tinct_accent4_rgba          # 100% opacity
$tinct_danger_rgba           # 100% opacity
$tinct_warning_rgba          # 100% opacity
$tinct_success_rgba          # 100% opacity
$tinct_info_rgba             # 100% opacity
```

## Colour Format

### RGB Format

Hyprland uses `rgb(RRGGBB)` format:

```conf
$tinct_background = rgb(1a1b26)
```

### RGBA Format

Hyprland uses `rgba(R, G, B, A)` where:
- R, G, B are 0-255 (decimal)
- A is 0.0-1.0 (float)

```conf
$tinct_background_rgba = rgba(26, 27, 38, 0.93)
```

## Colour Mapping

| Variable | Tinct Role | Common Usage |
|----------|------------|--------------|
| `$tinct_background` | Background | Solid backgrounds |
| `$tinct_background_rgba` | Background (93% opacity) | Transparent backgrounds |
| `$tinct_foreground` | Foreground | Text, labels |
| `$tinct_accent1` | Accent1 | Primary highlights, borders |
| `$tinct_accent1_rgba` | Accent1 (transparent) | Input field outlines |
| `$tinct_success` | Success | Check/success states |
| `$tinct_danger` | Danger | Error/fail states |
| `$tinct_warning` | Warning | Warning states |
| `$tinct_info` | Info | Information displays |

## Example Configuration

Complete example using Tinct colours:

```conf
# Source tinct colours
source = ~/.config/hypr/tinct-hyprlock.conf

general {
    hide_cursor = false
}

background {
    monitor =
    path = screenshot
    colour = $tinct_background
    blur_passes = 3
}

# Password input field
input-field {
    monitor =
    size = 20%, 5%
    outline_thickness = 3
    
    # Use transparent background, accent for outline
    inner_color = $tinct_background_rgba
    outer_color = $tinct_accent1_rgba
    
    # Success = green, fail = red
    check_color = $tinct_success_rgba
    fail_color = $tinct_danger_rgba
    
    font_color = $tinct_foreground
    fade_on_empty = false
    rounding = 15
    
    placeholder_text = <i>Input password...</i>
    fail_text = <i>$FAIL <b>($ATTEMPTS)</b></i>
    
    dots_spacing = 0.3
    
    position = 0, -20
    halign = center
    valign = center
}

# Time display
label {
    monitor =
    text = $TIME
    font_size = 90
    colour = $tinct_foreground
    
    position = -30, 0
    halign = right
    valign = top
}

# Date display
label {
    monitor =
    text = cmd[update:60000] date +"%A, %d %B %Y"
    font_size = 25
    colour = $tinct_foreground_muted
    
    position = -30, -150
    halign = right
    valign = top
}

# User info
label {
    monitor =
    text = Hi, $USER
    font_size = 20
    colour = $tinct_accent2
    
    position = 0, 80
    halign = center
    valign = center
}
```

## Customization

### Custom Templates

Extract the default template for customization:

```bash
tinct plugins templates dump -o hyprlock -l ~/.config/tinct/templates/hyprlock
```

This creates `~/.config/tinct/templates/hyprlock/tinct.conf.tmpl` which you can modify.

### Template Variables

Available in the template:

```go
{{ .SourceTheme }}        // "dark" or "light"
{{ .Background }}         // "#1a1b26"
{{ .BackgroundMuted }}    // "#16161e"
{{ .Foreground }}         // "#c0caf5"
{{ .ForegroundMuted }}    // "#a9b1d6"
{{ .Accent1 }}            // "#7aa2f7"
{{ .Accent2 }}            // "#bb9af7"
{{ .Accent3 }}            // "#7dcfff"
{{ .Accent4 }}            // "#9ece6a"
{{ .Danger }}             // "#f7768e"
{{ .Warning }}            // "#e0af68"
{{ .Success }}            // "#9ece6a"
{{ .Info }}               // "#7aa2f7"
```

### Helper Methods

Convert colours to different formats:

```go
// RGB format (RRGGBB without #)
{{ .BackgroundRGB }}      // "1a1b26"

// RGBA format (R, G, B, A)
{{ .BackgroundRGBA "ee" }} // "26, 27, 38, 0.93"
{{ .Accent1RGBA "ff" }}    // "122, 162, 247, 1.00"
{{ .Accent1RGBA "80" }}    // "122, 162, 247, 0.50"
```

## Integration with Hyprland

### Lock with Hyprlock

Add to `~/.config/hypr/hyprland.conf`:

```conf
# Lock keybind
bind = $mainMod, L, exec, hyprlock

# Lock on idle (optional)
exec-once = swayidle -w timeout 300 'hyprlock' timeout 600 'hyprctl dispatch dpms off' resume 'hyprctl dispatch dpms on'
```

## Pre-Execute Check

The plugin performs the following checks before generating:

1. Verifies `hyprlock` is installed and on `$PATH`
2. Creates config directory if it doesn't exist

If `hyprlock` is not found, the plugin will be skipped with a message.

## Troubleshooting

### Colours Not Applied

Ensure the source path is correct:

```bash
# Check if file exists
ls -l ~/.config/hypr/tinct-hyprlock.conf

# Verify hyprlock can read it
hyprlock --config ~/.config/hypr/hyprlock.conf
```

### Hyprlock Not Found

If pre-execute check fails:

```bash
# Check if hyprlock is installed
which hyprlock

# Install hyprlock (Arch)
pacman -S hyprlock

# Or build from source
git clone https://github.com/hyprwm/hyprlock
cd hyprlock
make all
sudo make install
```

### Source Directive Not Working

Ensure you're using an absolute path or `~/` prefix:

```conf
# Correct
source = ~/.config/hypr/tinct-hyprlock.conf
source = /home/username/.config/hypr/tinct-hyprlock.conf

# Incorrect
source = tinct-hyprlock.conf
source = ./tinct-hyprlock.conf
```

### RGBA Colours Look Wrong

Hyprland's RGBA format uses:
- RGB values in decimal (0-255)
- Alpha as float (0.0-1.0)

If colours look incorrect, check that you're using the RGBA variables (ending in `_rgba`), not the RGB ones.

## Advanced Usage

### Gradients

Hyprlock supports colour gradients:

```conf
input-field {
    # Gradient from accent1 to accent2
    outer_color = $tinct_accent1_rgba $tinct_accent2_rgba 45deg
    check_color = $tinct_success_rgba $tinct_accent3_rgba 120deg
    fail_color = $tinct_danger_rgba $tinct_warning_rgba 40deg
}
```

### Dynamic Backgrounds

Use screenshot with colour overlay:

```conf
background {
    monitor =
    path = screenshot
    colour = $tinct_background_rgba  # Transparent overlay
    blur_passes = 3
}
```

### Multiple Labels

Create themed status displays:

```conf
# System info
label {
    monitor =
    text = cmd[update:5000] uname -r
    colour = $tinct_info
    font_size = 14
    position = 10, 10
    halign = left
    valign = bottom
}

# Weather (if you have a script)
label {
    monitor =
    text = cmd[update:600000] ~/.config/hypr/scripts/weather.sh
    colour = $tinct_accent3
    font_size = 16
    position = -10, 10
    halign = right
    valign = bottom
}
```

## Example Output

```conf
# Hyprlock colour theme generated by Tinct
# Detected theme: dark

# ============================================================================
# Tinct Colour Variables
# ============================================================================

$tinct_background = rgb(1a1b26)
$tinct_background_muted = rgb(16161e)
$tinct_foreground = rgb(c0caf5)
$tinct_foreground_muted = rgb(a9b1d6)

$tinct_accent1 = rgb(7aa2f7)
$tinct_accent2 = rgb(bb9af7)
$tinct_accent3 = rgb(7dcfff)
$tinct_accent4 = rgb(9ece6a)

$tinct_danger = rgb(f7768e)
$tinct_warning = rgb(e0af68)
$tinct_success = rgb(9ece6a)
$tinct_info = rgb(7aa2f7)

# ============================================================================
# RGBA Variants (with alpha channel)
# ============================================================================

$tinct_background_rgba = rgba(26, 27, 38, 0.93)
$tinct_background_muted_rgba = rgba(22, 22, 30, 0.80)
$tinct_foreground_rgba = rgba(192, 202, 245, 1.00)
$tinct_foreground_muted_rgba = rgba(169, 177, 214, 0.87)

$tinct_accent1_rgba = rgba(122, 162, 247, 1.00)
$tinct_accent2_rgba = rgba(187, 154, 247, 1.00)
$tinct_accent3_rgba = rgba(125, 207, 255, 1.00)
$tinct_accent4_rgba = rgba(158, 206, 106, 1.00)

$tinct_danger_rgba = rgba(247, 118, 142, 1.00)
$tinct_warning_rgba = rgba(224, 175, 104, 1.00)
$tinct_success_rgba = rgba(158, 206, 106, 1.00)
$tinct_info_rgba = rgba(122, 162, 247, 1.00)
```

## Resources

- [Hyprlock Documentation](https://wiki.hyprland.org/Hypr-Ecosystem/hyprlock/)
- [Hyprlock GitHub](https://github.com/hyprwm/hyprlock)
- [Hyprland Colours Documentation](https://wiki.hyprland.org/Configuring/Variables/#colours)
- [Example Configurations](https://github.com/hyprwm/hyprlock/blob/main/assets/example.conf)

## See Also

- [Hyprland Plugin](../hyprland/README.md) - Main Hyprland window manager theming
- [Waybar Plugin](../waybar/README.md) - Status bar theming
- [Plugin Development Guide](../../../../docs/PLUGIN-STANDARD.md)