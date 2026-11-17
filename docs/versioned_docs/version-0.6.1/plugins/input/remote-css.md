---
title: remote-css
sidebar_position: 5
plugin:
  type: input
  name: remote-css
  source: builtin
  source_type: url
  description: Fetch colours from remote CSS variables
  service: remote CSS
  requires: []
  optional: []
  requires_network: true
  requires_credentials: []
  produces_wallpaper: false
---

# remote-css

Fetches a CSS file from an HTTPS URL and extracts colours from CSS custom properties (variables) and inline colour values. This is the right plugin for design-system stylesheets, framework theme files, and anywhere a `:root { --colour: #hex; }` block already encodes the palette you want to use.

## Installation

Built into tinct — nothing to install separately.

## Quick start

```bash
# Pull a palette from a CSS theme file
tinct generate -i remote-css \
  --remote-css.url "https://raw.githubusercontent.com/user/theme/main/colors.css" \
  -o kitty
```

## Configuration

No credentials. URLs must be HTTPS by default; tinct blocks plain HTTP and link-local / private addresses to avoid SSRF surprises. To allow plain `http://` URLs (useful for local development), set `TINCT_ALLOW_INSECURE_HTTP=1`.

The parser looks at every CSS custom property and every inline colour value it can recognise. Selectors (`:root`, `.theme-dark`, `@media (prefers-color-scheme: dark)`) are not interpreted — every colour found is added to the palette. Variable names like `--background`, `--bg`, `--text`, `--accent`, `--danger`, `--success`, etc. are auto-mapped to tinct roles; everything else falls through to the auto-categoriser.

Recognised colour syntaxes:

| Form | Example |
|------|---------|
| Hex 6 | `#1e1e2e` |
| Hex 3 | `#1e2` |
| Hex 8 (with alpha) | `#1e1e2eff` |
| RGB / RGBA | `rgb(30, 30, 46)`, `rgba(30, 30, 46, 1)` |
| HSL / HSLA | `hsl(240, 21%, 15%)` |

Use `--remote-css.map source=role` to rebind custom-property names when the source vocabulary doesn't match tinct's:

```bash
tinct generate -i remote-css \
  --remote-css.url "https://example.com/theme.css" \
  --remote-css.map "brand-primary=accent1,brand-secondary=accent2,surface=background,on-surface=foreground" \
  -o kitty
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote-css.url` | _(required)_ | HTTPS URL of the CSS document |
| `--remote-css.timeout` | `10s` | HTTP timeout |
| `--remote-css.map` | `{}` | Map source CSS variable names to tinct roles |

## Output

A raw palette built from every colour value the parser found, with role hints derived from variable names (and `--remote-css.map`, when supplied). The unmapped colours flow through the categoriser as usual.

This plugin does **not** produce a wallpaper.

## Troubleshooting

### `invalid URL: …`

URLs must use HTTPS and resolve to a public address by default. Either switch to HTTPS, point at a public host, or opt out with `TINCT_ALLOW_INSECURE_HTTP=1` (use sparingly).

### No colours found

The stylesheet has no recognisable colour values. The parser ignores font, shadow, gradient and animation values that aren't a bare hex/rgb/hsl token; if the source encodes colours in variables that resolve through `calc()` or `color-mix()`, those won't be picked up either.

### Wrong role mapping

Custom-property names vary wildly between design systems. The auto-mapper handles common idioms (`--bg`, `--text`, `--accent`, `--primary`, `--danger`, `--red`, etc.) but leaves everything else uncategorised. Use `--remote-css.map` to pin the exact variables you care about, or fetch the file and load it through the [`file`](./file.md) plugin instead.

### Timeout

Increase `--remote-css.timeout` (e.g. `30s`) for slow endpoints, or check whether the host is reachable from this machine.

## Related plugins

- [`remote-json`](./remote-json.md) — same idea, but for JSON palettes and JSONPath queries
- [`file`](./file.md) — load colours from a local file or specify them inline
