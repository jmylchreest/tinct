# Plugin Lock File Technical Documentation

## Overview

The plugin lock file (`.tinct-plugins.json`) is a JSON configuration file that persists plugin enable/disable state and external plugin metadata to disk. This document explains its structure, behavior, and implementation details.

## Purpose

The lock file serves three main purposes:

1. **Persistent Configuration**: Store plugin enable/disable preferences across sessions
2. **External Plugin Registry**: Track metadata for externally added plugins
3. **Project-Specific Settings**: Allow per-project plugin configurations

## File Structure

```json
{
  "enabled_plugins": [
    "output:hyprland",
    "output:notify",
    "input:image"
  ],
  "disabled_plugins": [
    "output:tailwind"
  ],
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

## Fields

### `enabled_plugins` (array of strings)

**Purpose**: Whitelist mode - only listed plugins are enabled.

**Format**: Each entry is either:
- `"plugin-name"` - matches any plugin type with that name
- `"type:plugin-name"` - matches specific type and name (e.g., `"output:hyprland"`)

**Behavior**:
- If this field exists and has entries, **only** these plugins are enabled
- All other plugins are automatically disabled
- This is the most restrictive mode
- Takes precedence over `disabled_plugins`

**Example**:
```json
{
  "enabled_plugins": ["output:hyprland", "input:image"]
}
```
Result: Only hyprland output and image input are enabled. All other plugins (including file input, tailwind output) are disabled.

### `disabled_plugins` (array of strings)

**Purpose**: Blacklist mode - listed plugins are disabled.

**Format**: Same as `enabled_plugins`:
- `"plugin-name"` - disables any plugin with that name
- `"type:plugin-name"` - disables specific type and name

**Behavior**:
- Only applies if `enabled_plugins` is empty or not present
- Listed plugins are disabled
- All other plugins use their default enabled state
- Less restrictive than whitelist mode

**Example**:
```json
{
  "disabled_plugins": ["output:tailwind"]
}
```
Result: Tailwind is disabled. All other plugins use their defaults (hyprland enabled, image enabled, file enabled).

### `external_plugins` (object/map)

**Purpose**: Registry of externally added plugins with their metadata.

**Format**: Object with plugin names as keys, metadata objects as values:

```json
{
  "plugin-name": {
    "path": "absolute/path/to/plugin",
    "type": "output",
    "source": "original/source/path",
    "version": "1.0.0"
  }
}
```

**Metadata Fields**:
- `path` (string, required): Absolute path to the executable plugin file
- `type` (string, required): Plugin type - `"input"` or `"output"`
- `source` (string, required): Original source location (file path or URL)
- `version` (string, optional): Plugin version if available

**Behavior**:
- Automatically populated by `tinct plugins add`
- Used to track which plugins are external vs built-in
- Path is used to locate and execute the plugin
- Cleaned up by `tinct plugins delete`

## File Location Priority

Tinct searches for the lock file in this order:

1. **Explicit Path**: If `--lock-file /path/to/file` is specified
2. **Current Directory**: `./.tinct-plugins.json`
3. **Home Directory**: `~/.tinct-plugins.json`

**First match wins**: Once found, Tinct uses that file and stops searching.

### Location Strategies

**Project-Specific Configuration**:
```bash
cd my-project
tinct plugins enable tailwind  # Creates ./tinct-plugins.json
```

**Global Configuration**:
```bash
cd ~
tinct plugins disable tailwind  # Creates ~/.tinct-plugins.json
```

**Explicit Configuration**:
```bash
tinct plugins list --lock-file /etc/tinct/plugins.json
```

## Priority System

The complete priority hierarchy for determining if a plugin is enabled:

```
1. Lock File (highest priority)
   ├─ If enabled_plugins exists and has entries → WHITELIST MODE
   │  └─ Plugin in list? → ENABLED
   │  └─ Plugin not in list? → DISABLED
   │
   └─ If enabled_plugins is empty/absent → Check disabled_plugins
      └─ Plugin in disabled_plugins? → DISABLED
      └─ Plugin not in disabled_plugins? → Check Environment

2. Environment Variables
   ├─ If TINCT_ENABLED_PLUGINS is set → WHITELIST MODE
   │  └─ Plugin in list? → ENABLED
   │  └─ Plugin not in list? → DISABLED
   │
   └─ If TINCT_DISABLED_PLUGINS is set → BLACKLIST MODE
      └─ Plugin in list? → DISABLED
      └─ Plugin not in list? → Check Plugin Default

3. Plugin Default (lowest priority)
   └─ Return plugin.Enabled() value
```

### Key Priority Rules

1. **Lock file overrides environment variables**
2. **Environment variables override plugin defaults**
3. **Whitelist mode (enabled_plugins) overrides blacklist mode (disabled_plugins)**
4. **Type-specific overrides work**: `"output:tailwind"` is more specific than `"tailwind"`

## Implementation Details

### Data Structures

The Go struct definition:

```go
type PluginLock struct {
    EnabledPlugins  []string                       `json:"enabled_plugins,omitempty"`
    DisabledPlugins []string                       `json:"disabled_plugins,omitempty"`
    ExternalPlugins map[string]ExternalPluginMeta  `json:"external_plugins,omitempty"`
}

type ExternalPluginMeta struct {
    Path    string `json:"path"`
    Type    string `json:"type"`
    Source  string `json:"source"`
    Version string `json:"version,omitempty"`
}
```

### Loading Process

```go
func loadPluginLock() (*PluginLock, string, error) {
    // 1. Determine lock file path (CLI flag > current dir > home dir)
    // 2. Read file contents
    // 3. Parse JSON
    // 4. Return lock and path
}
```

### Manager Integration

```go
func createManagerFromLock(lock *PluginLock) *Manager {
    if lock == nil {
        return manager.NewFromEnv()
    }
    
    config := manager.Config{
        EnabledPlugins:  lock.EnabledPlugins,
        DisabledPlugins: lock.DisabledPlugins,
    }
    
    return manager.New(config)
}
```

The manager uses this config to determine plugin state via `IsInputEnabled()` and `IsOutputEnabled()` methods.

### Enable/Disable Logic

**Enable Plugin**:
```go
// 1. Remove from disabled_plugins if present
// 2. Add to enabled_plugins if not present
// 3. Save lock file
```

**Disable Plugin**:
```go
// 1. Remove from enabled_plugins if present
// 2. Add to disabled_plugins if not present
// 3. Save lock file
```

## Common Patterns

### Pattern 1: Fresh Start (No Lock File)

**State**: No `.tinct-plugins.json` exists

**Behavior**:
- Falls back to environment variables
- If no env vars, uses plugin defaults
- Built-in plugins use their `Enabled()` method
- Most built-in plugins default to `true`

### Pattern 2: Whitelist Mode

**Lock File**:
```json
{
  "enabled_plugins": ["output:hyprland", "input:image"]
}
```

**Behavior**:
- **Only** hyprland output and image input are enabled
- All other plugins are disabled (even if their default is enabled)
- Most restrictive mode

### Pattern 3: Blacklist Mode

**Lock File**:
```json
{
  "disabled_plugins": ["output:tailwind"]
}
```

**Behavior**:
- Tailwind is explicitly disabled
- All other plugins use their defaults
- More permissive than whitelist

### Pattern 4: Mixed Mode (Whitelist Wins)

**Lock File**:
```json
{
  "enabled_plugins": ["output:hyprland"],
  "disabled_plugins": ["output:tailwind"]
}
```

**Behavior**:
- Whitelist takes precedence
- Only hyprland is enabled
- disabled_plugins is **ignored** when enabled_plugins exists
- Tailwind is disabled (not in whitelist)

### Pattern 5: External Plugins

**Lock File**:
```json
{
  "disabled_plugins": [],
  "external_plugins": {
    "notify": {
      "path": "/home/user/.local/share/tinct/plugins/notify-send.py",
      "type": "output",
      "source": "./contrib/notify-send.py"
    }
  }
}
```

**Behavior**:
- External plugin is registered and tracked
- Plugin is enabled by default (not in disabled list)
- Path is used to locate plugin for execution

## Command Effects

### `tinct plugins enable <name>`

**Operations**:
1. Loads or creates lock file
2. Removes plugin from `disabled_plugins` (if present)
3. Adds plugin to `enabled_plugins`
4. Saves lock file

**Result**: Plugin is explicitly enabled (whitelist mode)

### `tinct plugins disable <name>`

**Operations**:
1. Loads or creates lock file
2. Removes plugin from `enabled_plugins` (if present)
3. Adds plugin to `disabled_plugins`
4. Saves lock file

**Result**: Plugin is explicitly disabled (blacklist mode if no enabled_plugins)

### `tinct plugins add <name> <path>`

**Operations**:
1. Loads or creates lock file
2. Copies plugin to `~/.local/share/tinct/plugins/`
3. Makes plugin executable
4. Adds entry to `external_plugins`
5. Saves lock file

**Result**: External plugin is registered and available

### `tinct plugins delete <name>`

**Operations**:
1. Loads lock file
2. Deletes plugin file from disk
3. Removes from `external_plugins`
4. Removes from `enabled_plugins` and `disabled_plugins`
5. Saves lock file

**Result**: External plugin is completely removed

## Edge Cases

### Empty Lock File

```json
{}
```

**Behavior**: Falls back to environment variables, then plugin defaults. Effectively same as no lock file.

### Only External Plugins

```json
{
  "external_plugins": {
    "notify": { "path": "...", "type": "output", "source": "..." }
  }
}
```

**Behavior**: External plugins are registered but all plugins (built-in and external) use their defaults for enabled state.

### Malformed Plugin Names

**Lock File**:
```json
{
  "enabled_plugins": ["invalid:format:name"]
}
```

**Behavior**: Plugin won't match anything. Effectively disables all plugins (whitelist mode with no valid entries).

### Duplicate Entries

**Lock File**:
```json
{
  "enabled_plugins": ["hyprland", "output:hyprland", "hyprland"]
}
```

**Behavior**: All forms are valid. Plugin matching is liberal - matches either plain name or type:name. Duplicates are harmless.

## Testing Considerations

### Unit Testing with Lock Files

**Approach 1: Temporary Files**
```go
func TestWithLockFile(t *testing.T) {
    tmpFile := filepath.Join(t.TempDir(), ".tinct-plugins.json")
    // Write test lock file
    // Run tests with --lock-file tmpFile
}
```

**Approach 2: In-Memory Config**
```go
func TestPluginManager(t *testing.T) {
    config := manager.Config{
        EnabledPlugins: []string{"output:hyprland"},
    }
    mgr := manager.New(config)
    // Test manager behavior
}
```

### Integration Testing

Use lock files to set up known states:
```bash
# Test whitelist mode
echo '{"enabled_plugins":["output:hyprland"]}' > .tinct-plugins.json
./tinct plugins list  # Should show only hyprland enabled

# Test blacklist mode
echo '{"disabled_plugins":["output:tailwind"]}' > .tinct-plugins.json
./tinct plugins list  # Should show tailwind disabled
```

## Best Practices

### Version Control

**Do**: Commit `.tinct-plugins.json` for reproducible project builds
```bash
git add .tinct-plugins.json
git commit -m "Add project plugin configuration"
```

**Don't**: Commit personal/machine-specific paths in external_plugins
```json
{
  "external_plugins": {
    "my-plugin": {
      "path": "/home/alice/.local/share/tinct/plugins/plugin.py"  // ❌ Don't commit
    }
  }
}
```

### Global vs Project Configuration

**Global** (`~/.tinct-plugins.json`):
- Personal preferences
- Globally disabled plugins
- System-wide external plugins

**Project** (`./.tinct-plugins.json`):
- Project requirements
- Team-shared configuration
- Build reproducibility

### Documentation

Document your lock file in project README:
```markdown
## Tinct Configuration

This project uses Tinct with the following plugins:
- `hyprland` - Window manager theme generation
- `notify` - Build notifications

Plugin configuration is stored in `.tinct-plugins.json`.
```

## Troubleshooting

### Problem: Unexpected Plugin State

**Diagnosis**:
```bash
# Check which lock file is being used
tinct plugins list --verbose

# Check lock file contents
cat .tinct-plugins.json
cat ~/.tinct-plugins.json

# Check environment
echo $TINCT_ENABLED_PLUGINS
echo $TINCT_DISABLED_PLUGINS
```

### Problem: Lock File Not Found

**Solution**: Create explicitly:
```bash
tinct plugins enable <plugin-name>  # Auto-creates lock file
# or
echo '{}' > .tinct-plugins.json
```

### Problem: Lock File Conflicts

**Scenario**: Project lock file conflicts with global lock file.

**Solution**: Project lock file takes precedence (current directory searched first). Remove project lock file to use global:
```bash
rm .tinct-plugins.json
```

## Summary

The plugin lock file is a simple but powerful JSON configuration system that:

1. **Persists** plugin preferences to disk
2. **Supports** both whitelist and blacklist modes
3. **Tracks** external plugin metadata
4. **Integrates** with environment variables and plugin defaults
5. **Enables** project-specific and global configurations

The priority system ensures predictable behavior: lock file > environment > defaults, with whitelist mode being the most restrictive.