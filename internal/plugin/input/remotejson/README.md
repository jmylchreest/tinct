# RemoteJSON Input Plugin

**Type:** Input Plugin  
**Built-in:** Yes  
**Language:** Go

Fetch color palettes from remote JSON APIs with optional JSONPath queries.

## Overview

The `remotejson` plugin fetches color palettes from HTTP(S) JSON endpoints. It supports JSONPath queries for extracting colors from complex JSON structures and can map source color names to Tinct's semantic roles.

## Features

- ✅ Fetch from any HTTP(S) JSON endpoint
- ✅ JSONPath queries for complex JSON structures
- ✅ Color format auto-detection (hex, rgb, hsl)
- ✅ Role mapping (map source names to Tinct roles)
- ✅ Configurable timeout
- ✅ Perfect for theme repositories (Catppuccin, Dracula, etc.)

## Usage

### Basic Fetch

```bash
tinct generate -i remote-json \
  --remote-json.url "https://example.com/palette.json" \
  -o hyprland,kitty
```

### With JSONPath Query

```bash
tinct generate -i remote-json \
  --remote-json.url "https://api.example.com/themes/dark.json" \
  --remote-json.query "$.colors" \
  -o hyprland
```

### With Role Mapping

```bash
tinct generate -i remote-json \
  --remote-json.url "https://example.com/theme.json" \
  --remote-json.map "base=background,text=foreground,blue=accent1" \
  -o hyprland
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote-json.url` | *(required)* | JSON URL to fetch |
| `--remote-json.query` | *(optional)* | JSONPath query (e.g., `$.colors`) |
| `--remote-json.timeout` | `10s` | HTTP request timeout |
| `--remote-json.map` | `{}` | Map source names to roles |

## Supported Color Formats

The plugin auto-detects and parses:
- **Hex:** `#RRGGBB`, `RRGGBB`, `#RGB`
- **RGB:** `rgb(R, G, B)`, `{"r": R, "g": G, "b": B}`
- **HSL:** `hsl(H, S%, L%)`

## Examples

### Catppuccin Mocha

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.mocha.colors" \
  -o hyprland,kitty
```

### Custom API

```bash
tinct generate -i remote-json \
  --remote-json.url "https://myapi.com/theme/dark" \
  --remote-json.map "bg=background,fg=foreground,primary=accent1" \
  -o hyprland
```

## Related

- **[Input Plugin Guide](../README.md)**
- **[RemoteCSS Plugin](../remotecss/README.md)**
- **[File Plugin](../file/README.md)**

## License

MIT License - Same as Tinct
