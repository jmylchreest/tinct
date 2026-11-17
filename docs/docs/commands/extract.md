---
sidebar_position: 3
---

# tinct extract

Extract colours from a source without generating output configurations.

## Synopsis

```bash
tinct extract -i <input-plugin> [flags]
```

## Description

The `extract` command extracts a colour palette without generating any configuration files. Use it to:

- Preview colour extraction results
- Export palettes to JSON for external use
- Test input plugin configurations
- Debug colour extraction issues

## Required flags

| Flag | Description |
|------|-------------|
| `-i, --input` | Input plugin to use |

## Optional flags

| Flag | Default | Description |
|------|---------|-------------|
| `-c, --image.colours` | `16` | Number of colours the image plugin extracts (1–256) |
| `-p, --image.path` | | Path, directory, or URL for the image plugin |
| `--count` | `32` | Number of colours in the resulting palette |
| `-o, --output` | | Output file path (default: stdout) |
| `-f, --format` | `palette` | Output format: `palette`, `hex`, `rgb`, `json`, `categorised` |
| `--colour` | | Colour override as `role=hex`, repeatable |
| `--preview` | `false` | Show visual preview |
| `--backend` | `kmeans` | Extraction algorithm |
| `--plugin-args` | | Arguments for an external plugin, as `<plugin>=<json>` |

Each input plugin contributes its own `--<plugin>.<flag>` options — `--image.*`
for the built-in image plugin, and for installed external plugins whatever
their `--plugin-info` declares. Run `tinct extract --help` to see the set
available on your machine.

## AI request parameters

The `google-genai`, `minimax`, `openai` and `openrouter` inputs share `--ai.params`. Pass one quoted JSON object containing fields in the selected provider's request-body format:

```bash
tinct extract -i openai \
  --ai.prompt "misty mountains at dawn" \
  --ai.params '{"quality":"high","size":"3840x2160"}' \
  --aspect-ratio 16:9 --format json --output palette.json
```

This generates and caches an image and writes a palette, without running wallpaper or theme output plugins. Generation requires API credentials and may incur provider charges.

Objects merge recursively into the generated request. Explicit scalars, arrays and `null` replace defaults; invalid JSON or a non-object value is rejected. Unknown keys are passed through for server validation. Use one `--ai.params` object for all overrides; repeated flags do not accumulate values.

| Provider | Example request-body overrides |
|----------|--------------------------------|
| [Google Gemini](../plugins/input/google-genai.md) | `{"generationConfig":{"imageConfig":{"imageSize":"2K"}}}` |
| [MiniMax H3](../plugins/input/minimax.md) | `{"size":"3840x2160","output_format":"png"}` |
| [OpenAI Images](../plugins/input/openai.md) | `{"quality":"high","size":"3840x2160"}` |
| [OpenRouter Chat Completions](../plugins/input/openrouter.md) | `{"image_config":{"aspect_ratio":"21:9"}}` |

Fields and accepted values depend on the endpoint and model; tinct does not translate them between providers. Set credentials and endpoints using environment variables and provider flags, not this JSON. Parameters affect automatic cache keys. When using `--cache-filename`, pass `--cache-overwrite` to regenerate after changing parameters.

The same flag works with `tinct generate`. See each provider's guide for model selection, limits and complete examples.

## Output formats

### palette (default)

Human-readable list with role assignments:

```
Theme type: dark

background: #1e1e2e
foreground: #cdd6f4
accent1:    #89b4fa
accent2:    #f5c2e7
...
```

### hex

Simple list of hex colours:

```
#1e1e2e
#cdd6f4
#89b4fa
#f5c2e7
```

### json

Full JSON with all metadata:

```json
{
  "colours": {
    "background": {
      "hex": "#1e1e2e",
      "rgb": {"r": 30, "g": 30, "b": 46},
      "rgba": {"r": 30, "g": 30, "b": 46, "a": 255}
    }
  },
  "theme_type": "dark"
}
```

## Examples

### Basic extraction

```bash
# Extract and display palette
tinct extract -i image -p ~/wallpaper.jpg

# Extract 8 colours
tinct extract -i image -p ~/wallpaper.jpg -c 8
```

### Save to file

```bash
# Save as JSON
tinct extract -i image -p ~/wallpaper.jpg -o palette.json --format json

# Save hex list
tinct extract -i image -p ~/wallpaper.jpg -o colors.txt --format hex
```

### Visual preview

```bash
# Show coloured preview in terminal
tinct extract -i image -p ~/wallpaper.jpg --preview
```

### From different sources

```bash
# From URL
tinct extract -i image -p "https://example.com/wallpaper.jpg"

# From remote JSON
tinct extract -i remote-json \
  --remote-json.url "https://github.com/catppuccin/palette/raw/main/palette.json" \
  --remote-json.query "colors.mocha"
```

### From an external input plugin

Installed external plugins take their arguments the same way here as on
`generate` — either as dedicated flags or as JSON:

```bash
# Dedicated flags, registered from the plugin's own --plugin-info
tinct extract -i random --random.seed 12345 --random.count 8

# The equivalent JSON form
tinct extract -i random --plugin-args random='{"seed":12345,"count":8}'
```

## Use cases

### Pipeline integration

```bash
# Get colours for a script
COLORS=$(tinct extract -i image -p ~/wallpaper.jpg --format hex)
```

### Comparison

```bash
# Compare extraction with different colour counts
tinct extract -i image -p ~/wallpaper.jpg -c 8 --preview
tinct extract -i image -p ~/wallpaper.jpg -c 16 --preview
```

### External tool input

```bash
# Export for use with other tools
tinct extract -i image -p ~/wallpaper.jpg -o palette.json --format json
```

## See also

- [generate](./generate.md) - Generate full configurations
- [Colour extraction](../concepts/color-extraction.md) - How extraction works
- [Input plugins](../plugins/input/index.md) - Available input sources
