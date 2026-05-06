---
sidebar_position: 1
---

# Output plugins

Output plugins generate configuration files for applications and devices.

## Terminals

Terminal emulators with colour palette support:

| Plugin | Description | Config format |
|--------|-------------|---------------|
| [alacritty](./terminals/alacritty.md) | Alacritty terminal | TOML |
| [ghostty](./terminals/ghostty.md) | Ghostty terminal | Config file |
| [kitty](./terminals/kitty.md) | Kitty terminal | Config file |
| [konsole](./terminals/konsole.md) | KDE Konsole | .colorscheme |
| [ptyxis](./terminals/ptyxis.md) | GNOME Ptyxis | .palette |
| [warp](./terminals/warp.md) | Warp terminal | YAML |

## Desktop environments

Desktop theming for GNOME, KDE, and toolkits:

| Plugin | Description | Requirements |
|--------|-------------|--------------|
| [gnome-shell](./desktop/gnome-shell.md) | GNOME Shell theme | User Themes extension |
| [kde-plasma](./desktop/kde-plasma.md) | KDE Plasma colours | Plasma desktop |
| [gtk3](./desktop/gtk3.md) | GTK3 applications | adw-gtk3-theme |
| [gtk4](./desktop/gtk4.md) | GTK4 applications | - |
| [libadwaita](./desktop/libadwaita.md) | Libadwaita apps | - |
| [qt5](./desktop/qt5.md) | Qt5 applications | qt5ct |
| [qt6](./desktop/qt6.md) | Qt6 applications | qt6ct |

## Hyprland ecosystem

Window manager and related tools:

| Plugin | Description | Features |
|--------|-------------|----------|
| [hyprland](/docs/plugins/output/hyprland) | Hyprland compositor | Colour variables |
| [hyprlock](./hyprland/hyprlock.md) | Screen locker | Wallpaper support |
| [hyprpaper](./hyprland/hyprpaper.md) | Wallpaper manager | Auto-apply |

## Bars and launchers

Status bars, notification daemons, and application launchers:

| Plugin | Description | Config format |
|--------|-------------|---------------|
| [waybar](./bars-launchers/waybar.md) | Wayland bar | CSS |
| [dunst](./bars-launchers/dunst.md) | Notifications | dunstrc |
| [swayosd](./bars-launchers/swayosd.md) | On-screen display | SCSS |
| [fuzzel](./bars-launchers/fuzzel.md) | App launcher | INI |
| [walker](./bars-launchers/walker.md) | App launcher | JSON |
| [wofi](./bars-launchers/wofi.md) | App launcher | CSS |

## Editors and multiplexers

Text editors and terminal multiplexers:

| Plugin | Description | Config format |
|--------|-------------|---------------|
| [neovim](./editors/neovim.md) | Neovim editor | Lua |
| [zellij](./editors/zellij.md) | Terminal multiplexer | KDL |

## Special purpose

Export and utility plugins:

| Plugin | Description | Purpose |
|--------|-------------|---------|
| [markdown](./special/markdown.md) | Theme export | Save portable themes |
| [template](./special/template.md) | Custom templates | User-defined output |
| [histui](./special/histui.md) | Notification daemon | histui theming |

## Using output plugins

### All plugins

```bash
tinct generate -i image -p ~/wallpaper.jpg -o all
```

### Specific plugins

```bash
tinct generate -i image -p ~/wallpaper.jpg -o kitty,hyprland,waybar
```

### By category (manual)

```bash
# Just terminals
-o alacritty,ghostty,kitty

# Just Hyprland
-o hyprland,hyprpaper,hyprlock

# Just GNOME
-o gnome-shell,gtk4,libadwaita
```

## Common features

### Output paths

Each plugin writes to application-specific locations. Most use XDG directories:

- `~/.config/<app>/`
- `~/.local/share/<app>/`

### Post-generation hooks

Many plugins support automatic reload:

```bash
# Kitty with reload
tinct generate -i image -p ~/wallpaper.jpg -o kitty --kitty.reload
```

### File tracking

Generated files are tracked in `~/.config/tinct/.tinct-manifest.json`.

## See also

- [Plugin overview](../overview.md) - Architecture
- [Plugin development](../../plugin-development/index.md) - Creating custom plugins
