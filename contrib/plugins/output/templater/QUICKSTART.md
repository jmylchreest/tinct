# Templater - Quick Start Guide

## Installation

### 1. Build the Plugin

```bash
cd contrib/plugins/output/templater
go build -o /.local/share/tinct/plugins/tinct-plugin-templater
```

### 2. Register with Tinct

Create or edit `~/.config/tinct/plugins.json`:

```json
{
  "external_plugins": {
    "templater": {
      "path": "/home/YOUR_USERNAME/.local/share/tinct/plugins/templater",
      "type": "output",
      "version": "1.0.0"
    }
  },
  "enabled_plugins": [
    "output:templater"
  ]
}
```

### 3. Create Configuration Directory

```bash
mkdir -p ~/.config/tinct/templates
```

### 4. Create Configuration File

Copy the example configuration:

```bash
cp contrib/plugins/output/templater/example-config.yaml \
   ~/.config/tinct/templater.yaml
```

Or create a minimal one:

```yaml
templates:
  - name: "my-app-colours"
    description: "My app colour configuration"
    template_path: "~/.config/tinct/templates/myapp.tmpl"
    output_path: "~/.config/myapp/colours.conf"
    enabled: true

settings:
  create_dirs: true
  backup: true
  verbose: false
```

## Creating Your First Template

### Example 1: Simple CSS Variables

Create `~/.config/tinct/templates/css-vars.tmpl`:

```css
:root {
  --bg: {{ (get . "background").Hex }};
  --fg: {{ (get . "foreground").Hex }};
  --accent: {{ (get . "accent1").Hex }};
}
```

Add to config:

```yaml
templates:
  - name: "css-vars"
    template_path: "~/.config/tinct/templates/css-vars.tmpl"
    output_path: "~/myproject/colours.css"
    enabled: true
```

### Example 2: i3 Window Manager

Create `~/.config/tinct/templates/i3.tmpl`:

```
client.focused {{ (get . "accent1").Hex }} {{ (get . "background").Hex }}
```

## Usage

### Basic Generation

```bash
# Generate from image
tinct generate --input image -p wallpaper.jpg --outputs templater

# Generate from palette file
tinct generate --input file -p theme.json --outputs templater
```

### With Custom Config

```bash
tinct generate --input image -p wallpaper.jpg --outputs templater \
  --plugin-args 'templater={"config":"~/my-config.yaml"}'
```

### Process Specific Templates Only

```bash
tinct generate --input image -p wallpaper.jpg --outputs templater \
  --plugin-args 'templater={"templates":["css-vars","i3"]}'
```

### Verbose Mode

```bash
tinct generate --input image -p wallpaper.jpg --outputs templater \
  --plugin-args 'templater={"verbose":true}'
```

### Dry Run

```bash
tinct generate --input image -p wallpaper.jpg --outputs templater --dry-run
```

## Template Syntax

### Available Functions

- `get` - Get colour: `{{ (get . "background").Hex }}`
- `has` - Check if exists: `{{ if has . "accent4" }}...{{ end }}`
- `themeType` - Get theme type: `{{ themeType . }}` (returns "dark" or "light")
- `seq` - Generate sequence: `{{ range seq 1 4 }}{{ . }}{{ end }}`
- `ansi` - Get ANSI colour: `{{ (ansi . 0).Hex }}` (for terminal themes)

### Colour Methods

```
{{ $colour := get . "accent1" }}
{{ $colour.Hex }}        → #89b4fa
{{ $colour.HexNoHash }}  → 89b4fa
{{ $colour.RGB }}        → rgb(137,180,250)
{{ $colour.RGBA }}       → rgba(137,180,250,1)
{{ $colour.R }}          → 137
{{ $colour.G }}          → 180
{{ $colour.B }}          → 250
{{ $colour.RGBDecimal }} → 137,180,250
```

### Available Colour Roles

Core colours:
- `background`, `backgroundMuted`
- `foreground`, `foregroundMuted`

Accents:
- `accent1`, `accent2`, `accent3`, `accent4`
- `accent1Muted`, `accent2Muted`, etc.

Semantic:
- `danger`, `warning`, `success`, `info`, `notification`

Surface:
- `surface`, `onSurface`, `surfaceVariant`, `onSurfaceVariant`
- `outline`, `border`, `borderMuted`, `outlineVariant`

And many more! Use `has` to check availability.

## Common Patterns

### Conditional Colours

```css
/* Only include accent4 if it exists */
{{- if has . "accent4" }}
--accent-4: {{ (get . "accent4").Hex }};
{{- end }}
```

### Loop Through Accents

```css
{{- range seq 1 4 }}
  {{- $role := printf "accent%d" . }}
  {{- if has $ $role }}
--accent-{{ . }}: {{ (get $ $role).Hex }};
  {{- end }}
{{- end }}
```

### RGB Components

```css
--colour-rgb: {{ (get . "accent1").R }}, {{ (get . "accent1").G }}, {{ (get . "accent1").B }};
```

### Theme-Specific Values

```css
{{ if eq (themeType .) "dark" }}
  --shadow: rgba(0, 0, 0, 0.5);
{{ else }}
  --shadow: rgba(0, 0, 0, 0.2);
{{ end }}
```

## Troubleshooting

### Plugin Not Found

```bash
# Verify plugin is executable
chmod +x ~/.local/share/tinct/plugins/templater

# Test plugin info
~/.local/share/tinct/plugins/templater --plugin-info
```

### Config File Not Found

```bash
# Check config exists
ls -la ~/.config/tinct/templater.yaml

# Use custom config path
tinct generate ... --plugin-args 'templater={"config":"~/my-config.yaml"}'
```

### Template Errors

Enable verbose mode to see detailed error messages:

```bash
tinct generate ... --plugin-args 'templater={"verbose":true}'
```

### Colour Role Not Found

Always use `has` before `get` for optional colours:

```
{{- if has . "accent4" }}
{{ (get . "accent4").Hex }}
{{- else }}
#default-colour
{{- end }}
```

## Next Steps

- Check `README.md` for complete documentation
- See `templates/` directory for more examples
- Explore the example config: `example-config.yaml`
- Create your own templates for your favorite applications!
