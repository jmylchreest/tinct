---
title: awww
sidebar_position: 1

plugin:
  type: output
  name: awww
  category: special
  source: builtin
  app: awww wallpaper daemon
  app_url: https://codeberg.org/LGFae/awww
  requires: []
  optional: [awww]
  pattern: single-file
  default_output_dir: ~/.config/awww
  generated_files: [tinct-awww.sh]
  reload:
    method: wallpaper-apply
    command: awww img <wallpaper>
    user_action_required: false
---

# awww

Wallpaper integration for [awww](https://codeberg.org/LGFae/awww) (An Answer to your Wayland Wallpaper Woes), an animated wallpaper daemon for Wayland compositors implementing the wlr-layer-shell protocol. The plugin generates a persistence shell script and, if awww-daemon is running, applies the wallpaper directly using the configured transition.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o awww` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o awww
```

If awww-daemon is already running the wallpaper is applied immediately with the default fade transition. The generated `tinct-awww.sh` is also made executable so you can re-run it from your compositor autostart to restore the wallpaper after a daemon restart or fresh login.

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-awww.sh` | `~/.config/awww/tinct-awww.sh` | POSIX shell script that waits for `awww-daemon` and calls `awww img` with the wallpaper path and the transition options chosen at generate time. Marked executable. |

When no wallpaper source is available (e.g. you used a non-image input plugin) the script falls back to `awww clear <hex>` using the theme background colour.

## Integration

awww is unusual: there is no central config file. Tinct's role is to write a self-contained shell script that drives `awww` on demand. Wire it into your compositor autostart so it re-runs on login:

- **Hyprland**: `exec-once = ~/.config/awww/tinct-awww.sh`
- **Sway**: `exec ~/.config/awww/tinct-awww.sh`
- **Generic Wayland session**: `~/.config/awww/tinct-awww.sh &`

The script polls `awww query` for up to ~5 s, so it tolerates being launched before the daemon is fully up.

## Reload behaviour

### Automatic

After writing the script the plugin probes `awww query`. If the daemon answers, it runs the equivalent of:

```
awww img --transition-type <type> --transition-step <step> --transition-fps <fps> [--resize <mode>] [--fill-color <rrggbb>] [-o <outputs>] <wallpaper>
```

This is awww's IPC interface — the new wallpaper is swapped in using the configured transition without restarting the daemon. If `awww query` fails (daemon not running, binary missing) the wallpaper step is skipped silently; the script on disk still works on the next login.

### Manual fallback

If you need to re-apply the wallpaper without regenerating the theme:

```bash
~/.config/awww/tinct-awww.sh
```

Or invoke `awww` directly:

```bash
awww img --transition-type simple ~/Pictures/wallpaper.jpg
```

## Uninstall / revert

1. **Remove the autostart entry**: delete the `exec`/`exec-once` line that points at `~/.config/awww/tinct-awww.sh` from your compositor config.

2. **Delete the generated script**:

   ```bash
   rm ~/.config/awww/tinct-awww.sh
   ```

3. **Drop the current wallpaper**: awww keeps the last image in memory until the daemon exits. To clear it:

   ```bash
   awww clear 000000   # or any hex colour
   ```

   To pick a different image, run `awww img <new-image>` directly. The plugin does not track the previous wallpaper, so reverting to whatever you had before tinct must be done by hand.

4. **External state**: this plugin only writes to `~/.config/awww/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--awww.output-dir` | `~/.config/awww` | Override the output directory |
| `--awww.transition-type` | `simple` | Transition effect (`simple`, `left`, `right`, `top`, `bottom`, `wipe`, `grow`, `center`, `outer`, `any`, `random`) |
| `--awww.transition-step` | `2` | Transition step (1–255; smaller = smoother) |
| `--awww.transition-fps` | `30` | Transition FPS (1–65535) |
| `--awww.resize` | `` | Resize mode (`crop`, `fit`, `no`) |
| `--awww.fill-color` | `` | Fill colour for `fit` mode as `rrggbb`; empty uses awww's default |
| `--awww.outputs` | `` | Comma-separated list of output names to target (default: all) |

## Colour role mapping

awww is a wallpaper daemon, not a theming target — it does not consume tinct's palette. The single point of integration is the fallback `awww clear <hex>` invocation, which uses:

| awww command | Tinct role | Notes |
|---|---|---|
| `awww clear <hex>` | `background` | Used only when no wallpaper image is supplied |

## Customising the template

Dump the default template to make it your own:

```bash
tinct plugins templates dump -o awww -l ~/.config/tinct/templates/awww
```

This writes `~/.config/tinct/templates/awww/tinct-awww.sh.tmpl`. Tinct prefers your version over the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### Wallpaper not applied at generate time

`awww-daemon` must be running for the IPC call to succeed. Start it (typically from your compositor autostart) and re-run:

```bash
awww-daemon &
awww query                                  # should respond
tinct generate -i image -p wallpaper.jpg -o awww
```

If verbose mode shows `Skipping wallpaper application (awww-daemon not running)`, the daemon was unreachable; the script on disk is still correct and will work on next login.

### Wallpaper not restored after reboot

awww has no built-in persistence — it forgets the wallpaper when the daemon exits. Wire `~/.config/awww/tinct-awww.sh` into your compositor autostart (see Integration).

### High CPU when displaying a large GIF

awww caches GIF frames on first display. For very large GIFs, downscale them first:

```bash
gifsicle --resize-fit 1920x1080 large.gif > optimised.gif
tinct generate -i image -p optimised.gif -o awww
```

### Wrong output targeted

Use `--awww.outputs` with the output names reported by `awww query` (or `wlr-randr`). For example: `--awww.outputs=DP-1,HDMI-A-1`.

## Related plugins

- [wbg](../wbg/README.md) — simpler single-shot Wayland wallpaper setter, no transitions.
- [hyprpaper](../hyprpaper/README.md) — Hyprland's native wallpaper daemon.
