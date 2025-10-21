# Plugin Sources Documentation

## Overview

Tinct supports multiple source types for adding and updating external plugins. This allows you to install plugins from local files, HTTP URLs, or Git repositories.

## Table of Contents

- [Source Types](#source-types)
- [Local Files](#local-files)
- [HTTP/HTTPS URLs](#httphttps-urls)
- [Git Repositories](#git-repositories)
- [Source Format Reference](#source-format-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Source Types

Tinct automatically detects the source type based on the format:

| Source Type | Detection | Example |
|-------------|-----------|---------|
| **Local File** | File path (relative or absolute) | `./plugin.sh` |
| **HTTP/HTTPS** | Starts with `http://` or `https://` | `https://example.com/plugin.sh` |
| **Git Repository** | Contains `.git` suffix or GitHub/GitLab/Bitbucket domain | `https://github.com/user/repo.git` |

## Local Files

### Description

Install plugins from files on your local filesystem. Useful for:
- Development and testing
- Plugins stored in project repositories
- Private plugins not published online

### Format

```
/path/to/plugin.sh              # Absolute path
./relative/path/plugin.sh       # Relative path
../other-dir/plugin.py          # Parent directory
~/my-plugins/plugin.sh          # Home directory
```

### Examples

```bash
# Add plugin from local file
tinct plugins add my-plugin ./contrib/my-plugin.sh

# Add plugin from absolute path
tinct plugins add system-plugin /usr/local/share/tinct/plugins/system.sh

# Add plugin from home directory
tinct plugins add personal-plugin ~/my-tinct-plugins/personal.py
```

### Lock File Entry

```json
{
  "external_plugins": {
    "my-plugin": {
      "path": "/home/user/.local/share/tinct/plugins/my-plugin.sh",
      "type": "output",
      "source": "./contrib/my-plugin.sh",
      "version": "1.0.0"
    }
  }
}
```

### How It Works

1. Resolves relative path to absolute path
2. Verifies file exists
3. Copies file to plugin directory (`~/.local/share/tinct/plugins/`)
4. Sets execute permissions (`chmod +x`)
5. Stores original source path in lock file

### Update Behavior

When running `tinct plugins update`, the plugin is re-copied from the original source path. This allows:
- Keeping plugins in sync with source files
- Testing changes to plugin code
- Version control of plugin sources

## HTTP/HTTPS URLs

### Description

Download plugins directly from web servers. Useful for:
- Published plugins hosted on static servers
- GitHub raw file URLs
- CDN-hosted plugins
- Quick installation from URLs

### Format

```
https://example.com/path/to/plugin.sh
http://server.local/plugins/plugin.py
https://raw.githubusercontent.com/user/repo/main/plugin.sh
```

### Examples

```bash
# Add plugin from HTTPS URL
tinct plugins add web-plugin https://example.com/plugins/notify.sh

# Add plugin from GitHub raw URL
tinct plugins add github-plugin https://raw.githubusercontent.com/user/tinct-plugin/main/plugin.sh

# Add plugin from CDN
tinct plugins add cdn-plugin https://cdn.example.com/tinct/plugins/theme.py
```

### Lock File Entry

```json
{
  "external_plugins": {
    "web-plugin": {
      "path": "/home/user/.local/share/tinct/plugins/notify.sh",
      "type": "output",
      "source": "https://example.com/plugins/notify.sh",
      "version": "1.0.0"
    }
  }
}
```

### How It Works

1. Downloads file via HTTP GET request
2. Verifies HTTP 200 OK response
3. Saves to plugin directory
4. Extracts filename from URL or uses plugin name
5. Sets execute permissions
6. Stores URL in lock file

### Update Behavior

When running `tinct plugins update`, the plugin is re-downloaded from the URL. This allows:
- Fetching latest version of plugins
- Automatic updates when source changes
- Team synchronization via shared URLs

### Security Considerations

⚠️ **Important**: Only download plugins from trusted sources
- Verify HTTPS certificates are valid
- Check source reputation before installing
- Review plugin code before use
- Use specific version URLs when possible

## Git Repositories

### Description

Clone plugins from Git repositories. Useful for:
- Versioned plugin distribution
- Open-source plugins on GitHub/GitLab
- Private Git repositories
- Plugin development workflows

### Format

#### Basic Git URL
```
https://github.com/user/repo.git
https://gitlab.com/user/repo.git
git@github.com:user/repo.git
```

#### With File Path
```
https://github.com/user/repo.git:path/to/plugin.sh
https://github.com/user/repo.git:bin/plugin.py
```

The `:path/to/file` syntax specifies which file in the repository to use as the plugin.

### Examples

#### Example 1: Auto-Detect Plugin File

```bash
# Tinct will look for common plugin files in the repository
tinct plugins add repo-plugin https://github.com/user/tinct-plugin.git
```

**Auto-detection searches for:**
1. `<plugin-name>` (exact match)
2. `<plugin-name>.sh`
3. `<plugin-name>.py`
4. `plugin.sh`
5. `plugin.py`
6. `bin/<plugin-name>`
7. `bin/<plugin-name>.sh`
8. `bin/<plugin-name>.py`

#### Example 2: Specify File Path

```bash
# Explicitly specify which file to use
tinct plugins add notify https://github.com/user/tinct-plugins.git:notify-send.py

# Plugin in subdirectory
tinct plugins add theme https://github.com/user/tinct-plugins.git:plugins/theme.sh

# Plugin in bin directory
tinct plugins add custom https://github.com/user/tinct-plugins.git:bin/custom-plugin
```

#### Example 3: GitHub Repository

```bash
# Add plugin from GitHub
tinct plugins add github-notify https://github.com/tinct-contrib/notify-send.git

# With specific file
tinct plugins add github-notify https://github.com/tinct-contrib/notify-send.git:notify.py
```

#### Example 4: Private Repository

```bash
# SSH URL for private repos (requires SSH key setup)
tinct plugins add private-plugin git@github.com:company/private-tinct-plugins.git:plugin.sh
```

### Lock File Entry

```json
{
  "external_plugins": {
    "repo-plugin": {
      "path": "/home/user/.local/share/tinct/plugins/plugin.sh",
      "type": "output",
      "source": "https://github.com/user/tinct-plugin.git:plugin.sh",
      "version": "1.0.0"
    }
  }
}
```

### How It Works

1. Creates temporary directory
2. Clones repository with `git clone --depth 1` (shallow clone)
3. Extracts specified file (or auto-detects)
4. Copies file to plugin directory
5. Sets execute permissions
6. Cleans up temporary directory
7. Stores Git URL in lock file

### Update Behavior

When running `tinct plugins update`, the repository is re-cloned and the plugin file is extracted again. This allows:
- Fetching latest commits from repository
- Automatic updates when source changes
- Team synchronization via Git
- Version control integration

### Requirements

- Git must be installed and in PATH
- Network access to Git server
- SSH keys configured for private repositories

### Security Considerations

✅ **Recommended**:
- Use HTTPS URLs for public repositories
- Use SSH URLs for private repositories with key authentication
- Clone from trusted sources only
- Review repository contents before use
- Pin to specific tags/commits for stability (future feature)

## Source Format Reference

### Quick Reference Table

| Format | Example | Type |
|--------|---------|------|
| `./path/file.sh` | `./contrib/plugin.sh` | Local |
| `/absolute/path` | `/usr/local/share/plugin.sh` | Local |
| `~/home/path` | `~/my-plugins/plugin.sh` | Local |
| `https://url` | `https://example.com/plugin.sh` | HTTP |
| `http://url` | `http://localhost/plugin.sh` | HTTP |
| `*.git` | `https://github.com/user/repo.git` | Git |
| `github.com/*` | `https://github.com/user/repo` | Git |
| `gitlab.com/*` | `https://gitlab.com/user/repo` | Git |
| `git@*` | `git@github.com:user/repo.git` | Git |
| `*.git:path` | `https://github.com/user/repo.git:plugin.sh` | Git + Path |

### File Path Syntax (Git)

```
<git-url>:<file-path>
```

- `<git-url>`: Any valid Git repository URL
- `:`: Separator (required when specifying path)
- `<file-path>`: Path to plugin file within repository

**Examples**:
```
https://github.com/user/repo.git:plugin.sh
https://github.com/user/repo.git:src/plugins/theme.py
git@github.com:user/repo.git:bin/notify
```

## Examples

### Example 1: Local Development

```bash
# Add plugin from project
cd my-tinct-project
tinct plugins add dev-plugin ./plugins/dev.sh

# Edit plugin
vim plugins/dev.sh

# Update installed version
tinct plugins update

# Test
tinct generate -i image -p test.jpg -o dev-plugin
```

### Example 2: Team Collaboration

```bash
# Developer A: Add plugin to project
tinct plugins add team-notify ./contrib/notify.py
git add .tinct-plugins.json contrib/notify.py
git commit -m "Add notification plugin"
git push

# Developer B: Sync plugins
git pull
tinct plugins update  # Installs notify.py automatically
tinct plugins list    # Shows team-notify
```

### Example 3: Public Plugin from GitHub

```bash
# Install plugin from GitHub
tinct plugins add awesome-theme https://github.com/tinct-contrib/awesome-theme.git

# Plugin auto-detects theme.sh in repository
tinct plugins list
# output:awesome-theme *         enabled    External plugin (source: https://github.com...)

# Use plugin
tinct generate -i image -p wallpaper.jpg -o awesome-theme
```

### Example 4: Multi-File Repository

```bash
# Repository structure:
# tinct-plugins/
#   ├── notify.py
#   ├── theme.sh
#   └── util.sh

# Add specific plugins from same repo
tinct plugins add notify https://github.com/user/tinct-plugins.git:notify.py
tinct plugins add theme https://github.com/user/tinct-plugins.git:theme.sh

# Both plugins installed from same repository
tinct plugins list
```

### Example 5: HTTP Plugin with Updates

```bash
# Add plugin from web server
tinct plugins add web-plugin https://cdn.example.com/tinct/plugin-v1.sh

# Plugin author releases v2
# Update to fetch latest version
tinct plugins update

# Plugin automatically updated to v2
```

### Example 6: Private Git Repository

```bash
# Setup: Configure SSH key for GitHub
ssh-keygen -t ed25519 -C "your@email.com"
cat ~/.ssh/id_ed25519.pub  # Add to GitHub

# Add plugin from private repo
tinct plugins add private-plugin git@github.com:company/private-plugins.git:plugin.sh

# Works for team members with access
git clone <project>
tinct plugins update  # Requires SSH key
```

## Best Practices

### For Plugin Users

1. **Local Development**: Use local files during development
   ```bash
   tinct plugins add dev-plugin ./my-plugin.sh
   ```

2. **Production**: Use Git repositories or stable HTTP URLs
   ```bash
   tinct plugins add prod-plugin https://github.com/user/plugin.git:plugin.sh
   ```

3. **Version Control**: Commit `.tinct-plugins.json` to Git
   ```bash
   git add .tinct-plugins.json
   git commit -m "Add plugin configuration"
   ```

4. **Regular Updates**: Keep plugins up to date
   ```bash
   tinct plugins update
   ```

5. **Review Before Installing**: Check plugin source code
   ```bash
   # For Git repos
   git clone https://github.com/user/plugin.git
   cat plugin.git/plugin.sh  # Review code
   
   # Then install
   tinct plugins add plugin https://github.com/user/plugin.git
   ```

### For Plugin Authors

1. **GitHub Release**: Host plugins on GitHub
   ```bash
   # Repository structure
   tinct-plugin/
     ├── plugin.sh          # Auto-detected
     ├── README.md
     └── LICENSE
   ```

2. **Multiple Plugins**: Use subdirectories
   ```bash
   tinct-plugins/
     ├── notify/notify.py
     ├── theme/theme.sh
     └── README.md
   
   # Users install with path
   tinct plugins add notify https://github.com/user/tinct-plugins.git:notify/notify.py
   ```

3. **Version Tags**: Tag releases (future support)
   ```bash
   git tag v1.0.0
   git push --tags
   ```

4. **Documentation**: Include usage in README
   ```markdown
   ## Installation
   
   ```bash
   tinct plugins add my-plugin https://github.com/user/my-plugin.git
   ```
   ```

5. **Testing**: Provide test palette files
   ```bash
   # In repository
   test/sample-palette.json
   
   # In README
   cat test/sample-palette.json | ./plugin.sh
   ```

### For Teams

1. **Centralized Plugins**: Store plugins in project repository
   ```
   project/
     ├── contrib/plugins/
     │   ├── notify.py
     │   └── theme.sh
     └── .tinct-plugins.json
   ```

2. **Shared Lock File**: Commit and share
   ```bash
   git add .tinct-plugins.json
   git commit -m "Update plugin configuration"
   ```

3. **CI/CD Integration**: Update plugins in pipeline
   ```yaml
   # .github/workflows/build.yml
   - name: Install Tinct plugins
     run: tinct plugins update
   ```

4. **Documentation**: Document required plugins
   ```markdown
   ## Setup
   
   1. Clone repository
   2. Run `tinct plugins update`
   3. All plugins automatically installed
   ```

## Troubleshooting

### Problem: Git clone fails

**Error**: `failed to clone repository: exit status 128`

**Solutions**:
- Check Git is installed: `git --version`
- Verify URL is correct
- Check network connectivity
- For private repos, verify SSH key is configured

### Problem: HTTP download fails

**Error**: `failed to download plugin: HTTP 404`

**Solutions**:
- Verify URL is correct and accessible
- Check file exists at URL
- Try downloading manually: `curl -O <url>`
- Check for redirects or authentication

### Problem: Plugin file not found in Git repo

**Error**: `could not find plugin file in repository`

**Solutions**:
- Specify file path explicitly: `repo.git:path/to/plugin.sh`
- Check file exists in repository
- Ensure file name matches plugin name
- Use auto-detection search paths

### Problem: Permission denied

**Error**: `failed to make plugin executable: permission denied`

**Solutions**:
- Check plugin directory permissions
- Ensure you have write access to `~/.local/share/tinct/plugins/`
- Run with appropriate permissions

## Future Enhancements

### Planned Features

1. **Version Pinning**: Pin plugins to specific versions
   ```bash
   tinct plugins add plugin https://github.com/user/plugin.git@v1.0.0
   ```

2. **Checksums**: Verify plugin integrity
   ```json
   {
     "external_plugins": {
       "plugin": {
         "source": "https://example.com/plugin.sh",
         "checksum": "sha256:abc123..."
       }
     }
   }
   ```

3. **Plugin Registry**: Central plugin repository
   ```bash
   tinct plugins search notification
   tinct plugins install @tinct/notify
   ```

4. **Branch/Tag Support**: Install from specific Git refs
   ```bash
   tinct plugins add plugin https://github.com/user/plugin.git#develop
   tinct plugins add plugin https://github.com/user/plugin.git#v1.0.0
   ```

## See Also

- [Plugin Management Guide](PLUGINS.md)
- [Plugin Protocol Specification](../contrib/PLUGIN-PROTOCOL.md)
- [Plugin Lock File Documentation](PLUGIN-LOCK-FILE.md)
- [External Plugin Development](../contrib/README.md)