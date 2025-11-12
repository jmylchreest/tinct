# Contributing to the Official Plugin Repository

Thank you for your interest in contributing to the Tinct official plugin repository!

## Adding Your Plugin

There are two ways to add a plugin to this repository:

### 1. Submit a Pull Request

Submit a PR to add your plugin to `repository.json`. Use the repository manager tool to help:

```bash
# Build the tool
go build -o tinct-repo-manager ./cmd/tinct-repo-manager

# Add your plugin
./tinct-repo-manager add \
  --plugin my-plugin \
  --type output \
  --url "https://github.com/you/tinct-plugin-my-plugin/releases/download/v1.0.0/my-plugin_1.0.0_Linux_x86_64.tar.gz" \
  --platform linux_amd64 \
  --version 1.0.0 \
  --manifest contrib/repository/repository.json

# Validate
./tinct-repo-manager validate --manifest contrib/repository/repository.json

# Commit and create PR
git add contrib/repository/repository.json
git commit -m "Add my-plugin v1.0.0"
git push origin add-my-plugin
```

### 2. Open an Issue

If you prefer, open an issue with:
- Plugin name
- Plugin type (input or output)
- Description
- Repository URL
- Latest release tag
- Platform support

We'll add it for you!

## Plugin Requirements

To be included in the official repository, plugins must:

1. **Implement `--plugin-info`** - Return valid JSON with:
   ```json
   {
     "name": "plugin-name",
     "type": "input|output",
     "version": "1.0.0",
     "protocol_version": "0.0.1",
     "description": "What the plugin does",
     "author": "Your Name",
     "license": "MIT"
   }
   ```

2. **Be publicly accessible** - Hosted on GitHub or similar with public releases

3. **Have a license** - Open source license (MIT, Apache, GPL, etc.)

4. **Work correctly** - Test your plugin thoroughly before submission

5. **Follow security best practices** - No known vulnerabilities or malicious code

6. **Have documentation** - Include a README with usage instructions

## Plugin Naming

- Plugins should be named `tinct-plugin-<name>`
- Choose a descriptive, unique name
- Avoid generic names (e.g., `output`, `theme`)

## Release Process

When you release a new version of your plugin:

1. Create a GitHub release with proper versioning (e.g., `v1.0.1`)
2. Include compiled binaries for supported platforms
3. Open a PR or issue to update the repository
4. We'll sync your new version automatically

## Platform Support

Provide builds for as many platforms as possible:
- `linux_amd64` - Linux x86-64 (most common)
- `linux_arm64` - Linux ARM64 (Raspberry Pi, etc.)
- `darwin_amd64` - macOS Intel
- `darwin_arm64` - macOS Apple Silicon
- `any` - Platform-independent scripts

## Script Plugins

Script-based plugins (Python, Bash, etc.) should:
- Be executable (`chmod +x`)
- Include proper shebang (`#!/usr/bin/env python3`)
- List runtime dependencies in metadata
- Be self-contained or document external dependencies

## Quality Standards

Plugins should:
- Have clear, concise descriptions
- Include usage examples in README
- Handle errors gracefully
- Provide helpful error messages
- Support `--help` flag
- Follow the [Plugin Protocol](../../docs/PLUGIN-PROTOCOLS.md)

## Questions?

- Open an issue for questions
- Join discussions in GitHub Discussions
- Check existing plugins for examples

## Code of Conduct

Be respectful and constructive. We want this to be a welcoming community for all contributors.

Thank you for contributing to Tinct!
