---
title: wbg
sidebar_position: 1
plugin:
  type: output
  name: wbg
  category: special
  source: builtin
  app: wbg
  app_url: 'https://codeberg.org/dnkl/wbg'
  requires: []
  optional:
    - wbg
  pattern: single-file
  default_output_dir: ~/.config/wbg
  generated_files:
    - tinct-wbg.sh
  reload:
    method: wallpaper-apply
    command: pkill wbg; wbg <wallpaper>
    user_action_required: false
---

# wbg

Wallpaper integration for [wbg](https://codeberg.org/dnkl/wbg), the minimal Wayland wallpaper application from the dnkl/foot ecosystem. wbg is a single-shot foreground process — it displays one image scaled-to-fit on all monitors until killed. The plugin writes a small shell script that captures the current wallpaper and, when wbg is available, restarts the process so the new image takes effect immediately.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o wbg` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o wbg
```

If `wbg` is on `$PATH` the wallpaper switches immediately. The generated `tinct-wbg.sh` is made executable so you can re-run it from your compositor autostart to restore the wallpaper after a fresh login.

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-wbg.sh` | `~/.config/wbg/tinct-wbg.sh` | POSIX shell script that kills any existing `wbg` and `exec`s a fresh one with the wallpaper path. Marked executable. |

When no wallpaper source is available (e.g. you used a non-image input plugin) the script is generated with usage hints rather than a working `exec` line — wbg can't display a solid colour, so there's nothing to do.

## Integration

wbg has no config file — it's a CLI argument. Tinct's role is to write a self-contained shell script that re-launches wbg with the chosen image. Wire it into your compositor autostart so the wallpaper survives a reboot:

- **Hyprland**: `exec-once = ~/.config/wbg/tinct-wbg.sh`
- **Sway**: `exec ~/.config/wbg/tinct-wbg.sh`
- **Generic Wayland session**: `~/.config/wbg/tinct-wbg.sh &`

## Reload behaviour

### Automatic

After writing the script the plugin's `Wallpaper` hook runs:

```
killall wbg
wbg /absolute/path/to/wallpaper       # detached, released
```

wbg is a foreground process with no IPC, so the only way to change the image is to kill and restart it. The replacement process is detached from tinct so it continues running after the command exits.

If `wbg` isn't on `$PATH` the wallpaper step is skipped silently; the script on disk still works the first time wbg becomes available.

### Manual fallback

Re-run the generated script:

```bash
~/.config/wbg/tinct-wbg.sh
```

Or do the kill-and-restart by hand:

```bash
killall wbg 2>/dev/null
wbg ~/Pictures/wallpaper.jpg &
```

## Uninstall / revert

1. **Remove the autostart entry**: delete the `exec`/`exec-once` line that points at `~/.config/wbg/tinct-wbg.sh` from your compositor config.

2. **Delete the generated script**:

   ```bash
   rm ~/.config/wbg/tinct-wbg.sh
   ```

3. **Drop the current wallpaper**:

   ```bash
   killall wbg
   ```

   wbg exits and the compositor falls back to whatever the compositor's own background is (typically the default colour for empty space). The plugin does not track the previous wallpaper, so restoring a specific earlier image must be done by hand.

4. **External state**: this plugin only writes to `~/.config/wbg/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--wbg.output-dir` | `~/.config/wbg` | Override the output directory |

## Colour role mapping

wbg displays an image. It does not consume tinct's palette. There is no colour role mapping.

## Customising the template

Dump the default template to override it:

```bash
tinct plugins templates dump -o wbg -l ~/.config/tinct/templates/wbg
```

This writes `~/.config/tinct/templates/wbg/tinct-wbg.sh.tmpl`. Tinct uses your version in preference to the embedded default. Common reasons to override: change the `killall` invocation (e.g. `pkill -x wbg`), add `--stretch`, or hard-code a specific output.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### Wallpaper didn't change

`wbg` must be installed for the kill-and-restart to do anything. Check:

```bash
which wbg
```

If it's missing, install wbg from your distro (often `extra/wbg` or build from source). The generated script will start working the next time you run it.

### Multiple wbg instances

The plugin runs `killall wbg` before starting the new one. If you still end up with multiple processes, another autostart entry (e.g. your compositor's built-in background, or a leftover from a previous wallpaper tool) is racing tinct. Remove the duplicate launcher.

### Image format unsupported

wbg supports PNG, JPEG, JPEG XL, WebP, and SVG (subject to compile-time options). For GIFs or unsupported formats, convert first (`magick`, `cwebp`, `gifsicle`) or switch to the [awww](./awww.md) plugin, which handles animated wallpapers.

### Script exits immediately

The script ends with `exec wbg ...`, which replaces the shell process. When run from a terminal you'll see wbg take over the terminal foreground; ctrl-C kills wbg. Run with `&` (or via autostart) for normal daemon-style use.

## Related plugins

- [awww](./awww.md) — animated wallpaper daemon with IPC and transitions; richer alternative to wbg.
- [hyprpaper](../hyprland/hyprpaper.md) — Hyprland's native wallpaper daemon.
