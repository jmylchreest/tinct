---
title: gtk3
sidebar_position: 4

plugin:
  type: output
  name: gtk3
  category: desktop
  source: builtin
  app: GTK 3
  app_url: https://docs.gtk.org/gtk3/
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/gtk-3.0
  generated_files: [tinct-gtk3.css]
  reload:
    method: none
    user_action_required: true
---

# gtk3

Generates a per-user CSS override for [GTK 3](https://docs.gtk.org/gtk3/) applications (GIMP, Inkscape, Pidgin, and most non-libadwaita GNOME 3-era apps). The file defines the `adw-gtk3` / libadwaita-compatible `@define-color` symbols — `window_bg_color`, `accent_bg_color`, `sidebar_bg_color`, and so on — so any GTK 3 app that respects those palette names picks up tinct's colours.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o gtk3` works out of the box.

To actually see the colours applied you also need the [adw-gtk-theme](https://github.com/lassekongo83/adw-gtk3) (a.k.a. `adw-gtk3`) installed and selected as the GTK 3 theme — see [Integration](#integration).

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o gtk3
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-gtk3.css` | `~/.config/gtk-3.0/tinct-gtk3.css` | Palette overrides; defines `@define-color` for the libadwaita-style colour names |

## Integration

GTK 3 does **not** auto-discover files dropped into `~/.config/gtk-3.0/`. Two manual steps are required, both of which the plugin will print at the end of generation if missing:

1. **Import the palette** from `~/.config/gtk-3.0/gtk.css` (create the file if it doesn't exist):

   ```css
   @import url("tinct-gtk3.css");
   ```

2. **Install the `adw-gtk3` theme and select it** in `~/.config/gtk-3.0/settings.ini`. The colour overrides only take effect when the active theme actually consults the libadwaita palette names tinct sets; the default Adwaita GTK 3 theme does not.

   ```bash
   # Arch / CachyOS
   sudo pacman -S adw-gtk-theme

   # Debian / Ubuntu
   sudo apt install adw-gtk3
   ```

   ```ini
   # ~/.config/gtk-3.0/settings.ini
   [Settings]
   gtk-theme-name=adw-gtk3-dark
   ```

The plugin's `PostExecute` inspects both files: if either the `@import` line is missing or the theme isn't `adw-gtk3`-something, it prints a banner explaining what to add. There is no automatic edit — both files belong to the user.

## Reload behaviour

### Automatic

None. GTK 3 reads `gtk.css` and `settings.ini` once at application start; there is no per-process or system-wide reload mechanism for them.

### Manual fallback

Close and reopen any running GTK 3 application to pick up the new palette. The colours are baked into the GtkStyleContext at app launch.

## Uninstall / revert

1. **Remove the include line** from `~/.config/gtk-3.0/gtk.css`:

   ```bash
   sed -i '/@import url("tinct-gtk3.css")/d' ~/.config/gtk-3.0/gtk.css
   ```

   If `gtk.css` is now empty, delete it.

2. **Delete the generated file**:

   ```bash
   rm ~/.config/gtk-3.0/tinct-gtk3.css
   ```

3. **Reload to drop the theme** — close and reopen each GTK 3 app, or restart your session.

4. **External state**: optionally revert your GTK 3 theme back to the default Adwaita by editing `~/.config/gtk-3.0/settings.ini` (`gtk-theme-name=Adwaita`) and uninstalling `adw-gtk-theme` if you no longer want it. The tinct plugin itself never writes to `settings.ini` — that was a manual step you took during [Integration](#integration) and only you know whether to undo it.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--gtk3.output-dir` | `~/.config/gtk-3.0` | Override the output directory |

## Colour role mapping

| `@define-color` name | Tinct role |
|---|---|
| `window_bg_color` | `background` |
| `window_fg_color` | `foreground` |
| `view_bg_color` | `surface` |
| `view_fg_color` | `onSurface` |
| `accent_bg_color` / `accent_color` | `accent1` |
| `accent_fg_color` | `onAccent1` |
| `headerbar_bg_color` | `surface` |
| `headerbar_fg_color` | `onSurface` |
| `headerbar_backdrop_color` | `surfaceVariant` |
| `headerbar_shade_color` | `border` |
| `card_bg_color` / `popover_bg_color` / `dialog_bg_color` | `surface` |
| `card_fg_color` / `popover_fg_color` / `dialog_fg_color` | `onSurface` |
| `card_shade_color` | `border` |
| `sidebar_bg_color` / `sidebar_backdrop_color` | `surfaceVariant` |
| `sidebar_fg_color` | `onSurfaceVariant` |
| `sidebar_shade_color` | `border` |
| `warning_bg_color` / `warning_color` | `warning` |
| `warning_fg_color` | `onWarning` |
| `error_bg_color` / `error_color` / `destructive_bg_color` / `destructive_color` | `danger` |
| `error_fg_color` / `destructive_fg_color` | `onDanger` |
| `success_bg_color` / `success_color` | `success` |
| `success_fg_color` | `onSuccess` |

## Customising the template

```bash
tinct plugins templates dump -o gtk3 -l ~/.config/tinct/templates/gtk3
```

This creates `~/.config/tinct/templates/gtk3/tinct-gtk3.css.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Plugin printed `GTK3 Theme Configuration Required` banner

`PostExecute` ran but detected that either `gtk.css` doesn't `@import` the palette, or `settings.ini` isn't using an `adw-gtk3*` theme. Apply both fixes in the [Integration](#integration) section and re-run `tinct generate -o gtk3`. The banner will disappear once both conditions are satisfied.

### Colours look slightly off / only partly applied

The default GNOME `Adwaita` GTK 3 theme ignores the libadwaita palette names. You need `adw-gtk3-dark` (or the light variant) selected in `settings.ini`. Verify with:

```bash
grep gtk-theme-name ~/.config/gtk-3.0/settings.ini
```

If that still doesn't help, the app you're looking at may be running through Flatpak with its own bundled GTK 3 stack — Flatpak apps don't read `~/.config/gtk-3.0` by default. Use `flatpak override --user --filesystem=~/.config/gtk-3.0:ro` to expose the host config.

### `Plugin skipped: gtk3 config directory does not exist`

`Hooks.RequiredDirs` requires `~/.config/gtk-3.0` to already exist (taken as a signal that GTK 3 is installed). Create it: `mkdir -p ~/.config/gtk-3.0` and re-run.

## Related plugins

- [gtk4](../gtk4/README.md) — sibling plugin for the GTK 4 stack.
- [libadwaita](../libadwaita/README.md) — for the modern libadwaita-only apps; writes alongside gtk4.
- [gnome-shell](../gnome-shell/README.md) — themes the GNOME compositor itself.
