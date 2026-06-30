---
title: qt5
sidebar_position: 8
plugin:
  type: output
  name: qt5
  category: desktop
  source: builtin
  app: Qt 5
  app_url: 'https://doc.qt.io/qt-5/'
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/qt5ct/colors
  generated_files:
    - tinct.conf
  reload:
    method: ipc
    command: KDE D-Bus theme reload (KGlobalSettings)
    user_action_required: true
---

# qt5

Generates a [qt5ct](https://github.com/desktop-app/qt5ct) custom colour scheme for [Qt 5](https://doc.qt.io/qt-5/) applications on non-Plasma desktops. qt5ct is Qt 5's "configuration tool" — it provides a `QPlatformTheme` plugin that injects a user-defined `QPalette` into every Qt 5 app at startup. The file written here is one of the colour schemes qt5ct can be pointed at; selecting it in `qt5ct.conf` (or via the qt5ct GUI) makes Qt 5 apps adopt tinct's palette.

For KDE Plasma users, prefer the [`kde-plasma`](./kde-plasma.md) plugin — Plasma's colour scheme system is richer and applies system-wide.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o qt5` works out of the box.

You also need `qt5ct` installed and `QT_QPA_PLATFORMTHEME=qt5ct` set in your environment (typically in `~/.config/environment.d/*.conf` or `~/.profile`). The plugin's `PreExecute` aborts if `~/.config/qt5ct/` doesn't exist; create the directory and run `qt5ct` at least once to initialise.

```bash
# Arch / CachyOS
sudo pacman -S qt5ct

# Debian / Ubuntu
sudo apt install qt5ct
```

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o qt5
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.conf` | `~/.config/qt5ct/colors/tinct.conf` | Custom qt5ct colour scheme; 22-slot `[ColorScheme]` ini for active / inactive / disabled `QPalette` states |

## Integration

qt5ct discovers user colour schemes from `~/.config/qt5ct/colors/*.conf` and lists them under **qt5ct → Appearance → Color scheme → Custom**. The plugin writes into that directory, so the scheme called "tinct" appears in the dropdown automatically.

To select it:

1. Ensure `QT_QPA_PLATFORMTHEME=qt5ct` is exported (check with `env | grep QT_QPA`).
2. Run `qt5ct` and pick **Color scheme → Custom → tinct**.

Or, edit `~/.config/qt5ct/qt5ct.conf` directly:

```ini
[Appearance]
custom_palette=true
color_scheme_path=/home/<user>/.config/qt5ct/colors/tinct.conf
```

The plugin's `PostExecute` checks `~/.config/qt5ct/qt5ct.conf` for a `color_scheme_path=` line containing `tinct.conf`; if absent, it prints an actionable banner.

## Reload behaviour

### Automatic

After writing the file, the plugin calls the shared KDE D-Bus theme-reload helper (`kdedbus.ReloadTheme`). On KDE Plasma this triggers a global Qt repalette; on non-KDE desktops the D-Bus call is a no-op (the relevant services don't exist) and the plugin logs a verbose-mode note.

> qt5ct itself has no IPC reload. Once `qt5ct.conf` references the scheme, every newly-launched Qt 5 app picks it up — but already-running Qt 5 apps continue to use the palette they read at startup.

### Manual fallback

Close and reopen the Qt 5 applications you want to retheme. There is no qt5ct-equivalent of `dunstctl reload`.

## Uninstall / revert

1. **Stop pointing qt5ct at the tinct scheme** — edit `~/.config/qt5ct/qt5ct.conf`:

   ```ini
   [Appearance]
   custom_palette=false
   # color_scheme_path=...   # comment out or delete
   ```

   Or in the qt5ct GUI: **Appearance → Color scheme → `<pick a built-in>`**.

2. **Delete the generated file**:

   ```bash
   rm ~/.config/qt5ct/colors/tinct.conf
   ```

3. **Reload** — close and reopen Qt 5 apps (or restart your session for the global change to take effect).

4. **External state**: this plugin only writes to `~/.config/qt5ct/colors/`. It does **not** modify `qt5ct.conf` itself — selecting the scheme there was your manual step in [Integration](#integration) and you have to undo it in step 1.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--qt5.output-dir` | `~/.config/qt5ct/colors` | Override the output directory |

## Colour role mapping

qt5ct stores each `QPalette` state (active / inactive / disabled) as a 22-comma-separated `#AARRGGBB` list. The slot order is fixed by Qt:

| Slot | Qt role | Tinct role (active) | Tinct role (inactive / disabled) |
|---|---|---|---|
| 0 | `WindowText` | `foreground` | `foreground` @ 70% / 50% alpha |
| 1 | `Button` | `surface` | `surface` |
| 2 | `Light` | `surfaceVariant` | `surfaceVariant` |
| 3 | `Midlight` | `border` (from `outline`/`border`/`foregroundMuted`) | same |
| 4 | `Dark` | `background` | `background` |
| 5 | `Mid` | `background` | `background` |
| 6 | `Text` | `foreground` | `foreground` @ 70% / 50% alpha |
| 7 | `BrightText` | `foreground` | `foreground` |
| 8 | `ButtonText` | `foreground` | `foreground` @ 70% / 50% alpha |
| 9 | `Base` | `background` | `background` |
| 10 | `Window` | `surface` | `surface` |
| 11 | `AlternateBase` | `border` | `border` |
| 12 | `ToolTipBase` | `accent1` | `accent1` (disabled = `surface`) |
| 13 | `ToolTipText` | `background` | `foreground` @ 70% / 50% alpha |
| 14 | `Text` (re-use) | `accent1` | `accent1` @ 70% / 50% alpha |
| 15 | `Highlight` (interim) | `foreground` | `foreground` @ 70% |
| 16 | `Button` (re-use) | `surface` | `surface` |
| 17 | `BrightText` (re-use) | `#ffffffff` literal | `#ffffffff` literal |
| 18 | `Link` (interim) | `background` | `background` |
| 19 | `LinkVisited` (interim) | `foreground` | `foreground` |
| 20 | `Highlight` (translucent) | `accent1` @ 50% alpha | `accent1` @ 70% / 30% alpha |
| 21 | `HighlightedText` | `accent1` | `accent1` (disabled = `surface`) |

(Some Qt slot indices appear twice in qt5ct's 22-slot format; that's how qt5ct serialises the palette.)

## Customising the template

```bash
tinct plugins templates dump -o qt5 -l ~/.config/tinct/templates/qt5
```

This creates `~/.config/tinct/templates/qt5/tinct.conf.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### `Plugin skipped: qt5ct config directory does not exist`

`PreExecute` couldn't find `~/.config/qt5ct/`. Install `qt5ct` and run it once (or `mkdir -p ~/.config/qt5ct/colors`) to create the config directory, then retry.

### Plugin printed `Qt5 Theme Configuration Required` banner

`PostExecute` checked `~/.config/qt5ct/qt5ct.conf` and didn't find a `color_scheme_path=` line referencing `tinct.conf`. Follow the steps in [Integration](#integration): set `QT_QPA_PLATFORMTHEME=qt5ct` in your environment and select the tinct scheme in the qt5ct GUI (or set `custom_palette=true` and `color_scheme_path=...` in `qt5ct.conf` directly).

### Qt 5 apps still look like Adwaita / Breeze

Either `QT_QPA_PLATFORMTHEME` isn't set (qt5ct only kicks in when Qt is told to use it), or your Qt apps are running under Flatpak with their own bundled Qt stack. Check `env QT_DEBUG_PLUGINS=1 <app>` to see what platform-theme plugin Qt actually loaded.

### Already-open apps didn't update

qt5ct reads the palette at app start. Close and reopen the apps; there is no Qt-side IPC to repalette running apps from a config file change (the D-Bus reload only works on KDE Plasma).

### I'm on KDE Plasma — does this plugin do anything useful?

Probably not. Plasma supplies its own `QPlatformTheme` (`plasma`) that overrides `QT_QPA_PLATFORMTHEME=qt5ct`. Use the [`kde-plasma`](./kde-plasma.md) plugin instead, which writes a proper KDE colour scheme.

## Related plugins

- [qt6](./qt6.md) — sibling plugin for Qt 6 apps (writes to `~/.config/qt6ct/colors/`).
- [kde-plasma](./kde-plasma.md) — preferred for KDE Plasma users; applies to both Qt 5 and Qt 6 apps system-wide.
- [gtk3](./gtk3.md), [gtk4](./gtk4.md) — pair with these so Qt and GTK apps look consistent.
