# RemoteCSS Input Plugin

**Type:** Input Plugin  
**Built-in:** Yes  
**Language:** Go

Extract colour palettes from remote CSS files by parsing CSS variables and hex colour codes.

## Overview

The `remotecss` plugin fetches CSS files from HTTP(S) URLs and extracts colours from CSS custom properties (variables) and inline hex codes. It's useful for importing colour schemes from CSS frameworks, design systems, or theme files.

## Features

- ✅ Fetch from any HTTP(S) CSS endpoint
- ✅ Parse CSS custom properties (`--colour-name: #hex`)
- ✅ Extract inline hex codes
- ✅ Parse rgb() and hsl() functions
- ✅ Role mapping (map CSS variable names to Tinct roles)
- ✅ Configurable timeout

## Usage

### Basic Fetch

```bash
tinct generate -i remote-css \
  --remote-css.url "https://example.com/theme.css" \
  -o hyprland,kitty
```

### With Role Mapping

```bash
tinct generate -i remote-css \
  --remote-css.url "https://example.com/variables.css" \
  --remote-css.map "primary=background,secondary=foreground,accent=accent1" \
  -o hyprland
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote-css.url` | *(required)* | CSS URL to fetch |
| `--remote-css.timeout` | `10s` | HTTP request timeout |
| `--remote-css.map` | `{}` | Map CSS variable names to roles |

## Supported CSS Formats

The plugin extracts colours from:

### CSS Custom Properties
```css
:root {
  --bg-primary: #1e1e2e;
  --fg-primary: #cdd6f4;
  --accent-blue: #89b4fa;
}
```

### Inline Hex Codes
```css
.theme-dark {
  background: #1e1e2e;
  colour: #cdd6f4;
}
```

### RGB/HSL Functions
```css
:root {
  --colour-bg: rgb(30, 30, 46);
  --colour-fg: hsl(227, 64%, 88%);
}
```

## Colour Format Support

- **Hex:** `#RRGGBB`, `#RGB`
- **RGB:** `rgb(R, G, B)`, `rgba(R, G, B, A)`
- **HSL:** `hsl(H, S%, L%)`, `hsla(H, S%, L%, A)`

## Examples

### Tailwind CSS

```bash
tinct generate -i remote-css \
  --remote-css.url "https://unpkg.com/tailwindcss@latest/dist/tailwind.css" \
  -o hyprland
```

### Custom Design System

```bash
tinct generate -i remote-css \
  --remote-css.url "https://mysite.com/design-system.css" \
  --remote-css.map "colour-primary=background,colour-text=foreground,colour-accent=accent1" \
  -o hyprland,kitty
```

### GitHub Dark Theme

```bash
tinct generate -i remote-css \
  --remote-css.url "https://github.githubassets.com/assets/dark.css" \
  -o hyprland
```

## How It Works

1. **Fetch CSS** - Download CSS file from URL
2. **Parse Variables** - Extract CSS custom properties
3. **Extract Colours** - Find hex codes, rgb(), hsl() values
4. **Map Roles** - Apply user-specified role mappings
5. **Return Palette** - Raw colours (categorisation happens separately)

## Role Mapping

Map CSS variable names to Tinct semantic roles:

```bash
--remote-css.map "bg-primary=background,text-primary=foreground,blue-500=accent1"
```

Without mapping, colours are extracted but not assigned to specific roles (auto-categorisation will assign them).

## Troubleshooting

### "CSS URL is required"

**Solution:**
```bash
tinct generate -i remote-css --remote-css.url "https://example.com/theme.css" -o hyprland
```

### "Failed to fetch CSS"

**Problem:** URL unreachable or invalid.

**Solution:**
- Verify URL in browser
- Check internet connection
- Increase timeout: `--remote-css.timeout 30s`

### "No colours found"

**Problem:** CSS file doesn't contain recognizable colour formats.

**Solution:** Check that CSS file contains hex codes or CSS variables with colours.

## Related

- **[Input Plugin Guide](../README.md)**
- **[RemoteJSON Plugin](../remotejson/README.md)**
- **[File Plugin](../file/README.md)**

## Testing

```bash
# Test with dry-run
tinct generate -i remote-css \
  --remote-css.url "https://example.com/theme.css" \
  --dry-run --verbose

# Preview extracted colours
tinct extract -i remote-css \
  --remote-css.url "https://example.com/theme.css" \
  --preview
```

## License

MIT License - Same as Tinct
