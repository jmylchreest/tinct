---
title: rofi
sidebar_position: 4

plugin:
  type: output
  name: rofi
  category: bars-launchers
  source: builtin
  app: rofi
  app_url: https://github.com/davatorium/rofi
  requires: [rofi]
  optional: []
  pattern: two-file
  default_output_dir: ~/.config/rofi/themes
  generated_files: [tinct-colours.rasi, tinct.rasi]
  reload:
    method: none
    user_action_required: false
---

# rofi

Generates a [rofi](https://github.com/davatorium/rofi) theme as two `.rasi` files: a palette (`tinct-colours.rasi`) declaring every tinct role as a named rasi colour, and a main theme (`tinct.rasi` by default) that `@import`s the palette and applies opinionated styling for `window`, `inputbar`, `listview`, `element` and friends. Drop them in the user theme directory and select the main file via `@theme "tinct"` or `rofi -theme tinct`.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o rofi` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o rofi
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colours.rasi` | `~/.config/rofi/themes/tinct-colours.rasi` | **Palette only.** Declares every tinct role as a named rasi colour inside `* { ... }`. No widget rules. Importable into your own theme via `@import "tinct-colours.rasi"`. |
| `<theme-name>.rasi` | `~/.config/rofi/themes/tinct.rasi` | **Main theme.** `@import`s the palette above, then styles `window`, `inputbar`, `listview`, `element`, etc. Filename matches `--rofi.theme-name` (default `tinct`). |

## Integration

Rofi looks for theme files in `~/.config/rofi/themes/` (and a few system locations). Select the tinct theme by adding one line to your `~/.config/rofi/config.rasi`:

```rasi
@theme "tinct"
```

Or invoke rofi with the theme on the command line, no config edit required:

```bash
rofi -show drun -theme tinct
```

Rofi resolves `tinct` to `~/.config/rofi/themes/tinct.rasi`, which in turn `@import`s `tinct-colours.rasi` — both files must live alongside each other.

If you prefer to write your own rasi theme but want the tinct palette, skip `@theme "tinct"` and instead `@import "tinct-colours.rasi"` from your own theme:

```rasi
@import "tinct-colours.rasi"

window {
    background-color: @surface;
    border-color:     @accent1;
}
```

## Reload behaviour

### Automatic

None. Rofi reads its theme on every launch — there is no daemon to reload.

### Manual fallback

Launch rofi again. The next invocation picks up the new theme.

## Uninstall / revert

1. **Remove the theme selector** from your `~/.config/rofi/config.rasi`:

   ```bash
   sed -i '/@theme "tinct"/d' ~/.config/rofi/config.rasi
   ```

   (Skip this step if you only ever invoked rofi with `-theme tinct` on the command line.)

2. **Delete the generated files**:

   ```bash
   rm ~/.config/rofi/themes/tinct.rasi ~/.config/rofi/themes/tinct-colours.rasi
   ```

3. **Reload to drop the theme**: no action needed. The next rofi invocation uses your previous theme.

4. **External state**: this plugin only writes to `~/.config/rofi/themes/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--rofi.output-dir` | `~/.config/rofi/themes` | Override the output directory |
| `--rofi.theme-name` | `tinct` | Theme filename (without extension). Output is `<theme-name>.rasi`. |

## Colour role mapping

Every tinct role is declared as a named rasi colour in the `* { ... }` block of `tinct-colours.rasi`; the main theme `@import`s it and references the names via `@<name>`.

| rasi name | Tinct role |
|---|---|
| `background` | `background` |
| `background-muted` | `backgroundMuted` |
| `foreground` | `foreground` |
| `foreground-muted` | `foregroundMuted` |
| `surface` | `surface` |
| `on-surface` | `onSurface` |
| `surface-variant` | `surfaceVariant` |
| `on-surface-variant` | `onSurfaceVariant` |
| `surface-container-lowest` | `surfaceContainerLowest` |
| `surface-container-low` | `surfaceContainerLow` |
| `surface-container` | `surfaceContainer` |
| `surface-container-high` | `surfaceContainerHigh` |
| `surface-container-highest` | `surfaceContainerHighest` |
| `outline` | `outline` |
| `outline-variant` | `outlineVariant` |
| `border` | `border` |
| `accent1` … `accent4` | `accent1` … `accent4` |
| `on-accent1` … `on-accent4` | `onAccent1` … `onAccent4` |
| `danger` | `danger` |
| `on-danger` | `onDanger` |
| `warning` | `warning` |
| `on-warning` | `onWarning` |
| `success` | `success` |
| `on-success` | `onSuccess` |
| `info` | `info` |
| `on-info` | `onInfo` |

Widget defaults wired in the template:

| rofi widget | Property | Resolved role |
|---|---|---|
| `window` | `background-color` | `surface` |
| `window` | `border-color` | `accent1` |
| `inputbar` | `background-color` | `surfaceContainer` |
| `inputbar` | `text-color` | `onSurface` |
| `prompt` | `text-color` | `accent1` |
| `element selected` | `background-color` / `text-color` | `accent1` / `onAccent1` |
| `element selected.urgent` | `background-color` / `text-color` | `danger` / `onDanger` |
| `message` | `background-color` / `text-color` | `surfaceContainerLow` / `onSurfaceVariant` |

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o rofi -l ~/.config/tinct/templates/rofi
```

This writes `~/.config/tinct/templates/rofi/tinct.rasi.tmpl` and `~/.config/tinct/templates/rofi/tinct-colours.rasi.tmpl`. Tinct uses your versions in preference to the embedded defaults; override one, both, or neither.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### `Failed to parse theme: tinct`

Rofi looks for `<name>.rasi` in `~/.config/rofi/themes/`. Confirm the file landed there:

```bash
ls -l ~/.config/rofi/themes/tinct.rasi
rofi -dump-theme -theme tinct | head
```

If you used `--rofi.theme-name=other`, the file is `other.rasi` and you select it with `@theme "other"` or `rofi -theme other`.

### Theme didn't change

You're either still using the old rofi process (rofi exits after every menu, but if it didn't, kill it) or your `config.rasi` references a different theme via `@theme`. Run `rofi -dump-theme | head -5` to see which theme rofi is actually loading.

### `rofi: command not found`

The plugin lists `rofi` as a required binary. Install it via your distro's package manager.

## Related plugins

- [fuzzel](../fuzzel/README.md) — Wayland-native launcher alternative.
- [wofi](../wofi/README.md) — another Wayland-native launcher.
- [walker](../walker/README.md) — modern modular launcher.
