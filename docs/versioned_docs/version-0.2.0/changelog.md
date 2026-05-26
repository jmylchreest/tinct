---
sidebar_position: 100
sidebar_label: Changelog
toc_max_heading_level: 2
---

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## [0.2.0]
*2026-03-18*

### Fixed

- fix: implement JSON-stdio structured response protocol (v0.2.0) (c65bbb3)

### Changed

- refactor: replace aptabase telemetry with statsfactory SDK (ab6d98a)

## [0.1.28]
*2026-03-14*

### Fixed

- fix: restore wallpaper format detection for URLs without file extensions (60810ea)

### Changed

- refactor: remove security.ValidateFilePath in favour of plugin-managed output paths (52770fb)
- refactor: eliminate 3 remaining clone groups (RPC client, path utils, sync finalization) (0104c9a)
- refactor: extract shared two-pass tar extraction into extractFromTar (aab71bb)
- refactor: eliminate code duplication across 6 clone groups (da18f5a)

## [0.1.27]
*2026-03-12*

## [0.1.26]
*2026-02-20*

### Added

- feat: centralise config to tinct.toml, async batched telemetry, remove plugin enable/disable (6b11cf8)

### Fixed

- fix: resolve lint warnings (errcheck, errname, gocognit) (0a8151c)

## [0.1.25]
*2026-02-19*

### Added

- feat: add anonymous usage telemetry via Aptabase (b3c137b)
- feat: add wbg (simple Wayland wallpaper) output plugin (7127b75)
- feat: add awww (animated Wayland wallpaper daemon) output plugin (e8049b8)
- feat: add Midnight Commander (mc) skin plugin and xterm256 color conversion (46bed02)
- feat: add OpenCode theme output plugin (0044489)

### Fixed

- fix: deploy docs from latest main after versioning (b7d0c7b)

## [0.1.24]
*2026-02-10*

### Added

- feat: add Midnight Commander (mc) skin plugin and xterm256 color conversion (3d3541a)

## [0.1.23]
*2026-02-10*

### Added

- feat: add OpenCode theme output plugin (ba85ac0)

### Fixed

- fix: changelog TOC and release workflow (64f4995)

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
