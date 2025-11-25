# Tinct Plugins Reference

Complete reference of all built-in and available plugins for tinct.

## Input Plugins

Input plugins extract or generate color palettes from various sources.

| Plugin | Description | Key Features |
|--------|-------------|--------------|
| **image** | Extract colors from images | JPEG, PNG, GIF, WebP support; Ambient edge/corner extraction; Configurable sampling methods; K-means clustering |
| **google-genai** | Generate images with Google Gemini | AI image generation with Imagen; Automatic color extraction; Wallpaper integration |
| **remote-json** | Fetch from JSON URLs | JSONPath query support; Flexible color format parsing; Theme repository integration |
| **remote-css** | Extract from CSS files | CSS variable parsing; Hex color extraction; Remote URL fetching |
| **file** | Load from saved palettes | Manual color specification; Role assignments; Saved palette loading |

## Output Plugins

Output plugins generate configuration files for applications and devices.

| Plugin | Description | Notes |
|--------|-------------|-------|
| **alacritty** | Alacritty terminal emulator | TOML/YAML configuration |
| **dunst** | Dunst notification daemon | INI-style configuration |
| **fuzzel** | Fuzzel application launcher | INI-style configuration |
| **ghostty** | Ghostty terminal emulator | Config file with palette support |
| **gnome-shell** | GNOME Shell desktop theme | Automatic gsettings application; User theme extension support |
| **gtk3** | GTK3 applications | CSS themes; Requires adw-gtk3-theme |
| **gtk4** | GTK4 applications | CSS themes for traditional GTK4 apps |
| **hyprland** | Hyprland compositor | Color variable configuration |
| **hyprlock** | Hyprlock screen locker | Colors and wallpaper integration |
| **hyprpaper** | Hyprpaper wallpaper manager | Automatic wallpaper application |
| **kde-plasma** | KDE Plasma desktop environment | `.colors` scheme files; Light/dark theme detection |
| **kitty** | Kitty terminal emulator | Config file with 256-color palette |
| **libadwaita** | Libadwaita applications | CSS themes for modern GNOME apps |
| **ptyxis** | Ptyxis terminal emulator | Palette files with window dressing support |
| **neovim** | Neovim text editor | Lua color scheme generation |
| **qt5** | Qt5 applications | qt5ct configuration; Requires QT_QPA_PLATFORMTHEME=qt5ct |
| **qt6** | Qt6 applications | qt6ct configuration; Requires QT_QPA_PLATFORMTHEME=qt6ct |
| **swayosd** | SwayOSD on-screen display | SCSS configuration |
| **walker** | Walker application launcher | JSON configuration |
| **waybar** | Waybar status bar | CSS styling |
| **wofi** | Wofi application launcher | CSS styling |
| **zed** | Zed code editor | JSON theme files (external plugin) |
| **zellij** | Zellij terminal multiplexer | KDL configuration |

## Plugin Details

### Qt/KDE Plugins

Tinct provides comprehensive support for Qt and KDE Plasma theming:

#### qt5 Plugin

Generates color schemes for Qt5 applications using qt5ct.

**Requirements:**
- qt5ct installed
- `QT_QPA_PLATFORMTHEME=qt5ct` environment variable set
- `~/.config/qt5ct/` directory exists

**Usage:**
```bash
# Generate Qt5 theme
tinct generate -i image -p wallpaper.jpg -o qt5

# Select theme in qt5ct GUI
qt5ct
```

**Output:** `~/.config/qt5ct/colors/tinct.conf`

**Format:** 22 colors per state (active/inactive/disabled) in #AARRGGBB format

#### qt6 Plugin

Generates color schemes for Qt6 applications using qt6ct.

**Requirements:**
- qt6ct installed
- `QT_QPA_PLATFORMTHEME=qt6ct` environment variable set
- `~/.config/qt6ct/` directory exists

**Usage:**
```bash
# Generate Qt6 theme
tinct generate -i image -p wallpaper.jpg -o qt6

# Select theme in qt6ct GUI
qt6ct
```

**Output:** `~/.config/qt6ct/colors/tinct.conf`

**Format:** 22 colors per state (active/inactive/disabled) in #AARRGGBB format

#### kde-plasma Plugin

Generates global color schemes for KDE Plasma desktop environment.

**Requirements:**
- KDE Plasma installed
- Either `~/.config/kdeglobals` or `~/.config/plasmarc` exists

**Usage:**
```bash
# Generate KDE Plasma color scheme
tinct generate -i image -p wallpaper.jpg -o kde-plasma

# Apply via System Settings or command line
plasma-apply-colorscheme TinctDark  # or TinctLight
```

**Output:** `~/.local/share/color-schemes/TinctDark.colors` (or `TinctLight.colors`)

**Format:** INI-style `.colors` file with RGB decimal values

**Color Sections:**
- Colors:Button
- Colors:Complementary
- Colors:Header
- Colors:Selection
- Colors:Tooltip
- Colors:View
- Colors:Window
- ColorEffects (Disabled/Inactive)
- WM (Window Manager colors)

### GNOME Plugins

#### gnome-shell Plugin

Generates GNOME Shell theme with automatic application.

**Requirements:**
- GNOME Shell installed
- User Themes extension installed and enabled

**Installation:**
```bash
# Fedora/RHEL
sudo dnf install gnome-shell-extension-user-theme

# Ubuntu/Debian
sudo apt install gnome-shell-extension-user-theme

# Arch
sudo pacman -S gnome-shell-extension-user-theme

# Or via browser
# Visit https://extensions.gnome.org/extension/19/user-themes/
```

After installation, enable the extension:
```bash
gnome-extensions enable user-theme@gnome-shell-extensions.gcampax.github.com
```

Then log out and back in (or restart GNOME Shell on X11 with `Alt+F2` → `r`).

**Usage:**
```bash
# Generate GNOME Shell theme
tinct generate -i image -p wallpaper.jpg -o gnome-shell
```

**Features:**
- Automatic gsettings application
- User theme extension detection and validation
- Wallpaper integration
- Flicker-free theme reloading via A/B theme switching

**Output:** 
- `~/.local/share/themes/tinct-a/gnome-shell/gnome-shell.css`
- `~/.local/share/themes/tinct-b/gnome-shell/gnome-shell.css`

(User Themes extension searches `~/.themes` and `~/.local/share/themes` - both locations are valid)

**How Theme Reloading Works:**

The plugin generates two identical themes (`tinct-a` and `tinct-b`) and alternates between them when updating. This approach:

1. Avoids visual flashing that would occur when switching to the default theme
2. Forces the user-theme extension to reload the CSS (which only watches gsettings changes, not file changes)
3. Provides seamless theme updates without requiring manual shell restarts

The user-theme extension monitors the `org.gnome.shell.extensions.user-theme name` gsettings key. When this value changes, it reloads the theme CSS. By toggling between `tinct-a` and `tinct-b`, tinct can update the theme appearance without the user seeing any intermediate default theme.

#### ptyxis Plugin

Generates color palettes for the Ptyxis terminal emulator with window dressing support.

**Requirements:**
- Ptyxis terminal installed

**Usage:**
```bash
# Generate Ptyxis palette
tinct generate -i image -p wallpaper.jpg -o ptyxis
```

**Features:**
- Complete 16-color ANSI palette support
- Light and dark theme variants
- Window dressing (headerbar theming based on terminal state)
- Bell notification styling
- Remote session indicators
- Superuser/root session highlighting

**Output:**
- Native: `~/.local/share/org.gnome.Ptyxis/palettes/tinct.palette`
- Flatpak (Stable): `~/.var/app/app.devsuite.Ptyxis/data/ptyxis/palettes/tinct.palette`
- Flatpak (Nightly): `~/.var/app/org.gnome.Ptyxis.Devel/data/ptyxis/palettes/tinct.palette`

**Window Dressing:**

Ptyxis supports "window dressing" - coloring the window headerbar based on terminal state:

- **BellBackground:** Colors headerbar when terminal bell is triggered (uses warning color)
- **RemoteBackground:** Colors headerbar during remote SSH sessions (uses info color)
- **SuperuserBackground:** Colors headerbar when running as root (uses danger color)

These visual indicators help distinguish terminal states at a glance.

**Note:** Ptyxis uses its own palette-based theming system. The libadwaita plugin does not affect Ptyxis's headerbar - window chrome theming is controlled by the palette file itself through the special background fields.

**Applying the Palette:**

After generation, select the "Tinct" palette in Ptyxis preferences. For Ptyxis 49+, you can also install palettes via command line:

```bash
ptyxis --install-palette ~/.local/share/org.gnome.Ptyxis/palettes/tinct.palette
```

#### gtk3 / gtk4 / libadwaita Plugins

Generate CSS themes for GTK applications.

**gtk3:** Legacy GTK3 apps (GIMP, older apps)
**gtk4:** Traditional GTK4 apps
**libadwaita:** Modern GNOME apps (GNOME Settings, Files, etc.)

**Note:** GTK3 requires `adw-gtk3-theme` to be installed for proper styling.

### Hyprland Ecosystem

Tinct provides full integration with the Hyprland ecosystem:

- **hyprland:** Color variables for compositor
- **hyprpaper:** Wallpaper management with auto-application
- **hyprlock:** Lock screen theming with wallpaper support

### Terminal Emulators

All terminal plugins support 256-color palettes with ANSI color mapping:

- **alacritty:** TOML configuration
- **ghostty:** Modern terminal with palette support
- **kitty:** Advanced terminal with 256-color palette
- **ptyxis:** GNOME terminal with window dressing support (see detailed section above)

### LED & External Devices

While tinct doesn't include built-in hardware control plugins, it provides the foundation for creating them:

**Available Examples (in contrib/plugins/):**
- WLED ambient lighting
- OpenRGB peripheral control
- Philips Hue integration

**How to Use:**
```bash
# Extract positional colors for ambient lighting
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8 \
  -o wled-ambient
```

See [External Plugins Guide](../contrib/plugins/README.md) for creating device controllers.

## Plugin Configuration

### Environment Variables

```bash
# Enable only specific plugins
export TINCT_ENABLED_PLUGINS="image,hyprland,kitty,qt5"

# Disable specific plugins
export TINCT_DISABLED_PLUGINS="waybar,dunst"
```

### Lock File Configuration

Plugins can be configured via `.tinct-plugins.json`:

```json
{
  "version": "1",
  "enabled_plugins": ["image", "hyprland", "kitty", "qt5", "kde-plasma"],
  "disabled_plugins": ["waybar"],
  "external_plugins": {
    "zed": {
      "name": "zed",
      "path": "/home/user/.local/share/tinct/plugins/tinct-plugin-zed",
      "type": "output",
      "version": "0.1.0"
    }
  }
}
```

## Plugin Management Commands

```bash
# List all plugins
tinct plugins list

# Install external plugin
tinct plugins install <github-user>/<repo>

# Add plugin repository
tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/refs/heads/main/repository/repository.json

# Sync plugins
tinct plugins sync
```

## Creating Custom Plugins

Tinct supports two types of plugins:

1. **Built-in plugins (Go):** Compiled into the binary for maximum performance
2. **External plugins (Any language):** Standalone executables using JSON-stdio protocol

**See also:**
- [Plugin Development Guide](DEVELOPMENT.md) - Creating built-in Go plugins
- [External Plugins Guide](../contrib/plugins/README.md) - Creating plugins in any language
- [Plugin Protocols](PLUGIN-PROTOCOLS.md) - Technical protocol specifications
- [Plugin Hooks](plugin-hooks.md) - Pre/post-execution hooks

## Plugin Status Legend

When running `tinct plugins list`, you'll see status indicators:

- **O** = On-demand (loaded when used)
- **E** = Enabled (always loaded)
- **D** = Disabled (not loaded)
- **\*** = External plugin
- **Y** = Compatible with current tinct version
- **N** = Incompatible with current tinct version
