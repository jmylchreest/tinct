---
sidebar_position: 1
---

# Colour extraction

Tinct uses K-means clustering to extract meaningful colours from images. This page explains how the extraction process works.

## K-means clustering

K-means is an unsupervised machine learning algorithm that groups similar colours together. Tinct applies it to images as follows:

1. **Sampling**: Read all pixels from the image
2. **Clustering**: Group similar colours into K clusters (default: 16)
3. **Centroids**: Calculate the centre colour of each cluster
4. **Weighting**: Assign weights based on cluster size (colour frequency)

The result is a palette of representative colours with relative importance weights.

## Configuring extraction

Control the number of colours extracted:

```bash
# Extract 8 colours
tinct extract -i image -p ~/wallpaper.jpg -c 8

# Extract 24 colours
tinct extract -i image -p ~/wallpaper.jpg -c 24
```

## Seed modes

Tinct supports different seed modes for deterministic or varied results:

| Mode | Description |
|------|-------------|
| `content` | Seed based on image content (default) |
| `filepath` | Seed based on file path |
| `manual` | Use a specific seed value |
| `random` | Random seed each time |

```bash
# Deterministic based on content
tinct generate -i image -p ~/wallpaper.jpg --image.seed-mode content

# Random results
tinct generate -i image -p ~/wallpaper.jpg --image.seed-mode random

# Pin an explicit seed
tinct generate -i image -p ~/wallpaper.jpg --image.seed-mode manual --image.seed-value 42
```

## Ambient edge extraction

For monitor bias lighting and LED strip synchronisation, tinct can extract colours from image edges:

```bash
tinct generate -i image -p ~/wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8
```

### Region configurations

| Regions | Positions |
|---------|-----------|
| 4 | top, right, bottom, left |
| 8 | corners and edges |
| 12 | corners, edges, and inner points |
| 16 | Full grid including centre positions |

### Sampling methods

| Method | Description |
|--------|-------------|
| `average` | Average colour of the edge region |
| `dominant` | Most common colour in the region |

```bash
tinct generate -i image -p ~/wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8 \
  --image.samplingMethod dominant
```

## Remote image support

Tinct can fetch and cache images from URLs:

```bash
tinct generate -i image -p "https://example.com/wallpaper.jpg" -o all
```

Images are cached locally for subsequent runs. Configure caching in `~/.config/tinct/tinct.toml`:

```toml
[cache]
images = true
overwrite = false
# dir = ""      # default: ~/.cache/tinct/images
# filename = "" # default: content-hash based filename
```

Environment variables take priority over the config file:

```bash
export TINCT_IMAGE_CACHE=true
export TINCT_IMAGE_CACHE_DIR=~/.cache/tinct/images
export TINCT_IMAGE_CACHE_OVERWRITE=false
```

## Colour weights

Extracted colours include weights representing their prominence in the source:

- Weights are normalised to sum to 1.0
- Larger clusters (more pixels) have higher weights
- Weights influence role assignment priority

## Next steps

- [Material Design 3](./material-design.md) - How colours are categorised
- [Colour roles](./color-roles.md) - All 49+ semantic roles
- [Theme detection](./theme-detection.md) - Light/dark auto-detection
