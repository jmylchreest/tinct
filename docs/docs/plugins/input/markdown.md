---
title: markdown
sidebar_position: 3
plugin:
  type: input
  name: markdown
  source: builtin
  source_type: file
  description: Load theme from tinct markdown theme file
  requires: []
  optional: []
  requires_network: false
  requires_credentials: []
  produces_wallpaper: true
---

# markdown

Loads a theme that was previously exported by the `markdown` output plugin. These markdown files bundle the full categorised palette, all metadata, and (optionally) a base64-embedded wallpaper. Use this plugin to back up themes, share them across machines, or roundtrip a generation pipeline without re-running extraction.

## Installation

Built into tinct — nothing to install separately.

## Quick start

```bash
# Restore a previously saved theme to multiple outputs
tinct generate -i markdown \
  --markdown.path ~/.config/tinct/themes/mountain-sunset.md \
  -o hyprland,kitty,waybar
```

The theme file is produced by the companion output plugin:

```bash
# Save a theme (output plugin)
tinct generate -i image -p ~/wallpaper.jpg \
  -o markdown \
  --markdown.name "Mountain Sunset" \
  --markdown.include-thumbnail
```

## Configuration

No credentials, no network. The theme file must be a tinct markdown export (see the `markdown` output plugin for the format). Wallpapers may be embedded inline as a base64 data URI or referenced as an external path; both are handled.

When the wallpaper is embedded, the plugin extracts it to a real file so output plugins that set wallpapers (`hyprpaper`, `awww`, `wbg`, etc.) can use it. By default the extracted file is written next to the theme file; override with `--markdown.wallpaper-dir`.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--markdown.path` | _(required)_ | Path to the tinct markdown theme file |
| `--markdown.extract-wallpaper` | `true` | Extract an embedded wallpaper to a real file |
| `--markdown.wallpaper-dir` | _(same dir as theme)_ | Override the directory the wallpaper is extracted to |

## Output

A categorised palette with the full role set preserved from the source theme file. Theme type (dark/light) is taken from the file's metadata.

This plugin **provides a wallpaper** when the theme file contains one:

- Embedded wallpapers are written to the configured directory and the extracted path is exported as `.WallpaperPath`.
- External references are resolved to an absolute path; `.WallpaperRawPath` preserves the original (possibly relative) string from the file for display.
- If the file has no wallpaper, no path is exported and wallpaper-aware output plugins will skip that step.

## Troubleshooting

### `--markdown.path is required`

Pass the theme file with `--markdown.path`. There's no auto-discovery.

### `failed to parse theme file`

The file isn't a recognised tinct markdown export, or it's been hand-edited beyond what the parser tolerates. Try regenerating the file via the `markdown` output plugin and diffing for layout changes.

### Wallpaper-aware output plugins don't set a wallpaper

Either the theme file doesn't include a wallpaper, or you disabled extraction with `--markdown.extract-wallpaper=false`. Re-export the theme with `--markdown.include-wallpaper` on the output side, or supply a wallpaper through a different input.

### Permission denied writing the extracted wallpaper

The default extraction directory is the same directory as the theme file, which may be read-only (e.g. when the file lives in `/etc` or a synced repo). Point `--markdown.wallpaper-dir` at a writable location such as `~/.cache/tinct/`.

## Related plugins

- [`file`](./file.md) — load a categorised JSON palette without the markdown wrapping
- [`image`](./image.md) — extract a fresh palette from an image instead of replaying one
