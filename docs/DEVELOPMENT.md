# Development Guidelines for Tinct

This document provides comprehensive guidelines for developers and LLMs working on the Tinct project. Following these guidelines ensures consistency, maintainability, and quality across the codebase.

## Table of Contents

- [Language and Tooling](#language-and-tooling)
- [Architecture Principles](#architecture-principles)
- [SOLID Principles in Go](#solid-principles-in-go)
- [Code Style and Best Practices](#code-style-and-best-practices)
- [Project Structure](#project-structure)
- [CLI Framework (Cobra/Viper)](#cli-framework-cobraviper)
- [Plugin System](#plugin-system)
- [Hook System](#hook-system)
- [Error Handling](#error-handling)
- [Testing](#testing)
- [Documentation](#documentation)

## Language and Tooling

### Go Version

**Minimum Version: Go 1.25+**

 **[Go 1.25 Language Specification](https://go.dev/ref/spec)**

#### Go 1.25+ Features to Leverage

Go 1.25 builds upon the modern Go foundation with enhanced features:

- **Generics Enhancements**
  - Improved type inference for generic functions
  - Better constraint satisfaction checking
  - Type parameter instantiation improvements

- **Range Over Function Types (Enhanced)**
  - Iterator patterns with `range` over custom functions
  - Cleaner collection iteration without explicit methods
  - Example: `for v := range myFunc { ... }`

- **Standard Library Improvements**
  - `slices` package: Clone, Compact, Sort, Contains, etc.
  - `maps` package: Clone, Copy, DeleteFunc, Equal, etc.
  - `cmp` package: Ordered types and comparison functions
  - Enhanced `errors` package with Join and multi-error handling

- **Performance Optimizations**
  - Profile-Guided Optimization (PGO) for production builds
  - Improved compiler optimizations
  - Better escape analysis and inlining

- **Error Handling**
  - `errors.Join()` for combining multiple errors
  - Better error wrapping and unwrapping with `%w`
  - Error tree traversal with `errors.Is()` and `errors.As()`

- **Context Package Enhancements**
  - Better cancellation propagation
  - Improved timeout and deadline handling
  - Context value type safety improvements

- **Structured Logging** (log/slog)
  - Built-in structured logging support
  - JSON and text formatters
  - Leveled logging (Debug, Info, Warn, Error)
  - Context-aware logging

**Example Usage in Tinct:**

```go
// Use generics for type-safe plugin registry
type Registry[T any] struct {
    items map[string]T
    mu    sync.RWMutex
}

// Use slices package for operations
import "slices"
colors = slices.Clone(originalColors)
slices.SortFunc(colors, compareColors)

// Use structured logging
import "log/slog"
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("extracting colors", "algorithm", "kmeans", "count", 16)

// Use range over functions for custom iterators
for color := range palette.Colors() {
    // process color
}

// Use errors.Join for multiple errors
if err := errors.Join(err1, err2, err3); err != nil {
    return fmt.Errorf("multiple errors occurred: %w", err)
}
```

**Note:** Always check the [official Go 1.25 release notes](https://go.dev/doc/go1.25) for the complete list of features and changes.

#### Quick Reference: Go 1.25 Key Features

**1. Generics Type Inference Improvements**
```go
// Better inference for nested generic calls
func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// Type parameters inferred from usage
numbers := []int{1, 2, 3}
strings := Map(numbers, strconv.Itoa) // T and U inferred
```

**2. Range Over Function Types**
```go
// Custom iterators without explicit interfaces
func (p *Palette) Colors() func(func(color.Color) bool) {
    return func(yield func(color.Color) bool) {
        for _, c := range p.colors {
            if !yield(c) {
                return
            }
        }
    }
}

// Usage with range
for color := range palette.Colors() {
    process(color)
}
```

**3. Standard Library Enhancements**
```go
// slices package
import "slices"
sorted := slices.Clone(colors)
slices.SortFunc(sorted, compareColors)
unique := slices.Compact(sorted)

// maps package
import "maps"
copy := maps.Clone(original)
maps.DeleteFunc(copy, func(k string, v int) bool { return v == 0 })

// cmp package
import "cmp"
result := cmp.Or(value1, value2, defaultValue) // First non-zero
```

**4. Enhanced Error Handling**
```go
// Join multiple errors
err := errors.Join(
    validateInput(),
    checkPermissions(),
    processData(),
)

// Unwrap multiple errors
for _, e := range err.(interface{ Unwrap() []error }).Unwrap() {
    log.Error("error", "detail", e)
}
```

**5. Structured Logging (log/slog)**
```go
import "log/slog"

// Create logger
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Structured logging with context
logger.Info("extracting colors",
    "algorithm", "kmeans",
    "count", 16,
    "duration", time.Since(start))

logger.Error("extraction failed",
    "error", err,
    "input", inputPath)

// Context-aware logging
ctx := context.WithValue(ctx, "request_id", uuid.New())
logger.InfoContext(ctx, "processing request")
```

**6. Performance Features**
```go
// Profile-Guided Optimization (PGO)
// 1. Build with default profile
go build -o tinct ./cmd/tinct

// 2. Collect profile
./tinct generate image.jpg -cpuprofile=default.pgo

// 3. Rebuild with PGO
go build -pgo=default.pgo -o tinct ./cmd/tinct
// Results in 5-15% performance improvement
```

**Resources:**
-  [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
-  [Go Language Specification](https://go.dev/ref/spec)
-  [Effective Go](https://go.dev/doc/effective_go)
-  [Go Blog](https://go.dev/blog/)

### CLI Framework
- **Cobra**: Command-line interface structure and commands
- **Viper**: Configuration management (files, environment variables, flags)
- Use Cobra's built-in flag binding with Viper for seamless config integration

### Development Tools
- `go fmt` / `gofmt` for formatting
- `golangci-lint` for comprehensive linting
- `go vet` for static analysis
- `staticcheck` for additional checks
- `go mod` for dependency management
- `goreleaser` for automated releases and distribution

## Architecture Principles

### Core Design Philosophy

Tinct follows a plugin-based architecture with the following components:

1. **Core Engine**: Colour extraction and palette generation
2. **Source Plugin System**: Extensible wallpaper generation and fetching
3. **Output Plugin System**: Extensible configuration file generators
4. **Hook System**: Pre/post execution lifecycle management
5. **Configuration**: Centralized config management via Viper

### High-Level Flow

```
Source Plugins (optional) -> Input (Image/Colors) -> Color Extraction -> Palette Generation -> Output Plugins -> Hooks
      |                                                    |
  (AI Gen, Fetch)                                   Configuration
```

**Flow Variations:**

1. **Direct Image Input**: `Image -> Extract -> Generate -> Output Plugins`
2. **AI Wallpaper Generation**: `AI Source Plugin -> Generate Image -> Extract -> Generate -> Output Plugins`
3. **Wallpaper Fetching**: `Fetch Source Plugin -> Download -> Extract -> Generate -> Output Plugins`
4. **Color Codes**: `Color Codes -> Generate -> Output Plugins`

### Key Architectural Goals

- **Extensibility**: Easy to add new colour extraction algorithms and output plugins
- **Composability**: Plugins work independently but can be chained
- **Testability**: Each component is independently testable
- **Performance**: Efficient color processing, concurrent plugin execution where safe
- **User Experience**: Clear CLI interface, helpful error messages, sensible defaults

## SOLID Principles in Go

### Single Responsibility Principle (SRP)

Each package, type, and function should have one clear purpose.

**Good Example:**
```go
// Good: Each type has a single responsibility
type ColorExtractor struct {
    algorithm Algorithm
}

type PaletteGenerator struct {
    extractor *ColorExtractor
    config    GeneratorConfig
}

type PluginExecutor struct {
    plugins []Plugin
    hooks   *HookManager
}
```

**Bad Example:**
```go
// Bad: God object doing too much
type Tinct struct {
    // handles extraction, generation, plugin execution, hooks, config, etc.
}
```

### Open/Closed Principle (OCP)

Design for extension without modification. Use interfaces extensively.

**Good Example:**
```go
// Open for extension via interface implementation
type ColorExtractor interface {
    Extract(ctx context.Context, input io.Reader) ([]color.Color, error)
}

type KMeansExtractor struct{}
type MedianCutExtractor struct{}
type DominantColorExtractor struct{}

// All implement ColorExtractor without modifying existing code
```

### Liskov Substitution Principle (LSP)

Subtypes must be substitutable for their base types. In Go, this means interface implementations must honor contracts.

**Good Example:**
```go
type Plugin interface {
    // Name returns the unique plugin identifier
    Name() string
    
    // Execute generates output based on the palette
    // Must be idempotent - multiple calls with same palette produce same result
    Execute(ctx context.Context, palette *Palette) error
}

// All implementations must respect the idempotency contract
```

### Interface Segregation Principle (ISP)

Clients shouldn't depend on interfaces they don't use. Keep interfaces small and focused.

**Good Example:**
```go
// Small, focused interfaces
type PluginValidator interface {
    Validate() error
}

type PluginExecutor interface {
    Execute(ctx context.Context, palette *Palette) error
}

type PluginCleaner interface {
    Cleanup() error
}

// Plugins implement only what they need
```

**Bad Example:**
```go
// Bad: Fat interface forcing unnecessary implementations
type Plugin interface {
    Execute(ctx context.Context, palette *Palette) error
    Validate() error
    Cleanup() error
    GetMetadata() Metadata
    GetDependencies() []string
    RequiresRoot() bool
    // ... many more methods
}
```

### Dependency Inversion Principle (DIP)

Depend on abstractions, not concretions. Use interfaces for dependencies.

**Good Example:**
```go
// High-level module depends on abstraction
type Application struct {
    extractor ColorExtractor      // interface
    generator PaletteGenerator    // interface
    executor  PluginExecutor      // interface
    hooks     HookManager         // interface
}

// Low-level implementations are injected
func NewApplication(
    extractor ColorExtractor,
    generator PaletteGenerator,
    executor PluginExecutor,
    hooks HookManager,
) *Application {
    return &Application{
        extractor: extractor,
        generator: generator,
        executor:  executor,
        hooks:     hooks,
    }
}
```

## Code Style and Best Practices

### Naming Conventions

- **Packages**: Short, lowercase, no underscores (e.g., `palette`, `plugin`, `extract`)
- **Interfaces**: Descriptive names, often end in -er for single-method (e.g., `Extractor`, `Generator`)
- **Types**: PascalCase (e.g., `ColorPalette`, `PluginConfig`)
- **Functions/Methods**: PascalCase for exported, camelCase for unexported
- **Variables**: Descriptive camelCase; avoid single letters except in tight scopes (loop iterators)
- **Constants**: PascalCase or ALL_CAPS for package-level constants

### Idiomatic Go Patterns

#### Accept Interfaces, Return Structs
```go
// Good
func ProcessPalette(p Palette) *Result { // concrete return
    // ...
}

func NewExecutor(plugins []Plugin) *Executor { // interface parameter
    // ...
}
```

#### Use Context for Cancellation and Timeouts
```go
func (e *Extractor) Extract(ctx context.Context, img image.Image) ([]color.Color, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // proceed with extraction
    }
    // ...
}
```

#### Error Handling
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to extract colors: %w", err)
}

// Use errors.Is and errors.As for checking
if errors.Is(err, ErrInvalidFormat) {
    // handle specific error
}

// Join multiple errors (Go 1.20+)
return errors.Join(err1, err2, err3)
```

#### Defer for Cleanup
```go
func (p *Plugin) Execute(ctx context.Context, palette *Palette) error {
    f, err := os.Create(p.outputPath)
    if err != nil {
        return err
    }
    defer f.Close() // ensures cleanup
    
    // work with file
}
```

#### Use Table-Driven Tests
```go
func TestColorExtraction(t *testing.T) {
    tests := []struct {
        name    string
        input   image.Image
        want    int
        wantErr bool
    }{
        {"8 colors", testImage1, 8, false},
        {"16 colors", testImage2, 16, false},
        {"invalid", nil, 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

#### Use Functional Options for Complex Constructors
```go
type ExtractorOption func(*Extractor)

func WithAlgorithm(algo Algorithm) ExtractorOption {
    return func(e *Extractor) {
        e.algorithm = algo
    }
}

func WithColorCount(count int) ExtractorOption {
    return func(e *Extractor) {
        e.colorCount = count
    }
}

func NewExtractor(opts ...ExtractorOption) *Extractor {
    e := &Extractor{
        algorithm:  defaultAlgorithm,
        colorCount: 16,
    }
    for _, opt := range opts {
        opt(e)
    }
    return e
}
```

### Modern Go Features Usage

#### Generics for Type Safety
```go
// Use generics for plugin registry
type Registry[T any] struct {
    items map[string]T
    mu    sync.RWMutex
}

func (r *Registry[T]) Register(name string, item T) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.items[name] = item
}

func (r *Registry[T]) Get(name string) (T, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    item, ok := r.items[name]
    return item, ok
}
```

#### Slices Package for Collection Operations
```go
import "slices"

// Use slices package for common operations
colors = slices.Clone(originalColors)
slices.Sort(colors)
unique := slices.Compact(slices.Clone(colors))

if slices.Contains(supportedFormats, format) {
    // ...
}
```

## Project Structure

```
tinct/
 cmd/
    tinct/
        main.go              # Entry point
 internal/
    app/
       app.go              # Application orchestration
    color/
       extractor.go        # Colour extraction interfaces
       kmeans.go           # K-means implementation
       mediancut.go        # Median cut implementation
       palette.go          # Palette types and operations
    config/
       config.go           # Configuration types and loading
    plugin/
       plugin.go           # Plugin interface and registry
       executor.go         # Plugin execution logic
       builtin/            # Built-in plugins
           alacritty.go
           kitty.go
           ...
    hook/
       hook.go             # Hook interface and manager
       types.go            # Hook types (pre/post)
    image/
        loader.go           # Image loading
        processor.go        # Image preprocessing
 pkg/
    palette/                # Public palette API
        palette.go
 plugins/                     # External plugin examples
    example/
 configs/
    tinct.yaml              # Example config
 docs/
    plugins.md              # Plugin development guide
    hooks.md                # Hook system guide
    algorithms.md           # Colour extraction algorithms
 go.mod
 go.sum
 README.md
 DEVELOPMENT.md              # This file
 LICENSE
 .golangci.yml               # Linter configuration
```

### Package Organization

- **`cmd/`**: Application entry points (main packages)
- **`internal/`**: Private application code, not importable by other projects
- **`pkg/`**: Public library code, safe for external use
- **`configs/`**: Configuration file examples
- **`docs/`**: Additional documentation
- **`plugins/`**: Example external plugins

## CLI Framework (Cobra/Viper)

### Cobra Command Structure

```go
// cmd/tinct/main.go
var rootCmd = &cobra.Command{
    Use:   "tinct",
    Short: "Generate colour palettes from images or colours",
    Long:  `Tinct extracts colour palettes and applies them via plugins.`,
}

var generateCmd = &cobra.Command{
    Use:   "generate [image]",
    Short: "Generate a palette from an image",
    Args:  cobra.MaximumNArgs(1),
    RunE:  runGenerate,
}

func init() {
    cobra.OnInitialize(initConfig)
    
    // Global flags
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    rootCmd.PersistentFlags().IntP("colors", "c", 16, "number of colors to extract")
    
    // Command flags
    generateCmd.Flags().StringP("algorithm", "a", "kmeans", "extraction algorithm")
    generateCmd.Flags().StringSliceP("plugins", "p", []string{}, "plugins to execute")
    
    // Bind flags to viper
    viper.BindPFlag("colors", rootCmd.PersistentFlags().Lookup("colors"))
    viper.BindPFlag("algorithm", generateCmd.Flags().Lookup("algorithm"))
    
    rootCmd.AddCommand(generateCmd)
}
```

### Viper Configuration

```go
func initConfig() {
    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        home, err := os.UserHomeDir()
        cobra.CheckErr(err)
        
        viper.AddConfigPath(home)
        viper.AddConfigPath(".")
        viper.SetConfigName(".tinct")
        viper.SetConfigType("yaml")
    }
    
    viper.SetEnvPrefix("TINCT")
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err == nil {
        fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
    }
}
```

### Configuration Precedence

1. Explicit CLI flags (highest priority)
2. Environment variables (prefixed with TINCT_)
3. Configuration file
4. Default values (lowest priority)

## Plugin System

Tinct supports **two types of plugins** for maximum flexibility:

1. **Source Plugins**: Generate or fetch wallpapers (AI generation, web fetching, local selection)
2. **Output Plugins**: Generate application configuration files from palettes

### Plugin Architecture

Plugins can be implemented in two ways:

1. **Built-in Plugins**: Compiled directly into the Tinct binary
2. **External Plugins**: Standalone executables discovered and loaded at runtime

### External Plugin System

For external plugins, Tinct uses the **HashiCorp go-plugin** architecture:

- **Cross-platform**: Works on Windows, Linux, macOS, FreeBSD
- **Language-agnostic**: Plugins can be written in any language
- **Process isolation**: Plugin crashes don't affect the main application
- **Versioning**: Protocol version negotiation between plugin and host
- **Security**: Plugins run in separate processes with limited access

**External Plugin Discovery:**
```
~/.config/tinct/plugins/          # User plugins
/usr/local/share/tinct/plugins/   # System-wide plugins
./plugins/                        # Local development plugins
```

**Plugin Naming Convention:**
- Source plugins: `tinct-source-<name>`
- Output plugins: `tinct-output-<name>`

**Example:**
```
~/.config/tinct/plugins/
 tinct-source-myai          # Custom AI provider
 tinct-source-wallhaven     # Wallhaven API fetcher
 tinct-output-hyprland      # Hyprland compositor
 tinct-output-waybar        # Waybar status bar
```

### External Plugin Implementation

**Recommended Approach: HashiCorp go-plugin**

Tinct uses [HashiCorp's go-plugin](https://github.com/hashicorp/go-plugin) for external plugins:

```bash
go get github.com/hashicorp/go-plugin
```

**Plugin Protocol Definition** (`internal/plugin/protocol.go`):

```go
package plugin

import (
    "context"
    "github.com/hashicorp/go-plugin"
    "net/rpc"
)

// ProtocolVersion is the version of the plugin protocol
const ProtocolVersion = 1

// Handshake is used to verify plugin compatibility
var Handshake = plugin.HandshakeConfig{
    ProtocolVersion:  ProtocolVersion,
    MagicCookieKey:   "TINCT_PLUGIN",
    MagicCookieValue: "tinct_color_palette",
}

// PluginMap maps plugin types to their RPC implementations
var PluginMap = map[string]plugin.Plugin{
    "source": &SourcePluginRPC{},
    "output": &OutputPluginRPC{},
}

// SourcePluginRPC is the RPC implementation for source plugins
type SourcePluginRPC struct {
    Impl SourcePlugin
}

func (p *SourcePluginRPC) Server(*plugin.MuxBroker) (interface{}, error) {
    return &SourcePluginRPCServer{Impl: p.Impl}, nil
}

func (p *SourcePluginRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
    return &SourcePluginRPCClient{client: c}, nil
}

// SourcePluginRPCServer is the RPC server
type SourcePluginRPCServer struct {
    Impl SourcePlugin
}

func (s *SourcePluginRPCServer) Name(args interface{}, resp *string) error {
    *resp = s.Impl.Name()
    return nil
}

func (s *SourcePluginRPCServer) Generate(args *SourceParams, resp *[]byte) error {
    img, err := s.Impl.Generate(context.Background(), *args)
    if err != nil {
        return err
    }
    // Encode image to bytes
    *resp = encodeImage(img)
    return nil
}

// SourcePluginRPCClient is the RPC client
type SourcePluginRPCClient struct {
    client *rpc.Client
}

func (c *SourcePluginRPCClient) Name() string {
    var resp string
    c.client.Call("Plugin.Name", new(interface{}), &resp)
    return resp
}

func (c *SourcePluginRPCClient) Generate(ctx context.Context, params SourceParams) (image.Image, error) {
    var resp []byte
    err := c.client.Call("Plugin.Generate", params, &resp)
    if err != nil {
        return nil, err
    }
    return decodeImage(resp)
}

// Similar implementations for OutputPluginRPC...
```

**Plugin Discovery** (`internal/plugin/discovery.go`):

```go
package plugin

import (
    "os"
    "path/filepath"
    "strings"
)

// PluginDirs returns directories to search for plugins
func PluginDirs() []string {
    home, _ := os.UserHomeDir()
    return []string{
        filepath.Join(home, ".config", "tinct", "plugins"),
        "/usr/local/share/tinct/plugins",
        "/usr/share/tinct/plugins",
        "./plugins",
    }
}

// DiscoverPlugins finds all available external plugins
func DiscoverPlugins() ([]string, error) {
    var plugins []string
    
    for _, dir := range PluginDirs() {
        entries, err := os.ReadDir(dir)
        if err != nil {
            continue // Directory doesn't exist or not readable
        }
        
        for _, entry := range entries {
            if entry.IsDir() {
                continue
            }
            
            name := entry.Name()
            // Match tinct-source-* or tinct-output-*
            if strings.HasPrefix(name, "tinct-source-") ||
               strings.HasPrefix(name, "tinct-output-") {
                plugins = append(plugins, filepath.Join(dir, name))
            }
        }
    }
    
    return plugins, nil
}

// LoadExternalPlugin loads a plugin from a path
func LoadExternalPlugin(path string) (*plugin.Client, error) {
    client := plugin.NewClient(&plugin.ClientConfig{
        HandshakeConfig: Handshake,
        Plugins:         PluginMap,
        Cmd:             exec.Command(path),
        AllowedProtocols: []plugin.Protocol{
            plugin.ProtocolNetRPC,
        },
    })
    
    rpcClient, err := client.Client()
    if err != nil {
        return nil, fmt.Errorf("failed to create RPC client: %w", err)
    }
    
    return client, nil
}
```

**External Plugin Example (Go)** (`examples/plugins/wallhaven/main.go`):

```go
package main

import (
    "context"
    "image"
    "log"
    
    "github.com/hashicorp/go-plugin"
    "github.com/jmylchreest/tinct/internal/plugin"
)

// WallhavenPlugin fetches wallpapers from Wallhaven
type WallhavenPlugin struct {
    apiKey string
}

func (p *WallhavenPlugin) Name() string {
    return "wallhaven"
}

func (p *WallhavenPlugin) Generate(ctx context.Context, params plugin.SourceParams) (image.Image, error) {
    // Fetch from Wallhaven API
    width, height := params.ResolveDimensions()
    
    // Build API request
    url := fmt.Sprintf("https://wallhaven.cc/api/v1/search?resolution=%dx%d", width, height)
    if len(params.Tags) > 0 {
        url += "&q=" + strings.Join(params.Tags, "+")
    }
    
    // Fetch and decode image
    // ... implementation ...
    
    return img, nil
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: plugin.Handshake,
        Plugins: map[string]plugin.Plugin{
            "source": &plugin.SourcePluginRPC{
                Impl: &WallhavenPlugin{
                    apiKey: os.Getenv("WALLHAVEN_API_KEY"),
                },
            },
        },
    })
}
```

**Build External Plugin:**
```bash
# Build the plugin
cd examples/plugins/wallhaven
go build -o tinct-source-wallhaven

# Install to user plugins directory
mkdir -p ~/.config/tinct/plugins
cp tinct-source-wallhaven ~/.config/tinct/plugins/

# Make executable
chmod +x ~/.config/tinct/plugins/tinct-source-wallhaven
```

**External Plugin Example (Python)** (`examples/plugins/reddit/tinct-source-reddit`):

```python
#!/usr/bin/env python3
"""
External Tinct source plugin for Reddit wallpapers
Uses HashiCorp go-plugin protocol over RPC
"""

import sys
import json
from xmlrpc.server import SimpleXMLRPCServer

class RedditWallpaperPlugin:
    def name(self):
        return "reddit"
    
    def generate(self, params):
        """Generate/fetch wallpaper from Reddit"""
        import praw
        import requests
        from PIL import Image
        from io import BytesIO
        
        # Parse params
        width = params.get('width', 3840)
        height = params.get('height', 2160)
        subreddit = params.get('extra', {}).get('subreddit', 'wallpapers')
        
        # Fetch from Reddit
        reddit = praw.Reddit(client_id=os.environ['REDDIT_CLIENT_ID'],
                            client_secret=os.environ['REDDIT_SECRET'],
                            user_agent='tinct')
        
        # Get top wallpaper
        for post in reddit.subreddit(subreddit).hot(limit=10):
            if post.url.endswith(('.jpg', '.png')):
                resp = requests.get(post.url)
                img = Image.open(BytesIO(resp.content))
                
                # Resize if needed
                if img.size != (width, height):
                    img = img.resize((width, height), Image.LANCZOS)
                
                # Encode to bytes
                buf = BytesIO()
                img.save(buf, format='PNG')
                return buf.getvalue()
        
        raise Exception("No suitable wallpaper found")

if __name__ == '__main__':
    # RPC server setup for go-plugin protocol
    server = SimpleXMLRPCServer(('localhost', 0))
    server.register_instance(RedditWallpaperPlugin())
    
    # Handshake
    print(f"1|1|tcp|localhost:{server.server_address[1]}|grpc")
    sys.stdout.flush()
    
    server.serve_forever()
```

**Using External Plugins:**

```bash
# List discovered plugins
tinct plugins list

# Show plugin info
tinct plugins info wallhaven

# Use external plugin
tinct generate --source wallhaven --tags "nature,mountains"

# Use with configuration
tinct generate --source reddit --subreddit wallpapers --size 4k
```

**Plugin Configuration** (`~/.config/tinct/config.yaml`):

```yaml
plugins:
  # External plugin discovery
  discovery:
    enabled: true
    dirs:
      - ~/.config/tinct/plugins
      - /usr/local/share/tinct/plugins
    
  # Plugin-specific configuration
  wallhaven:
    api_key_env: WALLHAVEN_API_KEY
    default_category: general
    purity: sfw
  
  reddit:
    client_id_env: REDDIT_CLIENT_ID
    secret_env: REDDIT_SECRET
    default_subreddit: wallpapers
```

**Security Considerations:**

1. **Plugin Verification**: 
   ```go
   // Verify plugin signature (optional)
   func VerifyPluginSignature(path string) error {
       // Check file hash against known signatures
       // Use SHA256 checksums stored in trusted location
   }
   ```

2. **Sandboxing**:
   - Plugins run in separate processes (go-plugin provides this)
   - Set resource limits (CPU, memory, network)
   - Use AppArmor/SELinux profiles for Linux

3. **Permissions**:
   ```yaml
   plugins:
     security:
       allowed_network: true        # Allow network access
       allowed_filesystem: limited  # Limited filesystem access
       timeout: 60s                 # Maximum execution time
   ```

**Alternative: Simple Executable Protocol**

For simpler plugins without go-plugin dependency:

```bash
#!/bin/bash
# tinct-source-simple
# Simple external plugin using JSON protocol

# Read params from stdin
params=$(cat)

# Extract parameters
prompt=$(echo "$params" | jq -r '.prompt')
width=$(echo "$params" | jq -r '.width // 3840')
height=$(echo "$params" | jq -r '.height // 2160')

# Generate/fetch wallpaper
# Output image to stdout as base64
curl -s "https://picsum.photos/${width}/${height}" | base64

# Exit 0 on success
exit 0
```

**Usage:**
```bash
echo '{"width":3840,"height":2160}' | ~/.config/tinct/plugins/tinct-source-simple
```

### Source Plugin Interface

Source plugins provide wallpapers that will be used for palette generation.

```go
package plugin

import (
    "context"
    "image"
)

// SourcePlugin generates or fetches wallpapers
type SourcePlugin interface {
    // Name returns the unique plugin identifier
    Name() string
    
    // Generate produces a wallpaper based on input parameters
    // params can include prompts, URLs, search terms, etc.
    Generate(ctx context.Context, params SourceParams) (image.Image, error)
}

// SourceParams provides input for source plugins
type SourceParams struct {
    // Prompt for AI generation (e.g., "cyberpunk city at night")
    Prompt string
    
    // URL for fetching (e.g., "https://unsplash.com/random")
    URL string
    
    // Tags for search/filtering
    Tags []string
    
    // Dimensions (default: 3840x2160 - 4K)
    Width  int
    Height int
    
    // Size preset (e.g., "4k", "ultrawide", "dual", "triple")
    // Overrides Width/Height if set
    SizePreset string
    
    // Additional plugin-specific parameters
    Extra map[string]any
}

// Common size presets for wallpaper generation
const (
    SizePreset4K         = "4k"          // 3840x2160
    SizePreset1080p      = "1080p"       // 1920x1080
    SizePreset1440p      = "1440p"       // 2560x1440
    SizePresetUltrawide  = "ultrawide"   // 3440x1440 (21:9)
    SizePresetSuperwide  = "superwide"   // 5120x1440 (32:9)
    SizePresetDual       = "dual"        // 3840x1080 (2x 1920x1080)
    SizePresetTriple     = "triple"      // 5760x1080 (3x 1920x1080)
    SizePresetDual4K     = "dual-4k"     // 7680x2160 (2x 3840x2160)
)

// ResolveDimensions returns the actual width and height based on preset or explicit values
func (p *SourceParams) ResolveDimensions() (width, height int) {
    if p.SizePreset != "" {
        switch p.SizePreset {
        case SizePreset4K:
            return 3840, 2160
        case SizePreset1080p:
            return 1920, 1080
        case SizePreset1440p:
            return 2560, 1440
        case SizePresetUltrawide:
            return 3440, 1440
        case SizePresetSuperwide:
            return 5120, 1440
        case SizePresetDual:
            return 3840, 1080
        case SizePresetTriple:
            return 5760, 1080
        case SizePresetDual4K:
            return 7680, 2160
        }
    }
    
    // Use explicit dimensions or defaults
    width = p.Width
    height = p.Height
    
    // Default to 4K if not specified
    if width == 0 {
        width = 3840
    }
    if height == 0 {
        height = 2160
    }
    
    return width, height
}

// SourceType indicates the source plugin category
type SourceType int

const (
    SourceTypeAI SourceType = iota      // AI image generation
    SourceTypeFetch                      // Fetch from URL/API
    SourceTypeLocal                      // Local file selection
    SourceTypeRandom                     // Random selection
)

// SourceMetadata provides source plugin information
type SourceMetadata struct {
    Name        string
    Description string
    Type        SourceType
    RequiresAPI bool        // Does it need an API key?
    APIKeyEnv   string      // Environment variable for API key
}

// SourceMetadataProvider is an optional interface for source plugin metadata
type SourceMetadataProvider interface {
    SourceMetadata() SourceMetadata
}
```

### Output Plugin Interface

Output plugins generate configuration files from colour palettes.

```go
package plugin

import "context"

// OutputPlugin represents a palette output plugin
type OutputPlugin interface {
    // Name returns the unique plugin identifier
    Name() string
    
    // Execute generates output based on the palette
    Execute(ctx context.Context, palette *color.Palette) error
}

// Validator is an optional interface for plugins that need validation
type Validator interface {
    Validate() error
}

// Cleaner is an optional interface for cleanup operations
type Cleaner interface {
    Cleanup() error
}

// Metadata provides plugin information
type Metadata struct {
    Name        string
    Description string
    Version     string
    Author      string
}

// MetadataProvider is an optional interface for plugin metadata
type MetadataProvider interface {
    Metadata() Metadata
}
```

### Plugin Registration

```go
// Separate registries for different plugin types
var (
    sourceRegistry = NewRegistry[SourcePlugin]()
    outputRegistry = NewRegistry[OutputPlugin]()
)

// Source Plugin Registration
func RegisterSource(name string, p SourcePlugin) {
    sourceRegistry.Register(name, p)
}

func GetSource(name string) (SourcePlugin, bool) {
    return sourceRegistry.Get(name)
}

func ListSources() []string {
    return sourceRegistry.List()
}

// Output Plugin Registration
func RegisterOutput(name string, p OutputPlugin) {
    outputRegistry.Register(name, p)
}

func GetOutput(name string) (OutputPlugin, bool) {
    return outputRegistry.Get(name)
}

func ListOutputs() []string {
    return outputRegistry.List()
}
```

### Built-in Output Plugin Example

```go
package builtin

type AlacrittyPlugin struct {
    outputPath string
    backup     bool
}

func NewAlacrittyPlugin(outputPath string) *AlacrittyPlugin {
    return &AlacrittyPlugin{
        outputPath: outputPath,
        backup:     true,
    }
}

func (p *AlacrittyPlugin) Name() string {
    return "alacritty"
}

func (p *AlacrittyPlugin) Execute(ctx context.Context, palette *color.Palette) error {
    // Generate alacritty.yml color configuration
    // ...
    return nil
}

func (p *AlacrittyPlugin) Metadata() plugin.Metadata {
    return plugin.Metadata{
        Name:        "alacritty",
        Description: "Generate Alacritty terminal color scheme",
        Version:     "1.0.0",
    }
}

func init() {
    plugin.RegisterOutput("alacritty", NewAlacrittyPlugin("~/.config/alacritty/colors.yml"))
}
```

### Built-in Source Plugin Examples

#### AI Wallpaper Generator

```go
package builtin

import (
    "context"
    "image"
    "log/slog"
)

// AIWallpaperPlugin generates wallpapers using AI (e.g., DALL-E, Stable Diffusion)
type AIWallpaperPlugin struct {
    apiKey     string
    provider   string // "dalle", "stablediffusion", "midjourney"
    logger     *slog.Logger
}

func NewAIWallpaperPlugin(provider, apiKey string) *AIWallpaperPlugin {
    return &AIWallpaperPlugin{
        apiKey:   apiKey,
        provider: provider,
        logger:   slog.Default(),
    }
}

func (p *AIWallpaperPlugin) Name() string {
    return "ai-wallpaper"
}

func (p *AIWallpaperPlugin) Generate(ctx context.Context, params plugin.SourceParams) (image.Image, error) {
    width, height := params.ResolveDimensions()
    
    p.logger.Info("generating AI wallpaper",
        "provider", p.provider,
        "prompt", params.Prompt,
        "dimensions", fmt.Sprintf("%dx%d", width, height),
        "preset", params.SizePreset)
    
    // Call AI API based on provider
    switch p.provider {
    case "dalle":
        return p.generateWithDALLE(ctx, params)
    case "stablediffusion":
        return p.generateWithStableDiffusion(ctx, params)
    default:
        return nil, fmt.Errorf("unsupported AI provider: %s", p.provider)
    }
}

func (p *AIWallpaperPlugin) SourceMetadata() plugin.SourceMetadata {
    return plugin.SourceMetadata{
        Name:        "ai-wallpaper",
        Description: "Generate wallpapers using AI (DALL-E, Stable Diffusion, etc.)",
        Type:        plugin.SourceTypeAI,
        RequiresAPI: true,
        APIKeyEnv:   "TINCT_AI_API_KEY",
    }
}

func init() {
    apiKey := os.Getenv("TINCT_AI_API_KEY")
    provider := os.Getenv("TINCT_AI_PROVIDER")
    if provider == "" {
        provider = "stablediffusion"
    }
    plugin.RegisterSource("ai-wallpaper", NewAIWallpaperPlugin(provider, apiKey))
}
```

#### Random Wallpaper Fetcher

```go
package builtin

// RandomWallpaperPlugin fetches random wallpapers from online sources
type RandomWallpaperPlugin struct {
    sources []string // URLs or API endpoints
    logger  *slog.Logger
}

func NewRandomWallpaperPlugin(sources []string) *RandomWallpaperPlugin {
    return &RandomWallpaperPlugin{
        sources: sources,
        logger:  slog.Default(),
    }
}

func (p *RandomWallpaperPlugin) Name() string {
    return "random-fetch"
}

func (p *RandomWallpaperPlugin) Generate(ctx context.Context, params plugin.SourceParams) (image.Image, error) {
    width, height := params.ResolveDimensions()
    
    p.logger.Info("fetching random wallpaper", 
        "sources", p.sources,
        "dimensions", fmt.Sprintf("%dx%d", width, height))
    
    // Use provided URL or pick random source
    url := params.URL
    if url == "" {
        url = p.sources[rand.Intn(len(p.sources))]
    }
    
    // Append dimensions to URL if it supports them
    url = fmt.Sprintf("%s/%dx%d", url, width, height)
    
    // Fetch and decode image
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch wallpaper: %w", err)
    }
    defer resp.Body.Close()
    
    img, _, err := image.Decode(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to decode image: %w", err)
    }
    
    return img, nil
}

func (p *RandomWallpaperPlugin) SourceMetadata() plugin.SourceMetadata {
    return plugin.SourceMetadata{
        Name:        "random-fetch",
        Description: "Fetch random wallpapers from Unsplash, Pexels, etc.",
        Type:        plugin.SourceTypeFetch,
        RequiresAPI: false,
    }
}

func init() {
    defaultSources := []string{
        "https://source.unsplash.com/random/2560x1440",
        "https://picsum.photos/2560/1440",
    }
    plugin.RegisterSource("random-fetch", NewRandomWallpaperPlugin(defaultSources))
}
```

### Plugin Protocol Version

**Current Protocol Version: 1.0**

Tinct implements a plugin protocol versioning system to ensure compatibility between the main binary and external plugins. This prevents runtime errors from incompatible plugins built against different API versions.

#### Version Format

Protocol versions follow **MAJOR.MINOR** format:
- **MAJOR version**: Incremented for breaking/incompatible API changes
- **MINOR version**: Incremented for backward-compatible additions

#### Compatibility Rules

1. **Major version must match exactly** - Breaking changes require recompilation
2. **Minor version can be higher** - Forward compatible additions are allowed
3. **Plugins without protocol_version** - Allowed but should be updated (legacy support)

#### Plugin Implementation

All external plugins **must** report their protocol version in the `--plugin-info` response:

**Go Plugin Example:**
```go
type PluginInfo struct {
    Name            string `json:"name"`
    Type            string `json:"type"`
    Version         string `json:"version"`
    ProtocolVersion string `json:"protocol_version"` // Required!
    Description     string `json:"description"`
    Author          string `json:"author"`
}

func main() {
    if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
        info := PluginInfo{
            Name:            "my-plugin",
            Type:            "output",
            Version:         "1.0.0",
            ProtocolVersion: "1.0", // Current protocol version
            Description:     "My custom plugin",
            Author:          "Your Name",
        }
        
        encoder := json.NewEncoder(os.Stdout)
        encoder.SetIndent("", "  ")
        encoder.Encode(info)
        os.Exit(0)
    }
    // ... rest of plugin logic
}
```

**Shell Script Plugin Example:**
```bash
#!/bin/bash

if [ "$1" = "--plugin-info" ]; then
  cat <<'EOF'
{
  "name": "my-plugin",
  "type": "output",
  "version": "1.0.0",
  "protocol_version": "1.0",
  "description": "My custom shell plugin",
  "author": "Your Name"
}
EOF
  exit 0
fi

# ... rest of plugin logic
```

#### Version Checking

Tinct automatically checks protocol version compatibility when loading external plugins:

```go
// In internal/plugin/manager/manager.go
func (m *Manager) RegisterExternalPlugin(name, pluginType, path, description string) error {
    // Query plugin info
    pluginInfo, err := queryPluginInfo(path)
    if err != nil {
        return fmt.Errorf("failed to query plugin info: %w", err)
    }

    // Check protocol version compatibility
    if pluginInfo.ProtocolVersion != "" {
        compatible, err := protocol.IsCompatible(pluginInfo.ProtocolVersion)
        if err != nil || !compatible {
            return fmt.Errorf(
                "plugin '%s' protocol version %s is incompatible with tinct %s: %s",
                name,
                pluginInfo.ProtocolVersion,
                protocol.ProtocolVersion,
                err.Error(),
            )
        }
    }
    
    // Register plugin...
}
```

#### Error Messages

When protocol versions are incompatible, users see clear error messages:

```
Error: plugin 'my-plugin' protocol version 2.0 is incompatible with tinct 1.0: 
incompatible major version: plugin is 2.0, tinct requires 1.x
```

#### Migration Guide

When protocol version changes occur:

**For Plugin Authors:**
1. Update `protocol_version` field in plugin-info response
2. Update plugin code to match new API contract
3. Rebuild and test plugin
4. Update plugin documentation

**For Major Version Changes (1.x → 2.x):**
- Review breaking changes in CHANGELOG
- Update all API calls to new signatures
- Test thoroughly with new tinct version
- Consider maintaining separate plugin versions for backward compatibility

**For Minor Version Changes (1.0 → 1.1):**
- Optional: Add support for new features
- No changes required if not using new features
- Plugin remains compatible with older tinct versions

#### Best Practices

1. **Always include protocol_version** - Don't rely on legacy support
2. **Test with target tinct version** - Verify compatibility before release
3. **Document required version** - Specify minimum tinct version in plugin README
4. **Use semantic versioning** - Follow semver for plugin version field
5. **Handle version gracefully** - Provide clear error messages for version mismatches

## Hook System

### Hook Interface

```go
package hook

import "context"

// Type represents when a hook should execute
type Type int

const (
    TypePreExecution Type = iota
    TypePostExecution
    TypeOnError
)

// Hook represents a lifecycle hook
type Hook interface {
    // Type returns when this hook should execute
    Type() Type
    
    // Execute runs the hook
    Execute(ctx context.Context, state *State) error
}

// State provides context to hooks
type State struct {
    Palette     *color.Palette
    PluginNames []string
    Errors      []error
}

// Manager handles hook registration and execution
type Manager interface {
    Register(hook Hook)
    ExecutePre(ctx context.Context, state *State) error
    ExecutePost(ctx context.Context, state *State) error
    ExecuteOnError(ctx context.Context, state *State) error
}
```

### Hook Example

```go
// Example: Check if required tools are installed before execution
type ReadinessHook struct{}

func (h *ReadinessHook) Type() hook.Type {
    return hook.TypePreExecution
}

func (h *ReadinessHook) Execute(ctx context.Context, state *hook.State) error {
    for _, pluginName := range state.PluginNames {
        // Check if plugin dependencies are met
    }
    return nil
}

// Example: Reload services after palette generation
type ReloadHook struct {
    services []string
}

func (h *ReloadHook) Type() hook.Type {
    return hook.TypePostExecution
}

func (h *ReloadHook) Execute(ctx context.Context, state *hook.State) error {
    for _, svc := range h.services {
        // Reload service (e.g., systemctl reload)
    }
    return nil
}
```

## Error Handling

### Error Types

```go
package errors

import "errors"

var (
    // Domain errors
    ErrInvalidImage       = errors.New("invalid image format")
    ErrInvalidColorCount  = errors.New("invalid color count")
    ErrPluginNotFound     = errors.New("plugin not found")
    ErrHookFailed         = errors.New("hook execution failed")
    
    // Sentinel errors for specific cases
    ErrNoColorsExtracted  = errors.New("no colors extracted")
    ErrConfigNotFound     = errors.New("configuration not found")
)

// PluginError wraps plugin execution errors
type PluginError struct {
    Plugin string
    Err    error
}

func (e *PluginError) Error() string {
    return fmt.Sprintf("plugin %s failed: %v", e.Plugin, e.Err)
}

func (e *PluginError) Unwrap() error {
    return e.Err
}
```

### Error Handling Best Practices

1. **Always wrap errors with context**: Use `fmt.Errorf("context: %w", err)`
2. **Define sentinel errors**: For expected error conditions
3. **Use custom error types**: For errors that need additional data
4. **Check errors with `errors.Is` and `errors.As`**: Not equality
5. **Return errors, don't panic**: Except for truly unrecoverable situations
6. **Log errors at boundaries**: Where they're handled, not everywhere

## Testing

### Test Organization

```go
// color/extractor_test.go
package color_test // Use _test package for black-box testing

import (
    "testing"
    "tinct/internal/color"
)

func TestExtractor(t *testing.T) {
    // Test implementation
}

// For testing internals, use same package
// color/palette_internal_test.go
package color

import "testing"

func TestInternalFunction(t *testing.T) {
    // Can access unexported symbols
}
```

### Test Patterns

```go
// Table-driven tests
func TestPaletteGeneration(t *testing.T) {
    tests := []struct {
        name       string
        colors     []color.Color
        count      int
        wantColors int
        wantErr    bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GeneratePalette(tt.colors, tt.count)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if len(got.Colors) != tt.wantColors {
                t.Errorf("got %d colors, want %d", len(got.Colors), tt.wantColors)
            }
        })
    }
}

// Subtests for related scenarios
func TestPluginExecution(t *testing.T) {
    t.Run("successful execution", func(t *testing.T) {
        // ...
    })
    
    t.Run("missing dependencies", func(t *testing.T) {
        // ...
    })
    
    t.Run("invalid palette", func(t *testing.T) {
        // ...
    })
}

// Use testify for assertions (optional but recommended)
import "github.com/stretchr/testify/assert"

func TestSomething(t *testing.T) {
    result := DoSomething()
    assert.NoError(t, result.Error)
    assert.Equal(t, expected, result.Value)
}
```

### Test Coverage

- Aim for >80% coverage on critical paths
- 100% coverage on public APIs
- Test error paths, not just happy paths
- Use coverage tool: `go test -cover ./...`

### Benchmarks

```go
func BenchmarkColorExtraction(b *testing.B) {
    img := loadTestImage()
    extractor := NewKMeansExtractor()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = extractor.Extract(context.Background(), img)
    }
}
```

## Documentation

### Package Documentation

```go
// Package palette provides colour palette generation and manipulation.
//
// The palette package defines the core types and operations for working
// with colour palettes. It supports multiple colour spaces and provides
// utilities for palette optimization and colour harmony analysis.
//
// Example usage:
//
//     p := palette.New(colors)
//     optimized := p.Optimize()
//     hex := optimized.ToHex()
//
package palette
```

### Function Documentation

```go
// Extract analyzes the image and extracts the specified number of dominant colors.
// It uses the configured algorithm (K-means, median cut, etc.) to identify colors
// that best represent the image.
//
// The context can be used to cancel long-running extractions. If the image is nil
// or extraction fails, an error is returned.
//
// Example:
//
//     colors, err := extractor.Extract(ctx, img)
//     if err != nil {
//         return fmt.Errorf("extraction failed: %w", err)
//     }
//
func (e *Extractor) Extract(ctx context.Context, img image.Image) ([]color.Color, error) {
    // implementation
}
```

### README and User Documentation

- Keep README.md user-focused with examples
- Use docs/ directory for detailed guides
- Include architecture diagrams where helpful
- Provide plugin development guide
- Document configuration options thoroughly

## Summary Checklist

When implementing features, ensure:

- [ ] Uses Go 1.21+ features appropriately
- [ ] Follows SOLID principles
- [ ] Idiomatic Go code (passes `go vet`, `golangci-lint`)
- [ ] Proper error handling with context
- [ ] Interfaces used for dependencies
- [ ] Comprehensive tests (table-driven where applicable)
- [ ] Documented public APIs
- [ ] Context used for cancellation
- [ ] No goroutine leaks
- [ ] Thread-safe when accessed concurrently
- [ ] Cobra commands properly structured
- [ ] Viper configuration integrated
- [ ] Plugin interface implemented correctly
- [ ] Hooks registered and executed appropriately

---

**Note**: These guidelines are living documentation. Update them as the project evolves and new patterns emerge.