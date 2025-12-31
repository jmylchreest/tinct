---
sidebar_position: 2
---

# Plugin protocols

Tinct supports two plugin communication protocols with automatic detection.

## Protocol comparison

| Feature | JSON-stdio | Go-Plugin (RPC) |
|---------|-----------|-----------------|
| **Languages** | Any (shell, Python, Ruby, etc.) | Go |
| **Communication** | JSON over stdin/stdout | RPC over stdio |
| **Process model** | New process per invocation | Persistent process (reused) |
| **Startup cost** | High (fork+exec each time) | Low (RPC to running process) |
| **Isolation** | Basic process isolation | Enhanced with crash recovery |
| **Error handling** | stderr text | Structured RPC errors |
| **State** | Stateless | Can maintain state |
| **Complexity** | Minimal | Moderate |

## JSON-stdio protocol

Best for simple plugins written in any language.

### When to use

- Writing plugins in shell scripts, Python, Ruby, or any language
- Plugin is simple and short-lived
- Maximum portability
- Minimal dependencies

### How it works

1. **Plugin info query**
   ```bash
   $ my-plugin --plugin-info
   ```
   Returns:
   ```json
   {
     "name": "my-plugin",
     "type": "output",
     "version": "1.0.0",
     "protocol_version": "0.0.1",
     "description": "My plugin",
     "plugin_protocol": "json-stdio"
   }
   ```

2. **Execution**
   - Tinct spawns the plugin process
   - Sends palette data via stdin as JSON
   - Reads response from stdout
   - Process exits

### Palette data format

Plugins receive palette data as JSON on stdin:

```json
{
  "colours": {
    "background": {
      "role": "background",
      "hex": "#1e1e2e",
      "rgb": {"r": 30, "g": 30, "b": 46},
      "rgba": {"r": 30, "g": 30, "b": 46, "a": 255}
    },
    "foreground": {
      "role": "foreground",
      "hex": "#cdd6f4",
      "rgb": {"r": 205, "g": 214, "b": 244},
      "rgba": {"r": 205, "g": 214, "b": 244, "a": 255}
    }
  },
  "theme_type": "dark",
  "all_colours": [...]
}
```

## Go-plugin protocol

Best for performance-critical Go plugins.

### When to use

- Writing plugins in Go
- Plugin does heavy computation
- Need better error handling and crash recovery
- Plugin maintains state between invocations
- Want health monitoring

### How it works

1. **Plugin info query** (same as JSON-stdio)
   ```bash
   $ my-plugin --plugin-info
   ```
   Returns:
   ```json
   {
     "name": "my-plugin",
     "type": "output",
     "version": "1.0.0",
     "protocol_version": "0.0.1",
     "description": "My Go plugin",
     "plugin_protocol": "go-plugin"
   }
   ```

2. **Execution**
   - Tinct spawns plugin process once
   - Establishes RPC connection over stdio
   - Makes RPC calls for Generate/PreExecute/PostExecute
   - Process stays alive for reuse
   - Automatic crash recovery if plugin dies

### Performance comparison

| Protocol | 100 Invocations | Avg per Call |
|----------|-----------------|--------------|
| JSON-stdio | 5.2s | 52ms |
| Go-Plugin | 0.8s | 8ms |

Go-Plugin is approximately 6x faster for repeated invocations due to process reuse.

## Automatic detection

Tinct automatically detects which protocol to use:

1. Runs `plugin --plugin-info`
2. Parses the `plugin_protocol` field
3. Uses the appropriate executor:
   - `"go-plugin"` - RPC executor
   - `"json-stdio"` or omitted - JSON executor

No configuration needed.

## Wallpaper support

Input plugins can provide wallpaper images to output plugins.

### JSON-stdio format

Return a response object with `wallpaper_path`:

```json
{
  "colours": [...],
  "wallpaper_path": "/path/to/wallpaper.png"
}
```

### Go-plugin format

Implement the `WallpaperPath()` method:

```go
func (p *MyInputPlugin) WallpaperPath() string {
    return p.imagePath
}
```

## See also

- [Creating plugins](/docs/plugins/external/creating)
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin)
