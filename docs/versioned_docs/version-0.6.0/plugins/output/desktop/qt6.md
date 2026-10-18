---
title: qt6
sidebar_position: 9
plugin:
  type: output
  name: qt6
  category: desktop
  source: builtin
  app: Qt 6
  app_url: 'https://doc.qt.io/qt-6/'
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/qt6ct/colors
  generated_files:
    - tinct.conf
  reload:
    method: ipc
    command: KDE D-Bus theme reload (KGlobalSettings)
    user_action_required: true
---

# qt6

Generates a [qt6ct](https://github.com/trialuser02/qt6ct) custom colour scheme for [Qt 6](https://doc.qt.io/qt-6/) applications on non-Plasma desktops. qt6ct is the Qt 6 successor to qt5ct — a `QPlatformTheme` plugin that injects a user-defined `QPalette` into every Qt 6 app at startup. The file written here is one of the colour schemes qt6ct can be pointed at; selecting it in `qt6ct.conf` (or via the qt6ct GUI) makes Qt 6 apps adopt tinct's palette.

For KDE Plasma users, prefer the [`kde-plasma`](./kde-plasma.md) plugin — Plasma's colour scheme system is richer and applies system-wide.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o qt6` works out of the box.

You also need `qt6ct` installed and `QT_QPA_PLATFORMTHEME=qt6ct` set in your environment (typically in `~/.config/environment.d/*.conf` or `~/.profile`). The plugin's `PreExecute` aborts if `~/.config/qt6ct/` doesn't exist; create the directory and run `qt6ct` at least once to initialise.

```bash
# Arch / CachyOS
sudo pacman -S qt6ct

# Debian / Ubuntu
sudo apt install qt6ct
```

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o qt6
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.conf` | `~/.config/qt6ct/colors/tinct.conf` | Custom qt6ct colour scheme; 22-slot `[ColorScheme]` ini for active / inactive / disabled `QPalette` states |

## Integration

qt6ct discovers user colour schemes from `~/.config/qt6ct/colors/*.conf` and lists them under **qt6ct → Appearance → Color scheme → Custom**. The plugin writes into that directory, so the scheme called "tinct" appears in the dropdown automatically.

To select it:

1. Ensure `QT_QPA_PLATFORMTHEME=qt6ct` is exported (check with `env | grep QT_QPA`).
2. Run `qt6ct` and pick **Color scheme → Custom → tinct**.

Or, edit `~/.config/qt6ct/qt6ct.conf` directly:

```ini
[Appearance]
custom_palette=true
color_scheme_path=/home/<user>/.config/qt6ct/colors/tinct.conf
```

The plugin's `PostExecute` checks `~/.config/qt6ct/qt6ct.conf` for a `color_scheme_path=` line containing `tinct.conf`; if absent, it prints an actionable banner.

## Reload behaviour

### Automatic

After writing the file, the plugin calls the shared KDE D-Bus theme-reload helper (`kdedbus.ReloadTheme`). On KDE Plasma this triggers a global Qt repalette; on non-KDE desktops the D-Bus call is a no-op (the relevant services don't exist) and the plugin logs a verbose-mode note.

> qt6ct itself has no IPC reload. Once `qt6ct.conf` references the scheme, every newly-launched Qt 6 app picks it up — but already-running Qt 6 apps continue to use the palette they read at startup.

### Manual fallback

Close and reopen the Qt 6 applications you want to retheme. There is no qt6ct-equivalent of `dunstctl reload`.

## Uninstall / revert

1. **Stop pointing qt6ct at the tinct scheme** — edit `~/.config/qt6ct/qt6ct.conf`:

   ```ini
   [Appearance]
   custom_palette=false
   # color_scheme_path=...   # comment out or delete
   ```

   Or in the qt6ct GUI: **Appearance → Color scheme → `<pick a built-in>`**.

2. **Delete the generated file**:

   ```bash
   rm ~/.config/qt6ct/colors/tinct.conf
   ```

3. **Reload** — close and reopen Qt 6 apps (or restart your session for the global change to take effect).

4. **External state**: this plugin only writes to `~/.config/qt6ct/colors/`. It does **not** modify `qt6ct.conf` itself — selecting the scheme there was your manual step in [Integration](#integration) and you have to undo it in step 1.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--qt6.output-dir` | `~/.config/qt6ct/colors` | Override the output directory |

## Colour role mapping

qt6ct stores each `QPalette` state (active / inactive / disabled) as a 22-comma-separated `#AARRGGBB` list. The slot order is fixed by Qt:

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

(Some Qt slot indices appear twice in qt6ct's 22-slot format; that's how qt6ct serialises the palette.)

## Customising the template

```bash
tinct plugins templates dump -o qt6 -l ~/.config/tinct/templates/qt6
```

This creates `~/.config/tinct/templates/qt6/tinct.conf.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### `Plugin skipped: qt6ct config directory does not exist`

`PreExecute` couldn't find `~/.config/qt6ct/`. Install `qt6ct` and run it once (or `mkdir -p ~/.config/qt6ct/colors`) to create the config directory, then retry.

### Plugin printed `Qt6 Theme Configuration Required` banner

`PostExecute` checked `~/.config/qt6ct/qt6ct.conf` and didn't find a `color_scheme_path=` line referencing `tinct.conf`. Follow the steps in [Integration](#integration): set `QT_QPA_PLATFORMTHEME=qt6ct` in your environment and select the tinct scheme in the qt6ct GUI (or set `custom_palette=true` and `color_scheme_path=...` in `qt6ct.conf` directly).

### Qt 6 apps still look like Fusion / Breeze

Either `QT_QPA_PLATFORMTHEME` isn't set (qt6ct only kicks in when Qt is told to use it), or your Qt apps are running under Flatpak with their own bundled Qt stack. Check `env QT_DEBUG_PLUGINS=1 <app>` to see what platform-theme plugin Qt actually loaded.

### Already-open apps didn't update

qt6ct reads the palette at app start. Close and reopen the apps; there is no Qt-side IPC to repalette running apps from a config file change (the D-Bus reload only works on KDE Plasma).

### I'm on KDE Plasma — does this plugin do anything useful?

Probably not. Plasma supplies its own `QPlatformTheme` (`plasma`) that overrides `QT_QPA_PLATFORMTHEME=qt6ct`. Use the [`kde-plasma`](./kde-plasma.md) plugin instead, which writes a proper KDE colour scheme that applies to both Qt 5 and Qt 6 apps.

## Related plugins

- [qt5](./qt5.md) — sibling plugin for legacy Qt 5 apps (writes to `~/.config/qt5ct/colors/`).
- [kde-plasma](./kde-plasma.md) — preferred for KDE Plasma users; applies to both Qt 5 and Qt 6 apps system-wide.
- [gtk3](./gtk3.md), [gtk4](./gtk4.md) — pair with these so Qt and GTK apps look consistent.
