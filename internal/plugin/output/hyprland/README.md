---
title: hyprland
sidebar_position: 1

plugin:
  type: output
  name: hyprland
  category: hyprland
  source: builtin
  app: Hyprland
  app_url: https://hyprland.org/
  requires: []
  optional: [hyprctl]
  pattern: two-file
  default_output_dir: ~/.config/hypr/themes
  generated_files: [tinct-colours.conf, tinct.conf]
  reload:
    method: none
    user_action_required: true
---

# hyprland

Generates colour variables for [Hyprland](https://hyprland.org/), the dynamic tiling Wayland compositor. The plugin emits a palette file (`$variable` declarations consumable from any `hyprland.conf`) plus an example config that shows how to apply those variables to common Hyprland sections (`general`, `decoration`, `group`, window rules).

## Installation

Built into tinct — nothing to install separately. `tinct generate -o hyprland` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o hyprland
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colours.conf` | `~/.config/hypr/themes/tinct-colours.conf` | **Palette.** Defines every tinct colour role as a Hyprland `$variable` (e.g. `$accent1RGB`, `$backgroundRGBDec`). Source this from `hyprland.conf`. |
| `tinct.conf` | `~/.config/hypr/themes/tinct.conf` | **Example config.** Demonstrates wiring the palette into `general` (borders), `decoration` (shadow), `group` (tab bar) and `windowrule` lines. Optional — disable with `--hyprland.generate-stub=false`. |

The plugin auto-detects the installed Hyprland version (`hyprctl version` or `hyprland --version`) and selects a version-specific template if one is bundled (currently `0.53` is the only versioned variant).

## Integration

Hyprland uses an include-based config. Add one line to your `~/.config/hypr/hyprland.conf` to load the palette:

```ini
source = ~/.config/hypr/themes/tinct-colours.conf
```

After sourcing, every `$<role>RGB` and `$<role>RGBDec` variable is available everywhere in your Hyprland config:

```ini
general {
    col.active_border = $accent1RGB $accent2RGB 45deg
    col.inactive_border = $backgroundMutedRGB
}
```

The generated `tinct.conf` is an **example**, not something you have to use. Copy the parts you want into your own config, or `source = ~/.config/hypr/themes/tinct.conf` if you want the whole reference layout applied directly.

## Reload behaviour

### Automatic

None by default. The plugin writes the files and exits without touching Hyprland.

Pass `--hyprland.reload` to make the plugin run `hyprctl reload` after generation. `hyprctl` is listed as an optional binary; if it isn't on `$PATH` the reload is skipped.

### Manual fallback

Trigger a Hyprland config reload by hand:

```bash
hyprctl reload
```

Hyprland reloads atomically — open windows keep their state, only the colours change.

## Uninstall / revert

1. **Remove the include line** from your `~/.config/hypr/hyprland.conf`:

   ```bash
   sed -i '/themes\/tinct-colours\.conf/d;/themes\/tinct\.conf/d' ~/.config/hypr/hyprland.conf
   ```

2. **Delete the generated files**:

   ```bash
   rm ~/.config/hypr/themes/tinct-colours.conf ~/.config/hypr/themes/tinct.conf
   ```

3. **Reload to drop the theme**:

   ```bash
   hyprctl reload
   ```

4. **External state**: this plugin only writes into `~/.config/hypr/themes/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--hyprland.output-dir` | `~/.config/hypr/themes` | Override the output directory |
| `--hyprland.generate-stub` | `true` | Generate the `tinct.conf` example alongside the palette |
| `--hyprland.stub-path` | _(unset)_ | Custom filename for the stub (basename only; written into the output dir) |
| `--hyprland.reload` | `false` | Run `hyprctl reload` after writing the files |

## Colour role mapping

The palette file declares every tinct role as a Hyprland `$variable`; the example `tinct.conf` wires a subset of them into common sections.

| Hyprland section | Property | Tinct role |
|---|---|---|
| `general` | `col.active_border` | `accent1` → `accent2` gradient |
| `general` | `col.inactive_border` | `backgroundMuted` |
| `decoration.shadow` | `color` | `background` @ 0.8 alpha |
| `misc` | `background_color` | `background` |
| `group` | `col.border_active` | `accent2` |
| `group` | `col.border_inactive` | `backgroundMuted` |
| `group.groupbar` | `col.active` | `accent1` |
| `group.groupbar` | `col.inactive` | `backgroundMuted` |
| `group.groupbar` | `text_color` | `foreground` |
| `windowrule bordercolor` | `title:.*[Ee]rror.*` | `danger` |
| `windowrule bordercolor` | `title:.*[Ww]arning.*` | `warning` |
| `windowrule bordercolor` | `title:.*[Ss]uccess.*` | `success` |

Every role tinct knows about is declared in the palette file in two formats — `$<role>RGB` for direct `rgb(...)` use, and `$<role>RGBDec` for `rgba(..., 0.50)` opacity expressions.

## Customising the template

Extract the default templates to override them:

```bash
tinct plugins templates dump -o hyprland -l ~/.config/tinct/templates/hyprland
```

This creates `~/.config/tinct/templates/hyprland/tinct-colours.conf.tmpl` and `tinct.conf.tmpl`. Tinct uses your versions in preference to the embedded defaults.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### `unknown variable: $accent1RGB`

You sourced `tinct.conf` (the example) but not `tinct-colours.conf` (the palette), or you sourced them in the wrong order. The palette must be sourced **before** any line that references its variables.

### Theme files written but colours unchanged

The plugin doesn't reload Hyprland by default. Run `hyprctl reload`, or pass `--hyprland.reload` next time.

### `hyprctl: command not found`

`hyprctl` ships with Hyprland; if you're missing it you're not actually running Hyprland or your install is broken. Generation still succeeds — the file is written either way — but `--hyprland.reload` becomes a no-op.

### Hyprland version not detected

The plugin reads the version from `hyprctl version` (running session) or `hyprland --version` (binary on PATH). If neither succeeds, tinct falls back to the default (unversioned) template. Override with custom templates if you need a specific layout.

## Related plugins

- [hyprlock](../hyprlock/README.md) — lock screen styling for Hyprland.
- [hyprpaper](../hyprpaper/README.md) — wallpaper manager for Hyprland.
- [waybar](../waybar/README.md) — status bar commonly paired with Hyprland.
- [dunst](../dunst/README.md) — notification daemon styling.
