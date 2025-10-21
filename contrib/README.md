# Tinct External Plugins

Simple guide for creating and using external plugins with Tinct.

## Quick Start

### Installing Plugins

**From GitHub Releases** (recommended):
```bash
# Install compiled binary plugin
tinct plugins add https://github.com/jmylchreest/tinct/releases/download/v1.0.0/tinct-plugin-random_v1.0.0_Linux_x86_64.tar.gz:random

# Install script-based plugin
tinct plugins add https://github.com/jmylchreest/tinct/releases/download/v1.0.0/tinct-plugin-scripts_v1.0.0.tar.gz:notify-send
```

**From Local Files**:
```bash
# Add a local plugin
tinct plugins add ./contrib/notify-send.py

# List plugins
tinct plugins list

# Use a plugin
tinct generate -i image -p wallpaper.jpg -o notify-send
```

See [Plugin Packages Documentation](../docs/PLUGIN-PACKAGES.md) for detailed installation instructions.

### Creating Plugins

1. **Create executable script** (any language)
2. **Implement `--plugin-info`** to return JSON metadata
3. **Read palette from stdin** as JSON
4. **Write output to stdout**

## Plugin Protocol

### 1. Plugin Info (Required)

When called with `--plugin-info`, return JSON metadata:

```bash
./plugin.sh --plugin-info
```

Output:
```json
{
  "name": "my-plugin",
  "type": "output",
  "version": "1.0.0",
  "description": "What this plugin does",
  "author": "Plugin Author"
}
```

**Note:** The `enabled` field has been removed from the plugin protocol. All plugins (both 
built-in and external) are disabled by default and must be explicitly enabled via:
- CLI flags (`-o plugin-name` or `-i plugin-name`)
- Plugin lock file (`.tinct-plugins.json`)
- Environment variables (`TINCT_ENABLED_PLUGINS`)

This ensures explicit user control over which plugins run.

### 2. Input (stdin)

#### For Output Plugins

Tinct sends the colour palette as JSON via stdin:

```json
{
  "colours": {
    "background": {"hex": "#1a1b26", "rgb": {"r": 26, "g": 27, "b": 38}},
    "foreground": {"hex": "#c0caf5", "rgb": {"r": 192, "g": 202, "b": 245}},
    "accent1": {"hex": "#7aa2f7", "rgb": {"r": 122, "g": 162, "b": 247}}
  },
  "theme_type": 1,
  "all_colours": [...],
  "plugin_args": {},
  "dry_run": false
}
```

Fields:
- `colours` - Semantic colour roles (background, foreground, accents, etc.)
- `theme_type` - 0=auto, 1=dark, 2=light
- `all_colours` - Array of all extracted colours
- `plugin_args` - Custom arguments from `--plugin-args`
- `dry_run` - If true, don't write files or execute actions

#### For Input Plugins

Tinct sends generation options as JSON via stdin:

```json
{
  "verbose": false,
  "dry_run": false,
  "colour_overrides": [],
  "plugin_args": {}
}
```

Fields:
- `verbose` - Enable verbose logging to stderr
- `dry_run` - Don't perform actual generation (return empty/minimal output)
- `colour_overrides` - Manual colour specifications (e.g., ["background=#1e1e2e"])
- `plugin_args` - Custom arguments from `--plugin-args`

**Input plugins must output a complete CategorisedPalette:**

```json
{
  "colours": {
    "background": {"hex": "#1e1e2e", "rgb": {"r": 30, "g": 30, "b": 46}, ...},
    "foreground": {"hex": "#cdd6f4", "rgb": {"r": 205, "g": 214, "b": 244}, ...},
    "accent1": {"hex": "#f38ba8", ...}
  },
  "all_colours": [...],
  "theme_type": 1
}
```

Required fields for each colour:
- `hex` - Hex color code (e.g., "#1e1e2e")
- `rgb` - RGB values (r, g, b: 0-255)
- `hue` - HSL hue (0-360)
- `saturation` - HSL saturation (0-1)
- `luminance` - Luminance value (0-1)
- `role` - Semantic role name
- `is_light` - Whether the color is light (luminance > 0.5)

**Note:** The `Colour` field (Go's `color.Color` interface) will be reconstructed from RGB values automatically.

### 3. Output (stdout)

Write any output to stdout. This is displayed by Tinct.

- **Output plugins:** Write status messages, file paths, or results
- **Input plugins:** Write the complete CategorisedPalette as JSON

### 4. Errors and Logging (stderr)

Write errors and verbose logging to stderr. Exit with non-zero code on failure.

Use stderr for:
- Verbose logging (when `verbose: true`)
- Error messages
- Progress updates
- Debug information

This keeps stdout clean for the actual output (palette JSON or generated files).

## Minimal Examples

### Output Plugin

```bash
#!/bin/bash
# minimal-plugin.sh

if [ "$1" = "--plugin-info" ]; then
  cat <<EOF
{
  "name": "minimal",
  "type": "output",
  "version": "1.0.0",
  "description": "Minimal example"
}
EOF
  exit 0
fi

# Read palette from stdin
PALETTE=$(cat)

# Extract dry-run flag
DRY_RUN=$(echo "$PALETTE" | jq -r '.dry_run // false')

# Get background colour
BG=$(echo "$PALETTE" | jq -r '.colours.background.hex // "N/A"')

if [ "$DRY_RUN" = "true" ]; then
  echo "Would process background: $BG"
else
  echo "Processing background: $BG"
  # Do actual work here
fi
```

### Input Plugin

```bash
#!/bin/bash
# minimal-input.sh

if [ "$1" = "--plugin-info" ]; then
  cat <<EOF
{
  "name": "minimal-input",
  "type": "input",
  "version": "1.0.0",
  "description": "Minimal input example"
}
EOF
  exit 0
fi

# Read input options from stdin
OPTIONS=$(cat)

# Extract flags
VERBOSE=$(echo "$OPTIONS" | jq -r '.verbose // false')
DRY_RUN=$(echo "$OPTIONS" | jq -r '.dry_run // false')

if [ "$VERBOSE" = "true" ]; then
  echo "Generating palette..." >&2
fi

if [ "$DRY_RUN" = "true" ]; then
  echo "Dry-run mode" >&2
  # Return empty palette
  echo '{"colours":{},"all_colours":[],"theme_type":0}'
  exit 0
fi

# Generate and output a complete palette
cat <<EOF
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
    }
  },
  "all_colours": [],
  "theme_type": 1
}
EOF
```

## Plugin Arguments

Users can pass custom arguments:

```bash
tinct generate -i image -p wall.jpg -o my-plugin \
  --plugin-args 'my-plugin={"format":"json","verbose":true}'
```

Access in plugin:
```bash
FORMAT=$(echo "$PALETTE" | jq -r '.plugin_args.format // "default"')
```

## Dry-Run Mode

Always check the `dry_run` flag:

```python
palette = json.load(sys.stdin)
dry_run = palette.get("dry_run", False)

if dry_run:
    print("Would write: config.conf")
else:
    with open("config.conf", "w") as f:
        f.write(content)
```

## Plugin Management

```bash
# Add plugin
tinct plugins add ./my-plugin.sh

# Enable/disable
tinct plugins enable my-plugin
tinct plugins disable my-plugin
tinct plugins clear my-plugin  # Reset to default

# Enable/disable all
tinct plugins enable all
tinct plugins disable all
tinct plugins clear all       # Clear all config

# Delete plugin
tinct plugins delete my-plugin --force
```

## Available Colours

Semantic colour roles in `colours` object:

- `background` / `backgroundMuted`
- `foreground` / `foregroundMuted`
- `accent1` / `accent2` / `accent3` / `accent4`
- `danger` / `warning` / `success` / `info`
- `notification`

Access them:
```bash
BG=$(echo "$PALETTE" | jq -r '.colours.background.hex')
FG=$(echo "$PALETTE" | jq -r '.colours.foreground.hex')
ACCENT=$(echo "$PALETTE" | jq -r '.colours.accent1.hex')
```

## Examples

### Bash Plugin

See `example-minimal.sh` for a complete bash example with:
- Plugin info handling
- Dry-run support
- Plugin arguments
- Colour extraction

### Python Output Plugin

See `notify-send.py` for a Python output plugin example with:
- Desktop notifications
- Custom arguments
- Error handling
- System integration

### Go Input Plugin

See `random.go` for a Go input plugin example with:
- Complete palette generation with semantic roles
- Input options handling (verbose, dry-run, plugin args)
- Proper CategorisedPalette output
- Compiled binary plugin
- Reproducible generation via seed

Build and use:
```bash
# Build the plugin
cd contrib
go build -o random random.go

# Test plugin info
./random --plugin-info

# Test palette generation
echo '{"plugin_args":{"count":8,"seed":12345}}' | ./random | jq .

# Test with verbose
echo '{"verbose":true}' | ./random 2>&1 | head

# Test with Tinct
tinct plugins add ./random
tinct generate -i random --preview
```

## Best Practices

1. **Always handle `--plugin-info`** - Required for plugin discovery
2. **Check `dry_run` flag** - Don't write files or execute actions in dry-run
3. **Handle `verbose` flag** - Log helpful information to stderr when true
4. **Handle missing colours gracefully** - Not all roles may be present (output plugins)
5. **Return complete palettes** - Include all required fields (input plugins)
6. **Exit with proper codes** - 0 for success, non-zero for errors
7. **Write errors/logs to stderr** - Keeps stdout clean for output
8. **Make scripts executable** - `chmod +x plugin.sh`
9. **Test standalone** - Test with sample JSON before integrating
10. **Document requirements** - List dependencies in description or comments
11. **Don't include `enabled` field** - It's ignored by the plugin manager

## Testing Plugins

### Testing Output Plugins

```bash
# Test plugin info
./output-plugin.sh --plugin-info

# Test with sample palette
echo '{"colours":{"background":{"hex":"#1a1b26"}},"all_colours":[],"theme_type":1,"dry_run":false}' | ./output-plugin.sh

# Test dry-run
echo '{"colours":{"background":{"hex":"#1a1b26"}},"all_colours":[],"theme_type":1,"dry_run":true}' | ./output-plugin.sh

# Test with plugin args
echo '{"colours":{"background":{"hex":"#1a1b26"}},"plugin_args":{"format":"json"}}' | ./output-plugin.sh
```

### Testing Input Plugins

```bash
# Test plugin info
./input-plugin.sh --plugin-info

# Test basic generation
echo '{}' | ./input-plugin.sh | jq .

# Test with verbose
echo '{"verbose":true}' | ./input-plugin.sh 2>&1

# Test dry-run
echo '{"dry_run":true}' | ./input-plugin.sh

# Test with plugin args
echo '{"plugin_args":{"seed":12345}}' | ./input-plugin.sh | jq .

# Validate output structure
echo '{}' | ./input-plugin.sh | jq 'has("colours") and has("all_colours") and has("theme_type")'
```

## Plugin Sources

Plugins can be added from:

- **GitHub Releases** (recommended): `tinct plugins add https://github.com/jmylchreest/tinct/releases/download/v1.0.0/tinct-plugin-random_v1.0.0_Linux_x86_64.tar.gz:random`
- **Local files**: `tinct plugins add ./plugin.sh`
- **HTTP URLs**: `tinct plugins add https://example.com/plugin.sh`
- **Git repos**: `tinct plugins add https://github.com/user/repo.git:path/to/plugin.sh`
- **Archives**: Automatically extracts from `.tar.gz`, `.tgz`, or `.zip` files

Plugins are copied to `~/.local/share/tinct/plugins/` and tracked in `.tinct-plugins.json`.

### Official Plugin Packages

Official plugins are distributed as separate downloadable packages:

- **Compiled plugins**: Platform-specific binaries (e.g., `random` for Go)
- **Script plugins**: Platform-independent scripts (e.g., `notify-send`, `example-minimal`)

For complete list and installation instructions, see [Plugin Packages Documentation](../docs/PLUGIN-PACKAGES.md).

## Input vs Output Plugins

### Input Plugins
- **Purpose:** Generate or extract colour palettes
- **Input:** Receive `GenerateOptions` (verbose, dry_run, colour_overrides, plugin_args)
- **Output:** Return complete `CategorisedPalette` as JSON
- **Examples:** `image` (extract from image), `file` (load from file), `random` (generate random)

### Output Plugins
- **Purpose:** Generate configuration files or perform actions with palettes
- **Input:** Receive `CategorisedPalette` with colours
- **Output:** Write files, send notifications, or display results
- **Examples:** `hyprland` (config files), `tailwind` (CSS/JS), `notify-send` (notifications)

## Plugin Distribution

### As a User

Install plugins from GitHub releases:
```bash
# Check latest release at https://github.com/jmylchreest/tinct/releases
tinct plugins add https://github.com/jmylchreest/tinct/releases/download/v1.0.0/tinct-plugin-random_v1.0.0_Linux_x86_64.tar.gz:random
```

### As a Developer

To distribute your plugin officially:

1. Add your plugin to `contrib/`
2. Update `.goreleaser.yml` to build/package it
3. Submit a pull request

See [Plugin Packages Documentation](../docs/PLUGIN-PACKAGES.md) for packaging guidelines.

## More Information

- [Plugin Packages](../docs/PLUGIN-PACKAGES.md) - Distribution and installation
- [Plugin Lock File](../docs/PLUGIN-LOCK-FILE.md) - Lock file format
- [External Plugin Execution](../docs/EXTERNAL-PLUGIN-EXECUTION.md) - How plugins are executed
- Example plugins: `contrib/example-minimal.sh`, `contrib/notify-send.py`, `contrib/random.go`
