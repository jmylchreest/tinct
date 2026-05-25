---
title: libadwaita
sidebar_position: 6

plugin:
  type: output
  name: libadwaita
  category: desktop
  source: builtin
  app: Libadwaita
  app_url: https://gnome.pages.gitlab.gnome.org/libadwaita/
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/gtk-4.0
  generated_files: [tinct-libadwaita.css]
  reload:
    method: none
    user_action_required: true
---

# libadwaita

Generates a per-user CSS override for [libadwaita](https://gnome.pages.gitlab.gnome.org/libadwaita/) — the design library used by modern GNOME apps (Files, Text Editor, Calculator, Console, Settings, and most third-party apps that follow the GNOME HIG). The file sets the libadwaita named-colour CSS custom properties (`--window-bg-color`, `--accent-bg-color`, `--sidebar-bg-color`, `--card-bg-color`, etc.) so the app palette tracks tinct's colours.

For traditional non-libadwaita GTK 4 apps, use the [`gtk4`](../gtk4/README.md) plugin instead — it writes a different file into the same directory.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o libadwaita` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o libadwaita
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct-libadwaita.css` | `~/.config/gtk-4.0/tinct-libadwaita.css` | libadwaita named-colour overrides (`:root` CSS custom properties) |

## Integration

Libadwaita reads its user overrides from `~/.config/gtk-4.0/gtk.css` (it shares this file with GTK 4). The plugin's `PostExecute` checks for the import line; if it isn't there, it prints a banner with the exact text to add. The plugin never edits `gtk.css` itself — that file belongs to the user.

```css
/* ~/.config/gtk-4.0/gtk.css */
@import url("tinct-libadwaita.css");
```

If you also use the [`gtk4`](../gtk4/README.md) plugin, you'll have two `@import` lines (order doesn't matter — they define disjoint sets of variables):

```css
@import url("tinct-gtk4.css");
@import url("tinct-libadwaita.css");
```

> **System-wide recolouring**: libadwaita apps only honour the user `gtk.css` when the [`ADW_DEBUG_HC` / accent-colour recolouring path](https://gnome.pages.gitlab.gnome.org/libadwaita/doc/main/styles-and-appearance.html#custom-styles) is in effect. Recent libadwaita (1.4+) reads the named-colour overrides directly; older versions may need `GTK_THEME=Adwaita:dark` or similar to pick them up.

## Reload behaviour

### Automatic

None at the plugin level. However, libadwaita 1.4+ ships a small live-update feature: it watches `~/.config/gtk-4.0/gtk.css` and reloads the named-colour palette in already-running apps when the file changes. This means **already-open libadwaita apps may pick up the new palette without a restart**, but only if your libadwaita version supports it. Behaviour is per-app and not guaranteed.

### Manual fallback

Close and reopen the libadwaita application to pick up the new palette deterministically.

## Uninstall / revert

1. **Remove the include line** from `~/.config/gtk-4.0/gtk.css`:

   ```bash
   sed -i '/@import url("tinct-libadwaita.css")/d' ~/.config/gtk-4.0/gtk.css
   ```

   If `gtk.css` is now empty, delete it.

2. **Delete the generated file**:

   ```bash
   rm ~/.config/gtk-4.0/tinct-libadwaita.css
   ```

3. **Reload to drop the theme** — close and reopen each libadwaita app. With libadwaita 1.4+, removing the import line and saving `gtk.css` may revert running apps automatically.

4. **External state**: this plugin only writes to `~/.config/gtk-4.0/`. No further cleanup is required. (Note: the `gtk4` plugin also writes into `~/.config/gtk-4.0/`; if you have both, run their uninstalls separately.)

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--libadwaita.output-dir` | `~/.config/gtk-4.0` | Override the output directory |

## Colour role mapping

| CSS custom property | Tinct role |
|---|---|
| `--window-bg-color` / `--view-bg-color` / `--card-bg-color` / `--dialog-bg-color` / `--popover-bg-color` / `--thumbnail-bg-color` | `background` |
| `--window-fg-color` / `--view-fg-color` / `--card-fg-color` / `--dialog-fg-color` / `--popover-fg-color` / `--thumbnail-fg-color` | `foreground` |
| `--headerbar-bg-color` / `--headerbar-backdrop-color` / `--secondary-sidebar-bg-color` / `--secondary-sidebar-backdrop-color` | `surface` |
| `--headerbar-fg-color` / `--secondary-sidebar-fg-color` | `onSurface` |
| `--headerbar-border-color` / `--secondary-sidebar-border-color` | `border` |
| `--sidebar-bg-color` / `--sidebar-backdrop-color` | `surfaceVariant` |
| `--sidebar-fg-color` | `onSurfaceVariant` |
| `--sidebar-border-color` | `border` |
| `--accent-bg-color` / `--accent-color` | `accent1` |
| `--accent-fg-color` | `onAccent1` |
| `--destructive-bg-color` / `--destructive-color` / `--error-bg-color` / `--error-color` | `danger` |
| `--destructive-fg-color` / `--error-fg-color` | `onDanger` |
| `--success-bg-color` / `--success-color` | `success` |
| `--success-fg-color` | `onSuccess` |
| `--warning-bg-color` / `--warning-color` | `warning` |
| `--warning-fg-color` | `onWarning` |
| `--headerbar-shade-color` / `--sidebar-shade-color` / `--secondary-sidebar-shade-color` / `--card-shade-color` / `--popover-shade-color` / `--shade-color` / `--scrollbar-outline-color` | static `rgba(0, 0, 0, …)` (not a tinct role) |

## Customising the template

```bash
tinct plugins templates dump -o libadwaita -l ~/.config/tinct/templates/libadwaita
```

This creates `~/.config/tinct/templates/libadwaita/tinct-libadwaita.css.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Plugin printed `Libadwaita Theme Configuration Required` banner

`PostExecute` looked at `~/.config/gtk-4.0/gtk.css` and didn't find `@import url("tinct-libadwaita.css")`. Add it as shown in [Integration](#integration), then re-run.

### Theme generates but my apps look unchanged

Older libadwaita (≤ 1.3) ignores the named-colour CSS custom properties tinct sets; only the system accent colour is honoured. Check your version:

```bash
pkg-config --modversion libadwaita-1
```

If you're on a recent version and apps still look unchanged, the app may be a non-libadwaita GTK 4 app — try the [`gtk4`](../gtk4/README.md) plugin instead.

### Flatpak apps don't pick up the theme

Flatpak GNOME apps don't read host `~/.config/gtk-4.0` by default. Grant access:

```bash
flatpak override --user --filesystem=~/.config/gtk-4.0:ro
```

Some apps further sandbox themselves and ignore even this — there is no general workaround besides running the app natively.

### `Plugin skipped: ~/.config/gtk-4.0 does not exist`

`Hooks.RequiredDirs` requires `~/.config/gtk-4.0` to already exist (taken as a signal that the GTK 4 / libadwaita stack is installed). Create it: `mkdir -p ~/.config/gtk-4.0` and re-run.

## Related plugins

- [gtk4](../gtk4/README.md) — companion plugin for traditional GTK 4 apps; writes a different file in the same `~/.config/gtk-4.0/` directory.
- [gtk3](../gtk3/README.md) — sibling plugin for legacy GTK 3 apps.
- [gnome-shell](../gnome-shell/README.md) — themes the GNOME compositor itself.
