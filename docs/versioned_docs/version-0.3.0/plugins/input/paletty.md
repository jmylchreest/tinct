---
sidebar_position: 6
---

# paletty

Fetch terminal palettes from [paletty.dev](https://paletty.dev).

paletty is shipped as an **external plugin** in
[`contrib/plugins/input/paletty`](https://github.com/jmylchreest/tinct/tree/main/contrib/plugins/input/paletty)
and uses the 0.3.0+ plugin protocol so it can pass role hints and a theme
hint back to Tinct.

## Description

The `paletty` plugin pulls a curated terminal palette from paletty.dev's
public JSON API and maps it onto Tinct's semantic roles using paletty's
known schema (core background/foreground plus the 16 ANSI slots).

## Installation

```bash
cd contrib/plugins/input/paletty
go build -o tinct-plugin-paletty
tinct plugins add ./tinct-plugin-paletty --type input
tinct plugins enable paletty
```

## Usage

```bash
tinct generate -i paletty --plugin-arg palette="<id-or-url>" [flags]
```

## Plugin arguments

| Argument | Default | Description |
|----------|---------|-------------|
| `palette` | *(required)* | Palette ID or full paletty.dev URL |
| `map` | — | Extra role→paletty-key mappings (JSON object), merged with defaults |
| `timeout` | `10s` | HTTP timeout |
| `base_url` | `https://paletty.dev` | Override base URL (advanced) |

## Examples

### Paste a URL straight from the browser

```bash
tinct generate -i paletty \
  --plugin-arg palette=https://paletty.dev/p/MDRHC0lqRj/midnight-a \
  -o ghostty,kitty
```

### Use the bare ID

```bash
tinct generate -i paletty --plugin-arg palette=MDRHC0lqRj -o all
```

### Pin extra accent slots

```bash
tinct generate -i paletty \
  --plugin-arg palette=MDRHC0lqRj \
  --plugin-arg map='{"accent1":"ansi.normal.6","accent2":"ansi.bright.5"}' \
  -o kitty
```

## Default role mapping

The plugin applies these mappings automatically:

| Tinct role | Paletty key |
|------------|-------------|
| background | `core.background` |
| foreground | `core.foreground` |
| danger | `ansi.normal.1` (red) |
| success | `ansi.normal.2` (green) |
| warning | `ansi.normal.3` (yellow) |
| info | `ansi.normal.4` (blue) |
| notification | `ansi.normal.5` (magenta) |

Remaining ANSI slots (`ansi.normal.0`, `6`, `7`, the bright variants, cursor
and selection colours) are still added to the palette so Tinct's categoriser
can pick accents from them. Use `map` to override or extend.

## Available keys for `map`

- `core.background`, `core.foreground`
- `core.cursor`, `core.cursorText`
- `core.selectionBackground`, `core.selectionForeground`
- `ansi.normal.0` … `ansi.normal.7`
- `ansi.bright.0` … `ansi.bright.7`

Map any of these to roles like `accent1`-`accent4`, `accent1Muted`, `info`,
etc.

## Theme hint

When `--theme auto` (the default) is in effect, this plugin advises the
categoriser based on paletty's `is_dark` field. Override with `--theme dark`
or `--theme light` if you disagree with the source's grading.

## Without this plugin (remote-json fallback)

paletty.dev's API also works directly with the built-in
[remote-json](./remote-json) plugin — useful if you'd rather not install
another binary, or if you want to point the same flow at a self-hosted
mirror:

```bash
tinct generate -i remote-json \
  --remote-json.url "https://paletty.dev/api/palettes/MDRHC0lqRj" \
  --remote-json.query "$.colors" \
  --remote-json.map "core.background=background,core.foreground=foreground,ansi.normal.1=danger,ansi.normal.2=success,ansi.normal.3=warning,ansi.normal.4=info,ansi.normal.5=notification" \
  -o ghostty
```

The dedicated `paletty` plugin is preferred — it accepts a URL or ID
directly, supplies the same default mapping for free, and surfaces the
`is_dark` theme hint.

## Troubleshooting

### `invalid paletty palette reference`

The value supplied to `palette` must be either a bare ID (alphanumeric, 6–32
characters) or a paletty.dev URL of the form
`https://paletty.dev/p/<id>/<slug>`.

### Palette has no colours

The API returned a response without a `colors` map — the palette ID may be
wrong or no longer exists.

### Timeouts

Increase with `--plugin-arg timeout=30s`, or check network access to
paletty.dev.

## See also

- [remote-json](./remote-json) — the generic alternative used as a fallback
- [paletty.dev](https://paletty.dev) — the gallery itself
