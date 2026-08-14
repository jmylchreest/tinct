---
title: noctalia
sidebar_position: 10

plugin:
  type: output
  name: noctalia
  category: bars-launchers
  source: external
  app: Noctalia
  app_url: https://github.com/noctalia-dev/noctalia-shell
  version: 0.1.0
  protocol_version: 0.3.0
  repository: https://github.com/jmylchreest/tinct
  install: tinct plugins install noctalia
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/noctalia/palettes
  generated_files: [tinct.json]
  reload:
    method: watch
    user_action_required: false
---

# noctalia

Generates a custom colour palette for the [Noctalia shell](https://github.com/noctalia-dev/noctalia-shell), a Wayland desktop shell (bar, dock, launcher, control center, notifications) for compositors like niri and Hyprland.

Noctalia ships its **own** Material You palette generator that derives colours from your wallpaper. This plugin doesn't replace that — it writes Noctalia's *custom palette* file so that, when you switch Noctalia to `source = "custom"`, its colours match the rest of your tinct-themed desktop (terminals, niri, editors, …) instead of being generated independently.

## Installation

```bash
tinct plugins install noctalia
```

Or build from this repo and add the binary manually:

```bash
cd contrib/plugins/output/noctalia && go build -o tinct-plugin-noctalia .
tinct plugins add ./tinct-plugin-noctalia
```

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o noctalia
```

Then point Noctalia at the generated palette **once** (see [Integration](#integration)).

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.json` | `~/.config/noctalia/palettes/tinct.json` | **Custom palette.** A `dark` and `light` variant, each with the 16 Material-3 roles Noctalia uses plus a terminal colour block. |

## Integration

Noctalia selects where its palette comes from in its config. Set the source to `custom` and name the palette `tinct` (the file is `tinct.json`, referenced without the extension).

**Noctalia v5** (current, native — TOML at `~/.config/noctalia/*.toml`):

```toml
[theme]
mode           = "dark"      # dark | light | auto
source         = "custom"    # use the palette file instead of wallpaper generation
custom_palette = "tinct"     # reads ~/.config/noctalia/palettes/tinct.json
```

**Noctalia v4** (Quickshell/QML) stores schemes at `~/.config/noctalia/colorschemes/<name>/<name>.json` and selects them in **Settings → Color Scheme**. Point `--noctalia.output-dir` accordingly if you run v4 (see [Flags](#flags)).

## Colour role mapping

tinct's role model is already Material Design 3, so the mapping is close to 1:1:

| Noctalia role | Tinct role |
|---|---|
| `mPrimary` / `mOnPrimary` | `accent1` / `onAccent1` |
| `mSecondary` / `mOnSecondary` | `accent2` / `onAccent2` |
| `mTertiary` / `mOnTertiary` | `accent3` / `onAccent3` |
| `mError` / `mOnError` | `danger` / `onDanger` |
| `mSurface` / `mOnSurface` | `surface` / `onSurface` |
| `mSurfaceVariant` / `mOnSurfaceVariant` | `surfaceVariant` / `onSurfaceVariant` |
| `mOutline` | `outline` |
| `mShadow` | `shadow` |
| `mHover` / `mOnHover` | `accent4` / `onAccent4` |
| `terminal.normal` / `terminal.bright` | tinct ANSI palette (16 colours) |
| `terminal.foreground` / `background` | `foreground` / `background` |

Both `dark` and `light` variants are written in one file when tinct produces a dual theme; a single-theme run still writes a valid file (the required `dark` key is always present).

## Reload behaviour

### Automatic

Noctalia v5 watches its config and palette files and **hot-reloads** the active palette when `tinct.json` changes — no action required.

### Manual fallback

Noctalia v4 needs the scheme re-selected (or the shell restarted) after a *new* scheme file appears; existing-scheme edits reload on selection. There is no dedicated "reload colour scheme" IPC command — the known nudge to force a re-apply is toggling dark mode twice:

```bash
qs -c noctalia-shell ipc call darkMode toggle
qs -c noctalia-shell ipc call darkMode toggle
```

## Uninstall / revert

1. In Noctalia's config, set `[theme] source` back to `wallpaper` (or `builtin`).
2. Delete the palette file:
   ```bash
   rm ~/.config/noctalia/palettes/tinct.json
   ```
3. `tinct plugins delete noctalia` to drop the plugin.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--noctalia.output-dir` | `~/.config/noctalia/palettes` | Output directory for the palette file (use `~/.config/noctalia/colorschemes/tinct` for Noctalia v4) |

Flags can also be passed as JSON via `--plugin-args`, e.g. `tinct generate ... -o noctalia --plugin-args noctalia='{"output-dir":"/some/path"}'`.

## Customising the template

```bash
tinct plugins templates dump -o noctalia -l ~/.config/tinct/templates/noctalia
```

This writes `templates/palette.json.tmpl`, rendered once per variant. tinct uses your version in preference to the embedded default.

## Troubleshooting

### Colours unchanged after generating

Noctalia is still on its wallpaper/builtin generator. Set `[theme] source = "custom"` and `custom_palette = "tinct"` (v5), or select the scheme in Settings (v4).

### Plugin skipped ("Noctalia not installed")

The plugin skips when `~/.config/noctalia` does not exist. Create it, or pass an explicit `--noctalia.output-dir`.

## Related plugins

- [waybar](../../../../internal/plugin/output/waybar/README.md) — status bar with a similar role; pair either with niri/Hyprland.
- [niri](../../../../internal/plugin/output/niri/README.md) — compositor colours for the niri scrollable-tiling WM.
- [fuzzel](../../../../internal/plugin/output/fuzzel/README.md) — launcher commonly paired with these shells.
