# Tinct Plugin Repository Template

This is a template for creating your own Tinct plugin repository. Use this as a starting point to host and distribute your plugins.

## Quick Start

### 1. Use This Template

Click "Use this template" on GitHub or:

```bash
git clone https://github.com/user/tinct-plugin-repository-template my-plugins
cd my-plugins
rm -rf .git
git init
git add .
git commit -m "Initial commit"
```

### 2. Customize Repository

Edit `repository.json`:
```json
{
  "version": "1",
  "name": "my-plugins",
  "description": "My awesome Tinct plugin repository",
  "url": "https://github.com/YOUR_USERNAME/YOUR_REPO",
  "maintained_by": "Your Name <your@email.com>",
  "last_updated": "2025-01-10T12:00:00Z",
  "plugins": {}
}
```

### 3. Push to GitHub

```bash
git remote add origin https://github.com/YOUR_USERNAME/YOUR_REPO.git
git branch -M main
git push -u origin main
```

### 4. Add Plugins

Use the helper script:

```bash
./scripts/add-plugin.sh my-plugin https://github.com/user/tinct-plugin-my-plugin output
```

Or manually edit `repository.json` following the examples below.

## Adding to Tinct

Users can add your repository with:

```bash
tinct plugins repo add my-repo https://raw.githubusercontent.com/YOUR_USERNAME/YOUR_REPO/main/repository.json
```

Then search and install plugins:

```bash
tinct plugins search
tinct plugins install plugin-name
```

## Directory Structure

```
.
├── repository.json          # Main manifest (edit this!)
├── plugins/                 # Git submodules for plugins (optional)
│   └── example-plugin/     # Add plugins as submodules
├── scripts/
│   ├── add-plugin.sh       # Helper to add new plugins
│   └── update-manifest.sh  # Auto-update manifest
├── docs/
│   └── plugins/            # Plugin documentation
├── CONTRIBUTING.md         # Contribution guidelines
└── README.md               # This file
```

## Repository Manifest Format

The `repository.json` file lists all available plugins:

```json
{
  "version": "1",
  "name": "my-plugins",
  "description": "My plugin repository",
  "url": "https://github.com/user/repo",
  "maintained_by": "Your Name <email>",
  "last_updated": "2025-01-10T12:00:00Z",
  
  "plugins": {
    "plugin-name": {
      "name": "plugin-name",
      "type": "input|output",
      "description": "Plugin description",
      "repository": "https://github.com/user/plugin-repo",
      "author": "Author Name",
      "license": "MIT",
      "tags": ["tag1", "tag2"],
      
      "versions": [
        {
          "version": "1.0.0",
          "released": "2025-01-10T12:00:00Z",
          "compatibility": ">=1.0.0",
          "changelog_url": "https://github.com/user/plugin/releases/tag/v1.0.0",
          
          "downloads": {
            "linux_amd64": {
              "url": "https://github.com/user/plugin/releases/download/v1.0.0/plugin_linux_amd64.tar.gz",
              "checksum": "sha256:abc123...",
              "size": 2048576
            },
            "darwin_amd64": {
              "url": "https://github.com/user/plugin/releases/download/v1.0.0/plugin_darwin_amd64.tar.gz",
              "checksum": "sha256:def456...",
              "size": 2150400
            }
          }
        }
      ]
    }
  }
}
```

## Adding a Plugin

### Method 1: Using the Script

```bash
./scripts/add-plugin.sh plugin-name https://github.com/user/tinct-plugin-name output
```

This will:
1. Add the plugin as a git submodule
2. Fetch the latest release information
3. Prompt you to update the manifest

### Method 2: Manual

1. **Add as Submodule** (optional):
   ```bash
   git submodule add https://github.com/user/plugin-repo plugins/plugin-name
   ```

2. **Get Release Info**:
   ```bash
   # Download release assets
   curl -LO https://github.com/user/plugin/releases/download/v1.0.0/plugin_linux_amd64.tar.gz
   
   # Calculate checksum
   sha256sum plugin_linux_amd64.tar.gz
   # Output: abc123... plugin_linux_amd64.tar.gz
   ```

3. **Update repository.json**:
   - Add plugin entry under `plugins`
   - Include all platform downloads
   - Use checksums with `sha256:` prefix
   - Update `last_updated` timestamp

4. **Commit and Push**:
   ```bash
   git add repository.json plugins/
   git commit -m "Add plugin-name v1.0.0"
   git push
   ```

## Platform Identifiers

Use these keys in the `downloads` section:

- `linux_amd64` - Linux x86-64
- `linux_arm64` - Linux ARM64
- `darwin_amd64` - macOS Intel
- `darwin_arm64` - macOS Apple Silicon
- `windows_amd64` - Windows 64-bit
- `any` - Platform-independent (scripts)

## Validation

Before pushing changes:

```bash
# Validate JSON syntax
jq empty repository.json

# View formatted output
jq . repository.json

# List all plugins
jq '.plugins | keys[]' repository.json
```

## Scripts

### add-plugin.sh

Helper script to add new plugins. Adds plugin as submodule and fetches release information.

Usage:
```bash
./scripts/add-plugin.sh <plugin-name> <plugin-url> [type]
```

Example:
```bash
./scripts/add-plugin.sh awesome https://github.com/user/tinct-plugin-awesome output
```

### update-manifest.sh

Updates the manifest with latest releases from plugin submodules.

Usage:
```bash
./scripts/update-manifest.sh
```

## Best Practices

1. **Always Include Checksums**: Required for security (SHA256 with `sha256:` prefix)
2. **Test Before Publishing**: Install plugins locally to verify
3. **Keep Versions Ordered**: Newest first, oldest last
4. **Document Changes**: Update CHANGELOG.md
5. **Use Semantic Versioning**: MAJOR.MINOR.PATCH format
6. **Link Changelogs**: Include `changelog_url` for each version
7. **Support Multiple Platforms**: When possible, provide builds for Linux, macOS, Windows

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines on:
- Submitting plugins
- Code review process
- Testing requirements
- Documentation standards

## Plugin Developer Guide

If you're a plugin developer wanting to be included:

1. **Create your plugin repository**
2. **Implement `--plugin-info` flag** (required)
3. **Create GitHub releases** with binaries
4. **Submit a PR** or issue with:
   - Plugin name and description
   - Repository URL
   - Latest release tag
   - Platform support

See the [Plugin Development Guide](https://github.com/jmylchreest/tinct/blob/main/contrib/README.md) for details.

## Testing

Test your repository locally:

```bash
# Add as local file path
tinct plugins repo add test file:///path/to/your/repo/repository.json

# Search and install
tinct plugins search
tinct plugins install your-plugin
```

## Automation

### GitHub Actions

Set up automated manifest updates with GitHub Actions:

```yaml
# .github/workflows/update-manifest.yml
name: Update Manifest
on:
  schedule:
    - cron: '0 0 * * *'  # Daily
  workflow_dispatch:

jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive
      
      - name: Update submodules
        run: git submodule update --remote
      
      - name: Update manifest
        run: ./scripts/update-manifest.sh
      
      - name: Commit changes
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add repository.json
          git diff --quiet && git diff --staged --quiet || \
            (git commit -m "Auto-update manifest" && git push)
```

## Hosting Options

- **GitHub** (recommended): Free, built-in CDN, easy raw file access
- **GitLab**: Free private repos, self-hosted option
- **GitHub Pages**: Custom domains, automatic updates
- **Custom Server**: Full control, authentication support

## Resources

- [Tinct Documentation](https://github.com/jmylchreest/tinct)
- [Plugin Repository Guide](https://github.com/jmylchreest/tinct/blob/main/docs/PLUGIN-REPOSITORY.md)
- [Hosting Guide](https://github.com/jmylchreest/tinct/blob/main/docs/HOSTING-PLUGIN-REPOSITORY.md)

## Support

- Report issues: https://github.com/YOUR_USERNAME/YOUR_REPO/issues
- Tinct issues: https://github.com/jmylchreest/tinct/issues
- Discussions: https://github.com/jmylchreest/tinct/discussions

## License

This template is released under the MIT License. See LICENSE for details.

Your repository and plugins may use different licenses - specify them in the manifest.