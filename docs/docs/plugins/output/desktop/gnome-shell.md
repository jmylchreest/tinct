---
sidebar_position: 1
---

# gnome-shell

Generate themes for [GNOME Shell](https://wiki.gnome.org/Projects/GnomeShell) panel, overview, and system menus.

## Description

GNOME Shell is the core user interface of the GNOME desktop environment. The plugin generates CSS files that style the panel, overview, app menu, and system dialogs.

## Requirements

- GNOME Shell installed
- User Themes extension enabled
- `gsettings` and `gnome-extensions` commands available

### Installing User Themes extension

```bash
# Fedora/RHEL
sudo dnf install gnome-shell-extension-user-theme

# Ubuntu/Debian
sudo apt install gnome-shell-extension-user-theme

# Arch
sudo pacman -S gnome-shell-extension-user-theme
```

After installation, enable the extension:

```bash
gnome-extensions enable user-theme@gnome-shell-extensions.gcampax.github.com
```

Then log out and back in for the extension to activate.

## Output path

```
~/.local/share/themes/tinct-a/gnome-shell/gnome-shell.css
~/.local/share/themes/tinct-b/gnome-shell/gnome-shell.css
```

Two theme variants (tinct-a and tinct-b) are generated to enable flicker-free theme reloading.

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o gnome-shell
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--gnome-shell.output-dir` | `~/.local/share/themes` | Output directory for theme files |

## Automatic application

The plugin automatically applies the theme using two strategies:

1. **D-Bus reload** (preferred): Calls `Main.loadTheme()` via D-Bus for instant theme reload
2. **gsettings toggle**: Falls back to toggling between tinct-a and tinct-b themes

If the source image is a wallpaper file, it will also be set as the desktop background.

## Themed elements

The GNOME Shell theme styles:

- Top panel (background, text, icons)
- Overview mode
- App menu and system menus
- Notification popups
- Lock screen elements
- Workspace switcher

## See also

- [GNOME Shell extensions](https://extensions.gnome.org/)
- [libadwaita](/docs/plugins/output/desktop/libadwaita)
- [gtk4](/docs/plugins/output/desktop/gtk4)
