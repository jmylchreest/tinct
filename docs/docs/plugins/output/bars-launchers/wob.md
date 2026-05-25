---
title: wob
sidebar_position: 9
plugin:
  type: output
  name: wob
  category: bars-launchers
  source: external
  app: wob
  app_url: 'https://github.com/francma/wob'
  version: 0.0.0
  protocol_version: 0.2.0
  repository: 'https://github.com/jmylchreest/tinct'
  install: tinct plugins install wob
  requires:
    - wob
  optional:
    - tail
  pattern: two-file
  default_output_dir: ~/.config/wob/themes
  generated_files:
    - tinct.ini
  reload:
    method: watch
    command: tinct-plugin-wob send (auto-restart on config change)
    user_action_required: false
---

# wob

Generates a `tinct.ini` theme overlay for [wob](https://github.com/francma/wob), the Wayland Overlay Bar — the bar you see when adjusting volume or brightness. The same binary also ships a wrapper subcommand (`tinct-plugin-wob start|send|stop|status`) that runs wob as a background process driven by a FIFO, automatically restarting it when the theme on disk has been rewritten. This means the bar's colours follow your latest tinct palette without you having to remember to kill/restart wob after every `tinct generate`.

## Installation

### Via the official plugin repository

```bash
tinct plugins install wob
```

### Build from source

```bash
cd contrib/plugins/output/wob
go build -ldflags "-X main.Version=0.0.0" -o tinct-plugin-wob
install -m 0755 tinct-plugin-wob ~/.local/bin/
```

### Install wob itself

The `wob` binary must be on `$PATH` for the wrapper subcommands (`start`/`send`) to work. The theme-generation path runs without wob installed but is pointless on its own.

| Distro | Install |
|---|---|
| Arch | `sudo pacman -S wob` |
| Debian/Ubuntu | `sudo apt install wob` |
| Fedora | `sudo dnf install wob` |
| Build from source | https://github.com/francma/wob |

### Verify

```bash
which tinct-plugin-wob
which wob
tinct-plugin-wob --plugin-info | jq .
```

The plugin uses tinct's JSON-stdio protocol (v0.2.0).

## Quick start

Generate the theme:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o wob
```

Start wob with the tinct theme overlaid on your base config (typically wired into your compositor autostart — see [Integration](#integration)):

```bash
tinct-plugin-wob start \
    --base-config ~/.config/wob/base.ini \
    --append-config ~/.config/wob/themes/tinct.ini
```

Send a value (e.g. from a volume keybind):

```bash
tinct-plugin-wob send 45
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.ini` | `~/.config/wob/themes/tinct.ini` | wob INI overlay — defines `default`/`low`/`normal`/`critical` styles plus overflow colours. Designed to be layered on top of a user-supplied `base.ini` via the wrapper's `--append-config`. |

You also need a `base.ini` of your own that defines the bar's geometry (size, anchor, timeout, etc.). A minimal example is shipped alongside this plugin at `base.ini.example`:

```ini
[default]
height = 50
width = 400
border_offset = 4
border_size = 2
bar_padding = 3
anchor = top
margin = 20
timeout = 2000
```

Copy it to `~/.config/wob/base.ini` and edit to taste.

## Integration

wob is a long-running process that reads percentages from a FIFO. The supplied `tinct-plugin-wob` wrapper handles start-up, FIFO management, and config-change detection.

### Hyprland

Autostart wob with the tinct overlay (add to `~/.config/hypr/hyprland.conf` or an `exec.conf` snippet):

```conf
exec-once = tinct-plugin-wob start \
    --base-config ~/.config/wob/base.ini \
    --append-config ~/.config/wob/themes/tinct.ini
```

Wire keybinds to the `send` subcommand:

```conf
# Volume controls
bindel = , XF86AudioRaiseVolume, exec, wpctl set-volume @DEFAULT_AUDIO_SINK@ 5%+ && tinct-plugin-wob send $(wpctl get-volume @DEFAULT_AUDIO_SINK@ | awk '{print int($2*100+0.5)}')
bindel = , XF86AudioLowerVolume, exec, wpctl set-volume @DEFAULT_AUDIO_SINK@ 5%- && tinct-plugin-wob send $(wpctl get-volume @DEFAULT_AUDIO_SINK@ | awk '{print int($2*100+0.5)}')
bindel = , XF86AudioMute, exec, wpctl set-mute @DEFAULT_AUDIO_SINK@ toggle && tinct-plugin-wob send 0

# Brightness controls (uses current/max two-argument form)
bindel = , XF86MonBrightnessUp, exec, brightnessctl s 5%+ && tinct-plugin-wob send $(brightnessctl get) $(brightnessctl m)
bindel = , XF86MonBrightnessDown, exec, brightnessctl s 5%- && tinct-plugin-wob send $(brightnessctl get) $(brightnessctl m)
```

The `send` subcommand also accepts `--style <name>` to override the default style (`low`, `normal`, `critical`, `warning`).

### Sway / other wlroots compositors

The same wrapper works on any compositor — wire `tinct-plugin-wob start` into your autostart and `tinct-plugin-wob send …` into your keybinds. wob itself is compositor-agnostic.

### Without the wrapper

If you'd rather use raw wob, the theme file is a plain INI overlay you can pass to wob directly:

```bash
cat ~/.config/wob/base.ini ~/.config/wob/themes/tinct.ini > /tmp/wob.ini
mkfifo /tmp/wob.fifo
tail -f /tmp/wob.fifo | wob -c /tmp/wob.ini &
echo 50 > /tmp/wob.fifo
```

You're then responsible for FIFO lifecycle, restart-on-config-change, and PID management — exactly what the wrapper handles for you.

## Reload behaviour

### Automatic

When you invoke `tinct-plugin-wob send`, the wrapper checks the mtime of the merged base+theme config it last started wob with. If tinct has rewritten `tinct.ini` since wob was launched, the wrapper:

1. Stops the running wob process (SIGTERM).
2. Re-merges the configs and starts a fresh wob with the new theme.
3. Forwards the percentage value to the new instance.

The next volume/brightness keypress after a `tinct generate` is what triggers the restart — there is no explicit reload step.

### Manual fallback

To force a restart immediately after regenerating the theme (without waiting for the next `send`):

```bash
tinct-plugin-wob stop
tinct-plugin-wob start \
    --base-config ~/.config/wob/base.ini \
    --append-config ~/.config/wob/themes/tinct.ini
```

Or just send any percentage to trigger the auto-restart:

```bash
tinct-plugin-wob send 0
```

## Uninstall / revert

1. **Remove the autostart and keybind lines** from your compositor config:

       grep -n tinct-plugin-wob ~/.config/hypr/*.conf

   Drop those lines (or replace with raw `wob` invocations using your own base.ini if you want to keep wob running without tinct's overlay).

2. **Stop the running wob instance**:

       tinct-plugin-wob stop

3. **Delete the generated file**:

       rm ~/.config/wob/themes/tinct.ini
       rmdir --ignore-fail-on-non-empty ~/.config/wob/themes

4. **Reload to drop the styling**: wob is already stopped after step 2. Restart it (without `--append-config tinct.ini`) if you want plain-base wob back:

       wob -c ~/.config/wob/base.ini < /tmp/wob.fifo

5. **External state**: the wrapper creates runtime files under `$XDG_RUNTIME_DIR/wob/` (FIFO at `wob.fifo`, merged config cache, PID file). These are cleaned by `tinct-plugin-wob stop`. To purge any stragglers:

       rm -rf "${XDG_RUNTIME_DIR:-/run/user/$UID}/wob"

   To remove the plugin binary:

       tinct plugins uninstall wob
       # or, for a source build:
       rm ~/.local/bin/tinct-plugin-wob

## Flags

The plugin itself takes no `--wob.*` flags — the output path is fixed to `~/.config/wob/themes/tinct.ini`.

The `tinct-plugin-wob` wrapper subcommands accept:

| Subcommand | Flag | Description |
|---|---|---|
| `start` | `--base-config FILE` | Path to your base wob.ini (required for theming) |
| `start` | `--append-config FILE` | Overlay config (repeatable, last wins) |
| `send` | `--style NAME` | Force a style (`low`, `normal`, `warning`, `critical`) for this value |

Environment variables for the wrapper:

| Var | Default | Description |
|---|---|---|
| `WOB_BIN` | `wob` | Override the wob binary name on PATH |
| `WOB_PIPE` | `wob.fifo` | FIFO name under runtime dir |
| `WOB_MERGED_CONFIG` | `wob-merged.ini` | Cached merged-config filename |

## Colour role mapping

The generated `tinct.ini` defines a default style plus three named styles. Each is keyed off the percentage sent to `wob` (default is used unless a `--style` override is supplied).

| wob style | wob key | Tinct role |
|---|---|---|
| `default` (root) | `background_color` | `background` @ ~80% alpha (CC suffix) |
| `default` | `border_color` | `foreground` |
| `default` | `bar_color` | `accent1` |
| `default` | `overflow_background_color` | `background` @ ~80% alpha |
| `default` | `overflow_border_color` | `danger` |
| `default` | `overflow_bar_color` | `danger` |
| `[style.low]` | `bar_color` / `border_color` | `success` |
| `[style.normal]` | `bar_color` / `border_color` | `accent1` |
| `[style.critical]` | `bar_color` / `border_color` | `danger` |

Overflow colours (used when the value sent exceeds 100) fall back to `danger` across every style except `low`/`normal`, which use `warning`.

## Customising the template

```bash
tinct plugins templates dump -o wob -l ~/.config/tinct/templates/wob
$EDITOR ~/.config/tinct/templates/wob/tinct.ini.tmpl
```

Tinct prefers your version over the embedded default. See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### Plugin not found by tinct

```bash
which tinct-plugin-wob
tinct plugins list | grep wob
```

If the binary isn't on `$PATH`, reinstall via `tinct plugins install wob` or symlink your source build into `~/.local/bin/`.

### "wob is already running" but it isn't

The wrapper tracks state via a PID file under `$XDG_RUNTIME_DIR/wob/wob.pid`. If wob crashed without cleanup, remove the stale state:

```bash
rm -f "${XDG_RUNTIME_DIR:-/run/user/$UID}/wob/wob.pid"
tinct-plugin-wob start --base-config ~/.config/wob/base.ini --append-config ~/.config/wob/themes/tinct.ini
```

### Values don't display

Confirm wob is running, the FIFO exists, and your compositor allows wob's overlay layer:

```bash
tinct-plugin-wob status
ls -la "${XDG_RUNTIME_DIR:-/run/user/$UID}/wob/"   # expect prw------- wob.fifo
tinct-plugin-wob send 50                            # should pop the bar
```

If `send` succeeds but nothing visible appears, check your compositor's layer-shell support and that wob's `anchor`/`margin` aren't off-screen.

### Theme colours don't update after `tinct generate`

The wrapper restarts wob only on the *next* `send` call after the config rewrite. Either send any value (`tinct-plugin-wob send 0`) or restart explicitly with `tinct-plugin-wob stop && tinct-plugin-wob start …`.

### `--append-config` overrides not applying

Order matters — last `--append-config` wins. If you have multiple layers, the file passed last to `tinct-plugin-wob start` is the one whose values take precedence in the merged config.

## Related plugins

- [swayosd](./swayosd.md) — an alternative on-screen display for Wayland with GTK CSS theming.
- [dunst](./dunst.md) — desktop notifications styled in tinct colours.
