---
sidebar_position: 1
---

# Templating

Customise how tinct generates configuration files.

## Overview

Tinct uses Go's `text/template` syntax to generate configuration files for each output plugin. You can customise these templates to adjust output format, add custom settings, or support additional configuration options.

## Quick start

### View available templates

```bash
# List templates for a plugin
tinct templates list --plugin hyprland
```

### Dump templates for editing

```bash
# Export all templates for a plugin
tinct templates dump --plugin hyprland

# Export a specific template
tinct templates dump --plugin hyprland --file tinct-colours.conf.tmpl
```

Templates are copied to `~/.config/tinct/templates/<plugin>/`.

### Edit and use

After editing, tinct automatically uses your custom templates:

```bash
tinct generate -i image -p ~/wallpaper.jpg -o hyprland
```

## Template locations

Templates are loaded in this order (first found wins):

1. **Custom**: `~/.config/tinct/templates/<plugin>/<filename>`
2. **Versioned**: Embedded templates for specific target versions
3. **Default**: Embedded default templates

## Basic syntax

Templates use Go template syntax with tinct-specific functions:

```go
# Theme: {{ themeType . }}

background {{ get . "background" | hex }}
foreground {{ get . "foreground" | hex }}

# With alpha
overlay {{ get . "scrim" | rgba }}
```

### Accessing colours

```go
{{ get . "background" | hex }}      # #1e1e2e
{{ get . "foreground" | rgb }}      # rgb(205, 214, 244)
{{ get . "accent1" | withAlpha 0.5 | rgba }}  # rgba(137, 180, 250, 0.50)
```

### Conditionals

```go
{{ if has . "border" }}
border {{ get . "border" | hex }}
{{ else }}
border {{ get . "outline" | hex }}
{{ end }}
```

### Variables

```go
{{- $bg := get . "background" }}
{{- $fg := get . "foreground" }}

background {{ $bg | hex }}
foreground {{ $fg | hex }}
```

### Loops

```go
{{ range allRoles . }}
# {{ . }}: {{ get $ . | hex }}
{{ end }}
```

## Reset to defaults

Remove custom templates to restore default behaviour:

```bash
rm -rf ~/.config/tinct/templates/hyprland
```

## Next steps

- [Functions reference](./functions.md) - All available template functions
- [Colour access](./color-access.md) - Patterns for accessing colours
- [Format conversion](./format-conversion.md) - Output format options
- [Versioned templates](./versioned.md) - Version-specific templates
