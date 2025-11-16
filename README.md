# Tinct

> An extensible colour palette generator and theme manager for unified theming across your entire environment

Generate colour palettes from multiple sources (images, remote JSON/CSS themes, manual lists) with intelligent categorisation, role hints, weighting, and positional extraction. Apply them consistently to terminals, window managers, text editors, notification daemons, and even external LED lighting. Built with Go and designed for extensibility.

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Features

- **Multiple Input Sources**: Extract from images, fetch from remote JSON/CSS, or provide manual colour lists
- **Role Hints & Weighting**: Override auto-categorisation with explicit role assignments and colour importance weights
- **Positional Extraction**: Extract edge/corner colours from images for ambient lighting and LED synchronization
- **Smart Categorisation**: Auto-assigns background, foreground, accent, and semantic colours with WCAG contrast checking
- **Theme-Aware**: Detects or forces dark/light themes with accessibility compliance
- **Highly Extensible**: Plugin system for inputs (image, remote JSON/CSS, file) and outputs (applications, LED devices)
- **External Device Support**: Send colours to LED strips, smart lights, and other RGB peripherals
- **Unified Theming**: Apply consistent colour schemes across your entire environment

## Quick Start

### Installation

**Arch Linux (AUR):**
```bash
# Using yay
yay -S tinct-bin

# Using paru
paru -S tinct-bin
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
# Add official plugin repository (for external plugins)
tinct plugins repo add official https://github.com/jmylchreest/tinct-plugins

# Extract and preview colours from image
tinct extract --preview wallpaper.jpg

# Generate themes from local wallpaper
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all

# Use Catppuccin Mocha theme from remote JSON
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o all

# Generate image with Google Gemini and apply theme (requires API key)
export GOOGLE_API_KEY="AIzaSyC7xM4deUpK3Y.eX4MPleHiJkL"
tinct generate -i google-genai \
  --prompt "beautiful sunset over the rolling mountains and vineyards of tuscany" \
  -o all

# Preview categorized palette with role assignments
tinct extract --categorise --preview wallpaper.jpg
```

## Available Plugins

### Input Plugins
| Plugin | Description |
|--------|-------------|
| **image** | Extract from images (JPEG, PNG, GIF, WebP) with optional ambient edge/corner extraction |
| **google-genai** | Generate images with Google Gemini Imagen and extract colours |
| **remote-json** | Fetch from JSON URLs with JSONPath queries |
| **remote-css** | Extract from CSS files (variables, hex codes) |
| **file** | Load from saved palettes or manual colour specifications |

### Output Plugins

**Terminals:**
| Plugin | Application |
|--------|-------------|
| **alacritty** | Alacritty terminal emulator |
| **kitty** | Kitty terminal emulator |

**Window Managers & Compositors:**
| Plugin | Application |
|--------|-------------|
| **hyprland** | Hyprland compositor (colour themes) |
| **hyprpaper** | Hyprpaper wallpaper manager (wallpaper config and auto-apply) |
| **hyprlock** | Hyprlock screen locker (colours and wallpaper) |

**Bars & Notifications:**
| Plugin | Application |
|--------|-------------|
| **waybar** | Waybar status bar |
| **dunst** | Dunst notification daemon |
| **swayosd** | SwayOSD on-screen display |

**Launchers:**
| Plugin | Application |
|--------|-------------|
| **fuzzel** | Fuzzel application launcher |
| **walker** | Walker application launcher |
| **wofi** | Wofi application launcher |

**Editors & Multiplexers:**
| Plugin | Application |
|--------|-------------|
| **neovim** | Neovim text editor (Lua colour schemes) |
| **zellij** | Zellij terminal multiplexer (KDL) |

**External Devices & Custom:**
- Write custom output plugins to control LED strips (e.g., WLED, Philips Hue, Govee)
- Use ambient edge/corner extraction to sync bias lighting with your wallpaper
- Support for any device with a JSON/HTTP API or command-line interface

See [Plugin Wishlist](docs/PLUGINS-WISHLIST.md) for planned plugins and [External Plugins Guide](contrib/plugins/README.md) for creating device controllers.

**Note on Plugin Templates:** The current application plugin templates are based on online examples and personal configurations. They may benefit from refactoring for broader adoption. Contributions are significantly welcome, especially for plugins that make sense to be shipped and managed as part of Tinct. If you have expertise with any of these applications or want to add support for new ones, please see [Contributing](docs/DEVELOPMENT.md).

## Template Functions Reference

Tinct uses Go templates for generating configuration files. Templates have access to the colour palette and various helper functions.

### Common Template Functions

| Function | Description | Example |
|----------|-------------|---------|
| `get . "role"` | Get colour by role name | `{{ get . "background" \| hex }}` → `#1e1e2e` |
| `has . "role"` | Check if role exists | `{{ if has . "border" }}...{{ end }}` |
| `ansi . "colour"` | Get closest ANSI colour | `{{ ansi . "red" \| hex }}` |
| `themeType .` | Get theme type | `{{ themeType . }}` → `dark` |
| `.WallpaperPath` | Wallpaper path (if available) | `{{ .WallpaperPath }}` → `/path/to/wallpaper.jpg` |

### Colour Format Functions

| Function | Output Format | Example |
|----------|---------------|---------|
| `hex` | `#RRGGBB` | `{{ get . "accent1" \| hex }}` → `#89b4fa` |
| `rgb` | `R, G, B` | `{{ get . "accent1" \| rgb }}` → `137, 180, 250` |
| `rgbDecimal` | `R G B` | `{{ get . "accent1" \| rgbDecimal }}` → `137 180 250` |
| `rgbSpaces` | `R G B` (space-separated) | `{{ get . "accent1" \| rgbSpaces }}` → `137 180 250` |
| `rgba` | `R, G, B, A` | `{{ get . "scrim" \| rgba }}` → `30, 30, 46, 0.9` |
| `hsl` | `H, S%, L%` | `{{ get . "accent1" \| hsl }}` → `217, 92%, 76%` |

### Alpha Channel Functions

| Function | Description | Example |
|----------|-------------|---------|
| `withAlpha colour alpha` | Set alpha channel (0.0-1.0) | `{{ withAlpha (get . "background") 0.9 \| rgba }}` |

### Iteration

```go
{{- range .AllColours }}
colour{{ .Index }} = {{ .Colour | hex }}
{{- end }}
```

**Available Fields:** `.Index`, `.Colour`, `.Role`, `.Luminance`, `.Weight`

See [Template Guide](docs/TEMPLATE_GUIDE.md) for complete documentation.

## Plugin Development

Tinct supports two types of plugins:

- **Built-in plugins** (Go): Compiled into the main binary for maximum performance
- **External plugins** (Any language): Standalone executables using JSON-stdio or gRPC protocols

For detailed plugin development guides:
- **[Built-in Plugin Development](docs/DEVELOPMENT.md)** - Creating Go plugins
- **[External Plugin Development](contrib/plugins/README.md)** - Creating plugins in any language (Bash, Python, etc.)

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
export TINCT_ENABLED_PLUGINS="hyprland,kitty"
```

### Plugin Lock File Configuration

Tinct uses a lock file (`.tinct-plugins.json`) to manage plugin state and configuration. The lock file is searched in the following order:
1. Current directory (`./.tinct-plugins.json`)
2. Home directory (`~/.tinct-plugins.json`)

**Lock File Structure:**

```json
{
  "version": "1",
  "enabled_plugins": [
    "image",
    "hyprland",
    "kitty"
  ],
  "disabled_plugins": [
    "waybar"
  ],
  "external_plugins": {
    "notify-send": {
      "name": "notify-send",
      "path": "/home/user/.local/share/tinct/plugins/notify-send.py",
      "type": "output",
      "version": "1.0.0",
      "description": "Send desktop notifications",
      "source": {
        "type": "local",
        "path": "./contrib/plugins/output/notify-send.py"
      },
      "installed_at": "2024-01-15T10:30:00Z"
    }
  }
}
```

**Configuration Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Lock file format version |
| `enabled_plugins` | array | List of explicitly enabled plugins by name |
| `disabled_plugins` | array | List of explicitly disabled plugins by name |
| `external_plugins` | object | Map of external plugin names to their metadata |

**External Plugin Metadata:**

| Field | Required | Description |
|-------|----------|-------------|
| `name` | ✓ | Plugin identifier (from `--plugin-info`) |
| `path` | ✓ | Absolute path to plugin executable |
| `type` | ✓ | Plugin type: `input` or `output` |
| `version` |  | Plugin version string |
| `description` |  | Human-readable description |
| `source` |  | Structured source information (type, path, url, etc.) |
| `installed_at` |  | ISO 8601 timestamp of installation |
| `config` |  | Plugin-specific configuration (see below) |

**Priority Order:**
1. Lock file settings (highest priority)
2. Environment variables (`TINCT_ENABLED_PLUGINS`, `TINCT_DISABLED_PLUGINS`)
3. Plugin defaults (lowest priority)

### Plugin-Specific Configuration

Plugins can store configuration in the lock file under the `config` field. This configuration is applied automatically when the plugin is used.

**Image Plugin Configuration:**

The `image` input plugin supports deterministic k-means clustering via seed configuration:

```json
{
  "external_plugins": {
    "image": {
      "name": "image",
      "type": "input",
      "config": {
        "seed": {
          "mode": "content",
          "value": null
        }
      }
    }
  }
}
```

**Seed Modes:**

| Mode | Description | Use Case |
|------|-------------|----------|
| `content` | Hash of image pixel data (default) | Same image → same colours (content-based) |
| `filepath` | Hash of absolute file path | Same location → same colours (path-based) |
| `manual` | User-provided seed value | Reproducible results with custom seed |
| `random` | Non-deterministic random seed | Different colours each run |

**Example: Manual Seed Configuration**

```json
{
  "config": {
    "seed": {
      "mode": "manual",
      "value": 42
    }
  }
}
```

Or via command-line flags:

```bash
# Use content-based seed (default)
tinct generate -i image -p wallpaper.jpg --image.seed-mode content

# Use filepath-based seed
tinct generate -i image -p wallpaper.jpg --image.seed-mode filepath

# Use manual seed
tinct generate -i image -p wallpaper.jpg --image.seed-mode manual --image.seed-value 42

# Use random seed (non-deterministic)
tinct generate -i image -p wallpaper.jpg --image.seed-mode random
```

**How Configuration is Passed to Plugins:**

1. **Built-in Plugins:** CLI flags (e.g., `--image.seed-mode`) are parsed and applied directly
2. **External Plugins:** Configuration is passed via JSON on stdin during plugin initialization
3. **Lock File Config:** Merged with CLI flags (CLI flags take precedence)

**Environment Variables:**

```bash
# Enable only specific plugins (whitelist mode)
export TINCT_ENABLED_PLUGINS="image,hyprland,kitty"

# Disable specific plugins (blacklist mode)
export TINCT_DISABLED_PLUGINS="waybar,dunst"
```

## Documentation

- **[Theme Cookbook](docs/THEME-COOKBOOK.md)**: Ready-to-use commands for popular themes (Catppuccin, Dracula, Tokyo Night, Nord, Gruvbox, and more)
- **[Template Guide](docs/TEMPLATE_GUIDE.md)**: Comprehensive template reference and customization guide
- **[Plugin Development](contrib/README.md)**: Create custom input/output plugins
- **[Plugin Hooks](docs/plugin-hooks.md)**: Pre/post-execution hooks
- **[External Plugins](docs/external-plugins.md)**: Write plugins in any language
- **[Setup Guide](docs/SETUP.md)**: Detailed installation and configuration
- **[Development Guide](docs/DEVELOPMENT.md)**: Contributing and architecture

## Examples

### Generate themes from wallpaper
```bash
# Extract colours and apply to all configured applications
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all

# hyprpaper plugin automatically:
# - Generates wallpaper config with preload/wallpaper directives
# - Applies wallpaper immediately (if hyprpaper is running)
# - Updates all monitors or preserves existing assignments
```

### Use Catppuccin Mocha theme
```bash
# Fetch Catppuccin Mocha from remote JSON and apply to all apps
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o all

# Apply to specific applications only
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o hyprland,kitty,waybar,walker
```

### Generate with AI (Google Gemini)
```bash
# Set up your Google API key (get one from https://aistudio.google.com/apikey)
export GOOGLE_API_KEY="AIzaSyC7x8K9mN3pQ2rT5vW8yZ1aBcDeFgHiJkL"

# Generate image with Gemini Imagen and extract colours
tinct generate -i google-genai \
  --prompt "beautiful sunset over the rolling mountains and vineyards of tuscany" \
  -o all

# The google-genai plugin:
# - Generates an image using Google's Imagen model
# - Saves the image to your pictures directory
# - Automatically extracts colours from the generated image
# - Provides the image path to wallpaper plugins (hyprpaper, hyprlock)
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
│   ├── colour/             # Colour extraction & categorisation
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
- Customise colour extraction and categorisation logic
- Integrate with external tools and workflows

Plugins can be written in any language using the external plugin interface (stdin/stdout JSON protocol) or as native Go plugins for maximum performance.

## License

MIT License - see [LICENSE](LICENSE) file.

## Author

**John Mylchreest**
- Email: jmylchreest@gmail.com
- GitHub: [@jmylchreest](https://github.com/jmylchreest)

## Contributing

Contributions welcome! See [Development Guide](docs/DEVELOPMENT.md) and [Plugin Development Guide](contrib/README.md).

---

Made with Go 1.25+
