---
title: remote-json
sidebar_position: 4
plugin:
  type: input
  name: remote-json
  source: builtin
  source_type: url
  description: Fetch colours from remote JSON with JSONPath queries
  service: remote JSON
  requires: []
  optional: []
  requires_network: true
  requires_credentials: []
  produces_wallpaper: false
---

# remote-json

Fetches a JSON document from an HTTPS URL and extracts a colour palette from it. An optional JSONPath query selects a subtree before colour parsing, and an optional `name=role` mapping rebinds source keys to tinct's semantic roles. This is the right plugin for theme repositories (Catppuccin, Dracula, Tokyo Night) and palette services that expose a JSON endpoint.

## Installation

Built into tinct — nothing to install separately.

## Quick start

```bash
# Catppuccin Mocha from the upstream palette JSON
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o kitty,waybar
```

## Configuration

No credentials. URLs must be HTTPS by default; tinct blocks plain HTTP and link-local / private addresses to avoid SSRF surprises. To allow plain `http://` URLs (useful for local development), set `TINCT_ALLOW_INSECURE_HTTP=1`.

Responses are parsed as a single document. The plugin recognises three shapes:

```json
// 1. Flat array of hex strings
["#1e1e2e", "#cdd6f4", "#89b4fa"]

// 2. Object with hex values
{
  "background": "#1e1e2e",
  "foreground": "#cdd6f4",
  "accent":     "#89b4fa"
}

// 3. Nested object — use --remote-json.query to dive in
{
  "colors": {
    "mocha": { "base": "#1e1e2e", "text": "#cdd6f4", "blue": "#89b4fa" }
  }
}
```

Colour values may be `#RRGGBB`, `#RGB`, `rgb(r,g,b)`, or named CSS colours. Use `--remote-json.map source=role` to rebind keys when the source schema doesn't line up with tinct's role names:

```bash
# Map a paletty.dev palette into tinct roles
tinct generate -i remote-json \
  --remote-json.url "https://paletty.dev/api/palettes/MDRHC0lqRj" \
  --remote-json.query "$.colors" \
  --remote-json.map "core.background=background,core.foreground=foreground,ansi.normal.1=danger,ansi.normal.2=success,ansi.normal.3=warning,ansi.normal.4=info,ansi.normal.5=notification" \
  -o ghostty
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote-json.url` | _(required)_ | HTTPS URL of the JSON document |
| `--remote-json.query` | _(none)_ | JSONPath query to extract a subtree (e.g. `$.colors`, `colors.mocha`) |
| `--remote-json.timeout` | `10s` | HTTP timeout |
| `--remote-json.map` | `{}` | Map source keys to tinct roles (`name=role,name2=role2`) |

## Output

A raw palette of however many colours the source document exposes. Role hints are populated when keys can be matched (either directly to tinct role names or through `--remote-json.map`). Unmapped colours fall through to the auto-categoriser.

This plugin does **not** produce a wallpaper.

## Troubleshooting

### `invalid URL: …`

URLs must use HTTPS and resolve to a public address by default. Either switch to HTTPS, point at a public host, or opt out with `TINCT_ALLOW_INSECURE_HTTP=1` (use sparingly).

### `failed to parse JSON`

The response isn't JSON, or the JSONPath query selected something that isn't an array of strings or an object of hex values. Try the URL with `curl | jq` first and refine the `--remote-json.query` accordingly.

### Timeout

Increase `--remote-json.timeout` (e.g. `30s`) for slow endpoints, or check whether the host is reachable from this machine.

### Wrong roles after extraction

The plugin auto-maps common key names (`background`, `bg`, `base` → background; `accent`, `primary` → accent1; `red`/`error` → danger; etc.), but every theme repo uses its own vocabulary. When in doubt, dump the source via `curl`, then build an explicit `--remote-json.map` for the exact keys you care about.

## Related plugins

- [`remote-css`](./remote-css.md) — same idea, but parses CSS custom properties from a remote stylesheet
- [`file`](./file.md) — load colours from a local file or specify them inline
