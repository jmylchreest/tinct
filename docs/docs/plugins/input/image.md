---
sidebar_position: 2
---

# image

Extract colour palettes from images using K-means clustering.

## Description

The `image` plugin analyses images to extract dominant colours. It supports:

- Local files (JPEG, PNG, GIF, WebP)
- Remote URLs (auto-cached)
- Ambient edge/corner extraction for LED sync

## Usage

```bash
tinct generate -i image -p <path-or-url> [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--image.path` | | Path or URL to image (also `-p`) |
| `--image.colors` | `16` | Number of colours to extract |
| `--image.seed` | `content` | Seed mode: `content`, `filepath`, `manual`, `random` |
| `--image.extractAmbience` | `false` | Extract edge/corner colours |
| `--image.ambienceRegions` | `8` | Regions for ambient: 4, 8, 12, 16 |
| `--image.samplingMethod` | `average` | Method: `average` or `dominant` |

## Examples

### Basic extraction

```bash
# From local file
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all

# From URL
tinct generate -i image -p "https://example.com/wallpaper.jpg" -o all
```

### Control extraction

```bash
# Extract more colours
tinct generate -i image -p ~/wallpaper.jpg --image.colors 24 -o all

# Deterministic results
tinct generate -i image -p ~/wallpaper.jpg --image.seed filepath -o all
```

### Ambient lighting

```bash
# 8-region edge extraction
tinct generate -i image -p ~/wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8 \
  -o all

# With dominant colour sampling
tinct generate -i image -p ~/wallpaper.jpg \
  --image.extractAmbience \
  --image.samplingMethod dominant \
  -o all
```

## How it works

1. **Load image**: Read pixels from file or cached URL
2. **Cluster**: Apply K-means to group similar colours
3. **Weight**: Calculate prominence based on cluster size
4. **Background detection**: Find most suitable background colour
5. **Foreground selection**: Choose contrasting text colour
6. **Accent assignment**: Assign vibrant colours to accent roles
7. **Semantic generation**: Create danger, warning, success, info

## Ambient extraction

When `--image.extractAmbience` is enabled:

### 8 regions (default)

```
TL  T  TR
L       R
BL  B  BR
```

### 12 regions

Adds inner corner positions.

### 16 regions

Full grid including center positions.

### Output roles

Ambient colours are assigned to positional roles:

- `positionTopLeft`, `positionTop`, `positionTopRight`
- `positionLeft`, `positionRight`
- `positionBottomLeft`, `positionBottom`, `positionBottomRight`

Use these in LED control plugins for monitor bias lighting.

## Caching

Remote images are cached locally:

```bash
# Configure caching
export TINCT_IMAGE_CACHE=true
export TINCT_IMAGE_CACHE_DIR=~/.cache/tinct/images
export TINCT_IMAGE_CACHE_OVERWRITE=false
```

## Seed modes

| Mode | Behaviour |
|------|-----------|
| `content` | Same image = same colours (default) |
| `filepath` | Same path = same colours |
| `manual` | Specify seed value |
| `random` | Different each run |

## Wallpaper provision

The image plugin provides the input image as a wallpaper path. Output plugins like `hyprpaper` use this to:

1. Copy the image to a standard location
2. Update configuration to use it
3. Trigger wallpaper refresh

## See also

- [Colour extraction](/docs/concepts/color-extraction)
- [google-genai](/docs/plugins/input/google-genai) - AI image generation
