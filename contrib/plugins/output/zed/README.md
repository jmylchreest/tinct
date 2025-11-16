# Tinct Zed Theme Plugin

An external output plugin for [tinct](https://github.com/jmylchreest/tinct) that generates theme files for the [Zed editor](https://zed.dev).

## Features

- **Full Zed Theme Support**: Generates complete Zed theme JSON with 100+ color mappings
- **Semantic Color Mapping**: Maps tinct's semantic roles to Zed's extensive UI elements
- **ANSI Terminal Colors**: Automatic ANSI color matching for terminal themes
- **Template-Based**: Uses tinct's public template API for maintainable theme generation
- **Auto-Detection**: Automatically detects dark/light theme type
- **Material Design 3**: Leverages tinct's Material Design 3 surface colors

## Installation

### Build from Source

```bash
cd contrib/plugins/output/zed
go build -o tinct-plugin-zed
```

### Install to PATH

```bash
# Build with version information
go build -ldflags "-X main.Version=0.1.0 -X main.Commit=$(git rev-parse HEAD) -X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o tinct-plugin-zed

# Install to ~/.local/bin (or any directory in your PATH)
mkdir -p ~/.local/bin
mv tinct-plugin-zed ~/.local/bin/
chmod +x ~/.local/bin/tinct-plugin-zed
```

## Usage

The plugin uses tinct's go-plugin protocol and is discovered automatically when in your PATH.

### Generate from Image

```bash
tinct generate -i file --path ~/wallpaper.png -o zed
```

### Generate from Google Gemini

```bash
export GOOGLE_API_KEY="your-api-key"
tinct generate -i google-genai --prompt "sunset over mountains" -o zed
```

### Specify Custom Output Directory

```bash
tinct generate -i file --path ~/wallpaper.png -o zed --output-dir ~/.config/zed/themes
```

## Output

The plugin generates a single file:

```
~/.config/zed/themes/tinct.json
```

## Applying the Theme to Zed

After generating the theme, you need to activate it in Zed:

### Method 1: Command Palette (Recommended)

1. Open Zed
2. Press `Cmd+Shift+P` (macOS) or `Ctrl+Shift+P` (Linux/Windows) to open the command palette
3. Type "theme" and select **"theme selector: toggle"**
4. Search for "Tinct" in the theme list
5. Select the theme to apply it

### Method 2: Settings File

Add the theme to your Zed settings file:

```bash
# Open Zed settings
zed ~/.config/zed/settings.json
```

Add or update the `theme` setting:

```json
{
  "theme": "Tinct"
}
```

### Method 3: Settings UI

1. Open Zed
2. Go to **Settings** (Cmd+, on macOS, Ctrl+, on Linux)
3. Search for "theme"
4. Select "Tinct" from the theme dropdown

### Automatic Theme Reload

Zed automatically watches the themes directory for changes. When you regenerate the theme with tinct, Zed will reload it automatically if it's already selected.

```bash
# Generate new theme from a different wallpaper
tinct generate -i file --path ~/new-wallpaper.png -o zed

# Zed will automatically reload the theme if "Tinct" is selected
```

### Verifying Theme Installation

Check that Zed recognizes your theme:

```bash
# List all available Zed themes (including custom ones)
ls -la ~/.config/zed/themes/

# View the generated theme
cat ~/.config/zed/themes/tinct.json | jq '.name'
```

The output should show `"Tinct"`.

## Theme Structure

The generated theme includes:

### UI Elements
- **Borders**: outline, accent1, borderMuted
- **Surfaces**: background, surface, surfaceContainer variants
- **Editor**: background, foreground, line numbers, gutters
- **Panels**: status bar, title bar, toolbar, tabs
- **Scrollbars**: with semi-transparent styling

### Status Indicators
- **Error**: danger role (red)
- **Warning**: warning role (yellow/orange)
- **Success**: success role (green)
- **Info**: info role (blue)
- **Hint**: info role

### Syntax Highlighting
- **Comments**: foregroundMuted (italic)
- **Keywords**: accent4 (purple/pink)
- **Functions**: accent1 (primary accent)
- **Strings**: success (green)
- **Types**: accent2 (secondary accent)
- **Constants**: accent3
- **Operators**: onSurface

### Terminal Colors
All 16 ANSI colors using tinct's perceptual color matching:
- Standard: black, red, green, yellow, blue, magenta, cyan, white
- Bright variants: bright_black through bright_white

## Color Mapping

The plugin leverages tinct's semantic color roles:

| Zed Element | Tinct Role | Description |
|-------------|------------|-------------|
| background | background | Editor background |
| surface | surface | Panels, sidebars |
| border | outline | Primary borders |
| text | foreground | Primary text |
| text.accent | accent1 | Highlighted text |
| error | danger | Error indicators |
| warning | warning | Warning indicators |
| success | success | Success indicators |
| terminal.ansi.* | ANSI mapping | Automatic color matching |

## Template Customization

The theme template uses tinct's public template API. To customize:

```bash
# Extract the template
mkdir -p ~/.config/tinct/templates/zed
cp templates/theme.json.tmpl ~/.config/tinct/templates/zed/

# Edit the template
$EDITOR ~/.config/tinct/templates/zed/theme.json.tmpl
```

### Available Template Functions

- `get . "roleName"` - Get color by semantic role
- `ansi . "colorName"` - Get ANSI color by name
- `hex` - Format as #RRGGBB
- `hexAlpha` - Format as #RRGGBBAA
- `withAlpha value alpha` - Set alpha (0.0-1.0)
- `themeType .` - Get "dark" or "light"

Example:
```json
"background": "{{ get . \"background\" | hex }}",
"accent": "{{ get . \"accent1\" | hex }}",
"transparent": "{{ withAlpha (get . \"surface\") 0.5 | hexAlpha }}",
"terminal.red": "{{ ansi . \"red\" | hex }}"
```

## Technical Details

### Plugin Protocol
- **Type**: go-plugin (HashiCorp)
- **Protocol Version**: 0.0.1
- **Plugin Type**: output

### Dependencies
- `github.com/jmylchreest/tinct/pkg/colour` - Color utilities
- `github.com/jmylchreest/tinct/pkg/template` - Template helpers
- `github.com/jmylchreest/tinct/pkg/plugin` - Plugin protocol
- `github.com/hashicorp/go-plugin` - Plugin framework

### Architecture

This plugin demonstrates tinct's public API:

1. **Receives** PaletteData via go-plugin RPC
2. **Converts** to colour.ThemeData with PaletteHelper
3. **Renders** template using template.TemplateFuncs()
4. **Outputs** JSON to ~/.config/zed/themes/

## Troubleshooting

### Plugin Not Found

Ensure the binary is in your PATH:
```bash
which tinct-plugin-zed
```

### Theme Not Loading in Zed

1. Check the output file exists:
   ```bash
   ls -la ~/.config/zed/themes/tinct.json
   ```

2. Validate JSON syntax:
   ```bash
   jq . ~/.config/zed/themes/tinct.json
   ```

3. Restart Zed to reload themes

### Template Errors

Check tinct output for template rendering errors:
```bash
tinct generate -i file --path ~/wallpaper.png -o zed --verbose
```

## Examples

### Dark Theme from Image
```bash
tinct generate -i file --path ~/dark-wallpaper.png -o zed
```

### Light Theme with Prompt
```bash
export GOOGLE_API_KEY="your-key"
tinct generate -i google-genai --prompt "bright sunny beach" -o zed
```

### Preview Generated Theme
```bash
cat ~/.config/zed/themes/tinct.json | jq '.themes[0].style | keys'
```

## Contributing

This plugin is part of the tinct project. Contributions welcome!

## License

Same as tinct - check the main repository for license information.

## See Also

- [Tinct Documentation](https://github.com/jmylchreest/tinct)
- [Zed Theme Documentation](https://zed.dev/docs/themes)
- [Zed Theme Schema](https://zed.dev/schema/themes/v0.2.0.json)
