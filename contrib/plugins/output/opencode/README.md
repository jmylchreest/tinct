# Tinct OpenCode Theme Plugin

An external output plugin for [tinct](https://github.com/jmylchreest/tinct) that generates theme files for [OpenCode](https://opencode.ai).

## Features

- **Full OpenCode Theme Support**: Generates complete OpenCode custom theme JSON with all ~50 color tokens
- **Semantic Color Mapping**: Maps tinct's semantic roles to OpenCode's theme tokens
- **Color Definitions**: Uses OpenCode's `defs` system for reusable color references
- **Dual-Theme Support**: Can generate both dark and light theme variants
- **Template-Based**: Uses tinct's public template API for maintainable theme generation
- **Auto-Detection**: Automatically detects dark/light theme type
- **Material Design 3**: Leverages tinct's Material Design 3 surface colors

## Installation

### Build from Source

```bash
cd contrib/plugins/output/opencode
go build -o tinct-plugin-opencode
```

### Install to PATH

```bash
# Build with version information
go build -ldflags "-X main.Version=0.1.0 -X main.Commit=$(git rev-parse HEAD) -X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o tinct-plugin-opencode

# Install to ~/.local/bin (or any directory in your PATH)
mkdir -p ~/.local/bin
mv tinct-plugin-opencode ~/.local/bin/
chmod +x ~/.local/bin/tinct-plugin-opencode
```

## Usage

The plugin uses tinct's go-plugin protocol and is discovered automatically when in your PATH.

### Generate from Image

```bash
tinct generate -i file --path ~/wallpaper.png -o opencode
```

### Generate from Google Gemini

```bash
export GOOGLE_API_KEY="your-api-key"
tinct generate -i google-genai --prompt "sunset over mountains" -o opencode
```

### Specify Custom Output Directory

```bash
tinct generate -i file --path ~/wallpaper.png -o opencode --output-dir ~/.config/opencode/themes
```

## Output

The plugin generates theme files in:

```
~/.config/opencode/themes/tinct.json
```

When dual-theme mode is active, an additional file is generated:

```
~/.config/opencode/themes/tinct-light.json   (or tinct-dark.json)
```

## Applying the Theme to OpenCode

After generating the theme, activate it in OpenCode:

### Method 1: Command (Recommended)

1. Open OpenCode
2. Type `/theme` and select **"tinct"** from the list

### Method 2: Config File

Add the theme to your OpenCode config:

```bash
vim ~/.config/opencode/opencode.json
```

```json
{
  "$schema": "https://opencode.ai/config.json",
  "theme": "tinct"
}
```

### Theme Reload

OpenCode does not currently support hot-reloading themes. After regenerating the theme with tinct, you will need to restart OpenCode or use `/theme` to re-select the theme for changes to take effect.

There is an [open feature request](https://github.com/anomalyco/opencode/issues/815) for theme hot-reload support.

**Tip:** If you are also using tinct to theme your terminal emulator, you can use OpenCode's built-in `system` theme as an alternative, which adapts to your terminal's ANSI colors in real time.

## Theme Structure

The generated theme maps tinct's semantic color system to OpenCode's theme tokens:

### Color Definitions

The `defs` section defines reusable colors from tinct's palette that are referenced throughout the theme:

| Definition | Tinct Role | Description |
|------------|------------|-------------|
| background | background | Primary background |
| foreground | foreground | Primary text |
| surface | surface | Panel/sidebar surfaces |
| accent1-4 | accent1-4 | Accent colors |
| danger | danger | Error/danger color |
| warning | warning | Warning color |
| success | success | Success color |
| info | info | Informational color |

### Theme Token Mapping

| OpenCode Token | Tinct Role | Description |
|----------------|------------|-------------|
| primary | accent1 | Primary accent |
| secondary | accent2 | Secondary accent |
| accent | accent3 | Tertiary accent |
| error | danger | Error indicators |
| warning | warning | Warning indicators |
| success | success | Success indicators |
| info | info | Info indicators |
| text | foreground | Primary text |
| textMuted | foregroundMuted | Muted/secondary text |
| background | background | Main background |
| backgroundPanel | surfaceContainerLow | Panel background |
| backgroundElement | surfaceContainer | Element background |
| border | outline | Primary border |
| borderActive | accent1 | Active/focused border |
| borderSubtle | outlineVariant | Subtle border |

### Diff Colors

| OpenCode Token | Tinct Role |
|----------------|------------|
| diffAdded | success |
| diffRemoved | danger |
| diffContext | onSurfaceVariant |
| diffHunkHeader | accent2 |
| diffAddedBg | success @ 15% alpha |
| diffRemovedBg | danger @ 15% alpha |

### Markdown Rendering

| OpenCode Token | Tinct Role |
|----------------|------------|
| markdownHeading | accent1 |
| markdownLink | accent2 |
| markdownCode | success |
| markdownEmph | accent4 |
| markdownStrong | accent1 |

### Syntax Highlighting

| OpenCode Token | Tinct Role |
|----------------|------------|
| syntaxComment | foregroundMuted |
| syntaxKeyword | accent4 |
| syntaxFunction | accent1 |
| syntaxString | success |
| syntaxNumber | accent3 |
| syntaxType | accent2 |
| syntaxVariable | foreground |
| syntaxOperator | onSurface |
| syntaxPunctuation | onSurfaceVariant |

## Template Customization

The theme template uses tinct's public template API. To customize:

```bash
# Extract the template
mkdir -p ~/.config/tinct/templates/opencode
cp templates/theme.json.tmpl ~/.config/tinct/templates/opencode/

# Edit the template
$EDITOR ~/.config/tinct/templates/opencode/theme.json.tmpl
```

### Available Template Functions

- `get . "roleName"` - Get color by semantic role
- `hex` - Format as #RRGGBB
- `hexAlpha` - Format as #RRGGBBAA
- `withAlpha value alpha` - Set alpha (0.0-1.0)
- `themeType .` - Get "dark" or "light"

Example:
```json
"background": "{{ get . \"background\" | hex }}",
"accent": "{{ get . \"accent1\" | hex }}",
"transparent": "{{ withAlpha (get . \"surface\") 0.5 | hexAlpha }}"
```

## Technical Details

### Plugin Protocol
- **Type**: go-plugin (HashiCorp)
- **Protocol Version**: 0.1.0
- **Plugin Type**: output

### Dependencies
- `github.com/jmylchreest/tinct/pkg/colour` - Color utilities
- `github.com/jmylchreest/tinct/pkg/template` - Template helpers
- `github.com/jmylchreest/tinct/pkg/plugin` - Plugin protocol
- `github.com/hashicorp/go-plugin` - Plugin framework

### Architecture

1. **Receives** PaletteData via go-plugin RPC
2. **Converts** to colour.ThemeData with PaletteHelper
3. **Renders** template using template.TemplateFuncs()
4. **Outputs** JSON to ~/.config/opencode/themes/

## Troubleshooting

### Plugin Not Found

Ensure the binary is in your PATH:
```bash
which tinct-plugin-opencode
```

### Theme Not Loading in OpenCode

1. Check the output file exists:
   ```bash
   ls -la ~/.config/opencode/themes/tinct.json
   ```

2. Validate JSON syntax:
   ```bash
   jq . ~/.config/opencode/themes/tinct.json
   ```

3. Restart OpenCode to reload themes

### Template Errors

Check tinct output for template rendering errors:
```bash
tinct generate -i file --path ~/wallpaper.png -o opencode --verbose
```

## Examples

### Dark Theme from Image
```bash
tinct generate -i file --path ~/dark-wallpaper.png -o opencode
```

### Light Theme with Prompt
```bash
export GOOGLE_API_KEY="your-key"
tinct generate -i google-genai --prompt "bright sunny beach" -o opencode
```

### Generate for Multiple Apps
```bash
# Generate themes for terminal + OpenCode simultaneously
tinct generate -i file --path ~/wallpaper.png -o ghostty,opencode,neovim
```

### Preview Generated Theme
```bash
cat ~/.config/opencode/themes/tinct.json | jq '.theme'
```

## Contributing

This plugin is part of the tinct project. Contributions welcome!

## License

Same as tinct - check the main repository for license information.

## See Also

- [Tinct Documentation](https://github.com/jmylchreest/tinct)
- [OpenCode Theme Documentation](https://opencode.ai/docs/themes/)
- [OpenCode Theme Schema](https://opencode.ai/theme.json)
