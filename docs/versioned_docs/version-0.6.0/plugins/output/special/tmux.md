---
title: tmux
sidebar_position: 1
plugin:
  type: output
  name: tmux
  category: special
  source: builtin
  app: tmux
  app_url: 'https://github.com/tmux/tmux'
  requires:
    - tmux
  optional: []
  pattern: two-file
  default_output_dir: ~/.config/tmux
  generated_files:
    - tinct-palette.conf
    - tinct.conf
  reload:
    method: ipc
    command: tmux source-file <output-dir>/tinct.conf
    user_action_required: false
---

# tmux

Generates a colour theme for [tmux](https://github.com/tmux/tmux), the terminal multiplexer. Two files are produced: a pure palette (`tinct-palette.conf`) exposing every tinct role as a `@tinct-<role>` user option, and a main config (`tinct.conf`) that sources the palette and applies opinionated styling to the status bar, window tabs, panes, message line, copy mode and clock. You source whichever fits your workflow from your existing `tmux.conf`.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o tmux` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o tmux
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-palette.conf` | `~/.config/tmux/tinct-palette.conf` | **Palette only.** A flat list of `set -gq @tinct-<role> "#..."` declarations covering every tinct role. Reference these from your own status line via `#{@tinct-bg}`, `#{@tinct-accent1}`, etc. Contains no styling. |
| `tinct.conf` | `~/.config/tmux/tinct.conf` | **Main config.** `source-file`s the palette above, then applies opinionated styling to `status-style`, `window-status-*`, `pane-*-border-style`, `message-style`, `mode-style`, and copy/clock styles using the `#{@tinct-*}` user options. |

The `@tinct-*` user options are the integration point — both for the bundled styling and for any custom `status-left` / `status-right` format strings you wire up yourself.

## Integration

tmux doesn't auto-load drop-in files, so you source one of the two generated files from your existing `tmux.conf`. Pick whichever fits:

**Use the bundled styling.** Sources the palette and applies tinct's opinionated styles:

```tmux
source-file ~/.config/tmux/tinct.conf
```

**Use only the palette.** Source just the variables, then write your own status/style lines that reference `#{@tinct-*}`:

```tmux
source-file ~/.config/tmux/tinct-palette.conf

# your own styling, e.g.:
set -g status-style "bg=#{@tinct-surface},fg=#{@tinct-foreground}"
set -g pane-active-border-style "fg=#{@tinct-accent1}"
```

Place either line near the end of `tmux.conf` so it overrides earlier settings. After saving, the plugin reloads tmux for you (see below).

## Reload behaviour

### Automatic

After generation, the plugin checks for a running tmux server (`tmux ls`). If one is found, it runs `tmux source-file ~/.config/tmux/tinct.conf` so existing sessions pick up the new palette immediately. When no server is running, the plugin silently skips the reload — there's nothing to reload into.

### Manual fallback

If the automatic reload didn't take (e.g. you weren't in tmux at generate time, or your `tmux.conf` overrides styles defined later), source it yourself from any pane:

```bash
tmux source-file ~/.config/tmux/tmux.conf
```

…or, with your prefix bound to `Ctrl+b` (default), press **prefix-r** if your `tmux.conf` has the conventional reload binding:

```tmux
bind r source-file ~/.config/tmux/tmux.conf \; display "Config reloaded"
```

Existing windows pick up the new styles immediately; no restart required.

## Uninstall / revert

1. **Remove the source line(s)** from your `~/.config/tmux/tmux.conf`:

   Delete whichever you added — `source-file ~/.config/tmux/tinct.conf` and/or `source-file ~/.config/tmux/tinct-palette.conf`.

2. **Delete the generated files**:

   ```bash
   rm ~/.config/tmux/tinct.conf ~/.config/tmux/tinct-palette.conf
   ```

3. **Reload to drop the theme**:

   ```bash
   tmux source-file ~/.config/tmux/tmux.conf
   ```

   tmux resets to its built-in default styling (or whatever your `tmux.conf` defines directly).

4. **External state**: this plugin only writes to `~/.config/tmux/`. No further cleanup is required. The `@tinct-*` user options disappear from tmux's option table on the next config reload.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--tmux.output-dir` | `~/.config/tmux` | Override the output directory |

## Colour role mapping

### User options exposed by `tinct-palette.conf`

Every tinct role is declared as a `@tinct-<role>` user option. Use them in your own format strings via `#{@tinct-<role>}`.

| User option | Tinct role |
|---|---|
| `@tinct-background` / `@tinct-foreground` | `background` / `foreground` |
| `@tinct-background-muted` / `@tinct-foreground-muted` | `backgroundMuted` / `foregroundMuted` |
| `@tinct-surface` / `@tinct-on-surface` | `surface` / `onSurface` |
| `@tinct-surface-variant` / `@tinct-on-surface-variant` | `surfaceVariant` / `onSurfaceVariant` |
| `@tinct-surface-container-lowest` … `@tinct-surface-container-highest` | `surfaceContainerLowest` … `surfaceContainerHighest` |
| `@tinct-outline` / `@tinct-outline-variant` / `@tinct-border` | `outline` / `outlineVariant` / `border` |
| `@tinct-accent1` … `@tinct-accent4` | `accent1` … `accent4` |
| `@tinct-on-accent1` … `@tinct-on-accent4` | `onAccent1` … `onAccent4` |
| `@tinct-danger` / `@tinct-on-danger` | `danger` / `onDanger` |
| `@tinct-warning` / `@tinct-on-warning` | `warning` / `onWarning` |
| `@tinct-success` / `@tinct-on-success` | `success` / `onSuccess` |
| `@tinct-info` / `@tinct-on-info` | `info` / `onInfo` |

### Styles applied by `tinct.conf`

| tmux option | Background / Foreground |
|---|---|
| `status-style` | `@tinct-surface` / `@tinct-foreground` |
| `status-left-style` | `@tinct-accent1` / `@tinct-on-accent1` (bold) |
| `status-right-style` | `@tinct-surface-container-high` / `@tinct-foreground` |
| `window-status-style` | `@tinct-surface` / `@tinct-foreground-muted` |
| `window-status-current-style` | `@tinct-accent1` / `@tinct-on-accent1` (bold) |
| `window-status-activity-style` | `@tinct-surface` / `@tinct-warning` |
| `window-status-bell-style` | `@tinct-surface` / `@tinct-danger` (bold) |
| `pane-border-style` | fg `@tinct-outline` |
| `pane-active-border-style` | fg `@tinct-accent1` |
| `message-style` | `@tinct-accent2` / `@tinct-on-accent2` |
| `message-command-style` | `@tinct-accent3` / `@tinct-on-accent3` |
| `mode-style` | `@tinct-accent1` / `@tinct-on-accent1` |
| `clock-mode-colour` | `@tinct-accent2` |
| `copy-mode-match-style` | `@tinct-warning` / `@tinct-on-warning` |
| `copy-mode-current-match-style` | `@tinct-accent1` / `@tinct-on-accent1` |

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o tmux -l ~/.config/tinct/templates/tmux
```

This writes `~/.config/tinct/templates/tmux/tinct.conf.tmpl` and `~/.config/tinct/templates/tmux/tinct-palette.conf.tmpl`. Tinct uses your versions in preference to the embedded defaults; override one, both, or neither.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Nothing changed after `tinct generate`

You either forgot to source the file or forgot to reload. Check both:

```bash
grep tinct.conf ~/.config/tmux/tmux.conf
tmux source-file ~/.config/tmux/tmux.conf
```

If your `tmux.conf` sets `status-style` (or other themed options) *after* the `source-file` line, those explicit values win — move the `source-file` line to the bottom of `tmux.conf`.

### Status bar looks fine but `#{@tinct-accent1}` is empty in my custom format

User options are set with `set -gq @tinct-*` (the `q` is "quiet — no error if it exists"). They live in tmux's global option table and are scoped to the server. If you start a brand-new tmux server after deleting the palette file, the options won't be set — that's expected. Re-source the palette (or re-generate and re-source).

### "ambiguous option" or "unknown option" when sourcing

The templates target a recent tmux (>= 3.2 — `@tinct-*` user options need 3.2+, the styles themselves work back to 3.0). On older tmux versions, source the palette manually and inline the colours into your style lines instead of using `#{@tinct-*}`.

## Related plugins

- [alacritty](../terminals/alacritty.md), [ghostty](../terminals/ghostty.md), [kitty](../terminals/kitty.md), [warp](../terminals/warp.md), [wezterm](../terminals/wezterm.md) — the terminal you're running tmux inside.
- [neovim](../editors/neovim.md) — pair the multiplexer styling with an editor theme that uses the same palette.
