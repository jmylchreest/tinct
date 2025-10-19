# External Plugins Guide

Complete guide for developing and using external plugins with Tinct.

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Plugin Types](#plugin-types)
- [Quick Start](#quick-start)
- [Go Plugins](#go-plugins)
- [Python Plugins](#python-plugins)
- [Shell Script Plugins](#shell-script-plugins)
- [Plugin Discovery](#plugin-discovery)
- [Configuration](#configuration)
- [Security](#security)
- [Testing](#testing)
- [Publishing](#publishing)
- [Troubleshooting](#troubleshooting)

## Overview

Tinct supports external plugins that extend functionality without requiring modifications to the core application. External plugins run as separate processes, providing:

- **Process Isolation**: Plugin crashes don't affect Tinct
- **Language Flexibility**: Write plugins in any language
- **Easy Distribution**: Simple binary or script distribution
- **Version Independence**: Plugins don't need recompilation when Tinct updates

## Architecture

### Plugin Communication

Tinct uses **HashiCorp's go-plugin** protocol for robust plugin communication:

```
┌────────────────────────────────────────────────────────┐
│                    Tinct (Host)                        │
│                                                        │
│  ┌──────────────────────────────────────────────────┐  │
│  │           Plugin Manager                         │  │
│  │  - Discovery                                     │  │
│  │  - Loading                                       │  │
│  │  - RPC Communication                             │  │
│  └──────────────┬───────────────────────────────────┘  │
└─────────────────┼──────────────────────────────────────┘
                  │ RPC (net/rpc or gRPC)
                  │
    ┌─────────────┼────────────┬──────────────┐
    │             │            │              │
    ▼            ▼            ▼              ▼
┌────────┐   ┌────────┐   ┌──────────┐   ┌─────────┐
│ Plugin │   │ Plugin │   │  Plugin  │   │ Plugin  │
│   A    │   │   B    │   │    C     │   │    D    │
│  (Go)  │   │(Python)│   │  (Bash)  │   │ (Rust)  │
└────────┘   └────────┘   └──────────┘   └─────────┘
```

### Plugin Protocol

**Protocol Version**: 1

**Handshake**:
- Magic Cookie: `TINCT_PLUGIN=tinct_color_palette`
- Protocol Version: `1`

**Communication Methods**:
1. **RPC over stdio** (recommended for simplicity)
2. **gRPC** (for high performance)

## Plugin Types

### Source Plugins

Generate or fetch wallpapers.

**Interface:**
```go
type SourcePlugin interface {
    Name() string
    Generate(ctx context.Context, params SourceParams) (image.Image, error)
}
```

**Naming**: `tinct-source-<name>`

**Examples**:
- `tinct-source-wallhaven` - Fetch from Wallhaven
- `tinct-source-bing` - Bing daily wallpaper
- `tinct-source-custom` - Custom API integration

### Output Plugins

Generate configuration files from palettes.

**Interface:**
```go
type OutputPlugin interface {
    Name() string
    Execute(ctx context.Context, palette *Palette) error
}
```

**Naming**: `tinct-output-<name>`

**Examples**:
- `tinct-output-hyprland` - Hyprland compositor
- `tinct-output-waybar` - Waybar status bar
- `tinct-output-neovim` - Neovim color scheme

## Quick Start

### Install Plugin Template

```bash
# Clone plugin template
git clone https://github.com/jmylchreest/tinct-plugin-template.git my-plugin
cd my-plugin

# Initialize
./init.sh wallhaven source  # Creates tinct-source-wallhaven
```

### Minimal Go Plugin

**`main.go`**:
```go
package main

import (
    "context"
    "github.com/hashicorp/go-plugin"
    tinct "github.com/jmylchreest/tinct/pkg/plugin"
)

type MySource struct{}

func (s *MySource) Name() string {
    return "mysource"
}

func (s *MySource) Generate(ctx context.Context, params tinct.SourceParams) (image.Image, error) {
    // Your implementation
    return fetchWallpaper(params)
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: tinct.Handshake,
        Plugins: map[string]plugin.Plugin{
            "source": &tinct.SourcePluginRPC{Impl: &MySource{}},
        },
    })
}
```

**Build and Install**:
```bash
go build -o tinct-source-mysource
mkdir -p ~/.config/tinct/plugins
cp tinct-source-mysource ~/.config/tinct/plugins/
chmod +x ~/.config/tinct/plugins/tinct-source-mysource
```

**Use**:
```bash
tinct generate --source mysource
```

## Go Plugins

### Full Example: Wallhaven Source Plugin

**`wallhaven/main.go`**:
```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "image"
    "log"
    "net/http"
    "os"
    
    "github.com/hashicorp/go-plugin"
    "github.com/jmylchreest/tinct/pkg/plugin"
)

const apiBase = "https://wallhaven.cc/api/v1"

type WallhavenPlugin struct {
    apiKey string
    client *http.Client
}

func NewWallhavenPlugin() *WallhavenPlugin {
    return &WallhavenPlugin{
        apiKey: os.Getenv("WALLHAVEN_API_KEY"),
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

func (p *WallhavenPlugin) Name() string {
    return "wallhaven"
}

func (p *WallhavenPlugin) Generate(ctx context.Context, params plugin.SourceParams) (image.Image, error) {
    width, height := params.ResolveDimensions()
    
    // Build search URL
    url := fmt.Sprintf("%s/search?resolution=%dx%d&sorting=toplist", 
        apiBase, width, height)
    
    if len(params.Tags) > 0 {
        url += "&q=" + strings.Join(params.Tags, "+")
    }
    
    if p.apiKey != "" {
        url += "&apikey=" + p.apiKey
    }
    
    // Search for wallpaper
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := p.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result struct {
        Data []struct {
            Path string `json:"path"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    if len(result.Data) == 0 {
        return nil, fmt.Errorf("no wallpapers found")
    }
    
    // Download wallpaper
    imgResp, err := http.Get(result.Data[0].Path)
    if err != nil {
        return nil, err
    }
    defer imgResp.Body.Close()
    
    img, _, err := image.Decode(imgResp.Body)
    return img, err
}

func (p *WallhavenPlugin) Metadata() plugin.SourceMetadata {
    return plugin.SourceMetadata{
        Name:        "wallhaven",
        Description: "Fetch wallpapers from Wallhaven.cc",
        Type:        plugin.SourceTypeFetch,
        RequiresAPI: false,
        APIKeyEnv:   "WALLHAVEN_API_KEY",
    }
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: plugin.Handshake,
        Plugins: map[string]plugin.Plugin{
            "source": &plugin.SourcePluginRPC{
                Impl: NewWallhavenPlugin(),
            },
        },
    })
}
```

**`go.mod`**:
```go
module github.com/yourusername/tinct-source-wallhaven

go 1.25

require (
    github.com/hashicorp/go-plugin v1.6.0
    github.com/jmylchreest/tinct v0.1.0
)
```

**Build**:
```bash
go build -o tinct-source-wallhaven
```

### Full Example: Hyprland Output Plugin

**`hyprland/main.go`**:
```go
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "text/template"
    
    "github.com/hashicorp/go-plugin"
    "github.com/jmylchreest/tinct/pkg/plugin"
)

type HyprlandPlugin struct {
    configPath string
}

func NewHyprlandPlugin() *HyprlandPlugin {
    home, _ := os.UserHomeDir()
    return &HyprlandPlugin{
        configPath: filepath.Join(home, ".config", "hypr", "colors.conf"),
    }
}

func (p *HyprlandPlugin) Name() string {
    return "hyprland"
}

func (p *HyprlandPlugin) Execute(ctx context.Context, palette *plugin.Palette) error {
    tmpl := `# Tinct generated colors for Hyprland
$background = rgb({{.Background}})
$foreground = rgb({{.Foreground}})
$cursor = rgb({{.Cursor}})

{{range $i, $color := .Colors}}
$color{{$i}} = rgb({{$color}})
{{end}}

# Apply to Hyprland
general {
    col.active_border = $color4 $color5 45deg
    col.inactive_border = $background
}

decoration {
    col.shadow = $background
}
`
    
    t, err := template.New("hyprland").Parse(tmpl)
    if err != nil {
        return err
    }
    
    f, err := os.Create(p.configPath)
    if err != nil {
        return err
    }
    defer f.Close()
    
    return t.Execute(f, map[string]interface{}{
        "Background": palette.Background.Hex(),
        "Foreground": palette.Foreground.Hex(),
        "Cursor":     palette.Cursor.Hex(),
        "Colors":     palette.Colors,
    })
}

func (p *HyprlandPlugin) Metadata() plugin.Metadata {
    return plugin.Metadata{
        Name:        "hyprland",
        Description: "Generate Hyprland color configuration",
        Version:     "1.0.0",
    }
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: plugin.Handshake,
        Plugins: map[string]plugin.Plugin{
            "output": &plugin.OutputPluginRPC{
                Impl: NewHyprlandPlugin(),
            },
        },
    })
}
```

## Python Plugins

### Python Plugin Helper Library

**Install**:
```bash
pip install tinct-plugin
```

### Example: Reddit Source Plugin

**`tinct-source-reddit`**:
```python
#!/usr/bin/env python3
"""Tinct source plugin for Reddit wallpapers"""

import os
import sys
from io import BytesIO
from tinct_plugin import SourcePlugin, serve_plugin

import praw
import requests
from PIL import Image

class RedditWallpaperPlugin(SourcePlugin):
    def __init__(self):
        self.reddit = praw.Reddit(
            client_id=os.getenv('REDDIT_CLIENT_ID'),
            client_secret=os.getenv('REDDIT_SECRET'),
            user_agent='tinct-plugin'
        )
    
    def name(self):
        return "reddit"
    
    def generate(self, params):
        """Fetch wallpaper from Reddit"""
        width = params.get('width', 3840)
        height = params.get('height', 2160)
        subreddit = params.get('extra', {}).get('subreddit', 'wallpapers')
        
        # Search for appropriate resolution
        query = f"self:no {width}x{height}"
        
        for post in self.reddit.subreddit(subreddit).search(
            query, time_filter='week', limit=50
        ):
            if post.url.endswith(('.jpg', '.png')):
                try:
                    resp = requests.get(post.url, timeout=10)
                    img = Image.open(BytesIO(resp.content))
                    
                    # Resize if needed
                    if img.size != (width, height):
                        img = img.resize((width, height), Image.LANCZOS)
                    
                    return img
                except Exception as e:
                    print(f"Failed to fetch {post.url}: {e}", file=sys.stderr)
                    continue
        
        raise Exception("No suitable wallpaper found")
    
    def metadata(self):
        return {
            'name': 'reddit',
            'description': 'Fetch wallpapers from Reddit',
            'type': 'fetch',
            'requires_api': True,
            'api_key_env': 'REDDIT_CLIENT_ID'
        }

if __name__ == '__main__':
    serve_plugin(RedditWallpaperPlugin())
```

**Make executable**:
```bash
chmod +x tinct-source-reddit
```

### Example: Neovim Output Plugin

**`tinct-output-neovim`**:
```python
#!/usr/bin/env python3
"""Tinct output plugin for Neovim color scheme"""

import os
from pathlib import Path
from tinct_plugin import OutputPlugin, serve_plugin

class NeovimPlugin(OutputPlugin):
    def __init__(self):
        self.colors_dir = Path.home() / '.config' / 'nvim' / 'colors'
        self.colors_dir.mkdir(parents=True, exist_ok=True)
    
    def name(self):
        return "neovim"
    
    def execute(self, palette):
        """Generate Neovim color scheme"""
        scheme_name = 'tinct'
        output_file = self.colors_dir / f'{scheme_name}.vim'
        
        with open(output_file, 'w') as f:
            f.write(f'''\" Tinct generated color scheme
\" Name: {scheme_name}

set background=dark
highlight clear
if exists("syntax_on")
  syntax reset
endif
let g:colors_name = "{scheme_name}"

\" UI Colors
hi Normal guifg={palette['foreground']} guibg={palette['background']}
hi Cursor guifg={palette['background']} guibg={palette['cursor']}
hi Visual guibg={palette['colors'][1]}
hi CursorLine guibg={palette['colors'][0]}
hi StatusLine guifg={palette['foreground']} guibg={palette['colors'][8]}
hi StatusLineNC guifg={palette['colors'][7]} guibg={palette['colors'][0]}

\" Syntax Colors
hi Comment guifg={palette['colors'][8]}
hi Constant guifg={palette['colors'][1]}
hi String guifg={palette['colors'][2]}
hi Identifier guifg={palette['colors'][4]}
hi Function guifg={palette['colors'][4]}
hi Statement guifg={palette['colors'][5]}
hi Keyword guifg={palette['colors'][5]}
hi Type guifg={palette['colors'][3]}
hi Special guifg={palette['colors'][6]}
''')
        
        print(f"Generated Neovim color scheme: {output_file}")
    
    def metadata(self):
        return {
            'name': 'neovim',
            'description': 'Generate Neovim color scheme',
            'version': '1.0.0'
        }

if __name__ == '__main__':
    serve_plugin(NeovimPlugin())
```

## Shell Script Plugins

### Simple JSON Protocol

For shell scripts, use a simple JSON stdin/stdout protocol:

**`tinct-source-bing`**:
```bash
#!/bin/bash
# Tinct source plugin: Bing daily wallpaper

set -e

# Read params from stdin
params=$(cat)

# Extract parameters
width=$(echo "$params" | jq -r '.width // 3840')
height=$(echo "$params" | jq -r '.height // 2160')

# Fetch Bing daily wallpaper
api_url="https://www.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=en-US"
image_url=$(curl -s "$api_url" | jq -r '.images[0].url')
full_url="https://www.bing.com${image_url}"

# Download and resize
tmp_file=$(mktemp).jpg
curl -s "$full_url" -o "$tmp_file"

# Resize with ImageMagick
convert "$tmp_file" -resize "${width}x${height}^" -gravity center -extent "${width}x${height}" -

# Cleanup
rm -f "$tmp_file"
```

**Make executable**:
```bash
chmod +x tinct-source-bing
```

**`tinct-output-i3`**:
```bash
#!/bin/bash
# Tinct output plugin: i3 window manager

set -e

# Read palette from stdin
palette=$(cat)

# Extract colors
bg=$(echo "$palette" | jq -r '.background')
fg=$(echo "$palette" | jq -r '.foreground')
c0=$(echo "$palette" | jq -r '.colors[0]')
c1=$(echo "$palette" | jq -r '.colors[1]')
c2=$(echo "$palette" | jq -r '.colors[2]')
c4=$(echo "$palette" | jq -r '.colors[4]')

# Generate i3 config
config_file="$HOME/.config/i3/colors"
cat > "$config_file" <<EOF
# Tinct generated colors for i3
set \$bg     $bg
set \$fg     $fg
set \$c0     $c0
set \$c1     $c1
set \$c2     $c2
set \$c4     $c4

# Window colors
#                       border  background text    indicator
client.focused          \$c4    \$c4       \$bg    \$c2
client.unfocused        \$c0    \$c0       \$fg    \$c0
client.focused_inactive \$c0    \$c0       \$fg    \$c0
client.urgent           \$c1    \$c1       \$bg    \$c1

# Bar colors
bar {
    colors {
        background \$bg
        statusline \$fg
        separator  \$c0
        
        #                  border  background text
        focused_workspace  \$c4    \$c4       \$bg
        active_workspace   \$c0    \$c0       \$fg
        inactive_workspace \$bg    \$bg       \$fg
        urgent_workspace   \$c1    \$c1       \$bg
    }
}
EOF

echo "Generated i3 colors: $config_file"

# Reload i3
i3-msg reload >/dev/null 2>&1 || true
```

## Plugin Discovery

### Discovery Paths

Tinct searches for plugins in the following order:

1. `~/.config/tinct/plugins/` (user plugins)
2. `~/.local/share/tinct/plugins/` (user plugins, XDG)
3. `/usr/local/share/tinct/plugins/` (system-wide)
4. `/usr/share/tinct/plugins/` (system-wide)
5. `./plugins/` (development)

### Plugin Listing

```bash
# List all discovered plugins
tinct plugins list

# Output:
# Source Plugins:
#   wallhaven    Fetch wallpapers from Wallhaven.cc
#   reddit       Fetch wallpapers from Reddit
#   bing         Bing daily wallpaper
#
# Output Plugins:
#   hyprland     Generate Hyprland color configuration
#   neovim       Generate Neovim color scheme
#   waybar       Generate Waybar colors
```

### Plugin Info

```bash
# Show detailed plugin information
tinct plugins info wallhaven

# Output:
# Name:        wallhaven
# Type:        Source Plugin
# Description: Fetch wallpapers from Wallhaven.cc
# Version:     1.0.0
# Path:        ~/.config/tinct/plugins/tinct-source-wallhaven
# API Key:     WALLHAVEN_API_KEY (optional)
# Supported:   ✓
```

## Configuration

### Plugin Configuration

**`~/.config/tinct/config.yaml`**:
```yaml
plugins:
  # Discovery settings
  discovery:
    enabled: true
    paths:
      - ~/.config/tinct/plugins
      - ~/.local/share/tinct/plugins
      - /usr/local/share/tinct/plugins
    
    # Skip plugins (by name)
    skip:
      - old-plugin
  
  # Default timeouts
  timeout: 60s
  
  # Per-plugin configuration
  wallhaven:
    api_key_env: WALLHAVEN_API_KEY
    default_category: general
    purity: sfw
    
  reddit:
    client_id_env: REDDIT_CLIENT_ID
    secret_env: REDDIT_SECRET
    default_subreddit: wallpapers
    
  hyprland:
    config_path: ~/.config/hypr/colors.conf
    reload_command: hyprctl reload
```

### Environment Variables

```bash
# Plugin-specific API keys
export WALLHAVEN_API_KEY="your-key-here"
export REDDIT_CLIENT_ID="your-id"
export REDDIT_SECRET="your-secret"

# Plugin paths (override defaults)
export TINCT_PLUGIN_PATH="~/.config/tinct/plugins:/usr/local/share/tinct/plugins"

# Plugin debugging
export TINCT_PLUGIN_DEBUG=1
```

## Security

### Plugin Verification

**Checksum Verification**:
```bash
# Generate checksums for your plugins
cd ~/.config/tinct/plugins
sha256sum * > checksums.txt

# Tinct verifies on load (optional)
tinct config set plugins.verify_checksums true
```

### Plugin Sandboxing

**Resource Limits** (Linux):
```yaml
plugins:
  security:
    # Maximum execution time
    timeout: 60s
    
    # CPU limit (percentage)
    cpu_limit: 50
    
    # Memory limit
    memory_limit: 512M
    
    # Network access
    allow_network: true
    
    # Filesystem access
    filesystem_access: limited  # limited, readonly, full
```

### Trusted Plugins

```yaml
plugins:
  trusted:
    # Only allow these plugins
    whitelist:
      - wallhaven
      - reddit
      - hyprland
    
    # Or block specific plugins
    blacklist:
      - suspicious-plugin
```

## Testing

### Manual Testing

```bash
# Test plugin directly
echo '{"width":1920,"height":1080}' | ~/.config/tinct/plugins/tinct-source-wallhaven

# Test with Tinct in debug mode
TINCT_PLUGIN_DEBUG=1 tinct generate --source wallhaven --size 1080p
```

### Automated Testing

**`plugin_test.go`**:
```go
package main

import (
    "context"
    "testing"
    
    "github.com/jmylchreest/tinct/pkg/plugin"
)

func TestWallhavenPlugin(t *testing.T) {
    p := NewWallhavenPlugin()
    
    params := plugin.SourceParams{
        Width:  1920,
        Height: 1080,
        Tags:   []string{"nature"},
    }
    
    img, err := p.Generate(context.Background(), params)
    if err != nil {
        t.Fatalf("Generate failed: %v", err)
    }
    
    if img.Bounds().Dx() != 1920 || img.Bounds().Dy() != 1080 {
        t.Errorf("Wrong dimensions: got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
    }
}
```

## Publishing

### GitHub Release

```yaml
# .goreleaser.yml for your plugin
project_name: tinct-source-wallhaven

builds:
  - binary: tinct-source-wallhaven
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
```

### Plugin Registry (Planned)

```bash
# Publish to Tinct plugin registry
tinct plugin publish wallhaven \
  --repo github.com/yourusername/tinct-source-wallhaven \
  --version 1.0.0

# Users can install
tinct plugin install wallhaven
```

### Installation Script

```bash
#!/bin/bash
# install.sh for your plugin

PLUGIN_NAME="tinct-source-wallhaven"
VERSION="1.0.0"
INSTALL_DIR="$HOME/.config/tinct/plugins"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
esac

# Download
URL="https://github.com/yourusername/${PLUGIN_NAME}/releases/download/v${VERSION}/${PLUGIN_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"
curl -sL "$URL" | tar xz

chmod +x "$PLUGIN_NAME"
echo "Installed $PLUGIN_NAME to $INSTALL_DIR"
```

## Troubleshooting

### Plugin Not Found

```bash
# Check discovery paths
tinct plugins paths

# List all discovered plugins
tinct plugins list

# Check if plugin is executable
ls -la ~/.config/tinct/plugins/tinct-source-*

# Make plugin executable
chmod +x ~/.config/tinct/plugins/tinct-source-wallhaven
```

### Plugin Fails to Load

```bash
# Enable debug logging
export TINCT_PLUGIN_DEBUG=1
tinct generate --source wallhaven

# Check plugin handshake
~/.config/tinct/plugins/tinct-source-wallhaven
# Should output: 1|1|tcp|...
```

### Plugin Timeout

```yaml
# Increase timeout in config
plugins:
  timeout: 120s  # Increase from default 60s
```

### API Key Issues

```bash
# Verify environment variables
env | grep -i api

# Check plugin config
tinct config get plugins.wallhaven
```

## Examples Repository

Find more examples at:
- **Official Examples**: https://github.com/jmylchreest/tinct-plugins
- **Community Plugins**: https://github.com/topics/tinct-plugin

## See Also

- [DEVELOPMENT.md](../DEVELOPMENT.md) - Core plugin interfaces
- [README.md](../README.md) - Built-in plugins
- [ai-wallpaper.md](ai-wallpaper.md) - AI generation plugins

---

**Happy plugin development! 🔌**