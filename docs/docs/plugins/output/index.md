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

| Plugin | Description |
|--------|-------------|
| [alacritty](./terminals/alacritty.md) | TOML colour theme for Alacritty, the GPU-accelerated cross-platform terminal emulator. |
| [ghostty](./terminals/ghostty.md) | Single-file colour theme for Ghostty, Mitchell Hashimoto's GPU-accelerated terminal emulator. |
| [foot](./terminals/foot.md) | Colour theme for foot, the fast, lightweight Wayland terminal emulator from the author of fuzzel. |
| [kitty](./terminals/kitty.md) | Single-file colour theme for kitty, Kovid Goyal's GPU-accelerated terminal emulator. |
| [konsole](./terminals/konsole.md) | .colorscheme file for Konsole, the KDE terminal emulator. |
| [warp](./terminals/warp.md) | YAML custom-theme file for Warp, the Rust-based AI-first terminal. |
| [wezterm](./terminals/wezterm.md) | TOML colour scheme for WezTerm, Wez Furlong's GPU-accelerated, Lua-configurable terminal. |
| [ptyxis](./terminals/ptyxis.md) | Ptyxis terminal colour palette. |

## Desktop Environments

| Plugin | Description |
|--------|-------------|
| [niri](./desktop/niri.md) | Themed colour configuration for niri, the scrollable-tiling Wayland compositor. |
| [gnome-shell](./desktop/gnome-shell.md) | CSS theme for GNOME Shell, the compositor and panel of the GNOME desktop. |
| [gtk3](./desktop/gtk3.md) | Per-user CSS override for GTK 3 applications (GIMP, Inkscape, Pidgin, and most non-libadwaita GNOME 3-era apps). |
| [gtk4](./desktop/gtk4.md) | Per-user CSS override for traditional GTK 4 applications — i.e. GTK 4 apps that do not use libadwaita. |
| [libadwaita](./desktop/libadwaita.md) | Per-user CSS override for libadwaita — the design library used by modern GNOME apps (Files, Text Editor, Calculator, Console,… |
| [kde-plasma](./desktop/kde-plasma.md) | KDE Plasma .colors colour schemes for the Plasma desktop and KDE applications. |
| [qt5](./desktop/qt5.md) | Qt5ct custom colour scheme for Qt 5 applications on non-Plasma desktops. qt5ct is Qt 5's "configuration tool" — it provides a… |
| [qt6](./desktop/qt6.md) | Qt6ct custom colour scheme for Qt 6 applications on non-Plasma desktops. qt6ct is the Qt 6 successor to qt5ct — a QPlatformTh… |

## Hyprland Ecosystem

| Plugin | Description |
|--------|-------------|
| [hyprland](./hyprland/hyprland.md) | Colour variables for Hyprland, the dynamic tiling Wayland compositor. |
| [hyprlock](./hyprland/hyprlock.md) | Colour palette for Hyprlock, the screen locker for Hyprland. |
| [hyprpaper](./hyprland/hyprpaper.md) | Hyprpaper wallpaper configuration that wires the wallpaper image picked up by tinct into Hyprpaper's preload / wallpaper dire… |

## Bars & Launchers

| Plugin | Description |
|--------|-------------|
| [wofi](./bars-launchers/wofi.md) | Colour theme for Wofi, the GTK-based application launcher / dmenu replacement for Wayland. |
| [waybar](./bars-launchers/waybar.md) | Colour theme for Waybar, the highly customisable status bar for Wayland compositors. |
| [walker](./bars-launchers/walker.md) | GTK CSS theme for Walker, a Wayland-native application launcher. |
| [dunst](./bars-launchers/dunst.md) | Urgency-based colour theme for Dunst, the lightweight notification daemon for X11 and Wayland. |
| [fuzzel](./bars-launchers/fuzzel.md) | Colour-only INI theme for Fuzzel, the lightweight Wayland application launcher. |
| [rofi](./bars-launchers/rofi.md) | Rofi theme as two .rasi files: a palette (tinct-colours.rasi) declaring every tinct role as a named rasi colour, and a main t… |
| [swayosd](./bars-launchers/swayosd.md) | GTK CSS stylesheet for SwayOSD, the on-screen-display server for Sway and other wlroots compositors. |
| [keylightd-tray](./bars-launchers/keylightd-tray.md) | CSS theme for keylightd-tray, the Elgato Key Light system-tray applet that ships with keylightd. |
| [awob](./bars-launchers/awob.md) | Shared palette plus a worked-example theme for awob, an animated Wayland Overlay Bar. |
| [dunstify](./bars-launchers/dunstify.md) | Pops a desktop notification via dunstify (or notify-send) every time tinct generate finishes. |
| [wob](./bars-launchers/wob.md) | Tinct.ini theme overlay for wob, the Wayland Overlay Bar — the bar you see when adjusting volume or brightness. |
| [noctalia](./bars-launchers/noctalia.md) | Custom colour palette for the Noctalia shell, a Wayland desktop shell (bar, dock, launcher, control center, notifications) fo… |

## Editors & Multiplexers

| Plugin | Description |
|--------|-------------|
| [neovim](./editors/neovim.md) | Lua colorscheme for Neovim, plus a matching lualine status-line theme. |
| [zellij](./editors/zellij.md) | KDL-format theme file for Zellij, the Rust terminal workspace. |
| [helix](./editors/helix.md) | TOML theme for Helix, the modal post-modern text editor written in Rust. |
| [opencode](./editors/opencode.md) | Complete theme file for OpenCode, the AI-assisted coding agent. |
| [zed](./editors/zed.md) | Complete theme file for the Zed editor — a fast Rust-based collaborative code editor. |

## Special Purpose

| Plugin | Description |
|--------|-------------|
| [yazi](./special/yazi.md) | Yazi flavor pack — Yazi's directory-based theme format. |
| [wbg](./special/wbg.md) | Wallpaper integration for wbg, the minimal Wayland wallpaper application from the dnkl/foot ecosystem. wbg is a single-shot foreground pr… |
| [tmux](./special/tmux.md) | Colour theme for tmux, the terminal multiplexer. |
| [rosec](./special/rosec.md) | ROsec prompt theme — the colours used by the rosec-prompt interactive prompt for backgrounds, borders, labels, inputs, and th… |
| [mc](./special/mc.md) | Skin for Midnight Commander, the venerable terminal file manager. |
| [markdown](./special/markdown.md) | Exports the current palette as a portable Markdown theme file with YAML frontmatter. |
| [histui](./special/histui.md) | GTK4 / libadwaita stylesheet for histui, a notification daemon. |
| [btop](./special/btop.md) | Btop theme file — the colours behind the resource monitor's CPU/memory/network/process boxes, gradient meters, temperature sc… |
| [awww](./special/awww.md) | Wallpaper integration for awww (An Answer to your Wayland Wallpaper Woes), an animated wallpaper daemon for Wayland compositors implement… |
| [spicetify](./special/spicetify.md) | Spicetify color.ini colour scheme so the Spotify desktop client picks up your tinct palette. |
| [templater](./special/templater.md) | A meta-plugin: rather than targeting one app, templater renders an arbitrary set of user-supplied Go text/template files against the curr… |

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

Most plugins reload the target application themselves after writing — kitty is
signalled with `SIGUSR1`, waybar with `SIGUSR2`, and so on. A few make it
opt-in via a flag:

```bash
# Hyprland with reload
tinct generate -i image -p ~/wallpaper.jpg -o hyprland --hyprland.reload
```

Each plugin's page documents its reload behaviour.

### File tracking

Generated files are tracked in `~/.local/share/tinct/manifest.json`.

## See also

- [Plugin overview](../overview.md) - Architecture
- [Plugin development](../../plugin-development/index.md) - Creating custom plugins
