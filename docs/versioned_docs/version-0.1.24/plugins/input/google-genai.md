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
tinct generate -i google-genai --prompt "<description>" [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--prompt` | | Image description prompt (required) |

## Examples

### Basic generation

```bash
export GOOGLE_API_KEY="your-api-key"

# Nature scene
tinct generate -i google-genai \
  --prompt "sunset over rolling hills of tuscany" \
  -o all

# Abstract
tinct generate -i google-genai \
  --prompt "abstract geometric patterns in blue and purple" \
  -o all

# Cityscape
tinct generate -i google-genai \
  --prompt "neon-lit city street at night in the rain" \
  -o all
```

### With specific outputs

```bash
tinct generate -i google-genai \
  --prompt "misty forest at dawn" \
  -o hyprland,kitty,waybar
```

### Force theme type

```bash
# Ensure dark theme
tinct generate -i google-genai \
  --prompt "northern lights over snowy mountains" \
  -o all -t dark

# Ensure light theme
tinct generate -i google-genai \
  --prompt "bright sunny beach with white sand" \
  -o all -t light
```

## How it works

1. **Prompt submission**: Send description to Gemini
2. **Image generation**: Gemini generates image using Imagen
3. **Caching**: Image saved locally (base64)
4. **Extraction**: K-means clustering on generated image
5. **Categorisation**: Standard colour role assignment
6. **Wallpaper**: Image provided to output plugins

## Generated image location

Generated images are cached in:

```
~/.cache/tinct/genai/
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
--prompt "golden hour over lavender fields in provence"

# Cool blue palette
--prompt "icebergs floating in arctic waters under aurora"

# Earthy natural tones
--prompt "ancient redwood forest with morning mist"

# Cyberpunk
--prompt "rainy tokyo street with neon signs at night"
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

- [image](/docs/plugins/input/image) - Extract from existing images
- [openrouter](/docs/plugins/input/openrouter) - Alternative AI generation
