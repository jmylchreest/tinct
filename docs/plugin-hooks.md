# Plugin Hooks System

Tinct's output plugins support optional pre-execution and post-execution hooks that enable advanced functionality such as environment validation, configuration reloading, and service management.

## Overview

The hook system provides two optional interfaces that plugins can implement:

1. **PreExecuteHook**: Validate environment and skip plugin if prerequisites aren't met
2. **PostExecuteHook**: Perform actions after successful file generation

Both hooks are completely optional - plugins work fine without implementing them.

## PreExecuteHook Interface

### Purpose

The `PreExecuteHook` interface allows plugins to validate their environment before attempting to generate files. This is useful for:

- Checking if required executables exist on `$PATH`
- Verifying configuration directories exist
- Validating environment prerequisites
- Gracefully skipping execution when conditions aren't met

### Interface Definition

```go
type PreExecuteHook interface {
    // PreExecute runs before Generate(). Returns:
    //   - skip: if true, plugin is skipped (not an error, just bypassed)
    //   - reason: human-readable explanation for skipping
    //   - error: actual error that should stop execution
    PreExecute(ctx context.Context) (skip bool, reason string, err error)
}
```

### Return Values

- **skip**: `true` to skip the plugin gracefully, `false` to continue
- **reason**: Human-readable message explaining why the plugin was skipped
- **error**: Only return an error for actual failures; return `skip=true` for expected conditions

### Example Implementation

```go
func (p *KittyPlugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
    // Check if kitty executable exists on PATH
    _, err = exec.LookPath("kitty")
    if err != nil {
        return true, "kitty executable not found on $PATH", nil
    }

    // Check if config directory exists
    configDir := p.DefaultOutputDir()
    if _, err := os.Stat(configDir); os.IsNotExist(err) {
        return true, fmt.Sprintf("kitty config directory not found: %s", configDir), nil
    }

    return false, "", nil
}
```

### Best Practices

1. **Use skip for expected conditions**: If kitty isn't installed, that's expected - return `skip=true`, not an error
2. **Return errors for unexpected failures**: System errors, permission issues, etc. should return actual errors
3. **Provide helpful reason messages**: Users should understand why a plugin was skipped
4. **Respect the context**: Check for cancellation: `if ctx.Err() != nil { return true, "cancelled", ctx.Err() }`
5. **Keep it fast**: Pre-hooks should complete quickly (< 1 second)

### Context Timeout

The CLI sets a 5-second timeout for pre-execution hooks. If your hook takes longer, it will be cancelled.

## PostExecuteHook Interface

### Purpose

The `PostExecuteHook` interface allows plugins to perform actions after successful file generation. Common use cases:

- Reload application configuration
- Send signals to running processes (SIGHUP, SIGUSR1)
- Restart services via systemctl
- Notify users of changes (desktop notifications)
- Trigger dependent processes

### Interface Definition

```go
type PostExecuteHook interface {
    // PostExecute runs after successful Generate() and file writing.
    // The writtenFiles slice contains the paths that were written.
    // Errors are logged but don't fail the overall operation.
    PostExecute(ctx context.Context, writtenFiles []string) error
}
```

### Parameters

- **ctx**: Context for cancellation and timeout (10-second timeout)
- **writtenFiles**: Slice of absolute paths to files that were successfully written

### Example Implementation

```go
func (p *KittyPlugin) PostExecute(ctx context.Context, writtenFiles []string) error {
    if !p.reloadConfig {
        return nil // Feature not enabled
    }

    // Send SIGUSR1 to all running kitty instances to reload config
    cmd := exec.CommandContext(ctx, "killall", "-SIGUSR1", "kitty")
    if err := cmd.Run(); err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            if exitErr.ExitCode() == 1 {
                return fmt.Errorf("no running kitty instances found to reload")
            }
        }
        return fmt.Errorf("failed to reload kitty config: %w", err)
    }

    return nil
}
```

### Best Practices

1. **Make it optional**: Use flags to enable post-execution actions (`--kitty.reload`)
2. **Handle missing processes gracefully**: Not running ≠ failure
3. **Respect dry-run mode**: Post-hooks are NOT called in dry-run mode
4. **Use context-aware commands**: `exec.CommandContext(ctx, ...)` respects cancellation
5. **Return descriptive errors**: Users should know what failed and why
6. **Don't panic**: Errors are logged but won't fail the generation process

### Context Timeout

The CLI sets a 10-second timeout for post-execution hooks. Operations should complete within this time.

## Error Handling

### PreExecute Errors

- **skip=true, error=nil**: Plugin is gracefully skipped, logged in verbose mode
- **skip=false, error!=nil**: Plugin execution is stopped, error is displayed
- **skip=true, error!=nil**: Error is logged but plugin is still skipped

### PostExecute Errors

- Errors are logged with a warning symbol () but don't fail the overall operation
- Other plugins continue to execute
- Generation is still considered successful

## Integration with CLI

### Execution Flow

1. Validate plugin configuration
2. **Run PreExecute hook** (if implemented)
   - If skip=true, move to next plugin
   - If error!=nil, display error and continue to next plugin
3. Generate files
4. Write files to disk (unless dry-run)
5. **Run PostExecute hook** (if implemented and not dry-run)
   - If error!=nil, log warning but continue

### Execution Order

Hooks are executed in a specific order to ensure proper coordination:

1. **Global pre-generate hook** (`~/.config/tinct/hooks/pre-generate.sh`)
2. **All plugin PreExecute hooks** (validation phase)
3. **All plugin Generate() calls** (file generation)
4. **All plugin PostExecute hooks** (for successful plugins only)
5. **Global post-generate hook** (`~/.config/tinct/hooks/post-generate.sh`)

This ensures that:
- All validation happens before any generation
- All files are written before any post-processing
- Plugins that were skipped don't run post-hooks
- Global hooks can coordinate across all plugins

### User Feedback

```bash
# Verbose mode shows all hooks
$ tinct generate -i image -p wallpaper.jpg --outputs kitty --verbose
Running global pre-generate hook: /home/user/.config/tinct/hooks/pre-generate.sh
   Starting theme generation...
Skipping kitty: kitty executable not found on $PATH
Running global post-generate hook: /home/user/.config/tinct/hooks/post-generate.sh
   Theme generation complete!

# Post-hook success (verbose mode)
$ tinct generate -i image -p wallpaper.jpg --outputs kitty --kitty.reload --verbose
Running global pre-generate hook: /home/user/.config/tinct/hooks/pre-generate.sh
 Output plugin: kitty
   Generate Kitty terminal colour theme configuration
   /home/user/.config/kitty/tinct.conf (2190 bytes)
   kitty post-hook completed
Running global post-generate hook: /home/user/.config/tinct/hooks/post-generate.sh

# Post-hook failure (always shown)
$ tinct generate -i image -p wallpaper.jpg --outputs kitty --kitty.reload
   /home/user/.config/kitty/tinct.conf (2190 bytes)
   kitty post-hook failed: no running kitty instances found to reload
```

## Global Hook Scripts

Tinct supports global hook scripts that run before and after all plugin execution. These are useful for:
- Backing up existing configurations
- Coordinating multiple applications
- Sending notifications
- Logging theme changes
- Custom pre/post processing

### Location

Global hooks are shell scripts located at:
- `~/.config/tinct/hooks/pre-generate.sh` - Runs before any plugin execution
- `~/.config/tinct/hooks/post-generate.sh` - Runs after all plugins complete

### Requirements

- Scripts must be executable (`chmod +x`)
- Scripts should exit with code 0 for success
- Scripts are executed with a 30-second timeout

### Environment Variables

Global hooks receive these environment variables:
- `TINCT_HOOK` - Hook name ("pre-generate" or "post-generate")
- `TINCT_VERSION` - Tinct version

### Example Pre-Generate Hook

```bash
#!/bin/sh
# ~/.config/tinct/hooks/pre-generate.sh

echo " Backing up existing themes..."

# Backup kitty theme
if [ -f ~/.config/kitty/tinct.conf ]; then
    cp ~/.config/kitty/tinct.conf ~/.config/kitty/tinct.conf.backup
fi

# Backup hyprland theme
if [ -f ~/.config/hypr/tinct-colours.conf ]; then
    cp ~/.config/hypr/tinct-colours.conf ~/.config/hypr/tinct-colours.conf.backup
fi

echo " Backup complete"
```

### Example Post-Generate Hook

```bash
#!/bin/sh
# ~/.config/tinct/hooks/post-generate.sh

echo " Theme generation complete!"

# Send desktop notification
if command -v notify-send >/dev/null 2>&1; then
    notify-send "Tinct" "Theme updated successfully" --icon=applications-graphics
fi

# Update wallpaper if pywal is installed
if command -v wal >/dev/null 2>&1; then
    echo "Updating wallpaper colors..."
    wal -i ~/Pictures/current-wallpaper.jpg -q
fi

# Restart dependent services
systemctl --user restart waybar 2>/dev/null || true
```

### Dry-Run Behavior

Global hooks are **not executed** in dry-run mode (`--dry-run`).

## Built-in Plugin Hooks

### Kitty Plugin

**PreExecute:**
- Checks for `kitty` executable on `$PATH`
- Verifies config directory exists

**PostExecute:**
- Controlled by `--kitty.reload` flag
- Sends `SIGUSR1` signal to all running kitty instances
- Causes kitty to reload configuration without restart

**Usage:**
```bash
tinct generate -i image -p wallpaper.jpg --outputs kitty --kitty.reload
```

### Hyprland Plugin

**PreExecute:**
- Checks for `hyprctl` executable on `$PATH`
- Verifies hyprland is currently running
- Checks themes directory exists

**PostExecute:**
- Controlled by `--hyprland.reload` flag
- Runs `hyprctl reload` to reload configuration
- Full config reload including theme changes

**Usage:**
```bash
tinct generate -i image -p wallpaper.jpg --outputs hyprland --hyprland.reload
```

## Creating Custom Hooks

### Step 1: Import Required Packages

```go
import (
    "context"
    "os/exec"
    
    "github.com/jmylchreest/tinct/internal/plugin/output"
)
```

### Step 2: Implement the Interface(s)

```go
type MyPlugin struct {
    outputDir    string
    reloadConfig bool
}

func (p *MyPlugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
    // Check prerequisites
    _, err = exec.LookPath("myapp")
    if err != nil {
        return true, "myapp not found on $PATH", nil
    }
    
    return false, "", nil
}

func (p *MyPlugin) PostExecute(ctx context.Context, writtenFiles []string) error {
    if !p.reloadConfig {
        return nil
    }
    
    // Reload configuration
    cmd := exec.CommandContext(ctx, "myapp-reload")
    return cmd.Run()
}
```

### Step 3: Register Flags

```go
func (p *MyPlugin) RegisterFlags(cmd *cobra.Command) {
    cmd.Flags().StringVar(&p.outputDir, "myapp.output-dir", "", "Output directory")
    cmd.Flags().BoolVar(&p.reloadConfig, "myapp.reload", false, "Reload config after generation")
}
```

## Testing Hooks

### Unit Testing PreExecute

```go
func TestPlugin_PreExecute_MissingExecutable(t *testing.T) {
    plugin := New()
    
    ctx := context.Background()
    skip, reason, err := plugin.PreExecute(ctx)
    
    if err != nil {
        t.Errorf("PreExecute should not return error for missing executable: %v", err)
    }
    
    if !skip {
        t.Error("PreExecute should skip when executable is missing")
    }
    
    if reason == "" {
        t.Error("PreExecute should provide a reason for skipping")
    }
}
```

### Unit Testing PostExecute

```go
func TestPlugin_PostExecute_Disabled(t *testing.T) {
    plugin := New()
    plugin.reloadConfig = false
    
    ctx := context.Background()
    err := plugin.PostExecute(ctx, []string{"/tmp/test.conf"})
    
    if err != nil {
        t.Errorf("PostExecute should not error when disabled: %v", err)
    }
}
```

### Integration Testing

```bash
# Test pre-hook skipping
$ tinct generate -i image -p test.jpg --outputs myapp --myapp.output-dir /nonexistent --verbose
Skipping myapp: config directory not found: /nonexistent

# Test post-hook execution
$ tinct generate -i image -p test.jpg --outputs myapp --myapp.reload
 Done! Generated 1 output plugin(s)
```

## Advanced Patterns

### Conditional Skipping

```go
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
    // Skip in CI environments
    if os.Getenv("CI") == "true" {
        return true, "skipping in CI environment", nil
    }
    
    // Skip if running as root (safety check)
    if os.Geteuid() == 0 {
        return true, "refusing to run as root", nil
    }
    
    return false, "", nil
}
```

### Multi-Process Reloading

```go
func (p *Plugin) PostExecute(ctx context.Context, writtenFiles []string) error {
    if !p.reloadAll {
        return nil
    }
    
    // Find all running instances
    cmd := exec.CommandContext(ctx, "pgrep", "-x", "myapp")
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("no running instances found")
    }
    
    // Send reload signal to each
    pids := strings.Split(strings.TrimSpace(string(output)), "\n")
    for _, pid := range pids {
        syscall.Kill(syscall.Atoi(pid), syscall.SIGHUP)
    }
    
    return nil
}
```

### Notification Integration

```go
func (p *Plugin) PostExecute(ctx context.Context, writtenFiles []string) error {
    // Send desktop notification
    cmd := exec.CommandContext(ctx, "notify-send",
        "Theme Updated",
        fmt.Sprintf("Generated %d theme files", len(writtenFiles)))
    
    // Don't fail if notification fails
    _ = cmd.Run()
    
    return nil
}
```

## Troubleshooting

### Hook Not Running

1. Check if hook is implemented: `var _ output.PreExecuteHook = (*Plugin)(nil)`
2. Verify verbose output: `--verbose` shows hook execution
3. Check timeout: Hooks have strict timeouts (5s pre, 10s post)

### PreExecute Always Skips

1. Check executable path: `which myapp`
2. Verify directory permissions: `ls -la ~/.config/myapp`
3. Test in verbose mode to see the reason

### PostExecute Fails

1. Check if process is running: `pgrep myapp`
2. Verify reload command works manually: `myapp-reload`
3. Check for permission issues: some reload actions need sudo

### Context Cancellation

If operations take too long, context will be cancelled. Handle gracefully:

```go
select {
case <-ctx.Done():
    return ctx.Err()
default:
    // Continue operation
}
```

## Security Considerations

1. **Validate user input**: Never pass unsanitized paths to shell commands
2. **Use exec.Command, not shell**: Avoid shell injection vulnerabilities
3. **Check permissions**: Don't require root unless absolutely necessary
4. **Fail safely**: Errors should not leave system in bad state
5. **Timeout operations**: Always use context-aware operations

## External Plugin Hooks

External plugins (scripts/binaries) can also implement hooks by supporting special command-line flags.

### PreExecute Hook

External plugins can implement pre-execution checks by handling the `--pre-execute` flag:

```bash
#!/bin/bash
# my-plugin.sh

if [ "$1" = "--pre-execute" ]; then
    # Check prerequisites
    if ! command -v myapp >/dev/null 2>&1; then
        echo "myapp not found"
        exit 1  # Exit code 1 = skip plugin
    fi
    
    exit 0  # Exit code 0 = continue
fi

# Normal plugin execution
# Read palette JSON from stdin
palette=$(cat)
# Generate output...
```

**Exit codes:**
- `0` - Continue with plugin execution
- `1` - Skip plugin (reason from stdout)
- `2+` - Error (message from stderr)

### PostExecute Hook

External plugins can implement post-execution actions by handling the `--post-execute` flag:

```bash
#!/bin/bash
# my-plugin.sh

if [ "$1" = "--post-execute" ]; then
    # Read written files from stdin
    files=$(cat)
    
    # Reload application
    if pgrep -x myapp >/dev/null; then
        pkill -SIGHUP myapp
        echo "Reloaded myapp"
    fi
    
    exit 0
fi

# Normal plugin execution...
```

The post-execute hook receives JSON on stdin:
```json
{
  "written_files": [
    "/home/user/.config/myapp/theme.conf",
    "/home/user/.config/myapp/colors.conf"
  ]
}
```

### Full External Plugin Example

```bash
#!/bin/bash
# advanced-plugin.sh

# Pre-execution hook
if [ "$1" = "--pre-execute" ]; then
    if ! command -v myapp >/dev/null 2>&1; then
        echo "myapp not installed"
        exit 1
    fi
    
    if [ ! -d ~/.config/myapp ]; then
        echo "myapp config directory not found"
        exit 1
    fi
    
    exit 0
fi

# Post-execution hook
if [ "$1" = "--post-execute" ]; then
    # Reload all running instances
    if pgrep -x myapp >/dev/null; then
        pkill -SIGUSR1 myapp
        echo "Sent reload signal to myapp"
    else
        echo "No running instances to reload" >&2
        exit 1
    fi
    
    exit 0
fi

# Plugin info
if [ "$1" = "--info" ]; then
    cat <<EOF
{
  "name": "myapp",
  "description": "Generate MyApp theme configuration",
  "version": "1.0.0"
}
EOF
    exit 0
fi

# Normal execution - generate theme
palette=$(cat)

# Extract colors and generate config
echo "$palette" | jq -r '.colours.background.hex' > ~/.config/myapp/bg.txt
echo "$palette" | jq -r '.colours.foreground.hex' > ~/.config/myapp/fg.txt

echo "Generated theme files"
```

### Timeout

External plugin hooks have the same timeouts as built-in plugins:
- PreExecute: 5 seconds
- PostExecute: 10 seconds

## See Also

- [Plugin Development Guide](../contrib/README.md)
- [Output Plugin Interface](../internal/plugin/output/plugin.go)
- [Built-in Plugins](../internal/plugin/output/)
- [External Plugin Protocol](../contrib/README.md#plugin-protocol)