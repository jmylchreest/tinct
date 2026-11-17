---
title: helix
sidebar_position: 5
plugin:
  type: output
  name: helix
  category: editors
  source: builtin
  app: Helix
  app_url: 'https://helix-editor.com/'
  requires:
    - hx
  optional: []
  pattern: two-file
  default_output_dir: ~/.config/helix/themes
  generated_files:
    - tinct-palette.toml
    - tinct.toml
  reload:
    method: none
    user_action_required: true
---

# helix

Generates a TOML theme for [Helix](https://helix-editor.com/), the modal post-modern text editor written in Rust. The output is split into a palette file (named colours) and a theme file that `inherits =` the palette and binds UI / syntax keys to those names — so future runs rewrite the palette without touching key bindings the user may have edited.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o helix` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o helix
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-palette.toml` | `~/.config/helix/themes/tinct-palette.toml` | **Palette** — `[palette]` block only. Defines named colours (`background`, `accent1`, `danger`, …). Tinct rewrites this on every run. |
| `tinct.toml` | `~/.config/helix/themes/tinct.toml` | **Theme** — `inherits = "tinct-palette"` plus all `ui.*`, `diagnostic.*`, syntax (`keyword`, `function`, `string`, …), and markup key bindings. Safe to fork / edit by hand. |

## Generated format

`tinct-palette.toml`:

```toml
[palette]
background       = "#1a1b26"
backgroundMuted  = "#16161e"
foreground       = "#c0caf5"
foregroundMuted  = "#737aa2"
accent1          = "#7aa2f7"
accent2          = "#bb9af7"
accent3          = "#7dcfff"
accent4          = "#e0af68"
danger           = "#f7768e"
warning          = "#e0af68"
success          = "#9ece6a"
info             = "#7dcfff"
# … plus on-* variants, surface / outline tones
```

`tinct.toml`:

```toml
inherits = "tinct-palette"

"ui.background"   = { bg = "background" }
"ui.text"         = "foreground"
"ui.cursor"       = { fg = "background", bg = "accent1" }
"ui.statusline"   = { fg = "foreground", bg = "surface" }
"ui.selection"    = { bg = "backgroundMuted" }
"ui.linenr"       = "foregroundMuted"
"ui.linenr.selected" = "accent1"

"diagnostic.error"   = { underline = { color = "danger",  style = "curl" } }
"diagnostic.warning" = { underline = { color = "warning", style = "curl" } }

"keyword"  = "accent2"
"function" = "accent3"
"string"   = "success"
"constant" = "accent4"
"type"     = "accent1"
"comment"  = { fg = "foregroundMuted", modifiers = ["italic"] }
```

## Integration

Set this in your `~/.config/helix/config.toml`:

```toml
theme = "tinct"
```

Or, inside a running Helix session, run:

```
:theme tinct
```

`:theme` takes effect immediately for the current session but does not survive a restart — set the `theme` key in `config.toml` to make it permanent.

## Reload behaviour

### Automatic

**None** — Helix has no IPC channel for theme reload. After generation, running Helix sessions continue with their currently-loaded theme until the user reloads explicitly. New sessions started after generation pick up the new theme via `config.toml` automatically.

### Manual fallback

Inside any running Helix session:

```
:config-reload
:theme tinct
```

`:config-reload` re-reads `config.toml`; `:theme tinct` swaps the theme in-place without restarting. If your `config.toml` already has `theme = "tinct"`, `:config-reload` alone is enough — but `:theme tinct` is the safer one-liner because it works regardless of whether you've set the config-file value.

## Uninstall / revert

1. **Remove the config line** from `~/.config/helix/config.toml`:

   Delete `theme = "tinct"` (or change it back to your previous theme name).

2. **Delete the generated files**:

   ```bash
   rm ~/.config/helix/themes/tinct.toml ~/.config/helix/themes/tinct-palette.toml
   ```

3. **Reload to drop the theme**: inside Helix, run `:theme default` (or any other installed theme name), or simply restart Helix.

4. **External state**: this plugin only writes to `~/.config/helix/themes/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--helix.output-dir` | `~/.config/helix/themes` | Override the output directory |

## Colour role mapping

### UI

| Helix key | Tinct role |
|---|---|
| `ui.background` (bg) | `background` |
| `ui.text` | `foreground` |
| `ui.text.inactive` | `foregroundMuted` |
| `ui.cursor` (fg / bg) | `background` / `accent1` |
| `ui.cursor.match` | `warning` (bold) |
| `ui.cursor.select` (bg) | `accent2` |
| `ui.selection` (bg) | `backgroundMuted` |
| `ui.selection.primary` (bg) | `accent1Muted` |
| `ui.linenr` | `foregroundMuted` |
| `ui.linenr.selected` | `accent1` (bold) |
| `ui.statusline` (fg / bg) | `foreground` / `surface` |
| `ui.statusline.inactive` (fg / bg) | `foregroundMuted` / `backgroundMuted` |
| `ui.statusline.normal` (fg / bg) | `onAccent1` / `accent1` |
| `ui.statusline.insert` (fg / bg) | `onAccent4` / `accent4` |
| `ui.statusline.select` (fg / bg) | `onAccent2` / `accent2` |
| `ui.bufferline.active` (fg / bg) | `foreground` / `background` (bold) |
| `ui.popup` / `ui.menu` (fg / bg) | `foreground` / `surface` |
| `ui.menu.selected` (fg / bg) | `onAccent1` / `accent1` (bold) |
| `ui.virtual.indent-guide` | `outlineVariant` |
| `ui.virtual.ruler` (bg) | `surfaceVariant` |

### Diagnostics & diff

| Helix key | Tinct role |
|---|---|
| `diagnostic.error` | `danger` (curl underline) |
| `diagnostic.warning` | `warning` (curl underline) |
| `diagnostic.info` | `info` (curl underline) |
| `diagnostic.hint` | `success` (curl underline) |
| `diff.plus` | `success` |
| `diff.minus` | `danger` |
| `diff.delta` | `warning` |

### Syntax

| Helix key | Tinct role | Notes |
|---|---|---|
| `keyword`, `keyword.control` | `accent2` | conditional/repeat get italics |
| `keyword.control.return`, `.exception` | `danger` | |
| `keyword.control.import`, `keyword.directive` | `accent3` | |
| `function`, `function.method` | `accent3` | builtin italicised |
| `function.macro`, `function.special` | `accent4` | |
| `string` | `success` | |
| `string.regexp` | `warning` | |
| `constant`, `constant.numeric` | `accent4` | |
| `constant.character` | `success` | |
| `type`, `type.builtin` | `accent1` | builtin italicised |
| `type.enum.variant` | `accent4` | |
| `attribute`, `constructor` | `accent3` | |
| `variable` | `foreground` | |
| `variable.builtin` | `accent2` (italic) | |
| `variable.other.member` | `accent3` | |
| `label`, `tag` | `accent2` | |
| `namespace` | `accent1` | |
| `comment` | `foregroundMuted` (italic) | |
| `operator`, `punctuation` | `foregroundMuted` | brackets stay `foreground` for parity |
| `special` | `warning` | |

### Markup

| Helix key | Tinct role |
|---|---|
| `markup.heading.1` | `accent1` (bold) |
| `markup.heading.2` | `accent2` (bold) |
| `markup.heading.3` | `accent3` (bold) |
| `markup.heading.4..6` | `accent4` |
| `markup.list.*` | `accent2` |
| `markup.link.url` | `info` (underlined) |
| `markup.link.text` | `accent3` |
| `markup.raw.*` | `success` |
| `markup.quote` | `foregroundMuted` (italic) |

## Customising the template

Extract the default templates to override either or both:

```bash
tinct plugins templates dump -o helix -l ~/.config/tinct/templates/helix
```

This writes `~/.config/tinct/templates/helix/tinct-palette.toml.tmpl` and `~/.config/tinct/templates/helix/tinct.toml.tmpl`. Tinct uses your version in preference to the embedded default — you can override just one file (tinct will fall back to the embedded copy of the other).

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### `theme not found: tinct`

Helix searches `~/.config/helix/themes/` (and the runtime themes directory shipped with the binary). Confirm the path:

```bash
ls -l ~/.config/helix/themes/tinct.toml ~/.config/helix/themes/tinct-palette.toml
```

Both files must be present. The main `tinct.toml` references `tinct-palette` via `inherits =`; if the palette file is missing, Helix logs a parse error and falls back to its built-in default theme.

### Colours don't change after `:theme tinct`

Helix caches the parsed theme. If you regenerated while the editor was open and `:theme tinct` shows the old colours, run `:config-reload` first — that re-reads the theme file. If colours still look stale, the palette file may not have been rewritten (e.g. permissions error during generation); check `tinct generate -o helix --verbose` output.

### `inherits` chain error

The plugin emits exactly one inheritance step (`tinct.toml` → `tinct-palette.toml`). If you customise `tinct.toml` to `inherits` from a different palette, the colour names in this README's mapping table will only resolve if your palette defines the same keys (`background`, `accent1`, `danger`, `surface`, …). Either keep the inherited palette name, or add the keys you use to your own palette.

### Theme appears but syntax highlighting looks wrong

Helix's syntax keys are tree-sitter scope names, not generic categories. If a specific language's highlights look off (e.g. Rust lifetimes, Python decorators), it's almost always a missing scope mapping — copy the template (see *Customising* above) and add the scope name with a colour from the palette.

## Related plugins

- [neovim](./neovim.md) — the other modal editor target. Different theme format but the same colour-role vocabulary.
- [ghostty](../terminals/ghostty.md), [wezterm](../terminals/wezterm.md), [kitty](../terminals/kitty.md) — pair the editor palette with a matching terminal scheme.
