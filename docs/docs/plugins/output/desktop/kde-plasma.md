---
sidebar_position: 2
---

# kde-plasma

Generate colour schemes for [KDE Plasma](https://kde.org/plasma-desktop/) desktop environment.

## Description

KDE Plasma is the default desktop environment for many Linux distributions. The plugin generates `.colors` files that style the entire Plasma desktop including panels, windows, and KDE applications.

## Output path

```
~/.local/share/color-schemes/TinctDark1.colors
~/.local/share/color-schemes/TinctDark2.colors
~/.local/share/color-schemes/TinctLight1.colors
~/.local/share/color-schemes/TinctLight2.colors
```

Four files are generated (two variants each for dark and light themes) to work around `plasma-apply-colorscheme` not reloading when the scheme name is unchanged.

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o kde-plasma
```

### Dual-theme generation

When using dual theme mode, both light and dark variants are generated:

```bash
tinct generate -i image -p ~/wallpaper.jpg -o kde-plasma --dual-theme
```

The plugin will detect your system colour preference and apply the matching theme.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--kde-plasma.output-dir` | `~/.local/share/color-schemes` | Output directory for colour scheme files |

## Automatic application

The plugin automatically:

1. Detects system colour preference (light/dark)
2. Selects the appropriate theme variant (1 or 2) to force reload
3. Applies the scheme using `plasma-apply-colorscheme`
4. Triggers D-Bus reload for KWin and Plasma Shell
5. Sets wallpaper using `plasma-apply-wallpaperimage` (if source is an image file)

## Requirements

For automatic application:
- `plasma-apply-colorscheme` (part of plasma-workspace)
- `plasma-apply-wallpaperimage` (optional, for wallpaper setting)
- `kreadconfig5` (for detecting current scheme)

## Themed elements

The KDE Plasma colour scheme affects:

- Plasma panels and widgets
- KDE applications (Dolphin, Kate, Konsole, etc.)
- Window decorations
- System dialogs
- Accent colours throughout the desktop

## Manual application

If auto-application fails:

```bash
plasma-apply-colorscheme TinctDark1
# or
plasma-apply-colorscheme TinctLight1
```

## See also

- [KDE Plasma documentation](https://docs.kde.org/stable5/en/)
- [konsole](../terminals/konsole.md)
- [qt5](./qt5.md)
- [qt6](./qt6.md)
