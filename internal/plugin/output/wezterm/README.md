---
title: wezterm
sidebar_position: 6

plugin:
  type: output
  name: wezterm
  category: terminals
  source: builtin
  app: WezTerm
  app_url: https://wezfurlong.org/wezterm/
  requires: []
  optional: [wezterm]
  pattern: single-file
  default_output_dir: ~/.config/wezterm/colors
  generated_files: [tinct.toml]
  reload:
    method: none
    user_action_required: false
---

# wezterm

Generates a TOML colour scheme for [WezTerm](https://wezfurlong.org/wezterm/), Wez Furlong's GPU-accelerated, Lua-configurable terminal. The file lives in WezTerm's `colors/` directory where it's auto-discovered as a named scheme; you activate it with one line of Lua.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o wezterm` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o wezterm
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.toml` | `~/.config/wezterm/colors/tinct.toml` | Single-file colour scheme — `[colors]` (fg/bg, cursor, selection, scrollbar, split, visual bell, compose cursor), `ansi = [...]`, `brights = [...]`, and the `[colors.tab_bar.*]` blocks for the tab strip. |

Override the filename with `--wezterm.theme-name=<name>` (the file becomes `<name>.toml` and you activate it with `config.color_scheme = "<name>"`).

## Integration

WezTerm scans `~/.config/wezterm/colors/` automatically. Activate the scheme from your `wezterm.lua`:

```lua
config.color_scheme = "tinct"
```

If you don't already have a `wezterm.lua`, this minimal config is enough:

```lua
local wezterm = require 'wezterm'
local config = wezterm.config_builder()
config.color_scheme = "tinct"
return config
```

## Reload behaviour

### Automatic

None from the plugin — WezTerm watches its config files (including everything in `colors/`) and reloads automatically. As soon as `tinct generate` finishes writing `tinct.toml`, every open WezTerm window re-renders with the new colours; you don't need to send any signal.

### Manual fallback

If you've set `automatically_reload_config = false`, reload by hand:

- In-app: default key chord `Ctrl+Shift+R` (configurable), or `Cmd+Shift+R` on macOS.
- Restart WezTerm.

## Uninstall / revert

1. **Remove the config line** from your `wezterm.lua`:

   Delete `config.color_scheme = "tinct"` (or change it back to your previous scheme name).

2. **Delete the generated file**:

   ```bash
   rm ~/.config/wezterm/colors/tinct.toml
   ```

3. **Reload to drop the theme**:

   Saving `wezterm.lua` triggers WezTerm's auto-reload. Otherwise press `Ctrl+Shift+R`.

4. **External state**: this plugin only writes to `~/.config/wezterm/colors/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--wezterm.output-dir` | `~/.config/wezterm/colors` | Override the output directory |
| `--wezterm.theme-name` | `tinct` | Theme name — sets both the filename (`<name>.toml`) and the value to use in `config.color_scheme = "<name>"` |

## Colour role mapping

### `[colors]`

| WezTerm key | Tinct role |
|---|---|
| `foreground` | `foreground` |
| `background` | `background` |
| `cursor_bg` / `cursor_border` | `accent1` |
| `cursor_fg` | `background` |
| `selection_fg` | `background` |
| `selection_bg` | `accent1` |
| `scrollbar_thumb` | `outline` |
| `split` | `outlineVariant` |
| `visual_bell` | `warning` |
| `compose_cursor` | `accent2` |
| `ansi[0..7]` | ANSI `black`..`white` |
| `brights[0..7]` | ANSI `brightblack`..`brightwhite` |

### `[colors.tab_bar.*]`

| WezTerm key | Tinct role |
|---|---|
| `background` | `surfaceContainerLow` |
| `active_tab.bg_color` / `fg_color` | `accent1` / `onAccent1` (bold) |
| `inactive_tab.bg_color` / `fg_color` | `surfaceContainer` / `foregroundMuted` |
| `inactive_tab_hover.bg_color` / `fg_color` | `surfaceContainerHigh` / `foreground` |
| `new_tab.bg_color` / `fg_color` | `surfaceContainerLow` / `foregroundMuted` |
| `new_tab_hover.bg_color` / `fg_color` | `surfaceContainer` / `foreground` |

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o wezterm -l ~/.config/tinct/templates/wezterm
```

This writes `~/.config/tinct/templates/wezterm/tinct.toml.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### "color scheme not found: tinct"

WezTerm searches `colors/` next to the config file it loaded. Confirm the path:

```bash
ls -l ~/.config/wezterm/colors/tinct.toml
```

If the file isn't there, the plugin's output directory is somewhere else — re-check `--wezterm.output-dir` and `TINCT_PLUGIN_WEZTERM_OUTPUT_DIR`. On Windows without `$XDG_CONFIG_HOME`, the default is `%APPDATA%/wezterm/colors`.

### Tab bar still looks default

Tab-bar colours only apply when WezTerm's native tab bar is enabled (`use_fancy_tab_bar = true` is default; the fancy bar uses its own colour pipeline). If you set `enable_tab_bar = false`, the `[colors.tab_bar.*]` block is ignored — that's WezTerm's behaviour, not the plugin's.

### Edits to `tinct.toml` get overwritten

Don't edit `tinct.toml` directly — every `tinct generate -o wezterm` clobbers it. Customise the template instead (see above) so your edits are reapplied each run.

## Related plugins

- [alacritty](../alacritty/README.md), [ghostty](../ghostty/README.md), [kitty](../kitty/README.md) — other GPU-accelerated terminals.
- [tmux](../tmux/README.md) — pair the terminal palette with a matching multiplexer status bar.
