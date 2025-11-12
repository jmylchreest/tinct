# Tinct External Plugins

Quick guide for creating external plugins with Tinct. Plugins can be written in **any language** using JSON-stdio, or in **Go** using go-plugin for better performance.

## Quick Start

```bash
# Add a plugin
tinct plugins add ./my-plugin.sh

# Enable and use
tinct plugins enable my-plugin
tinct generate -i image -o my-plugin
```

## Two Plugin Protocols

Tinct automatically detects which protocol your plugin uses:

| Protocol | Best For | Speed | Languages |
|----------|----------|-------|-----------|
| **JSON-stdio** | Scripts, simple plugins | ~52ms/call | Any (bash, Python, etc.) |
| **go-plugin** | Complex Go plugins | ~8ms/call | Go only |

### JSON-stdio (Simple)
- Write in any language
- JSON in via stdin, JSON out via stdout
- New process per invocation
- Minimal dependencies

### go-plugin (Advanced)
- Write in Go using RPC
- Process reuse (6.5x faster)
- Automatic crash recovery
- Built-in health monitoring

**Detection:** Set `"plugin_protocol": "go-plugin"` in `--plugin-info` output.

## Creating Plugins

### JSON-stdio Plugin (Bash Example)

```bash
#!/bin/bash
# my-plugin.sh

if [ "$1" = "--plugin-info" ]; then
  cat <<EOF
{
  "name": "my-plugin",
  "type": "output",
  "version": "1.0.0",
  "protocol_version": "0.0.1",
  "description": "My plugin",
  "plugin_protocol": "json-stdio"
}
EOF
  exit 0
fi

# Read palette from stdin
PALETTE=$(cat)
DRY_RUN=$(echo "$PALETTE" | jq -r '.dry_run // false')
BG=$(echo "$PALETTE" | jq -r '.colours.background.hex')

if [ "$DRY_RUN" = "true" ]; then
  echo "Would process: $BG"
else
  echo "Background: $BG"
  # Do actual work
fi
```

### go-plugin Plugin (Go Example)

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "github.com/hashicorp/go-plugin"
    "github.com/jmylchreest/tinct/internal/plugin/protocol"
)

type MyPlugin struct{}

func (p *MyPlugin) Generate(ctx context.Context, palette protocol.PaletteData) (map[string][]byte, error) {
    bg := palette.Colours["background"].Hex
    return map[string][]byte{
        "output.txt": []byte(fmt.Sprintf("Background: %s\n", bg)),
    }, nil
}

func (p *MyPlugin) PreExecute(ctx context.Context) (bool, string, error) {
    return false, "", nil
}

func (p *MyPlugin) PostExecute(ctx context.Context, files []string) error {
    return nil
}

func (p *MyPlugin) GetMetadata() protocol.PluginInfo {
    return protocol.PluginInfo{
        Name:            "my-plugin",
        Type:            "output",
        Version:         "1.0.0",
        ProtocolVersion: protocol.ProtocolVersion,
        Description:     "My Go plugin",
        PluginProtocol:  "go-plugin", // Important!
    }
}

func main() {
    if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
        json.NewEncoder(os.Stdout).Encode((&MyPlugin{}).GetMetadata())
        os.Exit(0)
    }
    
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: protocol.Handshake,
        Plugins: map[string]plugin.Plugin{
            "output": &protocol.OutputPluginRPC{Impl: &MyPlugin{}},
        },
    })
}
```

Build: `go build -o tinct-plugin-myplugin`

**Note**: All Tinct plugins should follow the naming convention `tinct-plugin-{name}` to avoid conflicts with system binaries.

## Plugin Info (Required)

All plugins must respond to `--plugin-info`:

```json
{
  "name": "my-plugin",
  "type": "input|output",
  "version": "1.0.0",
  "protocol_version": "0.0.1",
  "description": "What this plugin does",
  "plugin_protocol": "json-stdio"  // or "go-plugin"
}
```

## Input/Output

### Output Plugins

**Input (stdin):**
```json
{
  "colours": {
    "background": {"hex": "#1a1b26", "rgb": {"r": 26, "g": 27, "b": 38}},
    "foreground": {"hex": "#c0caf5", "rgb": {"r": 192, "g": 202, "b": 245}}
  },
  "theme_type": 1,
  "plugin_args": {},
  "dry_run": false
}
```

**Output (stdout):** Status messages or generated content

### Input Plugins

**Input (stdin):**
```json
{
  "verbose": false,
  "dry_run": false,
  "colour_overrides": [],
  "plugin_args": {}
}
```

**Output (stdout):** Complete palette JSON (same format as output plugins receive)

## Key Fields

- `dry_run` - Don't write files/execute actions when true
- `verbose` - Log to stderr when true
- `plugin_args` - Custom arguments from `--plugin-args my-plugin='{"key":"value"}'`
- `colours` - Semantic roles: `background`, `foreground`, `accent1-4`, `danger`, `warning`, etc.

## Plugin Management

```bash
# Install
tinct plugins add ./plugin.sh
tinct plugins add https://example.com/plugin.tar.gz:plugin

# Enable/Disable
tinct plugins enable my-plugin
tinct plugins disable my-plugin
tinct plugins clear my-plugin  # Reset

# List
tinct plugins list

# Delete
tinct plugins delete my-plugin
```

## Examples

### Simple Notification (bash + JSON-stdio)

```bash
#!/bin/bash
# notify.sh - 15 lines
if [ "$1" = "--plugin-info" ]; then
  echo '{"name":"notify","type":"output","version":"1.0.0","description":"Send notifications"}'
  exit 0
fi

PALETTE=$(cat)
if [ "$(echo "$PALETTE" | jq -r '.dry_run')" != "true" ]; then
  notify-send "Theme Generated" "Tinct updated your theme"
fi
```

### All Go Plugins Use go-plugin

All Go-based plugins in this directory use the go-plugin protocol:

- **`input/random/`** - Random colour palette generator
  - RPC-based palette generation
  - Reproducible via seed
  - Process reuse for performance

- **`output/wob/`** - Wayland Overlay Bar theme generator
  - Generates wob INI themes
  - Includes wrapper for starting/controlling wob
  - Proper directory structure with plugin interface

- **`output/templater/`** - Template-based configuration file generator
  - Go templates for custom config files
  - YAML configuration
  - 600+ lines converted to go-plugin

- **`output/dunstify/`** - Desktop notifications
  - Auto-detects dunstify or notify-send
  - Fallback to notify-send if dunstify unavailable
  - Clean exit if neither found

## When to Use go-plugin?

**Use JSON-stdio** for:
- Shell scripts
- Simple notifications
- File wrappers
- Quick prototypes

**Use go-plugin** for:
- Heavy computation
- Plugins called repeatedly
- State between calls
- Better error handling

## Testing

```bash
# Test plugin info
./plugin.sh --plugin-info

# Test with sample data
echo '{"colours":{"background":{"hex":"#1a1b26"}},"dry_run":false}' | ./plugin.sh

# Test dry-run
echo '{"dry_run":true}' | ./plugin.sh
```

## Best Practices

1. Always handle `--plugin-info`
2. Check `dry_run` flag
3. Write errors to stderr, output to stdout
4. Make scripts executable (`chmod +x`)
5. Use `plugin_protocol` field for go-plugin
6. Test standalone before integrating

## More Information

- [Plugin Protocols](../../docs/PLUGIN-PROTOCOLS.md) - Detailed protocol comparison
- [Hybrid Architecture](../../docs/HYBRID-PLUGIN-SUMMARY.md) - Implementation details
- [Plugin Packages](../../docs/PLUGIN-PACKAGES.md) - Distribution guide
