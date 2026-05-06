---
sidebar_position: 2
---

# ghostty

Generate colour themes for the [Ghostty](https://ghostty.org/) terminal emulator.

## Description

Ghostty is a fast, feature-rich terminal emulator. The plugin generates a configuration file with colour palette.

## Output path

```
~/.config/ghostty/tinct-theme
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o ghostty
```

## Configuration

Reference the theme in your Ghostty config:

```
theme = tinct-theme
```

Or use include:

```
config-file = ~/.config/ghostty/tinct-theme
```

## Generated output

```
background = #1e1e2e
foreground = #cdd6f4

palette = 0=#45475a
palette = 1=#f38ba8
palette = 2=#a6e3a1
palette = 3=#f9e2af
palette = 4=#89b4fa
palette = 5=#f5c2e7
palette = 6=#94e2d5
palette = 7=#bac2de
palette = 8=#585b70
palette = 9=#f38ba8
palette = 10=#a6e3a1
palette = 11=#f9e2af
palette = 12=#89b4fa
palette = 13=#f5c2e7
palette = 14=#94e2d5
palette = 15=#a6adc8
```

## See also

- [Ghostty documentation](https://ghostty.org/docs)
- [alacritty](./alacritty.md)
- [kitty](./kitty.md)
