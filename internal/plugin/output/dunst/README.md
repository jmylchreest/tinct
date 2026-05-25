---
title: dunst
sidebar_position: 2

plugin:
  type: output
  name: dunst
  category: bars-launchers
  source: builtin
  app: Dunst
  app_url: https://dunst-project.org/
  requires: [dunst]
  optional: [dunstctl]
  pattern: drop-in
  default_output_dir: ~/.config/dunst/dunstrc.d
  generated_files: [60-tinct.conf]
  reload:
    method: ipc
    command: dunstctl reload
    user_action_required: false
---

# dunst

Generates an urgency-based colour theme for [Dunst](https://dunst-project.org/), the lightweight notification daemon for X11 and Wayland. The output is a single drop-in config file that dunst loads automatically; no manual config changes are required.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o dunst` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o dunst
```

## Generated files

| File | Path | Role |
|------|------|------|
| `60-tinct.conf` | `~/.config/dunst/dunstrc.d/60-tinct.conf` | Drop-in conf with `[global]` frame styling plus `[urgency_low]`, `[urgency_normal]`, `[urgency_critical]` sections. |

The `60-` prefix orders the file after most other drop-ins (00–50) so colour settings override them, but before any high-priority overrides (70–99) the user maintains by hand.

## Generated format

```ini
[global]
frame_color = "#89b4fa"
separator_color = "#313244"

[urgency_low]
background = "#1e1e2e"
foreground = "#cdd6f4"
frame_color = "#89b4fa"

[urgency_normal]
background = "#1e1e2e"
foreground = "#cdd6f4"
frame_color = "#89b4fa"

[urgency_critical]
background = "#1e1e2e"
foreground = "#cdd6f4"
frame_color = "#f38ba8"
```

## Integration

**No manual configuration required.** Dunst auto-loads every `.conf` file in `~/.config/dunst/dunstrc.d/` after reading the main `dunstrc`. The plugin writes directly into that directory; the file becomes active on the next dunst reload.

If you don't already have a `dunstrc.d/` directory, the plugin creates it.

## Reload behaviour

### Automatic

After writing the file, the plugin runs:

```
dunstctl reload
```

This is dunst's IPC reload — it re-reads all config files in-place without restarting the daemon. Notifications already on screen are preserved.

If `dunstctl` isn't on `$PATH`, the auto-reload is skipped silently. Generation still succeeds; the theme just won't appear until dunst restarts.

### Manual fallback

If `dunstctl` is unavailable or the IPC reload fails for any reason:

```bash
killall dunst && dunst &
```

Or, if dunst is managed by systemd:

```bash
systemctl --user restart dunst
```

## Uninstall / revert

1. **Remove the config line**: none added. The plugin uses dunst's drop-in directory, so no edits to `dunstrc` or any other file ever happened.

2. **Delete the generated file**:

   ```bash
   rm ~/.config/dunst/dunstrc.d/60-tinct.conf
   ```

3. **Reload to drop the theme**:

   ```bash
   dunstctl reload
   ```

4. **External state**: this plugin only writes to `~/.config/dunst/dunstrc.d/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dunst.output-dir` | `~/.config/dunst/dunstrc.d` | Override the output directory |

## Colour role mapping

| Dunst section | Property | Tinct role | Notes |
|---|---|---|---|
| `[global]` | `frame_color` | `backgroundMuted` | Default frame for all urgencies that don't override it |
| `[global]` | `separator_color` | `frame` | Matches `frame_color` (dunst keyword) |
| `[urgency_low]` | `background` | `background` | |
| `[urgency_low]` | `foreground` | `foregroundMuted` | Lower-priority text |
| `[urgency_low]` | `frame_color` | `info` @ 80% alpha | |
| `[urgency_low]` | `highlight` | `info` | Progress bars |
| `[urgency_normal]` | `background` | `background` | |
| `[urgency_normal]` | `foreground` | `foreground` | |
| `[urgency_normal]` | `frame_color` | `accent1` | |
| `[urgency_normal]` | `highlight` | `accent1` | |
| `[urgency_critical]` | `background` | `danger` @ 93% alpha | Bright background to grab attention |
| `[urgency_critical]` | `foreground` | `background` | Inverted for contrast against red |
| `[urgency_critical]` | `frame_color` | `danger` | |
| `[urgency_critical]` | `highlight` | `warning` | |

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o dunst -l ~/.config/tinct/templates/dunst
```

This creates `~/.config/tinct/templates/dunst/tinct.dunstrc.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Theme not visible after generation

Check that dunst loaded the drop-in:

```bash
ls -l ~/.config/dunst/dunstrc.d/60-tinct.conf
dunstctl reload
notify-send "Test" "Should now be themed"
```

If the file exists but the theme isn't applied, your `dunstrc` may have settings that take precedence. Drop-ins layer on top of `dunstrc`, so explicit colour values in `dunstrc` override what dunst loads from `dunstrc.d/`. Remove the conflicting lines from `dunstrc`.

### `dunstctl: command not found`

`dunstctl` ships in the same package as `dunst` on most distros, but some minimal installs split it out. The plugin treats it as optional — generation still succeeds, but you'll need to restart dunst manually for changes to apply.

### Critical notifications time out

By default the template sets `timeout = 0` (never timeout) for `urgency_critical`. If yours expire anyway, your `dunstrc` is overriding the drop-in. Either remove the `timeout` line from `dunstrc`'s `[urgency_critical]` section, or move your settings into a higher-numbered drop-in (e.g. `~/.config/dunst/dunstrc.d/70-overrides.conf`).

## Related plugins

- [waybar](../waybar/README.md) — status bar styling, often configured alongside dunst.
- [hyprlock](../hyprlock/README.md) — lock screen styling for Hyprland users.
