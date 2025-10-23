# Tinct 

> A modern, blazingly fast colour palette generator for your terminal and beyond

Tinct extracts vibrant colour palettes from images using advanced clustering algorithms. Perfect for generating colour schemes from wallpapers and applying them system-wide to your favourite applications.

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

##  Features

-  **Blazingly Fast**: K-means clustering optimised for speed (< 0.01s for most images)
-  **Smart Categorisation**: Automatically categorises colours by role (background, foreground, accents)
-  **Colour Previews**: See actual colours in your terminal with ANSI colour blocks
-  **Theme-Aware**: Auto-detects or forces dark/light theme with WCAG contrast checking
-  **Semantic Colours**: Intelligent assignment of danger, warning, success, info, notification colours
-  **Muted Variants**: Automatically generates muted background/foreground for inactive states
-  **Remote Palettes**: Fetch themes from URLs (JSON, CSS) with role mapping
-  **[Theme Cookbook](docs/THEME-COOKBOOK.md)**: Ready-to-use commands for Catppuccin, Dracula, Tokyo Night, Nord, Gruvbox, Solarized
-  **Multiple Formats**: Output as hex, RGB, JSON, or categorised
-  **Wide Format Support**: JPEG, PNG, GIF, WebP
-  **Simple CLI**: Built with Cobra for intuitive command-line usage
-  **Zero Dependencies**: Self-contained binary with no runtime dependencies
-  **Well Tested**: Comprehensive unit tests with >80% coverage
-  **Accessibility**: WCAG 2.0 compliant contrast ratio calculations

##  Quick Start

### Installation

#### From Source

```bash
# Clone the repository
git clone https://github.com/jmylchreest/tinct.git
cd tinct

# Build
go build -o tinct ./cmd/tinct

# Install (optional)
go install ./cmd/tinct
```

#### Using Go Install

```bash
go install github.com/jmylchreest/tinct/cmd/tinct@latest
```

### Basic Usage

```bash
# Extract 16 colours (default) from an image
tinct extract wallpaper.jpg

# Extract 8 colours with preview
tinct extract --preview --colours 8 wallpaper.png

# Categorise colours for theming
tinct extract --categorise wallpaper.jpg

# Categorise with previews
tinct extract --categorise --preview wallpaper.jpg

# Output as JSON
tinct extract --format json wallpaper.jpg

# Save to file
tinct extract --output palette.txt wallpaper.jpg
```

### Popular Themes

See the **[Theme Cookbook](docs/THEME-COOKBOOK.md)** for ready-to-use commands for popular themes like Catppuccin, Dracula, Tokyo Night, and more!

**Quick example - Catppuccin Mocha:**
```bash
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.mocha.colors" \
  --remote-json.map base=background,text=foreground,red=danger,green=success \
  --theme dark \
  --outputs waybar
```

### Using Remote Palettes

Tinct can fetch and use pre-defined color palettes from popular themes using specialized remote plugins:

#### Remote JSON Plugin (with JSONPath queries)

```bash
# Catppuccin Mocha from GitHub (with nested JSON path)
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.colors.mocha" \
  --remote-json.map base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info \
  --outputs waybar

# Simple flat JSON palette
tinct generate --input remote-json \
  --remote-json.url "https://example.com/theme.json" \
  --remote-json.map primary=background,secondary=foreground \
  --outputs hyprland,kitty
```

#### Remote CSS Plugin (extracts CSS variables)

```bash
# Fetch colors from a CSS theme file
tinct generate --input remote-css \
  --remote-css.url "https://example.com/theme.css" \
  --remote-css.map color-base=background,color-text=foreground,color-accent=accent1 \
  --outputs waybar

# Extract all colors without mapping (automatic categorization)
tinct generate --input remote-css \
  --remote-css.url "https://example.com/colors.css" \
  --preview
```

#### Saving Palettes for Reuse

```bash
# Save palette for later use
tinct generate --input remote-json \
  --remote-json.url "https://example.com/palette.json" \
  --remote-json.map base=background,text=foreground \
  --save-palette my-theme.txt

# Then use the saved palette
tinct generate --input file --file.path my-theme.txt --outputs hyprland,kitty
```

**File Input Formats:**

Without role hints (automatic categorization):
```
1e1e2e
cdd6f4
f38ba8
a6e3a1
```

With explicit role hints (override auto-categorization):
```
background=#1e1e2e
foreground=#cdd6f4
danger=#f38ba8
success=#a6e3a1
```

##  Usage

### Extract Command

Extract a colour palette from an image:

```bash
tinct extract <image> [flags]
```

#### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--colours` | `-c` | `16` | Number of colours to extract (1-256) |
| `--algorithm` | `-a` | `kmeans` | Extraction algorithm (currently: kmeans) |
| `--format` | `-f` | `hex` | Output format (hex, rgb, json, categorised) |
| `--output` | `-o` | `stdout` | Output file path |
| `--categorise` | | `false` | Categorise colours by role (background, foreground, accent) |
| `--theme` | | `auto` | Theme type for categorisation (auto, dark, light) |
| `--preview` | | `false` | Show colour previews in terminal |
| `--verbose` | `-v` | `false` | Enable verbose output |
| `--quiet` | `-q` | `false` | Suppress non-error output |

### Examples

#### Basic Extraction with Previews

```bash
# Extract 8 colours and display with ANSI colour blocks
tinct extract wallpaper.jpg --preview --colours 8
```

Output:
```
 #ff0000
 #00ff00
 #0000ff
 #ffff00
 #ff00ff
 #00ffff
 #808080
 #ff8000
```

#### Colour Categorisation

```bash
# Auto-detect theme and categorise colours
tinct extract wallpaper.jpg --categorise --preview --colours 8
```

Output:
```
Theme Type: dark

Core Colours:
    background           #0000ff (luminance: 0.07)
    background-muted     #0b0b65 (luminance: 0.22)
    foreground           #ffff00 (luminance: 0.93)
    foreground-muted     #f3f399 (luminance: 0.78)
  Contrast ratio: 8.00:1

Accent Colours:
    accent-1             #ff0000 (sat: 1.00)
    accent-2             #00ff00 (sat: 1.00)
    accent-3             #ff00ff (sat: 1.00)
    accent-4             #00ffff (sat: 1.00)

Semantic Colours:
    danger               #ff0000
    warning              #ff8000
    success              #00ff00
    info                 #00ffff
    notification         #ff00ff
```

#### Theme-Specific Categorisation

```bash
# Force light theme (dark text on light background)
tinct extract wallpaper.jpg --categorise --theme light --preview
```

Output:
```
Theme Type: light

Core Colours:
    background           #ffff00 (luminance: 0.93)
    background-muted     #f3f399 (luminance: 0.78)
    foreground           #0000ff (luminance: 0.07)
    foreground-muted     #0b0b65 (luminance: 0.22)
  Contrast ratio: 8.00:1

Accent Colours:
    accent-1             #ff0000 (sat: 1.00)
    accent-2             #00ff00 (sat: 1.00)
  ...
```

#### Categorised JSON Output

```bash
# Get categorised colours as JSON for scripting
tinct extract wallpaper.jpg --categorise --format json --colours 6
```

Output:
```json
{
  "colours": {
    "background": {
      "role": "background",
      "hex": "#0000ff",
      "rgb": {"r": 0, "g": 0, "b": 255},
      "luminance": 0.07,
      "is_light": false,
      "hue": 240,
      "saturation": 1.0
    },
    "background-muted": {
      "role": "background-muted",
      "hex": "#0b0b65",
      "rgb": {"r": 11, "g": 11, "b": 101},
      "luminance": 0.22,
      "is_light": false
    },
    "foreground": {
      "role": "foreground",
      "hex": "#ffff00",
      "rgb": {"r": 255, "g": 255, "b": 0},
      "luminance": 0.93,
      "is_light": true
    },
    "foreground-muted": {
      "role": "foreground-muted",
      "hex": "#f3f399",
      "luminance": 0.78,
      "is_light": true
    },
    "accent-1": {...},
    "danger": {...},
    "warning": {...},
    "success": {...}
  },
  "theme_type": "dark"
}
```

#### Basic JSON Output

```bash
tinct extract wallpaper.jpg --format json --colours 4
```

Output:
```json
{
  "count": 4,
  "colors": [
    {
      "hex": "#ff0000",
      "rgb": {"r": 255, "g": 0, "b": 0}
    },
    {
      "hex": "#00ff00",
      "rgb": {"r": 0, "g": 255, "b": 0}
    }
  ]
}
```

#### RGB Format

```bash
tinct extract wallpaper.jpg --format rgb --preview --colours 5
```

Output:
```
  rgb(255, 0, 0)
  rgb(0, 255, 0)
  rgb(0, 0, 255)
  rgb(255, 255, 0)
  rgb(128, 128, 128)
```

#### Save Palette to File

```bash
# Extract and save hex colours
tinct extract wallpaper.jpg --output colours.txt

# Extract and save as JSON
tinct extract wallpaper.jpg --format json --output palette.json

# Extract categorised palette
tinct extract wallpaper.jpg --categorise --output theme.txt
```

### Version Command

Display version information:

```bash
tinct version
```

Output:
```
tinct version dev (go1.25.1, linux/amd64)
```

### Built-in Output Plugins

Tinct includes built-in plugins for generating theme configurations:

- **hyprland**: Hyprland window manager colour themes
  - Generates `~/.config/hypr/tinct-colours.conf`
  - Creates example stub configuration showing usage
  - Includes semantic colours and indexed palette

- **kitty**: Kitty terminal colour themes  
  - Generates `~/.config/kitty/tinct.conf`
  - Configures foreground, background, cursor, tabs, borders
  - Includes full 16-colour ANSI palette

- **waybar**: Waybar status bar colour themes
  - Generates `~/.config/waybar/tinct-colours.css`
  - Creates example CSS with GTK `@define-color` format
  - Includes comprehensive module styling and semantic colours
  - Optional auto-reload with `--waybar.reload` flag

To use a plugin, include the generated file in your configuration:

```bash
# Kitty example - add to ~/.config/kitty/kitty.conf
include ~/.config/kitty/tinct.conf

# Hyprland example - add to ~/.config/hypr/hyprland.conf  
source = ~/.config/hypr/tinct-colours.conf

# Waybar example - add to ~/.config/waybar/style.css
@import "tinct-colours.css";
```

### Plugin Management

Tinct features a powerful plugin system for managing input and output plugins. Plugins can be discovered from repositories, installed, and managed with a lock file for reproducibility.

#### Quick Start

```bash
# Add the official plugin repository
tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/main/repository.json

# Search for plugins
tinct plugins search notification

# Install a plugin
tinct plugins install random

# Sync plugins from lock file (on new machine)
tinct plugins sync
```

#### Plugin Repositories

Discover and install plugins from curated repositories:

```bash
# Add a repository
tinct plugins repo add official <manifest-url>

# List repositories
tinct plugins repo list

# Update repository manifests
tinct plugins repo update

# Search for plugins
tinct plugins search random
tinct plugins search --type input
tinct plugins search --tag notification

# Browse all available plugins
tinct plugins browse

# Show plugin details
tinct plugins info random
```

#### Installing Plugins

Install plugins from repositories or direct URLs:

```bash
# Install from repository (searches all repos)
tinct plugins install random

# Install specific version
tinct plugins install random@1.2.0

# Install from specific repository
tinct plugins install random --repo official

# Install from HTTP URL
tinct plugins add https://example.com/plugin.py

# Install local plugin
tinct plugins add ./my-plugin.sh
```

#### Lock File & Sync

Tinct tracks installed plugins in `.tinct-plugins.json`, making it easy to sync configurations across machines:

```bash
# Install missing plugins from lock file
tinct plugins sync

# Verify installed plugins match lock file
tinct plugins verify

# Reinstall all plugins
tinct plugins sync --force

# Remove plugins not in lock file
tinct plugins clean
```

**Workflow Example:**
```bash
# On machine 1: Install plugins normally
tinct plugins install random notify-send

# Commit lock file to dotfiles
cp ~/.config/tinct/.tinct-plugins.json ~/dotfiles/tinct/

# On machine 2: Clone and sync
git clone https://github.com/user/dotfiles ~/dotfiles
ln -sf ~/dotfiles/tinct/.tinct-plugins.json ~/.config/tinct/
tinct plugins sync  # Automatically installs random and notify-send
```

#### List & Manage Plugins

View and manage installed plugins:

```bash
# List all plugins
tinct plugins list

# Enable/disable plugins
tinct plugins enable hyprland
tinct plugins disable hyprland

# Update plugins to latest versions
tinct plugins update random
tinct plugins update --all

# Remove a plugin
tinct plugins delete notify --force
```

#### Plugin Priority

Plugin state is determined by (highest to lowest priority):
1. **Plugin lock file** (`.tinct-plugins.json`)
2. **Environment variables** (`TINCT_ENABLED_PLUGINS`, `TINCT_DISABLED_PLUGINS`)
3. **Plugin defaults**

**Environment Variables:**
```bash
# Enable only specific plugins (whitelist mode)
export TINCT_ENABLED_PLUGINS="output:hyprland,input:image"

# Disable specific plugins (blacklist mode)
export TINCT_DISABLED_PLUGINS="output:notify"
```

#### Custom Templates

Customize how output plugins generate configuration files by overriding their embedded templates:

```bash
# List available templates
tinct plugins templates list

# Extract templates for customization
tinct plugins templates dump -o hyprland,kitty

# Extract specific plugin templates
tinct plugins templates dump -o hyprland

# Overwrite existing custom templates
tinct plugins templates dump -o hyprland --force
```

Custom templates are stored in `~/.config/tinct/templates/{plugin-name}/` and automatically override the embedded versions. This allows you to:
- Customize output format and structure
- Add custom comments or documentation
- Modify variable names or color assignments
- Integrate with specific application configurations

**Example: Customizing Hyprland colors**
```bash
# Extract the template
tinct plugins templates dump -o hyprland

# Edit it
$EDITOR ~/.config/tinct/templates/hyprland/tinct-colours.conf.tmpl

# Generate output (uses your custom template)
tinct generate --input image -p wallpaper.jpg --outputs hyprland
```

See [Custom Templates Documentation](docs/CUSTOM-TEMPLATES.md) for detailed usage and examples.

For comprehensive plugin development guides, see [contrib/README.md](contrib/README.md).

**Repository Hosting:**
- [Host Your Own Repository](docs/HOSTING-PLUGIN-REPOSITORY.md) - Complete guide for hosting and maintaining plugin repositories
- [Repository Template](docs/repository-template/) - Template structure for creating your own plugin repository

##  Features in Detail

### Plugin Hooks

Tinct's output plugins support optional pre-execution and post-execution hooks for advanced functionality:

**Pre-Execution Hooks:**
- Check if required executables exist on `$PATH`
- Verify configuration directories exist
- Validate environment prerequisites
- Skip plugin execution gracefully if conditions aren't met

**Post-Execution Hooks:**
- Reload application configuration after theme generation
- Send signals to running processes
- Restart services
- Notify users of changes

**Example Usage:**

```bash
# Generate kitty theme and reload kitty instances
tinct generate -i image -p wallpaper.jpg --outputs kitty --kitty.reload

# Generate hyprland theme and reload hyprland
tinct generate -i image -p wallpaper.jpg --outputs hyprland --hyprland.reload

# Pre-hooks automatically skip plugins when prerequisites aren't met
tinct generate -i image -p wallpaper.jpg --outputs kitty --verbose
# Output: ⊘ Skipping kitty: kitty executable not found on $PATH
```

### Colour Categorisation

Tinct intelligently categorises extracted colours by their role in a colour scheme:

#### Core Roles
- **background**: Darkest (dark theme) or lightest (light theme) colour
- **background-muted**: Slightly adjusted background for inactive elements
- **foreground**: Highest contrast with background (WCAG compliant)
- **foreground-muted**: Reduced contrast foreground for secondary text

#### Accent Roles
- **accent-1**: Most vibrant colour (highest saturation)
- **accent-1-muted**: Muted variant of accent-1 for inactive/secondary states
- **accent-2**: Second most vibrant colour
- **accent-2-muted**: Muted variant of accent-2 for inactive/secondary states
- **accent-3**: Third most vibrant colour
- **accent-3-muted**: Muted variant of accent-3 for inactive/secondary states
- **accent-4**: Fourth most vibrant colour
- **accent-4-muted**: Muted variant of accent-4 for inactive/secondary states

Each accent has an automatically generated muted variant positioned immediately after it in the indexed palette (e.g., colour4=accent1, colour5=accent1Muted). These muted variants are perfect for inactive UI elements like disabled buttons, inactive tabs, or secondary borders.

#### Semantic Roles (enhanced for visibility)
- **danger**: Red hues (0°) - errors, destructive actions
- **warning**: Orange hues (45°) - warnings, caution
- **success**: Green hues (120°) - success, confirmation
- **info**: Blue hues (210°) - information, help
- **notification**: Purple hues (285°) - notifications, badges

**Semantic Color Enhancement**: Automatically boosts saturation (minimum 60%) and adjusts lightness for optimal visibility and meaning. If your palette lacks vibrant colors in specific hue ranges, Tinct generates appropriate fallback colors to ensure all semantic roles are available and distinct.

**Contrast Checking**: Ensures foreground and background colours meet WCAG AA standards (4.5:1 contrast ratio) for accessibility. Semantic colors are guaranteed minimum 3:1 contrast with background.

**Theme Detection**: Automatically detects whether to use a dark or light theme based on average luminance, or allows manual override with `--theme`. Semantic colors are adjusted based on theme (lighter for dark themes, darker for light themes).

**Muted Variants**: Automatically generates muted variants by adjusting luminance by 15% and reducing saturation by 20%, perfect for inactive UI elements.

See [docs/semantic-colors.md](docs/semantic-colors.md) for detailed information on semantic color enhancement.

### K-means Algorithm

Fast clustering algorithm that groups similar colours together and returns cluster centroids.

- **Speed**: Blazingly fast (< 0.01s for most images)
- **Optimisation**: Smart pixel sampling with K-means++ initialisation
- **Quality**: Excellent for extracting dominant colours
- **Best for**: General purpose colour extraction

**Usage:**
```bash
tinct extract image.jpg --algorithm kmeans --colours 16
```

**Performance Optimisations**:
- Aggressive sampling for large images (max 2000 pixels)
- K-means++ initialisation for better starting centroids
- Early convergence detection (stops when < 1% of points change)
- Maximum 20 iterations (reduced from 50)

### ANSI Colour Previews

Use the `--preview` flag to see actual colour blocks in your terminal:

```bash
tinct extract image.jpg --preview
```

This uses ANSI 24-bit true colour escape codes to display colour blocks () in your terminal. Works in most modern terminals including:
- iTerm2, Terminal.app (macOS)
- GNOME Terminal, Konsole, Alacritty, Kitty (Linux)
- Windows Terminal (Windows 10+)

### Coming Soon

- **Median Cut**: Perceptually balanced colour extraction
- **Dominant Colour**: Frequency-based colour extraction
- **Saturation Adjustment**: Boost or reduce colour vibrancy
- **Custom Contrast Ratios**: AAA compliance option (7:1)
- **Colour Harmony**: Complementary, triadic, analogous schemes

##  Development

### Prerequisites

- Go 1.25 or higher
### Building

```bash
# Build binary
go build -o tinct ./cmd/tinct

# Install to $GOPATH/bin
go install ./cmd/tinct
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

### Linting

```bash
# Run golangci-lint
golangci-lint run

# Format code
go fmt ./...

# Run go vet
go vet ./...
```

### Project Structure

```
tinct/
 cmd/tinct/              # Main application entry point
 internal/               # Private application code
    cli/                # CLI commands (Cobra)
    colour/             # Colour extraction algorithms
       extractor.go    # Extractor interface & factory
       kmeans.go       # K-means implementation
       palette.go      # Palette type & conversions
       categorisation.go # Smart categorisation
       ansi.go         # Terminal colour previews
    image/              # Image loading utilities
    version/            # Version info (ldflags)
 testdata/               # Test images and data
 plan.md                 # Development roadmap
```

##  Roadmap

### Phase 1: Foundation  **COMPLETE**
-  CLI framework with Cobra
-  K-means colour extraction (blazingly fast - <5ms!)
-  Multiple output formats (hex, rgb, json, categorised)
-  Colour categorisation with WCAG contrast checking
-  Theme detection (auto, dark, light)
-  Semantic colour assignment (danger, warning, success, etc.)
-  Muted colour variants
-  ANSI terminal colour previews
-  Version management
-  British English spelling throughout

### Phase 2: Enhanced Features (Coming Soon)
- [ ] Additional algorithms (Median Cut, Dominant Colour)
- [ ] Configuration file support (Viper)
- [ ] Advanced output formats (HTML preview, CSS variables)
- [ ] Saturation and brightness adjustments
- [ ] AAA contrast compliance mode
- [ ] Colour harmony analysis

### Phase 3: Plugin System
- [x] Output plugins (Hyprland , Kitty )
- [ ] More output plugins (Alacritty, i3, etc.)
- [ ] Source plugins (wallpaper fetchers)
- [x] External plugin support 

### Phase 4: AI Integration
- [ ] AI wallpaper generation (Stable Diffusion, DALL-E)
- [ ] Custom size support
- [ ] Prompt management

See [plan.md](plan.md) for detailed roadmap and implementation status.

##  Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feat/amazing-feature`)
5. Open a Pull Request

### Commit Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions/changes
- `perf:` - Performance improvements
- `chore:` - Maintenance tasks

##  License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

##  Author

**John Mylchreest**
- Email: jmylchreest@gmail.com
- GitHub: [@jmylchreest](https://github.com/jmylchreest)

##  Acknowledgments

- Inspired by [pywal](https://github.com/dylanaraps/pywal)
- Built with [Cobra](https://github.com/spf13/cobra) CLI framework
- K-means++ algorithm for optimal centroid initialisation
- WCAG 2.0 guidelines for accessibility compliance

##  Performance

**Before optimisation**: 4 minutes+ (reported issue)
**After optimisation**: ~3ms average (**80,000x faster!**)

Performance test on 400×400 image:
```
Run 1: 4ms
Run 2: 3ms
Run 3: 3ms
Average: 3.3ms
```

**Optimisations**:
- Aggressive pixel sampling (max 2000 samples)
- K-means++ initialisation
- Early convergence detection
- Reduced max iterations

---

**Made with  and Go 1.25+**