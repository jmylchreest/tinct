---
sidebar_position: 4
---

# Plugin repositories

Manage and share external plugins.

## Adding repositories

Add a plugin repository from GitHub or any Git URL:

```bash
# Add from GitHub
tinct plugins add https://github.com/user/tinct-plugins

# Add with custom name
tinct plugins add https://github.com/user/tinct-plugins --name community

# Add from other Git hosts
tinct plugins add https://gitlab.com/user/tinct-plugins
```

## Syncing repositories

Update all registered repositories:

```bash
tinct plugins sync
```

This fetches the latest plugin manifests from all registered repositories.

## Installing plugins

After syncing, install plugins by name:

```bash
# Install a specific plugin
tinct plugins install wled

# Install multiple plugins
tinct plugins install wled openrgb keylightd
```

## Listing plugins

```bash
# List all available plugins (built-in and external)
tinct plugins

# List only external plugins
tinct plugins --external

# Show details for a specific plugin
tinct plugins info wled
```

## Removing plugins

```bash
# Remove an installed plugin
tinct plugins remove wled

# Remove a repository
tinct plugins remove-repo community
```

## Repository structure

A plugin repository should contain a `manifest.json`:

```json
{
  "name": "community-plugins",
  "version": "1.0.0",
  "description": "Community tinct plugins",
  "plugins": [
    {
      "name": "wled",
      "type": "output",
      "version": "1.0.0",
      "description": "WLED LED strip controller",
      "url": "https://github.com/user/tinct-output-wled/releases/download/v1.0.0",
      "binaries": {
        "linux-amd64": "tinct-output-wled-linux-amd64",
        "linux-arm64": "tinct-output-wled-linux-arm64",
        "darwin-amd64": "tinct-output-wled-darwin-amd64",
        "darwin-arm64": "tinct-output-wled-darwin-arm64"
      }
    }
  ]
}
```

## Publishing plugins

### Create a release

1. Build your plugin for multiple platforms:

```bash
GOOS=linux GOARCH=amd64 go build -o tinct-output-myplugin-linux-amd64
GOOS=linux GOARCH=arm64 go build -o tinct-output-myplugin-linux-arm64
GOOS=darwin GOARCH=amd64 go build -o tinct-output-myplugin-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o tinct-output-myplugin-darwin-arm64
```

2. Create a GitHub release with the binaries

3. Update your `manifest.json` with the new version

### Standalone installation

For plugins not in a repository, provide an installation script:

```bash
#!/bin/bash
# install.sh

PLUGIN_NAME="tinct-output-myplugin"
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
URL="https://github.com/user/${PLUGIN_NAME}/releases/download/v${VERSION}/${PLUGIN_NAME}-${OS}-${ARCH}"

mkdir -p "$INSTALL_DIR"
curl -sL "$URL" -o "$INSTALL_DIR/$PLUGIN_NAME"
chmod +x "$INSTALL_DIR/$PLUGIN_NAME"

echo "Installed $PLUGIN_NAME to $INSTALL_DIR"
```

## Security

### Checksum verification

Repositories can include checksums for plugin binaries:

```json
{
  "plugins": [
    {
      "name": "wled",
      "binaries": {
        "linux-amd64": {
          "url": "...",
          "sha256": "abc123..."
        }
      }
    }
  ]
}
```

Tinct verifies checksums when installing.

### Trusted repositories

Only install plugins from sources you trust. Review plugin code before running, especially for plugins that:

- Access network resources
- Modify system files
- Require elevated permissions

## See also

- [Creating plugins](/docs/plugins/external/creating) - Write your own plugins
- [Protocols](/docs/plugins/external/protocols) - Communication details
