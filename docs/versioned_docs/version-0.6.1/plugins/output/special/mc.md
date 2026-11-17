---
title: mc
sidebar_position: 1
plugin:
  type: output
  name: mc
  category: special
  source: builtin
  app: Midnight Commander
  app_url: 'https://midnight-commander.org/'
  requires:
    - mc
  optional: []
  pattern: single-file
  default_output_dir: ~/.local/share/mc/skins
  generated_files:
    - tinct.ini
  reload:
    method: none
    user_action_required: true
---

# mc

Generates a [skin](https://midnight-commander.org/wiki/doc/common/skins) for [Midnight Commander](https://midnight-commander.org/), the venerable terminal file manager. The output is a single INI file dropped into MC's skin directory. Colours are converted from tinct's 24-bit hex values to the nearest xterm-256 indices via the `xterm256` template function, so the skin renders correctly on any 256-colour terminal.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o mc` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o mc
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.ini` | `~/.local/share/mc/skins/tinct.ini` | **Skin.** MC INI format covering all themable sections — `[core]` (panel, input, gauges), `[dialog]`, `[error]`, `[filehighlight]` (directory / executable / symlink / archive colours), `[menu]`, `[popupmenu]`, `[buttonbar]`, `[statusbar]`, `[help]`, `[editor]`, `[viewer]`, `[diffviewer]`, plus the line-drawing and widget glyph blocks. |

`tinct.ini` is declared `256colors = true`; on an 8-colour terminal MC will fall back to its default skin.

## Integration

MC discovers skins by filename in `$XDG_DATA_HOME/mc/skins/` (default `~/.local/share/mc/skins/`). The skin's "name" is the filename without `.ini` — so this plugin's output is selectable as `tinct`.

Three ways to apply it:

1. **Per-launch flag**:

   ```bash
   mc -S tinct
   ```

2. **Persistent in `~/.config/mc/ini`** (under the `[Midnight-Commander]` section):

   ```ini
   skin=tinct
   ```

3. **From inside MC**: `F9` → `Options` → `Appearance` → pick `tinct` from the list.

## Reload behaviour

### Automatic

None. MC reads its skin once at startup; there is no hot-reload signal, IPC, or watcher.

### Manual fallback

Quit and relaunch MC after regenerating the skin:

```bash
mc -S tinct
```

Alternatively, while inside MC, `F9` → `Options` → `Appearance` and re-select `tinct` — this re-reads the skin file without exiting the process.

## Uninstall / revert

1. **Remove the config line** from `~/.config/mc/ini`:

   ```bash
   sed -i '/^skin=tinct$/d' ~/.config/mc/ini
   ```

   Also drop `-S tinct` from any launcher script / shell alias.

2. **Delete the generated file**:

   ```bash
   rm ~/.local/share/mc/skins/tinct.ini
   ```

3. **Reload to drop the theme**: quit and relaunch MC. The next launch will fall back to the default skin (or whatever you set `skin=` to).

4. **External state**: this plugin only writes to `~/.local/share/mc/skins/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--mc.output-dir` | `~/.local/share/mc/skins` | Override the output directory |

## Colour role mapping

MC uses xterm-256 indices, not hex; the plugin converts every tinct role at template time via the `xterm256` function and emits indices like `colorN`.

| MC section | Key | Tinct role | Notes |
|---|---|---|---|
| `[core]` | `_default_` | `foreground` / `background` | Default panel colour |
| `[core]` | `selected` | `onAccent1` / `accent1` | Highlight bar |
| `[core]` | `marked` | `accent2` | Tagged files |
| `[core]` | `input` | `foreground` / `surfaceContainerLow` | Input fields |
| `[core]` | `header` | `accent1` | Panel header (bold) |
| `[dialog]` | `_default_` | `onSurface` / `surface` | Dialog background |
| `[dialog]` | `dtitle` | `accent1` | Dialog title (bold) |
| `[error]` | `_default_` | `foreground` / `danger` | Error dialog body |
| `[filehighlight]` | `directory` | `accent1` | Bold |
| `[filehighlight]` | `executable` | `success` | |
| `[filehighlight]` | `symlink` / `hardlink` | `accent3` | |
| `[filehighlight]` | `stalelink` / `core` | `danger` | Broken / coredump |
| `[filehighlight]` | `device` | `accent4` | Bold |
| `[filehighlight]` | `special` | `warning` | Sockets / pipes |
| `[filehighlight]` | `archive` | `warning` | Bold |
| `[filehighlight]` | `doc` | `info` | Documents |
| `[filehighlight]` | `source` / `database` | `accent2` | |
| `[filehighlight]` | `media` / `graph` | `accent4` | |
| `[menu]` / `[popupmenu]` | `_default_` | `onSurface` / `surface(High)` | Top menu bar |
| `[menu]` / `[popupmenu]` | `menusel` | `onAccent1` / `accent1` | Highlighted item |
| `[buttonbar]` | `button` / `hotkey` | `foreground` / `surfaceContainer`, `accent1` | F-key bar |
| `[statusbar]` | `_default_` | `foreground` / `surfaceContainer` | Bottom status line |
| `[editor]` | `editmarked` | `onAccent1` / `accent1` | Selection |
| `[editor]` | `editframeactive` | `accent1` | Active frame |
| `[editor]` | `editnonprintable` | `danger` | Control chars |
| `[diffviewer]` | `added` / `removed` / `changed` | `success` / `danger` / `info` | Diff colours |

Bold and other attributes appear after the second `;` in MC's `fg;bg;attr` triple (`accent1;;bold` for "accent1 on default with bold").

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o mc -l ~/.config/tinct/templates/mc
```

This creates `~/.config/tinct/templates/mc/tinct.ini.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors — `xterm256` is the function that maps a tinct hex colour to its nearest xterm-256 index.

## Troubleshooting

### MC ignores the skin

MC searches both `$XDG_DATA_HOME/mc/skins` (default `~/.local/share/mc/skins`) and `/usr/share/mc/skins`. Confirm the file exists at the user path:

```bash
ls -l ~/.local/share/mc/skins/tinct.ini
```

Then either invoke `mc -S tinct` or check that `skin=tinct` is under the `[Midnight-Commander]` section of `~/.config/mc/ini` (not under any other section).

### Colours look washed out / wrong

MC needs a 256-colour terminal. If you see eight-colour-style output, your terminal isn't advertising 256 colours. Set `TERM=xterm-256color` (or `tmux-256color`, etc.) and relaunch MC.

### Some elements unthemed

MC's UI has many sections; if you find a section the default template doesn't cover (`[input]` variations, `[shadow]`, etc.), dump the template (see above) and add the missing block. MC silently uses its default colours for any key the skin omits.

### Editor / viewer colours don't update

The internal editor and viewer (`mcedit`, `mcview`) inherit from the same skin, but only on launch. Reopen the file in MC after regenerating, or apply the skin live via `F9 → Options → Appearance`.

## Related plugins

- [tmux](./tmux.md) — multiplexer theming for the surrounding shell.
- [neovim](../editors/neovim.md) — external editor for users who prefer it over `mcedit`.
- [ghostty](../terminals/ghostty.md) / [kitty](../terminals/kitty.md) — terminal emulator themes so MC's surroundings match.
