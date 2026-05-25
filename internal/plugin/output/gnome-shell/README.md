---
title: gnome-shell
sidebar_position: 3

plugin:
  type: output
  name: gnome-shell
  category: desktop
  source: builtin
  app: GNOME Shell
  app_url: https://wiki.gnome.org/Projects/GnomeShell
  requires: []
  optional: []
  pattern: flavor-pack
  default_output_dir: ~/.local/share/themes
  generated_files: [tinct-a/gnome-shell/gnome-shell.css, tinct-b/gnome-shell/gnome-shell.css]
  reload:
    method: ipc
    command: gsettings set org.gnome.shell.extensions.user-theme name tinct-a (or D-Bus Main.loadTheme())
    user_action_required: false
---

# gnome-shell

Generates a CSS theme for [GNOME Shell](https://wiki.gnome.org/Projects/GnomeShell), the compositor and panel of the GNOME desktop. The output styles the top panel, overview, dash, search, workspace switcher, notifications, system menu, OSD, and modal dialogs. Two identical theme packs (`tinct-a` and `tinct-b`) are written so the plugin can toggle between them via `gsettings` and force a flicker-free re-read of the CSS on every regeneration.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o gnome-shell` works out of the box.

The plugin can only apply the theme if the [User Themes](https://extensions.gnome.org/extension/19/user-themes/) GNOME Shell extension is **installed and enabled**. Without it, GNOME Shell refuses to load any custom shell theme. The plugin's `PreExecute` will detect a missing or disabled extension and abort generation with an actionable error message.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o gnome-shell
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-a/gnome-shell/gnome-shell.css` | `~/.local/share/themes/tinct-a/gnome-shell/gnome-shell.css` | Shell theme variant A |
| `tinct-b/gnome-shell/gnome-shell.css` | `~/.local/share/themes/tinct-b/gnome-shell/gnome-shell.css` | Shell theme variant B (identical content to A) |

Two packs are emitted because GNOME Shell does not re-read its CSS when the user-theme `name` key is set to the value it already holds. The plugin alternates between `tinct-a` and `tinct-b` on each generation so the change is observable and the new CSS is applied without a Shell restart.

## Integration

GNOME Shell discovers shell themes by scanning `~/.local/share/themes/<theme-name>/gnome-shell/gnome-shell.css` (per-user) and `/usr/share/themes/<theme-name>/gnome-shell/gnome-shell.css` (system). The plugin writes into the per-user location.

To actually apply a discovered shell theme, GNOME Shell relies on the **User Themes** extension. The plugin's prerequisites are:

1. `gnome-shell`, `gsettings`, and `gnome-extensions` on `$PATH`.
2. The User Themes extension installed at either `~/.local/share/gnome-shell/extensions/user-theme@gnome-shell-extensions.gcampax.github.com` or `/usr/share/gnome-shell/extensions/user-theme@gnome-shell-extensions.gcampax.github.com`. On most distros: `sudo pacman -S gnome-shell-extension-user-theme` / `sudo dnf install gnome-shell-extension-user-theme` / `sudo apt install gnome-shell-extension-user-theme`.
3. The extension enabled: `gnome-extensions enable user-theme@gnome-shell-extensions.gcampax.github.com` followed by a Shell restart (`Alt+F2` → `r` on X11; log out and back in on Wayland).

Once those are in place, no further manual configuration is required — `PostExecute` selects one of the two `tinct-*` packs via `gsettings`.

## Reload behaviour

### Automatic

After writing the CSS, the plugin attempts two strategies in order:

1. **D-Bus theme reload** (preferred). Invokes `Main.loadTheme()` on `org.gnome.Shell` via `org.gnome.Shell.Eval`. This re-reads the CSS in place without changing the user-theme setting. Skipped if GNOME Shell isn't on the session bus.
2. **gsettings toggle** (fallback). Reads the current `org.gnome.shell.extensions.user-theme name` value; if it is `tinct-a`, sets it to `tinct-b` (or vice-versa); if it is anything else, sets it to `tinct-a`. Toggling forces GNOME Shell to reload because the value changes.

If a wallpaper was supplied via `-p`, the plugin also sets `org.gnome.desktop.background picture-uri`, `picture-uri-dark`, and `picture-options=zoom`.

> **Wayland note**: `org.gnome.Shell.Eval` is disabled on GNOME Wayland sessions for unlocked users; the D-Bus path will fail and the gsettings toggle takes over. The toggle works on both X11 and Wayland.

### Manual fallback

If both the D-Bus path and the gsettings toggle fail (extension disabled mid-session, etc.):

```bash
gsettings set org.gnome.shell.extensions.user-theme name tinct-a
```

If GNOME Shell still doesn't pick it up, restart the Shell: `Alt+F2` → `r` (X11), or log out and back in (Wayland).

## Uninstall / revert

1. **Remove the config setting** — reset the user-theme back to GNOME's default:

   ```bash
   gsettings reset org.gnome.shell.extensions.user-theme name
   ```

2. **Delete the generated files**:

   ```bash
   rm -rf ~/.local/share/themes/tinct-a ~/.local/share/themes/tinct-b
   ```

3. **Reload to drop the theme** — the `gsettings reset` above is the reload. If the Shell doesn't pick it up immediately, restart the Shell (`Alt+F2` → `r` on X11; log out / log in on Wayland).

4. **External state**: this plugin **does mutate state outside `~/.config`**. Specifically:
   - `gsettings` key `org.gnome.shell.extensions.user-theme name` is set to `tinct-a` or `tinct-b`. Reset it as in step 1.
   - If `-p <wallpaper-path>` was passed at generate time, the plugin also wrote `org.gnome.desktop.background picture-uri`, `picture-uri-dark`, and `picture-options`. Reset these with:

     ```bash
     gsettings reset org.gnome.desktop.background picture-uri
     gsettings reset org.gnome.desktop.background picture-uri-dark
     gsettings reset org.gnome.desktop.background picture-options
     ```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--gnome-shell.output-dir` | `~/.local/share/themes/tinct/gnome-shell` (CLI help string); resolves at runtime to `~/.local/share/themes` | Override the base themes directory |

> The CLI help advertises a deeper default path, but the runtime `DefaultOutputDir()` is `~/.local/share/themes` — the plugin then writes `tinct-a/gnome-shell/...` and `tinct-b/gnome-shell/...` underneath it.

## Colour role mapping

| Selector | Property | Tinct role |
|---|---|---|
| `stage` | `color` | `foreground` |
| `#panel` | `background-color` / `color` | `background` / `foreground` |
| `.panel-button` | `color` | `foreground` |
| `.panel-button:hover` / `:active` / `:focus` / `:checked` | `background-color` / `color` | `accent1` / `onAccent1` |
| `#overview` | `background-color` | `background` |
| `#dash` / `.dash-background` | `background-color` | `surface` |
| `#dash` | `border` | `border` |
| `.dash-label` | `background-color` / `color` | `surface` / `onSurface` |
| `.app-well-app:hover`, `.show-apps:hover`, `.grid-search-result:hover` | `background-color` | `accent1` |
| `.search-entry` | `background-color` / `color` / `border-color` | `background` / `foreground` / `accent1` |
| `.workspace-background` | `background-color` | `surface` |
| `.workspace-background:hover` | `background-color` | `accent1` |
| `.window-clone-border` | `border` | `accent1` |
| `.notification-banner` | `background-color` / `color` / `border` | `background` / `foreground` / `accent1` |
| `.notification-banner:hover` | `background-color` | `surfaceVariant` |
| `.popup-menu` | `background-color` / `color` / `border` | `background` / `foreground` / `accent1` |
| `.popup-menu-item:hover` / `:focus` / `:active` | `background-color` / `color` | `accent1` / `onAccent1` |
| `.popup-separator-menu-item` | `background-color` | `outline` |
| `.button` | `background-color` / `color` / `border` | `surface` / `onSurface` / `border` |
| `.button:hover` / `:active` / `:focus` | `background-color` / `color` / `border-color` | `accent1` / `onAccent1` / `accent1` |
| `.message-list` | `background-color` / `border` | `background` / `accent1` |
| `.message` | `background-color` / `color` | `surface` / `onSurface` |
| `.message:hover` | `background-color` / `color` | `accent1` / `onAccent1` |
| `.osd-window` | `background-color` / `color` / `border` | `background` / `foreground` / `accent1` |
| `.modal-dialog` | `background-color` / `color` / `border` | `background` / `foreground` / `accent1` |

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o gnome-shell -l ~/.config/tinct/templates/gnome-shell
```

This creates `~/.config/tinct/templates/gnome-shell/gnome-shell.css.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### `User Themes extension is not installed`

`PreExecute` checked both `~/.local/share/gnome-shell/extensions/` and `/usr/share/gnome-shell/extensions/` for the `user-theme@gnome-shell-extensions.gcampax.github.com` directory and found neither. Install it from your distro's package repository (see the Integration section) or from [extensions.gnome.org](https://extensions.gnome.org/extension/19/user-themes/). After installing, enable it with `gnome-extensions enable user-theme@gnome-shell-extensions.gcampax.github.com` and restart the Shell.

### `User Themes extension is installed but not enabled`

```bash
gnome-extensions enable user-theme@gnome-shell-extensions.gcampax.github.com
```

Then log out and back in (Wayland) or `Alt+F2` → `r` (X11). Re-run `tinct generate`.

### Theme generates but Shell still looks unchanged

The D-Bus reload silently no-ops if `org.gnome.Shell` is not on the session bus (for example, the Shell crashed and was restarted as a different bus name). Run the gsettings toggle by hand:

```bash
gsettings set org.gnome.shell.extensions.user-theme name tinct-b
gsettings set org.gnome.shell.extensions.user-theme name tinct-a
```

If that still does nothing, the User Themes extension may have been disabled by a recent GNOME upgrade — re-enable it as above.

### Wallpaper didn't change

Wallpaper is only set when `-p <wallpaper-path>` is passed to `tinct generate`. The plugin sets `picture-uri`, `picture-uri-dark`, and `picture-options=zoom` via `gsettings` on a best-effort basis. Verify with:

```bash
gsettings get org.gnome.desktop.background picture-uri
```

## Related plugins

- [gtk3](../gtk3/README.md), [gtk4](../gtk4/README.md), [libadwaita](../libadwaita/README.md) — theme the applications running inside GNOME Shell.
- [dunst](../dunst/README.md) — alternative notification daemon for users who replace GNOME's notification stack.
