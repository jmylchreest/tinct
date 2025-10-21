# Plugin Management Guide

Tinct features a powerful plugin system that allows you to extend its functionality with both built-in and external plugins. This guide covers everything you need to know about managing plugins.

## Table of Contents

- [Overview](#overview)
- [Plugin Types](#plugin-types)
- [Managing Plugins](#managing-plugins)
- [Plugin Priority](#plugin-priority)
- [External Plugins](#external-plugins)
- [Writing Custom Plugins](#writing-custom-plugins)
- [Plugin Lock File](#plugin-lock-file)
- [Examples](#examples)

## Overview

Tinct uses a plugin architecture to handle:
- **Input Plugins**: Extract or load colour palettes from various sources
- **Output Plugins**: Generate configuration files or perform actions with palettes

Plugins can be:
- **Built-in**: Shipped with Tinct (e.g., `image`, `file`, `hyprland`)
- **External**: Custom scripts or programs that follow the plugin interface

## Plugin Types

### Built-in Input Plugins

| Plugin | Description | Status |
|--------|-------------|--------|
| `image` | Extract colours from image files (JPEG, PNG, GIF, WebP) | Enabled |
| `file` | Load palette from file or build from colour specifications | Enabled |

### Built-in Output Plugins

| Plugin | Description | Default Status |
|--------|-------------|----------------|
| `hyprland` | Generate Hyprland window manager colour theme | Enabled |
| `tailwind` | Generate Tailwind CSS / shadcn/ui theme configuration | Disabled |

## Managing Plugins

### Listing Plugins

View all available plugins with their current status:

```bash
tinct plugins list
```

Output:
```
Plugins:
------------------------------------------------------------------------------------------
PLUGIN                         STATUS     DESCRIPTION
------------------------------------------------------------------------------------------
input:file                     enabled    Load palette from file or build from colour specifications
input:image                    enabled    Extract colour palette from an image file
output:hyprland                enabled    Generate Hyprland colour theme configuration
output:notify *                enabled    External plugin (source: ./contrib/notify-send.py)
output:tailwind                disabled   Generate Tailwind CSS / shadcn/ui theme configuration

* = external plugin
```

**Note**: Plugins are shown with their fully qualified name (type:name) and external plugins are marked with an asterisk (*)

For verbose output with lock file location:

```bash
tinct plugins list --verbose
```

### Enabling Plugins

Enable a disabled plugin:

```bash
# By plugin name
tinct plugins enable tailwind

# With explicit type prefix
tinct plugins enable output:tailwind

# Specify type with flag
tinct plugins enable tailwind --type output
```

This adds the plugin to the enabled list in `.tinct-plugins.json`.

### Disabling Plugins

Disable an enabled plugin:

```bash
# By plugin name
tinct plugins disable tailwind

# With explicit type prefix
tinct plugins disable output:tailwind

# Specify type with flag
tinct plugins disable tailwind --type output
```

This adds the plugin to the disabled list in `.tinct-plugins.json`.

### Custom Lock File Location

Specify a custom lock file path:

```bash
tinct plugins list --lock-file /path/to/.tinct-plugins.json
tinct plugins enable hyprland --lock-file ~/.config/tinct/plugins.json
```

## Plugin Priority

Plugin enable/disable state is determined by the following priority order (highest to lowest):

1. **Plugin Lock File** (`.tinct-plugins.json`)
2. **Environment Variables** (`TINCT_ENABLED_PLUGINS`, `TINCT_DISABLED_PLUGINS`)
3. **Plugin Defaults** (set by plugin author)

### Whitelist Mode vs Blacklist Mode

**Whitelist Mode** (when `TINCT_ENABLED_PLUGINS` or lock file `enabled_plugins` is set):
- Only explicitly listed plugins are enabled
- All other plugins are disabled
- Most restrictive mode

**Blacklist Mode** (when `TINCT_DISABLED_PLUGINS` or lock file `disabled_plugins` is set):
- Explicitly listed plugins are disabled
- All other plugins use their default state
- Less restrictive than whitelist

### Environment Variables

Control plugins via environment variables:

```bash
# Enable only specific plugins (whitelist mode)
export TINCT_ENABLED_PLUGINS="output:hyprland,output:notify,input:image"
tinct generate -i image -p wallpaper.jpg

# Disable specific plugins (blacklist mode)
export TINCT_DISABLED_PLUGINS="output:tailwind"
tinct generate -i image -p wallpaper.jpg

# Clear to use lock file or plugin defaults
unset TINCT_ENABLED_PLUGINS
unset TINCT_DISABLED_PLUGINS
```

Format: Comma-separated list of `type:name` or just `name`

**Note**: The lock file takes precedence over environment variables.

## External Plugins

External plugins allow you to extend Tinct with custom functionality without modifying the core application.

### Adding External Plugins

Add a plugin from a local path:

```bash
# Add from local file
tinct plugins add notify ./contrib/notify-send.py

# Specify plugin type explicitly
tinct plugins add my-plugin ./path/to/plugin.sh --type output

# Force overwrite if plugin already exists
tinct plugins add notify ./contrib/notify-send.py --force
```

The plugin will be:
1. Copied to `~/.local/share/tinct/plugins/`
2. Made executable
3. Registered in `.tinct-plugins.json`

### Listing External Plugins

External plugins appear in the list with their metadata:

```bash
tinct plugins list
```

Output includes:
```
External Plugins:
--------------------------------------------------------------------------------
NAME                 TYPE       STATUS     SOURCE
--------------------------------------------------------------------------------
notify               output     enabled    ./contrib/notify-send.py
```

### Deleting External Plugins

Remove an external plugin:

```bash
# Interactive confirmation
tinct plugins delete notify

# Force delete without confirmation
tinct plugins delete notify --force
```

This will:
1. Delete the plugin file from `~/.local/share/tinct/plugins/`
2. Remove the plugin from `.tinct-plugins.json`
3. Remove from enabled/disabled lists

**Note**: Built-in plugins cannot be deleted.

### Updating External Plugins

Update all external plugins from their source locations:

```bash
# Update all external plugins
tinct plugins update

# Update with custom lock file
tinct plugins update --lock-file /path/to/.tinct-plugins.json
```

This command:
1. Reads the plugin lock file
2. For each external plugin, re-downloads/copies from its `source` field
3. Updates the plugin file in `~/.local/share/tinct/plugins/`
4. Preserves plugin metadata in the lock file

**Use Cases**:
- Keep plugins synchronized across machines
- Update plugins after pulling lock file changes from version control
- Restore plugins after system changes

**Example**:
```bash
$ tinct plugins update
Updating plugin 'notify' from ./contrib/notify-send.py...
  ✓ Updated: /home/user/.local/share/tinct/plugins/notify-send.py

Update complete: 1 succeeded, 0 failed
```

### Plugin Storage

External plugins are stored in:
```
~/.local/share/tinct/plugins/
```

Each plugin is a separate executable file.

## Writing Custom Plugins

See the [External Plugins README](../contrib/README.md) for detailed information on writing custom plugins.

### Plugin Interface Requirements

All external plugins must:

1. **Be executable** (`chmod +x plugin.sh`)
2. **Respond to `--plugin-info`** with JSON metadata
3. **Read input from stdin** (for output plugins)
4. **Write output to stdout**
5. **Exit with code 0** on success, non-zero on error

### Plugin Info Format

When called with `--plugin-info`, plugins must output:

```json
{
  "name": "plugin-name",
  "type": "output",
  "version": "1.0.0",
  "description": "Brief description",
  "enabled": true,
  "author": "Your Name",
  "requires": ["dependency1", "dependency2"]
}
```

### Color Palette Format

Output plugins receive JSON on stdin:

```json
{
  "theme_type": "dark",
  "colours": {
    "background": {
      "hex": "#1a1b26",
      "rgb": {"r": 26, "g": 27, "b": 38}
    },
    "foreground": {
      "hex": "#c0caf5",
      "rgb": {"r": 192, "g": 202, "b": 245}
    },
    "accent1": {
      "hex": "#7aa2f7",
      "rgb": {"r": 122, "g": 162, "b": 247}
    }
  },
  "all_colours": [
    {
      "hex": "#1a1b26",
      "rgb": {"r": 26, "g": 27, "b": 38},
      "role": "background"
    }
  ]
}
```

### Semantic Colour Roles

Tinct uses semantic colour roles (camelCase):

- `background` - Primary background colour
- `backgroundMuted` - Muted/secondary background
- `foreground` - Primary text colour
- `foregroundMuted` - Muted/secondary text
- `accent1` through `accent4` - Accent colours
- `danger` - Error/danger colour
- `warning` - Warning colour
- `success` - Success colour
- `info` - Info colour
- `notification` - Notification colour

## Plugin Lock File

The plugin lock file (`.tinct-plugins.json`) stores plugin configuration and metadata.

### Location

The lock file is searched in this order:
1. Path specified with `--lock-file` flag
2. Current directory: `./.tinct-plugins.json`
3. Home directory: `~/.tinct-plugins.json`

### Structure

```json
{
  "enabled_plugins": [
    "output:hyprland",
    "output:notify"
  ],
  "disabled_plugins": [
    "output:tailwind"
  ],
  "external_plugins": {
    "notify": {
      "path": "/home/user/.local/share/tinct/plugins/notify-send.py",
      "type": "output",
      "source": "./contrib/notify-send.py",
      "version": "1.0.0"
    }
  }
}
```

### Fields

- **enabled_plugins**: List of plugins to enable (whitelist mode when present)
- **disabled_plugins**: List of plugins to disable (blacklist mode)
- **external_plugins**: Metadata for registered external plugins

### Manual Editing

You can manually edit the lock file:

```json
{
  "disabled_plugins": ["output:tailwind"],
  "external_plugins": {}
}
```

The file will be validated when loaded.

## Examples

### Example 1: Disable Tailwind Globally

```bash
# Create lock file with tailwind disabled
tinct plugins disable tailwind

# Verify
tinct plugins list
# Output shows: output:tailwind    disabled    ...

# Generate will skip tailwind
tinct generate -i image -p wallpaper.jpg
```

### Example 2: Enable Only Specific Plugins

```bash
# Enable only hyprland and notify
tinct plugins enable hyprland
tinct plugins enable notify

# All other plugins are now disabled (whitelist mode)
tinct plugins list

# Generate with only these plugins
tinct generate -i image -p wallpaper.jpg
```

### Example 3: Add Custom Notification Plugin

```bash
# Add the notify plugin
tinct plugins add notify ./contrib/notify-send.py

# Enable it explicitly
tinct plugins enable notify

# Use it with generate
tinct generate -i image -p wallpaper.jpg -o hyprland,notify
```

### Example 4: Project-Specific Configuration

```bash
# Create project-specific lock file
cd my-project
tinct plugins enable tailwind --lock-file ./.tinct-plugins.json
tinct plugins disable hyprland --lock-file ./.tinct-plugins.json

# Generate uses project config
tinct generate -i image -p logo.png --lock-file ./.tinct-plugins.json
```

### Example 5: Temporary Override with Environment

```bash
# Lock file disables tailwind
cat .tinct-plugins.json
# { "disabled_plugins": ["output:tailwind"] }

# Override with environment variable (enable only tailwind)
TINCT_ENABLED_PLUGINS="output:tailwind" tinct generate -i image -p test.jpg
```

### Example 6: Testing External Plugin

```bash
# Test plugin info
~/.local/share/tinct/plugins/notify-send.py --plugin-info

# Test with sample palette
echo '{"theme_type":"dark","colours":{"accent1":{"hex":"#7aa2f7"}}}' | \
  ~/.local/share/tinct/plugins/notify-send.py
```

### Example 7: Version Control with Update

```bash
# Scenario: Working in a team with shared lock file

# 1. Clone repository with lock file
git clone https://github.com/team/project.git
cd project

# 2. Install plugins from lock file
tinct plugins update

# 3. All external plugins are now installed
tinct plugins list
# Shows all plugins including external ones marked with *

# 4. Use plugins
tinct generate -i image -p wallpaper.jpg -o hyprland,notify
```

### Example 8: List All Plugins Verbosely

```bash
# Show which lock file is being used
tinct plugins list --verbose

# Example output:
# Using lock file: /home/user/.tinct-plugins.json
#
# Plugins:
# ------------------------------------------------------------------------------------------
# PLUGIN                         STATUS     DESCRIPTION
# ------------------------------------------------------------------------------------------
# input:file                     enabled    ...
# output:notify *                enabled    External plugin (source: ...)
# ...
```

## Troubleshooting

### Plugin Not Found

**Problem**: `Error: unknown output plugin: notify`

**Solution**: Check if the plugin is registered:
```bash
tinct plugins list
```

If missing, add it:
```bash
tinct plugins add notify ./path/to/notify.py
```

### Plugin Disabled

**Problem**: `Error: output plugin 'tailwind' is currently disabled`

**Solution**: Enable the plugin:
```bash
tinct plugins enable tailwind
```

Or check why it's disabled:
```bash
tinct plugins list
cat .tinct-plugins.json
```

### Lock File Conflicts

**Problem**: Unexpected plugin state

**Solution**: Check lock file location and content:
```bash
tinct plugins list --verbose
cat .tinct-plugins.json
```

Remove lock file to reset to defaults:
```bash
rm .tinct-plugins.json
```

### External Plugin Not Executable

**Problem**: Plugin added but won't run

**Solution**: Make it executable:
```bash
chmod +x ~/.local/share/tinct/plugins/plugin-name
```

### Environment Override Not Working

**Problem**: Environment variable doesn't change plugin state

**Solution**: Lock file takes precedence. Either:
1. Remove the lock file
2. Edit the lock file
3. Use `--lock-file /dev/null` to ignore it

## Best Practices

1. **Version Control**: Add `.tinct-plugins.json` to your project for reproducible builds
2. **Plugin Updates**: Run `tinct plugins update` after pulling lock file changes
3. **Project-Specific**: Use project directories for different configurations
4. **Test First**: Use `--dry-run` and `--preview` to verify output before writing
5. **Document External Plugins**: Keep a README with plugin sources and requirements
6. **Backup**: The lock file is small - keep backups if critical
7. **Security**: Only add external plugins from trusted sources
8. **Validation**: Test external plugins independently before adding
9. **Lock File Sync**: Use `plugins update` to sync plugins across machines

## See Also

- [External Plugins README](../contrib/README.md) - Writing custom plugins
- [Main README](../README.md) - General Tinct usage
- [Examples](../examples/) - Example configurations and use cases