---
title: dunstify
sidebar_position: 8
plugin:
  type: output
  name: dunstify
  category: bars-launchers
  source: external
  app: dunstify / notify-send
  app_url: 'https://github.com/dunst-project/dunst'
  version: 0.0.0
  protocol_version: 0.3.0
  repository: 'https://github.com/jmylchreest/tinct'
  install: tinct plugins install dunstify
  requires: []
  optional:
    - dunstify
    - notify-send
  default_output_dir: ''
  generated_files: []
  reload:
    method: none
    user_action_required: false
---

# dunstify

Pops a desktop notification via [`dunstify`](https://github.com/dunst-project/dunst) (or [`notify-send`](https://gitlab.gnome.org/GNOME/libnotify)) every time `tinct generate` finishes. It writes no files of its own — it consumes the list of files written by the other output plugins in the run and shows a one-line summary. The original purpose is as a **reference implementation** of tinct's go-plugin RPC protocol, demonstrating that a plugin doesn't have to produce theme files to be useful.

## Installation

### Via the official plugin repository

```bash
tinct plugins install dunstify
```

### Build from source

```bash
cd contrib/plugins/output/dunstify
go build -ldflags "-X main.Version=0.0.0" -o tinct-plugin-dunstify
install -m 0755 tinct-plugin-dunstify ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-dunstify
tinct-plugin-dunstify --plugin-info | jq .
```

### Install a notification helper

One of `dunstify` or `notify-send` must be on `$PATH`:

| Distro | dunst (includes `dunstify`) | libnotify (provides `notify-send`) |
|---|---|---|
| Arch | `sudo pacman -S dunst` | `sudo pacman -S libnotify` |
| Debian/Ubuntu | `sudo apt install dunst` | `sudo apt install libnotify-bin` |
| Fedora | `sudo dnf install dunst` | `sudo dnf install libnotify` |

If neither is present the plugin skips cleanly — no error, no notification.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o hyprland,kitty,dunstify
```

The desktop notification appears once `tinct generate` finishes:

```
Theme Generated
Generated 8 theme files
```

The icon (`preferences-desktop-theme`), urgency (`low`), and timeout (5 s) are baked into the plugin — override them by editing `main.go` and rebuilding from source.

## Generated files

None. This plugin's effect is the notification itself — there is nothing to write to disk and nothing to clean up later.

## Integration

No configuration required. Add `dunstify` to the `-o` plugin list whenever you want a notification when generation finishes:

```bash
tinct generate -i image -p ~/wallpaper.jpg -o ghostty,waybar,dunstify
```

Or enable it permanently so it runs alongside every generate:

```bash
tinct plugins enable output:dunstify
```

## Reload behaviour

### Automatic

Not applicable — the plugin's *only* side-effect is the notification, dispatched in `PostExecute` after the other plugins have finished writing. There is no separate "reload" step.

### Manual fallback

If the notification didn't appear:

1. Confirm the notification daemon is actually running — `pgrep -x dunst` for dunst, otherwise check whichever daemon your DE provides.
2. Confirm one of the helper binaries is on `$PATH` — `command -v dunstify || command -v notify-send`.
3. Re-run `tinct generate ... -o dunstify --verbose` and check the daemon's log.

## Uninstall / revert

1. **Remove the config line**: this plugin has no config to remove. If you opted in permanently with `tinct plugins enable output:dunstify`, disable it:

       tinct plugins disable output:dunstify

2. **Delete the generated files**: none — this plugin writes nothing to disk. Nothing to remove.

3. **Reload to drop the styling**: not applicable. Once disabled, the plugin stops dispatching notifications immediately.

4. **External state**: this plugin only fires desktop notifications via the system notification daemon. No files outside `tinct generate`'s working set are touched. To remove the plugin binary:

       tinct plugins uninstall dunstify
       # or, for a source build:
       rm ~/.local/bin/tinct-plugin-dunstify

## Flags

This plugin has no configurable flags. To tune the notification (urgency, timeout, icon, body text), edit `main.go` and rebuild from source. See [Customising](#customising) below.

## Colour role mapping

Not applicable. This plugin does not consume any colour values from the palette — it ignores `palette.colours` entirely and reports only the count of files written by other plugins in the run.

## Customising

The dispatched notification is hard-coded. To change it:

1. Edit `contrib/plugins/output/dunstify/main.go`.
2. Rebuild: `go build -o tinct-plugin-dunstify`.
3. Reinstall: `install -m 0755 tinct-plugin-dunstify ~/.local/bin/`.

Common knobs:

| Knob | Where | Default |
|---|---|---|
| Urgency | `-u low` arg to `exec.CommandContext` | `low` |
| Timeout | `-t 5000` arg (ms) | `5000` (5 s) |
| Icon | `icon` constant in `PostExecute` | `preferences-desktop-theme` |
| Body text | `body` formatted in `PostExecute` | `Generated N theme file(s)` |

## Troubleshooting

### Plugin skips: "Neither dunstify nor notify-send found on $PATH"

Install one of them:

```bash
sudo pacman -S dunst       # Arch (provides dunstify)
sudo apt install dunst     # Debian/Ubuntu
sudo pacman -S libnotify   # Arch (provides notify-send)
```

### Plugin runs but no notification appears

Your notification daemon may not be running. For dunst:

```bash
pgrep -x dunst || dunst &
```

For GNOME, KDE, sway-notification-center, etc., check the relevant daemon. `notify-send` honours whatever DBus notification service is currently registered.

### Plugin runs but `--plugin-info` returns the wrong name

If you built from source without setting `-ldflags "-X main.Version=…"`, the binary reports `version: 0.0.0`. That's cosmetic and doesn't affect operation — set the ldflag if you care.

## Related plugins

- [dunst](./dunst.md) — the built-in dunst theming plugin (writes `~/.config/dunst/dunstrc.d/`). Pair `dunst` (style the daemon) with `dunstify` (announce regenerations) if you want both.

