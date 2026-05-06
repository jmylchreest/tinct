---
sidebar_position: 100
sidebar_label: Changelog
toc_max_heading_level: 2
---

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## [Unreleased]

## [0.3.1]
*2026-05-06*

## [0.3.0]
*2026-05-04*

## [0.2.4]
*2026-05-04*

### Added

- New external plugin **awob** for the [animated Wayland Overlay Bar](https://github.com/jmylchreest/awob).
  Writes `~/.config/awob/themes/_palettes/tinct.kdl` (a shared palette +
  named-style block) plus a `tinct/` theme directory containing
  `scene.kdl` and `manifest.toml`. Other awob themes can adopt the
  tinct palette by adding `import "../_palettes/tinct.kdl"` to their
  scene; the palette refreshes in place on every `tinct generate` and
  awob-daemon hot-reloads. Install with `tinct plugin install awob`.

### Removed

- **BREAKING**: AUR per-plugin packages (`tinct-plugin-random-bin`,
  `tinct-plugin-wob-bin`, `tinct-plugin-templater-bin`,
  `tinct-plugin-dunstify-bin`, `tinct-plugin-zed-bin`,
  `tinct-plugin-opencode-bin`, `tinct-plugin-keylightd-tray-bin`) are no
  longer built. Plugins should be installed via the plugin manager
  (`tinct plugin install <name>`).

  This is the canonical install path — it pins versions, sandboxes
  installs to `~/.local/share/tinct/plugins/`, and uninstalls cleanly.
  The `tinct-bin` AUR package for tinct itself remains supported. The
  goreleaser config keeps a commented-out `tinct-plugin-templater-bin`
  block as a copy-paste reference for anyone re-enabling AUR
  distribution for a specific plugin.

### Changed

- **BREAKING**: `ptyxis` is no longer a built-in output plugin. It has
  been migrated to an external plugin so its bespoke Flatpak detection,
  dual-mode `[Light]`/`[Dark]` palette generation, and live-reload
  install hint can iterate independently of tinct releases. Users who
  previously relied on the built-in plugin can install the external
  one with:

      tinct plugin install ptyxis

  Behaviour and output paths are unchanged once installed. The
  `--ptyxis.output-dir` flag becomes `--ptyxis.output-dir` on the
  external plugin (same name, same semantics) and continues to honour
  the `TINCT_PLUGIN_PTYXIS_OUTPUT_DIR` environment variable.

## [0.2.3]
*2026-05-03*

## [0.2.2]
*2026-04-10*

## [0.2.1]
*2026-03-18*

## [0.2.0]
*2026-03-18*

## [0.1.28]
*2026-03-14*

## [0.1.27]
*2026-03-12*

## [0.1.26]
*2026-02-20*

## [0.1.24]
*2026-02-10*

## [0.1.23]
*2026-02-10*

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
