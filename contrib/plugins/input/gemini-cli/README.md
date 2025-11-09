# Gemini CLI Input Plugin for Tinct

Generate images using Google's Gemini AI and extract color palettes from them.

## Overview

This plugin wraps the `gemini-cli` tool with the `nanobanana` extension to:
1. Generate images from text prompts using Gemini 2.5 Flash Image model
2. Cache generated images for reuse
3. Extract dominant colors using ImageMagick
4. Return colors to tinct for theme generation

## Prerequisites

### Required

- **gemini-cli** - Google Gemini CLI tool
  ```bash
  # Arch Linux
  sudo pacman -S gemini-cli
  ```

- **nanobanana** - Gemini CLI extension for image generation
  ```bash
  # Install via Gemini CLI extensions
  gemini extensions install https://github.com/gemini-cli-extensions/nanobanana
  ```

- **Gemini API Key** - Get a free API key at https://aistudio.google.com/apikey
  ```bash
  export GEMINI_API_KEY="your-api-key-here"
  # Or add to ~/.bashrc or ~/.zshrc for persistence
  ```
  
  Supported environment variables (in order of precedence):
  - `NANOBANANA_GEMINI_API_KEY`
  - `NANOBANANA_GOOGLE_API_KEY`
  - `GEMINI_API_KEY`
  - `GOOGLE_API_KEY`

### Build Requirements

- **Go 1.23+** - For building the plugin
  ```bash
  # Arch Linux
  sudo pacman -S go
  ```

## Installation

1. **Build the plugin**:
   ```bash
   cd contrib/plugins/input/gemini-cli
   go build -o gemini-cli
   ```

2. **Add the plugin**:
   ```bash
   tinct plugins add ./gemini-cli --type input
   ```

3. **Enable the plugin**:
   ```bash
   tinct plugins enable gemini-cli
   ```

4. **Verify installation**:
   ```bash
   tinct plugins list
   ```

## Usage

### Basic Example (32 Colors Default)

Generate a theme from an AI-generated image:

```bash
tinct generate -i gemini-cli -o kitty \
  --plugin-args 'gemini-cli={"prompt":"sunset over mountains with purple sky"}'
```

### With Multiple Outputs

```bash
tinct generate -i gemini-cli -o kitty,waybar,hyprland \
  --plugin-args 'gemini-cli={"prompt":"cyberpunk city at night with neon lights"}'
```

### With Ambient Lighting (Edge/Corner Extraction)

Extract additional colors from image edges for richer ambient effects:

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={"prompt":"sunset over mountains","extract_ambience":true,"regions":8}'
```

### Custom Color Count

For minimalist themes with fewer colors:

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={"prompt":"minimalist zen garden","count":16}'
```

### Advanced Configuration

Full control over caching and color extraction:

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={
    "prompt":"peaceful zen garden with cherry blossoms",
    "count":32,
    "extract_ambience":true,
    "regions":8,
    "sample_method":"average",
    "cache":true,
    "cache_dir":"~/.cache/tinct/gemini-cli",
    "cache_filename":"zen-garden.png",
    "cache_overwrite":false
  }'
```

## Plugin Arguments

| Argument | Type | Default | Description |
|----------|------|---------|-------------|
| `prompt` | string | *required* | Text description for image generation |
| `count` | number | 32 | Number of colors to extract from the image |
| `extract_ambience` | boolean | false | Extract edge/corner colors for ambient lighting |
| `regions` | number | 8 | Number of edge regions to extract (4, 8, 12, 16) |
| `sample_percent` | number | 10 | Percentage of edge to sample (1-50) |
| `sample_method` | string | `average` | Sampling method: `average` or `dominant` |
| `seed_mode` | string | `content` | K-means seed mode: `content`, `manual`, `random` |
| `seed_value` | number | 0 | Manual seed value (only with `seed_mode=manual`) |
| `cache` | boolean | true | Enable caching of generated images |
| `cache_dir` | string | `~/.cache/tinct/gemini-cli` | Directory to store cached images |
| `cache_filename` | string | *auto-generated* | Custom filename for cached image |
| `cache_overwrite` | boolean | false | Allow overwriting existing cached images |

## How It Works

1. **Image Generation**:
   - Calls `gemini "/generate <prompt> --count=1"` (positional prompt)
   - Nanobanana extension generates image using Gemini 2.5 Flash Image model
   - Image is saved to disk by nanobanana

2. **Caching**:
   - If caching is enabled, image is copied to cache directory
   - Filename is either custom or generated from prompt hash (SHA256)
   - Subsequent runs with same prompt reuse cached image (unless overwrite=true)

3. **Color Extraction**:
   - Loads image using tinct's SmartLoader
   - Uses tinct's k-means clustering algorithm for color extraction
   - Optionally extracts edge/corner colors for ambient lighting
   - Returns array of RGB colors to tinct

4. **Theme Generation**:
   - Tinct receives the color palette (32 colors by default)
   - Categorizes colors into semantic roles (background, foreground, accents, etc.)
   - Generates theme files for all enabled output plugins

## Examples

### Moody Dark Theme

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={"prompt":"dark moody forest at twilight with deep blues and purples"}'
```

### Vibrant Colorful Theme

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={"prompt":"vibrant tropical sunset with oranges, pinks, and purples","count":20}'
```

### Minimalist Monochrome

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={"prompt":"minimalist black and white abstract geometric shapes"}'
```

### Nature-Inspired

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={"prompt":"autumn forest with golden leaves and earthy tones"}'
```

## Caching Behavior

### Default (Cache Enabled)

```bash
# First run - generates new image
tinct generate -i gemini-cli -o kitty \
  --plugin-args 'gemini-cli={"prompt":"ocean waves"}'
# → Saves to ~/.cache/tinct/gemini-cli/gemini-<hash>.png

# Second run with same prompt - uses cached image
tinct generate -i gemini-cli -o waybar \
  --plugin-args 'gemini-cli={"prompt":"ocean waves"}'
# → Reuses cached image
```

### Force Regeneration

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={"prompt":"ocean waves","cache_overwrite":true}'
# → Generates new image even if cached version exists
```

### Disable Caching

```bash
tinct generate -i gemini-cli -o all \
  --plugin-args 'gemini-cli={"prompt":"ocean waves","cache":false}'
# → Uses temporary file, not cached
```

## Troubleshooting

### "gemini-cli not found"

Install gemini-cli:
```bash
sudo pacman -S gemini-cli
```

### "nanobanana extension not found"

Install the nanobanana extension:
```bash
gemini extensions install https://github.com/gemini-cli-extensions/nanobanana
```

Verify it's enabled:
```bash
gemini extensions list
# Should show: ✓ nanobanana (X.X.X)
#              Enabled (User): true
```

### "Failed to find generated image"

Check that nanobanana is working:
```bash
gemini "/generate test image --count=1"
# Should generate an image file
```

### "API key not configured"

Set your Gemini API key:
```bash
export GEMINI_API_KEY="your-key-here"
# Or add to ~/.bashrc or ~/.zshrc
```

### Build errors

Ensure you have Go 1.23+ and all dependencies:
```bash
go version
cd contrib/plugins/input/gemini-cli
go mod tidy
go build -o gemini-cli
```

## Tips

1. **Descriptive Prompts**: More detailed prompts generate better images
   - ✓ "cyberpunk city at night with neon blue and purple lights reflecting on wet streets"
   - ✗ "city"

2. **Color Count**: Adjust based on desired theme complexity
   - 32 colors: Default, rich palette (recommended)
   - 20-24 colors: Medium complexity themes
   - 16 colors: Standard terminal themes
   - 8-12 colors: Minimalist themes

3. **Ambient Lighting**: Enable `extract_ambience` for edge/corner colors
   ```bash
   tinct generate -i gemini-cli -o all \
     --plugin-args 'gemini-cli={"prompt":"your prompt","extract_ambience":true,"regions":8}'
   ```
   This extracts additional colors from image edges for ambient lighting effects

4. **Cache Management**: Use custom filenames for important themes
   ```bash
   tinct generate -i gemini-cli -o all \
     --plugin-args 'gemini-cli={"prompt":"your prompt","cache_filename":"my-favorite-theme.png"}'
   ```

5. **Reuse Images**: Generate once, try different extraction settings
   ```bash
   # First time - generates image with 32 colors
   tinct generate -i gemini-cli -o kitty \
     --plugin-args 'gemini-cli={"prompt":"sunset"}'
   
   # Reuse with ambient extraction
   tinct generate -i gemini-cli -o waybar \
     --plugin-args 'gemini-cli={"prompt":"sunset","extract_ambience":true}'
   
   # Reuse with different color count
   tinct generate -i gemini-cli -o hyprland \
     --plugin-args 'gemini-cli={"prompt":"sunset","count":24}'
   ```

## Integration with Other Plugins

Gemini-generated images can be used with tinct's image input plugin:

```bash
# Generate and cache an image
tinct generate -i gemini-cli -o kitty \
  --plugin-args='{
    "prompt":"abstract art",
    "cache_filename":"abstract.png"
  }'

# Use the cached image with more extraction options
tinct generate -i image -o all \
  --image.path ~/.cache/tinct/gemini-cli/abstract.png \
  --image.colours 24 \
  --image.extractAmbience true
```

## License

MIT License - See LICENSE file in repository root

## Credits

- Built for [Tinct](https://github.com/jmylchreest/tinct)
- Uses [gemini-cli](https://github.com/google-gemini/gemini-cli)
- Uses [nanobanana](https://github.com/gemini-cli-extensions/nanobanana) extension
- Powered by Google Gemini 2.5 Flash Image model
