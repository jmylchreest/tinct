---
sidebar_position: 1
---

# Output plugins

Output plugins generate configuration files for applications and devices.

## Terminals

Terminal emulators with colour palette support:

| Plugin | Description | Config format |
|--------|-------------|---------------|
| [alacritty](/docs/plugins/output/terminals/alacritty) | Alacritty terminal | TOML |
| [ghostty](/docs/plugins/output/terminals/ghostty) | Ghostty terminal | Config file |
| [kitty](/docs/plugins/output/terminals/kitty) | Kitty terminal | Config file |
| [konsole](/docs/plugins/output/terminals/konsole) | KDE Konsole | .colorscheme |
| [ptyxis](/docs/plugins/output/terminals/ptyxis) | GNOME Ptyxis | .palette |

## Desktop environments

Desktop theming for GNOME, KDE, and toolkits:

| Plugin | Description | Requirements |
|--------|-------------|--------------|
| [gnome-shell](/docs/plugins/output/desktop/gnome-shell) | GNOME Shell theme | User Themes extension |
| [kde-plasma](/docs/plugins/output/desktop/kde-plasma) | KDE Plasma colours | Plasma desktop |
| [gtk3](/docs/plugins/output/desktop/gtk3) | GTK3 applications | adw-gtk3-theme |
| [gtk4](/docs/plugins/output/desktop/gtk4) | GTK4 applications | - |
| [libadwaita](/docs/plugins/output/desktop/libadwaita) | Libadwaita apps | - |
| [qt5](/docs/plugins/output/desktop/qt5) | Qt5 applications | qt5ct |
| [qt6](/docs/plugins/output/desktop/qt6) | Qt6 applications | qt6ct |

## Hyprland ecosystem

Window manager and related tools:

| Plugin | Description | Features |
|--------|-------------|----------|
| [hyprland](/docs/plugins/output/hyprland) | Hyprland compositor | Colour variables |
| [hyprlock](/docs/plugins/output/hyprland/hyprlock) | Screen locker | Wallpaper support |
| [hyprpaper](/docs/plugins/output/hyprland/hyprpaper) | Wallpaper manager | Auto-apply |

## Bars and launchers

Status bars, notification daemons, and application launchers:

| Plugin | Description | Config format |
|--------|-------------|---------------|
| [waybar](/docs/plugins/output/bars-launchers/waybar) | Wayland bar | CSS |
| [dunst](/docs/plugins/output/bars-launchers/dunst) | Notifications | dunstrc |
| [swayosd](/docs/plugins/output/bars-launchers/swayosd) | On-screen display | SCSS |
| [fuzzel](/docs/plugins/output/bars-launchers/fuzzel) | App launcher | INI |
| [walker](/docs/plugins/output/bars-launchers/walker) | App launcher | JSON |
| [wofi](/docs/plugins/output/bars-launchers/wofi) | App launcher | CSS |

## Editors and multiplexers

Text editors and terminal multiplexers:

| Plugin | Description | Config format |
|--------|-------------|---------------|
| [neovim](/docs/plugins/output/editors/neovim) | Neovim editor | Lua |
| [zellij](/docs/plugins/output/editors/zellij) | Terminal multiplexer | KDL |

## Special purpose

Export and utility plugins:

| Plugin | Description | Purpose |
|--------|-------------|---------|
| [markdown](/docs/plugins/output/special/markdown) | Theme export | Save portable themes |
| [template](/docs/plugins/output/special/template) | Custom templates | User-defined output |
| [histui](/docs/plugins/output/special/histui) | Notification daemon | histui theming |

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

- [Plugin overview](/docs/plugins/overview) - Architecture
- [External plugins](/docs/plugins/external) - Custom plugins
