# Output Plugin Standard

This document defines the standard rules and conventions that all output plugins **MUST** follow to ensure consistency and predictability across the Tinct ecosystem.

## Table of Contents

1. [File Naming Convention](#file-naming-convention)
2. [Plugin Structure](#plugin-structure)
3. [Interface Requirements](#interface-requirements)
4. [Template Requirements](#template-requirements)
5. [Testing Requirements](#testing-requirements)
6. [Color Variable Naming](#color-variable-naming)
7. [Documentation Requirements](#documentation-requirements)
8. [Examples](#examples)

---

## File Naming Convention

All output plugins **MUST** follow this naming standard for generated files:

### Two-File Pattern (Preferred)

**Use when the application supports:**
- Color variables/constants (CSS variables, shell variables, etc.)
- Include/source/import directives

**Files generated:**
1. `tinct-colours.{ext}` - Color variable definitions only
2. `tinct.{ext}` - Example configuration that uses the color variables

**Benefits:**
- Users can modify example config without touching color definitions
- Color file can be imported into multiple configuration files
- Clear separation of concerns
- Easy to update colors without breaking configuration

### Single-File Pattern

**Use when the application:**
- Does NOT support variables, OR
- Does NOT support includes/imports

**File generated:**
- `tinct.{ext}` - Complete configuration with colors embedded directly

### Naming Examples

| Application | Pattern | Files Generated |
|-------------|---------|----------------|
| Hyprland | Two-file | `tinct-colours.conf`, `tinct.conf` |
| Waybar | Two-file | `tinct-colours.css`, `tinct.css` |
| Kitty | Single-file | `tinct.conf` |
| Alacritty | Two-file | `tinct-colours.toml`, `tinct.toml` |

### File Extension Rules

- **MUST** use the application's native configuration file extension
- **MUST** be lowercase (e.g., `.conf`, `.css`, `.toml`, not `.CONF`)
- **MUST** include the dot (e.g., `.conf` not `conf`)

---

## Plugin Structure

Each plugin **MUST** be in its own directory with the following structure:

```
internal/plugin/output/
 pluginname/
     pluginname.go              # Plugin implementation (REQUIRED)
     pluginname_test.go         # Unit tests (REQUIRED)
     tinct-colours.ext.tmpl     # Color definitions template (if two-file)
     tinct.ext.tmpl             # Config/example template (REQUIRED)
```

### Directory Naming

- **MUST** be lowercase
- **MUST** match the application name
- **MUST** be a single word (use hyphens if needed, e.g., `my-app`)

### Go File Naming

- **MUST** be named `{pluginname}.go`
- **MUST** be in package `package {pluginname}`
- Test file **MUST** be named `{pluginname}_test.go`

### Template File Naming

- **MUST** use the exact names: `tinct-colours.{ext}.tmpl` and/or `tinct.{ext}.tmpl`
- **MUST** have `.tmpl` extension for template files
- Template files **MUST** be embedded using `go:embed *.tmpl`

---

## Interface Requirements

All plugins **MUST** implement the `output.Plugin` interface:

```go
type Plugin interface {
    Name() string
    Description() string
    DefaultOutputDir() string
    Generate(*colour.CategorisedPalette) (map[string][]byte, error)
    RegisterFlags(*cobra.Command)
    Validate() error
}
```

### Method Requirements

#### `Name() string`

- **MUST** return lowercase plugin name
- **MUST** match the directory name
- **MUST** match the application name
- Example: `"waybar"`, `"hyprland"`, `"kitty"`

#### `Description() string`

- **MUST** be a short, descriptive sentence
- **SHOULD** start with "Generate" or "Create"
- Example: `"Generate Waybar status bar color theme"`

#### `DefaultOutputDir() string`

- **MUST** return the full default path for configuration files
- **MUST** expand `~` to user home directory
- **SHOULD** use `filepath.Join()` for cross-platform compatibility
- Example: `~/.config/waybar`, `~/.config/hypr`

#### `Generate(*colour.CategorisedPalette) (map[string][]byte, error)`

- **MUST** accept a non-nil palette (return error if nil)
- **MUST** return a map of filename -> file content
- **MUST** use exact filenames: `tinct-colours.{ext}` and/or `tinct.{ext}`
- **MUST NOT** include directory paths in the map keys (just filenames)
- Example return: `{"tinct-colours.css": []byte(...), "tinct.css": []byte(...)}`

#### `RegisterFlags(*cobra.Command)`

- **MUST** register plugin-specific flags
- **MUST** use prefix `{pluginname}.` for all flags
- **MUST** register at minimum: `{pluginname}.output-dir` flag
- Example: `--waybar.output-dir`, `--hyprland.generate-stub`

#### `Validate() error`

- **MUST** check if configuration is valid
- **SHOULD** return `nil` if plugin has no validation requirements
- **MAY** validate paths, check dependencies, etc.

### Optional Interfaces

Plugins **MAY** implement additional interfaces:

#### `VerbosePlugin`

```go
type VerbosePlugin interface {
    SetVerbose(bool)
}
```

- **SHOULD** implement if plugin has verbose logging
- **MUST** store verbose flag in plugin struct
- **MUST** respect verbose flag in `Generate()` method

#### `PreExecuteHook`

```go
type PreExecuteHook interface {
    PreExecute(context.Context) (skip bool, reason string, err error)
}
```

- **MAY** implement to check prerequisites
- **SHOULD** check if application is installed
- **SHOULD** check if config directory exists
- Return `skip=true` to skip plugin without error

#### `PostExecuteHook`

```go
type PostExecuteHook interface {
    PostExecute(context.Context, []string) error
}
```

- **MAY** implement to perform post-generation tasks
- **SHOULD** reload application configuration if safe
- Examples: send SIGUSR2 to reload, restart service

---

## Template Requirements

### Embedding

Templates **MUST** be embedded using:

```go
//go:embed *.tmpl
var templates embed.FS
```

### Template Loading

Plugins **MUST** use the template loader for user customization support:

```go
import tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"

loader := tmplloader.New("pluginname", templates)
if p.verbose {
    loader.WithVerbose(true, &verboseLogger{out: os.Stderr})
}
tmplContent, fromCustom, err := loader.Load("tinct-colours.ext.tmpl")
```

### Template Structure

#### Color Definitions Template (`tinct-colours.{ext}.tmpl`)

**MUST** include:
1. Header comment explaining the file purpose
2. Import/source instructions
3. Detected theme type: `{{ .SourceTheme }}`
4. Color definitions using semantic names
5. Usage documentation in comments

**Example structure:**

```
# File header with description
# Import instructions
# Detected theme: {{ .SourceTheme }}

# Color definitions
{{- range .Colors }}
variable_name = "{{ .Hex }}"
{{- end }}

# Usage instructions in comments
```

#### Configuration Template (`tinct.{ext}.tmpl`)

**MUST** include:
1. Header comment explaining the file purpose
2. Import/source statement for color definitions file (if two-file pattern)
3. Example configuration showing color usage
4. Comments explaining available variables
5. Examples for common use cases

### Template Data Structure

Plugins **MUST** define clear data structures for templates:

```go
type ColorsData struct {
    SourceTheme string            // "dark" or "light"
    Colors      []ColorDefinition // Semantic colors
}

type ColorDefinition struct {
    Name string // Semantic name (e.g., "background")
    Hex  string // Hex color (e.g., "#1a1b26")
}
```

### Color Mapping

Plugins **MUST** map semantic roles consistently:

```go
roleMapping := map[colour.ColourRole]string{
    colour.RoleBackground:      "background",
    colour.RoleBackgroundMuted: "background-muted",
    colour.RoleForeground:      "foreground",
    colour.RoleForegroundMuted: "foreground-muted",
    colour.RoleAccent1:         "accent1",
    colour.RoleAccent1Muted:    "accent1-muted",
    colour.RoleAccent2:         "accent2",
    colour.RoleAccent2Muted:    "accent2-muted",
    colour.RoleAccent3:         "accent3",
    colour.RoleAccent3Muted:    "accent3-muted",
    colour.RoleAccent4:         "accent4",
    colour.RoleAccent4Muted:    "accent4-muted",
    colour.RoleDanger:          "danger",
    colour.RoleWarning:         "warning",
    colour.RoleSuccess:         "success",
    colour.RoleInfo:            "info",
    colour.RoleNotification:    "notification",
}
```

---

## Testing Requirements

All plugins **MUST** have comprehensive tests:

### Required Tests

1. **`TestPlugin_Name`** - Verify correct plugin name
2. **`TestPlugin_Description`** - Verify non-empty description
3. **`TestPlugin_DefaultOutputDir`** - Verify valid default path
4. **`TestPlugin_Generate`** - Verify successful generation with valid palette
5. **`TestPlugin_GenerateNilPalette`** - Verify error on nil palette
6. **`TestPlugin_Validate`** - Verify validation logic
7. **`TestPrepareData`** - Verify template data preparation

### Optional Tests (if implemented)

- **`TestPlugin_SetVerbose`** - If implements VerbosePlugin
- **`TestPlugin_PreExecute`** - If implements PreExecuteHook
- **`TestPlugin_PostExecute`** - If implements PostExecuteHook
- **`TestPlugin_GenerateWithoutStub`** - If has stub generation flag

### Test Coverage

- **MUST** achieve >80% code coverage
- **MUST** test error cases
- **SHOULD** test with various palette configurations
- **SHOULD** verify generated content format

---

## Color Variable Naming

### Standard Semantic Names

All plugins **MUST** use these semantic color names:

| Category | Variable Names |
|----------|----------------|
| **Backgrounds** | `background`, `background-muted` |
| **Foregrounds** | `foreground`, `foreground-muted` |
| **Accents** | `accent1`, `accent2`, `accent3`, `accent4` |
| **Accent Muted** | `accent1-muted`, `accent2-muted`, `accent3-muted`, `accent4-muted` |
| **Status** | `danger`, `warning`, `success`, `info`, `notification` |

### Naming Convention

- **MUST** use lowercase with hyphens (kebab-case)
- **MUST NOT** use underscores or camelCase
- **MUST** use semantic names, not color names (e.g., `danger` not `red`)
- **MUST** use consistent naming across all plugins

### Application-Specific Syntax

While the semantic names are standard, the **syntax** adapts to the application:

- **Hyprland:** `$background`, `$foreground`
- **Waybar:** `@background`, `@foreground`
- **CSS:** `--background`, `--foreground`
- **Shell:** `$BACKGROUND`, `$FOREGROUND`
- **TOML/TOML:** `background`, `foreground`

---

## Documentation Requirements

### Code Documentation

Plugin files **MUST** include:

1. Package-level documentation
2. Struct documentation
3. Method documentation for public methods
4. Comments explaining non-obvious logic

**Example:**

```go
// Package waybar provides an output plugin for Waybar status bar color themes.
package waybar

// Plugin implements the output.Plugin interface for Waybar.
type Plugin struct {
    outputDir    string
    generateStub bool
    verbose      bool
}

// New creates a new Waybar output plugin with default settings.
func New() *Plugin {
    return &Plugin{
        outputDir:    "",
        generateStub: true,
        verbose:      false,
    }
}
```

### Template Documentation

Templates **MUST** include:

1. File purpose in header comment
2. How to use/import the file
3. Theme detection info
4. Available variables list
5. Usage examples

### README Updates

When adding a new plugin, **MUST** update:

1. Main output plugin README (`internal/plugin/output/README.md`)
2. Add plugin to examples list
3. Update any architecture diagrams
4. Include sample output

---

## Examples

### Complete Two-File Plugin Example

See `internal/plugin/output/waybar/` for a complete reference implementation:

- `waybar.go` - Full plugin implementation
- `waybar_test.go` - Comprehensive tests
- `tinct-colours.css.tmpl` - Color definitions
- `tinct.css.tmpl` - Example configuration

### Complete Single-File Plugin Example

See `internal/plugin/output/kitty/` for a single-file reference:

- `kitty.go` - Plugin implementation
- `kitty_test.go` - Tests
- `tinct.conf.tmpl` - Complete configuration

---

## Plugin Checklist

Use this checklist when creating a new plugin:

- [ ] Created directory: `internal/plugin/output/{pluginname}/`
- [ ] Created Go file: `{pluginname}.go`
- [ ] Implemented all required interface methods
- [ ] Created template file(s): `tinct-colours.{ext}.tmpl` and/or `tinct.{ext}.tmpl`
- [ ] Embedded templates with `go:embed *.tmpl`
- [ ] Used template loader for user customization support
- [ ] Created test file: `{pluginname}_test.go`
- [ ] All required tests passing
- [ ] Semantic color names follow standard
- [ ] File naming follows convention (two-file or single-file)
- [ ] Added plugin documentation comments
- [ ] Added template documentation comments
- [ ] RegisterFlags uses `{pluginname}.` prefix
- [ ] DefaultOutputDir expands `~` to home directory
- [ ] Generate returns correct map structure
- [ ] Validate implemented (even if returns nil)
- [ ] Updated main README with plugin example
- [ ] Tested with real palette data
- [ ] Code coverage >80%

---

## Validation

Before submitting a new plugin:

1. **Run tests:** `go test ./internal/plugin/output/{pluginname}/... -v`
2. **Check coverage:** `go test -cover ./internal/plugin/output/{pluginname}/...`
3. **Run linter:** `golangci-lint run ./internal/plugin/output/{pluginname}/...`
4. **Build binary:** `go build ./cmd/tinct`
5. **Test generation:** `./tinct apply --plugin {pluginname} test-image.png`
6. **Verify output files** exist and are valid

---

## Questions?

If you're unsure about any standard requirement:

1. Check existing reference plugins (waybar, hyprland, kitty)
2. Review the main output plugin README
3. Ask in discussions or issues

---

**Made with  and consistency in mind**