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
  --remote-json.query "$.mocha.colors" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2,teal=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Latte (Light, warm)

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.latte.colors" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2,teal=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Frappé (Dark, cool)

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.frappe.colors" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2,teal=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Macchiato (Dark, medium)

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.macchiato.colors" \
  --remote-json.map "base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info,pink=accent1,mauve=accent2,teal=accent3" \
  -o hyprland,kitty,waybar,dunst
```

## Dracula

[Dracula](https://draculatheme.com/) is a dark theme with vibrant colours.

```bash
# Create a Dracula palette file
cat > /tmp/dracula.json << 'EOF'
{
  "background": "#282a36",
  "foreground": "#f8f8f2",
  "selection": "#44475a",
  "comment": "#6272a4",
  "red": "#ff5555",
  "orange": "#ffb86c",
  "yellow": "#f1fa8c",
  "green": "#50fa7b",
  "purple": "#bd93f9",
  "cyan": "#8be9fd",
  "pink": "#ff79c6"
}
EOF

# Apply using remote-json
tinct generate -i remote-json \
  --remote-json.url "file:///tmp/dracula.json" \
  --remote-json.map "background=background,foreground=foreground,red=danger,green=success,yellow=warning,cyan=info,pink=accent1,purple=accent2,orange=accent3" \
  -o hyprland,kitty,waybar,dunst
```

## Tokyo Night

[Tokyo Night](https://github.com/enkia/tokyo-night-vscode-theme) is a dark theme inspired by Tokyo's night skyline.

### Tokyo Night Storm

```bash
cat > /tmp/tokyo-night-storm.json << 'EOF'
{
  "background": "#24283b",
  "foreground": "#c0caf5",
  "black": "#1d202f",
  "red": "#f7768e",
  "green": "#9ece6a",
  "yellow": "#e0af68",
  "blue": "#7aa2f7",
  "magenta": "#bb9af7",
  "cyan": "#7dcfff",
  "white": "#a9b1d6",
  "brightBlack": "#414868",
  "brightRed": "#f7768e",
  "brightGreen": "#9ece6a",
  "brightYellow": "#e0af68",
  "brightBlue": "#7aa2f7",
  "brightMagenta": "#bb9af7",
  "brightCyan": "#7dcfff",
  "brightWhite": "#c0caf5"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/tokyo-night-storm.json" \
  --remote-json.map "background=background,foreground=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2" \
  -o hyprland,kitty,waybar,dunst,neovim
```

### Tokyo Night (Classic)

```bash
cat > /tmp/tokyo-night.json << 'EOF'
{
  "background": "#1a1b26",
  "foreground": "#c0caf5",
  "red": "#f7768e",
  "green": "#9ece6a",
  "yellow": "#e0af68",
  "blue": "#7aa2f7",
  "magenta": "#bb9af7",
  "cyan": "#7dcfff"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/tokyo-night.json" \
  --remote-json.map "background=background,foreground=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2" \
  -o hyprland,kitty,waybar,dunst,neovim
```

## Nord

[Nord](https://www.nordtheme.com/) is an arctic, north-bluish colour palette.

```bash
cat > /tmp/nord.json << 'EOF'
{
  "polarNight0": "#2e3440",
  "polarNight1": "#3b4252",
  "polarNight2": "#434c5e",
  "polarNight3": "#4c566a",
  "snowStorm0": "#d8dee9",
  "snowStorm1": "#e5e9f0",
  "snowStorm2": "#eceff4",
  "frost0": "#8fbcbb",
  "frost1": "#88c0d0",
  "frost2": "#81a1c1",
  "frost3": "#5e81ac",
  "aurora0": "#bf616a",
  "aurora1": "#d08770",
  "aurora2": "#ebcb8b",
  "aurora3": "#a3be8c",
  "aurora4": "#b48ead"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/nord.json" \
  --remote-json.map "polarNight0=background,snowStorm2=foreground,aurora0=danger,aurora3=success,aurora2=warning,frost1=info,frost0=accent1,frost2=accent2,aurora4=accent3" \
  -o hyprland,kitty,waybar,dunst
```

## Gruvbox

[Gruvbox](https://github.com/morhetz/gruvbox) is a retro groove colour scheme.

### Gruvbox Dark

```bash
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
  "orange": "#fe8019"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/gruvbox-dark.json" \
  --remote-json.map "bg0=background,fg1=foreground,red=danger,green=success,yellow=warning,blue=info,purple=accent1,aqua=accent2,orange=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Gruvbox Light

```bash
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
  "orange": "#d65d0e"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/gruvbox-light.json" \
  --remote-json.map "bg0=background,fg1=foreground,red=danger,green=success,yellow=warning,blue=info,purple=accent1,aqua=accent2,orange=accent3" \
  -o hyprland,kitty,waybar,dunst
```

## Solarized

[Solarized](https://ethanschoonover.com/solarized/) is a precision colour scheme for terminals and GUIs.

### Solarized Dark

```bash
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
  --remote-json.map "base03=background,base0=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2,violet=accent3" \
  -o hyprland,kitty,waybar,dunst
```

### Solarized Light

```bash
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
  --remote-json.map "base3=background,base00=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2,violet=accent3" \
  -o hyprland,kitty,waybar,dunst
```

## One Dark

[One Dark](https://github.com/atom/one-dark-syntax) is Atom's iconic dark theme.

```bash
cat > /tmp/one-dark.json << 'EOF'
{
  "background": "#282c34",
  "foreground": "#abb2bf",
  "black": "#282c34",
  "red": "#e06c75",
  "green": "#98c379",
  "yellow": "#e5c07b",
  "blue": "#61afef",
  "magenta": "#c678dd",
  "cyan": "#56b6c2",
  "white": "#abb2bf"
}
EOF

tinct generate -i remote-json \
  --remote-json.url "file:///tmp/one-dark.json" \
  --remote-json.map "background=background,foreground=foreground,red=danger,green=success,yellow=warning,blue=info,magenta=accent1,cyan=accent2" \
  -o hyprland,kitty,waybar,dunst,neovim
```

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

Higher weights influence accent priority and selection during categorization.

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
- Need precise control over colour categorization
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
