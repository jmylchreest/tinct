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
| **Health checks** | None | Automatic monitoring |
| **Error handling** | stderr text | Structured RPC errors |
| **Bidirectional** | No | Yes (plugin can call back) |
| **State** | Stateless | Can maintain state |
| **Dependencies** | None | hashicorp/go-plugin |
| **Complexity** | Minimal | Moderate |

## Automatic detection

Tinct automatically detects which protocol to use:

1. Runs `plugin --plugin-info`
2. Parses the `plugin_protocol` field
3. Uses the appropriate executor

No configuration required.

## JSON-stdio protocol

Best for simple plugins written in any language.

### When to use

- Writing plugins in shell scripts, Python, Ruby, or any language
- Plugin is simple and short-lived
- Maximum portability
- Minimal dependencies

### How it works

**1. Plugin info query**

```bash
$ my-plugin --plugin-info
```

Returns:

```json
{
  "name": "my-plugin",
  "type": "output",
  "version": "1.0.0",
  "protocol_version": "0.2.0",
  "description": "My plugin",
  "plugin_protocol": "json-stdio"
}
```

### Optional metadata block

Plugins may include an optional `metadata` object describing routine behaviour. This lets the documentation pipeline diff a plugin's README frontmatter against its real runtime behaviour, catching drift.

```json
{
  "name": "my-plugin",
  "type": "output",
  "version": "1.0.0",
  "protocol_version": "0.2.0",
  "description": "My plugin",
  "plugin_protocol": "json-stdio",
  "metadata": {
    "required_binaries": ["myapp"],
    "optional_binaries": [],
    "default_output_dir": "~/.config/myapp",
    "generated_files": ["tinct.conf"],
    "pattern": "single-file",
    "reload": {
      "method": "signal",
      "command": "pkill -USR1 -x myapp",
      "user_action_required": false
    }
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `required_binaries` | string[] | Binaries that must exist on PATH for the plugin to run |
| `optional_binaries` | string[] | Binaries used opportunistically (e.g. reload helpers) |
| `default_output_dir` | string | Path the plugin writes to by default (`~` left unexpanded) |
| `generated_files` | string[] | Filenames the plugin produces inside `default_output_dir` |
| `pattern` | string | `two-file`, `single-file`, `drop-in`, or `flavor-pack` |
| `reload.method` | string | `ipc`, `signal`, `watch`, `wallpaper-apply`, or `none` |
| `reload.command` | string | Descriptive — the exact reload action the plugin performs |
| `reload.user_action_required` | bool | True if the user must manually reload (no automatic reload available) |

The entire `metadata` object is optional and additive — clients ignore it if absent. Internal plugins don't need to populate it because the manager uses `hooks.Spec` directly.

**2. Execution flow**

```
tinct                          plugin
  │                              │
  │──── spawn process ──────────▶│
  │                              │
  │──── JSON via stdin ─────────▶│
  │                              │
  │◀─── JSON via stdout ─────────│
  │     (status to stderr) ──────│
  │                              │
  │◀─── process exits ───────────│
  │                              │
```

### Input plugin data format

Input plugins receive configuration and must return palette data:

**Stdin (configuration):**
```json
{
  "source": "/path/to/image.jpg",
  "options": {
    "colours": 16,
    "algorithm": "kmeans"
  }
}
```

**Stdout (palette):**
```json
{
  "colours": {
    "background": {"hex": "#1e1e2e", "rgb": {"r": 30, "g": 30, "b": 46}},
    "foreground": {"hex": "#cdd6f4", "rgb": {"r": 205, "g": 214, "b": 244}},
    "accent1": {"hex": "#89b4fa", "rgb": {"r": 137, "g": 180, "b": 250}}
  },
  "theme_type": "dark",
  "wallpaper_path": "/path/to/image.jpg"
}
```

### Output plugin data format

Output plugins receive the full palette via stdin and must write a structured JSON response to stdout.

**Stdin (palette):**
```json
{
  "colours": {
    "background": {
      "role": "background",
      "hex": "#1e1e2e",
      "rgb": {"r": 30, "g": 30, "b": 46},
      "rgba": {"r": 30, "g": 30, "b": 46, "a": 255},
      "luminance": 0.027,
      "is_light": false,
      "hue": 240,
      "saturation": 0.21
    },
    "foreground": {
      "role": "foreground",
      "hex": "#cdd6f4",
      "rgb": {"r": 205, "g": 214, "b": 244},
      "rgba": {"r": 205, "g": 214, "b": 244, "a": 255}
    }
  },
  "theme_type": "dark",
  "all_colours": [...],
  "wallpaper_path": "/path/to/wallpaper.jpg"
}
```

**Stdout (structured response — protocol 0.2.0):**
```json
{
  "success": true,
  "files_written": [
    "/home/user/.config/myapp/colours.conf"
  ],
  "message": "Generated configuration"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `success` | boolean | Yes | Whether the plugin completed successfully |
| `files_written` | string[] | Yes | Absolute paths of files written by the plugin (empty array if none) |
| `message` | string | No | Human-readable status message |

**Important:** Plugins must write all informational/status output to **stderr**. Only the JSON response goes to **stdout**. Tinct displays stderr output to the user.

For plugins that report failure, set `success` to `false` with an explanatory `message`. The plugin should still exit with code 0 — Tinct treats non-zero exit codes as a process-level crash, distinct from a plugin-reported failure.

### Error handling

Write errors and status messages to stderr:

```bash
echo "Generating theme configuration..." >&2
echo "Error: config directory not found" >&2
```

For fatal errors, either write a failure response to stdout or exit non-zero:

```bash
# Option 1: Structured failure (preferred)
echo '{"success":false,"files_written":[],"message":"Config directory not found"}' 
exit 0

# Option 2: Process-level error (last resort)
echo "Error: config directory not found" >&2
exit 1
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

**1. Plugin info query** (same as JSON-stdio)

```bash
$ my-plugin --plugin-info
```

Returns:

```json
{
  "name": "my-plugin",
  "type": "output",
  "version": "1.0.0",
  "protocol_version": "0.2.0",
  "description": "My Go plugin",
  "plugin_protocol": "go-plugin"
}
```

**2. Execution flow**

```
tinct                          plugin (persistent)
  │                              │
  │──── spawn process ──────────▶│
  │                              │
  │──── RPC: Generate() ────────▶│
  │◀─── RPC response ────────────│
  │                              │
  │──── RPC: Generate() ────────▶│  (reused)
  │◀─── RPC response ────────────│
  │                              │
  │──── RPC: Kill() ────────────▶│
  │◀─── process exits ───────────│
```

### Performance comparison

| Protocol | 100 Invocations | Avg per Call |
|----------|-----------------|--------------|
| JSON-stdio | ~5.2s | ~52ms |
| Go-Plugin | ~0.8s | ~8ms |

Go-Plugin is approximately 6x faster for repeated invocations due to process reuse.

### RPC interface

Output plugins implement:

```go
type OutputPlugin interface {
    // Generate creates configuration files from palette data
    Generate(ctx context.Context, palette PaletteData) (map[string][]byte, error)

    // PreExecute validates environment (optional)
    PreExecute(ctx context.Context) (skip bool, reason string, err error)

    // PostExecute runs after file generation (optional)
    PostExecute(ctx context.Context, files []string) error

    // GetMetadata returns plugin information
    GetMetadata() PluginInfo
}
```

Input plugins implement:

```go
type InputPlugin interface {
    // Extract generates a palette from the source
    Extract(ctx context.Context, config InputConfig) (*PaletteData, error)

    // GetMetadata returns plugin information
    GetMetadata() PluginInfo

    // WallpaperPath returns path to wallpaper (optional)
    WallpaperPath() string
}
```

### Handshake configuration

```go
var Handshake = plugin.HandshakeConfig{
    ProtocolVersion:  1,
    MagicCookieKey:   "TINCT_PLUGIN",
    MagicCookieValue: "tinct",
}
```

## Wallpaper support

Input plugins can provide wallpaper images to output plugins.

### JSON-stdio format

Return `wallpaper_path` in the response:

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

Output plugins receive the wallpaper path in the palette data and can use it for applications like hyprpaper.

## Protocol version

The current protocol version is `0.2.0`. Plugins should declare this in their metadata:

```json
{
  "protocol_version": "0.2.0"
}
```

### Version history

| Version | Changes |
|---------|---------|
| `0.2.0` | Structured JSON response for output plugins (`success`, `files_written`, `message`). Plugins write files themselves and report paths. Informational output goes to stderr only. |
| `0.0.1` | Initial version. Output plugin stdout treated as freeform text, displayed to user. No file tracking for json-stdio plugins. |

Tinct checks version compatibility and warns if a plugin uses an incompatible version. Plugins using protocol versions older than `0.2.0` continue to work with legacy behavior (freeform stdout displayed to user, no file tracking).

## See also

- [Creating plugins](./creating.md) - Implementation examples
- [Lifecycle hooks](./hooks.md) - Pre/post execution
- [HashiCorp go-plugin](https://github.com/hashicorp/go-plugin) - RPC framework
