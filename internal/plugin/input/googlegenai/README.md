# Google Gen AI Input Plugin

Internal plugin for generating images using Google's Imagen models via the official Google Gen AI Go SDK.

## Features

- Direct integration with `google.golang.org/genai` SDK
- Support for Imagen 3, Imagen 4, Imagen 4 Ultra, and Imagen 4 Fast
- Automatic image caching
- Region/border colour sampling for ambient lighting
- Wallpaper passthrough support
- Multiple aspect ratios
- Deterministic colour extraction with content-based seeding

## Prerequisites

Set your Google API key:

```bash
export GOOGLE_API_KEY="your-api-key-here"
```

Get an API key at: https://aistudio.google.com/apikey

## Usage

### Basic Usage

```bash
tinct generate -i google-genai --prompt "sunset over mountains" -o kitty
```

### With Specific Model

```bash
tinct generate -i google-genai \
  --prompt "cyberpunk city at night" \
  --model "imagen-4.0-ultra-generate-001" \
  --aspect-ratio "21:9" \
  -o kitty
```

### With Ambient Lighting

```bash
tinct generate -i google-genai \
  --prompt "forest landscape" \
  --extract-ambience \
  --regions 16 \
  -o wled-ambient
```

### List Available Models

```bash
# In Go code
import "github.com/jmylchreest/tinct/internal/plugin/input/googlegenai"

googlegenai.ListModels()
```

## Available Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--prompt` | string | *required* | Text description for image generation |
| `--model` | string | `imagen-4.0-fast-generate-001` | Imagen model to use |
| `--aspect-ratio` | string | `16:9` | Image aspect ratio (1:1, 3:4, 4:3, 9:16, 16:9, 21:9) |
| `--negative-prompt` | string | - | Description of what to discourage |
| `--genai-backend` | string | `gemini-api` | Backend (gemini-api or vertex-ai) |
| `--count` | int | `32` | Number of colours to extract |
| `--extract-ambience` | bool | `false` | Extract edge/corner colours |
| `--regions` | int | `8` | Number of edge regions (4, 8, 12, 16) |
| `--sample-percent` | int | `10` | Percentage of edge to sample (1-50) |
| `--sample-method` | string | `average` | Sampling method (average or dominant) |
| `--seed-mode` | string | `content` | Seed mode (content, manual, random) |
| `--seed-value` | int64 | `0` | Manual seed value |
| `--cache` | bool | `true` | Enable image caching |
| `--cache-dir` | string | `~/.cache/tinct/google-genai` | Cache directory |
| `--cache-filename` | string | *auto* | Custom cache filename |
| `--cache-overwrite` | bool | `false` | Overwrite existing cache |

## Available Models

| Model ID | Description | Cost per Image |
|----------|-------------|----------------|
| `imagen-3.0-generate-002` | Imagen 3 (stable) | Standard |
| `imagen-4.0-fast-generate-001` | Imagen 4 Fast (default) | $0.02 |
| `imagen-4.0-generate-001` | Imagen 4 (flagship) | $0.04 |
| `imagen-4.0-ultra-generate-001` | Imagen 4 Ultra | $0.06 |

## Implementation Details

### Interfaces Implemented

- `input.Plugin` - Core plugin interface
- `input.WallpaperProvider` - Provides wallpaper path for `--set-wallpaper` flag

### Dependencies

- `google.golang.org/genai` v1.34.0+ - Official Google Gen AI Go SDK
- Uses tinct's internal utilities:
  - `internal/colour` - K-means colour extraction
  - `internal/image` - Smart image loading
  - `internal/plugin/input/shared/regions` - Region sampling

### Architecture

1. **Generate()** - Main entry point
   - Validates configuration
   - Determines cache path
   - Generates image via Google Gen AI SDK
   - Extracts colours using k-means
   - Optionally extracts region colours for ambient lighting

2. **Image Generation** - Calls `GenerateImages` API
   - Enhances prompt for wallpaper suitability
   - Configures aspect ratio and format
   - Handles safety filtering
   - Writes image bytes to cache

3. **Colour Extraction** - Uses tinct's k-means extractor
   - Supports content-based seeding for deterministic results
   - Combines main palette with region colours (if enabled)
   - Applies proper weights (90% main, 10% regions)

## Advantages Over External Plugin

The internal implementation provides several benefits over the external go-plugin version:

1. **No RPC Overhead** - Direct function calls instead of go-plugin RPC
2. **Shared Utilities** - Reuses tinct's internal colour extraction and image loading
3. **Type Safety** - Compile-time checks instead of runtime protocol marshaling
4. **Simpler Code** - ~500 lines vs ~600+ lines with plugin boilerplate
5. **Better Integration** - Native cobra flag registration
6. **Easier Maintenance** - Single codebase, no plugin versioning concerns

## Testing

The plugin can be tested with dry-run mode:

```bash
# Test validation
tinct generate -i google-genai --dry-run -o kitty
# Error: input plugin validation failed: prompt is required

# Test with prompt
tinct generate -i google-genai --prompt "test" --dry-run -o kitty
# Will attempt to generate (requires GOOGLE_API_KEY)
```

## Cost Considerations

- **Imagen 4 Fast**: $0.02/image (recommended for frequent use)
- **Imagen 4**: $0.04/image (balanced quality/cost)
- **Imagen 4 Ultra**: $0.06/image (highest quality)

Enable caching to minimize costs by reusing generated images.
