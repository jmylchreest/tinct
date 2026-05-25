---
title: walker
sidebar_position: 1

plugin:
  type: output
  name: walker
  category: bars-launchers
  source: builtin
  app: Walker
  app_url: https://github.com/abenz1267/walker
  requires: [walker]
  optional: []
  pattern: two-file
  default_output_dir: ~/.config/walker/themes
  generated_files: [tinct-colours.css, tinct/style.css]
  reload:
    method: none
    user_action_required: true
---

# walker

Generates a GTK CSS theme for [Walker](https://github.com/abenz1267/walker), a Wayland-native application launcher. The output is a pair of stylesheets — a colour-only file using GTK's `@define-color` directives plus a complete launcher stylesheet that imports those colour names — both written under Walker's `themes/` directory.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o walker` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o walker
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colours.css` | `~/.config/walker/themes/tinct-colours.css` | **Palette.** GTK `@define-color` block. Imported by `tinct/style.css` and reusable from any custom Walker theme. |
| `tinct/style.css` | `~/.config/walker/themes/tinct/style.css` | **Theme.** Complete Walker stylesheet — window, search input, list, item rows, scrollbars — that consumes the palette via `@import '../tinct-colours.css'`. |

## Generated format

### `tinct-colours.css`

```css
:root {
    --background: #1e1e2e;
    --foreground: #cdd6f4;
    --surface: #313244;
    --accent1: #89b4fa;
    --accent2: #f5c2e7;
    /* ... additional colours */
}
```

### `tinct/style.css`

```css
@import url("../tinct-colours.css");

window {
    background-color: var(--background);
    border-radius: 12px;
}

#search {
    background-color: var(--surface);
    color: var(--foreground);
}

.item:selected {
    background-color: var(--accent1);
}
```

## Integration

Walker selects a theme by directory name. The plugin writes its stylesheet to `~/.config/walker/themes/tinct/style.css`, so the theme name to point Walker at is `tinct`.

Set this once in Walker's `config.toml`:

```toml
theme = "tinct"
```

Walker will pick up `themes/tinct/style.css` (and through it, the shared `tinct-colours.css`) the next time it starts. If you maintain your own Walker stylesheet, you can ignore `tinct/style.css` and instead `@import '../tinct-colours.css'` from your own theme to use just the palette.

## Reload behaviour

### Automatic

None. Walker does not expose a hot-reload IPC or signal, so the plugin does not attempt one — generation writes the files and exits.

### Manual fallback

Walker reads its stylesheet on launch. To pick up a regenerated theme:

```bash
pkill -x walker
```

Walker is normally launched on-demand via a keybind, so the next invocation will load the updated CSS automatically.

## Uninstall / revert

1. **Remove the config line** from your Walker config (only if you set `theme = "tinct"`):

   ```bash
   sed -i '/theme = "tinct"/d' ~/.config/walker/config.toml
   ```

   If you were importing only the palette from your own theme, drop the `@import '../tinct-colours.css';` line from your stylesheet instead.

2. **Delete the generated files**:

   ```bash
   rm -rf ~/.config/walker/themes/tinct ~/.config/walker/themes/tinct-colours.css
   ```

3. **Reload to drop the theme**: kill Walker so the next launch re-reads the (now non-Tinct) theme:

   ```bash
   pkill -x walker
   ```

4. **External state**: this plugin only writes to `~/.config/walker/themes/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--walker.output-dir` | `~/.config/walker/themes` | Override the output directory |

## Colour role mapping

| GTK colour name | Tinct role | Notes |
|---|---|---|
| `@window_bg_color` | `background` | Launcher window background |
| `@theme_fg_color` | `foreground` | Default text colour |
| `@accent_bg_color` | `accent1` | Selection / hover highlight, item borders |
| `@error_bg_color` | `danger` | Error banner background |
| `@error_fg_color` | `onDanger` | Error banner text |

`tinct/style.css` derives further shades (input background, hover states, keybind footer) via GTK's `lighter()`/`darker()`/`alpha()` helpers, so the above five roles drive the whole launcher's appearance.

## Customising the template

Extract the default templates to override them:

```bash
tinct plugins templates dump -o walker -l ~/.config/tinct/templates/walker
```

This creates `~/.config/tinct/templates/walker/tinct-colours.css.tmpl` and `~/.config/tinct/templates/walker/tinct.css.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Theme not applied after generation

Confirm `theme = "tinct"` is set in `~/.config/walker/config.toml` and that the directory exists:

```bash
ls -l ~/.config/walker/themes/tinct/style.css
```

Then kill any running Walker instance (`pkill -x walker`); the next launch will pick up the new stylesheet.

### Colours look wrong / partially themed

`tinct/style.css` relies on `tinct-colours.css` being at `../tinct-colours.css`. If you moved or renamed either file, GTK falls back to its defaults for missing variables. Re-running `tinct generate -o walker` restores the expected layout.

### Custom Walker stylesheet stops loading

Walker resolves `theme` by directory name. If you renamed `themes/tinct/` or set `theme` to a directory that has no `style.css`, Walker will silently fall back to its built-in theme. Either restore the file or update `config.toml` to point at your custom theme.

## Related plugins

- [waybar](../waybar/README.md) — status bar styling that shares the `@define-color` GTK CSS pattern.
- [wofi](../wofi/README.md) — alternative GTK-based launcher.
- [fuzzel](../fuzzel/README.md) — non-GTK launcher for users who prefer a simpler config.
