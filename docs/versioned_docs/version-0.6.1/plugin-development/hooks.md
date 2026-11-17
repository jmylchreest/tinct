---
sidebar_position: 4
---

# Lifecycle hooks

Optional hooks for environment validation, reload, wallpaper application,
and instructional messages around file generation.

## Two ways to declare hooks

Tinct supports two complementary mechanisms:

1. **Declarative `hooks.Spec`** (recommended) — return a data struct from
   `Hooks() hooks.Spec`. The shared runner takes care of binary checks,
   directory auto-creation, reload commands, signal broadcasts, chmod,
   wallpaper application, and instructional messages. Both internal
   plugins and external Go plugins import the same package and share the
   exact same runner.

2. **Imperative `PreExecuteHook` / `PostExecuteHook`** — implement the
   methods yourself. Use this when the spec can't represent your
   plugin's behaviour (OR-conditions on prerequisites, content-aware
   inspection of existing config files, multi-strategy reload with
   per-strategy messaging, D-Bus calls, etc.).

The two are not exclusive — a plugin can implement both. The runner
evaluates the spec first on the pre-side (gating execution via
`RequiredBinaries` / `RequiredDirs`), then calls your imperative
`PreExecute` for any extra checks. On the post-side the imperative
`PostExecute` runs first (so it can emit custom warnings before the
routine reload), then the spec runs (chmod → reload → wallpaper →
instructions).

If a plugin returns no spec and implements no hooks, it works exactly as
today — neither mechanism is required.

## Declarative hooks: `hooks.Spec`

```go
import "github.com/jmylchreest/tinct/pkg/plugin/hooks"

func (p *Plugin) Hooks() hooks.Spec {
    return hooks.Spec{
        // ... fields ...
    }
}
```

### Spec fields

#### Pre-execute

| Field | Type | Behaviour |
|---|---|---|
| `RequiredBinaries` | `[]string` | Plugin is skipped if any of these are absent (resolved via `appdetect.IsPresentAny` — checks PATH, Flatpak, AppImage). |
| `OptionalBinaries` | `[]string` | Presence is recorded; absence emits a verbose-mode warning but does not skip. Use for "we can render but reload won't work" cases (`dunstctl`, `hyprctl`, `kitten`). |
| `RequiredDirs` | `[]string` | Plugin is skipped if any of these are absent. Use when the host app must be installed first (GTK/Qt config dirs). |
| `AutoCreateDir` | `bool` | When true, `MkdirAll(DefaultOutputDir())` on first run. |

#### Post-execute

| Field | Type | Behaviour |
|---|---|---|
| `Reload` | `*ReloadSpec` | Verb DSL: `VerbExec` (run a command), `VerbSignal` (send a signal to all matching processes). Failure is logged but non-fatal. |
| `ReloadFn` | `func(ctx) error` | In-process Go alternative for non-DSL reload paths. Mutually exclusive with `Reload`. |
| `MakeExecutable` | `[]string` | File basenames in the writeset to `chmod +x`. |
| `SupportsWallpaper` | `bool` | When true and `WallpaperPath` is set, calls the `Wallpaper` handler after reload. |
| `Wallpaper` | `func(ctx, path) error` | In-process wallpaper handler. Required when `SupportsWallpaper` is true. |
| `Instructions` | `string` | Optional message printed on success in verbose mode. Rendered as a `text/template` against `{OutputDir, WrittenFiles, WallpaperPath}`. |
| `InstructionsFn` | `func(ctx hooks.Context) string` | Function alternative for cases that need closure capture. Mutually exclusive with `Instructions`. |

### Reload verbs

```go
// Run a command — bare strings are exec shorthand.
Reload: &hooks.ReloadSpec{
    Verb: hooks.VerbExec,
    Args: []string{"dunstctl", "reload"},
}

// Send a Unix signal to all matching processes by name.
// On Windows this is a no-op with a verbose warning.
Reload: &hooks.ReloadSpec{
    Verb: hooks.VerbSignal,
    Args: []string{"kitty", "SIGUSR1"},
}
```

Recognised signal names: `SIGUSR1`, `SIGUSR2`, `SIGHUP`, `SIGTERM`,
`SIGINT`.

### Instructions templates

The `Instructions` string is rendered with these variables available:

- `{{.OutputDir}}` — the plugin's `DefaultOutputDir()` value
- `{{.WrittenFiles}}` — slice of paths written (use `{{index .WrittenFiles 0}}` for the first one)
- `{{.WallpaperPath}}` — set when a wallpaper was passed through

Example:

```go
Instructions: `   Theme written.

   Add to your config:
     import "{{index .WrittenFiles 0}}"`,
```

If your message needs more dynamic content (e.g. interpolating a struct
field that isn't a written file), use `InstructionsFn` instead:

```go
InstructionsFn: func(_ hooks.Context) string {
    return fmt.Sprintf("Set 'theme = %q' in your config.", p.themeName)
},
```

### Complete examples

**Simple cross-platform terminal** (alacritty, kitty, neovim, mc):

```go
import (
    "github.com/jmylchreest/tinct/pkg/plugin/hooks"
    "github.com/jmylchreest/tinct/pkg/plugin/paths"
)

func (p *Plugin) DefaultOutputDir() string {
    if p.outputDir != "" {
        return p.outputDir
    }
    return filepath.Join(paths.XDGConfigDir(), "kitty", "themes")
}

func (p *Plugin) Hooks() hooks.Spec {
    return hooks.Spec{
        RequiredBinaries: []string{"kitty"},
        OptionalBinaries: []string{"kitten"},
        AutoCreateDir:    true,
        Reload: &hooks.ReloadSpec{
            Verb: hooks.VerbSignal,
            Args: []string{"kitty", "SIGUSR1"},
        },
    }
}
```

**Wallpaper plugin** (awww, wbg):

```go
func (p *Plugin) Hooks() hooks.Spec {
    return hooks.Spec{
        OptionalBinaries:  []string{"awww"},
        AutoCreateDir:     true,
        MakeExecutable:    []string{"tinct-awww.sh"},
        SupportsWallpaper: true,
        Wallpaper:         p.setWallpaper,
    }
}
```

**Partial migration** — pre is declarative, post stays imperative:

```go
func (p *Plugin) Hooks() hooks.Spec {
    return hooks.Spec{
        RequiredDirs: []string{p.DefaultOutputDir()},
    }
}

// PostExecute kept for content-aware install messaging the spec
// can't represent.
func (p *Plugin) PostExecute(...) error { /* unchanged */ }
```

## Imperative hooks (escape hatches)

When the spec can't represent your behaviour, implement
`PreExecuteHook` / `PostExecuteHook` directly. Cases where this is the
right answer:

- **OR-semantics on prerequisites.** Tinct's `histui` plugin proceeds
  if EITHER the daemon is on `$PATH` OR `~/.config/histui/` exists.
  Spec fields are AND-only.
- **Content-aware messaging.** Tinct's `gtk3` and `gtk4` plugins inspect
  the existing `gtk.css` and `settings.ini` to emit different install
  messages depending on what's already configured. Templates can't read
  files.
- **Multi-strategy reload with per-strategy messaging.** Tinct's
  `konsole` plugin tries D-Bus first, falls back to `konsoleprofile`,
  and prints a different message per strategy.
- **D-Bus or other custom IPC.** Tinct's `kde-plasma` plugin uses
  KWin/Plasma D-Bus reload and isn't a single-shot exec.

### `PreExecuteHook` interface

```go
type PreExecuteHook interface {
    PreExecute(ctx context.Context) (skip bool, reason string, err error)
}
```

| `skip` | `error` | Result |
|---|---|---|
| `false` | `nil` | Continue with generation. |
| `true` | `nil` | Skip gracefully, log reason in verbose mode. |
| `true` | `err` | Skip and log error. |
| `false` | `err` | Stop execution, display error. |

The hook has a **5-second timeout**.

### `PostExecuteHook` interface

```go
type PostExecuteHook interface {
    PostExecute(ctx context.Context, execCtx ExecutionContext, writtenFiles []string) error
}
```

Errors are logged but never fail the overall operation. The hook has a
**10-second timeout**.

The execution order when a plugin implements both spec and imperative
hooks:

```
PreExecute lifecycle:
  1. spec.RequiredBinaries / OptionalBinaries / RequiredDirs / AutoCreateDir
  2. plugin.PreExecute(ctx)        if implemented

PostExecute lifecycle:
  1. plugin.PostExecute(ctx)       if implemented
  2. spec.MakeExecutable
  3. spec.Reload / spec.ReloadFn
  4. spec.Wallpaper                if SupportsWallpaper && WallpaperPath set
  5. spec.Instructions / InstructionsFn   in verbose mode
```

`DryRun` short-circuits the entire spec post-execute.

## User-overridable plugin config

Plugins should resolve user-overridable strings (like `DefaultOutputDir`)
through `internal/pluginconfig.Resolve`. It walks a fixed precedence
chain so users get a consistent experience across plugins:

```go
import "github.com/jmylchreest/tinct/internal/pluginconfig"

func (p *Plugin) DefaultOutputDir() string {
    return pluginconfig.Resolve("kitty", "output_dir", p.outputDir,
        filepath.Join(paths.XDGConfigDir(), "kitty", "themes"))
}
```

The arguments are:

1. **plugin name** — must match what `Name()` returns (used in
   `[plugin.<name>]` and `TINCT_PLUGIN_<NAME>_*`).
2. **key** — the field name in tinct.toml; underscores translate to
   uppercase letters in the env var.
3. **flag value** — the value cobra has already bound from
   `--<plugin>.<key>` (typically the plugin's struct field).
4. **fallback** — what to use when nothing else is set; usually a
   `paths.XDGConfigDir()`-prefixed path.

Resolution chain (highest priority first):

1. Cobra flag value (treated as set when the string is non-empty).
2. `TINCT_PLUGIN_<NAME>_<KEY>` environment variable.
3. `[plugin.<name>] <key>` in tinct.toml.
4. The fallback you passed.

Linux behaviour with no TOML/env is byte-identical to the fallback you
provide, so adopting Resolve in a new plugin is a no-op for users who
don't care about overrides — they keep getting the same default path
they always did.

## Cross-platform path resolution

Use the `pkg/plugin/paths` helpers instead of computing paths yourself —
they resolve correctly on Linux, macOS, and Windows:

```go
import "github.com/jmylchreest/tinct/pkg/plugin/paths"

// XDG everywhere (alacritty, kitty, neovim, ghostty on macOS)
filepath.Join(paths.XDGConfigDir(), "myapp")

// XDG_DATA_HOME (mc, etc.)
filepath.Join(paths.XDGDataDir(), "myapp")

// macOS-bundled apps (Ghostty, Zellij)
if os.Getenv("XDG_CONFIG_HOME") == "" && runtime.GOOS == "darwin" {
    return paths.MacOSAppSupport("com.example.MyApp")
}
return filepath.Join(paths.XDGConfigDir(), "myapp")
```

Resolution rules:

| Helper | Linux | macOS | Windows |
|---|---|---|---|
| `XDGConfigDir()` | `$XDG_CONFIG_HOME` or `~/.config` | `$XDG_CONFIG_HOME` or `~/.config` | `%APPDATA%` |
| `XDGDataDir()` | `$XDG_DATA_HOME` or `~/.local/share` | `$XDG_DATA_HOME` or `~/.local/share` | `%LOCALAPPDATA%` |
| `MacOSAppSupport(bundle)` | n/a (caller checks `runtime.GOOS`) | `~/Library/Application Support/<bundle>` | n/a |

## Environment variables for non-Go plugins

External plugins written in any language receive these environment
variables when tinct launches them. They let you make portable
decisions without re-implementing path resolution per-language:

| Variable | Value |
|---|---|
| `TINCT_GOOS` | `linux` / `darwin` / `windows` |
| `TINCT_GOARCH` | `amd64` / `arm64` / etc. |
| `TINCT_VERSION` | tinct's own version |
| `TINCT_HOME` | resolved `$HOME` |
| `TINCT_XDG_CONFIG_HOME` | resolved XDG config dir (per `pkg/plugin/paths`) |
| `TINCT_XDG_DATA_HOME` | resolved XDG data dir |

Python example:

```python
import os
goos = os.environ.get("TINCT_GOOS", "linux")
config_home = os.environ.get("TINCT_XDG_CONFIG_HOME",
                             os.path.expanduser("~/.config"))

if goos == "darwin":
    cmd = ["terminal-notifier", ...]
else:
    cmd = ["notify-send", ...]
```

The Go SDK (`pkg/plugin/hooks`, `pkg/plugin/paths`) is not available to
non-Go plugins, but every plugin protocol is unchanged — your imperative
`--pre-execute` / `--post-execute` JSON-stdio entry points keep working
exactly as before.
