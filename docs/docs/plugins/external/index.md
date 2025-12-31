---
sidebar_position: 1
---

# External plugins

Extend tinct with plugins written in any language.

## Overview

Tinct supports external plugins that run as separate processes, providing:

- **Language flexibility**: Write plugins in Go, Python, shell scripts, or any language
- **Process isolation**: Plugin crashes don't affect tinct
- **Easy distribution**: Simple binary or script distribution
- **Version independence**: Plugins don't need recompilation when tinct updates

## Plugin types

### Input plugins

Generate or fetch colour palettes from external sources.

**Naming**: `tinct-input-<name>`

**Examples**:
- `tinct-input-wallhaven` - Fetch wallpapers from Wallhaven.cc
- `tinct-input-reddit` - Get wallpapers from Reddit
- `tinct-input-bing` - Bing daily wallpaper

### Output plugins

Generate configuration files for applications not built into tinct.

**Naming**: `tinct-output-<name>`

**Examples**:
- `tinct-output-wled` - WLED LED strip controller
- `tinct-output-openrgb` - OpenRGB for PC peripherals
- `tinct-output-custom-app` - Your custom application

## Quick start

### Installing external plugins

```bash
# Add a plugin repository
tinct plugins add https://github.com/user/tinct-plugin-wled

# Sync all repositories
tinct plugins sync

# Install a specific plugin
tinct plugins install wled
```

### Using external plugins

External plugins work just like built-in ones:

```bash
# Use an external input plugin
tinct generate -i wallhaven -o kitty -o alacritty

# Use an external output plugin
tinct generate -i image -p ~/wallpaper.jpg -o wled
```

### List available plugins

```bash
# List all plugins (built-in and external)
tinct plugins

# Show plugin details
tinct plugins info wled
```

## Plugin discovery paths

Tinct searches for external plugins in:

1. `~/.config/tinct/plugins/` (user plugins)
2. `~/.local/share/tinct/plugins/` (XDG user data)
3. `/usr/local/share/tinct/plugins/` (system-wide)
4. `/usr/share/tinct/plugins/` (system-wide)
5. `./plugins/` (development)

## Protocol support

Tinct supports two communication protocols:

| Protocol | Languages | Process model | Use case |
|----------|-----------|---------------|----------|
| **JSON-stdio** | Any | New process each call | Simple plugins |
| **go-plugin** | Go | Persistent process | Performance-critical |

See [Protocols](/docs/plugins/external/protocols) for details.

## Next steps

- [Protocols](/docs/plugins/external/protocols) - Communication protocol details
- [Creating plugins](/docs/plugins/external/creating) - Write your own plugins
- [Repository management](/docs/plugins/external/repository) - Share and distribute plugins
