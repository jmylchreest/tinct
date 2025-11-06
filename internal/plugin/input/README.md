# Input Plugins

This directory contains **built-in input plugins** that extract or generate color palettes from various sources. Input plugins are the first step in Tinct's pipeline - they produce raw color palettes that are then categorized into semantic roles (background, foreground, accent1, etc.).

## Overview

Input plugins extract colors from sources like:
- **Images** - Wallpapers, photos, artwork
- **Remote APIs** - JSON theme APIs, CSS variable files
- **Files** - Saved palette files
- **Generators** - Algorithmic palette generation

The extracted colors are passed to the categorizer which assigns semantic roles based on luminance, contrast, and color theory.

## Built-in Input Plugins

| Plugin | Description | Source Type | Theme Detection |
|--------|-------------|-------------|-----------------|
| **image** | Extract from images using k-means clustering | Local files, HTTP(S) URLs | ✅ Auto-detects dark/light |
| **file** | Load from saved palette files | JSON, YAML files | ✅ Preserves theme type |
| **remotejson** | Fetch from JSON APIs with JSONPath queries | HTTP(S) URLs | ❌ Uses categorizer |
| **remotecss** | Extract from CSS files (variables, hex codes) | HTTP(S) URLs | ❌ Uses categorizer |

## Directory Structure

```
internal/plugin/input/
├── README.md              # This file
├── plugin.go              # Input plugin interface and registry
├── image/                 # Image extraction plugin
│   ├── image.go           # Main plugin implementation
│   └── seed.go            # K-means seed generation
├── file/                  # File loading plugin
│   ├── file.go            # Load from JSON/YAML
│   └── file_test.go       # Unit tests
├── remotejson/            # Remote JSON fetching plugin
│   └── remotejson.go      # Fetch and parse JSON
├── remotecss/             # Remote CSS extraction plugin
│   └── remotecss.go       # Parse CSS variables/hex codes
└── shared/                # Shared utilities
    └── regions/           # Ambient region extraction
        ├── README.md      # Region extraction docs
        └── regions.go     # Edge/corner color extraction
```

## Plugin Interface

### Required Methods

All input plugins must implement the `Plugin` interface:

```go
type Plugin interface {
    // Name returns the plugin identifier (used in CLI: -i <name>)
    Name() string

    // Description returns human-readable description
    Description() string

    // Version returns plugin version (e.g., "1.0.0")
    Version() string

    // Generate extracts colors from the source
    // Returns raw color palette (categorization happens separately)
    Generate(ctx context.Context, opts GenerateOptions) (*colour.Palette, error)

    // RegisterFlags registers CLI flags (e.g., --image.path)
    RegisterFlags(cmd *cobra.Command)

    // Validate checks required configuration is provided
    Validate() error
}
```

### Optional Interfaces

Input plugins can implement optional interfaces for additional functionality:

#### WallpaperProvider

Provides the source wallpaper path for wallpaper-setting plugins:

```go
type WallpaperProvider interface {
    // WallpaperPath returns path to source wallpaper image
    WallpaperPath() string
}
```

**Implementation:** `image` plugin  
**Used by:** Output plugins like `hyprpaper`, `hyprlock` to set wallpaper

#### ThemeHinter

Suggests theme type (dark/light) to help the categorizer:

```go
type ThemeHinter interface {
    // ThemeHint returns "dark", "light", "auto", or "" for no hint
    ThemeHint() string
}
```

**Implementation:** `image`, `file` plugins  
**Purpose:** Advisory only - categorizer makes final decision

## Plugin Details

### image Plugin

Extracts colors from images using k-means clustering.

**Features:**
- K-means clustering with configurable seed modes
- Supports local files and HTTP(S) URLs
- Optional ambient region extraction (edge/corner colors)
- Theme detection based on dominant luminance
- Wallpaper provider for output plugins

**Seed Modes:**
- `content` (default) - Hash of image pixels (deterministic per image)
- `filepath` - Hash of file path (deterministic per location)
- `manual` - User-provided seed value
- `random` - Non-deterministic random seed

**CLI Flags:**
```bash
--image.path, -p          # Image path or URL (required)
--image.algorithm         # Extraction algorithm (default: kmeans)
--image.colours           # Number of colors to extract (default: 16)
--image.extractAmbience   # Extract edge/corner regions
--image.regions           # Number of regions (4, 8, 12, 16)
--image.seed-mode         # Seed mode: content, filepath, manual, random
--image.seed-value        # Seed value (for manual mode)
```

**Example:**
```bash
tinct generate -i image -p wallpaper.jpg -o hyprland,kitty
```

**See:** [Ambient Region Extraction](shared/regions/README.md)

### file Plugin

Loads color palettes from saved files (JSON, YAML).

**Features:**
- Load from JSON or YAML palette files
- Preserves semantic roles and theme type
- Useful for reusing generated palettes
- Fast - no extraction needed

**File Format:**
```json
{
  "colours": [
    {"hex": "#1e1e2e", "role": "background"},
    {"hex": "#cdd6f4", "role": "foreground"},
    {"hex": "#f38ba8", "role": "accent1"}
  ],
  "theme_type": "dark"
}
```

**CLI Flags:**
```bash
--file.path               # Path to palette file (required)
```

**Example:**
```bash
tinct generate -i file -p saved-palette.json -o kitty
```

### remotejson Plugin

Fetches color palettes from remote JSON APIs.

**Features:**
- HTTP(S) URL support
- JSONPath queries for complex JSON structures
- Color format auto-detection (hex, rgb, hsl)
- Useful for theme repositories (Catppuccin, Dracula, etc.)

**CLI Flags:**
```bash
--remote-json.url         # JSON URL (required)
--remote-json.path        # JSONPath query (optional)
--remote-json.format      # Color format hint: hex, rgb, hsl (optional)
```

**Example:**
```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  -o hyprland,kitty
```

### remotecss Plugin

Extracts colors from remote CSS files.

**Features:**
- Parse CSS custom properties (variables)
- Extract hex color codes
- Useful for loading CSS frameworks or theme files

**CLI Flags:**
```bash
--remote-css.url          # CSS file URL (required)
```

**Example:**
```bash
tinct generate -i remote-css \
  --remote-css.url "https://example.com/theme.css" \
  -o waybar
```

## Creating a New Input Plugin

### Step-by-Step Guide

1. **Create plugin directory:**
   ```bash
   mkdir -p internal/plugin/input/myplugin
   cd internal/plugin/input/myplugin
   ```

2. **Implement the Plugin interface** (`myplugin.go`):
   ```go
   package myplugin

   import (
       "context"
       "fmt"
       "image/color"
       
       "github.com/jmylchreest/tinct/internal/colour"
       "github.com/jmylchreest/tinct/internal/plugin/input"
       "github.com/spf13/cobra"
   )

   type Plugin struct {
       sourcePath string
       verbose    bool
   }

   func New() *Plugin {
       return &Plugin{}
   }

   func (p *Plugin) Name() string {
       return "myplugin"
   }

   func (p *Plugin) Description() string {
       return "Extract colors from my source"
   }

   func (p *Plugin) Version() string {
       return "1.0.0"
   }

   func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
       cmd.Flags().StringVar(&p.sourcePath, "myplugin.path", "", "Source path (required)")
   }

   func (p *Plugin) Validate() error {
       if p.sourcePath == "" {
           return fmt.Errorf("source path is required (use --myplugin.path)")
       }
       return nil
   }

   func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
       p.verbose = opts.Verbose

       if opts.Verbose {
           fmt.Println("Extracting colors from:", p.sourcePath)
       }

       // Extract colors from your source
       colors := []color.Color{
           color.RGBA{R: 30, G: 30, B: 46, A: 255},   // Example colors
           color.RGBA{R: 205, G: 214, B: 244, A: 255},
       }

       return &colour.Palette{
           Colors: colors,
       }, nil
   }
   ```

3. **Implement optional interfaces** (if needed):
   ```go
   // If your plugin provides a wallpaper
   func (p *Plugin) WallpaperPath() string {
       return p.sourcePath
   }

   // If your plugin can detect theme type
   func (p *Plugin) ThemeHint() string {
       // Analyze your source and return "dark" or "light"
       return "dark"
   }
   ```

4. **Add tests** (`myplugin_test.go`):
   ```go
   package myplugin

   import (
       "context"
       "testing"

       "github.com/jmylchreest/tinct/internal/plugin/input"
   )

   func TestGenerate(t *testing.T) {
       p := New()
       p.sourcePath = "test-source"

       opts := input.GenerateOptions{
           Verbose: false,
           DryRun:  false,
       }

       palette, err := p.Generate(context.Background(), opts)
       if err != nil {
           t.Fatalf("Generate failed: %v", err)
       }

       if len(palette.Colors) == 0 {
           t.Error("Expected colors, got empty palette")
       }
   }
   ```

5. **Register the plugin** in `internal/cli/generate.go`:
   ```go
   import "github.com/jmylchreest/tinct/internal/plugin/input/myplugin"

   func init() {
       // ... existing plugins ...
       registerInputPlugin(myplugin.New())
   }
   ```

6. **Add README** (`internal/plugin/input/myplugin/README.md`):
   - Describe what the plugin does
   - Document CLI flags
   - Provide usage examples
   - List any special requirements

7. **Test the plugin:**
   ```bash
   go build ./cmd/tinct
   ./tinct generate -i myplugin --myplugin.path "source" -o hyprland --verbose
   ```

## Generate Options

Input plugins receive `GenerateOptions` during generation:

```go
type GenerateOptions struct {
    // Verbose enables detailed logging to stderr
    Verbose bool

    // DryRun indicates generation without side effects
    DryRun bool

    // ColourOverrides are manual color specifications (role=hex)
    // Example: ["background=#1e1e2e", "accent1=#f38ba8"]
    ColourOverrides []string

    // PluginArgs are custom arguments from --plugin-args
    PluginArgs map[string]any
}
```

**Best Practices:**
- Respect `Verbose` flag - log helpful info to stderr when true
- Respect `DryRun` flag - don't download/cache files or make network requests
- Handle `ColourOverrides` if appropriate (usually done by categorizer)
- Parse `PluginArgs` for plugin-specific configuration

## Color Extraction Best Practices

### Number of Colors

- **Too few** (< 8): May not capture color variety
- **Optimal** (12-24): Good balance for most use cases
- **Too many** (> 32): Diminishing returns, slower categorization

### Color Quality

1. **Remove duplicates** - Colors too similar muddy the palette
2. **Check viability** - Very dark/light colors may not categorize well
3. **Preserve variety** - Include range of hues, saturations, luminances

### Theme Detection

If implementing `ThemeHinter`:
- Calculate average luminance of extracted colors
- If avg luminance < 0.5 → probably dark theme
- If avg luminance > 0.5 → probably light theme
- Return "auto" if unsure (let categorizer decide)

## Testing

### Unit Tests

Test core extraction logic:

```bash
go test ./internal/plugin/input/myplugin/...
```

### Integration Tests

Test with actual Tinct commands:

```bash
# Test basic extraction
./tinct generate -i myplugin --myplugin.path "test" -o hyprland --dry-run

# Test with verbose
./tinct generate -i myplugin --myplugin.path "test" --verbose

# Test with preview
./tinct generate -i myplugin --myplugin.path "test" --preview
```

### Test Coverage

Aim for > 80% coverage:

```bash
go test -cover ./internal/plugin/input/myplugin/...
```

## Shared Utilities

### Ambient Region Extraction

The `shared/regions` package provides edge/corner color extraction for ambient lighting:

```go
import "github.com/jmylchreest/tinct/internal/plugin/input/shared/regions"

// Extract 8 regions (top, topRight, right, etc.)
regionColors := regions.ExtractRegions(img, regions.Config8Regions)
```

**See:** [Region Extraction README](shared/regions/README.md)

## Documentation

Each input plugin should document:

✅ **What it extracts from** - Source type (files, URLs, APIs)  
✅ **CLI flags** - All configuration options  
✅ **Usage examples** - Common use cases  
✅ **Requirements** - Dependencies or prerequisites  
✅ **Limitations** - Known issues or constraints

## Contributing

When adding or modifying input plugins:

1. **Follow the Plugin interface** - Implement all required methods
2. **Handle errors gracefully** - Return meaningful error messages
3. **Add comprehensive tests** - Unit tests for core logic
4. **Document thoroughly** - README + code comments
5. **Support dry-run** - Don't make side effects in dry-run mode
6. **Respect verbose flag** - Log helpful info when requested
7. **Consider optional interfaces** - Implement WallpaperProvider/ThemeHinter if applicable

## Resources

- **[Main Plugin README](../README.md)** - Overall plugin architecture
- **[Output Plugins](../output/README.md)** - Output plugin documentation
- **[Development Guide](../../../../docs/DEVELOPMENT.md)** - Contributing guidelines
- **[Template Guide](../../../../docs/TEMPLATE_GUIDE.md)** - For understanding categorized output

## Questions?

- Check existing plugin implementations in this directory
- Read the comprehensive development guide in `docs/DEVELOPMENT.md`
- Open an issue on GitHub for clarification
- See examples in contrib/plugins/input/ for external plugin patterns
