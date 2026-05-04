---
sidebar_position: 1
---

# alacritty

Generate colour themes for the [Alacritty](https://alacritty.org/) terminal emulator.

## Description

Alacritty is a modern, GPU-accelerated terminal emulator. The plugin generates a TOML colour scheme file.

## Output path

```
~/.config/alacritty/tinct-colors.toml
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o alacritty
```

## Configuration

Import the generated file in your `alacritty.toml`:

```toml
import = ["~/.config/alacritty/tinct-colors.toml"]
```

## Generated output

```toml
[colors]
[colors.primary]
background = "#1e1e2e"
foreground = "#cdd6f4"

[colors.normal]
black = "#45475a"
red = "#f38ba8"
green = "#a6e3a1"
yellow = "#f9e2af"
blue = "#89b4fa"
magenta = "#f5c2e7"
cyan = "#94e2d5"
white = "#bac2de"

[colors.bright]
black = "#585b70"
red = "#f38ba8"
green = "#a6e3a1"
yellow = "#f9e2af"
blue = "#89b4fa"
magenta = "#f5c2e7"
cyan = "#94e2d5"
white = "#a6adc8"
```

## ANSI colour mapping

Tinct maps semantic colours to ANSI positions:

| ANSI | Normal | Bright |
|------|--------|--------|
| 0 (black) | backgroundMuted | surface |
| 1 (red) | danger | danger |
| 2 (green) | success | success |
| 3 (yellow) | warning | warning |
| 4 (blue) | accent1 | accent1 |
| 5 (magenta) | accent2 | accent2 |
| 6 (cyan) | info | info |
| 7 (white) | foregroundMuted | foreground |

## See also

- [Alacritty documentation](https://alacritty.org/config-alacritty.html)
- [ghostty](/docs/plugins/output/terminals/ghostty)
- [kitty](/docs/plugins/output/terminals/kitty)
