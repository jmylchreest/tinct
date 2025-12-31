---
sidebar_position: 5
---

# Publishing plugins

Share your plugins with the community.

## Repository structure

A plugin repository contains a manifest describing available plugins:

```
my-tinct-plugins/
├── repository.json      # Plugin manifest
└── README.md            # Repository documentation
```

### Manifest format

```json
{
  "name": "my-plugins",
  "version": "1.0.0",
  "description": "My custom tinct plugins",
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
      },
      "checksums": {
        "linux-amd64": "sha256:abc123...",
        "linux-arm64": "sha256:def456..."
      }
    }
  ]
}
```

### Binary naming

Follow this convention for cross-platform binaries:

```
{plugin-name}-{os}-{arch}
```

Supported platforms:
- `linux-amd64`, `linux-arm64`
- `darwin-amd64`, `darwin-arm64`
- `windows-amd64`

## Building for multiple platforms

### Go plugins

```bash
#!/bin/bash
# build-all.sh

PLUGIN_NAME="tinct-output-myplugin"
VERSION="1.0.0"

platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    output="${PLUGIN_NAME}-${GOOS}-${GOARCH}"

    echo "Building ${output}..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "dist/${output}"
done

# Generate checksums
cd dist && sha256sum * > checksums.txt
```

### Python plugins

For Python plugins, provide a wrapper script:

```bash
#!/bin/bash
# tinct-output-myplugin

# Find Python interpreter
PYTHON="${PYTHON:-python3}"

# Run the actual plugin
exec "$PYTHON" "$(dirname "$0")/myplugin.py" "$@"
```

## Creating releases

### GitHub releases

1. Create a new release on GitHub
2. Upload binaries for each platform
3. Update `repository.json` with new version and URLs

Example release workflow:

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin]
        goarch: [amd64, arm64]

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags="-s -w" -o dist/tinct-output-myplugin-${{ matrix.goos }}-${{ matrix.goarch }}

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: binary-${{ matrix.goos }}-${{ matrix.goarch }}
          path: dist/

  release:
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Download artifacts
        uses: actions/download-artifact@v4
        with:
          path: dist
          merge-multiple: true

      - name: Create checksums
        run: |
          cd dist
          sha256sum * > checksums.txt

      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
```

## Standalone installation

For plugins not in a repository, provide an installation script:

```bash
#!/bin/bash
# install.sh

set -e

PLUGIN_NAME="tinct-output-myplugin"
VERSION="${1:-1.0.0}"
INSTALL_DIR="${HOME}/.config/tinct/plugins"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
esac

# Download URL
URL="https://github.com/user/${PLUGIN_NAME}/releases/download/v${VERSION}/${PLUGIN_NAME}-${OS}-${ARCH}"

echo "Installing ${PLUGIN_NAME} v${VERSION} for ${OS}/${ARCH}..."

# Create plugin directory
mkdir -p "$INSTALL_DIR"

# Download and install
curl -sL "$URL" -o "$INSTALL_DIR/$PLUGIN_NAME"
chmod +x "$INSTALL_DIR/$PLUGIN_NAME"

echo "Installed to $INSTALL_DIR/$PLUGIN_NAME"

# Verify installation
"$INSTALL_DIR/$PLUGIN_NAME" --plugin-info
```

## Security considerations

### Checksum verification

Always provide checksums for binaries:

```json
{
  "plugins": [{
    "name": "myplugin",
    "binaries": {
      "linux-amd64": "tinct-output-myplugin-linux-amd64"
    },
    "checksums": {
      "linux-amd64": "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
  }]
}
```

Tinct verifies checksums when installing from repositories.

### Code signing

For additional security, consider signing your releases:

```bash
# Sign with GPG
gpg --armor --detach-sign dist/tinct-output-myplugin-linux-amd64

# Users can verify
gpg --verify tinct-output-myplugin-linux-amd64.asc
```

### Trusted repositories

Advise users to only install plugins from sources they trust. Document what your plugin does and what permissions it requires.

## Submitting to official repository

To add your plugin to the official tinct-plugins repository:

1. Fork [jmylchreest/tinct-plugins](https://github.com/jmylchreest/tinct-plugins)
2. Add your plugin to `repository.json`
3. Submit a pull request with:
   - Plugin description
   - Build/release workflow
   - Link to source repository

Requirements:
- Open source under a compatible licence
- Follows naming conventions
- Provides binaries for common platforms
- Includes checksums

## Documentation

Provide clear documentation for your plugin:

```markdown
# tinct-output-myplugin

Tinct output plugin for MyApp.

## Installation

```bash
tinct plugins add https://github.com/user/tinct-plugins
tinct plugins sync
tinct plugins install myplugin
```

## Usage

```bash
tinct generate -i image -p wallpaper.jpg -o myplugin
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--myplugin.option` | `value` | Description |

## Configuration

Add to your `~/.config/myapp/config`:

```
include ~/.config/myapp/tinct.conf
```
```

## See also

- [Creating plugins](/docs/plugin-development/creating) - Plugin development
- [tinct plugins command](/docs/commands/plugins) - Plugin management
