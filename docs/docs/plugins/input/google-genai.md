---
sidebar_position: 3
---

# google-genai

Generate images and extract colours using Google Gemini AI.

## Description

The `google-genai` plugin uses Google's Gemini model with Imagen to:

1. Generate an image from a text prompt
2. Extract colours from the generated image
3. Provide the image as wallpaper

## Requirements

- Google API key from [Google AI Studio](https://aistudio.google.com/apikey)
- Set `GOOGLE_API_KEY` environment variable

## Usage

```bash
export GOOGLE_API_KEY="your-api-key"
tinct generate -i google-genai --ai.prompt "<description>" [flags]
```

### Storing your API key

Rather than pasting your key into every shell session, retrieve it from a secret manager:

```bash
# GNOME Keyring / libsecret
export GOOGLE_API_KEY=$(secret-tool lookup service google-genai)

# Bitwarden CLI
export GOOGLE_API_KEY=$(bw get password google-genai)

# KDE Wallet (via kwallet-query)
export GOOGLE_API_KEY=$(kwallet-query -r google-genai kdewallet)

# pass (password-store)
export GOOGLE_API_KEY=$(pass show google-genai)
```

## Flags

### AI flags (shared with openrouter)

| Flag | Default | Description |
|------|---------|-------------|
| `--ai.prompt` | | Image description prompt (required) |
| `--ai.model` | `auto` | AI model to use for image generation |
| `--ai.list-models` | `false` | List available models and exit |
| `--ai.no-extended-prompt` | `false` | Disable automatic wallpaper prompt enhancement |
| `--ai.no-negative-prompt` | `false` | Disable default negative prompt |
| `--ai.negative-prompt` | | Custom negative prompt to discourage certain elements |

### Plugin-specific flags

| Flag | Default | Description |
|------|---------|-------------|
| `--googlegenai.backend` | `gemini-api` | Backend: `gemini-api` or `vertex-ai` |
| `--googlegenai.image-size` | `2K` | Image size: `1K` or `2K` (Imagen models only) |

### Common flags

| Flag | Default | Description |
|------|---------|-------------|
| `--aspect-ratio` | `16:9` | Image aspect ratio (`1:1`, `3:4`, `4:3`, `9:16`, `16:9`, `21:9`) |
| `--count` | `32` | Number of colours to extract |
| `--seed-mode` | `content` | Seed mode: `content`, `manual`, `random` |

## Available models

| Model | Description |
|-------|-------------|
| `gemini-2.5-flash-image` | Default. Gemini with native image output |
| `imagen-3.0-generate-002` | Imagen 3 |
| `imagen-4.0-fast-generate-001` | Imagen 4 Fast |
| `imagen-4.0-generate-001` | Imagen 4 |
| `imagen-4.0-ultra-generate-001` | Imagen 4 Ultra |

List models with live pricing:

```bash
tinct generate -i google-genai --ai.list-models
```

## Examples

### Basic generation

```bash
export GOOGLE_API_KEY="your-api-key"

# Nature scene
tinct generate -i google-genai \
  --ai.prompt "sunset over rolling hills of tuscany" \
  -o all

# Abstract
tinct generate -i google-genai \
  --ai.prompt "abstract geometric patterns in blue and purple" \
  -o all

# Cityscape
tinct generate -i google-genai \
  --ai.prompt "neon-lit city street at night in the rain" \
  -o all
```

### With specific model and aspect ratio

```bash
tinct generate -i google-genai \
  --ai.prompt "cyberpunk city at night" \
  --ai.model "imagen-4.0-ultra-generate-001" \
  --aspect-ratio "21:9" \
  -o all
```

### With specific outputs

```bash
tinct generate -i google-genai \
  --ai.prompt "misty forest at dawn" \
  -o hyprland,kitty,waybar
```

### Force theme type

```bash
# Ensure dark theme
tinct generate -i google-genai \
  --ai.prompt "northern lights over snowy mountains" \
  -o all -t dark

# Ensure light theme
tinct generate -i google-genai \
  --ai.prompt "bright sunny beach with white sand" \
  -o all -t light
```

### Dry run

```bash
tinct generate -i google-genai \
  --ai.prompt "autumn forest" \
  --dry-run -o kitty
```

## How it works

1. **Prompt submission**: Send description to Gemini (prompt is automatically enhanced with wallpaper-optimised instructions unless `--ai.no-extended-prompt` is set)
2. **Image generation**: Gemini generates image using Imagen
3. **Caching**: Image saved locally
4. **Extraction**: K-means clustering on generated image
5. **Categorisation**: Standard colour role assignment
6. **Wallpaper**: Image provided to output plugins

## Generated image location

Generated images are cached in:

```
~/.cache/tinct/generated/google-genai/
```

The most recent generation is used as the wallpaper.

## Prompt tips

### For specific moods

| Mood | Prompt elements |
|------|-----------------|
| Calm | "serene", "peaceful", "misty", "soft" |
| Energetic | "vibrant", "dynamic", "bold", "neon" |
| Warm | "sunset", "golden hour", "amber", "autumn" |
| Cool | "twilight", "moonlit", "blue hour", "winter" |

### For better extraction

- Be specific about dominant colours you want
- Include lighting conditions (dawn, dusk, noon)
- Mention textures and materials
- Avoid overly complex scenes

### Example prompts

```bash
# Rich warm tones
--ai.prompt "golden hour over lavender fields in provence"

# Cool blue palette
--ai.prompt "icebergs floating in arctic waters under aurora"

# Earthy natural tones
--ai.prompt "ancient redwood forest with morning mist"

# Cyberpunk
--ai.prompt "rainy tokyo street with neon signs at night"
```

## API usage

Each generation makes one API call. Generation time varies (typically 5-15 seconds).

Monitor your API usage at [Google AI Studio](https://aistudio.google.com/).

## Troubleshooting

### API key not working

- Verify key at AI Studio
- Check GOOGLE_API_KEY is exported
- Ensure key has Gemini API access

### Poor colour extraction

- Try more specific prompts
- Include colour keywords
- Specify lighting conditions

### Generation fails

- Check internet connection
- Verify API quota
- Try simpler prompt

## See also

- [image](./image.md) - Extract from existing images
- [openrouter](./openrouter.md) - Alternative AI generation via OpenRouter
