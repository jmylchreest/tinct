---
sidebar_position: 2
---

# dunst

Generate colour themes for [Dunst](https://dunst-project.org/) notification daemon.

## Description

Dunst is a lightweight notification daemon for Linux. The plugin generates a configuration snippet that styles notification popups with your theme colours.

## Output path

```
~/.config/dunst/dunstrc.d/60-tinct.conf
```

The file is placed in the `dunstrc.d` directory with a numbered prefix to ensure proper configuration ordering.

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o dunst
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dunst.output-dir` | `~/.config/dunst/dunstrc.d` | Output directory for theme files |

## Configuration

The plugin uses Dunst's drop-in configuration directory (`dunstrc.d`). Files in this directory are loaded alphabetically after the main `dunstrc`.

No manual configuration is required - the generated file is automatically loaded.

## Automatic reload

After generation, tinct runs `dunstctl reload` to apply the new theme immediately. If reload fails, you can restart dunst manually:

```bash
killall dunst && dunst &
```

## Themed elements

The theme affects:

- Notification background colours
- Text colours
- Border colours
- Urgency-specific styling (low, normal, critical)

## Generated format

```ini
[global]
frame_color = "#89b4fa"
separator_color = "#313244"

[urgency_low]
background = "#1e1e2e"
foreground = "#cdd6f4"
frame_color = "#89b4fa"

[urgency_normal]
background = "#1e1e2e"
foreground = "#cdd6f4"
frame_color = "#89b4fa"

[urgency_critical]
background = "#1e1e2e"
foreground = "#cdd6f4"
frame_color = "#f38ba8"
```

## See also

- [Dunst documentation](https://dunst-project.org/documentation/)
- [histui](/docs/plugins/output/special/histui)
- [waybar](/docs/plugins/output/bars-launchers/waybar)
