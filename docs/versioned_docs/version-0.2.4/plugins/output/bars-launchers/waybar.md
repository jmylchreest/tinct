---
sidebar_position: 1
---

# waybar

Generate colour themes for [Waybar](https://github.com/Alexays/Waybar) status bar.

## Description

Waybar is a highly customizable status bar for Sway and wlroots-based compositors. The plugin generates CSS files with colour variable definitions and an optional styling stub.

## Output paths

```
~/.config/waybar/themes/tinct.css
~/.config/waybar/tinct-style.css
```

The plugin generates:
- `tinct.css` - CSS colour variable definitions
- `tinct-style.css` - Example styling stub (optional)

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o waybar
```

### Without auto-reload

```bash
tinct generate -i image -p ~/wallpaper.jpg -o waybar --waybar.reload=false
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--waybar.output-dir` | `~/.config/waybar/themes` | Output directory for theme files |
| `--waybar.generate-stub` | `true` | Generate example CSS stub |
| `--waybar.reload` | `true` | Reload Waybar after generation (SIGUSR2) |

## Configuration

Import the generated colours in your Waybar `style.css`:

```css
@import "themes/tinct.css";

/* Use the tinct colour variables */
* {
    color: var(--foreground);
    background-color: var(--background);
}

#workspaces button.focused {
    background-color: var(--accent1);
}

#battery.warning {
    color: var(--warning);
}

#battery.critical {
    color: var(--danger);
}
```

Or import the full stub file:

```css
@import "tinct-style.css";
```

## Automatic reload

When `--waybar.reload` is enabled (default), tinct sends SIGUSR2 to all running Waybar instances, triggering a live CSS reload.

## CSS variables

The generated `tinct.css` defines:

```css
:root {
    --background: #1e1e2e;
    --foreground: #cdd6f4;
    --surface: #313244;
    --accent1: #89b4fa;
    --accent2: #f5c2e7;
    --danger: #f38ba8;
    --success: #a6e3a1;
    --warning: #f9e2af;
    --info: #94e2d5;
    /* ... additional semantic colours */
}
```

## See also

- [Waybar documentation](https://github.com/Alexays/Waybar/wiki)
- [dunst](/docs/plugins/output/bars-launchers/dunst)
- [hyprland](/docs/plugins/output/hyprland)
