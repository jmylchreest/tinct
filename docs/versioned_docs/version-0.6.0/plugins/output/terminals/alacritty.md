---
title: alacritty
sidebar_position: 1
plugin:
  type: output
  name: alacritty
  category: terminals
  source: builtin
  app: Alacritty
  app_url: 'https://alacritty.org/'
  requires:
    - alacritty
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/alacritty
  generated_files:
    - tinct-colors.toml
  reload:
    method: none
    user_action_required: false
---

# alacritty

Generates a TOML colour theme for [Alacritty](https://alacritty.org/), the GPU-accelerated cross-platform terminal emulator. The output is a single fragment file you import from your main `alacritty.toml`; Alacritty watches its config tree and applies new colours to every open window the moment the file changes on disk.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o alacritty` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o alacritty
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colors.toml` | `~/.config/alacritty/tinct-colors.toml` | Single-file palette containing `[colors.*]` tables (primary, cursor, selection, search, hints, footer bar, normal, bright, dim). |

## Generated format

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

## Integration

Alacritty has no auto-include directory, so you import the file by hand. Add the following to your `alacritty.toml`:

```toml
[general]
import = [
  "~/.config/alacritty/tinct-colors.toml",
]
```

(Older Alacritty versions used a top-level `import = [...]` array — both forms still work.) Once the line is in place you never need to touch your `alacritty.toml` again; re-running `tinct generate` overwrites `tinct-colors.toml` and Alacritty picks it up live.

## Reload behaviour

### Automatic

None from the plugin — Alacritty's built-in `live_config.reload` (on by default) watches the imported file and reloads colours into every open window as soon as `tinct generate` finishes writing it.

### Manual fallback

If you've disabled `live_config.reload`, send Alacritty the reload key chord (default `Ctrl+Shift+R` on Linux, `Cmd+Shift+R` on macOS), or restart Alacritty.

## Uninstall / revert

1. **Remove the import line** from your `~/.config/alacritty/alacritty.toml`:

   Delete the entry for `tinct-colors.toml` from the `import = [...]` array (and drop the `[general]` section or array entirely if it now contains nothing else).

2. **Delete the generated file**:

   ```bash
   rm ~/.config/alacritty/tinct-colors.toml
   ```

3. **Reload to drop the theme**:

   Saving `alacritty.toml` triggers the live reload. If you've disabled it, press `Ctrl+Shift+R` (or `Cmd+Shift+R` on macOS).

4. **External state**: this plugin only writes to `~/.config/alacritty/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--alacritty.output-dir` | `~/.config/alacritty` | Override the output directory |

## Colour role mapping

| Alacritty table | Property | Tinct role |
|---|---|---|
| `[colors.primary]` | `background` | `background` |
| `[colors.primary]` | `foreground` | `foreground` |
| `[colors.primary]` | `dim_foreground` | `foregroundMuted` |
| `[colors.primary]` | `bright_foreground` | `foreground` |
| `[colors.cursor]` | `text` | `background` |
| `[colors.cursor]` | `cursor` | `accent1` |
| `[colors.vi_mode_cursor]` | `cursor` | `accent2` |
| `[colors.selection]` | `background` | `accent1` |
| `[colors.search.matches]` | `background` | `warning` |
| `[colors.search.focused_match]` | `background` | `accent1` |
| `[colors.hints.start]` | `background` | `warning` |
| `[colors.hints.end]` | `background` | `success` |
| `[colors.footer_bar]` | `foreground` | `foreground` |
| `[colors.footer_bar]` | `background` | `backgroundMuted` |
| `[colors.normal]` / `[colors.dim]` | `black`..`white` | ANSI `black`..`white` |
| `[colors.bright]` | `black`..`white` | ANSI `brightblack`..`brightwhite` |

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o alacritty -l ~/.config/tinct/templates/alacritty
```

This writes `~/.config/tinct/templates/alacritty/tinct-colors.toml.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Colours unchanged after `tinct generate`

Check that the import is wired up and the file was written:

```bash
grep tinct-colors ~/.config/alacritty/alacritty.toml
ls -l ~/.config/alacritty/tinct-colors.toml
```

If both look right but nothing changed visually, your `alacritty.toml` probably defines its own `[colors.*]` tables *after* the import — explicit values win. Move the offending blocks above the `[general]` section or delete them.

### "live_config.reload" was disabled

Re-enable it by adding `live_config_reload = true` to your `alacritty.toml`, or press the reload key chord after each generation.

### Import path with `~` doesn't expand

Alacritty does expand `~` in import paths since 0.13, but very old builds don't. If you're on an ancient version, use the full path:

```toml
import = ["/home/you/.config/alacritty/tinct-colors.toml"]
```

## Related plugins

- [ghostty](./ghostty.md), [kitty](./kitty.md), [wezterm](./wezterm.md) — other GPU-accelerated terminals.
- [tmux](../special/tmux.md) — pair the terminal palette with a matching multiplexer status bar.
