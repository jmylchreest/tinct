---
title: ghostty
sidebar_position: 2
plugin:
  type: output
  name: ghostty
  category: terminals
  source: builtin
  app: Ghostty
  app_url: 'https://ghostty.org/'
  requires:
    - ghostty
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/ghostty/themes
  generated_files:
    - tinct.conf
  reload:
    method: signal
    command: pkill -USR2 -x ghostty
    user_action_required: false
---

# ghostty

Generates a single-file colour theme for [Ghostty](https://ghostty.org/), Mitchell Hashimoto's GPU-accelerated terminal emulator. The file lives in Ghostty's `themes/` directory so it shows up alongside the built-in themes; you activate it with one config line.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o ghostty` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o ghostty
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.conf` | `~/.config/ghostty/themes/tinct.conf` | Single-file theme — `background`, `foreground`, cursor/selection colours, and the 16-colour `palette = N=#rrggbb` ANSI table. |

On macOS without `$XDG_CONFIG_HOME`, the path is `~/Library/Application Support/com.mitchellh.ghostty/themes/tinct.conf`. On Windows it's `%APPDATA%/ghostty/themes/tinct.conf`.

## Generated format

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

## Integration

Add one line to your Ghostty config (`~/.config/ghostty/config` on Linux, the equivalent platform path elsewhere):

```
theme = tinct
```

Ghostty searches `themes/` next to its config first, so the `tinct` name resolves to the file this plugin writes. You can also pick "tinct" interactively via `ghostty +list-themes` or the in-app theme picker.

## Reload behaviour

### Automatic

After writing the file, the plugin broadcasts `SIGUSR2` to every running `ghostty` process:

```
pkill -USR2 -x ghostty
```

Ghostty's signal handler re-reads its config and live-applies the new theme to every open window. No restart required.

On Windows the signal mechanism doesn't exist — the hook warns and skips it; the theme still loads on the next Ghostty start.

### Manual fallback

If `--ghostty.reload=false` was passed, or `ghostty` isn't running yet:

```bash
pkill -USR2 -x ghostty
```

…or restart Ghostty (`Cmd+Q` / close all windows then re-launch).

## Uninstall / revert

1. **Remove the config line** from your Ghostty config:

   Delete the `theme = tinct` line (or change it back to your previous theme).

2. **Delete the generated file**:

   ```bash
   rm ~/.config/ghostty/themes/tinct.conf
   ```

3. **Reload to drop the theme**:

   ```bash
   pkill -USR2 -x ghostty
   ```

4. **External state**: this plugin only writes to `~/.config/ghostty/themes/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ghostty.output-dir` | `~/.config/ghostty/themes` | Override the output directory |
| `--ghostty.reload` | `true` | Send `SIGUSR2` to running ghostty processes after generation |

## Colour role mapping

| Ghostty key | Tinct role |
|---|---|
| `background` | `background` |
| `foreground` | `foreground` |
| `cursor-color` | `accent1` |
| `cursor-text` | `background` |
| `selection-background` | `accent1` |
| `selection-foreground` | `foreground` |
| `palette = 0..7` | ANSI `black`..`white` |
| `palette = 8..15` | ANSI `bright-black`..`bright-white` |

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o ghostty -l ~/.config/tinct/templates/ghostty
```

This writes `~/.config/tinct/templates/ghostty/tinct.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### `theme = tinct` errors with "theme not found"

Ghostty searches `themes/` next to the config file it loaded. Confirm the path matches:

```bash
ghostty +show-config | grep -i theme
ls -l ~/.config/ghostty/themes/tinct.conf
```

If Ghostty is using a different config root (e.g. macOS Application Support), point the plugin at it with `--ghostty.output-dir` or set `TINCT_PLUGIN_GHOSTTY_OUTPUT_DIR`.

### Theme loaded but colours look stale

Auto-reload only happens if Ghostty was running when `tinct generate` finished. Re-run `pkill -USR2 -x ghostty` by hand, or pass `--ghostty.reload=true` (the default).

### `pkill: command not found` (Windows / minimal containers)

The signal broadcast is a no-op on Windows and exits silently if `pkill` is missing. The file is still written; restart Ghostty to load it.

## Related plugins

- [alacritty](./alacritty.md), [kitty](./kitty.md), [wezterm](./wezterm.md) — other GPU-accelerated terminals.
- [tmux](../special/tmux.md) — pair the terminal palette with a matching multiplexer status bar.
