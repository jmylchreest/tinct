---
sidebar_position: 5
---

# remote-json

Fetch colour palettes from remote JSON URLs.

## Description

The `remote-json` plugin fetches JSON data from URLs and extracts colours. Use it for:

- Theme repositories (Catppuccin, Dracula, etc.)
- Custom theme APIs
- Shared colour configurations

## Usage

```bash
tinct generate -i remote-json --remote-json.url "<url>" [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote-json.url` | | JSON URL (required) |
| `--remote-json.query` | | JSONPath query for nested data |
| `--remote-json.timeout` | `10s` | Request timeout |

## Examples

### Catppuccin themes

```bash
# Catppuccin Mocha (dark)
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o all

# Catppuccin Latte (light)
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.latte" \
  -o all -t light

# Catppuccin Frappe
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.frappe" \
  -o all
```

### Custom JSON

```bash
# Simple colour array
tinct generate -i remote-json \
  --remote-json.url "https://example.com/theme.json" \
  -o all

# Nested structure
tinct generate -i remote-json \
  --remote-json.url "https://example.com/themes.json" \
  --remote-json.query "themes.dark.colors" \
  -o all
```

## JSON format

### Simple array

```json
["#1e1e2e", "#cdd6f4", "#89b4fa", "#f5c2e7"]
```

### Object with hex values

```json
{
  "background": "#1e1e2e",
  "foreground": "#cdd6f4",
  "accent": "#89b4fa"
}
```

### Nested structure

```json
{
  "colors": {
    "mocha": {
      "base": "#1e1e2e",
      "text": "#cdd6f4",
      "blue": "#89b4fa"
    }
  }
}
```

Use `--remote-json.query` to navigate: `colors.mocha`

## JSONPath queries

Access nested data with dot notation:

| Query | Result |
|-------|--------|
| `colors` | Top-level colours object |
| `colors.mocha` | Mocha variant |
| `themes[0]` | First theme in array |
| `dark.palette` | Nested palette |

## Colour parsing

The plugin recognises various colour formats:

| Format | Example |
|--------|---------|
| Hex 6 | `#1e1e2e` |
| Hex 3 | `#1e2` |
| RGB | `rgb(30, 30, 46)` |
| Named | `red`, `blue` |

## Role mapping

Colours are mapped to roles based on:

1. **Key names**: `background`, `foreground`, `accent`
2. **Position**: First colour as background, second as foreground
3. **Luminance**: Dark colours to background roles

### Key name mapping

| JSON key | Tinct role |
|----------|------------|
| `background`, `bg`, `base` | background |
| `foreground`, `fg`, `text` | foreground |
| `accent`, `primary`, `blue` | accent1 |
| `secondary`, `pink` | accent2 |
| `red`, `error` | danger |
| `green`, `success` | success |
| `yellow`, `warning` | warning |

## Caching

Responses are cached briefly to avoid repeated requests:

- Cache duration: 15 minutes
- Location: In-memory

For permanent caching, download and use `file` plugin.

## Troubleshooting

### JSON parse error

- Verify URL returns valid JSON
- Check JSONPath query syntax
- Ensure colours are in recognised format

### Timeout

- Increase timeout: `--remote-json.timeout 30s`
- Check network connectivity
- Verify URL is accessible

### Wrong colours

- Check JSONPath navigates to correct location
- Verify colour format in source JSON
- Use `--preview` to inspect results

## See also

- [remote-css](/docs/plugins/input/remote-css) - Parse CSS variables
- [file](/docs/plugins/input/file) - Local colour specification
