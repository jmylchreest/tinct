---
sidebar_position: 2
---

# tinct generate

Generate theme configurations from a colour source and apply them to output plugins.

## Synopsis

```bash
tinct generate -i <input-plugin> [flags]
```

## Description

The `generate` command is the primary way to create and apply themes. It:

1. Extracts or loads colours using the specified input plugin
2. Categorises colours into semantic roles
3. Generates configuration files for each output plugin
4. Optionally triggers reload hooks

## Required flags

| Flag | Description |
|------|-------------|
| `-i, --input` | Input plugin to use (e.g., `image`, `file`, `remote-json`) |

## Common flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `all` | Output plugins (comma-separated or `all`) |
| `-p, --path` | | Path or URL for input (varies by plugin) |
| `-t, --theme` | `auto` | Theme type: `auto`, `dark`, or `light` |
| `-v, --verbose` | `false` | Enable verbose output |
| `-q, --quiet` | `false` | Suppress non-error output |
| `--preview` | `false` | Preview palette without writing files |
| `--dry-run` | `false` | Show what would be written |
| `--force` | `false` | Overwrite untracked files |
| `--save-palette` | | Save palette to JSON file |
| `--no-desaturation` | `false` | Disable colour desaturation |

## Plugin-specific flags

Each plugin may register additional flags with the format `--<plugin>.<flag>`:

```bash
# Image plugin
--image.colors 16
--image.extractAmbience
--image.ambienceRegions 8

# Markdown output
--markdown.name "My Theme"
--markdown.include-thumbnail
```

See individual plugin documentation for available flags.

## Examples

### Basic usage

```bash
# Generate from image
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all

# Generate from remote JSON
tinct generate -i remote-json \
  --remote-json.url "https://example.com/theme.json" \
  -o all
```

### Targeting specific plugins

```bash
# Only terminals
tinct generate -i image -p ~/wallpaper.jpg -o kitty,alacritty,ghostty

# Only Hyprland ecosystem
tinct generate -i image -p ~/wallpaper.jpg -o hyprland,hyprpaper,hyprlock
```

### Preview and dry-run

```bash
# See the extracted palette
tinct generate -i image -p ~/wallpaper.jpg -o all --preview

# See what files would be written
tinct generate -i image -p ~/wallpaper.jpg -o all --dry-run
```

### Force theme type

```bash
# Force dark theme
tinct generate -i image -p ~/wallpaper.jpg -o all -t dark

# Force light theme
tinct generate -i image -p ~/wallpaper.jpg -o all -t light
```

### Save palette

```bash
# Save extracted palette to JSON
tinct generate -i image -p ~/wallpaper.jpg -o all --save-palette palette.json
```

### AI generation

```bash
# Google Gemini
export GOOGLE_API_KEY="your-key"
tinct generate -i google-genai --ai.prompt "sunset over mountains" -o all

# OpenRouter
export OPENROUTER_API_KEY="your-key"
tinct generate -i openrouter --ai.prompt "cyberpunk city" -o all
```

## Output

On success, tinct outputs:

- List of generated files
- Theme type detected/used
- Any warnings or notes

With `--verbose`:

- Detailed colour extraction information
- Plugin processing details
- File write confirmations

## Exit codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | Plugin error |

## See also

- [extract](./extract.md) - Extract without generating
- [plugins](./plugins.md) - Manage plugins
- [Input plugins](../plugins/input/index.md) - Available input sources
- [Output plugins](../plugins/output/index.md) - Available outputs
