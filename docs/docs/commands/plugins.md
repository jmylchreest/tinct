---
sidebar_position: 4
---

# tinct plugins

Manage input and output plugins.

## Synopsis

```bash
tinct plugins <subcommand> [flags]
```

## Subcommands

| Command | Description |
|---------|-------------|
| `list` | List all available plugins |
| `repo add` | Add a plugin repository |
| `repo list` | List configured repositories |
| `repo remove` | Remove a repository |
| `install` | Install an external plugin |
| `sync` | Sync plugins from repositories |
| `enable` | Enable a plugin |
| `disable` | Disable a plugin |

## tinct plugins list

Show all available plugins with their status:

```bash
tinct plugins list
```

Output shows:

- Plugin name and type (input/output)
- Status indicators: Enabled (E), Disabled (D), On-demand (O)
- External plugins marked with asterisk (*)
- Version compatibility

### Status indicators

| Indicator | Meaning |
|-----------|---------|
| O | On-demand (loaded when used) |
| E | Enabled (always available) |
| D | Disabled (not loaded) |
| * | External plugin |
| Y | Compatible with current tinct |
| N | Incompatible version |

## tinct plugins repo

Manage plugin repositories.

### Add a repository

```bash
tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/refs/heads/main/repository/repository.json
```

### List repositories

```bash
tinct plugins repo list
```

### Remove a repository

```bash
tinct plugins repo remove official
```

## tinct plugins install

Install an external plugin from GitHub:

```bash
tinct plugins install <github-user>/<repo>
```

Example:

```bash
tinct plugins install jmylchreest/tinct-plugin-wled
```

The plugin is:

1. Downloaded from the repository
2. Verified (checksum if available)
3. Installed to `~/.local/share/tinct/plugins/`
4. Registered in `.tinct-plugins.json`

## tinct plugins sync

Synchronise plugins from all configured repositories:

```bash
tinct plugins sync
```

This:

- Fetches repository manifests
- Updates plugin metadata
- Reports available updates

## tinct plugins enable/disable

Control which plugins are active:

```bash
# Enable a plugin
tinct plugins enable neovim

# Disable a plugin
tinct plugins disable waybar
```

Configuration is saved to `.tinct-plugins.json`.

## Environment variables

Override plugin configuration with environment variables:

```bash
# Only enable specific plugins
export TINCT_ENABLED_PLUGINS="image,hyprland,kitty"

# Disable specific plugins
export TINCT_DISABLED_PLUGINS="waybar,dunst"
```

## Configuration file

Plugin settings are stored in `.tinct-plugins.json`:

```json
{
  "version": "1",
  "enabled_plugins": ["image", "hyprland", "kitty"],
  "disabled_plugins": ["waybar"],
  "external_plugins": {
    "wled": {
      "name": "wled",
      "path": "/home/user/.local/share/tinct/plugins/wled",
      "type": "output",
      "version": "0.1.0"
    }
  }
}
```

## Plugin paths

External plugins can be loaded from custom paths:

```bash
# Use a specific plugin binary
tinct generate -i image -p ~/wallpaper.jpg \
  -o my-custom-plugin \
  --plugin-path ~/projects/my-plugin/plugin
```

## Examples

### Set up official repository

```bash
# Add repository
tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/refs/heads/main/repository/repository.json

# Sync plugins
tinct plugins sync

# List available
tinct plugins list
```

### Focus on specific plugins

```bash
# Disable plugins you don't use
tinct plugins disable gnome-shell
tinct plugins disable kde-plasma

# Or use environment variable for session
export TINCT_ENABLED_PLUGINS="image,hyprland,kitty,waybar"
```

## See also

- [External plugins](/docs/plugins/external) - Creating custom plugins
- [Plugin repository](/docs/plugins/external/repository) - Repository system
