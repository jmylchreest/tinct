---
title: google-genai
sidebar_position: 6
plugin:
  type: input
  name: google-genai
  source: builtin
  source_type: api
  description: Generate images with Google Imagen and extract colours
  service: Google Gemini
  service_url: 'https://ai.google.dev/'
  requires: []
  optional: []
  requires_network: true
  requires_credentials:
    - GOOGLE_API_KEY
  produces_wallpaper: true
---

# google-genai

Generates an image with Google's Gemini and Imagen models, then extracts a colour palette from the result via k-means. The generated image is cached locally and exported as a wallpaper, so the same `tinct generate …` invocation produces both the palette and the wallpaper for output plugins like `hyprpaper`, `awww`, and `wbg`.

## Installation

Built into tinct — nothing to install separately.

## Quick start

```bash
export GOOGLE_API_KEY="…"
tinct generate -i google-genai \
  --ai.prompt "sunset over rolling tuscan hills" \
  -o hyprland,hyprpaper,kitty
```

The first run generates and caches the image; subsequent runs with the same prompt + model reuse the cached file unless you pass `--cache-overwrite`.

## Configuration / credentials

A Google API key is required. The free tier of [Google AI Studio](https://aistudio.google.com/apikey) is usually sufficient for personal use — sign in with a Google account, create an API key, and export it:

```bash
export GOOGLE_API_KEY="…"
```

Rather than pasting the key into every shell session, pull it from a secret manager:

```bash
# GNOME Keyring / libsecret
export GOOGLE_API_KEY=$(secret-tool lookup service google-genai)

# Bitwarden CLI
export GOOGLE_API_KEY=$(bw get password google-genai)

# KDE Wallet
export GOOGLE_API_KEY=$(kwallet-query -r google-genai kdewallet)

# pass (password-store)
export GOOGLE_API_KEY=$(pass show google-genai)
```

Generated images are cached under `~/.cache/tinct/generated/google-genai/`. Each file is keyed by a hash of the prompt and model, so re-running the same prompt is free.

### Available models

| Model | Notes |
|-------|-------|
| `gemini-2.5-flash-image` | Default. Gemini with native image output, fast and cheap |
| `imagen-3.0-generate-002` | Imagen 3 |
| `imagen-4.0-fast-generate-001` | Imagen 4 Fast |
| `imagen-4.0-generate-001` | Imagen 4 |
| `imagen-4.0-ultra-generate-001` | Imagen 4 Ultra (highest quality, slowest) |

Run `tinct generate -i google-genai --ai.list-models` to see the full live list.

## Flags

### Shared AI flags

| Flag | Default | Description |
|------|---------|-------------|
| `--ai.prompt` | _(required)_ | Text description of the image to generate |
| `--ai.model` | `auto` | Model ID; `auto` resolves to `gemini-2.5-flash-image` |
| `--ai.list-models` | `false` | List available models and exit |
| `--ai.no-extended-prompt` | `false` | Disable tinct's automatic wallpaper-orientation prompt enhancement |
| `--ai.no-negative-prompt` | `false` | Disable the default negative prompt |
| `--ai.negative-prompt` | _(none)_ | Custom negative prompt |

### Plugin-specific flags

| Flag | Default | Description |
|------|---------|-------------|
| `--googlegenai.backend` | `gemini-api` | Backend: `gemini-api` or `vertex-ai` |
| `--googlegenai.image-size` | `2K` | `1K` or `2K` — applies only to Imagen Standard/Ultra |

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

## Output

A raw palette of `--count` colours (default 32) extracted from the generated image via k-means. Theme type is inferred from average luminance.

This plugin **provides a wallpaper**: the cached generated image is exported as `.WallpaperPath`, so wallpaper-aware output plugins install it automatically. With `--cache=false` a temporary file is used and `.WallpaperPath` still points at it for the duration of the run.

## Costs & rate limits

Pricing depends on the model and changes regularly; the canonical source is the [Google AI pricing page](https://ai.google.dev/gemini-api/docs/pricing).

- The Gemini API has a free tier that covers Gemini 2.5 Flash Image generation at modest volumes — sufficient for personal wallpaper experimentation.
- Imagen 4 Standard and Ultra are billed per image and are not part of the free tier.
- Rate limits are per-API-key. If you hit a quota error, wait, switch models, or add a billing account.

Generation latency is typically 5–15 seconds; Imagen Ultra is noticeably slower.

## Troubleshooting

### `GOOGLE_API_KEY environment variable is required`

Export the variable in the same shell where you run `tinct`. Get a key at https://aistudio.google.com/apikey.

### `--ai.prompt is required`

Validation requires a prompt unless `--ai.list-models` is set. Even one word works; richer prompts produce richer palettes.

### `failed to generate image: …`

Common causes: API key lacks Gemini API access, quota exceeded, model name typo, or the prompt triggered Responsible AI filtering. Try `--ai.list-models` to confirm the model ID, then a simpler prompt to rule out filtering.

### Same prompt always returns the same image

Cache hits are intentional. Pass `--cache-overwrite` to force a fresh generation, or change the prompt/model so the cache key changes.

### Wallpaper-aware output plugins didn't install the image

Confirm `--cache` is enabled (default) so the file persists past the run. If you ran with `--cache=false`, the temp file is gone by the time a downstream tool tries to use it; re-run without that flag.

## Related plugins

- [`openrouter`](./openrouter.md) — same idea, but multi-provider via OpenRouter (often includes free image models)
- [`image`](./image.md) — use an existing image instead of generating one
