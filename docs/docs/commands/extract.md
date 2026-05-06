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
| `-c, --colors` | `16` | Number of colours to extract |
| `-p, --path` | | Path or URL for input |
| `-o, --output` | | Output file path |
| `--format` | `palette` | Output format: `palette`, `hex`, `json` |
| `--preview` | `false` | Show visual preview |
| `--backend` | `kmeans` | Extraction algorithm |

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
