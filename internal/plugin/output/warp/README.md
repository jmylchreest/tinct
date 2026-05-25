---
title: warp
sidebar_position: 5

plugin:
  type: output
  name: warp
  category: terminals
  source: builtin
  app: Warp
  app_url: https://www.warp.dev/
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.local/share/warp-terminal/themes
  generated_files: [tinct.yaml]
  reload:
    method: none
    user_action_required: true
---

# warp

Generates a YAML custom-theme file for [Warp](https://www.warp.dev/), the Rust-based AI-first terminal. The file lands in Warp's user themes directory and shows up in the appearance picker labelled "Tinct".

## Installation

Built into tinct — nothing to install separately. `tinct generate -o warp` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o warp
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.yaml` | `~/.local/share/warp-terminal/themes/tinct.yaml` | Single-file Warp custom theme — `name: Tinct`, accent / cursor / background / foreground, polarity hint, and the full normal+bright ANSI block. |

On macOS the path is `~/.warp/themes/tinct.yaml` (Warp's documented per-user themes location, not Application Support).

## Generated format

```yaml
name: Tinct
accent: '#7aa2f7'
cursor: '#7aa2f7'
background: '#1a1b26'
foreground: '#c0caf5'
details: lighter
terminal_colors:
  normal:
    black:   '#15161e'
    red:     '#f7768e'
    green:   '#9ece6a'
    yellow:  '#e0af68'
    blue:    '#7aa2f7'
    magenta: '#bb9af7'
    cyan:    '#7dcfff'
    white:   '#a9b1d6'
  bright:
    black:   '#414868'
    red:     '#f7768e'
    green:   '#9ece6a'
    yellow:  '#e0af68'
    blue:    '#7aa2f7'
    magenta: '#bb9af7'
    cyan:    '#7dcfff'
    white:   '#c0caf5'
```

## Integration

Warp's built-in file watcher scans the themes directory while the app is running, so no config edit is needed. To activate the theme:

- **Settings → Appearance → Themes → Tinct**, or
- Command Palette (`Cmd+P` / `Ctrl+P`) → "Switch Theme" → **Tinct**.

Warp remembers the selection per-workspace.

## Reload behaviour

### Automatic

None from the plugin — Warp watches the themes directory itself. New theme files appear in the picker within a few seconds; re-runs of `tinct generate` overwrite `tinct.yaml` in place and (if Tinct is already the active theme) Warp re-renders open panes with the new colours.

### Manual fallback

If the new theme doesn't show up:

1. Open Settings → Appearance → Themes. The list refreshes on focus.
2. If still missing, restart Warp.

## Uninstall / revert

1. **Remove the config line**: none added. The active theme is a Warp setting, not a config file edit — switch back to your previous theme via Settings → Appearance → Themes.

2. **Delete the generated file**:

   ```bash
   rm ~/.local/share/warp-terminal/themes/tinct.yaml   # Linux
   # or
   rm ~/.warp/themes/tinct.yaml                        # macOS
   ```

3. **Reload to drop the theme**:

   Pick a different theme in Settings → Appearance → Themes. Warp drops "Tinct" from the picker the next time it rescans the directory.

4. **External state**: this plugin only writes to Warp's themes directory. Warp itself stores the "currently selected theme" elsewhere — deleting the file leaves that setting pointing at a missing theme; Warp falls back to its default the next time the setting is read.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--warp.output-dir` | `~/.local/share/warp-terminal/themes` (Linux) / `~/.warp/themes` (macOS) | Override the output directory |

## Colour role mapping

| Warp key | Tinct role |
|---|---|
| `accent` | `accent1` |
| `cursor` | `accent1` |
| `background` | `background` |
| `foreground` | `foreground` |
| `details` | `lighter` (dark themes) / `darker` (light themes) — Warp uses this to pick its surface contrast |
| `terminal_colors.normal.black`..`white` | ANSI `black`..`white` |
| `terminal_colors.bright.black`..`white` | ANSI `bright-black`..`bright-white` |

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o warp -l ~/.config/tinct/templates/warp
```

This writes `~/.config/tinct/templates/warp/tinct.yaml.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### "Tinct" doesn't appear in the themes picker

Check the path. Warp loads themes from `~/.warp/themes/` on macOS and `~/.local/share/warp-terminal/themes/` on Linux. If you set `--warp.output-dir` or `TINCT_PLUGIN_WARP_OUTPUT_DIR` to something else, Warp won't find the file. Verify with:

```bash
ls -l ~/.warp/themes ~/.local/share/warp-terminal/themes 2>/dev/null
```

### Theme renders, but the accent colour is washed out

Warp uses the `details` field (`lighter` or `darker`) to derive its UI surface contrast. The plugin auto-selects the right value from the resolved theme type. If your palette is borderline (dark colours on a near-white background, or vice versa), nudge it by re-running tinct with a stronger source image.

### Warp ignores edits to `tinct.yaml`

Warp parses YAML strictly. If you've customised the template and your version contains a syntax error, Warp logs it to its developer console (Help → Developer → Logs) and silently keeps the previous theme.

## Related plugins

- [alacritty](../alacritty/README.md), [ghostty](../ghostty/README.md), [kitty](../kitty/README.md), [wezterm](../wezterm/README.md) — other modern terminals.
- [tmux](../tmux/README.md) — multiplexer status-bar styling for users who run tmux inside Warp.
