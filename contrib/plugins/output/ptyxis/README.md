---
title: ptyxis
sidebar_position: 8

plugin:
  type: output
  name: ptyxis
  category: terminals
  source: external
  app: Ptyxis
  app_url: https://gitlab.gnome.org/chergert/ptyxis
  version: 0.1.0
  protocol_version: 0.3.0
  repository: https://github.com/jmylchreest/tinct
  install: tinct plugins install ptyxis
  requires: []
  optional: [ptyxis]
  pattern: single-file
  default_output_dir: ~/.local/share/org.gnome.Ptyxis/palettes
  generated_files: [tinct.palette]
  reload:
    method: watch
    user_action_required: true
---

# ptyxis

Generates a [Ptyxis](https://gitlab.gnome.org/chergert/ptyxis) terminal colour palette. Ptyxis is GNOME's libadwaita-native terminal; it stores user palettes as `.palette` ini files in a known directory and watches that directory live via `GFileMonitor`. The plugin writes a single `tinct.palette` file into Ptyxis's palettes directory — the new palette appears in **Preferences → Appearance → Palette** without restarting Ptyxis.

When tinct is run with both primary and alternate themes (`--dual-theme`), the plugin emits a dual-mode palette with `[Light]` and `[Dark]` sections; Ptyxis switches between them based on the system colour preference. Otherwise a single-section palette is written and Ptyxis derives its mode from the background luminance.

> **Note.** This plugin was previously built into tinct and was migrated to an external plugin in v0.2.4. Behaviour is unchanged — same palette format, same install detection, same live reload.

## Installation

### Via the official plugin repository

```bash
tinct plugins install ptyxis
```

### Build from source

```bash
cd contrib/plugins/output/ptyxis
go build -ldflags "-X main.Version=0.1.0" -o tinct-plugin-ptyxis
install -m 0755 tinct-plugin-ptyxis ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-ptyxis
tinct-plugin-ptyxis --plugin-info | jq .
```

The plugin uses tinct's go-plugin RPC protocol and is discovered automatically once it's on `$PATH`.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o ptyxis
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.palette` | `~/.local/share/org.gnome.Ptyxis/palettes/tinct.palette` | Single-mode or dual-mode Ptyxis palette (ini format) |

### Install-location detection

The plugin auto-detects whether Ptyxis is installed natively or via Flatpak and writes to the corresponding palettes directory:

- **Native**: `$XDG_DATA_HOME/org.gnome.Ptyxis/palettes` (default `~/.local/share/org.gnome.Ptyxis/palettes`).
- **Flatpak**: `~/.var/app/app.devsuite.Ptyxis/data/org.gnome.Ptyxis/palettes`.

Override with `--ptyxis.output-dir=/path` or `TINCT_PLUGIN_PTYXIS_OUTPUT_DIR=/path`.

## Integration

**No manual configuration required.** Ptyxis lists every `.palette` file in its palettes directory under **Preferences → Appearance → Palette**. To activate the theme:

1. Open Ptyxis preferences.
2. Go to **Appearance → Palette**.
3. Select **Tinct**.

In dual-mode (`--dual-theme`), Ptyxis honours the system colour preference and picks the `[Light]` or `[Dark]` section automatically.

## Reload behaviour

### Automatic

Ptyxis's `PtyxisUserPalettes` watches the palettes directory via `GFileMonitor`. New or modified palettes are picked up live — there's no signal, IPC or restart involved. After the first time you select **Tinct**, subsequent regenerations are visible immediately if a Ptyxis window is open.

### Manual fallback

The first time the palette is written, the user has to **select it** in **Preferences → Appearance → Palette → Tinct**. After that, subsequent regenerations are automatic. If a regeneration doesn't appear:

- Confirm the file was written: `ls -l ~/.local/share/org.gnome.Ptyxis/palettes/tinct.palette` (or the Flatpak path).
- Restart Ptyxis: close and reopen.

## Uninstall / revert

1. **Remove the palette selection** in **Preferences → Appearance → Palette** — choose a different palette. (No config-file line was added — Ptyxis stores the palette choice in `gsettings`.)

2. **Delete the generated file**:

   ```bash
   # Native install
   rm ~/.local/share/org.gnome.Ptyxis/palettes/tinct.palette
   # Flatpak install
   rm ~/.var/app/app.devsuite.Ptyxis/data/org.gnome.Ptyxis/palettes/tinct.palette
   ```

3. **Reload**: Ptyxis's file watcher removes the palette from the picker live. If it doesn't, restart Ptyxis.

4. **External state**: this plugin only writes to Ptyxis's palettes directory. To also remove the plugin binary:

   ```bash
   tinct plugins uninstall ptyxis
   # or, for a source build:
   rm ~/.local/bin/tinct-plugin-ptyxis
   ```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ptyxis.output-dir` / `-o` | auto-detected | Override the output directory (native or Flatpak palettes dir) |

The `TINCT_PLUGIN_PTYXIS_OUTPUT_DIR` environment variable serves the same purpose as the flag.

## Colour role mapping

Ptyxis's `.palette` format is a small INI file. The plugin maps tinct's semantic roles onto Ptyxis's foreground/background/cursor keys and the standard 16-colour ANSI block.

| Ptyxis key | Tinct role |
|---|---|
| `Foreground` | `foreground` |
| `Background` | `background` |
| `Cursor` | `accent1` |
| `Color0`–`Color7` | ANSI `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white` |
| `Color8`–`Color15` | ANSI bright variants |
| `BellForeground` / `BellBackground` | `foreground` / `warning` |
| `RemoteForeground` / `RemoteBackground` | `foreground` / `info` |
| `SuperuserForeground` / `SuperuserBackground` | `foreground` / `danger` |

The ANSI block uses tinct's perceptual ANSI matcher (`ansi` template func) so colours sit close to standard expectations regardless of palette source.

## Customising the template

```bash
mkdir -p ~/.config/tinct/templates/ptyxis
tinct plugins templates dump -o ptyxis -l ~/.config/tinct/templates/ptyxis
$EDITOR ~/.config/tinct/templates/ptyxis/tinct.palette.tmpl
```

Tinct prefers your version over the embedded default. See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### Plugin skipped: "Ptyxis terminal is not installed"

```bash
which ptyxis
ls ~/.var/app/app.devsuite.Ptyxis 2>/dev/null
```

The plugin requires either the `ptyxis` binary on `$PATH` or the Flatpak data directory to be present. If you've installed Ptyxis in a non-standard location, set `--ptyxis.output-dir` or `TINCT_PLUGIN_PTYXIS_OUTPUT_DIR` directly.

### Palette not appearing in Ptyxis preferences

```bash
ls -l ~/.local/share/org.gnome.Ptyxis/palettes/tinct.palette
# or, for Flatpak:
ls -l ~/.var/app/app.devsuite.Ptyxis/data/org.gnome.Ptyxis/palettes/tinct.palette
```

If the file is present, open **Preferences → Appearance → Palette** in Ptyxis — the watcher should add the entry live. If Ptyxis is running under Flatpak but the plugin wrote to the native path (or vice versa), override the output directory.

### Wrong variant (light/dark) chosen in dual-mode

Ptyxis follows the system colour preference. If you want to lock a specific variant, generate without `--dual-theme` so the plugin writes a single-mode palette.

## Related plugins

- [ghostty](../../../../internal/plugin/output/ghostty/README.md), [kitty](../../../../internal/plugin/output/kitty/README.md), [foot](../../../../internal/plugin/output/foot/README.md), [alacritty](../../../../internal/plugin/output/alacritty/README.md), [wezterm](../../../../internal/plugin/output/wezterm/README.md) — other terminals tinct themes.
