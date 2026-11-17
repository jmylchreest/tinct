---
title: btop
sidebar_position: 1
plugin:
  type: output
  name: btop
  category: special
  source: builtin
  app: btop
  app_url: 'https://github.com/aristocratos/btop'
  requires:
    - btop
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/btop/themes
  generated_files:
    - tinct.theme
  reload:
    method: signal
    command: pkill -SIGUSR2 -x btop
    user_action_required: false
---

# btop

Generates a [btop](https://github.com/aristocratos/btop) theme file — the colours behind the resource monitor's CPU/memory/network/process boxes, gradient meters, temperature scales, and status text. btop is a curses-style TUI for terminals; it reads one `.theme` file at a time and re-reads it when sent `SIGUSR2`.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o btop` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o btop
```

The first time, you also need to point btop at the theme — see [Integration](#integration).

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.theme` | `~/.config/btop/themes/tinct.theme` | btop theme file. Sets every `theme[...]` key btop recognises: backgrounds, foregrounds, gradient endpoints for CPU/memory/network/disk graphs, box outlines, and meter backgrounds. |

The filename matches `--btop.theme-name` (default `tinct`). The filename without `.theme` is also the value you put into btop's `color_theme` config key.

## Integration

btop loads themes by name from `~/.config/btop/themes/`. After the first generate, tell btop to use it — either edit the config:

```ini
# ~/.config/btop/btop.conf
color_theme = "tinct"
```

…or pick it interactively: press `p` (options menu) → cycle to "tinct" in the theme list → `Esc` to save.

The theme name is `tinct` by default; if you used `--btop.theme-name=mything` then it's `"mything"` instead.

## Reload behaviour

### Automatic

After writing the theme file the plugin sends `SIGUSR2` to every running `btop` process (this is the verb declared in `Hooks().Reload`). btop traps that signal and re-reads its config + active theme, so the colours change live without exiting the TUI.

The reload only applies if `color_theme = "tinct"` is already set in `btop.conf`; otherwise btop re-reads the file but keeps using whichever theme it was already on. Use `--btop.reload=false` to skip the signal.

### Manual fallback

If the signal didn't reach btop (it wasn't running at generate time, or you're in a different mount namespace):

```bash
pkill -SIGUSR2 -x btop
```

Or restart the process, or press `p` in the running TUI and re-select the theme.

## Uninstall / revert

1. **Remove the config line**: edit `~/.config/btop/btop.conf` and either delete `color_theme = "tinct"` or change it to another theme name (e.g. `color_theme = "Default"`).

2. **Delete the generated file**:

   ```bash
   rm ~/.config/btop/themes/tinct.theme
   ```

3. **Reload to drop the theme**:

   ```bash
   pkill -SIGUSR2 -x btop
   ```

   Or restart btop. If `color_theme` still pointed at `tinct` you may briefly see a missing-theme warning until you set it to something else.

4. **External state**: this plugin only writes to `~/.config/btop/themes/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--btop.output-dir` | `~/.config/btop/themes` | Override the output directory |
| `--btop.theme-name` | `tinct` | Theme name; written as `<name>.theme` and used as the `color_theme` value |
| `--btop.reload` | `true` | Send `SIGUSR2` to running btop instances after writing the theme |

## Colour role mapping

| btop key | Tinct role | Notes |
|---|---|---|
| `main_bg` | `background` | Terminal background; leave empty in the template for transparent terminals |
| `main_fg` | `foreground` | Main text colour |
| `title` | `accent1` | Box titles |
| `hi_fg` | `accent2` | Highlighted shortcut keys |
| `selected_bg` / `selected_fg` | `accent1` / `onAccent1` | Selected row in process list |
| `inactive_fg` | `foregroundMuted` | Greyed-out text |
| `proc_misc` | `accent3` | Mini graphs in process box |
| `cpu_box`, `mem_box`, `net_box`, `proc_box` | `outline` | Box borders |
| `div_line` | `outlineVariant` | Divider line at top of each box |
| `temp_start` / `temp_mid` / `temp_end` | `success` / `warning` / `danger` | Temperature gradient (cool → hot) |
| `cpu_start` / `cpu_mid` / `cpu_end` | `accent1` / `accent2` / `accent3` | CPU usage gradient |
| `free_start` | `success` | Memory: free |
| `cached_start` | `info` | Memory: cached |
| `available_start` | `accent2` | Memory: available |
| `used_start` / `used_mid` / `used_end` | `accent1` / `warning` / `danger` | Memory: used (low → critical) |
| `download_*` / `upload_*` | `info`/`accent1`/`accent2` and `warning`/`accent3`/`danger` | Network throughput gradients |
| `process_start` / `process_mid` / `process_end` | `accent1` / `accent2` / `accent3` | Process box graph |
| `meter_bg` | `surfaceContainer` | Background of meter bars |

## Customising the template

Dump the default template to override it:

```bash
tinct plugins templates dump -o btop -l ~/.config/tinct/templates/btop
```

This writes `~/.config/tinct/templates/btop/tinct.theme.tmpl`. Tinct uses your copy in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### Theme didn't change

Check `color_theme` is actually set:

```bash
grep color_theme ~/.config/btop/btop.conf
```

If it's not `"tinct"` (or whatever `--btop.theme-name` you used), btop is reloading the file but ignoring it. Set it via `p` in the TUI or by editing the conf.

### `pkill: SIGUSR2: unknown signal`

Some `pkill` implementations want `-USR2` or numeric `-12`. The plugin uses Go's `signal` package so this only affects manual reload commands. Try:

```bash
pkill -USR2 -x btop
```

### Theme list is empty inside btop

btop only scans `~/.config/btop/themes/` (plus its built-in themes). If `tinct.theme` is in a different directory (because you set `--btop.output-dir`), move or symlink it back, or accept that btop won't find it.

### Background isn't transparent

`main_bg` is set to the theme background. To get terminal-transparent btop, override the template and set `theme[main_bg]=""` (empty string), or edit the generated file in place.

## Related plugins

- [neovim](../editors/neovim.md) — companion theme for the editor often run alongside btop.
- [tmux](./tmux.md) — tmux colours for users running btop inside a tmux pane.
