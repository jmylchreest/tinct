# Image Input Plugin

**Type:** Input Plugin  
**Built-in:** Yes  
**Language:** Go

Extract color palettes from images using k-means clustering with deterministic seed options and optional ambient region extraction for LED lighting.

## Overview

The `image` plugin is Tinct's primary color extraction plugin. It analyzes images (wallpapers, photos, artwork) and extracts a palette of representative colors using k-means clustering. The plugin supports both local files and remote HTTP(S) URLs, and can optionally extract edge/corner colors for ambient LED lighting setups.

## Features

- ✅ **K-means clustering** - Intelligent color extraction with configurable seed
- ✅ **Deterministic generation** - 4 seed modes for reproducible results
- ✅ **Local and remote sources** - Supports file paths and HTTP(S) URLs
- ✅ **Ambient region extraction** - Edge/corner colors for LED bias lighting
- ✅ **Theme detection** - Auto-detects dark/light themes from image luminance
- ✅ **Wallpaper provider** - Provides wallpaper path to output plugins
- ✅ **Smart loading** - Handles JPEG, PNG, GIF, WebP formats

## Usage

### Basic Extraction

```bash
# Extract from local image
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o hyprland,kitty

# Extract from URL
tinct generate -i image -p https://example.com/wallpaper.jpg -o hyprland
```

### With Custom Color Count

```bash
# Extract 24 colors instead of default 16
tinct generate -i image -p wallpaper.jpg -c 24 -o hyprland
```

### Ambient Lighting / LED Extraction

```bash
# Extract edge/corner colors for LED bias lighting
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.regions 8 \
  -o hyprland,wled-ambient

# 12 regions for ultrawide displays
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.regions 12 \
  -o hyprland
```

### Seed Modes (Deterministic Extraction)

```bash
# Content-based seed (default) - Same image → same colors
tinct generate -i image -p wallpaper.jpg --image.seed-mode content -o hyprland

# Filepath-based seed - Same location → same colors
tinct generate -i image -p wallpaper.jpg --image.seed-mode filepath -o hyprland

# Manual seed - Reproducible with specific seed
tinct generate -i image -p wallpaper.jpg \
  --image.seed-mode manual \
  --image.seed-value 42 \
  -o hyprland

# Random seed - Different colors each run
tinct generate -i image -p wallpaper.jpg --image.seed-mode random -o hyprland
```

## CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--image.path` | `-p` | *(required)* | Path to image file or HTTP(S) URL |
| `--image.algorithm` | `-a` | `kmeans` | Extraction algorithm (only kmeans supported) |
| `--image.colours` | `-c` | `16` | Number of colors to extract (1-256) |
| `--image.extractAmbience` | | `false` | Extract edge/corner regions for ambient lighting |
| `--image.regions` | | `8` | Number of regions to extract (4, 8, 12, 16) |
| `--image.sample-size` | | `10` | Percentage of edge to sample (1-50) |
| `--image.sample-method` | | `average` | Sampling method: `average` or `dominant` |
| `--image.seed-mode` | | `content` | Seed mode: `content`, `filepath`, `manual`, `random` |
| `--image.seed-value` | | `0` | Seed value (only used with `seed-mode=manual`) |

## Seed Modes

The image plugin supports 4 seed modes for k-means clustering, allowing you to control whether palette extraction is deterministic or random:

### `content` (Default)

Generates seed from image pixel data hash.

**Use case:** Same image content → same colors  
**Deterministic:** Yes  
**Changes if:** Image content changes

```bash
tinct generate -i image -p wallpaper.jpg --image.seed-mode content -o hyprland
```

### `filepath`

Generates seed from absolute file path hash.

**Use case:** Same file location → same colors  
**Deterministic:** Yes  
**Changes if:** File path changes

```bash
tinct generate -i image -p /home/user/wallpaper.jpg --image.seed-mode filepath -o hyprland
```

### `manual`

Uses user-provided seed value.

**Use case:** Reproducible results with custom seed  
**Deterministic:** Yes  
**Changes if:** You change the seed value

```bash
tinct generate -i image -p wallpaper.jpg \
  --image.seed-mode manual \
  --image.seed-value 12345 \
  -o hyprland
```

### `random`

Uses non-deterministic random seed.

**Use case:** Different colors each run  
**Deterministic:** No  
**Changes:** Every time

```bash
tinct generate -i image -p wallpaper.jpg --image.seed-mode random -o hyprland
```

## Ambient Region Extraction

The image plugin can extract colors from specific edge/corner regions for ambient LED lighting setups.

### Region Configurations

- **4 regions** - Corners only (budget setups)
- **8 regions** - Corners + mid-edges (recommended, default)
- **12 regions** - Corners + double edges (ultrawide displays)
- **16 regions** - Corners + triple edges (enthusiast setups)

### Example: 8 Regions

```
     [top]
[left]     [right]
  [Monitor]
[left]     [right]
   [bottom]
```

### Usage

```bash
# Enable ambient extraction with 8 regions
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.regions 8 \
  --image.sample-size 10 \
  --image.sample-method average \
  -o hyprland

# Use with WLED or other LED controllers
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.regions 8 \
  -o wled-ambient
```

### Sample Methods

- **`average`** (default) - Calculate average color of all pixels in region
- **`dominant`** - Find most frequent color in region (more vibrant)

### Weighting

Ambient colors are weighted at 10% of total palette weight to avoid over-representing edge colors in the final categorized palette.

**See:** [Region Extraction Documentation](../shared/regions/README.md)

## Supported Formats

- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- WebP (.webp)

## Interfaces Implemented

### Required: `Plugin`

```go
Name() string
Description() string
Version() string
Generate(ctx, opts) (*colour.Palette, error)
RegisterFlags(cmd)
Validate() error
```

### Optional: `WallpaperProvider`

```go
WallpaperPath() string  // Returns path to source image
```

Provides wallpaper path to output plugins like `hyprpaper`, `hyprlock` for wallpaper setting.

### Optional: `ThemeHinter`

```go
ThemeHint() string  // Returns "dark", "light", or "auto"
```

Suggests theme type based on average luminance of extracted colors.

## How It Works

1. **Load Image** - From local file or HTTP(S) URL using SmartLoader
2. **Calculate Seed** - Based on configured seed mode
3. **K-means Clustering** - Extract N most representative colors
4. **Extract Regions** (optional) - Sample edge/corner colors if enabled
5. **Return Palette** - Raw colors (categorization happens separately)

### K-means Clustering

The plugin uses k-means clustering to find the most representative colors in the image:

1. Initialize K random cluster centers (using configured seed)
2. Assign each pixel to nearest cluster
3. Recalculate cluster centers
4. Repeat until convergence
5. Return cluster centers as palette colors

**Seed control:** Different seeds produce different cluster initializations, affecting final colors.

## Lock File Configuration

The image plugin supports configuration via the lock file (`.tinct-plugins.json`):

```json
{
  "external_plugins": {
    "image": {
      "name": "image",
      "type": "input",
      "config": {
        "seed": {
          "mode": "content",
          "value": null
        }
      }
    }
  }
}
```

**Note:** This is primarily for documentation - the image plugin is built-in and doesn't require lock file configuration for basic usage.

## Examples

### Extract and Preview

```bash
# Extract colors and show preview in terminal
tinct extract --preview ~/Pictures/wallpaper.jpg

# Extract with categorization
tinct extract --categorise --preview ~/Pictures/wallpaper.jpg
```

### Different Color Counts

```bash
# Minimal palette (8 colors)
tinct generate -i image -p wallpaper.jpg -c 8 -o hyprland

# Standard palette (16 colors, default)
tinct generate -i image -p wallpaper.jpg -o hyprland

# Large palette (32 colors)
tinct generate -i image -p wallpaper.jpg -c 32 -o hyprland
```

### With Multiple Output Plugins

```bash
# Generate for multiple applications
tinct generate -i image -p wallpaper.jpg \
  -o hyprland,hyprpaper,hyprlock,kitty,waybar,dunst

# Wallpaper is automatically set by hyprpaper plugin
```

## Performance

- **Local files:** Very fast (< 1 second for most images)
- **Remote URLs:** Depends on download speed + processing
- **K-means:** O(n * k * i) where n=pixels, k=colors, i=iterations
- **Region extraction:** Minimal overhead (~5-10ms)

## Troubleshooting

### "Image path or URL is required"

**Problem:** No image path provided.

**Solution:**
```bash
tinct generate -i image -p /path/to/wallpaper.jpg -o hyprland
```

### "Invalid image path or URL"

**Problem:** Image file doesn't exist or URL is malformed.

**Solution:**
- Check file path is correct and file exists
- Ensure URL starts with `http://` or `https://`
- Test URL in browser first

### "Failed to load image"

**Problem:** Image format not supported or file corrupted.

**Solution:**
- Verify image format (JPEG, PNG, GIF, WebP)
- Try opening image in another program
- Convert to PNG if using exotic format

### "Colours must be between 1 and 256"

**Problem:** Invalid color count.

**Solution:**
```bash
tinct generate -i image -p wallpaper.jpg -c 16 -o hyprland
```

### "Invalid regions value"

**Problem:** Ambient regions count not 4, 8, 12, or 16.

**Solution:**
```bash
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.regions 8 \
  -o hyprland
```

## Related

- **[Input Plugin Guide](../README.md)** - Input plugin architecture
- **[Region Extraction](../shared/regions/README.md)** - Ambient lighting details
- **[File Plugin](../file/README.md)** - Load from saved palettes
- **[Main README](../../../../README.md)** - Project overview

## Testing

```bash
# Run unit tests
go test ./internal/plugin/input/image/...

# Test with verbose output
tinct generate -i image -p testdata/sample.jpg -o hyprland --verbose --dry-run
```

## Contributing

When modifying the image plugin:

1. Maintain backward compatibility with existing CLI flags
2. Add tests for new features
3. Update this README with new functionality
4. Consider performance impact of changes
5. Test with various image formats and sizes

## License

MIT License - Same as Tinct
