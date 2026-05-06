---
sidebar_position: 6
---

# qt5

Generate colour schemes for [Qt5](https://www.qt.io/) applications via qt5ct.

## Description

Qt5 is a cross-platform application framework used by many applications including VLC, qBittorrent, and various KDE applications. On non-KDE desktops, Qt5 applications can be styled using qt5ct.

The plugin generates a colour scheme file compatible with qt5ct.

## Output path

```
~/.config/qt5ct/colors/tinct.conf
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o qt5
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--qt5.output-dir` | `~/.config/qt5ct/colors` | Output directory for Qt5 colour scheme files |

## Requirements

### Install qt5ct

```bash
# Arch/CachyOS
sudo pacman -S qt5ct

# Fedora
sudo dnf install qt5ct

# Ubuntu/Debian
sudo apt install qt5ct
```

### Set environment variable

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, or `/etc/environment`):

```bash
export QT_QPA_PLATFORMTHEME=qt5ct
```

## Configuration

After generation, configure qt5ct to use the tinct colour scheme:

1. Run `qt5ct`
2. Go to **Appearance** tab
3. Set **Color scheme** to **Custom**
4. Select **tinct** from the dropdown
5. Click **Apply**

Changes will apply to Qt5 applications after restart.

## Live reload

On KDE Plasma, the plugin attempts to trigger a live theme reload via D-Bus. Running applications should update their colours immediately.

## Themed elements

The Qt5 colour scheme affects:

- Window backgrounds
- Button colours
- Text colours
- Selection highlights
- Tooltips
- Menu colours
- Scrollbars

## See also

- [qt5ct documentation](https://github.com/desktop-app/qt5ct)
- [qt6](./qt6.md)
- [kde-plasma](./kde-plasma.md)
