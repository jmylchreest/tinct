# OpenRouter Input Plugin

Internal plugin for generating images using OpenRouter.ai models via their REST API.

## Features

- Direct integration with OpenRouter.ai API
- Support for multiple image generation models (Flux, DALL-E, Gemini, etc.)
- Automatic model selection (cheapest/free models preferred)
- Live pricing from API with `--ai.list-models`
- Automatic image caching
- Region/border colour sampling for ambient lighting
- Wallpaper passthrough support
- Multiple aspect ratios
- Deterministic colour extraction with content-based seeding

## Prerequisites

Set your OpenRouter API key:

```bash
export OPENROUTER_API_KEY="your-api-key-here"
```

Get an API key at: https://openrouter.ai/keys

## Usage

### Basic Usage

The default model is `auto`, which automatically selects the cheapest available image generation model (preferring free models).

```bash
tinct generate -i openrouter --ai.prompt "sunset over mountains" -o kitty
```

### With Specific Model

```bash
tinct generate -i openrouter \
  --ai.prompt "cyberpunk city at night" \
  --ai.model "google/gemini-2.5-flash-image-preview" \
  --aspect-ratio "16:9" \
  -o kitty
```

### With Ambient Lighting

```bash
tinct generate -i openrouter \
  --ai.prompt "forest landscape" \
  --extract-ambience \
  --regions 16 \
  -o wled-ambient
```

### List Available Models

```bash
tinct generate -i openrouter --ai.list-models
```

This will show all available image generation models with their current pricing, sorted by cost (free first).

## Available Flags

The plugin uses three categories of flags:

### Shared AI Flags (ai.* prefix)

These flags are shared across all AI image generation plugins.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--ai.prompt` | string | *required* | Text description for AI image generation |
| `--ai.model` | string | `auto` | AI model to use (auto selects cheapest) |
| `--ai.list-models` | bool | `false` | List available AI models and exit |
| `--ai.no-extended-prompt` | bool | `false` | Disable automatic wallpaper prompt enhancements |
| `--ai.no-negative-prompt` | bool | `false` | Disable default negative prompt |
| `--ai.negative-prompt` | string | - | Custom negative prompt to discourage certain elements |

### Common Flags (unprefixed)

These flags are shared across multiple input plugins.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--count` | int | `32` | Number of colours to extract |
| `--aspect-ratio` | string | `16:9` | Image aspect ratio (1:1, 3:4, 4:3, 9:16, 16:9) |
| `--cache` | bool | `true` | Enable image caching |
| `--cache-dir` | string | `~/.cache/tinct/generated` | Cache directory |
| `--cache-filename` | string | *auto* | Custom cache filename |
| `--cache-overwrite` | bool | `false` | Overwrite existing cache |
| `--extract-ambience` | bool | `false` | Extract edge/corner colours |
| `--regions` | int | `8` | Number of edge regions (4, 8, 12, 16) |
| `--sample-percent` | int | `10` | Percentage of edge to sample (1-50) |
| `--sample-method` | string | `average` | Sampling method (average or dominant) |
| `--seed-mode` | string | `content` | Seed mode (content, manual, random) |
| `--seed-value` | int64 | `0` | Manual seed value |

### Plugin-Specific Flags (openrouter.* prefix)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--openrouter.prefer-free` | bool | `true` | Prefer free models when using auto selection |

## Auto Model Selection

When `--ai.model` is set to `auto` (the default), the plugin:

1. Fetches the list of available models from OpenRouter API
2. Filters to only models that support image output
3. Sorts by cost (cheapest first)
4. If `--openrouter.prefer-free` is enabled (default), selects a free model if available
5. Otherwise selects the cheapest model

This ensures you always use the most cost-effective option without needing to track model availability.

## Available Models

Use `--ai.list-models` to see current models and pricing. Common image-capable models include:

| Model ID | Description |
|----------|-------------|
| `google/gemini-2.5-flash-image` | Gemini Flash image generation |
| `black-forest-labs/flux.2-pro` | Flux 2 Pro |
| `black-forest-labs/flux.2-flex` | Flux 2 Flex |
| `sourceful/riverflow-v2-standard-preview` | Riverflow V2 Standard |
| Various others | See `--ai.list-models` for current list |

**Note:** Not all OpenRouter models appear in the API response used by `--ai.list-models`. You can use any valid model ID with `--ai.model`, even if not listed. For the complete list of image generation models, visit:
https://openrouter.ai/models?fmt=cards&input_modalities=image%2Ctext&output_modalities=image

## Implementation Details

### Interfaces Implemented

- `input.Plugin` - Core plugin interface
- `input.WallpaperProvider` - Provides wallpaper path for `--set-wallpaper` flag

### API Integration

Uses the OpenRouter.ai REST API:
- **Endpoint**: `https://openrouter.ai/api/v1/chat/completions`
- **Authentication**: Bearer token via `OPENROUTER_API_KEY`
- **Image Generation**: Uses `modalities: ["image", "text"]` parameter
- **Response**: Base64-encoded PNG images in the response message

### Dependencies

- Uses tinct's internal utilities:
  - `internal/colour` - K-means colour extraction
  - `internal/image` - Smart image loading
  - `internal/plugin/input/shared/regions` - Region sampling
  - `internal/plugin/input/shared/aiflags` - Shared AI flags
  - `internal/plugin/input/shared/commonflags` - Shared common flags

### Architecture

1. **Generate()** - Main entry point
   - Auto-selects model if set to "auto"
   - Validates configuration
   - Determines cache path
   - Generates image via OpenRouter API
   - Extracts colours using k-means
   - Optionally extracts region colours for ambient lighting

2. **Image Generation** - Calls chat/completions API
   - Enhances prompt for wallpaper suitability
   - Configures aspect ratio (model dependent)
   - Parses base64 image data from response
   - Writes image bytes to cache

3. **Model Selection** - Automatic cheapest model selection
   - Fetches models from `/api/v1/models`
   - Filters for image output capability
   - Sorts by pricing
   - Prefers free models by default

4. **Colour Extraction** - Uses tinct's k-means extractor
   - Supports content-based seeding for deterministic results
   - Combines main palette with region colours (if enabled)
   - Applies proper weights (90% main, 10% regions)

## Testing

The plugin can be tested with dry-run mode:

```bash
# Test validation
tinct generate -i openrouter --dry-run -o kitty
# Error: input plugin validation failed: --ai.prompt is required

# Test with prompt (dry-run skips API call)
tinct generate -i openrouter --ai.prompt "test" --dry-run -o kitty

# List models (does not require prompt)
tinct generate -i openrouter --ai.list-models
```

## Cost Considerations

- Use `auto` model selection (default) to always use the cheapest available model
- Enable `--openrouter.prefer-free` (enabled by default) to prioritize free models
- Enable caching to minimize costs by reusing generated images
- Use `--ai.list-models` to see current pricing before generating

## Comparison with Google GenAI Plugin

| Feature | OpenRouter | Google GenAI |
|---------|-----------|--------------|
| Model Variety | Many providers (Flux, DALL-E, etc.) | Google models only |
| Pricing | Variable, includes free options | Per-model pricing |
| Auto Selection | Cheapest model by default | Manual selection |
| API Key | `OPENROUTER_API_KEY` | `GOOGLE_API_KEY` |
| Negative Prompts | Via prompt enhancement | Native API support |
