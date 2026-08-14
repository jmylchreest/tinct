---
title: paletty
sidebar_position: 8

plugin:
  type: input
  name: paletty
  source: external
  source_type: api
  description: Fetch curated terminal palettes from paletty.dev
  service: paletty.dev
  service_url: https://paletty.dev
  version: 0.1.0
  protocol_version: 0.3.0
  repository: https://github.com/jmylchreest/tinct
  install: tinct plugins install paletty
  requires: []
  optional: []
  requires_network: true
  requires_credentials: []
  produces_wallpaper: false
---

# paletty

Fetches a curated terminal palette from [paletty.dev](https://paletty.dev)'s public JSON API and returns it as a tinct palette. Each paletty palette comes with a complete ANSI 16-colour set plus core background/foreground/cursor/selection slots; the plugin maps the conventional roles (`background`, `foreground`, `danger`, `success`, `warning`, `info`, `notification`) onto sensible defaults and surfaces paletty's `is_dark` metadata as a theme hint so the categoriser doesn't have to guess.

## Installation

### Via the official plugin repository

```bash
tinct plugins install paletty
```

### Build from source

```bash
cd contrib/plugins/input/paletty
go build -ldflags "-X main.Version=0.1.0" -o tinct-plugin-paletty
install -m 0755 tinct-plugin-paletty ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-paletty
tinct-plugin-paletty --plugin-info | jq .
```

The plugin uses tinct's go-plugin RPC protocol and is discovered automatically once it's on `$PATH`.

## Quick start

```bash
# By URL — paste from your browser
tinct generate -i paletty \
  --paletty.palette https://paletty.dev/p/MDRHC0lqRj/midnight-a \
  -o ghostty

# By palette ID
tinct generate -i paletty --paletty.palette MDRHC0lqRj -o ghostty

# Preview the palette without writing theme files
tinct extract -i paletty --paletty.palette MDRHC0lqRj
```

## Configuration / credentials

No credentials are required — paletty.dev exposes its palette catalogue via a public read-only API. The plugin needs network access to reach `https://paletty.dev`; behind a corporate proxy, set `HTTPS_PROXY` in the environment that runs tinct.

The plugin accepts either a paletty palette ID (e.g. `MDRHC0lqRj`) or a full paletty.dev URL — IDs are extracted from URLs of the form `/p/<id>/<slug>` or `/api/palettes/<id>`.

### Default role mapping

Paletty palettes ship 16 ANSI slots plus core entries. By default the plugin pins these roles:

| Tinct role | Paletty key |
|------------|-------------|
| `background` | `core.background` |
| `foreground` | `core.foreground` |
| `danger` | `ansi.normal.1` (red) |
| `success` | `ansi.normal.2` (green) |
| `warning` | `ansi.normal.3` (yellow) |
| `info` | `ansi.normal.4` (blue) |
| `notification` | `ansi.normal.5` (magenta) |

The remaining ANSI slots (cyan/white, bright variants, cursor/selection) are still added to the palette so tinct's categoriser can draw accents from them. Pin specific accents explicitly via the `map` arg:

```bash
tinct generate -i paletty \
  --paletty.palette MDRHC0lqRj \
  --paletty.map '{"accent1":"ansi.normal.6","accent2":"ansi.bright.5"}' \
  -o kitty
```

### Available paletty keys

The full schema returned by `GET /api/palettes/{id}`:

- `core.background`, `core.foreground`
- `core.cursor`, `core.cursorText`
- `core.selectionBackground`, `core.selectionForeground`
- `ansi.normal.0` … `ansi.normal.7`
- `ansi.bright.0` … `ansi.bright.7`

Any of these can appear on the right-hand side of a `map` entry.

## Flags

Once installed, tinct registers this plugin's arguments as `--paletty.<arg>`
flags. They can also be supplied as JSON via
`--plugin-args paletty='{"palette":"MDRHC0lqRj"}'`.

| Flag | Type | Default | Description |
|----------|------|---------|-------------|
| `--paletty.palette` | string | _(required)_ | paletty.dev palette ID or full URL |
| `--paletty.timeout` | duration | `10s` | HTTP timeout for the fetch |
| `--paletty.base_url` | string | `https://paletty.dev` | Override the base URL (advanced/testing) |
| `--paletty.map` | JSON object | _(none)_ | Extra role → paletty-key mappings merged on top of the defaults |

## Output

A palette of up to 22 colours (the full paletty palette in deterministic key order), with the default role mapping applied via tinct's protocol 0.3.0 role-hint mechanism. The theme hint is derived from paletty's `is_dark` field — override with `--theme dark` or `--theme light` if you disagree with the source's grading.

This plugin **does not provide a wallpaper**. Combine it with the `image` or `google-genai` input plugins (running in a separate invocation) if you want a wallpaper alongside the palette.

## Troubleshooting

### `plugin-arg 'palette' is required`

Pass either a palette ID or a paletty.dev URL via `--paletty.palette`. IDs are alphanumeric tokens of 6–32 characters; URLs of the form `https://paletty.dev/p/<id>/<slug>` and `https://paletty.dev/api/palettes/<id>` both work.

### `invalid paletty palette reference`

The argument didn't match the expected ID pattern or URL shape. Open the palette page in your browser and copy the URL directly — the ID lives in the path segment after `/p/`.

### `HTTP 404 Not Found`

The palette ID doesn't exist on paletty.dev, or the palette was deleted. Confirm the ID by loading `https://paletty.dev/api/palettes/<id>` in a browser; a 404 from that endpoint is the upstream confirming the palette is gone.

### Network errors / timeouts

The plugin enforces a 10-second HTTP timeout by default. Raise it for slow links:

```bash
tinct generate -i paletty \
  --paletty.palette MDRHC0lqRj \
  --paletty.timeout 30s \
  -o ghostty
```

If your network blocks paletty.dev outright, fall back to the built-in `remote-json` input plugin and point it at the JSON endpoint directly — see Related plugins below.

## Related plugins

- [`remote-json`](../../../../internal/plugin/input/remotejson/README.md) — generic JSON-from-URL palette source; works against paletty.dev's API if you'd rather not install a dedicated plugin.
- [`image`](../../../../internal/plugin/input/image/README.md) — extract a palette from an image instead of fetching a curated one.
