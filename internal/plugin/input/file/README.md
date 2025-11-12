# File Input Plugin

**Type:** Input Plugin  
**Built-in:** Yes  
**Language:** Go

Load colour palettes from saved files or build palettes from manual colour specifications.

## Overview

The `file` plugin loads previously saved colour palettes or builds palettes from manual colour specifications. It supports both JSON (categorized palettes) and simple text formats (hex colours with optional role assignments). This is useful for reusing generated palettes, sharing themes, or manually defining specific colours.

## Features

- ✅ **Multiple formats** - JSON (categorized palettes) and text (hex lists)
- ✅ **Role preservation** - Maintains semantic role assignments from saved palettes
- ✅ **Manual specifications** - Build palettes from command-line colour specs
- ✅ **Flexible syntax** - Supports both `colour` and `colour` spelling
- ✅ **Override support** - Merge file colours with manual overrides
- ✅ **Fast** - No extraction needed, instant loading

## Usage

### Load from JSON File

```bash
# Load previously saved categorized palette
tinct generate -i file --file.path saved-palette.json -o hyprland,kitty
```

### Load from Text File

```bash
# Load simple hex list
tinct generate -i file --file.path colours.txt -o hyprland
```

### Manual Colour Specifications

```bash
# Build palette from manual colours (no file)
tinct generate -i file \
  --colour "background=#1e1e2e" \
  --colour "foreground=#cdd6f4" \
  --colour "accent1=#f38ba8" \
  -o hyprland

# Mix file with overrides
tinct generate -i file --file.path base-palette.json \
  --colour "accent1=#ff0000" \
  -o hyprland
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--file.path` | *(optional)* | Path to palette file (JSON or text) |
| `--colour` | *(repeatable)* | Manual colour specification (role=hex) |

**Note:** Either `--file.path` or `--colour` must be provided (or both).

## File Formats

### JSON Format (Categorized Palette)

Load a complete categorized palette with roles and theme metadata:

```json
{
  "colours": {
    "background": {
      "hex": "#1e1e2e",
      "rgb": {"r": 30, "g": 30, "b": 46},
      "hue": 240,
      "saturation": 0.21,
      "luminance": 0.15,
      "role": "background",
      "is_light": false
    },
    "foreground": {
      "hex": "#cdd6f4",
      "rgb": {"r": 205, "g": 214, "b": 244},
      "hue": 227,
      "saturation": 0.64,
      "luminance": 0.88,
      "role": "foreground",
      "is_light": true
    },
    "accent1": {
      "hex": "#f38ba8",
      "rgb": {"r": 243, "g": 139, "b": 168},
      "role": "accent1"
    }
  },
  "all_colours": [],
  "theme_type": 1
}
```

This is the format output by `tinct extract --categorise`.

### Text Format (Hex Colours)

Simple line-based format with hex colours:

```
# My custom palette
# Lines starting with # are comments

# Background and foreground
background=#1e1e2e
foreground=#cdd6f4

# Accents
accent1=#f38ba8
accent2=#89b4fa
accent3=#a6e3a1

# Or just hex colours without roles
#ff5555
#50fa7b
#8be9fd
```

**Supported formats:**
- `role=hex` - Hex with role assignment
- `hex` - Hex without role (will be auto-categorized)
- `#comments` - Comments (ignored)
- Empty lines (ignored)

### Hex Colour Formats

All standard hex formats are supported:

```
#RRGGBB   →  #1e1e2e  (most common)
RRGGBB    →  1e1e2e   (without #)
#RGB      →  #1e2     (shorthand)
RGB       →  1e2      (shorthand without #)
```

## Role Names

When specifying roles, you can use any of these semantic role names:

### Core Semantic Roles

| Role | Description |
|------|-------------|
| `background` | Primary background colour |
| `backgroundMuted` | Secondary/muted background |
| `foreground` | Primary text/foreground colour |
| `foregroundMuted` | Secondary/muted foreground |
| `accent1` | Primary accent colour |
| `accent1Muted` | Muted primary accent |
| `accent2` | Secondary accent colour |
| `accent2Muted` | Muted secondary accent |
| `accent3` | Tertiary accent colour |
| `accent3Muted` | Muted tertiary accent |
| `accent4` | Quaternary accent colour |
| `accent4Muted` | Muted quaternary accent |
| `danger` | Error/danger colour (red) |
| `warning` | Warning colour (yellow/orange) |
| `success` | Success colour (green) |
| `info` | Info colour (blue) |
| `notification` | Notification accent colour |

### Position Roles (Ambient Lighting)

For LED bias lighting setups:

**8 Regions:**
- `positionTopLeft`, `positionTop`, `positionTopRight`
- `positionRight`, `positionLeft`
- `positionBottomLeft`, `positionBottom`, `positionBottomRight`

**12 Regions:** (adds inner positions)
- `positionTopLeftInner`, `positionTopCenter`, `positionTopRightInner`
- `positionRightTop`, `positionRightBottom`
- `positionBottomRightInner`, `positionBottomCenter`, `positionBottomLeftInner`
- `positionLeftBottom`, `positionLeftTop`

**16 Regions:** (adds outer/center positions)
- Additional positions for fine-grained LED control

**Note:** Role names are case-insensitive and ignore underscores/hyphens.

## Examples

### Save and Reload Palette

```bash
# Extract and save palette
tinct extract --categorise wallpaper.jpg > saved-palette.json

# Later, reload the exact same palette
tinct generate -i file --file.path saved-palette.json -o hyprland
```

### Create Custom Palette

Create `my-theme.txt`:
```
# My custom dark theme
background=#1a1b26
foreground=#c0caf5

accent1=#7aa2f7
accent2=#bb9af7
accent3=#7dcfff

danger=#f7768e
warning=#e0af68
success=#9ece6a
```

Use it:
```bash
tinct generate -i file --file.path my-theme.txt -o hyprland,kitty
```

### Manual Palette (No File)

```bash
# Build palette entirely from command line
tinct generate -i file \
  --colour "background=#282a36" \
  --colour "foreground=#f8f8f2" \
  --colour "accent1=#ff79c6" \
  --colour "accent2=#bd93f9" \
  --colour "accent3=#8be9fd" \
  --colour "accent4=#50fa7b" \
  --colour "danger=#ff5555" \
  --colour "warning=#ffb86c" \
  --colour "success=#50fa7b" \
  -o hyprland,kitty
```

### Override Specific Colours

```bash
# Load base palette but override accent colours
tinct generate -i file --file.path catppuccin-mocha.json \
  --colour "accent1=#ff0000" \
  --colour "accent2=#00ff00" \
  -o hyprland
```

### Quick Test Colours

```bash
# Test a specific colour combination
tinct generate -i file \
  --colour "background=#000000" \
  --colour "foreground=#ffffff" \
  --colour "accent1=#ff0000" \
  --preview
```

## Use Cases

### 1. Reuse Generated Palettes

```bash
# Generate once from image
tinct extract --categorise wallpaper.jpg > my-palette.json

# Reuse many times (instant, no re-extraction)
tinct generate -i file --file.path my-palette.json -o hyprland
tinct generate -i file --file.path my-palette.json -o kitty
tinct generate -i file --file.path my-palette.json -o waybar
```

### 2. Share Themes

```bash
# Save your theme
tinct extract --categorise my-wallpaper.jpg > my-awesome-theme.json

# Share the JSON file
# Others can use: tinct generate -i file --file.path my-awesome-theme.json -o ...
```

### 3. Manual Theme Creation

```bash
# Design a theme from scratch
tinct generate -i file \
  --colour "background=#1e1e2e" \
  --colour "foreground=#cdd6f4" \
  --colour "accent1=#f38ba8" \
  --colour "accent2=#89b4fa" \
  --preview

# Adjust colours until satisfied
# Then generate for all apps
tinct generate -i file \
  --colour "background=#1e1e2e" \
  --colour "foreground=#cdd6f4" \
  --colour "accent1=#f38ba8" \
  --colour "accent2=#89b4fa" \
  -o hyprland,kitty,waybar
```

### 4. Tweak Existing Palettes

```bash
# Load base theme and adjust specific colours
tinct generate -i file --file.path catppuccin.json \
  --colour "accent1=#ff6b9d" \
  --colour "background=#16161e" \
  -o hyprland
```

## Interfaces Implemented

### Required: `Plugin`

```go
Name() string
Description() string  
Version() string
Generate(ctx, opts) (*colour.Palette, error)
RegisterFlags(cmd)
Validate() error
```

**Note:** The file plugin does NOT implement `WallpaperProvider` or `ThemeHinter` since it doesn't have a source image.

## How It Works

1. **Load File** (if `--file.path` provided)
   - Try JSON format (categorized palette)
   - Fallback to text format (hex list with optional roles)
   
2. **Parse Manual Colours** (if `--colour` provided)
   - Parse role=hex specifications
   - Create colour objects with role hints

3. **Merge** (if both provided)
   - Start with file colours
   - Override with manual specifications
   - Role-based overrides replace existing role colours

4. **Return Palette**
   - Colours with role hints (if roles were specified)
   - Or raw colours for categorisation

## Validation

The plugin validates:
- ✅ At least one source provided (file or manual colours)
- ✅ Colour specifications use `role=hex` format
- ✅ Hex colours are valid (6 or 3 characters)
- ✅ Role names are recognized

## Error Messages

### "Must provide either --file.path or --colour specifications"

**Problem:** No input provided.

**Solution:**
```bash
tinct generate -i file --file.path palette.json -o hyprland
# or
tinct generate -i file --colour "background=#000000" -o hyprland
```

### "Invalid colour format 'xxx': expected 'role=hex'"

**Problem:** Manual colour missing `=` separator.

**Solution:**
```bash
tinct generate -i file --colour "accent1=#ff0000" -o hyprland
```

### "Unknown colour role 'xxx'"

**Problem:** Role name not recognized.

**Solution:** Use one of the supported role names (see [Role Names](#role-names) section).

### "Invalid hex colour length"

**Problem:** Hex colour is not 3 or 6 characters.

**Solution:**
```bash
# Use proper hex format
--colour "accent1=#ff0000"   # 6 chars
--colour "accent1=#f00"      # 3 chars (shorthand)
```

## Performance

- **File loading:** Near-instant (< 10ms)
- **JSON parsing:** Very fast for typical palette sizes
- **Text parsing:** Line-by-line, minimal overhead
- **No extraction:** Unlike image plugin, no processing needed

## Related

- **[Input Plugin Guide](../README.md)** - Input plugin architecture
- **[Image Plugin](../image/README.md)** - Extract from images
- **[RemoteJSON Plugin](../remotejson/README.md)** - Fetch from JSON APIs
- **[Main README](../../../../README.md)** - Project overview

## Testing

```bash
# Run unit tests
go test ./internal/plugin/input/file/...

# Test with file
tinct generate -i file --file.path testdata/palette.json -o hyprland --dry-run

# Test with manual colours
tinct generate -i file \
  --colour "background=#000000" \
  --colour "foreground=#ffffff" \
  --preview
```

## Contributing

When modifying the file plugin:

1. Maintain backward compatibility with existing file formats
2. Add tests for new formats or features
3. Update this README with new functionality
4. Consider error messages for invalid input
5. Keep parsing logic simple and robust

## License

MIT License - Same as Tinct
