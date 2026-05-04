---
sidebar_position: 5
---

# Versioned templates

Handle breaking changes in target application configuration.

## Overview

Some applications change their configuration syntax between versions. Versioned templates allow plugins to support multiple versions automatically.

## How it works

1. **Version detection**: Plugin detects installed application version
2. **Template selection**: Loader checks for version-specific templates
3. **Best match**: Selects highest version not exceeding target
4. **Fallback**: Uses default template if no match found

## Directory structure

Versioned templates are stored in subdirectories:

```
internal/plugin/output/hyprland/
├── tinct.conf.tmpl              # Default (older versions)
├── tinct-colours.conf.tmpl
└── templates/
    └── 0.53/
        └── tinct.conf.tmpl      # For Hyprland v0.53+
```

## Template resolution order

1. **Custom**: `~/.config/tinct/templates/<plugin>/` (highest priority)
2. **Versioned**: `templates/<version>/` (if target version set)
3. **Default**: Root-level templates (fallback)

## Example: Hyprland windowrule syntax

Hyprland v0.53 changed windowrule syntax:

**Default template (pre-0.53):**

```go
# Old syntax
windowrule = bordercolor $dangerRGB, title:^(.*[Ee]rror.*)$
windowrule = bordercolor $warningRGB, title:^(.*[Ww]arning.*)$
```

**templates/0.53/tinct.conf.tmpl:**

```go
# New syntax
windowrule {
    name = tinct-error-border
    border_color = $dangerRGB
    match:title = ^(.*[Ee]rror.*)$
}
```

## Version matching rules

- Uses semantic versioning (e.g., "0.53.0", "1.2.3")
- Finds highest version directory that is less than or equal to target version

Examples:

| Target | Available | Selected |
|--------|-----------|----------|
| 0.53.0 | 0.52, 0.53 | 0.53 |
| 0.52.5 | 0.52, 0.53 | 0.52 |
| 0.55.0 | 0.52, 0.53 | 0.53 (highest) |
| 0.50.0 | 0.52, 0.53 | default |

## Verbose logging

When verbose mode is enabled:

```
   Detected Hyprland version: 0.53.0
   Using versioned template: templates/0.53/tinct.conf.tmpl (target: 0.53.0)
```

Or when falling back:

```
   Could not detect Hyprland version, using default templates
   Using default embedded template: tinct.conf.tmpl
```

## Custom templates override versioning

Your custom templates always take priority:

```bash
# Your custom template at:
~/.config/tinct/templates/hyprland/tinct.conf.tmpl

# Will be used regardless of detected version
```

## Creating versioned templates

For plugin developers implementing version support:

### 1. Implement VersionedPlugin interface

```go
func (p *Plugin) GetTargetVersion() string {
    output, err := exec.Command("myapp", "--version").Output()
    if err != nil {
        return ""
    }
    return parseVersion(string(output))
}
```

### 2. Use versioned loader in Generate

```go
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
    loader := tmplloader.New("myplugin", templates)

    if version := p.GetTargetVersion(); version != "" {
        loader.WithTargetVersion(version)
    }

    content, _, err := loader.Load("theme.conf.tmpl")
    // ...
}
```

### 3. Embed versioned templates

```go
//go:embed *.tmpl templates
var templates embed.FS
```

## Best practices

1. **Graceful fallback**: Always provide a default template
2. **Document changes**: Comment what changed between versions
3. **Test both paths**: Verify old and new templates work
4. **Descriptive names**: Use meaningful names in new syntax

## See also

- [Functions reference](/docs/templating/functions)
- [template output plugin](/docs/plugins/output/special/template)
