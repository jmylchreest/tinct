# Theme Cookbook

Pre-built commands for applying popular colour themes to your desktop using Tinct's remote palette fetching capabilities.

## Table of Contents

- [Catppuccin](#catppuccin)
- [Dracula](#dracula)
- [Tokyo Night](#tokyo-night)
- [Nord](#nord)
- [Gruvbox](#gruvbox)
- [Solarized](#solarized)
- [One Dark](#one-dark)
- [Custom Themes](#custom-themes)

## Catppuccin

[Catppuccin](https://github.com/catppuccin/catppuccin) is a soothing pastel theme with multiple flavours.

### Mocha (Dark, warm)

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.mocha.colours" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2,teal=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Latte (Light, warm)

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.latte.colours" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2,teal=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Frappé (Dark, cool)

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.frappe.colours" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2,teal=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Macchiato (Dark, medium)

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.macchiato.colours" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2,teal=accent3" \
  -o hyprland,kitty,waybar,dunst
```

## Dracula

[Dracula](https://draculatheme.com/) is a dark theme with vibrant colours designed for reduced eye strain.

```bash
# Apply Dracula using official CSS variables from community gist
tinct generate -i remote-css \
  --remote-css.url "https://gist.githubusercontent.com/alexanderbazo/a689dcb972b84ed17721d14fe0a9ee89/raw/dracula.css" \
  --remote-css.map "Background=background,Foreground=foreground,Red=danger,Green=success,Yellow=warning,Cyan=info,Pink=accent1,Purple=accent2,Orange=accent3" \
  -o hyprland,kitty,waybar,dunst
```

**Colour Palette:**
- **Background colours**: Dark gray (`#282a36`) with lighter currentLine (`#44475a`)
- **Text colours**: Off-white foreground (`#f8f8f2`) with muted blue comments (`#6272a4`)
- **Accent colours**: Vibrant cyan, green, orange, pink, purple, red, and yellow

## Tokyo Night

[Tokyo Night](https://github.com/tokyo-night/tokyo-night-vscode-theme) is a dark theme by Enkia celebrating the lights of Downtown Tokyo at night, optimized for extended coding sessions.

**Note**: The official Tokyo Night theme JSON files contain extensive semantic tokens. For Tinct, we'll use simplified mappings of the core colours. For the complete theme, you can create a local JSON with the extracted colours below.

### Tokyo Night (Classic)

```bash
# Create local palette from official theme colours
cat > /tmp/tokyo-night.json << 'EOF'
{
  "background": "#1a1b26",
  "foreground": "#c0caf5",
  "red": "#f7768e",
  "green": "#9ece6a",
  "yellow": "#e0af68",
  "blue": "#7aa2f7",
  "magenta": "#bb9af7",
  "cyan": "#7dcfff",
  "comment": "#565f89"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/tokyo-night.json" \
  --remote-json.map "background=background,foreground=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2" \
  -o hyprland,kitty,waybar,dunst,neovim
```

### Tokyo Night Storm

```bash
# Storm variant with slightly lighter background
cat > /tmp/tokyo-night-storm.json << 'EOF'
{
  "background": "#24283b",
  "foreground": "#c0caf5",
  "red": "#f7768e",
  "green": "#9ece6a",
  "yellow": "#e0af68",
  "blue": "#7aa2f7",
  "magenta": "#bb9af7",
  "cyan": "#7dcfff",
  "comment": "#565f89"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/tokyo-night-storm.json" \
  --remote-json.map "background=background,foreground=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2" \
  -o hyprland,kitty,waybar,dunst,neovim
```

**Colour Reference** (from [official JSON](https://github.com/tokyo-night/tokyo-night-vscode-theme/blob/master/themes/tokyo-night-colour-theme.json)):
- Keywords: `#bb9af7` (purple), Strings: `#9ece6a` (green), Numbers: `#ff9e64` (orange)
- Functions: `#7aa2f7` (blue), Variables: `#c0caf5` (light blue), Comments: `#565f89` (muted)

## Nord

[Nord](https://www.nordtheme.com/) is an arctic, north-bluish colour palette with 16 carefully selected colours organized into four thematic groups.

```bash
# Apply Nord using official CSS from GitHub
tinct generate -i remote-css \
  --remote-css.url "https://raw.githubusercontent.com/nordtheme/nord/develop/src/nord.css" \
  --remote-css.map "nord0=background,nord6=foreground,nord11=danger,nord14=success,nord13=warning,nord8=info,nord7=accent1,nord9=accent2,nord15=accent3,nord10=accent4" \
  -o hyprland,kitty,waybar,dunst
```

**Colour Mapping Explained:**
- **Polar Night** (`nord0-3`): Dark backgrounds and UI elements → `nord0` as background
- **Snow Storm** (`nord4-6`): Light text colours → `nord6` (brightest) as foreground
- **Frost** (`nord7-10`): Bluish accent colours → mapped to primary accents and info
- **Aurora** (`nord11-15`): Colourful semantic colours → mapped to danger/success/warning

## Gruvbox

[Gruvbox](https://github.com/morhetz/gruvbox) is a retro groove colour scheme with high-contrast pastel colours designed for eye comfort.

**Note**: Gruvbox's canonical colour definitions are in the Vim colour scheme file. These colours are extracted from the [official repository](https://github.com/morhetz/gruvbox/blob/master/colours/gruvbox.vim).

### Gruvbox Dark

```bash
# Create palette from official Gruvbox dark (medium) colours
cat > /tmp/gruvbox-dark.json << 'EOF'
{
  "bg0": "#282828",
  "bg1": "#3c3836",
  "bg2": "#504945",
  "bg3": "#665c54",
  "fg0": "#fbf1c7",
  "fg1": "#ebdbb2",
  "fg2": "#d5c4a1",
  "fg3": "#bdae93",
  "red": "#fb4934",
  "green": "#b8bb26",
  "yellow": "#fabd2f",
  "blue": "#83a598",
  "purple": "#d3869b",
  "aqua": "#8ec07c",
  "orange": "#fe8019",
  "gray": "#928374"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/gruvbox-dark.json" \
  --remote-json.map "bg0=background,fg1=foreground,red=danger,green=success,yellow=warning,blue=info,purple=accent1,aqua=accent2,orange=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Gruvbox Light

```bash
# Create palette from official Gruvbox light (medium) colours
cat > /tmp/gruvbox-light.json << 'EOF'
{
  "bg0": "#fbf1c7",
  "bg1": "#ebdbb2",
  "bg2": "#d5c4a1",
  "bg3": "#bdae93",
  "fg0": "#282828",
  "fg1": "#3c3836",
  "fg2": "#504945",
  "fg3": "#665c54",
  "red": "#cc241d",
  "green": "#98971a",
  "yellow": "#d79921",
  "blue": "#458588",
  "purple": "#b16286",
  "aqua": "#689d6a",
  "orange": "#d65d0e",
  "gray": "#928374"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/gruvbox-light.json" \
  --remote-json.map "bg0=background,fg1=foreground,red=danger,green=success,yellow=warning,blue=info,purple=accent1,aqua=accent2,orange=accent3" \
  -o hyprland,kitty,waybar,dunst
```

**Colour Variants**: Gruvbox provides `hard`, `medium`, and `soft` variants with different background contrasts. The above uses `medium`. For `hard` use `#1d2021` (dark) or `#f9f5d7` (light) as `bg0`.

## Solarized

[Solarized](https://ethanschoonover.com/solarized/) by Ethan Schoonover is a precision colour scheme with sixteen colours (eight monotones, eight accents) engineered with precise CIELAB lightness relationships.

**Note**: Colours extracted from the [official specification](https://github.com/altercation/solarized). Solarized uses symmetric contrast across light and dark modes.

### Solarized Dark

```bash
# Create palette from official Solarized specification
cat > /tmp/solarized-dark.json << 'EOF'
{
  "base03": "#002b36",
  "base02": "#073642",
  "base01": "#586e75",
  "base00": "#657b83",
  "base0": "#839496",
  "base1": "#93a1a1",
  "base2": "#eee8d5",
  "base3": "#fdf6e3",
  "yellow": "#b58900",
  "orange": "#cb4b16",
  "red": "#dc322f",
  "magenta": "#d33682",
  "violet": "#6c71c4",
  "blue": "#268bd2",
  "cyan": "#2aa198",
  "green": "#859900"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/solarized-dark.json" \
  --remote-json.map "base03=background,base0=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2,violet=accent3,orange=accent4" \
  -o hyprland,kitty,waybar,dunst
```

### Solarized Light

```bash
# Create palette from official Solarized specification (same colours, different mapping)
cat > /tmp/solarized-light.json << 'EOF'
{
  "base03": "#002b36",
  "base02": "#073642",
  "base01": "#586e75",
  "base00": "#657b83",
  "base0": "#839496",
  "base1": "#93a1a1",
  "base2": "#eee8d5",
  "base3": "#fdf6e3",
  "yellow": "#b58900",
  "orange": "#cb4b16",
  "red": "#dc322f",
  "magenta": "#d33682",
  "violet": "#6c71c4",
  "blue": "#268bd2",
  "cyan": "#2aa198",
  "green": "#859900"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/solarized-light.json" \
  --remote-json.map "base3=background,base00=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2,violet=accent3,orange=accent4" \
  -o hyprland,kitty,waybar,dunst
```

**Design Notes**: 
- Dark mode uses `base03` (background) with `base0` (body text)
- Light mode uses `base3` (background) with `base00` (body text)
- The eight accent colours maintain consistent appearance across both modes

## One Dark

[One Dark](https://github.com/atom/one-dark-syntax) is Atom's iconic dark theme, included by default with the Atom editor.

**Note**: Colours extracted from the [official Atom syntax theme](https://github.com/atom/one-dark-syntax/blob/master/styles/colours.less). The theme uses HSL values internally which have been converted to hex.

```bash
# Create palette from official One Dark colours
cat > /tmp/one-dark.json << 'EOF'
{
  "background": "#282c34",
  "foreground": "#abb2bf",
  "comment": "#5c6370",
  "red": "#e06c75",
  "green": "#98c379",
  "yellow": "#e5c07b",
  "blue": "#61afef",
  "cyan": "#56b6c2",
  "magenta": "#c678dd",
  "orange": "#d19a66"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/one-dark.json" \
  --remote-json.map "background=background,foreground=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2,orange=accent3" \
  -o hyprland,kitty,waybar,dunst,neovim
```

**Colour Reference** (from [official source](https://github.com/atom/one-dark-syntax/blob/master/styles/colours.less)):
- `mono-1` (#abb2bf): Default text colour
- `hue-1` (#56b6c2): Cyan accent
- `hue-2` (#61afef): Blue accent  
- `hue-3` (#c678dd): Purple accent
- `hue-4` (#98c379): Green accent
- `hue-5` (#e06c75): Red accent
- `hue-6` (#d19a66): Orange accent

## Real-World CSS Extraction

### Extracting from GitHub Primer CSS

[GitHub Primer](https://primer.style/) is GitHub's open-source design system (MIT licensed) with semantic CSS variables. This is an excellent example of CSS variable mapping:

```bash
# Extract and preview colours from Primer CSS
tinct generate -i remote-css \
  --remote-css.url "https://unpkg.com/@primer/css/dist/primer.css" \
  --preview
```

Primer uses semantic CSS custom properties. You can map these variables to Tinct roles:

```bash
# Apply GitHub Primer colours with proper CSS variable mapping
tinct generate -i remote-css \
  --remote-css.url "https://unpkg.com/@primer/css/dist/primer.css" \
  --remote-css.map "bgColor-default=background,fgColor-default=foreground,fgColor-accent=accent1,bgColor-accent-emphasis=accent2,bgColor-danger-emphasis=danger,bgColor-success-emphasis=success" \
  -o hyprland,kitty,waybar,dunst
```

The mapping format is `css-var-name=role` where:
- `css-var-name` is the CSS variable name **without** the `--` prefix (e.g., `bgColor-default` for `--bgColor-default`)
- `role` is a Tinct semantic role (background, foreground, accent1, danger, success, etc.)

Primer's semantic naming makes it easy to map: `bgColor-*` for backgrounds, `fgColor-*` for text, with modifiers like `-default`, `-muted`, `-emphasis`.

### Extracting from DaisyUI Themes

[DaisyUI](https://daisyui.com/) is an open-source Tailwind component library with built-in themes (MIT licensed):

```bash
# Create a local JSON with DaisyUI's dark theme colours
cat > /tmp/daisyui-dark.json << 'EOF'
{
  "primary": "#661AE6",
  "secondary": "#D926AA",
  "accent": "#1FB2A5",
  "neutral": "#191D24",
  "base-100": "#2A303C",
  "info": "#3ABFF8",
  "success": "#36D399",
  "warning": "#FBBD23",
  "error": "#F87272"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/daisyui-dark.json" \
  --remote-json.map "base-100=background,neutral=foreground,error=danger,success=success,warning=warning,info=info,primary=accent1,secondary=accent2,accent=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Manual Palette Files

For complete control over your palette, you can create manual colour files with weights and role hints.

#### Simple Manual Palette (Just Weights)

The simplest approach - let Tinct auto-categorize colours, but control their importance:

```bash
cat > /tmp/my-palette.txt << 'EOF'
#1e1e2e:weight=10
#cdd6f4:weight=10
#f38ba8:weight=8
#a6e3a1:weight=8
#f9e2af:weight=7
#89dceb:weight=6
#f5c2e7:weight=5
#cba6f7:weight=4
#94e2d5:weight=3
EOF

tinct generate -i file --file.path /tmp/my-palette.txt \
  -o hyprland,kitty,waybar,dunst
```

Higher weights influence accent priority and selection during categorisation.

#### Complex Manual Palette (Roles + Weights + Positions)

For precise control over every aspect of the palette:

```bash
cat > /tmp/advanced-palette.txt << 'EOF'
# Core theme colours with explicit roles
background=#1e1e2e
foreground=#cdd6f4

# Semantic colours with roles and weights
#f38ba8:weight=10:role=danger
#a6e3a1:weight=10:role=success
#f9e2af:weight=10:role=warning
#89dceb:weight=10:role=info

# Accent colours with priority weights
#f5c2e7:weight=8:role=accent1
#cba6f7:weight=7:role=accent2
#94e2d5:weight=6:role=accent3

# Positional colours for ambient lighting
#3a2f4b:role=topLeft
#4a3f5b:role=top
#5a4f6b:role=topRight
#6a5f7b:role=right
#7a6f8b:role=bottomRight
#8a7f9b:role=bottom
#9a8fab:role=bottomLeft
#aa9fbb:role=left

# Additional accent colours (auto-categorized by weight)
#b4befe:weight=5
#fab387:weight=4
#eba0ac:weight=3
EOF

tinct generate -i file --file.path /tmp/advanced-palette.txt \
  -o hyprland,kitty,waybar,dunst,wled-ambient,openrgb-peripheral \
  --plugin-args 'wled-ambient={"host":"192.168.1.100"},openrgb-peripheral={"devices":["keyboard"]}'
```

This approach is perfect when you:
- Need precise control over colour categorisation
- Want to specify positional colours for LED strips
- Are migrating from another theme system
- Have specific accent priorities in mind

#### Custom Local Theme JSON

For your own theme palette using JSON format:

```bash
cat > ~/.config/tinct/my-theme.json << 'EOF'
{
  "background": "#1e1e2e",
  "foreground": "#cdd6f4",
  "danger": "#f38ba8",
  "success": "#a6e3a1",
  "warning": "#f9e2af",
  "info": "#89dceb",
  "accent1": "#f5c2e7",
  "accent2": "#cba6f7",
  "accent3": "#94e2d5"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file://$HOME/.config/tinct/my-theme.json" \
  -o hyprland,kitty,waybar,dunst
```

## Tips

### Quick Theme Testing

Use `--preview` to see how a theme looks before applying:

```bash
tinct generate -i remote-json \
  --remote-json.url "file:///tmp/dracula.json" \
  --remote-json.map "background=background,foreground=foreground,red=danger,green=success" \
  --preview
```

### Combining Themes with Images

Extract ambient colours from a wallpaper and combine with a theme:

```bash
# First, extract and preview positional colours from wallpaper
tinct extract -i image -p wallpaper.jpg --image.extractAmbience --categorise --preview

# Then apply a theme base with wallpaper-derived ambient lighting
tinct generate -i image -p wallpaper.jpg \
  --image.extractAmbience \
  --image.ambienceRegions 8 \
  -o wled-ambient,openrgb-peripheral \
  --plugin-args 'wled-ambient={"host":"192.168.1.100"}'
```

### All Plugins at Once

Apply a theme to your entire environment:

```bash
tinct generate -i remote-json \
  --remote-json.url "file:///tmp/catppuccin-mocha.json" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2" \
  -o hyprland,kitty,waybar,dunst,fuzzel,hyprlock,swayosd,wofi,neovim,zellij
```

### With External Devices

Sync your theme to LED peripherals:

```bash
# Apply Dracula to desktop + keyboard/mouse
tinct generate -i remote-json \
  --remote-json.url "file:///tmp/dracula.json" \
  --remote-json.map "background=background,foreground=foreground,pink=accent1,purple=accent2" \
  -o hyprland,kitty,waybar,openrgb-peripheral \
  --plugin-args 'openrgb-peripheral={"devices":["keyboard","mouse"]}'
```

## Contributing

Have a favourite theme that's missing? Please contribute by:
1. Testing the theme with Tinct
2. Adding the recipe to this cookbook
3. Submitting a pull request

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.
