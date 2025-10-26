# Tinct

> An extensible colour palette generator and theme manager for unified theming across your entire environment

Generate colour palettes from multiple sources (images, remote JSON/CSS themes, manual lists) with intelligent categorization, role hints, weighting, and positional extraction. Apply them consistently to terminals, window managers, text editors, notification daemons, and even external LED lighting. Built with Go and designed for extensibility.

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Features

- **Multiple Input Sources**: Extract from images, fetch from remote JSON/CSS, or provide manual colour lists
- **Role Hints & Weighting**: Override auto-categorization with explicit role assignments and colour importance weights
- **Positional Extraction**: Extract edge/corner colours from images for ambient lighting and LED synchronization
- **Smart Categorization**: Auto-assigns background, foreground, accent, and semantic colours with WCAG contrast checking
- **Theme-Aware**: Detects or forces dark/light themes with accessibility compliance
- **Highly Extensible**: Plugin system for inputs (image, remote JSON/CSS, file) and outputs (applications, LED devices)
- **External Device Support**: Send colours to LED strips, smart lights, and other RGB peripherals
- **Unified Theming**: Apply consistent colour schemes across your entire environment

## Quick Start

### Installation

```bash
go install github.com/jmylchreest/tinct/cmd/tinct@latest
```

Or build from source:
```bash
git clone https://github.com/jmylchreest/tinct.git
cd tinct && go build -o tinct ./cmd/tinct
```

### Basic Usage

```bash
# Extract and preview colours
tinct extract --preview wallpaper.jpg

# Generate themes for applications
tinct generate -i image -p wallpaper.jpg -o hyprland,kitty,waybar

# Use a remote theme (Catppuccin Mocha)
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.mocha.colors" \
  --remote-json.map base=background,text=foreground,red=danger,green=success \
  -o hyprland,kitty

# Preview categorized palette
tinct extract --categorise --preview wallpaper.jpg
```

## Available Plugins

### Input Plugins
- **image**: Extract from images (JPEG, PNG, GIF, WebP) with optional ambient edge/corner extraction
- **remote-json**: Fetch from JSON URLs with JSONPath queries
- **remote-css**: Extract from CSS files (variables, hex codes)
- **file**: Load from saved palettes

### Output Plugins

**Applications:**
- **hyprland**: Hyprland window manager
- **kitty**: Kitty terminal emulator
- **waybar**: Waybar status bar
- **dunst**: Dunst notification daemon
- **fuzzel**: Fuzzel application launcher
- **hyprlock**: Hyprlock screen locker
- **swayosd**: SwayOSD on-screen display
- **wofi**: Wofi application launcher
- **neovim**: Neovim text editor (Lua colorschemes)
- **zellij**: Zellij terminal multiplexer

**External Devices:**
- Write custom output plugins to control LED strips (e.g., WLED, Philips Hue, Govee)
- Use ambient edge/corner extraction to sync bias lighting with your wallpaper
- Support for any device with a JSON/HTTP API or command-line interface

See [PLUGINS-WISHLIST.md](PLUGINS-WISHLIST.md) for planned plugins and [External Plugins Guide](docs/external-plugins.md) for creating device controllers.

**Note on Plugin Templates:** The current application plugin templates are based on online examples and personal configurations. They may benefit from refactoring for broader adoption. Contributions are significantly welcome, especially for plugins that make sense to be shipped and managed as part of Tinct. If you have expertise with any of these applications or want to add support for new ones, please see [Contributing](DEVELOPMENT.md).

## Plugin Management

```bash
# List available plugins
tinct plugins list

# Install external plugin
tinct plugins install <github-user>/<repo> [<ref>]

# Add plugin repository
tinct plugins repo add <name> <url>

# Sync lock file with installed plugins
tinct plugins sync

# Enable/disable plugins
export TINCT_ENABLED_PLUGINS="output:hyprland,output:kitty"
```

## Documentation

- **[Theme Cookbook](docs/THEME-COOKBOOK.md)**: Ready-to-use commands for popular themes (Catppuccin, Dracula, Tokyo Night, Nord, Gruvbox, and more)
- **[Template Guide](docs/TEMPLATE_GUIDE.md)**: Comprehensive template reference and customization guide
- **[Plugin Development](contrib/README.md)**: Create custom input/output plugins
- **[Plugin Hooks](docs/plugin-hooks.md)**: Pre/post-execution hooks
- **[External Plugins](docs/external-plugins.md)**: Write plugins in any language
- **[Setup Guide](SETUP.md)**: Detailed installation and configuration
- **[Development Guide](DEVELOPMENT.md)**: Contributing and architecture

## Examples

### Generate for Multiple Apps
```bash
tinct generate -i image -p wallpaper.jpg -o hyprland,kitty,waybar,dunst
```

### Use Custom Colours
```bash
# Create palette file
cat > my-theme.txt << EOF
background=#1e1e2e
foreground=#cdd6f4
accent1=#f38ba8
danger=#f38ba8
success=#a6e3a1
warning=#f9e2af
info=#89dceb
EOF

# Apply to apps
tinct generate -i file --file.path my-theme.txt -o hyprland,kitty
```

### Ambient LED Lighting / External Devices
```bash
# Extract edge/corner colours for synchronized LED bias lighting
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8 \
  --preview

# WLED monitor ambient lighting (see contrib/plugins/output/wled-ambient.sh)
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8 \
  -o wled-ambient \
  --plugin-args 'wled-ambient={"host":"192.168.1.100","segments":[0]}'

# OpenRGB keyboard/mouse lighting (see contrib/plugins/output/openrgb-peripheral.sh)
tinct generate -i image -p wallpaper.jpg \
  -o openrgb-peripheral \
  --plugin-args 'openrgb-peripheral={"devices":["keyboard","mouse"]}'
```

Positional colours (topLeft, top, topRight, left, right, bottomLeft, bottom, bottomRight) are automatically extracted and categorized when `--image.extractAmbience` is enabled. These can be consumed by custom output plugins to control:
- **LED strips**: WLED, Govee, Philips Hue (see `wled-ambient.sh` example)
- **RGB peripherals**: Keyboards, mice, case lighting via OpenRGB (see `openrgb-peripheral.sh` example)
- **Smart lights**: HomeKit, Home Assistant integrations

Example plugins demonstrating device control are in `contrib/plugins/output/`. See [External Plugins Guide](docs/external-plugins.md) for writing your own device controllers.

### With Plugin Hooks
```bash
# Reload Kitty after generating theme
tinct generate -i image -p wallpaper.jpg -o kitty --kitty.reload
```

## Project Structure

```
tinct/
├── cmd/tinct/              # CLI entry point
├── internal/
│   ├── colour/             # Colour extraction & categorization
│   ├── plugin/             # Plugin system
│   │   ├── input/          # Input plugins
│   │   ├── output/         # Output plugins
│   │   └── manager/        # Plugin management
│   └── cli/                # Command handlers
├── docs/                   # Documentation
├── contrib/                # Plugin development guides
└── testdata/               # Test fixtures
```

## Architecture

Tinct uses a flexible plugin architecture that makes it easy to:
- Add new input sources (image formats, remote APIs, databases)
- Add new output targets (applications, devices, services)
- Customize colour extraction and categorization logic
- Integrate with external tools and workflows

Plugins can be written in any language using the external plugin interface (stdin/stdout JSON protocol) or as native Go plugins for maximum performance.

## License

MIT License - see [LICENSE](LICENSE) file.

## Author

**John Mylchreest**
- Email: jmylchreest@gmail.com
- GitHub: [@jmylchreest](https://github.com/jmylchreest)

## Contributing

Contributions welcome! See [DEVELOPMENT.md](DEVELOPMENT.md) and [contrib/README.md](contrib/README.md).

---

Made with Go 1.25+
