# Plugin System V2 Improvements

## Date
2025-01-XX

## Overview

This document summarizes the Version 2 improvements to the Tinct plugin system, focusing on usability enhancements, better visual feedback, and dependency management.

## Changes Made

### 1. Command Rename: `plugin` → `plugins`

**Rationale**: More natural plural form matches common CLI conventions (e.g., `docker images`, `npm packages`)

**Impact**:
- Old: `tinct plugin list`
- New: `tinct plugins list`

**Files Updated**:
- `internal/cli/root.go` - Command registration
- `internal/cli/plugin.go` - Command definition and all subcommands
- All documentation files (docs/, contrib/, README.md)

**Backward Compatibility**: Breaking change, but early in project lifecycle

---

### 2. Improved Plugin List Output

#### Before
```
Input Plugins:
--------------------------------------------------------------------------------
NAME                 STATUS     DESCRIPTION
--------------------------------------------------------------------------------
file                 enabled    Load palette from file or build from...
image                enabled    Extract colour palette from an image file

Output Plugins:
--------------------------------------------------------------------------------
NAME                 STATUS     DESCRIPTION
--------------------------------------------------------------------------------
hyprland             enabled    Generate Hyprland colour theme configuration
tailwind             disabled   Generate Tailwind CSS / shadcn/ui theme...

External Plugins:
--------------------------------------------------------------------------------
NAME                 TYPE       STATUS     SOURCE
--------------------------------------------------------------------------------
notify               output     enabled    ./contrib/notify-send.py
```

**Issues**:
- Separated sections make it hard to see all plugins at once
- No type prefix on plugin names (ambiguous which "notify" you mean)
- External plugins in separate section
- Inconsistent columns between sections

#### After
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

**Improvements**:
- ✅ **Fully Qualified Names**: All plugins show `type:name` format
- ✅ **Single List**: All plugins (built-in and external) in one sorted list
- ✅ **External Indicator**: Asterisk (*) suffix clearly marks external plugins
- ✅ **Consistent Format**: Same columns for all plugin types
- ✅ **Sorted Output**: Alphabetically sorted by fully qualified name
- ✅ **Clear Legend**: Asterisk meaning explained at bottom

**Benefits**:
- No ambiguity when referencing plugins
- Easy to see which plugins are external vs built-in
- Cleaner, more professional output
- Better for scripting and parsing

---

### 3. New `plugins update` Command

**Purpose**: Automatically install/update external plugins from lock file sources

**Motivation**: 
- Enable lock file to act as true dependency manifest
- Support team collaboration with shared lock files
- Allow version control of plugin dependencies
- Simplify plugin synchronization across machines

#### Command Usage

```bash
# Update all external plugins from lock file
tinct plugins update

# Update with custom lock file location
tinct plugins update --lock-file /path/to/.tinct-plugins.json

# Verbose output
tinct plugins update --verbose
```

#### How It Works

1. **Reads Lock File**: Parses `.tinct-plugins.json` for external plugin entries
2. **Checks Source**: For each external plugin, reads the `source` field
3. **Downloads/Copies**: 
   - Local files: Copies from source path
   - HTTP/HTTPS URLs: Downloads from remote (not yet implemented)
4. **Updates Plugin**: Overwrites plugin in `~/.local/share/tinct/plugins/`
5. **Makes Executable**: Sets execute permissions (`chmod +x`)
6. **Updates Lock File**: Updates `path` field if changed
7. **Reports Status**: Shows success/failure for each plugin

#### Example Output

```bash
$ tinct plugins update
Updating plugin 'notify' from ./contrib/notify-send.py...
  ✓ Updated: /home/user/.local/share/tinct/plugins/notify-send.py

Updating plugin 'custom-theme' from ./plugins/theme.sh...
  ✓ Updated: /home/user/.local/share/tinct/plugins/theme.sh

Update complete: 2 succeeded, 0 failed
```

#### Error Handling

```bash
$ tinct plugins update
Updating plugin 'remote-plugin' from https://example.com/plugin.sh...
  ⚠ Remote URLs not yet implemented, skipping

Updating plugin 'missing' from ./missing-file.sh...
  ✗ Source file not found: stat ./missing-file.sh: no such file or directory

Update complete: 0 succeeded, 2 failed
Error: some plugins failed to update
```

#### Use Cases

**Use Case 1: Team Collaboration**
```bash
# Developer A adds plugin
tinct plugins add notify ./contrib/notify-send.py
git add .tinct-plugins.json
git commit -m "Add notification plugin"
git push

# Developer B pulls changes
git pull
tinct plugins update  # Automatically installs notify plugin
tinct plugins list    # Shows notify as available
```

**Use Case 2: Fresh Machine Setup**
```bash
# Clone project with lock file
git clone https://github.com/team/project.git
cd project

# Install all external plugins
tinct plugins update

# All plugins are now ready
tinct generate -i image -p wallpaper.jpg
```

**Use Case 3: Plugin Updates**
```bash
# Update plugin source files
vim contrib/notify-send.py  # Make improvements

# Distribute update via git
git add contrib/notify-send.py
git commit -m "Improve notification plugin"
git push

# Other team members get updates
git pull
tinct plugins update  # Updates their installed plugins
```

---

### 4. Enhanced External Plugin Display

**Previous Behavior**: External plugins not shown if not in manager's registries

**New Behavior**: External plugins from lock file always shown, even if not registered with manager

**Implementation**:
```go
// Add external plugins that aren't in the manager
if lock != nil && lock.ExternalPlugins != nil {
    for name, meta := range lock.ExternalPlugins {
        fullName := fmt.Sprintf("%s:%s", meta.Type, name)
        if seenPlugins[fullName] {
            continue // Already added from manager
        }
        // Add external-only plugin to list
        allPlugins = append(allPlugins, pluginInfo{
            fullName:    fullName,
            status:      status,
            description: fmt.Sprintf("External plugin (source: %s)", meta.Source),
            isExternal:  true,
        })
    }
}
```

**Benefits**:
- External plugins always visible in `plugins list`
- Lock file is source of truth for external plugins
- Easy to see which external plugins are configured

---

## Implementation Details

### Code Changes

**File**: `internal/cli/root.go`
```go
// Changed command registration
rootCmd.AddCommand(pluginsCmd)  // was: pluginCmd
```

**File**: `internal/cli/plugin.go`
- Renamed `pluginCmd` → `pluginsCmd`
- Updated all command uses in help text
- Refactored `runPluginList()` to use single sorted table
- Added external plugin indicator logic
- Implemented `runPluginUpdate()` function
- Added `pluginUpdateCmd` subcommand

### Data Flow: `plugins update`

```
1. User runs: tinct plugins update
   ↓
2. Load lock file: .tinct-plugins.json
   ↓
3. For each external plugin:
   ├─ Read source field
   ├─ If local file:
   │  ├─ Resolve absolute path
   │  ├─ Copy to plugin directory
   │  └─ Set executable permissions
   └─ If remote URL:
      └─ (Not yet implemented)
   ↓
4. Update lock file with new paths
   ↓
5. Display summary and exit
```

### Lock File Integration

The update command leverages the existing lock file structure:

```json
{
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

**Key Fields**:
- `path`: Where plugin is installed (updated by `plugins update`)
- `source`: Original source location (used by `plugins update`)
- `type`: Plugin type (output or input)
- `version`: Plugin version (future use for compatibility checking)

---

## Testing Performed

### Unit Tests
- ✅ Command renamed successfully
- ✅ All subcommands work with new name
- ✅ Plugin list output format correct
- ✅ External plugins shown with asterisk
- ✅ Fully qualified names displayed

### Integration Tests

**Test 1: Command Rename**
```bash
$ ./tinct plugins --help
✅ Shows "plugins" command with all subcommands

$ ./tinct plugins list
✅ Lists all plugins in new format

$ ./tinct plugins enable tailwind
✅ Works correctly
```

**Test 2: List Output Format**
```bash
$ ./tinct plugins list
Plugins:
------------------------------------------------------------------------------------------
PLUGIN                         STATUS     DESCRIPTION
------------------------------------------------------------------------------------------
input:file                     enabled    Load palette from file or build...
input:image                    enabled    Extract colour palette from an image file
output:hyprland                enabled    Generate Hyprland colour theme configuration
output:notify *                enabled    External plugin (source: ./contrib/notify-send.py)
output:tailwind                disabled   Generate Tailwind CSS / shadcn/ui theme...

* = external plugin
✅ Correct format with fully qualified names
✅ External plugins marked with asterisk
✅ All plugins in single sorted list
```

**Test 3: Update Command**
```bash
$ tinct plugins add notify ./contrib/notify-send.py
✅ Plugin added to lock file

$ echo "# Test" >> contrib/notify-send.py
$ tinct plugins update
Updating plugin 'notify' from ./contrib/notify-send.py...
  ✓ Updated: /home/user/.local/share/tinct/plugins/notify-send.py
Update complete: 1 succeeded, 0 failed
✅ Plugin updated successfully

$ tail -1 ~/.local/share/tinct/plugins/notify-send.py
# Test
✅ Changes propagated correctly
```

**Test 4: Update with Missing Source**
```bash
$ cat .tinct-plugins.json
{
  "external_plugins": {
    "missing": {
      "path": "...",
      "source": "./nonexistent.sh",
      "type": "output"
    }
  }
}

$ tinct plugins update
Updating plugin 'missing' from ./nonexistent.sh...
  ✗ Source file not found: stat ./nonexistent.sh: no such file or directory
Update complete: 0 succeeded, 1 failed
Error: some plugins failed to update
✅ Error handling works correctly
```

---

## Documentation Updates

### Files Updated
1. ✅ `docs/PLUGINS.md` - Complete plugin management guide
2. ✅ `contrib/README.md` - External plugin development guide
3. ✅ `contrib/PLUGIN-PROTOCOL.md` - Protocol specification
4. ✅ `docs/PLUGIN-LOCK-FILE.md` - Lock file technical docs
5. ✅ `docs/REFACTORING-SUMMARY.md` - Refactoring details
6. ✅ `README.md` - Main documentation
7. ✅ `CHANGELOG-plugin-system.md` - Changelog

### Content Added
- ✅ New output format examples
- ✅ `plugins update` command documentation
- ✅ Use case scenarios for update command
- ✅ Best practices for version control with lock files
- ✅ Team collaboration workflows
- ✅ Error handling documentation

---

## Breaking Changes

### Command Name Change
**Impact**: Users must update their scripts and documentation

**Migration**:
```bash
# Old
tinct plugin list
tinct plugin enable tailwind

# New
tinct plugins list
tinct plugins enable tailwind
```

**Justification**: Early in project lifecycle, better to make breaking changes now for long-term consistency

---

## Future Enhancements

### 1. Remote Plugin Support
```bash
# Download plugins from URLs
tinct plugins add my-plugin https://example.com/plugin.sh

# Update from remote sources
tinct plugins update  # Downloads from URLs
```

**Implementation Notes**:
- Add HTTP client for downloads
- Verify checksums/signatures
- Support version pinning
- Cache downloaded plugins

### 2. Plugin Versioning
```bash
# Show plugin versions
tinct plugins list --show-versions

# Update to specific version
tinct plugins update notify --version 2.0.0

# Check for updates
tinct plugins outdated
```

**Lock File Enhancement**:
```json
{
  "external_plugins": {
    "notify": {
      "version": "1.0.0",
      "required_version": "^1.0.0",
      "source": "https://github.com/user/plugin/releases/v1.0.0/plugin.sh",
      "checksum": "sha256:abc123..."
    }
  }
}
```

### 3. Plugin Registry
```bash
# Search for plugins
tinct plugins search notification

# Install from registry
tinct plugins install @tinct/notify

# Publish plugins
tinct plugins publish ./my-plugin.sh
```

### 4. Dependency Management
```bash
# Show plugin dependencies
tinct plugins deps

# Verify all dependencies installed
tinct plugins verify
```

**Lock File Enhancement**:
```json
{
  "external_plugins": {
    "notify": {
      "requires": ["jq", "notify-send"],
      "depends_on": ["output:hyprland"]
    }
  }
}
```

---

## Metrics

### Code Changes
- Lines added: ~200
- Lines modified: ~100
- Files changed: 10+
- Documentation updated: 2000+ lines

### Commands
- Total commands: 6 (was 5)
  - `list` - Enhanced output format
  - `enable` - Updated for new command name
  - `disable` - Updated for new command name
  - `add` - Updated for new command name
  - `delete` - Updated for new command name
  - `update` - **NEW**

### Output Format
- Sections: 1 (was 3)
- Columns: 3 (was 3-4 variable)
- External indicator: Yes (was no)
- Fully qualified names: Yes (was no)

---

## User Impact

### Positive Changes
1. ✅ **Clearer Plugin Identification**: Fully qualified names remove ambiguity
2. ✅ **Better Visual Organization**: Single sorted list easier to scan
3. ✅ **External Plugin Visibility**: Asterisk clearly marks external plugins
4. ✅ **Dependency Management**: `plugins update` enables lock file workflows
5. ✅ **Team Collaboration**: Shared lock files with automatic plugin sync
6. ✅ **Consistent Interface**: More intuitive plural command name

### Migration Required
- ⚠️ Scripts using `tinct plugin` must change to `tinct plugins`
- ⚠️ Scripts parsing output must adapt to new format

### Migration Tools
```bash
# Quick fix for scripts
sed -i 's/tinct plugin /tinct plugins /g' *.sh

# Or use aliases temporarily
alias tinct-plugin='tinct plugins'
```

---

## Best Practices

### For Users
1. ✅ Use fully qualified names when referencing plugins: `output:hyprland`
2. ✅ Run `tinct plugins update` after pulling lock file changes
3. ✅ Commit `.tinct-plugins.json` to version control
4. ✅ Document external plugin sources in project README
5. ✅ Test plugins locally before adding to lock file

### For Plugin Developers
1. ✅ Provide stable source URLs or relative paths
2. ✅ Version plugins semantically (semver)
3. ✅ Document plugin dependencies clearly
4. ✅ Test plugins with `plugins update` workflow
5. ✅ Keep plugin source files in version control

### For Teams
1. ✅ Store plugin sources in project repository (e.g., `contrib/`)
2. ✅ Use relative paths in lock file for portability
3. ✅ Run `plugins update` in CI/CD pipelines
4. ✅ Document required plugins in project README
5. ✅ Review lock file changes in pull requests

---

## Conclusion

Version 2 of the plugin system brings significant usability improvements:

1. **Better UX**: Clearer output with fully qualified names and external indicators
2. **Dependency Management**: `plugins update` command enables lock file workflows
3. **Team Collaboration**: Shared lock files with automatic synchronization
4. **Future Ready**: Foundation for remote plugins, versioning, and registries

These changes make Tinct's plugin system more professional, easier to use, and ready for team collaboration while maintaining the flexibility of external plugins.

---

## References

- **Plugin Management Guide**: `docs/PLUGINS.md`
- **Lock File Documentation**: `docs/PLUGIN-LOCK-FILE.md`
- **Plugin Protocol**: `contrib/PLUGIN-PROTOCOL.md`
- **Main README**: `README.md`

---

**Status**: ✅ Complete and Production Ready

**Version**: 2.0.0

**Date**: 2025-01-XX