# AI Wallpaper Generation Guide

This guide covers using Tinct's AI-powered wallpaper generation feature to create custom wallpapers from text prompts and automatically generate colour palettes from them.

## 📋 Table of Contents

- [Overview](#overview)
- [Supported AI Providers](#supported-ai-providers)
- [Setup](#setup)
- [Basic Usage](#basic-usage)
- [Advanced Usage](#advanced-usage)
- [Prompt Engineering](#prompt-engineering)
- [Configuration](#configuration)
- [API Costs](#api-costs)
- [Troubleshooting](#troubleshooting)

## Overview

The AI wallpaper generation feature allows you to:

1. **Generate wallpapers from text prompts** using AI image generation services
2. **Automatically extract colour palettes** from the generated images
3. **Apply palettes system-wide** using output plugins
4. **Cache generated wallpapers** to avoid regeneration costs

**Workflow:**
```
Text Prompt → AI Generation → Wallpaper Image → Color Extraction → Palette → System Theme
```

## Supported AI Providers

### Stable Diffusion (Recommended)

**Best for:** High-quality, cost-effective generation

- **API:** [Stability AI](https://platform.stability.ai/)
- **Models:** SDXL, SD 1.5, SD 2.1
- **Cost:** ~$0.002 per image (1024x1024)
- **Speed:** 2-5 seconds
- **Quality:** Excellent for landscapes, abstract art, and general scenes

### DALL-E 3

**Best for:** Photorealistic images, complex compositions

- **API:** [OpenAI Platform](https://platform.openai.com/)
- **Models:** DALL-E 3, DALL-E 2
- **Cost:** $0.04-$0.08 per image
- **Speed:** 5-15 seconds
- **Quality:** Exceptional photorealism and prompt adherence

### Midjourney (via API)

**Best for:** Artistic, stylized images

- **API:** [Midjourney API](https://docs.midjourney.com/) (requires subscription)
- **Cost:** Subscription-based
- **Speed:** 10-30 seconds
- **Quality:** Highly artistic and creative results

## Setup

### 1. Get API Key

**Stable Diffusion:**
```bash
# Sign up at https://platform.stability.ai/
# Get API key from account settings
export TINCT_AI_API_KEY="sk-..."
export TINCT_AI_PROVIDER="stablediffusion"
```

**DALL-E:**
```bash
# Sign up at https://platform.openai.com/
# Get API key from API keys page
export TINCT_AI_API_KEY="sk-..."
export TINCT_AI_PROVIDER="dalle"
```

### 2. Configure Tinct

**Option 1: Environment Variables**
```bash
export TINCT_AI_API_KEY="your-api-key-here"
export TINCT_AI_PROVIDER="stablediffusion"  # or "dalle", "midjourney"
export TINCT_AI_MODEL="sdxl"                # optional: specific model
export TINCT_AI_QUALITY="high"              # low, medium, high
export TINCT_AI_CACHE_DIR="~/.cache/tinct/wallpapers"
```

**Option 2: Configuration File** (`~/.config/tinct/config.yaml`)
```yaml
sources:
  ai:
    enabled: true
    provider: stablediffusion  # dalle, stablediffusion, midjourney
    api_key_env: TINCT_AI_API_KEY
    
    # Model settings
    model: sdxl              # Provider-specific model name
    quality: high            # low, medium, high
    
    # Image settings
    default_size: 4k         # 4k (default), ultrawide, dual, triple, superwide, dual-4k
                             # or custom: "3840x2160", "5120x1440", etc.
    format: png              # png, jpg, webp
    
    # Generation settings
    steps: 30                # Inference steps (more = better quality, slower)
    cfg_scale: 7.0          # Prompt adherence (1-20, 7-9 recommended)
    seed: -1                # -1 for random, or specific seed for reproducibility
    
    # Caching
    cache_enabled: true
    cache_dir: ~/.cache/tinct/wallpapers
    cache_ttl: 7d           # Time to keep cached wallpapers
```

### 3. Verify Setup

```bash
# Test AI connection
tinct sources test ai-wallpaper

# Check available providers
tinct sources list

# Show AI plugin info
tinct sources info ai-wallpaper
```

## Basic Usage

### Simple Generation

```bash
# Generate wallpaper and apply palette (defaults to 4K: 3840x2160)
tinct generate --ai "cyberpunk city at night with neon lights"

# Shorter version
tinct gen --ai "serene mountain landscape at sunset"

# With specific color count
tinct generate --ai "ocean waves under starry sky" -c 8
```

### Specify Wallpaper Size

```bash
# Use size presets
tinct generate --ai "mountain vista" --size 4k          # 3840x2160 (default)
tinct generate --ai "cosmic scene" --size ultrawide     # 3440x1440 (21:9)
tinct generate --ai "city skyline" --size dual          # 3840x1080 (2x 1920x1080)
tinct generate --ai "panorama" --size triple            # 5760x1080 (3x 1920x1080)
tinct generate --ai "wide view" --size superwide        # 5120x1440 (32:9)
tinct generate --ai "dual 4k" --size dual-4k            # 7680x2160 (2x 4K)

# Custom dimensions
tinct generate --ai "custom scene" --size 5120x2880
tinct generate --ai "portrait" --size 1080x1920
```

### Specify Output Plugins

```bash
# Apply to specific applications with ultrawide size
tinct generate --ai "minimalist desert scene" \
               --size ultrawide \
               --outputs alacritty,kitty,rofi

# Apply to all configured outputs for triple monitor setup
tinct generate --ai "dark forest with fog" \
               --size triple \
               --outputs all
```

### Preview Before Applying

```bash
# Generate and preview palette without applying
tinct generate --ai "autumn forest colors" --preview

# Save wallpaper without applying
tinct generate --ai "cyberpunk alley" \
               --save-wallpaper ~/Pictures/wallpaper.png \
               --no-apply
```

## Advanced Usage

### Provider-Specific Settings

**Stable Diffusion:**
```bash
tinct generate --ai "futuristic cityscape" \
               --ai-provider stablediffusion \
               --ai-model sdxl \
               --ai-steps 50 \
               --ai-cfg-scale 8.0 \
               --size ultrawide
```

**DALL-E:**
```bash
tinct generate --ai "photorealistic mountain valley" \
               --ai-provider dalle \
               --ai-model dall-e-3 \
               --ai-quality hd \
               --ai-style vivid \
               --size 4k
```

### Reproducible Generation

```bash
# Use specific seed for reproducibility
tinct generate --ai "abstract geometric patterns" \
               --ai-seed 42

# Regenerate the same wallpaper later
tinct generate --ai "abstract geometric patterns" \
               --ai-seed 42
```

### Negative Prompts

```bash
# Specify what NOT to include
tinct generate --ai "peaceful landscape" \
               --ai-negative "people, buildings, cars, text"
```

### Batch Generation

```bash
# Generate multiple variations
tinct generate --ai "cosmic nebula" \
               --ai-batch 5 \
               --select-best

# Use different prompts from file
tinct generate --ai-prompts prompts.txt \
               --batch-mode sequential
```

### Style Modifiers

```bash
# Add style to prompt
tinct generate --ai "mountain scene" \
               --ai-style "oil painting, vibrant colors, impressionist"

# Use preset styles
tinct generate --ai "city street" \
               --ai-preset cyberpunk  # or: minimalist, vintage, abstract
```

## Prompt Engineering

### Best Practices

**1. Be Specific and Descriptive**
```bash
# ❌ Bad
tinct generate --ai "landscape"

# ✅ Good
tinct generate --ai "mountain landscape at golden hour, dramatic clouds, lake reflection"
```

**2. Include Mood and Atmosphere**
```bash
# ✅ Excellent
tinct generate --ai "serene Japanese garden, cherry blossoms, soft morning light, peaceful atmosphere"
```

**3. Specify Art Style**
```bash
# For distinct colour palettes
tinct generate --ai "cyberpunk city, neon lights, dark background, high contrast, digital art"
tinct generate --ai "minimalist desert, pastel colours, flat design, simple shapes"
```

**4. Control Color Palette**
```bash
# Warm colors
tinct generate --ai "sunset landscape, warm orange and red tones, golden hour"

# Cool colors
tinct generate --ai "arctic landscape, cool blue and white tones, winter atmosphere"

# Monochromatic
tinct generate --ai "forest scene, shades of green, natural tones"
```

### Prompt Templates

**Dark Theme:**
```
"[subject], dark background, moody lighting, deep colors, cinematic, high contrast"
```

**Light Theme:**
```
"[subject], bright and airy, soft pastel colors, clean aesthetic, minimalist"
```

**Vibrant Theme:**
```
"[subject], vibrant colors, saturated, bold, energetic, colorful"
```

**Nature Theme:**
```
"[subject], natural tones, organic, earthy colors, serene, peaceful"
```

**Cyberpunk Theme:**
```
"[subject], neon lights, dark background, purple and cyan tones, futuristic, digital"
```

### Example Prompts

```bash
# Workspace themes
tinct gen --ai "minimalist workspace, warm wood tones, soft lighting, clean desk"
tinct gen --ai "developer setup, dark theme, blue accent lights, modern tech"

# Nature themes
tinct gen --ai "misty mountain valley, pine trees, morning fog, cool tones"
tinct gen --ai "desert dunes at sunset, warm orange and red, dramatic sky"
tinct gen --ai "tropical beach, turquoise water, palm trees, golden sand"

# Abstract themes
tinct gen --ai "geometric patterns, pastel colors, modern abstract art"
tinct gen --ai "fluid gradients, purple to blue, smooth transitions, minimal"

# Seasonal themes
tinct gen --ai "autumn forest, golden leaves, warm sunlight, cozy atmosphere"
tinct gen --ai "winter landscape, snow-covered trees, cool blue tones, crisp"
```

## Configuration

### Complete Configuration Example

```yaml
sources:
  ai:
    enabled: true
    provider: stablediffusion
    api_key_env: TINCT_AI_API_KEY
    
    # Generation settings
    model: sdxl
    quality: high
    steps: 30
    cfg_scale: 7.5
    
    # Image settings
    default_size: 4k              # 4k, ultrawide, dual, triple, superwide, dual-4k
                                   # or custom like "3840x2160", "5120x1440"
    format: png
    
    # Size presets (optional overrides)
    size_presets:
      4k: 3840x2160
      ultrawide: 3440x1440        # 21:9
      superwide: 5120x1440        # 32:9
      dual: 3840x1080             # 2x 1920x1080
      triple: 5760x1080           # 3x 1920x1080
      dual-4k: 7680x2160          # 2x 3840x2160
      1080p: 1920x1080
      1440p: 2560x1440
    
    # Negative prompts (what to avoid)
    default_negative_prompt: "blurry, low quality, watermark, text, ugly, distorted"
    
    # Style presets
    presets:
      cyberpunk:
        prompt_suffix: "neon lights, dark background, futuristic, digital art"
        cfg_scale: 8.0
      
      minimalist:
        prompt_suffix: "clean, simple, minimal, flat design, pastel colors"
        cfg_scale: 7.0
      
      nature:
        prompt_suffix: "natural lighting, organic, serene, photorealistic"
        cfg_scale: 7.5
    
    # Caching
    cache_enabled: true
    cache_dir: ~/.cache/tinct/wallpapers
    cache_ttl: 7d
    cache_max_size: 1GB
    
    # Performance
    timeout: 60s
    retry_attempts: 3
    retry_delay: 5s

# Post-generation settings
post_generation:
  # Apply wallpaper to desktop
  set_wallpaper: true
  wallpaper_mode: fill  # fill, fit, stretch, center, tile
  
  # Save generated wallpaper
  save_to: ~/Pictures/tinct-wallpapers/
  filename_template: "tinct-{date}-{hash}.png"
  
  # Backup previous wallpaper
  backup_previous: true
  backup_dir: ~/Pictures/tinct-wallpapers/backups/
```

### Provider-Specific Configuration

**Stable Diffusion:**
```yaml
sources:
  ai:
    provider: stablediffusion
    model: sdxl  # or: sd-1.5, sd-2.1
    sampler: k_dpmpp_2m  # Sampling method
    clip_guidance: 0
```

**DALL-E:**
```yaml
sources:
  ai:
    provider: dalle
    model: dall-e-3  # or: dall-e-2
    quality: hd      # standard, hd
    style: vivid     # natural, vivid
```

## API Costs

### Cost Comparison (Approximate)

| Provider | Resolution | Cost per Image | Monthly (30 imgs) | Notes |
|----------|-----------|----------------|-------------------|-------|
| Stable Diffusion | 1024x1024 | $0.002 | $0.06 | Standard |
| Stable Diffusion | 2048x2048 | $0.008 | $0.24 | HD |
| Stable Diffusion | 3840x2160 (4K) | $0.016 | $0.48 | 4K |
| Stable Diffusion | 5760x1080 (Triple) | $0.012 | $0.36 | Multi-monitor |
| DALL-E 2 | 1024x1024 | $0.02 | $0.60 | Standard |
| DALL-E 3 | 1024x1024 | $0.04 | $1.20 | Standard |
| DALL-E 3 HD | 1024x1024 | $0.08 | $2.40 | High quality |

**Note**: Larger resolutions (ultrawide, triple, dual-4k) typically cost proportionally more based on total pixel count.

### Cost Optimization Tips

1. **Enable Caching**
   ```yaml
   cache_enabled: true
   cache_ttl: 7d  # Reuse wallpapers for a week
   ```

2. **Use Stable Diffusion for Regular Use**
   ```bash
   export TINCT_AI_PROVIDER=stablediffusion  # Most cost-effective
   ```

3. **Lower Resolution for Testing**
   ```bash
   tinct gen --ai "test prompt" --size 1080p  # Cheaper than 4K
   ```

4. **Batch Generate When Experimenting**
   ```bash
   # Generate once, get multiple variations
   tinct gen --ai "theme" --ai-batch 5 --ai-seed 42
   ```

## Troubleshooting

### API Key Not Working

```bash
# Verify API key is set
echo $TINCT_AI_API_KEY

# Test API connection
tinct sources test ai-wallpaper --verbose

# Check provider is correct
tinct config get sources.ai.provider
```

### Generation Too Slow

```bash
# Reduce steps (faster but lower quality)
tinct gen --ai "prompt" --ai-steps 20

# Use smaller resolution
tinct gen --ai "prompt" --ai-size 1920x1080

# Check provider status
tinct sources status ai-wallpaper
```

### Poor Quality Results

```bash
# Increase steps
tinct gen --ai "prompt" --ai-steps 50

# Adjust CFG scale (higher = more prompt adherence)
tinct gen --ai "prompt" --ai-cfg-scale 9.0

# Use more descriptive prompt
tinct gen --ai "detailed [subject], high quality, sharp focus, professional photography"

# Try different provider
tinct gen --ai "prompt" --ai-provider dalle
```

### Rate Limiting

```bash
# Enable caching to reduce API calls
tinct config set sources.ai.cache_enabled true

# Add retry delay
tinct config set sources.ai.retry_delay 10s

# Use cached wallpapers
tinct generate --use-cached
```

### Network Issues

```bash
# Increase timeout
tinct gen --ai "prompt" --timeout 120s

# Enable retries
tinct config set sources.ai.retry_attempts 5

# Check proxy settings
export HTTP_PROXY=http://proxy:port
export HTTPS_PROXY=https://proxy:port
```

## Examples

### Complete Workflow

```bash
# 1. Generate 4K wallpaper with AI (default)
tinct generate --ai "cosmic nebula, purple and blue, space, stars, high detail" \
               --ai-provider stablediffusion \
               --ai-quality high \
               --size 4k

# 2. Extract 16 colors
# (automatic)

# 3. Apply to terminals and window manager
# (automatic based on config)

# 4. Wallpaper saved to cache
# ~/.cache/tinct/wallpapers/cosmic-nebula-abc123.png

# 5. Palette saved
# ~/.cache/tinct/palettes/cosmic-nebula-abc123.json
```

### Multi-Monitor Workflow

```bash
# Generate for triple monitor setup (5760x1080)
tinct generate --ai "panoramic mountain landscape, wide vista" \
               --size triple \
               --ai-provider stablediffusion

# Generate for ultrawide (3440x1440)
tinct generate --ai "cinematic space scene" \
               --size ultrawide

# Generate for dual 4K setup (7680x2160)
tinct generate --ai "expansive cosmic vista" \
               --size dual-4k \
               --ai-quality high
```

### Daily Theme Rotation

```bash
# Create cron job for daily theme changes
# Add to crontab: crontab -e

# Daily at 9 AM - Random theme
0 9 * * * tinct generate --ai "$(shuf -n1 ~/.config/tinct/prompts.txt)"

# Every 4 hours - Cycle through themes
0 */4 * * * tinct generate --ai-prompts ~/.config/tinct/prompts.txt --random
```

### Integration with Scripts

```bash
#!/bin/bash
# theme-switcher.sh

THEMES=(
    "minimalist mountain landscape, pastel colors"
    "cyberpunk city, neon lights, dark"
    "serene ocean beach, tropical"
    "autumn forest, warm tones"
)

# Pick random theme
THEME="${THEMES[$RANDOM % ${#THEMES[@]}]}"

# Detect monitor setup and set size
if xrandr | grep -q "3440x1440"; then
    SIZE="ultrawide"
elif xrandr | grep -q "5760x1080"; then
    SIZE="triple"
else
    SIZE="4k"
fi

# Generate and apply
tinct generate --ai "$THEME" \
               --size "$SIZE" \
               --outputs all \
               --notify "Theme changed: $THEME"
```

## See Also

- [Plugin Development Guide](plugins.md)
- [Color Extraction Algorithms](algorithms.md)
- [Configuration Reference](configuration.md)
- [DEVELOPMENT.md](../DEVELOPMENT.md)

---

**Happy theming! 🎨🤖**