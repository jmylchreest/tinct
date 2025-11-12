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
# Extract and preview colours from image
tinct extract --preview wallpaper.jpg

# Generate themes from wallpaper (colours + wallpaper auto-applied)
tinct generate -i image -p wallpaper.jpg -o hyprland,hyprpaper,hyprlock,kitty,waybar

# Use a remote theme (Catppuccin Mocha)
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  -o hyprland,kitty,waybar

# Preview categorized palette with role assignments
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
- **hyprland**: Hyprland window manager (colour themes)
- **hyprpaper**: Hyprpaper wallpaper manager (wallpaper config and auto-apply)
- **hyprlock**: Hyprlock screen locker (colours and wallpaper)
- **kitty**: Kitty terminal emulator
- **waybar**: Waybar status bar
- **dunst**: Dunst notification daemon
- **fuzzel**: Fuzzel application launcher
- **swayosd**: SwayOSD on-screen display
- **wofi**: Wofi application launcher
- **neovim**: Neovim text editor (Lua colour schemes)
- **zellij**: Zellij terminal multiplexer

**External Devices:**
- Write custom output plugins to control LED strips (e.g., WLED, Philips Hue, Govee)
- Use ambient edge/corner extraction to sync bias lighting with your wallpaper
- Support for any device with a JSON/HTTP API or command-line interface

See [Plugin Wishlist](docs/PLUGINS-WISHLIST.md) for planned plugins and [External Plugins Guide](docs/external-plugins.md) for creating device controllers.

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

## Plugin Interface Reference

### Input Plugin Interface

Input plugins extract colours from various sources and return a palette.

| Method | Required | Description |
|--------|----------|-------------|
| `Name()` | ✓ | Plugin identifier (e.g., "image", "remote-json") |
| `Description()` | ✓ | Human-readable description |
| `Generate(ctx, opts)` | ✓ | Extract colours, return `*colour.Palette` |
| `RegisterFlags(cmd)` | ✓ | Register CLI flags (e.g., `--image.path`) |
| `Validate()` | ✓ | Check required inputs are provided |
| `WallpaperPath()` | Optional | Return wallpaper file path (image plugin only) |
| `ThemeHint()` | Optional | Suggest "dark" or "light" theme type |

**Generate Method Returns:**

```go
type Palette struct {
    Colours []colour.Colour  // Extracted colours
}
```

**Optional Interfaces:**

| Interface | Method | Purpose |
|-----------|--------|---------|
| `WallpaperProvider` | `WallpaperPath() string` | Provide wallpaper path to output plugins |
| `ThemeHinter` | `ThemeHint() string` | Suggest theme type ("dark", "light", "auto") |

### Output Plugin Interface

Output plugins generate configuration files from categorized colours.

| Method | Required | Description |
|--------|----------|-------------|
| `Name()` | ✓ | Plugin identifier (e.g., "kitty", "hyprland") |
| `Description()` | ✓ | Human-readable description |
| `Generate(palette)` | ✓ | Generate config files, return `map[string][]byte` |
| `RegisterFlags(cmd)` | ✓ | Register CLI flags (e.g., `--kitty.output-dir`) |
| `Validate()` | ✓ | Check plugin configuration is valid |
| `DefaultOutputDir()` | ✓ | Return default config directory path |
| `SetVerbose(bool)` | Optional | Receive verbose flag setting |
| `PreExecute(ctx)` | Optional | Run checks before generation |
| `PostExecute(ctx, execCtx, files)` | Optional | Run actions after file write |
| `SetWallpaperContext(path)` | Optional | Receive wallpaper path before generation |
| `GetEmbeddedFS()` | Optional | Expose embedded templates for management |

**Generate Method Returns:**

```go
map[string][]byte{
    "tinct.conf":         []byte("..."),  // Config file content
    "tinct-colours.conf": []byte("..."),  // Colour definitions (optional)
}
```

**Optional Interfaces:**

| Interface | Method | Purpose | Example Usage |
|-----------|--------|---------|---------------|
| `VerbosePlugin` | `SetVerbose(bool)` | Receive verbose flag | Control output verbosity |
| `PreExecuteHook` | `PreExecute(ctx) (skip, reason, err)` | Pre-generation checks | Verify app is installed |
| `PostExecuteHook` | `PostExecute(ctx, execCtx, files) error` | Post-generation actions | Reload config, set wallpaper |
| `WallpaperContextProvider` | `SetWallpaperContext(string)` | Receive wallpaper path | Include in templates |
| `TemplateProvider` | `GetEmbeddedFS() interface{}` | Expose templates | Template management |

**PostExecute ExecutionContext:**

```go
type ExecutionContext struct {
    DryRun        bool   // Is this a dry-run?
    Verbose       bool   // Verbose mode enabled?
    OutputDir     string // Output directory path
    WallpaperPath string // Wallpaper path (if available)
}
```

### Plugin Implementation Patterns

**Input Plugin Example:**
```go
type Plugin struct {
    path string  // From --image.path flag
}

func (p *Plugin) Generate(ctx, opts) (*colour.Palette, error) {
    // Extract colours from image
    return &colour.Palette{Colours: extractedColors}, nil
}

func (p *Plugin) WallpaperPath() string {
    return p.path  // Optional interface
}
```

**Output Plugin Example:**
```go
type Plugin struct {
    outputDir     string
    verbose       bool
    wallpaperPath string  // Optional, from SetWallpaperContext
}

func (p *Plugin) Generate(palette) (map[string][]byte, error) {
    // Generate config using templates
    return map[string][]byte{"tinct.conf": content}, nil
}

func (p *Plugin) PostExecute(ctx, execCtx, files) error {
    // Reload app config, set wallpaper, etc.
    return nil
}
```

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
# Extract colours and apply wallpaper
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o hyprland,hyprpaper,hyprlock,kitty,waybar

# hyprpaper plugin automatically:
# - Generates wallpaper config with preload/wallpaper directives
# - Applies wallpaper immediately (if hyprpaper is running)
# - Updates all monitors or preserves existing assignments
```

### Use custom colours (no wallpaper)
```bash
# Generate from colour specification
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  -o hyprland,hyprlock,kitty,waybar

# hyprpaper/hyprlock still work without wallpaper source
# - Config files generated with helpful placeholders
# - Wallpaper application skipped gracefully
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
