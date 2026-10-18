---
title: image
sidebar_position: 1
plugin:
  type: input
  name: image
  source: builtin
  source_type: image
  description: >-
    Extract colours from images (files, directories, URLs) with optional ambient
    lighting
  requires: []
  optional: []
  requires_network: false
  requires_credentials: []
  produces_wallpaper: true
---

# image

Extracts a colour palette from an image using k-means clustering. The image can be a local file, a directory (a random image is picked), or an HTTP(S) URL. The plugin also implements `WallpaperProvider`, so output plugins that handle wallpapers (`hyprpaper`, `awww`, `wbg`, etc.) can install the source image alongside the theme.

## Installation

Built into tinct — nothing to install separately.

## Quick start

```bash
# Local file
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o hyprland,kitty

# Random image from a directory (non-recursive)
tinct generate -i image -p ~/Pictures/wallpapers/ -o hyprland

# Remote URL (cached locally for wallpaper support)
tinct generate -i image -p https://example.com/wallpaper.jpg -o hyprland,hyprpaper
```

## Configuration

There are no credentials to manage and nothing to install. The plugin reads its cache defaults from `~/.config/tinct/tinct.toml` if present:

```toml
[cache]
images = true                     # enable caching of remote images
overwrite = false                 # re-download even when a cache hit exists
# dir = ""                        # default: ~/.cache/tinct/images
# filename = ""                   # default: hash of the URL
```

Environment variables override the config file, and CLI flags override the env vars:

| Variable | Default | Description |
|----------|---------|-------------|
| `TINCT_IMAGE_CACHE` | `true` | Cache remote images locally |
| `TINCT_IMAGE_CACHE_DIR` | `~/.cache/tinct/images` | Where cached images land |
| `TINCT_IMAGE_CACHE_FILENAME` | _(URL hash)_ | Override cached filename |
| `TINCT_IMAGE_CACHE_OVERWRITE` | `false` | Re-download on every run |

Supported formats: JPEG (`.jpg`, `.jpeg`), PNG (`.png`), GIF (`.gif`), WebP (`.webp`).

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--image.path` | `-p` | _(required)_ | Path to image file, directory, or HTTP(S) URL |
| `--image.colours` | `-c` | `16` | Number of colours to extract (1–256) |
| `--image.extractAmbience` | | `false` | Extract edge/corner colours for ambient LED lighting |
| `--image.regions` | | `8` | Edge/corner region count (4, 8, 12, 16) |
| `--image.sample-size` | | `10` | Percentage of each edge to sample (1–50) |
| `--image.sample-method` | | `average` | `average` or `dominant` |
| `--image.seed-mode` | | `content` | Seed mode: `content`, `filepath`, `manual`, `random` |
| `--image.seed-value` | | `0` | Seed value when `--image.seed-mode=manual` |
| `--image.cache` | | `true` | Cache remote images for wallpaper support |
| `--image.cache-dir` | | `~/.cache/tinct/images` | Cache directory |
| `--image.cache-filename` | | _(URL hash)_ | Cached filename |
| `--image.cache-overwrite` | | `false` | Allow overwriting an existing cached image |

### Seed modes

| Mode | Determinism | Use case |
|------|-------------|----------|
| `content` | Same pixel data → same palette | Default; reproducible per image |
| `filepath` | Same absolute path → same palette | Reproducible per slot |
| `manual` | Same `--image.seed-value` → same palette | Explicit reproducibility |
| `random` | Different each run | Re-roll until you like the result |

### Ambient region extraction

When `--image.extractAmbience` is set, edge and corner colours are sampled in addition to the k-means palette and weighted at ~10% of the total so they don't dominate categorisation. The region count maps to a physical layout:

- `4` — corners only (budget LED setups)
- `8` — corners + mid-edges (default, recommended)
- `12` — corners + double edges (ultrawide)
- `16` — corners + triple edges (enthusiast bias-light arrays)

Region colours are tagged with `positionTopLeft`, `positionTop`, `positionRight`, etc., which LED-control output plugins can pick up directly.

## Output

A raw palette of up to `--image.colours` colours (default 16), plus optional region colours when `--image.extractAmbience` is enabled. Dark/light is detected automatically from the average luminance of the extracted palette.

This plugin **provides a wallpaper**:

- For local files, the wallpaper path is the resolved absolute path (or the tilde-prefixed form if the user wrote `~/...`).
- For remote URLs, the wallpaper path is the cached local file under `~/.cache/tinct/images/<hash>.<ext>`. If caching is disabled (`--image.cache=false`), no wallpaper path is exported — colour extraction still works.

Templates can use either form via `.WallpaperPath` (canonical, for config files) or `.WallpaperRawPath` (literal user input, for display/logging).

## Troubleshooting

### `image path or URL is required`

Pass an image with `-p`/`--image.path`. Directories are valid input — they pick a random child image, non-recursively.

### `failed to load image`

The file isn't a supported format or is corrupt. Confirm the format is JPEG, PNG, GIF, or WebP and that the file opens in another viewer. For URLs, fetch the URL with `curl -I` to verify it returns a 200 with an image content type.

### Remote URL works but `hyprpaper`/`awww` don't set a wallpaper

Output plugins that set wallpapers need a local file path. Make sure `--image.cache` is enabled (it is by default) so the URL is downloaded and the cached path is exported. If you previously disabled caching via `TINCT_IMAGE_CACHE=false`, override it with `--image.cache=true`.

### Same image produces a different palette each run

You're using `--image.seed-mode=random`. Switch to `content` (default) or `manual` with a fixed `--image.seed-value` for reproducible extraction.

## Related plugins

- [`file`](./file.md) — load a saved tinct palette instead of re-extracting
- [`markdown`](./markdown.md) — restore a theme exported by the `markdown` output plugin (including the embedded wallpaper)
- [`google-genai`](./google-genai.md), [`openrouter`](./openrouter.md) — generate the source image with an AI model instead of supplying one
