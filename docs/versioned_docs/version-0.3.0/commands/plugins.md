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
| `search` | Search repositories for plugins |
| `install` | Install a plugin from a repository |
| `add` | Add an external plugin from a local file, HTTP URL, or Git repo |
| `delete` | Remove an external plugin |
| `update` | Update all external plugins from their sources |
| `sync` | Install missing plugins from the manifest |
| `clean` | Remove plugins not in the manifest |
| `repo add` | Add a plugin repository |
| `repo list` | List configured repositories |
| `repo delete` | Remove a repository |
| `repo update` | Refresh repository manifests |
| `repo info` | Show repository details |

## tinct plugins list

Show all available plugins:

```bash
tinct plugins list
```

Output shows:

- Plugin name and type (input/output)
- External plugins marked with asterisk (*)
- Version and description

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

## tinct plugins search

Search configured repositories for plugins:

```bash
tinct plugins search [query]
```

| Flag | Description |
|------|-------------|
| `--type` | Filter by plugin type (`input` or `output`) |
| `--tag` | Filter by tag (can be repeated) |
| `--author` | Filter by author |
| `--repo` | Search only in a specific repository |

Examples:

```bash
tinct plugins search random           # Search for "random"
tinct plugins search --type input     # List all input plugins
tinct plugins search --tag terminal   # Find terminal-related plugins
```

## tinct plugins install

Install a plugin from a configured repository:

```bash
tinct plugins install <plugin-name> [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--repository` | | Install from a specific repository |
| `--version` | `latest` | Plugin version to install |
| `-f, --force` | `false` | Force reinstall if already installed |

Examples:

```bash
tinct plugins install random                         # Install latest from any repo
tinct plugins install random --repository official   # From specific repo
tinct plugins install random --version 0.0.2         # Specific version
```

The plugin is:

1. Downloaded from the repository
2. Verified (checksum)
3. Installed to `~/.local/share/tinct/plugins/`
4. Registered in `~/.config/tinct/plugins.manifest.json`

## tinct plugins update

Update all external plugins from their original sources:

```bash
tinct plugins update
```

This checks each external plugin for newer versions and downloads updates. Plugins installed from repositories respect version pinning (use `--version latest` during install to always get the latest).

## tinct plugins sync

Install missing plugins from the manifest:

```bash
tinct plugins sync [flags]
```

| Flag | Description |
|------|-------------|
| `-f, --force` | Force reinstall all plugins |
| `--verify` | Verify checksums of existing plugins |
| `--skip-missing` | Continue on errors instead of stopping |

Useful for restoring plugins on a new machine from a shared `plugins.manifest.json`.

## Manifest

Plugin installation data is stored in `~/.config/tinct/plugins.manifest.json`:

```json
{
  "external_plugins": {
    "wled": {
      "name": "wled",
      "path": "/home/user/.local/share/tinct/plugins/wled",
      "type": "output",
      "version": "0.1.0",
      "source": {
        "type": "repository",
        "repository": "official",
        "plugin": "wled",
        "version": "0.1.0"
      }
    }
  }
}
```

Use the `--manifest-file` flag to specify an alternative manifest
path. The legacy `--lock-file` flag still works as a hidden alias.

A pre-rename `plugins.lock.json` is migrated to the new path
automatically on first read.

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

## See also

- [Plugin development](/docs/plugin-development) - Creating custom plugins
- [Publishing plugins](/docs/plugin-development/publishing) - Repository system
