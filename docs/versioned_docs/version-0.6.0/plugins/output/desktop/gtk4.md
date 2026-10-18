---
title: gtk4
sidebar_position: 5
plugin:
  type: output
  name: gtk4
  category: desktop
  source: builtin
  app: GTK 4
  app_url: 'https://docs.gtk.org/gtk4/'
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/gtk-4.0
  generated_files:
    - tinct-gtk4.css
  reload:
    method: none
    user_action_required: true
---

# gtk4

Generates a per-user CSS override for traditional [GTK 4](https://docs.gtk.org/gtk4/) applications — i.e. GTK 4 apps that do **not** use libadwaita. The file sets both modern `:root` CSS custom properties (`--theme-bg-color`, `--theme-selected-bg-color`, etc.) and the legacy `@define-color` symbols (`theme_bg_color`, `theme_fg_color`, `theme_selected_bg_color`, `borders`, status colours) so any GTK 4 app that consults either family picks up tinct's colours.

For libadwaita / modern GNOME apps, use the [`libadwaita`](./libadwaita.md) plugin instead — it writes a different file in the same directory.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o gtk4` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o gtk4
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-gtk4.css` | `~/.config/gtk-4.0/tinct-gtk4.css` | Palette overrides (`:root` custom properties and `@define-color` aliases) |

## Integration

GTK 4 does **not** auto-discover files dropped into `~/.config/gtk-4.0/`. The user has to import them from a `gtk.css` that GTK 4 reads at startup:

```css
/* ~/.config/gtk-4.0/gtk.css */
@import url("tinct-gtk4.css");
```

If `gtk.css` doesn't already exist, create it. The plugin's `PostExecute` checks for the import line; if it isn't there it prints a banner with the exact text to add. The plugin never edits `gtk.css` itself — that file belongs to the user.

GTK 4 does not need (and does not have) an equivalent of GTK 3's `adw-gtk3` theme for traditional apps; the default GTK 4 stylesheet honours these `@define-color` names directly.

## Reload behaviour

### Automatic

None. GTK 4 reads `gtk.css` once at application start; there is no per-process or system-wide reload mechanism for the user stylesheet.

### Manual fallback

Close and reopen any running GTK 4 application to pick up the new palette.

## Uninstall / revert

1. **Remove the include line** from `~/.config/gtk-4.0/gtk.css`:

   ```bash
   sed -i '/@import url("tinct-gtk4.css")/d' ~/.config/gtk-4.0/gtk.css
   ```

   If `gtk.css` is now empty, delete it.

2. **Delete the generated file**:

   ```bash
   rm ~/.config/gtk-4.0/tinct-gtk4.css
   ```

3. **Reload to drop the theme** — close and reopen each GTK 4 app, or restart your session.

4. **External state**: this plugin only writes to `~/.config/gtk-4.0/`. No further cleanup is required. (Note: the `libadwaita` plugin also writes into `~/.config/gtk-4.0/`; if you have both, run their uninstalls separately.)

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--gtk4.output-dir` | `~/.config/gtk-4.0` | Override the output directory |

## Colour role mapping

| CSS property / `@define-color` name | Tinct role |
|---|---|
| `--theme-bg-color` / `theme_bg_color` / `--theme-base-color` / `theme_base_color` | `background` |
| `--theme-fg-color` / `theme_fg_color` / `--theme-text-color` / `theme_text_color` | `foreground` |
| `--theme-bg-muted-color` | `backgroundMuted` |
| `--theme-fg-muted-color` | `foregroundMuted` |
| `--theme-selected-bg-color` / `theme_selected_bg_color` / `--theme-accent-color` | `accent1` |
| `--theme-selected-fg-color` / `theme_selected_fg_color` | `white` (literal, not a tinct role) |
| `--borders` / `borders` / `--theme-border-color` | `border` |
| `--warning-color` / `warning_color` | `warning` |
| `--error-color` / `error_color` | `danger` |
| `--success-color` / `success_color` | `success` |
| `--info-color` | `info` |

## Customising the template

```bash
tinct plugins templates dump -o gtk4 -l ~/.config/tinct/templates/gtk4
```

This creates `~/.config/tinct/templates/gtk4/tinct-gtk4.css.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Plugin printed `GTK4 Theme Configuration Required` banner

`PostExecute` checked `~/.config/gtk-4.0/gtk.css` and didn't find the `@import url("tinct-gtk4.css")` line. Add it as shown in [Integration](#integration), then re-run `tinct generate -o gtk4` — the banner disappears once the import is present.

### Theme generates but libadwaita apps look unchanged

This plugin only handles traditional GTK 4 apps. Modern GNOME apps (Files, Calculator, Text Editor, most Settings panels) use libadwaita, which ignores `theme_bg_color` and friends and instead consults `--window-bg-color`, `--accent-bg-color`, etc. Run the [`libadwaita`](./libadwaita.md) plugin too — it writes a separate file (`tinct-libadwaita.css`) into the same directory, and you'll need a second `@import` line for it.

### Flatpak apps don't pick up the theme

Flatpak GTK 4 apps don't read host `~/.config/gtk-4.0` by default. Either install the apps natively, or grant Flatpak access:

```bash
flatpak override --user --filesystem=~/.config/gtk-4.0:ro
```

### `Plugin skipped: gtk4 config directory does not exist`

`Hooks.RequiredDirs` requires `~/.config/gtk-4.0` to already exist. Create it: `mkdir -p ~/.config/gtk-4.0` and re-run.

## Related plugins

- [libadwaita](./libadwaita.md) — companion plugin for libadwaita apps; writes a different file in the same `~/.config/gtk-4.0/` directory.
- [gtk3](./gtk3.md) — sibling plugin for legacy GTK 3 apps.
- [gnome-shell](./gnome-shell.md) — themes the GNOME compositor itself.
