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
  service_url: 'https://platform.minimax.io/'
  requires: []
  optional: []
  requires_network: true
  requires_credentials:
    - MINIMAX_API_KEY
  produces_wallpaper: true
---

# minimax

Generates an image through [MiniMax](https://platform.minimax.io/) (the Hailuo image API), then extracts a palette from the result via k-means. The default is `MiniMax-H3-Image`, using the shared OpenAI-compatible Images transport. The older `image-01` model remains available explicitly.

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

The default H3 model uses an area tier of `2K` with the requested aspect ratio (about 2720×1536 for 16:9). Use `--ai.params '{"size":"3840x2160"}'` for exact 4K UHD output, or `--minimax.image-size 4K` for the larger 4K area tier. H3 area tiers are pixel-area targets, not long-edge lengths.

`image-01` accepts dimensions up to **2048px** per side (divisible by 8). With only an aspect ratio set, MiniMax renders at a low default (~720p), which looks soft on a high-DPI monitor. The plugin therefore derives explicit `width`/`height` from `--aspect-ratio` and `--minimax.image-size` (default `2K`), so `16:9` produces **2048×1152**. Pass `--minimax.image-size 1K` for faster/smaller renders, or an explicit long-edge pixel count (512–2048).

### Endpoint override

The default H3 model calls `https://api.minimax.io/v1/images/generations`. The configured `--minimax.base-url` retains the legacy `/v1/image_generation` value for compatibility; H3 derives the corresponding Images endpoint. Explicit `image-01` requests use the native endpoint. See the H3 section below for proxy URL formats.

### Model selection

MiniMax image models use a maintained catalogue, so `--ai.list-models` prints a maintained catalogue (works offline; no API key needed) rather than a live query. `--ai.model` defaults to `MiniMax-H3-Image`; you can pass any valid MiniMax image model ID.

## Flags

### Shared AI flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ai.params` | _(none)_ | JSON request-body overrides; see Custom API parameters below |
| `--ai.prompt` | _(required)_ | Text description of the image to generate |
| `--ai.model` | `MiniMax-H3-Image` | MiniMax image model ID |
| `--ai.list-models` | `false` | List the maintained model catalogue and exit |
| `--ai.no-extended-prompt` | `false` | Disable tinct's automatic wallpaper-orientation prompt enhancement |
| `--ai.no-negative-prompt` | `false` | Disable the default negative prompt |
| `--ai.negative-prompt` | _(none)_ | Custom negative prompt (appended to the prompt as an "Avoid: …" instruction) |

### Plugin-specific flags

| Flag | Default | Description |
|------|---------|-------------|
| `--minimax.base-url` | `https://api.minimax.io/v1/image_generation` | MiniMax image generation endpoint URL (override for proxies) |
| `--minimax.image-size` | `2K` | H3 area tier: `1K`, `2K`, `4K`; image-01 long-edge size: `1K`, `2K`, or pixels (512–2048) |

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

- [`google-genai`](./google-genai.md) — Google Gemini / Imagen image generation
- [`openrouter`](./openrouter.md) — multi-provider router across dozens of image models
- [`image`](./image.md) — use an existing image instead of generating one

## Custom API parameters

`--ai.params` accepts a JSON object merged recursively into the final provider request body. Explicit values override defaults; arrays and `null` replace their defaults. Unknown fields are forwarded to the server, and integer precision is preserved. For example:

```bash
# H3 (default)
--ai.params '{"size":"3840x2160","output_format":"png"}'

# Legacy image-01
--ai.model image-01 --ai.params '{"width":1536,"height":1024,"prompt_optimizer":false}'
```

Automatic cache keys include the parameters, model, prompt controls, aspect ratio, provider settings and letterbox processing. Existing cache entries using the old prompt/model-only hash are not reused. Custom `--cache-filename` values override this naming; use `--cache-overwrite` when changing settings. Metadata records explicit parameters. Credentials and endpoint configuration remain separate.

## H3 Image through the shared OpenAI-compatible transport

`MiniMax-H3-Image` is the default and uses `/v1/images/generations`. Select `--ai.model image-01` for the legacy model. An explicit H3 example:

```bash
tinct generate -i minimax \
  --ai.model MiniMax-H3-Image \
  --ai.prompt "misty mountains at dawn" \
  --ai.params '{"size":"3840x2160","output_format":"png"}' \
  -o hyprpaper,kitty
```

H3 uses the shared Images client; the native `image-01` response adapter is retained. Without explicit `size`, H3 receives `--aspect-ratio` and `--minimax.image-size` as an area level (`1K`, `2K`, or `4K`, default `2K`). When `--ai.params` supplies concrete dimensions, the generated aspect/area defaults are omitted because H3 does not allow them alongside explicit size. Other parameters are forwarded unchanged. Use `--ai.model` to select the model family and endpoint.

For H3, `--minimax.base-url` accepts an API root ending in `/v1`, an existing `/v1/image_generation` endpoint, or `/v1/images/generations`; tinct derives the corresponding Images endpoint. Requests use a ten-minute timeout and are not automatically retried.

The [official MiniMax CLI](https://github.com/MiniMax-AI/cli/blob/main/src/commands/image/generate.ts) documents the separate H3 and image-01 routes. [MiniMax H3 local serving](https://github.com/MiniMax-AI/MiniMax-H3-Image) also documents an OpenAI-style Images endpoint; local deployments can use the generic `openai` input directly, with their own model ID and server parameters.
