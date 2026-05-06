---
sidebar_position: 6
---

# remote-css

Extract colours from CSS variable definitions.

## Description

The `remote-css` plugin parses CSS files to extract colour variables. Use it for:

- CSS theme files
- Design system colours
- Existing CSS-based themes

## Usage

```bash
tinct generate -i remote-css --remote-css.url "<url>" [flags]
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote-css.url` | | CSS file URL (required) |
| `--remote-css.timeout` | `10s` | Request timeout |

## Examples

### Basic usage

```bash
tinct generate -i remote-css \
  --remote-css.url "https://example.com/theme.css" \
  -o all
```

### From GitHub

```bash
tinct generate -i remote-css \
  --remote-css.url "https://raw.githubusercontent.com/user/theme/main/colors.css" \
  -o all
```

## CSS format

The plugin parses CSS custom properties (variables):

```css
:root {
  --background: #1e1e2e;
  --foreground: #cdd6f4;
  --accent-primary: #89b4fa;
  --accent-secondary: #f5c2e7;
  --color-danger: #f38ba8;
  --color-warning: #f9e2af;
  --color-success: #a6e3a1;
  --color-info: #89dceb;
}
```

## Variable name mapping

Variable names are mapped to roles:

| CSS variable pattern | Tinct role |
|---------------------|------------|
| `--background`, `--bg` | background |
| `--foreground`, `--fg`, `--text` | foreground |
| `--accent`, `--primary` | accent1 |
| `--secondary` | accent2 |
| `--danger`, `--error`, `--red` | danger |
| `--warning`, `--yellow` | warning |
| `--success`, `--green` | success |
| `--info`, `--blue` | info |

## Colour formats

Recognised formats in CSS values:

| Format | Example |
|--------|---------|
| Hex 6 | `#1e1e2e` |
| Hex 3 | `#1e2` |
| Hex 8 | `#1e1e2eff` |
| RGB | `rgb(30, 30, 46)` |
| RGBA | `rgba(30, 30, 46, 1)` |
| HSL | `hsl(240, 21%, 15%)` |

## Handling complex CSS

The parser extracts from any valid CSS:

```css
/* Comments are ignored */
:root {
  --bg: #1e1e2e;
}

.theme-dark {
  --bg: #0a0a0a; /* This is also extracted */
}

@media (prefers-color-scheme: dark) {
  :root {
    --bg: #1e1e2e; /* And this */
  }
}
```

All colour values found are included in extraction.

## Combining with role hints

For precise control:

```bash
# Extract CSS then override roles
tinct generate -i remote-css \
  --remote-css.url "https://example.com/theme.css" \
  -o all
```

Or use `file` plugin with extracted values.

## Troubleshooting

### No colours found

- Verify URL returns valid CSS
- Check for CSS variables (--name: value)
- Ensure colours use recognised format

### Wrong role mapping

- Variable names may not match expected patterns
- Consider using `file` plugin for explicit mapping

### Timeout

- Increase timeout: `--remote-css.timeout 30s`
- Check URL accessibility

## See also

- [remote-json](./remote-json.md) - Parse JSON themes
- [file](./file.md) - Manual colour specification
