---
sidebar_position: 1
---

# hyprland

Generate colour themes for the [Hyprland](https://hyprland.org/) compositor.

## Description

Hyprland is a dynamic tiling Wayland compositor. The plugin generates configuration files that define colour variables for window borders, decorations, and other visual elements.

## Output paths

```
~/.config/hypr/themes/tinct-colours.conf
~/.config/hypr/themes/tinct.conf
```

The plugin generates two files:
- `tinct-colours.conf` - Colour variable definitions
- `tinct.conf` - Example stub showing how to use the colours

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o hyprland
```

### With automatic config reload

```bash
tinct generate -i image -p ~/wallpaper.jpg -o hyprland --hyprland.reload
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--hyprland.output-dir` | `~/.config/hypr/themes` | Output directory for theme files |
| `--hyprland.generate-stub` | `true` | Generate example configuration stub |
| `--hyprland.stub-path` | `""` | Custom path for stub file |
| `--hyprland.reload` | `false` | Reload Hyprland config after generation |

## Configuration

Include the generated colours in your Hyprland configuration.

Add to your `~/.config/hypr/hyprland.conf`:

```bash
source = ~/.config/hypr/themes/tinct-colours.conf
```

Then use the colour variables in your configuration:

```bash
general {
    col.active_border = $accent1
    col.inactive_border = $surface
}

decoration {
    col.shadow = $shadow
}
```

## Versioned templates

Tinct detects your installed Hyprland version and uses appropriate templates. This ensures compatibility as Hyprland's configuration syntax evolves.

Version detection uses `hyprctl version` or `hyprland --version`.

## Colour variables

The generated configuration defines these variables:

| Variable | Description |
|----------|-------------|
| `$background` | Primary background colour |
| `$foreground` | Primary text colour |
| `$surface` | Surface container colour |
| `$accent1` | Primary accent colour |
| `$accent2` | Secondary accent colour |
| `$danger` | Error/warning colour |
| `$success` | Success indicator colour |
| `$warning` | Warning indicator colour |
| `$info` | Information colour |
| `$shadow` | Shadow colour with alpha |

## Example stub

The generated `tinct.conf` shows how to apply colours to common elements:

```bash
# Tinct Theme Configuration
# Include this file in your hyprland.conf

general {
    col.active_border = $accent1 $accent2 45deg
    col.inactive_border = $surface
}

decoration {
    col.shadow = rgba(1e1e2eaa)
}

group {
    col.border_active = $accent1
    col.border_inactive = $surface
}
```

## See also

- [Hyprland documentation](https://wiki.hyprland.org/)
- [hyprlock](/docs/plugins/output/hyprland/hyprlock)
- [hyprpaper](/docs/plugins/output/hyprland/hyprpaper)
