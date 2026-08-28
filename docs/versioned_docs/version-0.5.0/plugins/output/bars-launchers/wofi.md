---
title: wofi
sidebar_position: 1
plugin:
  type: output
  name: wofi
  category: bars-launchers
  source: builtin
  app: Wofi
  app_url: 'https://hg.sr.ht/~scoopta/wofi'
  requires:
    - wofi
  optional: []
  pattern: two-file
  default_output_dir: ~/.config/wofi
  generated_files:
    - tinct-colors
    - style.css
  reload:
    method: none
    user_action_required: true
---

# wofi

Generates a colour theme for [Wofi](https://hg.sr.ht/~scoopta/wofi), the GTK-based application launcher / dmenu replacement for Wayland. The output is a numbered colour file (consumed via Wofi's `color=` directive) plus a GTK CSS stylesheet that references the palette by index.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o wofi` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o wofi
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colors` | `~/.config/wofi/tinct-colors` | **Palette.** Newline-separated list of hex colours. Wofi exposes each entry as `--wofi-color<n>` / `--wofi-rgb-color<n>` (one-based by line index). |
| `style.css` | `~/.config/wofi/style.css` | **Main stylesheet.** Complete Wofi GTK stylesheet — window, input, entry rows, selection states — referencing the palette via the `--wofi-color*` variables. |

## Generated format

### `tinct-colors`

```css
@define-color background #1e1e2e;
@define-color foreground #cdd6f4;
@define-color surface #313244;
@define-color accent1 #89b4fa;
@define-color accent2 #f5c2e7;
@define-color danger #f38ba8;
@define-color success #a6e3a1;
@define-color warning #f9e2af;
```

### `style.css`

```css
@import "tinct-colors";

window {
    margin: 0px;
    background-color: @background;
    border-radius: 8px;
    border: 2px solid @accent1;
}

#input {
    margin: 5px;
    background-color: @surface;
    color: @foreground;
    border: none;
    border-radius: 4px;
}

#entry:selected {
    background-color: @accent1;
    color: @background;
}
```

## Integration

Wofi loads two paths from your config (`~/.config/wofi/config`) or via command-line flags:

```ini
style=~/.config/wofi/style.css
color=~/.config/wofi/tinct-colors
```

Or invoke Wofi explicitly:

```bash
wofi --show drun --style ~/.config/wofi/style.css --color ~/.config/wofi/tinct-colors
```

The plugin writes `style.css` directly to Wofi's default config directory, so most users do not need any additional config — Wofi picks `style.css` up automatically and the `color=` line is the only line you may need to add (or pass via `--color`).

## Reload behaviour

### Automatic

None. Wofi has no daemon and no IPC reload — each invocation is a one-shot process that reads its stylesheet on startup.

### Manual fallback

Re-run Wofi. Any new invocation will read the regenerated files:

```bash
wofi --show drun
```

If a Wofi window is already on screen when you regenerate, dismiss it (Escape) and re-launch; the new theme appears on the next launch.

## Uninstall / revert

1. **Remove the config lines** from `~/.config/wofi/config` (only if you added them):

   ```bash
   sed -i '/tinct-colors\|tinct.*style\.css/d' ~/.config/wofi/config
   ```

   If you were passing `--style` / `--color` on the command line, drop those flags from your launcher script.

2. **Delete the generated files**:

   ```bash
   rm ~/.config/wofi/tinct-colors ~/.config/wofi/style.css
   ```

3. **Reload to drop the theme**: nothing to reload — the next `wofi` invocation will start without a custom stylesheet (or with whatever you point `style=` at).

4. **External state**: this plugin only writes to `~/.config/wofi/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--wofi.output-dir` | `~/.config/wofi` | Override the output directory |

## Colour role mapping

Wofi indexes the palette by line number in `tinct-colors`. The default template emits all 49 semantic roles via `allColors`; the stylesheet uses the first 12 by convention:

| Variable | Tinct role | Role |
|---|---|---|
| `--wofi-color0` | `background` | Window background |
| `--wofi-color1` | `backgroundMuted` | Input / hover backgrounds |
| `--wofi-color2` | `foreground` | Default text |
| `--wofi-color3` | `foregroundMuted` | Unselected entry text |
| `--wofi-color4` | `accent1` | Border, selection indicator |
| `--wofi-color5` | `accent2` | Reserved for user customisation |
| `--wofi-color6` | `accent3` | Reserved for user customisation |
| `--wofi-color7` | `accent4` | Reserved for user customisation |
| `--wofi-color8` | `danger` | Error states |
| `--wofi-color9` | `warning` | Warning states |
| `--wofi-color10` | `success` | Success states |
| `--wofi-color11` | `info` | Informational states |

Wofi also exposes `--wofi-rgb-color<n>` (comma-separated R,G,B) for use inside `rgba()` so the stylesheet can apply per-state alpha.

## Customising the template

Extract the default templates to override them:

```bash
tinct plugins templates dump -o wofi -l ~/.config/tinct/templates/wofi
```

This creates `~/.config/tinct/templates/wofi/tinct-colors` and `~/.config/tinct/templates/wofi/style.css.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Theme not picked up

Wofi only auto-loads `style.css` from its config dir; the colour file must be referenced explicitly. Add `color=~/.config/wofi/tinct-colors` to `~/.config/wofi/config`, or pass `--color ~/.config/wofi/tinct-colors` on the command line.

### Colours partially applied

If you maintain your own Wofi `style.css`, the plugin's `style.css` write will overwrite it. Back up your custom stylesheet, or point `--wofi.output-dir` somewhere else and `@import` the palette from your own stylesheet.

### `--wofi-color<n>` shows as a literal string

GTK's CSS parser silently emits the literal text when a custom property is undefined. Confirm Wofi is being invoked with `--color` (or `color=` in `~/.config/wofi/config`) pointing at the generated file.

## Related plugins

- [walker](./walker.md) — alternative Wayland launcher with similar GTK CSS theming.
- [fuzzel](./fuzzel.md) — non-GTK launcher for users who prefer a single-file config.
- [rofi](./rofi.md) — X11/XWayland launcher with its own colour file format.
