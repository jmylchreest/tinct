---
title: file
sidebar_position: 2
plugin:
  type: input
  name: file
  source: builtin
  source_type: file
  description: Load palette from file or colour overrides
  requires: []
  optional: []
  requires_network: false
  requires_credentials: []
  produces_wallpaper: false
---

# file

Loads a colour palette from a file on disk, or builds one inline from `--colour role=hex` overrides. The plugin accepts two on-disk formats: JSON (the canonical categorised-palette format tinct itself writes) and a simple text format (one hex per line, optionally `role=hex`). Use this plugin to replay a previously generated palette, hand-roll a custom palette, or pin specific roles when iterating on another input.

## Installation

Built into tinct — nothing to install separately.

## Quick start

```bash
# Load a saved categorised JSON palette
tinct generate -i file --file.path ~/.config/tinct/palettes/dracula.json -o kitty,dunst

# Build a palette inline from role overrides
tinct generate -i file \
  --colour background=#282a36 \
  --colour foreground=#f8f8f2 \
  --colour accent1=#bd93f9 \
  --colour danger=#ff5555 \
  -o kitty
```

## Configuration

No credentials, no network, no external prerequisites. You provide either `--file.path`, one or more `--colour role=hex` flags, or both (overrides take precedence over file values for the same role).

### File formats

**JSON** — a tinct `CategorisedPalette` (the format `tinct extract --json` and the `markdown` output plugin produce). Roles and the `allColours` array are preserved.

**Text** — one entry per line. Blank lines and `#` comments are ignored. Each line is either a bare hex value or `role=hex`:

```
# my-theme.txt
background = #1e1e2e
foreground = #cdd6f4
accent1    = #89b4fa
accent2    = #f5c2e7
danger     = #f38ba8
success    = #a6e3a1
warning    = #f9e2af
info       = #89dceb

# bare hex values are added to the palette without a role hint
#fab387
#94e2d5
```

Both `colour=` and `color=` spellings are accepted. `colourN`/`colorN` keys add an indexed colour without a role hint. Hex values may be `#RRGGBB`, `RRGGBB`, `#RGB`, or `RGB`.

### Recognised role names

Core semantics: `background`, `backgroundMuted`, `foreground`, `foregroundMuted`, `accent1`…`accent4`, `accent1Muted`…`accent4Muted`, `danger`, `warning`, `success`, `info`, `notification`.

Position hints for ambient lighting (used by LED output plugins): `positionTopLeft`, `positionTop`, `positionTopRight`, `positionLeft`, `positionRight`, `positionBottomLeft`, `positionBottom`, `positionBottomRight`, plus the inner/outer/centre variants for 12- and 16-region layouts. Names are case-insensitive and tolerate `_` or `-` separators (`background_muted` is equivalent to `backgroundMuted`).

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--file.path` | _(none)_ | Path to JSON or text palette file |
| `--colour` | `[]` | Inline override (`role=hex`), repeatable; also accepts `colour=` |

At least one of `--file.path` or `--colour` must be supplied.

## Output

A raw palette with whatever colours and role hints the file/overrides specify; if no role hints are present, tinct's categoriser auto-assigns roles using its standard luminance and vibrancy heuristics. This plugin does **not** produce a wallpaper — pair it with an image-based input if you need one.

## Troubleshooting

### `must provide either --file.path or --colour specifications`

Validation requires at least one source. Either point at a file or supply at least one inline override.

### `invalid colour format 'xyz': expected 'role=hex'`

Inline `--colour` flags must use `role=hex`. To add untagged colours, use a file instead — bare hex values are valid in the text format but not in `--colour`.

### `unknown colour role 'foo'`

The role name isn't in the recognised set above. Check the spelling against the role list, or use a `colourN=hex` form to add the colour without a role hint.

### Roles loaded from JSON don't match what was generated

The JSON format expects a tinct `CategorisedPalette`. Hand-edited JSON with a different shape (e.g. a raw `{"colors": [...]}` array) won't carry role hints; it falls through to the auto-categoriser. Use the text format with explicit `role=hex` lines instead, or use [`remote-json`](./remote-json.md) for arbitrary JSON shapes.

## Related plugins

- [`markdown`](./markdown.md) — load a theme exported by the `markdown` output plugin (palette + embedded wallpaper)
- [`remote-json`](./remote-json.md) — fetch arbitrary JSON from a URL and project it through a JSONPath query
- [`image`](./image.md) — extract a palette from a wallpaper instead of loading a saved one
