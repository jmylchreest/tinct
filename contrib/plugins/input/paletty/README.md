# Paletty Input Plugin

A Tinct external input plugin that fetches curated terminal palettes from
[paletty.dev](https://paletty.dev).

Uses Tinct's plugin protocol 0.3.0+, which allows external input plugins to
return role hints and a theme hint alongside the palette colours.

## Features

- Accepts a paletty palette ID *or* a full paletty.dev URL
- Built-in default role mapping (background, foreground, danger, success,
  warning, info, notification)
- Extra mappings via plugin args
- Theme hint surfaced from paletty's `is_dark` metadata

## Installation

### From source

```bash
cd contrib/plugins/input/paletty
go build -o tinct-plugin-paletty
tinct plugins add ./tinct-plugin-paletty --type input
tinct plugins enable paletty
```

## Usage

### By URL (easiest — paste from your browser)

```bash
tinct generate -i paletty \
  --plugin-arg palette=https://paletty.dev/p/MDRHC0lqRj/midnight-a \
  -o ghostty
```

### By ID

```bash
tinct generate -i paletty --plugin-arg palette=MDRHC0lqRj -o ghostty
```

### Override accent slots

The 6 ANSI slots that aren't covered by the default mapping (cyan, white, all
bright variants, cursor and selection colours) are still added to the palette
so Tinct's categoriser can pick accents from them. To pin one explicitly,
extend the role map:

```bash
tinct generate -i paletty \
  --plugin-arg palette=MDRHC0lqRj \
  --plugin-arg map='{"accent1":"ansi.normal.6","accent2":"ansi.bright.5"}' \
  -o kitty
```

## Plugin arguments

| Argument | Type | Default | Description |
|----------|------|---------|-------------|
| `palette` | string | *(required)* | paletty.dev palette ID or URL |
| `timeout` | duration | `10s` | HTTP timeout |
| `base_url` | string | `https://paletty.dev` | Override base URL (advanced/testing) |
| `map` | map | — | Extra role→paletty-key mappings, merged with defaults |

## Default role mapping

| Tinct role | Paletty key |
|------------|-------------|
| background | `core.background` |
| foreground | `core.foreground` |
| danger | `ansi.normal.1` (red) |
| success | `ansi.normal.2` (green) |
| warning | `ansi.normal.3` (yellow) |
| info | `ansi.normal.4` (blue) |
| notification | `ansi.normal.5` (magenta) |

## Available paletty keys

The full schema returned by `GET /api/palettes/{id}`:

- `core.background`, `core.foreground`
- `core.cursor`, `core.cursorText`
- `core.selectionBackground`, `core.selectionForeground`
- `ansi.normal.0` … `ansi.normal.7`
- `ansi.bright.0` … `ansi.bright.7`

Any can appear on the right-hand side of a `map` entry.

## Theme hint

When `--theme auto` (the default) is in effect, this plugin advises the
categoriser based on paletty's `is_dark` field. Override with `--theme dark`
or `--theme light` if you disagree with the source's grading.

## Without this plugin (remote-json fallback)

paletty.dev's API also works with the built-in `remote-json` plugin — useful
if you'd rather not install another binary:

```bash
tinct generate -i remote-json \
  --remote-json.url 'https://paletty.dev/api/palettes/MDRHC0lqRj' \
  --remote-json.query '$.colors' \
  --remote-json.map 'core.background=background,core.foreground=foreground,ansi.normal.1=danger,ansi.normal.2=success,ansi.normal.3=warning,ansi.normal.4=info,ansi.normal.5=notification' \
  -o ghostty
```

The dedicated plugin is preferred — it accepts a URL or ID directly and
supplies the same default mapping for free.

## License

MIT License - Same as Tinct.
