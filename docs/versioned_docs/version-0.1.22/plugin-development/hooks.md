---
sidebar_position: 4
---

# Lifecycle hooks

Optional hooks for environment validation and post-generation actions.

## Overview

Tinct's output plugins can implement two optional lifecycle hooks:

1. **PreExecuteHook** - Validate environment before generation
2. **PostExecuteHook** - Perform actions after file generation

Both hooks are completely optional. Plugins work without implementing them.

## PreExecuteHook interface

### Purpose

The `PreExecuteHook` interface allows plugins to validate their environment before generating files. Use cases:

- Check if required executables exist on `$PATH`
- Verify configuration directories exist
- Validate environment prerequisites
- Skip gracefully when target app isn't installed

### Interface definition

```go
type PreExecuteHook interface {
    // PreExecute runs before Generate(). Returns:
    //   - skip: if true, plugin is skipped (not an error, just bypassed)
    //   - reason: human-readable explanation for skipping
    //   - error: actual error that should stop execution
    PreExecute(ctx context.Context) (skip bool, reason string, err error)
}
```

### Return value behaviour

| skip | error | Result |
|------|-------|--------|
| false | nil | Continue with generation |
| true | nil | Skip gracefully, log reason in verbose mode |
| true | err | Skip and log error |
| false | err | Stop execution, display error |

### Timeout

PreExecute hooks have a **5-second timeout**. Operations should complete within this time.

### Example implementation

```go
func (p *KittyPlugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
    // Check for context cancellation
    if ctx.Err() != nil {
        return true, "cancelled", ctx.Err()
    }

    // Check if kitty executable exists on PATH
    _, err = exec.LookPath("kitty")
    if err != nil {
        // Not an error - just skip this plugin
        return true, "kitty executable not found on $PATH", nil
    }

    // Check if config directory exists
    configDir := p.DefaultOutputDir()
    if _, err := os.Stat(configDir); os.IsNotExist(err) {
        // Try to create it
        if err := os.MkdirAll(configDir, 0750); err != nil {
            return true, fmt.Sprintf("failed to create config directory: %s", configDir), nil
        }
    }

    return false, "", nil
}
```

### Best practices

1. **Use skip for expected conditions** - App not installed is expected, return `skip=true`, not an error
2. **Return errors for unexpected failures** - Permission issues, system errors should return actual errors
3. **Provide helpful reason messages** - Users should understand why a plugin was skipped
4. **Respect the context** - Check `ctx.Err()` for cancellation
5. **Keep it fast** - Pre-hooks should complete in under 5 seconds

## PostExecuteHook interface

### Purpose

The `PostExecuteHook` interface allows plugins to perform actions after successful file generation:

- Reload application configuration
- Send signals to running processes (SIGHUP, SIGUSR1)
- Restart services via systemctl
- Set wallpapers
- Notify users of changes

### Interface definition

```go
type PostExecuteHook interface {
    // PostExecute runs after successful Generate() and file writing.
    // The execCtx contains execution context including wallpaper path.
    // The writtenFiles contains the paths that were written.
    // Errors are logged but don't fail the overall operation.
    PostExecute(ctx context.Context, execCtx ExecutionContext, writtenFiles []string) error
}
```

### ExecutionContext struct

```go
type ExecutionContext struct {
    DryRun        bool   // Whether this is a dry-run (no files written)
    Verbose       bool   // Whether verbose output is enabled
    OutputDir     string // The output directory being used
    WallpaperPath string // Path to source wallpaper (from input plugin)
}
```

### Timeout

PostExecute hooks have a **10-second timeout**. Operations should complete within this time.

### Error handling

PostExecute errors are **logged but don't fail** the overall generation. This allows other plugins to continue even if one plugin's reload fails.

### Example implementation

```go
func (p *KittyPlugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, writtenFiles []string) error {
    // Don't do anything on dry-run
    if execCtx.DryRun {
        return nil
    }

    // Check for context cancellation
    if ctx.Err() != nil {
        return ctx.Err()
    }

    // Reload all kitty instances by sending SIGUSR1 signal
    cmd := exec.CommandContext(ctx, "killall", "-SIGUSR1", "kitty")
    if err := cmd.Run(); err != nil {
        // Check if it's just "no process found"
        if exitErr, ok := err.(*exec.ExitError); ok {
            if exitErr.ExitCode() == 1 {
                // No kitty instances running - not an error
                return nil
            }
        }
        return fmt.Errorf("failed to reload kitty: %w", err)
    }

    if execCtx.Verbose {
        fmt.Fprintf(os.Stderr, "   Kitty config reloaded\n")
    }

    return nil
}
```

### Using wallpaper path

For plugins that set wallpapers (like hyprpaper):

```go
func (p *HyprpaperPlugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, writtenFiles []string) error {
    if execCtx.DryRun || execCtx.WallpaperPath == "" {
        return nil
    }

    // Set wallpaper using hyprctl
    cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "wallpaper",
        fmt.Sprintf(",%s", execCtx.WallpaperPath))
    return cmd.Run()
}
```

### Best practices

1. **Check dry-run mode** - Don't perform actions when `execCtx.DryRun` is true
2. **Handle missing processes gracefully** - No running instances isn't a failure
3. **Use context-aware commands** - `exec.CommandContext(ctx, ...)` respects cancellation
4. **Return descriptive errors** - Help users understand what failed
5. **Don't panic** - Errors are logged, other plugins continue

## Other optional interfaces

Tinct also provides these optional interfaces:

### VerbosePlugin

Receive verbose logging settings:

```go
type VerbosePlugin interface {
    SetVerbose(verbose bool)
}
```

### TemplateProvider

Expose embedded templates for management commands:

```go
type TemplateProvider interface {
    GetEmbeddedFS() any
}
```

### VersionedPlugin

Support version-specific templates:

```go
type VersionedPlugin interface {
    GetTargetVersion() string
}
```

### DualThemePlugin

Generate both light and dark themes:

```go
type DualThemePlugin interface {
    GenerateDualTheme(primaryTheme, alternateTheme *colour.ThemeData) (map[string][]byte, error)
}
```

## Complete example

A plugin implementing all optional interfaces:

```go
type MyPlugin struct {
    outputDir string
    verbose   bool
    reload    bool
}

// Required Plugin interface methods...

func (p *MyPlugin) SetVerbose(verbose bool) {
    p.verbose = verbose
}

func (p *MyPlugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
    if _, err := exec.LookPath("myapp"); err != nil {
        return true, "myapp not found on $PATH", nil
    }
    return false, "", nil
}

func (p *MyPlugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, writtenFiles []string) error {
    if execCtx.DryRun || !p.reload {
        return nil
    }

    cmd := exec.CommandContext(ctx, "myapp", "--reload-config")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to reload myapp: %w", err)
    }

    if p.verbose {
        fmt.Fprintf(os.Stderr, "   myapp config reloaded\n")
    }
    return nil
}
```

## See also

- [Creating plugins](/docs/plugin-development/creating) - Plugin implementation
- [Protocols](/docs/plugin-development/protocols) - Communication details
