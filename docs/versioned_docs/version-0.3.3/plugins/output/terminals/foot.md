---
title: foot
sidebar_position: 2
plugin:
  type: output
  name: foot
  category: terminals
  source: builtin
  app: foot
  app_url: 'https://codeberg.org/dnkl/foot'
  requires:
    - foot
  optional: []
  pattern: two-file
  default_output_dir: ~/.config/foot
  generated_files:
    - tinct-colors.ini
    - tinct.ini
  reload:
    method: signal
    command: pkill -SIGUSR1 -x foot
    user_action_required: false
---

# foot

Generates a colour theme for [foot](https://codeberg.org/dnkl/foot), the fast, lightweight Wayland terminal emulator from the author of [fuzzel](https://codeberg.org/dnkl/fuzzel). The plugin emits two files: a `tinct-colors.ini` palette fragment with just the `[colors]` block (suitable for `include=` from your main `foot.ini`), and a `tinct.ini` example main config with reasonable font/scrollback defaults so first-time users can rename it directly to `foot.ini`.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o foot` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o foot
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colors.ini` | `~/.config/foot/tinct-colors.ini` | **Palette.** Just the `[colors]` section — every foot colour key in `RRGGBB` (hex without `#`) format. Include this from your `foot.ini`. |
| `tinct.ini` | `~/.config/foot/tinct.ini` | **Example main config.** An `include=` line for the palette plus reasonable font, padding, and scrollback defaults. Rename to `foot.ini` if you don't already have one. |

## Generated format

```ini
[colors]
foreground=cdd6f4
background=1e1e2e

selection-foreground=1e1e2e
selection-background=89b4fa

urls=89b4fa

regular0=1e1e2e
regular1=f38ba8
regular2=a6e3a1
regular3=f9e2af
regular4=89b4fa
regular5=f5c2e7
regular6=94e2d5
regular7=cdd6f4

bright0=313244
bright1=f38ba8
bright2=a6e3a1
bright3=f9e2af
bright4=89b4fa
bright5=f5c2e7
bright6=94e2d5
bright7=a6adc8
```

## Integration

foot supports an `include=` directive that pulls another `.ini` fragment into the running config. Add one line to your `~/.config/foot/foot.ini`:

```ini
include=~/.config/foot/tinct-colors.ini
```

Keep your structural settings (font, padding, keybindings, scrollback) in `foot.ini` and let the included file own the `[colors]` block.

If you don't have a `foot.ini` yet, **rename the generated `tinct.ini`** instead of editing it by hand:

```bash
mv ~/.config/foot/tinct.ini ~/.config/foot/foot.ini
```

The example `tinct.ini` already contains the `include=` line at the top, plus sensible defaults for font, padding and scrollback.

## Reload behaviour

### Automatic

After writing the files, the plugin sends `SIGUSR1` to every running `foot` process (foot's documented colour-reload signal). The reload is in-place — open terminals keep their content and scrollback; only the colours change.

The signal is sent silently at normal verbosity; the only user-facing message at normal verbosity is the reminder to add the `include=` line on your first run. Pass `-v` to see the signal broadcast in the log.

If no `foot` is running at generate time, the signal is a no-op and the next foot you launch picks up the new palette from the file on disk.

### Manual fallback

If for any reason the auto-reload didn't happen (foot wasn't running, or you generated on a different machine):

```bash
pkill -SIGUSR1 -x foot
```

`-x` matches the exact process name so this doesn't accidentally signal anything else.

## Uninstall / revert

1. **Remove the include line** from your `~/.config/foot/foot.ini`:

   ```bash
   sed -i '/tinct-colors\.ini/d' ~/.config/foot/foot.ini
   ```

   If you renamed `tinct.ini` to `foot.ini` directly, instead replace it with whatever config you want to use going forward (or delete it to revert foot to its defaults).

2. **Delete the generated files**:

   ```bash
   rm ~/.config/foot/tinct-colors.ini ~/.config/foot/tinct.ini
   ```

3. **Reload to drop the theme**:

   ```bash
   pkill -SIGUSR1 -x foot
   ```

4. **External state**: this plugin only writes to `~/.config/foot/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--foot.output-dir` | `~/.config/foot` | Override the output directory |

## Colour role mapping

All values are emitted in foot's `RRGGBB` (6-digit hex, no `#` prefix) format.

| foot key (`[colors]`) | Tinct role | Notes |
|---|---|---|
| `foreground` | `foreground` | Default text colour |
| `background` | `background` | Terminal background |
| `selection-foreground` | `background` | Inverted so selection stays legible |
| `selection-background` | `accent1` | Primary accent makes selections obvious |
| `urls` | `accent1` | Highlighted URL underline colour |
| `regular0` | `background` | ANSI black (matches bg) |
| `regular1` | `danger` | ANSI red |
| `regular2` | `success` | ANSI green |
| `regular3` | `warning` | ANSI yellow |
| `regular4` | `accent1` | ANSI blue |
| `regular5` | `accent2` | ANSI magenta |
| `regular6` | `accent3` | ANSI cyan |
| `regular7` | `foreground` | ANSI white |
| `bright0` | `backgroundMuted` | Bright black (subtle lift off bg) |
| `bright1` | `danger` | Bright red |
| `bright2` | `success` | Bright green |
| `bright3` | `warning` | Bright yellow |
| `bright4` | `accent1` | Bright blue |
| `bright5` | `accent2` | Bright magenta |
| `bright6` | `accent3` | Bright cyan |
| `bright7` | `foregroundMuted` | Bright white |

## Customising the template

Extract the default templates to override them:

```bash
tinct plugins templates dump -o foot -l ~/.config/tinct/templates/foot
```

This creates `~/.config/tinct/templates/foot/tinct-colors.ini.tmpl` and `tinct.ini.tmpl`. Tinct uses your versions in preference to the embedded defaults.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Colours unchanged in already-open foot windows

The auto-reload happens via `SIGUSR1`; if a window pre-dates the signal (e.g. it was launched after `tinct generate` finished writing but before the signal was delivered), trigger a reload manually:

```bash
pkill -SIGUSR1 -x foot
```

### `unknown option: include=`

foot has supported `include=` since 1.6 (2020). Upgrade to a current release; every Wayland-shipping distro carries one.

### Palette loaded but colours look wrong / partial

Your `foot.ini` probably has its own `[colors]` keys *after* the `include=` line. foot processes settings in order and the later value wins. Move the `include=` line below your `[colors]` block, or delete the overlapping keys from `foot.ini`.

### `foot: command not found`

The plugin lists `foot` as a required binary. Install it via your distro's package manager (`pacman -S foot`, `apt install foot`, etc.) or build from the [upstream repository](https://codeberg.org/dnkl/foot).

## Related plugins

- [alacritty](./alacritty.md), [ghostty](./ghostty.md), [kitty](./kitty.md), [wezterm](./wezterm.md) — other GPU-accelerated terminals.
- [fuzzel](../bars-launchers/fuzzel.md) — Wayland launcher from the same author; pairs naturally with foot.
- [tmux](../special/tmux.md) — match the terminal palette with your multiplexer status bar.
