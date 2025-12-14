# Tinct

> An extensible colour palette generator and theme manager for unified theming across your entire environment

Generate colour palettes from images, AI, remote themes, or manual specifications. Apply them consistently to terminals, desktop environments, window managers, and external devices. Built with Go for speed and extensibility.

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Quick Start

### Installation

**Arch Linux (AUR):**
```bash
yay -S tinct-bin      # Using yay
paru -S tinct-bin     # Using paru
```

**From Go:**
```bash
go install github.com/jmylchreest/tinct/cmd/tinct@latest
```

**From source:**
```bash
git clone https://github.com/jmylchreest/tinct.git
cd tinct && go build -o tinct ./cmd/tinct
```

### Basic Usage

```bash
# Extract colors from an image
tinct extract -i image -p ~/Pictures/wallpaper.jpg

# Apply theme to all configured applications
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all

# Use a specific theme (Catppuccin Mocha)
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o all

# Generate AI image and apply theme (requires Google API key)
export GOOGLE_API_KEY="your-api-key"
tinct generate -i google-genai \
  --prompt "beautiful sunset over mountains" \
  -o all
```

## Features

- **Multiple Input Sources** - Extract from images, generate with AI, fetch remote themes (JSON/CSS), or specify manually
- **Smart Categorisation** - Auto-assigns background, foreground, accent, and semantic colours with WCAG contrast checking
- **Unified Theming** - Apply consistent color schemes across 25+ applications and desktop environments
- **External Device Support** - Control LED strips, smart lights, and RGB peripherals
- **Plugin Architecture** - Extend with custom inputs/outputs in any language
- **Theme-Aware** - Automatic dark/light theme detection with accessibility compliance

## Supported Applications

Tinct supports 25+ applications across terminals, desktop environments, window managers, and more:

| Category | Applications |
|----------|-------------|
| **Terminals** | Alacritty, Ghostty, Kitty |
| **Desktop Environments** | GNOME Shell, KDE Plasma, GTK3/GTK4, Libadwaita, Qt5/Qt6 |
| **Window Managers** | Hyprland, Hyprpaper, Hyprlock |
| **Bars & Notifications** | Waybar, Dunst, SwayOSD |
| **Launchers** | Fuzzel, Walker, Wofi |
| **Editors & Multiplexers** | Neovim, Zed, Zellij |

**See [Plugin Reference](docs/PLUGINS.md) for complete plugin list and configuration.**

## Example Workflows

### Theme from Wallpaper

```bash
# Extract and apply
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all

# With ambient lighting extraction for LED sync
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8 \
  -o all,wled-ambient
```

### Use Popular Themes

```bash
# Catppuccin Mocha
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o all

# Tokyo Night
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/folke/tokyonight.nvim/main/extras/lua/tokyonight_colors.lua" \
  --remote-json.query "night" \
  -o hyprland,kitty,waybar
```

**See [Theme Cookbook](docs/THEME-COOKBOOK.md) for ready-to-use commands for Dracula, Nord, Gruvbox, and more.**

### AI-Generated Themes

```bash
# Get Google API key from https://aistudio.google.com/apikey
export GOOGLE_API_KEY="your-key-here"

# Generate image and apply theme
tinct generate -i google-genai \
  --prompt "sunset over rolling hills of tuscany" \
  -o all
```

### Save and Restore Themes

Save any theme to a portable markdown file with embedded wallpaper:

```bash
# Save theme from online wallpaper to a markdown file
tinct generate -i image \
  -p "https://example.com/wallpaper.jpg" \
  -o markdown \
  --markdown.name "Mountain Sunset" \
  --markdown.include-thumbnail

# The theme is saved to ~/.config/tinct/themes/mountain-sunset.md
```

Restore the theme later from the markdown file:

```bash
# Restore theme and wallpaper from the saved markdown file
tinct generate -i markdown \
  --markdown.path ~/.config/tinct/themes/mountain-sunset.md

# The wallpaper is automatically extracted and applied
```

This workflow allows you to:
- Share complete themes (colors + wallpaper) as single portable files
- Version control your themes in git
- Restore your exact setup on a new machine

## Documentation

| Document | Description |
|----------|-------------|
| [Plugin Reference](docs/PLUGINS.md) | Complete plugin list and configuration |
| [Theme Cookbook](docs/THEME-COOKBOOK.md) | Ready-to-use commands for popular themes |
| [Template Guide](docs/TEMPLATE_GUIDE.md) | Template functions and customization |
| [Setup Guide](docs/SETUP.md) | Detailed installation and configuration |
| [Development Guide](docs/DEVELOPMENT.md) | Creating built-in Go plugins |
| [External Plugins](contrib/plugins/README.md) | Creating plugins in any language |
| [Screenshots Guide](docs/SCREENSHOTS.md) | Adding screenshots to documentation |

## Plugin Management

```bash
# Add official plugin repository
tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/refs/heads/main/repository/repository.json

# List all plugins
tinct plugins list

# Install external plugin
tinct plugins install <github-user>/<repo>

# Enable/disable specific plugins
export TINCT_ENABLED_PLUGINS="image,hyprland,kitty,kde-plasma"
export TINCT_DISABLED_PLUGINS="waybar"
```

## Creating Plugins

Tinct supports two types of plugins:

1. **Built-in plugins (Go)** - Compiled into binary for maximum performance
2. **External plugins (Any language)** - Standalone executables using JSON-stdio protocol

**Quick External Plugin Example:**

```bash
#!/bin/bash
# contrib/plugins/output/my-app.sh

read -r palette
background=$(echo "$palette" | jq -r '.palette.background | .hex')

cat > ~/.config/my-app/colors.conf <<EOF
background = $background
EOF
```

Make it executable and use it:

```bash
chmod +x my-app.sh
tinct generate -i image -p wallpaper.jpg -o my-app --plugin-path ./my-app.sh
```

**See [External Plugins Guide](contrib/plugins/README.md) for detailed examples including LED controllers.**

## Project Structure

```
tinct/
├── cmd/tinct/                    # CLI entry point
├── internal/
│   ├── colour/                   # Color extraction & categorization
│   ├── plugin/
│   │   ├── input/                # Input plugins (image, AI, remote)
│   │   ├── output/               # Output plugins (apps, devices)
│   │   └── manager/              # Plugin management
│   └── cli/                      # Command handlers
├── docs/                         # Documentation
│   ├── PLUGINS.md                # Plugin reference
│   ├── THEME-COOKBOOK.md         # Theme recipes
│   ├── TEMPLATE_GUIDE.md         # Template reference
│   └── *.md                      # Other guides
├── contrib/plugins/              # Example external plugins
└── .github/screenshots/          # Documentation screenshots
```

## Advanced Features

### Ambient Lighting & External Devices

Extract positional colors for LED synchronization:

```bash
# Extract edge/corner colors
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8 \
  -o wled-ambient \
  --plugin-args 'wled-ambient={"host":"192.168.1.100"}'
```

Example plugins for external devices in `contrib/plugins/output/`:
- WLED ambient lighting (`wled-ambient.sh`)
- OpenRGB peripheral control (`openrgb-peripheral.sh`)
- Philips Hue integration (see contrib)

**See [External Plugins Guide](contrib/plugins/README.md) for writing device controllers.**

### Role Hints & Weighting

Override auto-categorization with explicit role assignments:

```bash
# Assign specific colors to roles
tinct generate -i file \
  --file.colors "#1e1e2e,#cdd6f4,#89b4fa" \
  --file.hints "background=0,foreground=1,accent1=2" \
  -o all
```

### Plugin Hooks

Automatically reload applications after theme changes:

```bash
# Reload Kitty after generating theme
tinct generate -i image -p wallpaper.jpg -o kitty --kitty.reload

# Hyprpaper automatically applies wallpaper
tinct generate -i image -p wallpaper.jpg -o hyprpaper
```

**See [Plugin Hooks](docs/plugin-hooks.md) for details.**

## Contributing

Contributions welcome! We especially need:

- Plugin templates for broader adoption
- New application/device support
- Documentation improvements
- Bug reports and feature requests

**See [Development Guide](docs/DEVELOPMENT.md) and [Plugin Wishlist](docs/PLUGINS-WISHLIST.md).**

## License

MIT License - see [LICENSE](LICENSE)

## Acknowledgments

Built with Go 1.25+. Color theory based on Material Design 3 and WCAG 2.1 guidelines.

---

**Note:** Application plugin templates are based on online examples and personal configurations. They may benefit from refactoring for broader adoption. Contributions are significantly welcome!
