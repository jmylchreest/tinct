---
title: openrouter
sidebar_position: 7

plugin:
  type: input
  name: openrouter
  source: builtin
  source_type: api
  description: Generate images with OpenRouter.ai models and extract colours
  service: OpenRouter
  service_url: https://openrouter.ai/
  requires: []
  optional: []
  requires_network: true
  requires_credentials: [OPENROUTER_API_KEY]
  produces_wallpaper: true
---

# openrouter

Generates an image through [OpenRouter](https://openrouter.ai/) — a multi-provider router that exposes dozens of image-capable models behind one API and one bill — then extracts a palette from the result via k-means. Unless you pin a specific model with `--ai.model`, tinct fetches the live model list at runtime, filters to image-capable models, and picks the cheapest (or, by default, a free one if available).

The generated image is cached locally and exported as a wallpaper, so the same `tinct generate …` invocation produces both the palette and the wallpaper for output plugins like `hyprpaper`, `awww`, and `wbg`.

## Installation

Built into tinct — nothing to install separately.

## Quick start

```bash
export OPENROUTER_API_KEY="…"
tinct generate -i openrouter \
  --ai.prompt "cyberpunk city street with neon signs at night" \
  -o hyprland,hyprpaper,kitty
```

## Configuration / credentials

An OpenRouter API key is required. Create one at [openrouter.ai/keys](https://openrouter.ai/keys) and export it:

```bash
export OPENROUTER_API_KEY="…"
```

Rather than pasting the key into every shell session, pull it from a secret manager:

```bash
# GNOME Keyring / libsecret
export OPENROUTER_API_KEY=$(secret-tool lookup service openrouter)

# Bitwarden CLI
export OPENROUTER_API_KEY=$(bw get password openrouter)

# pass (password-store)
export OPENROUTER_API_KEY=$(pass show openrouter)
```

Generated images are cached under `~/.cache/tinct/generated/openrouter/`. Each file is keyed by a hash of the prompt and model, so re-running the same prompt is free. File extensions are chosen based on the MIME type returned by the API (PNG / JPEG / WebP / GIF).

### Model selection

When `--ai.model` is `auto` (the default):

1. Fetch the current model list from the OpenRouter API.
2. Filter to models that include image output.
3. If `--openrouter.prefer-free` is `true` (default), pick a free model when one is available.
4. Otherwise, pick the cheapest (per-image / per-request / per-token, normalised against typical generation cost).

Use `--ai.list-models` to see what's currently available with live pricing.

## Flags

### Shared AI flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ai.prompt` | _(required)_ | Text description of the image to generate |
| `--ai.model` | `auto` | Specific model ID, or `auto` to let tinct pick the cheapest image-capable model |
| `--ai.list-models` | `false` | List available models with live pricing and exit |
| `--ai.no-extended-prompt` | `false` | Disable tinct's automatic wallpaper-orientation prompt enhancement |
| `--ai.no-negative-prompt` | `false` | Disable the default negative prompt |
| `--ai.negative-prompt` | _(none)_ | Custom negative prompt |

### Plugin-specific flags

| Flag | Default | Description |
|------|---------|-------------|
| `--openrouter.prefer-free` | `true` | When `--ai.model=auto`, prefer a free model over a paid one even when paid is cheaper-per-image (e.g. flat-fee plans) |

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

OpenRouter uses a credit balance. Different models bill differently (per-image, per-request, or per-token), and pricing changes regularly — the live source of truth is `--ai.list-models` and the [OpenRouter models page](https://openrouter.ai/models).

- Auto-selection with `--openrouter.prefer-free` (default) will pick a free image-capable model when one exists, which is the cheapest possible path.
- Free models often have stricter rate limits and lower quality than paid ones.
- Monitor your usage at the [OpenRouter activity dashboard](https://openrouter.ai/activity).

## Troubleshooting

### `OPENROUTER_API_KEY environment variable is required`

Export the variable in the same shell where you run `tinct`. Get a key at https://openrouter.ai/keys.

### `--ai.prompt is required`

Validation requires a prompt unless `--ai.list-models` is set.

### `failed to auto-select model`

OpenRouter's model list didn't return any image-capable model that matched the auto-selection filter. Pass `--ai.model` explicitly with a known image model ID (see `--ai.list-models`), or check whether your account has access to image models.

### `failed to generate image: …`

Common causes: insufficient credits, model temporarily unavailable, or the prompt was rejected by the underlying provider's safety filters. Try `--ai.list-models` to confirm the model is up, then a simpler prompt to rule out filtering. The wrapped error message usually names the upstream provider.

### Same prompt always returns the same image

Cache hits are intentional. Pass `--cache-overwrite` to force a fresh generation, or change the prompt/model so the cache key changes.

## Related plugins

- [`google-genai`](../googlegenai/) — same idea, but Google Gemini / Imagen directly (no router)
- [`image`](../image/) — use an existing image instead of generating one
