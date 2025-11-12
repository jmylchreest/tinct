# Plugin Wishlist

This document tracks requested and recommended output plugins for Tinct. Plugins are categorized by priority and implementation complexity.

## High Priority

### Terminal Emulators

#### WezTerm
- **Format**: TOML
- **Config Location**: `~/.config/wezterm/wezterm.lua` or `~/.config/wezterm/colours/`
- **Popularity**: High - modern GPU-accelerated terminal
- **Complexity**: Low - simple TOML colour mapping
- **Reference**: https://wezfurlong.org/wezterm/config/appearance.html#defining-your-own-colours

#### Alacritty
- **Format**: TOML (modern) or YAML (legacy)
- **Config Location**: `~/.config/alacritty/alacritty.toml`
- **Popularity**: Very High - most popular GPU-accelerated terminal
- **Complexity**: Low - straightforward colour scheme format
- **Reference**: https://alacritty.org/config-alacritty.html

#### Foot
- **Format**: INI-style config
- **Config Location**: `~/.config/foot/foot.ini` or `~/.config/foot/themes/`
- **Popularity**: Medium - lightweight Wayland terminal
- **Complexity**: Low - simple key=value format
- **Reference**: https://codeberg.org/dnkl/foot

### Terminal Multiplexers

#### Tmux
- **Format**: tmux.conf
- **Config Location**: `~/.tmux.conf` or `~/.config/tmux/tmux.conf`
- **Popularity**: Very High - ubiquitous terminal multiplexer
- **Complexity**: Low - status bar and pane styling
- **Reference**: https://github.com/tmux/tmux/wiki

## Medium Priority

### Application Launchers

#### Rofi
- **Format**: rasi (CSS-like)
- **Config Location**: `~/.config/rofi/themes/`
- **Popularity**: High - popular Wayland/X11 launcher
- **Complexity**: Medium - structured theme format with inheritance
- **Reference**: https://github.com/davatorium/rofi
- **Note**: Already have Fuzzel (similar purpose), but Rofi is more feature-rich

### Text Editors

#### Helix
- **Format**: TOML
- **Config Location**: `~/.config/helix/themes/`
- **Popularity**: Medium - modern modal editor
- **Complexity**: Low - simple TOML theme format
- **Reference**: https://docs.helix-editor.com/themes.html

### Window Managers

#### Sway
- **Format**: sway config
- **Config Location**: `~/.config/sway/config`
- **Popularity**: High - i3-compatible Wayland compositor
- **Complexity**: Low - border/bar colours
- **Reference**: https://github.com/swaywm/sway
- **Note**: Already have Hyprland, but Sway is more minimal

## Lower Priority

### Desktop Environments

#### GTK 3/4 Themes
- **Format**: CSS
- **Config Location**: `~/.themes/` or `~/.local/share/themes/`
- **Popularity**: High - used by many applications
- **Complexity**: High - complex CSS with many selectors
- **Reference**: https://docs.gtk.org/gtk4/css-overview.html

#### Qt Themes
- **Format**: Various (qss, conf)
- **Config Location**: Platform-dependent
- **Popularity**: High - used by KDE and Qt apps
- **Complexity**: High - complex theming system
- **Reference**: https://doc.qt.io/qt-6/stylesheet.html

### Development Tools

#### VS Code
- **Format**: JSON
- **Config Location**: `~/.vscode/extensions/` or workspace
- **Popularity**: Very High - most popular editor
- **Complexity**: Medium - comprehensive token scope mapping
- **Reference**: https://code.visualstudio.com/api/extension-guides/colour-theme

#### Zed
- **Format**: JSON
- **Config Location**: `~/.config/zed/themes/`
- **Popularity**: Growing - new Rust-based editor
- **Complexity**: Low - simple JSON theme format
- **Reference**: https://zed.dev/docs/themes

### Window Decorations

#### Mako (Notification Daemon)
- **Format**: INI-style config
- **Config Location**: `~/.config/mako/config`
- **Popularity**: High - popular Wayland notification daemon
- **Complexity**: Low - simple colour settings
- **Reference**: https://github.com/emersion/mako
- **Note**: Already have Dunst, but Mako is Wayland-native

### System UI

#### Plymouth (Boot Splash)
- **Format**: Plymouth theme format
- **Config Location**: `/usr/share/plymouth/themes/` or `/etc/plymouth/`
- **Popularity**: Medium - boot splash screen
- **Complexity**: Medium - requires understanding Plymouth theme structure
- **Reference**: https://www.freedesktop.org/wiki/Software/Plymouth/

#### GRUB Theme
- **Format**: GRUB theme format
- **Config Location**: `/boot/grub/themes/`
- **Popularity**: High - bootloader theming
- **Complexity**: Medium - custom image-based format
- **Reference**: https://www.gnu.org/software/grub/manual/grub/grub.html#Theme-file-format

#### reFINd Theme
- **Format**: reFINd theme format
- **Config Location**: `/boot/efi/EFI/refind/themes/`
- **Popularity**: High - EFI boot theming
- **Complexity**: Medium - custom image-based format
- **Reference**: https://refind.sourceforge.io/

## Specialized

### Browser Extensions

#### Firefox Userchrome
- **Format**: CSS
- **Config Location**: `~/.mozilla/firefox/*/chrome/userChrome.css`
- **Popularity**: Medium - Firefox UI theming
- **Complexity**: Medium - requires CSS knowledge and browser restart
- **Reference**: https://www.userchrome.org/

### Chat/Communication

#### Discord (BetterDiscord/Vencord)
- **Format**: CSS
- **Config Location**: Plugin-dependent
- **Popularity**: Medium - Discord theming
- **Complexity**: Medium - requires third-party client modifications
- **Reference**: https://betterdiscord.app/ or https://vencord.dev/

#### Slack
- **Format**: CSS injection
- **Config Location**: Various methods
- **Popularity**: Low - requires workarounds
- **Complexity**: High - no official theming support

## Community Requested

This section is reserved for plugins requested by the community. Please open an issue on GitHub to suggest new plugins.

---

## Implementation Notes

When implementing new plugins, consider:

1. **Popularity**: How widely used is this application?
2. **Complexity**: How difficult is the theme format to generate?
3. **Maintenance**: How often does the theme format change?
4. **Wayland Support**: Does it work on modern Wayland compositors?
5. **Semantic Mapping**: Can we meaningfully map our 49 colour roles to the application's theme elements?

## Contributing

To request a new plugin, please open an issue with:
- Application name and description
- Config file format and location
- Link to theming documentation
- Example theme file (if available)
