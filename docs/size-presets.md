# Wallpaper Size Presets Reference

Quick reference for wallpaper size presets supported by Tinct's source plugins.

## 📏 Standard Presets

### 4K (Default)
- **Preset**: `4k`
- **Resolution**: 3840x2160
- **Aspect Ratio**: 16:9
- **Use Case**: Standard 4K monitors, modern displays
- **Example**:
  ```bash
  tinct generate --ai "mountain landscape" --size 4k
  ```

### 1440p (2K)
- **Preset**: `1440p`
- **Resolution**: 2560x1440
- **Aspect Ratio**: 16:9
- **Use Case**: 1440p monitors, high-end gaming displays
- **Example**:
  ```bash
  tinct generate --ai "cosmic scene" --size 1440p
  ```

### 1080p (Full HD)
- **Preset**: `1080p`
- **Resolution**: 1920x1080
- **Aspect Ratio**: 16:9
- **Use Case**: Standard Full HD monitors
- **Example**:
  ```bash
  tinct generate --ai "minimalist design" --size 1080p
  ```

## 🖥️ Ultrawide Monitors

### Ultrawide (21:9)
- **Preset**: `ultrawide`
- **Resolution**: 3440x1440
- **Aspect Ratio**: 21:9
- **Use Case**: Ultrawide monitors (34" typical)
- **Example**:
  ```bash
  tinct generate --ai "panoramic vista" --size ultrawide
  ```

### Superwide (32:9)
- **Preset**: `superwide`
- **Resolution**: 5120x1440
- **Aspect Ratio**: 32:9
- **Use Case**: Super ultrawide monitors (49" typical), Samsung Odyssey G9
- **Example**:
  ```bash
  tinct generate --ai "expansive landscape" --size superwide
  ```

## 🖥️🖥️ Multi-Monitor Setups

### Dual Monitor (2x 1080p)
- **Preset**: `dual`
- **Resolution**: 3840x1080
- **Total Width**: 2x 1920
- **Aspect Ratio**: 32:9 (effective)
- **Use Case**: Two 1920x1080 monitors side-by-side
- **Example**:
  ```bash
  tinct generate --ai "wide cityscape" --size dual
  ```

### Triple Monitor (3x 1080p)
- **Preset**: `triple`
- **Resolution**: 5760x1080
- **Total Width**: 3x 1920
- **Aspect Ratio**: 48:9 (effective)
- **Use Case**: Three 1920x1080 monitors in a row
- **Example**:
  ```bash
  tinct generate --ai "panoramic scene, seamless" --size triple
  ```

### Dual 4K
- **Preset**: `dual-4k`
- **Resolution**: 7680x2160
- **Total Width**: 2x 3840
- **Aspect Ratio**: 32:9 (effective)
- **Use Case**: Two 4K monitors side-by-side
- **Example**:
  ```bash
  tinct generate --ai "ultra-wide vista" --size dual-4k
  ```

## 🎨 Custom Dimensions

You can specify any custom dimensions using `WIDTHxHEIGHT` format:

```bash
# Custom 5K display
tinct generate --ai "custom scene" --size 5120x2880

# Portrait orientation
tinct generate --ai "vertical art" --size 1080x1920

# Four monitor setup (2x2 grid of 1080p)
tinct generate --ai "massive scene" --size 3840x2160

# Custom ultrawide
tinct generate --ai "wide panorama" --size 3840x1600
```

## 📊 Size Comparison Chart

| Preset | Resolution | Aspect Ratio | Total Pixels | Relative Cost* |
|--------|-----------|--------------|--------------|----------------|
| 1080p | 1920x1080 | 16:9 | 2.1M | $ |
| 1440p | 2560x1440 | 16:9 | 3.7M | $$ |
| ultrawide | 3440x1440 | 21:9 | 5.0M | $$$ |
| 4k (default) | 3840x2160 | 16:9 | 8.3M | $$$$ |
| dual | 3840x1080 | 32:9 | 4.1M | $$ |
| superwide | 5120x1440 | 32:9 | 7.4M | $$$$ |
| triple | 5760x1080 | 48:9 | 6.2M | $$$ |
| dual-4k | 7680x2160 | 32:9 | 16.6M | $$$$$ |

*Relative cost for AI generation based on pixel count

## 🎯 Preset Selection Guide

### Choose **1080p** if:
- Testing prompts and themes
- Minimizing AI generation costs
- Standard 1080p display

### Choose **1440p** if:
- 2K display
- Balance between quality and cost
- Gaming monitor

### Choose **4k** (default) if:
- Modern 4K display
- High quality desktop wallpaper
- Single monitor setup
- Best quality/cost ratio for modern displays

### Choose **ultrawide** if:
- 34" ultrawide monitor (3440x1440)
- 21:9 aspect ratio display
- Cinematic viewing experience

### Choose **superwide** if:
- 49" super ultrawide monitor
- Samsung Odyssey G9 or similar
- Maximum horizontal space

### Choose **dual** if:
- Two 1920x1080 monitors side-by-side
- Want seamless wallpaper across both
- Budget-friendly multi-monitor

### Choose **triple** if:
- Three 1920x1080 monitors in a row
- Racing sim or flight sim setup
- Maximum immersion

### Choose **dual-4k** if:
- Two 4K monitors side-by-side
- Professional workstation
- Maximum resolution multi-monitor

## 💡 Tips for Multi-Monitor Wallpapers

### Prompt Engineering
For seamless multi-monitor wallpapers, include these keywords:
- "panoramic"
- "wide vista"
- "seamless"
- "continuous scene"
- "no center focus"

**Examples:**
```bash
# Good for triple monitor
tinct generate --ai "panoramic mountain vista, seamless landscape, no center focus" --size triple

# Good for dual monitor
tinct generate --ai "wide cityscape, continuous skyline, horizon centered" --size dual

# Good for ultrawide
tinct generate --ai "cinematic space scene, 21:9 aspect ratio, wide composition" --size ultrawide
```

### Avoid Center-Focused Compositions
For multi-monitor setups, avoid:
- Single subject in the center (will be split across bezels)
- Symmetrical compositions with center axis
- Portraits or centered objects

Prefer:
- Horizontal landscapes
- Continuous patterns
- Scenes with distributed visual interest

### Testing Multi-Monitor Wallpapers

```bash
# Generate and preview before applying
tinct generate --ai "your prompt" --size triple --preview

# Save for manual inspection
tinct generate --ai "your prompt" --size dual --save-wallpaper test.png

# If satisfied, apply to system
tinct generate --ai "your prompt" --size triple --outputs all
```

## 🔧 Configuration

Set default size in `~/.config/tinct/config.yaml`:

```yaml
sources:
  ai:
    default_size: 4k          # or ultrawide, dual, triple, etc.
  
  random_fetch:
    default_size: ultrawide   # Apply to fetch sources too

# Override per source
size_overrides:
  ai-wallpaper: dual-4k
  random-fetch: 4k
```

## 🌐 Provider Support

### AI Generation (Stable Diffusion, DALL-E)
- ✅ All presets supported
- ⚠️ Large sizes (dual-4k, triple) may be slower and more expensive
- 💡 Some providers may have maximum dimension limits

### Random Fetch (Unsplash, Pexels)
- ✅ Most presets supported
- ⚠️ Some APIs may not support very large sizes
- 💡 Falls back to largest available size

### Local Selection
- ✅ All presets supported (scaling applied if needed)
- 💡 Best quality with native resolution images

## 📝 Examples by Use Case

### Productivity Setup (Dual 4K)
```bash
tinct generate --ai "minimalist workspace, clean desk, soft lighting" \
               --size dual-4k \
               --outputs alacritty,kitty,rofi
```

### Gaming Setup (Ultrawide)
```bash
tinct generate --ai "cyberpunk cityscape, neon lights, dark atmosphere" \
               --size ultrawide \
               --ai-quality high
```

### Racing Sim (Triple)
```bash
tinct generate --ai "race track panorama, wide angle, horizon view" \
               --size triple \
               --ai-steps 50
```

### Standard Desktop (4K)
```bash
tinct generate --ai "serene nature scene, mountains and lake" \
               --size 4k \
               --outputs all
```

### Budget Testing (1080p)
```bash
tinct generate --ai "abstract geometric pattern" \
               --size 1080p \
               --preview
```

---

**See Also:**
- [AI Wallpaper Generation Guide](ai-wallpaper.md)
- [Configuration Reference](configuration.md)
- [README](../README.md)