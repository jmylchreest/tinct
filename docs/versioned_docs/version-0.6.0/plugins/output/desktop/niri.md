---
title: niri
sidebar_position: 1
plugin:
  type: output
  name: niri
  category: desktop
  source: builtin
  app: niri
  app_url: 'https://github.com/niri-wm/niri'
  requires: []
  optional:
    - niri
  pattern: two-file
  default_output_dir: ~/.config/niri
  generated_files:
    - tinct-colours.kdl
    - tinct.kdl
  reload:
    method: none
    user_action_required: false
---

# niri

Generates a themed colour configuration for [niri](https://github.com/niri-wm/niri), the scrollable-tiling Wayland compositor. The plugin emits a KDL include file that sets niri's themeable colour properties (`focus-ring`, `border`, `insert-hint`, `tab-indicator`, `shadow`, and the overview `backdrop-color`), plus an example file documenting how to wire it in.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o niri` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o niri
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colours.kdl` | `~/.config/niri/tinct-colours.kdl` | **Theme.** Sets niri's colour properties directly (`layout.focus-ring`, `layout.border`, `layout.tab-indicator`, `layout.insert-hint`, `layout.shadow`, `overview.backdrop-color`). Include this from `config.kdl`. |
| `tinct.kdl` | `~/.config/niri/tinct.kdl` | **Reference.** A read-only, commented copy of what gets applied, plus the list of available colour roles. Optional — disable with `--niri.generate-stub=false`. |

## Integration

Unlike Hyprland's `$variable` model, niri has no config variables. Instead it supports **includes that merge per-property**, so the colour file contains the actual themed config sections and overrides only the properties it sets.

Add one line near the **end** of your `~/.config/niri/config.kdl` — after your own `layout` block, so the merge takes precedence:

```kdl
include "tinct-colours.kdl"
```

niri includes are positional and merge per-property: only the properties set in the included file change; everything else in your config is preserved. Included files are **watched**, so niri live-reloads automatically whenever tinct rewrites the colour file — there is no reload step.

## Reload behaviour

### Automatic

niri watches `config.kdl` and every included file. When tinct regenerates `tinct-colours.kdl`, niri reloads the configuration automatically and applies the new colours to open windows. No action required.

### Manual fallback

If your config has an error and niri rejects the reload, validate it:

```bash
niri validate
```

## Uninstall / revert

1. **Remove the include line** from your `~/.config/niri/config.kdl`:

   ```bash
   sed -i '/include "tinct-colours\.kdl"/d' ~/.config/niri/config.kdl
   ```

2. **Delete the generated files**:

   ```bash
   rm ~/.config/niri/tinct-colours.kdl ~/.config/niri/tinct.kdl
   ```

   niri reloads automatically and reverts to your own colours.

3. **External state**: this plugin only writes into `~/.config/niri/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--niri.output-dir` | `~/.config/niri` | Override the output directory |
| `--niri.generate-stub` | `true` | Generate the `tinct.kdl` reference alongside the theme |
| `--niri.stub-path` | _(unset)_ | Custom filename for the stub (basename only; written into the output dir) |

## Colour role mapping

| niri section | Property | Tinct role |
|---|---|---|
| `layout.focus-ring` | `active-gradient` | `accent1` → `accent2` (45°) |
| `layout.focus-ring` | `inactive-color` | `backgroundMuted` |
| `layout.border` | `active-color` | `accent1` |
| `layout.border` | `inactive-color` | `backgroundMuted` |
| `layout.insert-hint` | `color` | `accent2` |
| `layout.tab-indicator` | `active-color` | `accent1` |
| `layout.tab-indicator` | `inactive-color` | `backgroundMuted` |
| `layout.shadow` | `color` | `shadow` (with alpha) |
| `overview` | `backdrop-color` | `scrim` |

## Customising the template

Extract the default templates to override them:

```bash
tinct plugins templates dump -o niri -l ~/.config/tinct/templates/niri
```

This creates `~/.config/tinct/templates/niri/tinct-colours.kdl.tmpl` and `tinct.kdl.tmpl`. Tinct uses your versions in preference to the embedded defaults.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Colours not changing

The include line must come **after** your own `layout` block in `config.kdl`. niri merges includes positionally — an include placed before your `layout` block is overridden by it. Move `include "tinct-colours.kdl"` to the end of the file.

### `focus-ring` colour ignored

niri's `focus-ring` uses either a solid colour **or** a gradient, and a gradient wins when both are present. The theme sets `active-gradient`; if you have an `active-color` defined *after* the include it takes effect instead. Remove the conflicting property or reorder the include.

### Config rejected after include

Run `niri validate` to see the parse error. A common cause is including the file at the wrong nesting level — `include` must be at the **top level** of `config.kdl`, not inside a `layout { }` block.

## Related plugins

- [waybar](../bars-launchers/waybar.md) — status bar commonly paired with niri.
- [fuzzel](../bars-launchers/fuzzel.md) — application launcher commonly paired with niri.
- [swayosd](../bars-launchers/swayosd.md) — on-screen display for volume/brightness.
- [hyprland](../hyprland/hyprland.md) — the other tiling Wayland compositor plugin.
