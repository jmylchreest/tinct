# Tinct Official Plugin Repository

Official plugin repository for Tinct - discover and install plugins for system-wide theming.

## Using This Repository

### Add the Repository

```bash
tinct plugins repo add official \
  https://raw.githubusercontent.com/jmylchreest/tinct/main/contrib/repository/repository.json
```

### Search and Install Plugins

```bash
# Search for plugins
tinct plugins search

# Install a plugin
tinct plugins install random

# Install specific version
tinct plugins install random --version 1.2.3
```

## Available Plugins

This repository contains:

### Input Plugins
- **random** - Generate random colour palettes with configurable seed and colour count

### Output Plugins
- **wob** - Wayland Overlay Bar theme generation
- **templater** - Template-based configuration file generator
- **dunstify** - Desktop notifications via dunstify or notify-send
- **notify-send** - Desktop notifications with colour palette information (script)
- **openrgb-peripheral** - OpenRGB peripheral lighting control (script)
- **wled-ambient** - WLED ambient monitor lighting (script)
- **example-minimal** - Minimal example plugin for development (script)

## For Repository Maintainers

This repository is maintained using `tinct-repo-manager`. See the [Repository Manager Guide](../../docs/REPOSITORY-MANAGER.md) for details.

### Quick Start

```bash
# Build the repo manager
go build -o tinct-repo-manager ./cmd/tinct-repo-manager

# Sync from configuration file (recommended)
./tinct-repo-manager sync \
  --config contrib/repository/sync-config.jsonl \
  --manifest contrib/repository/repository.json \
  --min-protocol-version 0.0.1 \
  --prune \
  --prune-remove-after 720h \
  --verbose

# Or sync from specific GitHub release
./tinct-repo-manager sync \
  --github jmylchreest/tinct \
  --version latest \
  --plugin-filter "tinct-plugin-*" \
  --exclude "tinct_*" \
  --manifest contrib/repository/repository.json \
  --min-protocol-version 0.0.1 \
  --prune \
  --prune-remove-after 720h

# List plugins
./tinct-repo-manager list \
  --manifest contrib/repository/repository.json

# Validate
./tinct-repo-manager validate \
  --manifest contrib/repository/repository.json
```

### Sync Configuration

The repository uses `sync-config.jsonl` to define all sync sources. This declarative configuration includes:
- GitHub releases (all versions of compiled plugins)
- Direct URLs to script-based plugins
- Automatic version detection from `--plugin-info` output

Example configuration:
```jsonl
{"type": "github", "repo": "jmylchreest/tinct", "version": "all", "filter": ["tinct-plugin-*"], "exclude": ["tinct_*"]}
{"type": "url", "url": "https://raw.githubusercontent.com/jmylchreest/tinct/main/contrib/plugins/output/tinct-plugin-notify-send.py", "plugin": "notify-send", "plugin_type": "output", "version": "-", "platform": "any", "runtime": "python3"}
```

Configuration features:
- **Version auto-detection**: Use `"version": "-"` to automatically detect version from plugin's `--plugin-info` output
- **Protocol filtering**: Use `--min-protocol-version 0.0.1` command flag to skip plugins with older protocol versions
- **Cascade filtering**: If version X fails protocol check, all older versions are automatically skipped
- **Metadata extraction**: All plugin information (type, description, author, license, tags) is automatically extracted from `--plugin-info`
- **Config placeholders**: Values like `plugin_type` in the config are validation placeholders; actual values come from `--plugin-info`

## Repository Structure

```json
{
  "version": "1",
  "name": "tinct-official",
  "description": "Official Tinct plugin repository",
  "url": "https://github.com/jmylchreest/tinct",
  "maintained_by": "Tinct Contributors",
  "last_updated": "2025-01-12T00:00:00Z",
  "plugins": {
    "plugin-name": {
      "name": "plugin-name",
      "type": "input|output",
      "description": "Plugin description",
      "repository": "https://github.com/user/repo",
      "author": "Author Name",
      "license": "MIT",
      "tags": ["tag1", "tag2"],
      "versions": [
        {
          "version": "1.0.0",
          "released": "2025-01-12T00:00:00Z",
          "compatibility": ">=1.0.0",
          "changelog_url": "https://github.com/user/repo/releases/tag/v1.0.0",
          "downloads": {
            "linux_amd64": {
              "url": "https://github.com/user/repo/releases/download/v1.0.0/plugin_linux_amd64.tar.gz",
              "checksum": "sha256:abc123...",
              "size": 2048576,
              "available": true
            }
          }
        }
      ]
    }
  }
}
```

## Automated Updates

This repository is automatically updated via GitHub Actions:
- **On Release**: When a new Tinct release is published
- **Weekly**: Prunes unavailable entries every Sunday at 2am UTC
- **On Tool Changes**: When the repo-manager tool is updated

The workflow:
1. Downloads the latest `tinct-repo-manager` binary from GitHub releases (or builds from source as fallback)
2. Executes `sync --config sync-config.jsonl --min-protocol-version 0.0.1` with the configuration
3. Validates the manifest
4. Commits changes if any plugins were updated

## Contributing

Want to add your plugin to this repository? See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

This repository manifest is released under the MIT License. Individual plugins may have different licenses - check each plugin's metadata.
