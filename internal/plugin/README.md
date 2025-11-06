# Internal Plugin System

This directory contains the **built-in plugin system** for Tinct. Built-in plugins are compiled directly into the main `tinct` binary and provide core functionality that all users have access to by default.

## Overview

The plugin system allows Tinct to be extended with:
- **Input plugins** - Extract or generate color palettes from various sources
- **Output plugins** - Apply palettes to applications, generate configs, or control devices

### Built-in vs External Plugins

| Aspect | Built-in (internal/) | External (contrib/) |
|--------|---------------------|---------------------|
| **Distribution** | Compiled into `tinct` binary | Separate downloads |
| **Language** | Go only | Any language |
| **Performance** | Direct function calls (fastest) | Process spawn (~8ms RPC or ~52ms JSON-stdio) |
| **Installation** | Included by default | Via `tinct plugins add` |
| **Location** | `internal/plugin/` | `contrib/plugins/` |
| **Examples** | image, hyprland, kitty | random, wob, templater |

## Directory Structure

```
internal/plugin/
├── README.md                  # This file
├── input/                     # Built-in input plugins
│   ├── image/                 # Extract from images
│   ├── file/                  # Load from files
│   ├── remotejson/            # Fetch from JSON URLs
│   ├── remotecss/             # Extract from CSS
│   └── shared/                # Shared utilities
│       └── regions/           # Ambient region extraction
├── output/                    # Built-in output plugins
│   ├── alacritty/             # Alacritty terminal
│   ├── dunst/                 # Dunst notifications
│   ├── fuzzel/                # Fuzzel launcher
│   ├── hyprland/              # Hyprland WM
│   ├── hyprlock/              # Hyprlock screen locker
│   ├── hyprpaper/             # Hyprpaper wallpaper manager
│   ├── kitty/                 # Kitty terminal
│   ├── neovim/                # Neovim editor
│   ├── swayosd/               # SwayOSD on-screen display
│   ├── waybar/                # Waybar status bar
│   ├── wofi/                  # Wofi launcher
│   ├── zellij/                # Zellij multiplexer
│   ├── common/                # Shared utilities
│   ├── template/              # Template engine
│   └── testing/               # Testing utilities
├── manager/                   # Plugin management
├── executor/                  # Plugin execution (RPC, JSON-stdio)
├── protocol/                  # Plugin protocols and interfaces
└── repository/                # Plugin repository management
```

## Plugin Interfaces

### Input Plugin Interface

Input plugins extract or generate color palettes.

**Required Methods:**
```go
type Plugin interface {
    Name() string                              // Plugin identifier
    Description() string                       // Human-readable description
    Generate(ctx, opts) (*colour.Palette, error)  // Extract/generate colors
    RegisterFlags(cmd *cobra.Command)          // Register CLI flags
    Validate() error                           // Validate configuration
}
```

**Optional Interfaces:**
```go
type WallpaperProvider interface {
    WallpaperPath() string  // Return wallpaper file path
}

type ThemeHinter interface {
    ThemeHint() string  // Suggest "dark", "light", or "auto"
}
```

**See:** [Input Plugins Documentation](input/)

### Output Plugin Interface

Output plugins generate configuration files or perform actions with palettes.

**Required Methods:**
```go
type Plugin interface {
    Name() string                              // Plugin identifier
    Description() string                       // Human-readable description
    Generate(palette) (map[string][]byte, error)  // Generate config files
    RegisterFlags(cmd *cobra.Command)          // Register CLI flags
    Validate() error                           // Validate configuration
    DefaultOutputDir() string                  // Default config directory
}
```

**Optional Interfaces:**
```go
type VerbosePlugin interface {
    SetVerbose(bool)  // Receive verbose flag
}

type PreExecuteHook interface {
    PreExecute(ctx) (skip bool, reason string, err error)  // Pre-generation checks
}

type PostExecuteHook interface {
    PostExecute(ctx, execCtx, files) error  // Post-generation actions
}

type WallpaperContextProvider interface {
    SetWallpaperContext(string)  // Receive wallpaper path
}

type TemplateProvider interface {
    GetEmbeddedFS() interface{}  // Expose embedded templates
}
```

**See:** [Output Plugins Documentation](output/README.md)

## Creating Built-in Plugins

### When to Create a Built-in Plugin

✅ **Create a built-in plugin when:**
- The functionality is essential for most users
- Performance is critical (direct function calls vs process spawn)
- Deep integration with Tinct core is needed
- The application is extremely popular (kitty, hyprland, etc.)

❌ **Create an external plugin when:**
- The functionality is niche or experimental
- The application is less common
- You want to support non-Go languages
- Distribution as a separate download is acceptable

### Adding a New Built-in Output Plugin

1. **Create plugin directory:**
   ```bash
   mkdir -p internal/plugin/output/myapp
   cd internal/plugin/output/myapp
   ```

2. **Create plugin implementation** (`myapp.go`):
   ```go
   package myapp

   import (
       "github.com/jmylchreest/tinct/internal/colour"
       "github.com/spf13/cobra"
   )

   type Plugin struct {
       outputDir string
       verbose   bool
   }

   func New() *Plugin {
       return &Plugin{}
   }

   func (p *Plugin) Name() string {
       return "myapp"
   }

   func (p *Plugin) Description() string {
       return "Generate themes for MyApp"
   }

   func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
       cmd.Flags().StringVar(&p.outputDir, "myapp.output-dir", p.DefaultOutputDir(), "Output directory")
   }

   func (p *Plugin) Validate() error {
       return nil
   }

   func (p *Plugin) DefaultOutputDir() string {
       return "~/.config/myapp"
   }

   func (p *Plugin) SetVerbose(v bool) {
       p.verbose = v
   }

   func (p *Plugin) Generate(palette *colour.CategorisedPalette) (map[string][]byte, error) {
       // Generate config content
       return map[string][]byte{
           "tinct.conf": []byte("# Generated by Tinct"),
       }, nil
   }
   ```

3. **Add embedded template** (optional):
   ```go
   import "embed"

   //go:embed templates/*.tmpl
   var templatesFS embed.FS

   func (p *Plugin) GetEmbeddedFS() interface{} {
       return templatesFS
   }
   ```

4. **Register plugin** in `internal/cli/generate.go`:
   ```go
   import "github.com/jmylchreest/tinct/internal/plugin/output/myapp"

   func init() {
       // ... existing plugins ...
       registerOutputPlugin(myapp.New())
   }
   ```

5. **Add README** (`internal/plugin/output/myapp/README.md`):
   - Describe what the plugin does
   - List configuration options
   - Provide usage examples
   - Document any special requirements

6. **Test the plugin:**
   ```bash
   go build ./cmd/tinct
   ./tinct generate -i image -p wallpaper.jpg -o myapp --verbose
   ```

### Adding a New Built-in Input Plugin

Similar process, but implement the input plugin interface. See existing plugins in `internal/plugin/input/` for examples.

## Plugin Architecture

### Plugin Manager (`manager/`)

Handles plugin discovery, loading, and lifecycle management.

**Key Responsibilities:**
- Register built-in plugins
- Load external plugins
- Manage enabled/disabled state
- Resolve plugin dependencies

### Plugin Executor (`executor/`)

Executes external plugins using different protocols.

**Supported Protocols:**
- **go-plugin RPC** - Process reuse, ~8ms overhead
- **JSON-stdio** - Simple, any language, ~52ms overhead

### Plugin Protocol (`protocol/`)

Defines interfaces and data structures for plugin communication.

**Key Types:**
- `PluginInfo` - Plugin metadata
- `PaletteData` - Color palette with metadata
- `InputOptions` - Input plugin options
- `ExecutionContext` - Execution environment

## Template System

Built-in output plugins use Go's `text/template` engine with custom functions.

**Available in templates:**
- `get . "role"` - Get color by role
- `has . "role"` - Check if role exists
- `hex` - Format as #RRGGBB
- `rgb` - Format as R, G, B
- `rgba` - Format as R, G, B, A
- `hsl` - Format as H, S%, L%
- `withAlpha color alpha` - Set alpha channel
- `themeType .` - Get theme type (dark/light)
- `.WallpaperPath` - Wallpaper path (if available)

**See:** [Template Guide](../../../docs/TEMPLATE_GUIDE.md)

## Testing

### Unit Tests

Each plugin should have unit tests:

```bash
# Test specific plugin
go test ./internal/plugin/output/myapp/...

# Test all plugins
go test ./internal/plugin/...
```

### Integration Tests

Test with actual Tinct commands:

```bash
# Test plugin end-to-end
./tinct generate -i image -p testdata/sample.jpg -o myapp --dry-run --verbose
```

### Test Utilities

Use `internal/plugin/output/testing` for test helpers:

```go
import "github.com/jmylchreest/tinct/internal/plugin/output/testing"

func TestPlugin(t *testing.T) {
    palette := testing.NewTestPalette()
    plugin := New()
    files, err := plugin.Generate(palette)
    // assertions...
}
```

## Documentation

Each plugin directory should have:

✅ **README.md** - Plugin overview and usage  
✅ **Code comments** - Document exported functions  
✅ **Examples** - Sample configurations  
✅ **Templates** - Document template variables (if applicable)

## Contributing

When adding or modifying built-in plugins:

1. **Follow existing patterns** - Match the style of existing plugins
2. **Add comprehensive documentation** - README + code comments
3. **Include tests** - Unit tests for core functionality
4. **Use templates** - Leverage the template system for configs
5. **Handle errors gracefully** - Return meaningful error messages
6. **Support dry-run** - Don't write files in dry-run mode
7. **Implement optional interfaces** - PreExecute, PostExecute, etc. when needed

## Resources

- **[Plugin Development Guide](../../../docs/DEVELOPMENT.md)** - Contributing guidelines
- **[External Plugins](../../../contrib/plugins/README.md)** - External plugin development
- **[Template Guide](../../../docs/TEMPLATE_GUIDE.md)** - Template syntax and functions
- **[Main README](../../../README.md)** - Project overview

## Questions?

- Check existing plugin implementations in this directory
- Read the comprehensive development guide in `docs/DEVELOPMENT.md`
- Open an issue on GitHub for clarification
- See [PLUGINS-WISHLIST.md](../../../docs/PLUGINS-WISHLIST.md) for requested plugins
