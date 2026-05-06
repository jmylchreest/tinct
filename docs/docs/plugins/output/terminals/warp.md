---
sidebar_position: 6
---

# warp

Generate colour themes for the [Warp](https://www.warp.dev/) terminal.

## Description

Warp is a Rust-based terminal with a YAML-driven custom-theme system. The
plugin writes a single `tinct.yaml` file into Warp's themes directory; the
running app picks it up automatically via its filesystem watcher and
exposes it under **Settings → Appearance → Theme**.

## Output path

| Platform | Default |
|----------|---------|
| Linux | `$XDG_DATA_HOME/warp-terminal/themes/tinct.yaml` (defaults to `~/.local/share/warp-terminal/themes/tinct.yaml`) |
| macOS | `~/.warp/themes/tinct.yaml` |

Override with `--warp.output-dir`, the `TINCT_PLUGIN_WARP_OUTPUT_DIR` env
var, or a `[plugin.warp] output_dir = "…"` entry in `tinct.toml`.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--warp.output-dir` | platform default | Override the themes directory |

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o warp
```

After the file is written, open Warp's settings and pick **Tinct** from the
theme list. Warp re-reads the file on every change, so subsequent
`tinct generate` runs apply live without needing a reload.

## Generated output

```yaml
name: Tinct
accent: '#7aa2f7'
cursor: '#7aa2f7'
background: '#1a1b26'
foreground: '#c0caf5'
details: lighter
terminal_colors:
  normal:
    black:   '#15161e'
    red:     '#f7768e'
    green:   '#9ece6a'
    yellow:  '#e0af68'
    blue:    '#7aa2f7'
    magenta: '#bb9af7'
    cyan:    '#7dcfff'
    white:   '#a9b1d6'
  bright:
    black:   '#414868'
    red:     '#f7768e'
    green:   '#9ece6a'
    yellow:  '#e0af68'
    blue:    '#7aa2f7'
    magenta: '#bb9af7'
    cyan:    '#7dcfff'
    white:   '#c0caf5'
```

## Notes

- `details` is set to `lighter` for dark themes and `darker` for light
  themes — this controls how Warp draws subtle UI accents and is one of
  the few non-colour fields the schema requires.
- Hot-reload "just works": no signal or restart needed. If the theme
  doesn't appear in the picker, confirm the file is in the directory
  shown above and check Warp logs for parse errors.
