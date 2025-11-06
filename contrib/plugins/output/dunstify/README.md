# Dunstify Plugin

**Type:** Output Plugin  
**Protocol:** go-plugin (RPC)  
**Language:** Go  
**Purpose:** Send desktop notifications when themes are generated

## Overview

The `dunstify` plugin sends desktop notifications via `dunstify` or `notify-send` when Tinct generates themes. This is a **demonstration plugin** showcasing the go-plugin protocol, which provides better performance than JSON-stdio through process reuse (~8ms vs ~52ms per call).

This plugin demonstrates:
- ✅ **go-plugin RPC protocol** - Process reuse for faster execution
- ✅ **PreExecute hook** - Check prerequisites before execution
- ✅ **PostExecute hook** - Act after theme files are written
- ✅ **Graceful fallback** - Works with both `dunstify` and `notify-send`
- ✅ **No file generation** - Notification-only plugin pattern

## Features

- Sends desktop notification when themes are generated
- Shows count of generated theme files
- Automatically detects available notification command
- Falls back from `dunstify` → `notify-send` → graceful skip
- Uses low urgency with 5-second timeout
- Includes themed icon (`preferences-desktop-theme`)

## Requirements

One of the following notification commands must be installed:

- **dunstify** (recommended) - Dunst notification utility with extended features
- **notify-send** - Standard freedesktop notification utility

### Installation

**Arch Linux:**
```bash
sudo pacman -S dunst  # Includes dunstify
# or
sudo pacman -S libnotify  # Provides notify-send
```

**Ubuntu/Debian:**
```bash
sudo apt install dunst
# or
sudo apt install libnotify-bin
```

**Fedora:**
```bash
sudo dnf install dunst
# or
sudo dnf install libnotify
```

## Building

```bash
cd contrib/plugins/output/dunstify
go build -o dunstify main.go
```

Or use GoReleaser (builds for multiple platforms):
```bash
goreleaser build --single-target --id plugin-dunstify
```

## Installation

### From GitHub Releases (Recommended)

```bash
# Download and install for your platform
tinct plugins add https://github.com/jmylchreest/tinct/releases/download/v1.0.0/tinct-plugin-dunstify_v1.0.0_Linux_x86_64.tar.gz:dunstify
```

### From Source

```bash
# Build the plugin
cd contrib/plugins/output/dunstify
go build -o dunstify main.go

# Add to Tinct
tinct plugins add ./dunstify
```

## Usage

### Basic Usage

```bash
# Generate theme with notification
tinct generate -i image -p wallpaper.jpg -o hyprland,kitty,dunstify

# The notification appears when generation completes
```

### Enable by Default

```bash
# Enable the plugin
tinct plugins enable output:dunstify

# Now it runs automatically with any output plugins
tinct generate -i image -p wallpaper.jpg -o hyprland,kitty
```

### Notification Example

When themes are generated, you'll see a desktop notification:

```
┌─────────────────────────────────────┐
│  🎨  Theme Generated                │
│                                     │
│  Generated 8 theme files            │
└─────────────────────────────────────┘
```

## How It Works

### Plugin Lifecycle

1. **PreExecute Hook** - Checks if `dunstify` or `notify-send` is available
   - If neither found → Skip gracefully (no error)
   - If found → Continue to Generate

2. **Generate** - No-op (notification plugins don't generate files)
   - Returns empty file map

3. **PostExecute Hook** - Sends desktop notification
   - Receives list of files written by other plugins
   - Counts files and sends notification
   - Tries `dunstify` first, falls back to `notify-send`

### go-plugin Protocol

This plugin uses the **go-plugin RPC protocol** instead of JSON-stdio:

**Benefits:**
- **6.5x faster** - Process reuse vs spawning new process each time
- **Automatic crash recovery** - If plugin crashes, it's restarted
- **Health monitoring** - Tinct can detect if plugin becomes unresponsive
- **Bidirectional streaming** - Can stream data if needed

**Trade-offs:**
- Must be written in Go
- Slightly more complex than JSON-stdio
- Requires hashicorp/go-plugin dependency

## Code Structure

```go
// PreExecute - Check if notification command is available
func (p *DunstifyPlugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
    if _, err := exec.LookPath("dunstify"); err == nil {
        return false, "", nil  // Found, continue
    }
    if _, err := exec.LookPath("notify-send"); err == nil {
        return false, "", nil  // Fallback, continue
    }
    return true, "Neither dunstify nor notify-send found", nil  // Skip gracefully
}

// Generate - No files to generate
func (p *DunstifyPlugin) Generate(ctx context.Context, palette protocol.PaletteData) (map[string][]byte, error) {
    return map[string][]byte{}, nil  // Empty map, no files
}

// PostExecute - Send notification after files are written
func (p *DunstifyPlugin) PostExecute(ctx context.Context, writtenFiles []string) error {
    summary := "Theme Generated"
    body := fmt.Sprintf("Generated %d theme files", len(writtenFiles))
    // Send notification via dunstify or notify-send
    return sendNotification(ctx, summary, body)
}
```

## Customization

To customize the notification, edit `main.go`:

### Change Urgency
```go
cmd := exec.CommandContext(ctx, dunstifyPath,
    "-u", "normal",  // low, normal, critical
    // ...
)
```

### Change Timeout
```go
cmd := exec.CommandContext(ctx, dunstifyPath,
    "-t", "10000",  // 10 seconds (in milliseconds)
    // ...
)
```

### Change Icon
```go
icon := "preferences-desktop-wallpaper"  // Use wallpaper icon instead
```

### Add More Details
```go
// Show first few generated files
if len(writtenFiles) > 0 {
    body = fmt.Sprintf("Generated:\n%s", strings.Join(writtenFiles[:3], "\n"))
}
```

## Testing

### Test Plugin Info
```bash
./dunstify --plugin-info
```

Expected output:
```json
{
  "name": "dunstify",
  "type": "output",
  "version": "0.0.1",
  "protocol_version": "0.0.1",
  "description": "Send desktop notifications via dunstify or notify-send",
  "plugin_protocol": "go-plugin"
}
```

### Test with Tinct
```bash
# Dry run (no files written, but notification sent)
tinct generate -i image -p wallpaper.jpg -o dunstify --dry-run

# Real run
tinct generate -i image -p wallpaper.jpg -o hyprland,dunstify
```

### Debug Mode
```bash
# Enable verbose output
tinct generate -i image -p wallpaper.jpg -o dunstify --verbose

# Check plugin detection
tinct plugins list | grep dunstify
```

## Troubleshooting

### "Neither dunstify nor notify-send found"

**Problem:** No notification command available.

**Solution:**
```bash
# Install dunst (includes dunstify)
sudo pacman -S dunst  # Arch
sudo apt install dunst  # Ubuntu/Debian

# Or install libnotify (includes notify-send)
sudo pacman -S libnotify
```

### "Plugin not found"

**Problem:** Plugin not registered with Tinct.

**Solution:**
```bash
# Add the plugin
tinct plugins add ./dunstify

# Verify it's registered
tinct plugins list
```

### Notification doesn't appear

**Problem:** Notification daemon not running.

**Solution:**
```bash
# Check if dunst is running
pgrep dunst

# Start dunst
dunst &

# Or use your system notification daemon
```

## Related Plugins

### Similar Output Plugins
- **notify-send.py** (contrib/plugins/output/) - Python script version, JSON-stdio protocol
- **hyprpaper** (internal) - Sets wallpaper, also uses PostExecute hook

### Plugin Development Resources
- [External Plugin Guide](../../README.md) - How to create plugins
- [go-plugin Documentation](https://github.com/hashicorp/go-plugin) - RPC protocol library
- [Plugin Protocol](../../../../internal/plugin/protocol/) - Tinct's plugin interfaces

## License

MIT License - Same as Tinct

## Contributing

To improve this plugin:
1. Fork the repository
2. Make changes to `contrib/plugins/output/dunstify/main.go`
3. Test with `go build && tinct plugins add ./dunstify`
4. Submit a pull request

Ideas for improvements:
- [ ] Show palette colors in notification
- [ ] Add notification action buttons (view files, revert, etc.)
- [ ] Support custom notification templates
- [ ] Add notification history/log
- [ ] Support notification urgency based on file count
