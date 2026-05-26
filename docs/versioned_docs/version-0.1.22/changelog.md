---
sidebar_position: 100
sidebar_label: Changelog
toc_max_heading_level: 2
---

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## [0.1.22]
*2026-01-02*

### Added
- Docusaurus documentation site at [jmylchreest.github.io/tinct](https://jmylchreest.github.io/tinct/)
- Versioned template system for breaking config changes
- Plugin development documentation with protocols, hooks, and publishing guides

### Changed
- Simplified README with links to documentation site

### Fixed
- hyprpaper v0.8.0 support with versioned templates

## [0.1.21]
*2025-12-31*

### Added
- histui output plugin for notification daemon theming

### Fixed
- Security annotations for config file reads (G304)

## [0.1.20]
*2025-12-31*

### Changed
- Tightened file and directory permissions for security

## [0.1.18]
*2025-12-24*

### Added
- ptyxis terminal output plugin

## [0.1.17]
*2025-12-14*

### Added
- konsole terminal output plugin

## [0.1.16]
*2025-12-14*

### Added
- zed editor output plugin (contrib)
- templater output plugin (contrib)
- keylightd-tray output plugin (contrib)
- wob bar output plugin (contrib)

## [0.1.15]
*2025-12-14*

### Added
- External plugin repository system
- Plugin installation from git repositories

## [0.1.14]
*2025-12-14*

### Added
- JSON-stdio protocol for external plugins
- go-plugin RPC protocol support

## [0.1.11]
*2025-12-08*

### Added
- markdown input/output plugins for theme portability
- Theme save and restore functionality

## [0.1.10]
*2025-11-26*

### Added
- openrouter input plugin for AI-generated themes
- google-genai input plugin for Gemini integration

## [0.1.8]
*2025-11-25*

### Added
- remote-json input plugin
- remote-css input plugin

## [0.1.5]
*2025-11-19*

### Added
- Ambient edge/corner colour extraction for LED sync
- Position-based colour roles

## [0.1.0]
*2025-11-16*

### Added
- Initial release
- Image colour extraction with K-means clustering
- Material Design 3 semantic colour roles
- Output plugins: alacritty, kitty, ghostty, hyprland, hyprlock, hyprpaper, waybar, dunst, gtk3, gtk4, qt5, qt6, gnome-shell, kde-plasma, neovim, zellij, fuzzel, walker, wofi, swayosd, libadwaita
- Template system with colour format conversion
- File tracking and management
