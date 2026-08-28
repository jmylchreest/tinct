---
title: keylightd-tray
sidebar_position: 7
plugin:
  type: output
  name: keylightd-tray
  category: bars-launchers
  source: external
  app: keylightd-tray
  app_url: 'https://github.com/jmylchreest/keylightd/tree/main/contrib/keylightd-tray'
  version: 0.1.0
  protocol_version: 0.3.0
  repository: 'https://github.com/jmylchreest/tinct'
  install: tinct plugins install keylightd-tray
  requires: []
  optional: []
  pattern: two-file
  default_output_dir: ~/.config/keylightd/keylightd-tray
  generated_files:
    - tinct-colours.css
    - tinct-custom.css
  reload:
    method: watch
    user_action_required: false
---

# keylightd-tray

Generates a CSS theme for [keylightd-tray](https://github.com/jmylchreest/keylightd/tree/main/contrib/keylightd-tray), the Elgato Key Light system-tray applet that ships with [keylightd](https://github.com/jmylchreest/keylightd). The output is a two-file pair — a palette file plus a theme application file — written into keylightd-tray's config directory. keylightd-tray watches its CSS files and re-renders the tray UI live as soon as the imported file changes.

## Installation

### Via the official plugin repository

```bash
tinct plugins install keylightd-tray
```

### Build from source

```bash
cd contrib/plugins/output/keylightd-tray
go build -ldflags "-X main.Version=0.1.0" -o tinct-plugin-keylightd-tray
install -m 0755 tinct-plugin-keylightd-tray ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-keylightd-tray
tinct-plugin-keylightd-tray --plugin-info | jq .
```

The plugin uses tinct's go-plugin RPC protocol and is discovered automatically once it's on `$PATH`.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o keylightd-tray
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colours.css` | `~/.config/keylightd/keylightd-tray/tinct-colours.css` | **Palette** — raw colour values exported as `--tinct-*` CSS custom properties |
| `tinct-custom.css` | `~/.config/keylightd/keylightd-tray/tinct-custom.css` | **Theme application** — `@import`s the palette and maps `--tinct-*` onto keylightd-tray's UI variables (`--bg-primary`, `--accent`, `--slider-track`, …) |

If `XDG_CONFIG_HOME` is set, both files land under `$XDG_CONFIG_HOME/keylightd/keylightd-tray/` instead.

## Integration

keylightd-tray loads `custom.css` from its config directory. To pull in the tinct theme, add a single `@import` line to that file (and leave it in place across regenerations — only the imported files are rewritten by tinct):

```bash
$EDITOR ~/.config/keylightd/keylightd-tray/custom.css
```

```css
@import url("tinct-custom.css");

/* Your additional customisations here */
```

The plugin prints a reminder in `--verbose` mode if `custom.css` is missing the import.

## Reload behaviour

### Automatic

keylightd-tray watches its CSS files via the standard GTK file-monitor and re-renders the tray UI as soon as the imported file changes. No signal, IPC or restart is needed — regenerate with tinct and the new colours appear immediately.

### Manual fallback

If the new colours don't appear, the watcher may have been disabled or the import may not be in place. Restart keylightd-tray:

```bash
pkill keylightd-tray && keylightd-tray &
```

## Uninstall / revert

1. **Remove the import line** from `~/.config/keylightd/keylightd-tray/custom.css`:

   ```bash
   $EDITOR ~/.config/keylightd/keylightd-tray/custom.css
   # delete the @import url("tinct-custom.css"); line
   ```

2. **Delete the generated files**:

   ```bash
   rm ~/.config/keylightd/keylightd-tray/tinct-colours.css \
      ~/.config/keylightd/keylightd-tray/tinct-custom.css
   ```

3. **Reload to drop the theme**: keylightd-tray re-renders on `custom.css` change. If it doesn't, restart it:

   ```bash
   pkill keylightd-tray && keylightd-tray &
   ```

4. **External state**: this plugin only writes to `~/.config/keylightd/keylightd-tray/`. To also remove the plugin binary:

   ```bash
   tinct plugins delete keylightd-tray
   # or, for a source build:
   rm ~/.local/bin/tinct-plugin-keylightd-tray
   ```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| (none) | — | Override the output directory via the `TINCT_PLUGIN_KEYLIGHTD_TRAY_OUTPUT_DIR` env var |

## Colour role mapping

### Palette file (`tinct-colours.css`)

The palette exports tinct's semantic roles as `--tinct-*` custom properties.

| CSS variable | Tinct role |
|---|---|
| `--tinct-background` | `background` |
| `--tinct-foreground` | `foreground` |
| `--tinct-surface` | `surface` |
| `--tinct-subtle` | `subtle` |
| `--tinct-overlay` | `overlay` |
| `--tinct-surface-container-lowest` | `surfaceContainerLowest` |
| `--tinct-surface-container-low` | `surfaceContainerLow` |
| `--tinct-surface-container` | `surfaceContainer` |
| `--tinct-surface-container-high` | `surfaceContainerHigh` |
| `--tinct-surface-container-highest` | `surfaceContainerHighest` |
| `--tinct-border-muted` | `borderMuted` |
| `--tinct-outline-variant` | `outlineVariant` |
| `--tinct-accent` | `accent1` (falls back to `accent`) |
| `--tinct-green` | `green` (falls back to `accent2`) |
| `--tinct-yellow` | `yellow` (falls back to `accent3`) |
| `--tinct-red` | `red` (falls back to `accent4`) |

### Theme application file (`tinct-custom.css`)

These map the palette onto keylightd-tray's UI variables.

| keylightd-tray variable | Palette variable | Used for |
|---|---|---|
| `--bg-primary` | `--tinct-background` | Window background |
| `--bg-secondary` / `--bg-tertiary` | `--tinct-surface` | Panel backgrounds |
| `--text-primary` | `--tinct-foreground` | Primary text |
| `--text-secondary` / `--text-muted` | `--tinct-subtle` | Secondary text |
| `--accent` | `--tinct-accent` | Sliders, interactive elements |
| `--success` | `--tinct-green` | Power "on" indicator |
| `--warning` | `--tinct-yellow` | Warning indicators |
| `--error` | `--tinct-red` | Error indicators |
| `--surface` | `--tinct-surface` | Surface fills |
| `--overlay` | `--tinct-overlay` | Overlay fills |
| `--slider-track` | `--tinct-surface-container-low` | Slider track background |
| `--input-bg` | `--tinct-surface-container-lowest` | Input field background |
| `--input-border` | `--tinct-border-muted` | Input field border |
| `--list-item-bg` | `--tinct-surface-container-lowest` | List item background |

## Customising the template

The plugin's CSS is generated in Go (no `text/template` template file), so customisation is best done by editing the generated `tinct-custom.css` mappings — keep `tinct-colours.css` untouched (it's regenerated) and add overrides in your own `custom.css` after the `@import` line:

```css
@import url("tinct-custom.css");

/* Override after the import */
:root {
  --accent: var(--tinct-accent);
  --slider-track: var(--tinct-surface-container);
}
```

## Troubleshooting

### Plugin not found by tinct

```bash
which tinct-plugin-keylightd-tray
tinct plugins list | grep keylightd-tray
```

If the binary isn't on `$PATH`, reinstall via `tinct plugins install keylightd-tray` or symlink the source build into `~/.local/bin/`.

### Theme not applied

```bash
cat ~/.config/keylightd/keylightd-tray/custom.css
```

Ensure the file contains `@import url("tinct-custom.css");`. If it doesn't, add it; the plugin prints this reminder when run with `--verbose`.

### Generated files missing

```bash
ls -l ~/.config/keylightd/keylightd-tray/tinct-*.css
tinct generate -i image -p ~/wallpaper.jpg -o keylightd-tray --verbose
```

Verbose mode shows the resolved output directory and any write errors.

## Related plugins

- [waybar](./waybar.md) — status bar that also uses CSS theming.
- [keylightd](https://github.com/jmylchreest/keylightd) — the upstream daemon and tray applet themed by this plugin.
