---
title: waybar
sidebar_position: 1

plugin:
  type: output
  name: waybar
  category: bars-launchers
  source: builtin
  app: Waybar
  app_url: https://github.com/Alexays/Waybar
  requires: [waybar]
  optional: []
  pattern: two-file
  default_output_dir: ~/.config/waybar/themes
  generated_files: [tinct.css, ../tinct-style.css]
  reload:
    method: signal
    command: pkill -SIGUSR2 -x waybar
    user_action_required: false
---

# waybar

Generates a colour theme for [Waybar](https://github.com/Alexays/Waybar), the highly customisable status bar for Wayland compositors. The plugin emits a GTK CSS palette file plus an example stylesheet that shows how to consume the palette, and (by default) sends `SIGUSR2` to running Waybar instances so the new theme takes effect without a restart.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o waybar` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o waybar
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.css` | `~/.config/waybar/themes/tinct.css` | **Palette.** A GTK `@define-color` block exposing all 49 tinct semantic roles as `@background`, `@accent1`, `@on-danger`, etc. Import this from your own `style.css`. |
| `tinct-style.css` | `~/.config/waybar/tinct-style.css` | **Example main config.** A complete reference `style.css` showing how to wire the palette into Waybar's standard modules (workspaces, clock, battery, tray). Drop in as `style.css` if you don't already have one. |

The example stylesheet is only generated when `--waybar.generate-stub=true` (the default). Disable it once you have your own `style.css` set up — the palette file (`tinct.css`) is the only file you actually need.

## Integration

Waybar uses a single `style.css` and resolves `@import` paths relative to it. Add this line to the top of `~/.config/waybar/style.css`:

```css
@import "themes/tinct.css";
```

Then reference the colour variables anywhere in your stylesheet:

```css
window#waybar {
    background: @background;
    color: @foreground;
}

#workspaces button.active {
    background: @accent1;
    color: @on-accent1;
}
```

If you don't already have a `style.css`, copy the generated example:

```bash
cp ~/.config/waybar/tinct-style.css ~/.config/waybar/style.css
```

## Reload behaviour

### Automatic

After writing the files, the plugin sends `SIGUSR2` to every running `waybar` process. Waybar treats `SIGUSR2` as a "reload config and stylesheet" signal — bars rebuild in-place without losing their position on screen.

This is the equivalent of running:

```bash
pkill -SIGUSR2 waybar
```

Disable with `--waybar.reload=false` if you want to manage reloads yourself (for example, when scripting many generations in a row).

### Manual fallback

If Waybar wasn't running at generate time, or the signal didn't take effect:

```bash
pkill -SIGUSR2 waybar
```

Or, in a Hyprland session, simply restart the bar:

```bash
pkill -x waybar && waybar &
```

## Uninstall / revert

1. **Remove the include line** from your `style.css`:

   ```bash
   sed -i '/@import "themes\/tinct\.css"/d' ~/.config/waybar/style.css
   ```

2. **Delete the generated files**:

   ```bash
   rm ~/.config/waybar/themes/tinct.css ~/.config/waybar/tinct-style.css
   ```

3. **Reload to drop the theme**:

   ```bash
   pkill -SIGUSR2 waybar
   ```

4. **External state**: this plugin only writes to `~/.config/waybar/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--waybar.output-dir` | `~/.config/waybar/themes` | Override the output directory for the palette file |
| `--waybar.generate-stub` | `true` | Also emit `../tinct-style.css` as a reference `style.css` |
| `--waybar.reload` | `true` | Send `SIGUSR2` to running Waybar instances after generation |

## Colour role mapping

The palette exposes every tinct semantic role as a GTK colour variable. Names are lowercase with hyphens (`@on-accent1`, not `@onAccent1`).

| Group | Variables | Tinct roles |
|---|---|---|
| Core | `@background`, `@background-muted`, `@foreground`, `@foreground-muted` | The four base palette colours |
| Accents | `@accent1` … `@accent4` (+ `-muted` variants) | Primary palette accents |
| On-accents | `@on-accent1` … `@on-accent4` | Foreground colours for text drawn on each accent |
| Semantic | `@danger`, `@warning`, `@success`, `@info`, `@notification` | Status colours |
| On-semantic | `@on-danger`, `@on-warning`, `@on-success`, `@on-info` | Foregrounds for semantic backgrounds |
| Surface | `@surface`, `@on-surface`, `@outline`, `@border`, plus `-variant`/`-muted` | Module backgrounds and borders |
| Container elevation | `@surface-container-lowest` … `@surface-container-highest` | Material-style elevation tiers |
| Inverse | `@inverse-surface`, `@inverse-on-surface`, `@inverse-primary` | Used for tooltips / inverted sections |
| Effects | `@scrim`, `@shadow` | Alpha baked in |

Use GTK's `alpha(@colour, 0.8)` function for translucency.

## Customising the template

Extract the default templates to override them:

```bash
tinct plugins templates dump -o waybar -l ~/.config/tinct/templates/waybar
```

This creates `~/.config/tinct/templates/waybar/tinct-colours.css.tmpl` (the palette) and `~/.config/tinct/templates/waybar/tinct.css.tmpl` (the example `style.css`). Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Waybar didn't reload after generation

`SIGUSR2` is sent to all processes named `waybar`. If your bar is wrapped (for example by a launcher script with a different process name) or runs in a sandbox where signals can't reach it, the auto-reload no-ops. Fall back to `pkill -x waybar && waybar &` from your normal shell.

### `@background` reported as undefined

The `@import` path is resolved relative to `style.css`. If you moved either file, GTK silently treats the variables as undefined and Waybar logs a CSS warning at startup. Check `journalctl --user -u waybar -e` (or wherever your Waybar logs go) and adjust the `@import` path.

### Theme looks fine but workspace buttons are unchanged

Most Waybar configs override workspace button styling explicitly. Search your `style.css` for `#workspaces button` and either delete the conflicting block or update it to use `@accent1`/`@on-accent1`.

### Tinct's example `style.css` overrides my customisations

`tinct-style.css` is written next to `style.css`, not on top of it — it only matters if you copy or symlink it as `style.css`. Pass `--waybar.generate-stub=false` to stop emitting it once you have your own.

## Related plugins

- [hyprland](../hyprland/README.md) — compositor colour exports often configured alongside Waybar.
- [hyprlock](../hyprlock/README.md) — lock screen styling for Hyprland users.
- [dunst](../dunst/README.md) — notification daemon, commonly themed with Waybar.
