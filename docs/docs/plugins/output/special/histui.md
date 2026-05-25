---
title: histui
sidebar_position: 1
plugin:
  type: output
  name: histui
  category: special
  source: builtin
  app: histui
  app_url: 'https://github.com/jmylchreest/histui'
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/histui/themes
  generated_files:
    - tinct-colors.css
  reload:
    method: none
    user_action_required: true
---

# histui

Generates a GTK4 / libadwaita stylesheet for [histui](https://github.com/jmylchreest/histui), a notification daemon. The output file does two jobs at once: it overrides the named libadwaita colours used by the default histui theme, and it exposes the full tinct palette as `--tinct-*` CSS custom properties for users writing bespoke themes.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o histui` works out of the box.

The plugin skips silently if neither `histuid` is on `$PATH` nor `~/.config/histui` exists, so it's safe to leave in a multi-output invocation on hosts where histui isn't installed.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o histui
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-colors.css` | `~/.config/histui/themes/tinct-colors.css` | GTK4 CSS that `@define-color`s the libadwaita palette (window/view/accent/headerbar/popover) and exposes every tinct role as a `--tinct-*` CSS custom property under the `*` selector. |

## Integration

`tinct-colors.css` is meant to be imported by your active histui theme, not loaded directly. In `~/.config/histui/themes/theme.css` (or whichever theme histui is configured to use):

```css
@import "tinct-colors.css";
@import "default/theme.css";
```

The `@define-color` block must come **before** any theme that consumes those names — that's why `tinct-colors.css` is imported first. The default histui theme then automatically picks up the overridden window/view/accent/etc. colours.

If you're writing a custom theme, prefer the CSS custom properties:

```css
.notification-popup {
    background-color: var(--tinct-background);
    color: var(--tinct-foreground);
    border-color: var(--tinct-border);
}
```

For translucent variants use GTK4's `alpha()` function:

```css
background-color: alpha(@window_bg_color, 0.85);
```

## Reload behaviour

### Automatic

None — `Hooks().Reload` is nil. The plugin only writes the file and prints an instruction line. histui watches its theme CSS for changes on some configurations, but tinct does not signal it.

### Manual fallback

Pick whichever your setup honours:

```bash
touch ~/.config/histui/themes/theme.css   # if histui hot-reloads on theme.css mtime
```

or

```bash
systemctl --user restart histui           # full restart
```

or kill and re-launch the daemon by hand.

## Uninstall / revert

1. **Remove the import line** from your histui theme:

   ```bash
   sed -i '/@import "tinct-colors.css";/d' ~/.config/histui/themes/theme.css
   ```

2. **Delete the generated file**:

   ```bash
   rm ~/.config/histui/themes/tinct-colors.css
   ```

3. **Reload to drop the styling**:

   ```bash
   touch ~/.config/histui/themes/theme.css
   # or: systemctl --user restart histui
   ```

4. **External state**: this plugin only writes to `~/.config/histui/themes/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--histui.output-dir` | `~/.config/histui/themes` | Override the output directory |

## Colour role mapping

| CSS name / property | Tinct role | Notes |
|---|---|---|
| `@define-color window_bg_color` | `background` | libadwaita window background |
| `@define-color window_fg_color` | `foreground` | libadwaita window foreground |
| `@define-color view_bg_color` | `surface` | Content area background |
| `@define-color view_fg_color` | `onSurface` | Content area foreground |
| `@define-color accent_bg_color` | `accent1` | Accent background |
| `@define-color accent_fg_color` | `onAccent1` | Accent foreground |
| `@define-color accent_color` | `accent1` | Plain accent |
| `@define-color borders` | `border` | Border colour |
| `@define-color error_color` | `danger` | Error / destructive |
| `@define-color warning_color` | `warning` | Warning |
| `@define-color success_color` | `success` | Success |
| `@define-color card_bg_color` | `surface` | Card background |
| `@define-color card_fg_color` | `onSurface` | Card foreground |
| `@define-color headerbar_bg_color` | `surfaceContainer` | Headerbar background |
| `@define-color headerbar_fg_color` | `foreground` | Headerbar foreground |
| `@define-color popover_bg_color` | `surfaceContainerHigh` | Popover background |
| `@define-color popover_fg_color` | `foreground` | Popover foreground |
| `--tinct-background`, `--tinct-foreground` | `background`, `foreground` | Core colours |
| `--tinct-accent{1,2,3,4}` and `--tinct-accent{1,2,3,4}-muted` | accents | All four accent ramps |
| `--tinct-on-accent{1,2,3,4}` | `onAccent{1..4}` | Accessible-on-accent foregrounds |
| `--tinct-danger`, `--tinct-warning`, `--tinct-success`, `--tinct-info` | semantics | Status colours |
| `--tinct-on-danger`, `--tinct-on-warning`, `--tinct-on-success`, `--tinct-on-info` | onSemantics | Accessible foregrounds for the above |
| `--tinct-surface`, `--tinct-on-surface`, `--tinct-surface-variant`, `--tinct-on-surface-variant` | surface family | Material-style surface roles |
| `--tinct-outline`, `--tinct-outline-variant`, `--tinct-border`, `--tinct-border-muted` | outlines | Lines and dividers |
| `--tinct-inverse-surface`, `--tinct-inverse-on-surface`, `--tinct-inverse-primary` | inverse | Inverted regions (snackbars etc.) |
| `--tinct-surface-container-{lowest,low,,high,highest}` | container elevations | Material elevation levels |

## Customising the template

Dump the default template to override it:

```bash
tinct plugins templates dump -o histui -l ~/.config/tinct/templates/histui
```

This writes `~/.config/tinct/templates/histui/tinct-colors.css.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### "Plugin skipped: histuid executable not found on $PATH and ~/.config/histui does not exist"

This is the deliberate skip condition: histui isn't installed *and* hasn't been configured. Install histui (which creates `~/.config/histui` on first run) or `mkdir -p ~/.config/histui` to opt in.

### Colours don't apply after generation

Two common causes:

- Your `theme.css` doesn't `@import "tinct-colors.css";`. Without the import, the overrides never load. See [Integration](#integration).
- The `@import` line is **after** the theme that uses those colours. `@define-color` is order-sensitive — put tinct-colors.css first.

### libadwaita app outside histui picked up the colours

`@define-color` only takes effect inside the stylesheet that loaded it. tinct-colors.css is loaded by histui's theme.css; other libadwaita apps are unaffected. Use the [gtk4](../desktop/gtk4.md) or [libadwaita](../desktop/libadwaita.md) plugins to theme those.

### Wrong elevation tint on popovers

`popover_bg_color` is mapped to `surfaceContainerHigh`. If you want a flatter look, override the template and point it at `surface` or `surfaceContainer`.

## Related plugins

- [libadwaita](../desktop/libadwaita.md) — system-wide libadwaita colour overrides.
- [gtk4](../desktop/gtk4.md) — GTK4 application theming.
- [dunst](../bars-launchers/dunst.md) — alternative notification daemon for X11/Wayland users not on histui.
