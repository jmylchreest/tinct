# Contributing to This Plugin Repository

Thank you for your interest in contributing to this Tinct plugin repository! This document provides guidelines for submitting plugins and improvements.

## Table of Contents

- [How to Contribute](#how-to-contribute)
- [Plugin Submission Guidelines](#plugin-submission-guidelines)
- [Plugin Requirements](#plugin-requirements)
- [Submission Process](#submission-process)
- [Code of Conduct](#code-of-conduct)

## How to Contribute

There are several ways to contribute:

1. **Submit a new plugin** - Add your plugin to the repository
2. **Update existing plugin** - Submit new versions or fixes
3. **Improve documentation** - Enhance README or guides
4. **Report issues** - Report broken links or incorrect information
5. **Suggest improvements** - Ideas for repository management

## Plugin Submission Guidelines

### Before Submitting

Ensure your plugin meets these criteria:

- [ ] Plugin is functional and tested
- [ ] Implements `--plugin-info` flag (required)
- [ ] Returns valid JSON from `--plugin-info`
- [ ] Reads input from stdin (if applicable)
- [ ] Writes output to stdout
- [ ] Handles errors gracefully
- [ ] Has proper documentation (README)
- [ ] Licensed (MIT, Apache, GPL, etc.)
- [ ] Uses semantic versioning (MAJOR.MINOR.PATCH)

### Quality Standards

Your plugin should:

- Be well-documented with usage examples
- Handle edge cases and errors
- Work across multiple platforms (when possible)
- Follow best practices for the language used
- Include tests (recommended)
- Have a clear purpose and description

## Plugin Requirements

### Mandatory: --plugin-info Flag

Your plugin MUST implement the `--plugin-info` flag that returns JSON:

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "type": "input|output",
  "description": "Short description of what the plugin does"
}
```

**Go Example:**
```go
if *pluginInfo {
    info := PluginInfo{
        Name:        "my-plugin",
        Version:     "1.0.0",
        Type:        "output",
        Description: "My awesome plugin",
    }
    json.NewEncoder(os.Stdout).Encode(info)
    return
}
```

**Python Example:**
```python
if "--plugin-info" in sys.argv:
    json.dump({
        "name": "my-plugin",
        "version": "1.0.0",
        "type": "output",
        "description": "My awesome plugin"
    }, sys.stdout)
    sys.exit(0)
```

**Bash Example:**
```bash
if [ "$1" = "--plugin-info" ]; then
    cat <<EOF
{
  "name": "my-plugin",
  "version": "1.0.0",
  "type": "output",
  "description": "My awesome plugin"
}
EOF
    exit 0
fi
```

### GitHub Releases

Your plugin must be distributed via GitHub Releases:

1. **Tag releases** using semantic versioning: `v1.0.0`, `v1.1.0`, etc.
2. **Upload binaries** for supported platforms
3. **Include checksums** (will be verified by maintainers)
4. **Write release notes** describing changes
5. **Link to documentation** in release

For compiled plugins (Go, Rust, etc.), we recommend using:
- [GoReleaser](https://goreleaser.com/) for Go
- [cargo-dist](https://github.com/axodotdev/cargo-dist) for Rust

## Submission Process

### Option 1: GitHub Issue (Easiest)

1. **Open an issue** titled "New Plugin: [plugin-name]"
2. **Provide information**:
   - Plugin name
   - Description (1-2 sentences)
   - Repository URL
   - Latest release tag
   - Supported platforms
   - Plugin type (input/output)
   - License
3. **Wait for review** - Maintainer will add your plugin

### Option 2: Pull Request (Preferred)

1. **Fork this repository**

2. **Add your plugin as a submodule**:
   ```bash
   git submodule add https://github.com/user/tinct-plugin-name plugins/plugin-name
   ```

3. **Update `repository.json`**:
   - Add your plugin entry under `plugins`
   - Include all version information
   - List all platform downloads with checksums
   - Update `last_updated` timestamp

4. **Generate checksums**:
   ```bash
   # Download your release assets
   curl -LO https://github.com/user/plugin/releases/download/v1.0.0/plugin_linux_amd64.tar.gz
   
   # Calculate SHA256
   sha256sum plugin_linux_amd64.tar.gz
   # Output: abc123... plugin_linux_amd64.tar.gz
   
   # Use in manifest as: "checksum": "sha256:abc123..."
   ```

5. **Test locally**:
   ```bash
   # Validate JSON
   jq empty repository.json
   
   # Test installation
   tinct plugins repo add test file:///path/to/your/fork/repository.json
   tinct plugins install your-plugin
   ```

6. **Submit pull request**:
   - Title: "Add [plugin-name] v[version]"
   - Description: Brief overview of what the plugin does
   - Checklist: Confirm requirements met

### Pull Request Template

```markdown
## Plugin Information

- **Name**: plugin-name
- **Version**: 1.0.0
- **Type**: input | output
- **Repository**: https://github.com/user/tinct-plugin-name
- **License**: MIT

## Description

Brief description of what the plugin does.

## Checklist

- [ ] Plugin implements `--plugin-info`
- [ ] Tested on multiple platforms (list which ones)
- [ ] GitHub release created with assets
- [ ] Checksums calculated and verified
- [ ] Documentation included (README)
- [ ] JSON manifest validated (`jq empty repository.json`)
- [ ] Plugin installed and tested locally
- [ ] Submodule added (if applicable)
- [ ] `last_updated` timestamp updated

## Testing

```bash
# Commands used to test
tinct plugins repo add test file:///path/to/fork/repository.json
tinct plugins install plugin-name
tinct generate -i image -o plugin-name test.png
```

## Additional Notes

Any additional information about the plugin, dependencies, or special requirements.
```

## Review Process

1. **Initial Review** (1-3 days)
   - Check plugin meets requirements
   - Verify checksums
   - Test installation on multiple platforms
   - Review code for security concerns

2. **Feedback** (if needed)
   - Requested changes or improvements
   - Questions about implementation

3. **Approval & Merge**
   - Plugin added to repository
   - Available to all users

4. **Announcement** (optional)
   - Plugin featured in repository README
   - Shared on social media/discussions

## Updating Existing Plugins

To update your plugin with a new version:

### Option 1: Issue
Open an issue titled "Update: [plugin-name] to v[new-version]" with:
- New version number
- Release URL
- Changelog summary

### Option 2: Pull Request
1. Update the submodule to new version
2. Add new version entry to manifest (at TOP of versions array)
3. Update checksums for all platforms
4. Submit PR titled "Update [plugin-name] to v[version]"

**Example manifest update:**
```json
"versions": [
  {
    "version": "1.1.0",  // NEW VERSION (first)
    "released": "2025-01-15T10:00:00Z",
    "downloads": { ... }
  },
  {
    "version": "1.0.0",  // OLD VERSION (after)
    "released": "2025-01-10T12:00:00Z",
    "downloads": { ... }
  }
]
```

## Platform Support

We encourage plugins that support multiple platforms:

**Minimum (required):**
- Linux amd64

**Recommended:**
- Linux amd64
- Linux arm64
- macOS amd64 (Intel)
- macOS arm64 (Apple Silicon)

**Nice to have:**
- Windows amd64
- FreeBSD amd64

**Script plugins** should work on all platforms (use "any" platform key).

## Security

### Security Review

All plugins undergo security review:
- Code review for obvious vulnerabilities
- Checksum verification
- License compatibility check
- No hardcoded credentials or API keys
- Safe handling of user input

### Reporting Security Issues

If you find a security issue in a plugin:
1. **Do not** open a public issue
2. Email the repository maintainer privately
3. Include: plugin name, vulnerability description, reproduction steps
4. Wait for response before public disclosure

## Documentation Standards

Your plugin should include:

### README.md
- Description and purpose
- Installation instructions
- Usage examples
- Configuration options
- Dependencies and requirements
- Troubleshooting guide
- License information

### Example:
```markdown
# Tinct Plugin: My Plugin

Short description of what it does.

## Installation

```bash
tinct plugins install my-plugin
```

## Usage

```bash
tinct generate -i image -o my-plugin wallpaper.png
```

## Configuration

This plugin accepts the following options via `--plugin-args`:

```json
{
  "my-plugin": {
    "option1": "value1",
    "option2": true
  }
}
```

## Requirements

- Python 3.8+
- python3-pillow

## License

MIT
```

## Version Management

### Semantic Versioning

Follow [semver](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes

### Version Retention

We typically keep:
- Latest version (always)
- Previous 2-3 minor versions
- Last patch of previous major versions

Very old versions may be removed to keep manifest clean.

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Accept constructive criticism
- Focus on what's best for the community
- Show empathy towards others

### Unacceptable Behavior

- Harassment or discriminatory language
- Trolling or insulting comments
- Publishing others' private information
- Spamming or excessive self-promotion

### Enforcement

Violations may result in:
1. Warning from maintainers
2. Temporary ban from repository
3. Permanent ban for severe/repeated violations

## Questions?

- Open an issue for general questions
- Join discussions: https://github.com/jmylchreest/tinct/discussions
- Email maintainer (see repository.json)

## License

By contributing, you agree that your contributions will be licensed under the same license as this repository (see LICENSE file).

Your plugin may use a different license - specify it in the plugin repository and manifest.

---

Thank you for contributing! 🎉