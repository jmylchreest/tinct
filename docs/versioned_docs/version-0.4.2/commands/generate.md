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

### Built-in plugins

Each built-in plugin registers its own cobra flags in the `--<plugin>.<flag>` namespace. They appear in `tinct generate --help` and have full type checking / completion:

```bash
# Image plugin
--image.colors 16
--image.extractAmbience
--image.ambienceRegions 8

# Markdown output
--markdown.name "My Theme"
--markdown.include-thumbnail
```

See each plugin's docs page for the available flags.

### External plugins

External plugins (anything installed via `tinct plugins install`, e.g. `paletty`, `awob`, `opencode`) don't have their flags compiled into the tinct binary, so they pass arguments through a different mechanism: the `--plugin-args` flag, with a JSON-encoded value per plugin:

```bash
# Paletty (input plugin, external) — fetch a specific palette
tinct generate -i paletty \
  --plugin-args 'paletty={"palette":"https://paletty.dev/p/MDRHC0lqRj/midnight-a"}' \
  -o all

# Spicetify (output plugin, external) with custom output dir
tinct generate -i image -p ~/wallpaper.jpg \
  -o spicetify \
  --plugin-args 'spicetify={"output-dir":"~/.config/spicetify/Themes/custom"}'
```

The key in `key=value` is the plugin name; the value is a JSON object whose keys are the args the plugin documents (see `tinct plugins list` or run the plugin's binary with `--plugin-info` to discover them).

`--plugin-args` is repeatable, one per plugin:

```bash
tinct generate -i paletty -o spicetify \
  --plugin-args 'paletty={"palette":"<url>"}' \
  --plugin-args 'spicetify={"output-dir":"<path>"}'
```

:::note Why two syntaxes?

Built-in plugins are linked into the tinct binary at compile time, so their flags can be registered with cobra and surfaced in `--help`. External plugins are separate executables discovered at runtime — tinct can't statically know their flag names, so it passes a generic key=value map through the plugin protocol. The dot-prefix flag namespace and the `--plugin-args` JSON are functionally equivalent; the difference is purely about what's knowable at compile time vs. runtime.

:::

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
