---
title: openai
sidebar_position: 8

plugin:
  type: input
  name: openai
  source: builtin
  source_type: api
  description: Generate wallpapers with OpenAI-compatible image APIs, including local servers
  service: OpenAI-compatible API
  service_url: https://developers.openai.com/api/docs/guides/image-generation
  requires: []
  optional: []
  requires_network: true
  requires_credentials: []
  produces_wallpaper: true
---

# openai

Generate a wallpaper through an OpenAI-compatible image endpoint, then extract its colour palette. Supports OpenAI directly and compatible local or hosted servers. No OpenAI SDK dependency is required.

## Quick start

```bash
export OPENAI_API_KEY="…"
tinct generate -i openai \
  --ai.prompt "misty mountains at dawn" \
  --ai.params '{"size":"3840x2160"}' \
  -o hyprland,hyprpaper,kitty
```

The default model for the official Images endpoint is `gpt-image-2.6-flare`. Model access, valid parameters and billing are determined by the server. API credentials are separate from a Codex subscription login.

## Local servers and compatible providers

```bash
tinct generate -i openai \
  --openai.base-url http://localhost:8080/v1 \
  --openai.api-key-env= \
  --ai.model your-image-model \
  --ai.prompt "a quiet forest" \
  --ai.params '{"size":"1024x1024"}' \
  --aspect-ratio 1:1 -o kitty
```

The base URL includes the API version prefix. The plugin appends `/images/generations` or `/chat/completions`. `OPENAI_BASE_URL` supplies the initial base URL; the flag overrides it. A custom endpoint requires an explicit `--ai.model`. An empty `--openai.api-key-env` disables authentication; otherwise the named environment variable supplies a bearer token. An unset key is allowed for custom servers.

A service must implement the selected endpoint and response schema. This does not automatically support ComfyUI workflows, arbitrary diffusion server APIs, or text-only chat models.

For OpenRouter-compatible image output through Chat Completions:

```bash
tinct generate -i openai \
  --openai.base-url https://openrouter.ai/api/v1 \
  --openai.protocol chat \
  --openai.api-key-env OPENROUTER_API_KEY \
  --ai.model YOUR_IMAGE_CAPABLE_MODEL \
  --ai.prompt "misty mountains" \
  --ai.params '{"image_config":{"aspect_ratio":"21:9"}}' \
  --aspect-ratio 21:9 -o kitty
```

The dedicated `-i openrouter` preset additionally provides live model discovery, pricing and automatic model selection. It uses the same shared image transport. See the [OpenRouter guide](../openrouter/) for complete examples and parameter guidance.

## Parameters

| Flag | Default | Purpose |
|------|---------|---------|
| `--openai.base-url` | `https://api.openai.com/v1` | API root, including version prefix; also reads `OPENAI_BASE_URL` |
| `--openai.protocol` | `images` | `images` or `chat` |
| `--openai.api-key-env` | `OPENAI_API_KEY` | Credential environment variable; empty disables authentication |
| `--ai.prompt` | required | Wallpaper description |
| `--ai.model` | `auto` | Official Images default, or explicit model ID for other endpoints |
| `--ai.params` | empty | JSON object merged into the final API request |
| `--ai.list-models` | `false` | Print default model and selection guidance; does not discover remote models |

Shared prompt, colour extraction, ambience, cache, seed and letterbox flags also apply. `--aspect-ratio` guides the prompt and post-processing; use `--ai.params` to set exact API dimensions. Set a matching aspect ratio, or use `--trim-letterbox=false` to preserve the server image dimensions.

Objects merge recursively. Explicit scalars, arrays and `null` replace generated defaults. Unknown fields are forwarded unchanged; the server decides whether they are valid. Integers retain their precision. For example:

```bash
--ai.params '{"quality":"high"}'
```

This is request-body configuration, not configuration of credentials, URLs or headers. Streaming is not supported. The plugin consumes the first returned image even if the request asks for several; requesting multiple images can incur additional charges.

## Output and caching

The Images protocol accepts `data[].b64_json` or `data[].url`. Chat mode accepts the OpenRouter extension `choices[].message.images[].image_url.url`, containing a data URL or HTTP(S) URL. Downloads do not receive the API authorization header. Text-only responses produce a clear error.

Generated images are stored under `~/.cache/tinct/generated/openai/` and exposed to wallpaper output plugins. Metadata includes the effective request, endpoint and protocol, without the API key. Cache hashes include generation parameters, endpoint, model, prompt options and post-processing settings. JSON whitespace and object key order do not affect the hash. `--cache-filename` deliberately overrides automatic naming; use `--cache-overwrite` when reusing a custom name with different settings. `--cache=false` writes to a temporary directory, retaining the image for wallpaper consumers.

## Provider compatibility

Checked against provider documentation on 2026-11-17:

| Service | Image-generation API | How tinct uses it |
|---------|----------------------|-------------------|
| OpenAI / compatible local servers | Images API | `openai`, protocol `images` |
| OpenRouter image-output models | Chat Completions with an image-output extension | `openrouter`, or `openai` with protocol `chat` |
| MiniMax H3 Image | OpenAI-style `/v1/images/generations` | `minimax` with H3 selected, or generic `openai` |
| MiniMax `image-01` | Native `/v1/image_generation` | Existing `minimax` adapter, with shared `--ai.params` |
| Google Gemini / Imagen | Native GenAI SDK | Existing `google-genai` adapter, with raw request-body overrides |
| Anthropic Claude | No native raster image output | Anthropic text driver in `pkg/ai`; no native wallpaper plugin |

MiniMax documents [OpenAI](https://platform.minimax.io/docs/api-reference/text-openai-api) and [Anthropic](https://platform.minimax.io/docs/api-reference/text-anthropic-api) compatibility for its chat models. Its newer [H3 model](https://github.com/MiniMax-AI/MiniMax-H3-Image) supports the Images protocol and the `minimax` plugin uses the shared client for that family. Its [image-01 endpoint](https://platform.minimax.io/docs/api-reference/image-generation-t2i) uses a separate wire format, so it cannot simply inherit the OpenAI Images adapter.

[Claude does not natively generate photos or illustrations](https://support.claude.com/en/articles/9002504-can-claude-produce-images). OpenRouter offers an [image-generation server tool](https://openrouter.ai/docs/guides/features/server-tools/image-generation), including access through Anthropic Messages. That is a provider-specific tool workflow, not an Anthropic image-generation standard. The server-tool response path and Anthropic Messages protocol are not implemented here; use a directly image-capable model through the supported image transports.

For provider implementers, `pkg/ai` supplies reusable image and text drivers independent of plugin flags. `ai.ImageRequest.Params` applies recursive JSON overrides, while `shared/aiflags` handles CLI parsing and cache fingerprints. The OpenRouter plugin and MiniMax H3 path reuse these public drivers. Native providers retain their response adapters and share parameter handling.
