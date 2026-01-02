---
sidebar_position: 7
---

# qt6

Generate colour schemes for [Qt6](https://www.qt.io/) applications via qt6ct.

## Description

Qt6 is the current version of the Qt application framework. On non-KDE desktops, Qt6 applications can be styled using qt6ct.

The plugin generates a colour scheme file compatible with qt6ct.

## Output path

```
~/.config/qt6ct/colors/tinct.conf
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o qt6
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--qt6.output-dir` | `~/.config/qt6ct/colors` | Output directory for Qt6 colour scheme files |

## Requirements

### Install qt6ct

```bash
# Arch/CachyOS
sudo pacman -S qt6ct

# Fedora
sudo dnf install qt6ct

# Ubuntu/Debian
sudo apt install qt6ct
```

### Set environment variable

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, or `/etc/environment`):

```bash
export QT_QPA_PLATFORMTHEME=qt6ct
```

## Configuration

After generation, configure qt6ct to use the tinct colour scheme:

1. Run `qt6ct`
2. Go to **Appearance** tab
3. Set **Color scheme** to **Custom**
4. Select **tinct** from the dropdown
5. Click **Apply**

Changes will apply to Qt6 applications after restart.

## Live reload

On KDE Plasma, the plugin attempts to trigger a live theme reload via D-Bus. Running applications should update their colours immediately.

## Themed elements

The Qt6 colour scheme affects:

- Window backgrounds
- Button colours
- Text colours
- Selection highlights
- Tooltips
- Menu colours
- Scrollbars

## Qt5 vs Qt6

Many applications are migrating from Qt5 to Qt6. If you have a mix of Qt5 and Qt6 applications, use both plugins:

```bash
tinct generate -i image -p ~/wallpaper.jpg -o qt5 -o qt6
```

## See also

- [qt6ct documentation](https://github.com/trialuser02/qt6ct)
- [qt5](/docs/plugins/output/desktop/qt5)
- [kde-plasma](/docs/plugins/output/desktop/kde-plasma)
