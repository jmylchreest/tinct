---
title: minimax
sidebar_position: 8

plugin:
  type: input
  name: minimax
  source: builtin
  source_type: api
  description: Generate images with MiniMax (Hailuo) and extract colours
  service: MiniMax
  service_url: https://platform.minimax.io/
  requires: []
  optional: []
  requires_network: true
  requires_credentials: [MINIMAX_API_KEY]
  produces_wallpaper: true
---

# minimax

Generates an image through [MiniMax](https://platform.minimax.io/) (the Hailuo image API), then extracts a palette from the result via k-means. MiniMax's text-to-image endpoint is fast and inexpensive, with a single image model (`image-01`).

The generated image is cached locally and exported as a wallpaper, so the same `tinct generate …` invocation produces both the palette and the wallpaper for output plugins like `hyprpaper`, `awww`, and `wbg`.

## Installation

Built into tinct — nothing to install separately.

## Quick start

```bash
export MINIMAX_API_KEY="…"
tinct generate -i minimax \
  --ai.prompt "cyberpunk city street with neon signs at night" \
  -o hyprland,hyprpaper,kitty
```

## Configuration / credentials

A MiniMax API key is required. Create one at [platform.minimax.io](https://platform.minimax.io/) and export it:

```bash
export MINIMAX_API_KEY="…"
```

Rather than pasting the key into every shell session, pull it from a secret manager:

```bash
# GNOME Keyring / libsecret
export MINIMAX_API_KEY=$(secret-tool lookup service minimax)

# Bitwarden CLI
export MINIMAX_API_KEY=$(bw get password minimax)

# pass (password-store)
export MINIMAX_API_KEY=$(pass show minimax)
```

Generated images are cached under `~/.cache/tinct/generated/minimax/`. Each file is keyed by a hash of the prompt and model, so re-running the same prompt is free. The file extension is chosen from the image format MiniMax returns (typically JPEG).

### Resolution

`image-01` accepts dimensions up to **2048px** per side (divisible by 8). With only an aspect ratio set, MiniMax renders at a low default (~720p), which looks soft on a high-DPI monitor. The plugin therefore derives explicit `width`/`height` from `--aspect-ratio` and `--minimax.image-size` (default `2K`), so `16:9` produces **2048×1152**. Pass `--minimax.image-size 1K` for faster/smaller renders, or an explicit long-edge pixel count (512–2048).

### Endpoint override

By default the plugin calls `https://api.minimax.io/v1/image_generation`. To route through a proxy or a compatible host, set `--minimax.base-url` to the full endpoint URL.

### Model selection

MiniMax exposes a single image model and no models-list API, so `--ai.list-models` prints a maintained catalogue (works offline; no API key needed) rather than a live query. `--ai.model` defaults to `image-01`; you can pass any valid MiniMax image model ID.

## Flags

### Shared AI flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ai.prompt` | _(required)_ | Text description of the image to generate |
| `--ai.model` | `image-01` | MiniMax image model ID |
| `--ai.list-models` | `false` | List the maintained model catalogue and exit |
| `--ai.no-extended-prompt` | `false` | Disable tinct's automatic wallpaper-orientation prompt enhancement |
| `--ai.no-negative-prompt` | `false` | Disable the default negative prompt |
| `--ai.negative-prompt` | _(none)_ | Custom negative prompt (appended to the prompt as an "Avoid: …" instruction) |

### Plugin-specific flags

| Flag | Default | Description |
|------|---------|-------------|
| `--minimax.base-url` | `https://api.minimax.io/v1/image_generation` | MiniMax image generation endpoint URL (override for proxies) |
| `--minimax.image-size` | `2K` | Target long-edge resolution: `1K`, `2K`, or pixels (512–2048) |

### Shared image flags

| Flag | Default | Description |
|------|---------|-------------|
| `--count` | `32` | Number of colours to extract |
| `--aspect-ratio` | `16:9` | `1:1`, `3:4`, `4:3`, `9:16`, `16:9`, `21:9` |
| `--cache` | `true` | Cache generated images |
| `--cache-dir` | `~/.cache/tinct/generated` | Override cache root |
| `--cache-filename` | _(prompt hash)_ | Override cached filename |
| `--cache-overwrite` | `false` | Re-generate even if a cache hit exists |
| `--extract-ambience` | `false` | Also sample edge/corner colours for ambient lighting |
| `--regions` | `8` | Edge region count (4, 8, 12, 16) |
| `--sample-percent` | `10` | Percent of each edge to sample |
| `--sample-method` | `average` | `average` or `dominant` |
| `--seed-mode` | `content` | Seed mode for k-means extraction |
| `--seed-value` | `0` | Seed value when `--seed-mode=manual` |
| `--trim-letterbox` | `true` | Trim solid letterbox borders baked into the generated image, then restore the requested aspect (`--trim-letterbox=false` to disable) |

## Output

A raw palette of `--count` colours (default 32) extracted from the generated image via k-means. Theme type is inferred from average luminance.

This plugin **provides a wallpaper**: the cached generated image is exported as `.WallpaperPath`, so wallpaper-aware output plugins install it automatically. With `--cache=false` a temporary file is used and `.WallpaperPath` still points at it for the duration of the run.

## Costs & rate limits

MiniMax bills per generated image and is one of the cheaper providers. Pricing changes over time — the live source of truth is the [MiniMax platform pricing](https://platform.minimax.io/). The figures shown by `--ai.list-models` are indicative only.

## Troubleshooting

### `MINIMAX_API_KEY environment variable is required`

Export the variable in the same shell where you run `tinct`. Get a key at https://platform.minimax.io/.

### `--ai.prompt is required`

Validation requires a prompt unless `--ai.list-models` is set.

### `API error <code>: <message>`

MiniMax reports failures in the response body even on an HTTP 200 (e.g. `1008 insufficient balance`, rate limiting, or a prompt rejected by safety filters). The code and message come straight from MiniMax — check your balance and simplify the prompt to rule out filtering.

### Same prompt always returns the same image

Cache hits are intentional. Pass `--cache-overwrite` to force a fresh generation, or change the prompt/model so the cache key changes.

## Related plugins

- [`google-genai`](../googlegenai/) — Google Gemini / Imagen image generation
- [`openrouter`](../openrouter/) — multi-provider router across dozens of image models
- [`image`](../image/) — use an existing image instead of generating one
