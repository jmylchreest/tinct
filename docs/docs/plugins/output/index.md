---
sidebar_position: 1
---

# Output plugins

Output plugins generate configuration files for applications and devices.

{/* AUTO-GENERATED PLUGIN TABLES — DO NOT EDIT BETWEEN THE MARKERS.
    Regenerated from in-tree READMEs by docs/scripts/sync-plugin-readmes.mjs.
    Plugins are grouped by `plugin.category` frontmatter; section
    headings and ordering live in the script's `renderOutputTables`
    function. To add curated copy beneath a category, do it after the
    closing marker. */}
{/* BEGIN AUTO-PLUGIN-TABLE */}

## Terminals

| Plugin | App | Source |
|--------|-----|--------|
| [alacritty](./terminals/alacritty.md) | Alacritty | builtin |
| [ghostty](./terminals/ghostty.md) | Ghostty | builtin |
| [foot](./terminals/foot.md) | foot | builtin |
| [kitty](./terminals/kitty.md) | kitty | builtin |
| [konsole](./terminals/konsole.md) | Konsole | builtin |
| [warp](./terminals/warp.md) | Warp | builtin |
| [wezterm](./terminals/wezterm.md) | WezTerm | builtin |

## Desktop environments

| Plugin | App | Source |
|--------|-----|--------|
| [gnome-shell](./desktop/gnome-shell.md) | GNOME Shell | builtin |
| [gtk3](./desktop/gtk3.md) | GTK 3 | builtin |
| [gtk4](./desktop/gtk4.md) | GTK 4 | builtin |
| [libadwaita](./desktop/libadwaita.md) | Libadwaita | builtin |
| [kde-plasma](./desktop/kde-plasma.md) | KDE Plasma | builtin |
| [qt5](./desktop/qt5.md) | Qt 5 | builtin |
| [qt6](./desktop/qt6.md) | Qt 6 | builtin |

## Hyprland ecosystem

| Plugin | App | Source |
|--------|-----|--------|
| [hyprland](./hyprland/hyprland.md) | Hyprland | builtin |
| [hyprlock](./hyprland/hyprlock.md) | Hyprlock | builtin |
| [hyprpaper](./hyprland/hyprpaper.md) | Hyprpaper | builtin |

## Bars and launchers

| Plugin | App | Source |
|--------|-----|--------|
| [wofi](./bars-launchers/wofi.md) | Wofi | builtin |
| [waybar](./bars-launchers/waybar.md) | Waybar | builtin |
| [walker](./bars-launchers/walker.md) | Walker | builtin |
| [dunst](./bars-launchers/dunst.md) | Dunst | builtin |
| [fuzzel](./bars-launchers/fuzzel.md) | Fuzzel | builtin |
| [rofi](./bars-launchers/rofi.md) | rofi | builtin |
| [swayosd](./bars-launchers/swayosd.md) | SwayOSD | builtin |

## Editors and multiplexers

| Plugin | App | Source |
|--------|-----|--------|
| [neovim](./editors/neovim.md) | Neovim | builtin |
| [zellij](./editors/zellij.md) | Zellij | builtin |
| [helix](./editors/helix.md) | Helix | builtin |
| [opencode](./editors/opencode.md) | OpenCode | external |

## Special purpose

| Plugin | App | Source |
|--------|-----|--------|
| [yazi](./special/yazi.md) | Yazi | builtin |
| [wbg](./special/wbg.md) | wbg | builtin |
| [tmux](./special/tmux.md) | tmux | builtin |
| [rosec](./special/rosec.md) | ROsec | builtin |
| [mc](./special/mc.md) | Midnight Commander | builtin |
| [markdown](./special/markdown.md) | Tinct Markdown theme export | builtin |
| [histui](./special/histui.md) | histui | builtin |
| [btop](./special/btop.md) | btop | builtin |
| [awww](./special/awww.md) | awww wallpaper daemon | builtin |
| [spicetify](./special/spicetify.md) | Spicetify | external |

{/* END AUTO-PLUGIN-TABLE */}

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
