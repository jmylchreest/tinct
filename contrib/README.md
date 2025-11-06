# Contrib Directory

This directory contains **community contributions, examples, and external resources** for Tinct that are not part of the core application but are useful for users and developers.

## Contents

### 📦 [`plugins/`](plugins/)
External plugins that extend Tinct's functionality. These plugins run as separate processes and can be written in any language.

**Types:**
- **Input plugins** (`plugins/input/`): Generate or extract color palettes from various sources
- **Output plugins** (`plugins/output/`): Apply palettes to applications, generate configs, or control external devices

**Distribution:**
- Compiled Go plugins are built and packaged by GoReleaser as separate downloads
- Script-based plugins (Python, Bash) are packaged together in the `tinct-plugin-scripts` archive
- Users can install via `tinct plugins add <source>` or download from GitHub releases

See [plugins/README.md](plugins/README.md) for the complete external plugin development guide.

### 🎨 [`palettes/`](palettes/)
Example color palette definitions in various formats that can be used with Tinct's input plugins.

**Use with:**
```bash
tinct generate -i file -p contrib/palettes/catppuccin-mocha.json -o hyprland,kitty
```

## What Goes in Contrib?

### ✅ Belongs in Contrib
- **External plugins** - Standalone executables that extend Tinct via the plugin protocol
- **Example configurations** - Sample configs, palettes, or templates for users to reference
- **Integration scripts** - Scripts that help integrate Tinct with other tools
- **Experimental features** - New functionality being tested before core integration
- **Third-party integrations** - Plugins for external services or proprietary tools

### ❌ Belongs in Core (`internal/`)
- **Built-in plugins** - Plugins compiled into the main `tinct` binary
- **Core functionality** - Essential features all users need
- **Stable APIs** - Well-tested interfaces used by the core application

## External vs Built-in Plugins

### External Plugins (contrib/plugins/)
- **Distribution:** Separate downloads, user-installable
- **Languages:** Any language (Go, Python, Bash, etc.)
- **Execution:** Run as separate processes
- **Location:** `contrib/plugins/`
- **Installation:** Via `tinct plugins add` command
- **Examples:** `random`, `wob`, `templater`, `dunstify`

### Built-in Plugins (internal/plugin/)
- **Distribution:** Compiled into main `tinct` binary
- **Language:** Go only
- **Execution:** Direct function calls (faster)
- **Location:** `internal/plugin/input/` or `internal/plugin/output/`
- **Installation:** Included by default
- **Examples:** `image`, `hyprland`, `kitty`, `waybar`

## Directory Structure

```
contrib/
├── README.md              # This file
├── plugins/               # External plugins
│   ├── README.md          # External plugin development guide
│   ├── input/             # Input plugins
│   │   └── random/        # Random palette generator (Go)
│   └── output/            # Output plugins
│       ├── example-minimal.sh     # Minimal bash example
│       ├── notify-send.py         # Desktop notifications (Python)
│       ├── wled-ambient.sh        # WLED LED control
│       ├── openrgb-peripheral.sh  # OpenRGB device control
│       ├── dunstify/              # Dunst notification plugin (Go)
│       ├── templater/             # Generic template engine (Go)
│       └── wob/                   # Wayland overlay bar (Go)
└── palettes/              # Example palette files
    └── (example palettes in JSON/CSS format)
```

## Contributing

### Adding a New Plugin

1. **Choose location:**
   - External plugin → `contrib/plugins/input/` or `contrib/plugins/output/`
   - Built-in plugin → `internal/plugin/input/` or `internal/plugin/output/`

2. **Create plugin directory:**
   ```bash
   mkdir -p contrib/plugins/output/my-plugin
   cd contrib/plugins/output/my-plugin
   ```

3. **Implement plugin protocol:**
   - Must respond to `--plugin-info` with JSON metadata
   - For output plugins: read `CategorisedPalette` from stdin
   - For input plugins: output `CategorisedPalette` to stdout
   - See [plugins/README.md](plugins/README.md) for details

4. **Add README.md:**
   - Document what the plugin does
   - List requirements and dependencies
   - Provide usage examples
   - Include installation instructions

5. **Add to GoReleaser (optional):**
   - Edit `.goreleaser.yml` to build and package your plugin
   - Compiled plugins get separate architecture-specific downloads
   - Script plugins are bundled in `tinct-plugin-scripts` archive

6. **Submit PR:**
   - Include tests if applicable
   - Update this README to list your plugin
   - Add to [docs/PLUGINS-WISHLIST.md](../docs/PLUGINS-WISHLIST.md) if it's a new plugin type

### Plugin Guidelines

- **Make it executable:** `chmod +x plugin.sh`
- **Handle `--plugin-info`:** Required for discovery
- **Support dry-run:** Check `dry_run` flag from stdin
- **Use stderr for logs:** Keep stdout clean for JSON output
- **Exit with proper codes:** 0 for success, non-zero for errors
- **Document dependencies:** List required tools/libraries
- **Include examples:** Show how to test and use the plugin

## Examples

### Installing External Plugins

```bash
# From GitHub releases (recommended)
tinct plugins add https://github.com/jmylchreest/tinct/releases/download/v1.0.0/tinct-plugin-random_v1.0.0_Linux_x86_64.tar.gz:random

# From local file
tinct plugins add ./contrib/plugins/output/notify-send.py

# From contrib directory
cd contrib/plugins/output
tinct plugins add ./example-minimal.sh
```

### Using External Plugins

```bash
# Generate with external input plugin
tinct generate -i random --random.seed 12345 -o hyprland,kitty

# Use external output plugin
tinct generate -i image -p wallpaper.jpg -o notify-send

# Combine built-in and external
tinct generate -i image -p wallpaper.jpg -o hyprland,kitty,notify-send,wob
```

## Resources

- **[Plugin Development Guide](plugins/README.md)** - How to create external plugins
- **[External Plugins Documentation](../docs/external-plugins.md)** - Comprehensive external plugin guide
- **[Built-in Plugins](../internal/plugin/README.md)** - Core plugin documentation
- **[Main README](../README.md)** - Project overview and usage
- **[Development Guide](../docs/DEVELOPMENT.md)** - Contributing to Tinct

## Questions?

- Check existing plugin examples in this directory
- Read the [plugin development guide](plugins/README.md)
- Open an issue on GitHub
- See [PLUGINS-WISHLIST.md](../docs/PLUGINS-WISHLIST.md) for planned plugins
