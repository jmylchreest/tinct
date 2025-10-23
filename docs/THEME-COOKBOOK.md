# Theme Cookbook 

A collection of ready-to-use commands for popular color themes. Copy, paste, and enjoy beautiful terminal themes!

---

## Quick Reference

| Theme | Variants | Plugin | Command |
|-------|----------|--------|---------|
| [Catppuccin](#catppuccin) | Mocha, Macchiato, Frappé, Latte | `remote-json` | See below |
| [Dracula](#dracula) | Original | `remote-json` | See below |
| [Tokyo Night](#tokyo-night) | Storm, Night, Moon | `remote-json` | See below |
| [Nord](#nord) | Original | `remote-json` | See below |
| [Gruvbox](#gruvbox) | Dark, Light | `remote-json` | See below |
| [Solarized](#solarized) | Dark, Light | `remote-json` | See below |

---

## Catppuccin

Beautiful pastel theme with four flavors.

### Mocha (Dark) 

Warm dark theme with rich colors.

```bash
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.mocha.colors" \
  --remote-json.map base=background,mantle=backgroundMuted,text=foreground,subtext0=foregroundMuted,red=danger,peach=warning,green=success,sky=info,sapphire=accent1,blue=accent2,mauve=accent3,pink=accent4 \
  --theme dark \
  --outputs waybar
```

**Colors:**
- Background: `#1e1e2e` (Base)
- Foreground: `#cdd6f4` (Text)
- Accent: `#89b4fa` (Blue)

### Macchiato (Dark) 

Slightly lighter than Mocha with vibrant accents.

```bash
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.macchiato.colors" \
  --remote-json.map base=background,mantle=backgroundMuted,text=foreground,subtext0=foregroundMuted,red=danger,peach=warning,green=success,sky=info,sapphire=accent1,blue=accent2,mauve=accent3,pink=accent4 \
  --theme dark \
  --outputs waybar
```

**Colors:**
- Background: `#24273a` (Base)
- Foreground: `#cad3f5` (Text)
- Accent: `#8aadf4` (Blue)

### Frappé (Dark) 

Cool and refreshing dark theme.

```bash
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.frappe.colors" \
  --remote-json.map base=background,mantle=backgroundMuted,text=foreground,subtext0=foregroundMuted,red=danger,peach=warning,green=success,sky=info,sapphire=accent1,blue=accent2,mauve=accent3,pink=accent4 \
  --theme dark \
  --outputs waybar
```

**Colors:**
- Background: `#303446` (Base)
- Foreground: `#c6d0f5` (Text)
- Accent: `#8caaee` (Blue)

### Latte (Light) 

Warm light theme for daytime use.

```bash
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.latte.colors" \
  --remote-json.map base=background,mantle=backgroundMuted,text=foreground,subtext0=foregroundMuted,red=danger,peach=warning,green=success,sky=info,sapphire=accent1,blue=accent2,mauve=accent3,pink=accent4 \
  --theme light \
  --outputs waybar
```

**Colors:**
- Background: `#eff1f5` (Base)
- Foreground: `#4c4f69` (Text)
- Accent: `#1e66f5` (Blue)

### Save Catppuccin for Reuse

```bash
# Save Mocha palette
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.mocha.colors" \
  --remote-json.map base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info \
  --save-palette ~/.config/tinct/catppuccin-mocha.txt

# Use saved palette
tinct generate --input file \
  --file.path ~/.config/tinct/catppuccin-mocha.txt \
  --theme dark \
  --outputs hyprland,kitty,waybar
```

---

## Dracula

A dark theme with high contrast and vibrant colors.

### Original (Dark) 

```bash
# Using a community JSON source
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/dracula/dracula-theme/master/colors.json" \
  --remote-json.map background=background,foreground=foreground,red=danger,orange=warning,green=success,cyan=info,purple=accent1,pink=accent2 \
  --theme dark \
  --outputs waybar
```

**Colors:**
- Background: `#282a36`
- Foreground: `#f8f8f2`
- Purple: `#bd93f9`
- Pink: `#ff79c6`

### Manual Dracula (Alternative)

If the JSON source isn't available, create a palette file:

```bash
cat > ~/.config/tinct/dracula.txt << 'EOF'
background=#282a36
foreground=#f8f8f2
accent1=#bd93f9
accent2=#ff79c6
accent3=#8be9fd
accent4=#50fa7b
danger=#ff5555
warning=#ffb86c
success=#50fa7b
info=#8be9fd
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/dracula.txt \
  --theme dark \
  --outputs waybar
```

---

## Tokyo Night

A clean, dark theme inspired by the city at night.

### Storm (Dark) 

Deeper, stormier variant.

```bash
cat > ~/.config/tinct/tokyo-night-storm.txt << 'EOF'
background=#24283b
foreground=#c0caf5
accent1=#7aa2f7
accent2=#bb9af7
accent3=#7dcfff
accent4=#9ece6a
danger=#f7768e
warning=#e0af68
success=#9ece6a
info=#7dcfff
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/tokyo-night-storm.txt \
  --theme dark \
  --outputs waybar
```

### Night (Dark) 

The original, most popular variant.

```bash
cat > ~/.config/tinct/tokyo-night.txt << 'EOF'
background=#1a1b26
foreground=#c0caf5
accent1=#7aa2f7
accent2=#bb9af7
accent3=#7dcfff
accent4=#9ece6a
danger=#f7768e
warning=#e0af68
success=#9ece6a
info=#7dcfff
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/tokyo-night.txt \
  --theme dark \
  --outputs waybar
```

### Moon (Dark) 

Softer, more muted variant.

```bash
cat > ~/.config/tinct/tokyo-night-moon.txt << 'EOF'
background=#222436
foreground=#c8d3f5
accent1=#82aaff
accent2=#c099ff
accent3=#86e1fc
accent4=#c3e88d
danger=#ff757f
warning=#ffc777
success=#c3e88d
info=#86e1fc
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/tokyo-night-moon.txt \
  --theme dark \
  --outputs waybar
```

---

## Nord

An arctic, north-bluish color palette.

### Original (Dark) 

```bash
cat > ~/.config/tinct/nord.txt << 'EOF'
background=#2e3440
backgroundMuted=#3b4252
foreground=#eceff4
foregroundMuted=#d8dee9
accent1=#88c0d0
accent2=#81a1c1
accent3=#5e81ac
accent4=#8fbcbb
danger=#bf616a
warning=#ebcb8b
success=#a3be8c
info=#81a1c1
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/nord.txt \
  --theme dark \
  --outputs waybar
```

**Colors:**
- Background: `#2e3440` (Polar Night)
- Foreground: `#eceff4` (Snow Storm)
- Accent: `#88c0d0` (Frost)

---

## Gruvbox

Retro groove colors with warm, earthy tones.

### Dark (Hard Contrast) 

```bash
cat > ~/.config/tinct/gruvbox-dark.txt << 'EOF'
background=#1d2021
backgroundMuted=#282828
foreground=#ebdbb2
foregroundMuted=#d5c4a1
accent1=#fb4934
accent2=#fabd2f
accent3=#b8bb26
accent4=#83a598
danger=#fb4934
warning=#fabd2f
success=#b8bb26
info=#83a598
notification=#d3869b
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/gruvbox-dark.txt \
  --theme dark \
  --outputs waybar
```

### Light (Hard Contrast) 

```bash
cat > ~/.config/tinct/gruvbox-light.txt << 'EOF'
background=#f9f5d7
backgroundMuted=#ebdbb2
foreground=#3c3836
foregroundMuted=#504945
accent1=#cc241d
accent2=#d79921
accent3=#98971a
accent4=#458588
danger=#cc241d
warning=#d79921
success=#98971a
info=#458588
notification=#b16286
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/gruvbox-light.txt \
  --theme light \
  --outputs waybar
```

---

## Solarized

Precision colors for machines and people.

### Dark 

```bash
cat > ~/.config/tinct/solarized-dark.txt << 'EOF'
background=#002b36
backgroundMuted=#073642
foreground=#fdf6e3
foregroundMuted=#eee8d5
accent1=#268bd2
accent2=#2aa198
accent3=#859900
accent4=#b58900
danger=#dc322f
warning=#cb4b16
success=#859900
info=#268bd2
notification=#6c71c4
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/solarized-dark.txt \
  --theme dark \
  --outputs waybar
```

### Light 

```bash
cat > ~/.config/tinct/solarized-light.txt << 'EOF'
background=#fdf6e3
backgroundMuted=#eee8d5
foreground=#002b36
foregroundMuted=#073642
accent1=#268bd2
accent2=#2aa198
accent3=#859900
accent4=#b58900
danger=#dc322f
warning=#cb4b16
success=#859900
info=#268bd2
notification=#6c71c4
EOF

tinct generate --input file \
  --file.path ~/.config/tinct/solarized-light.txt \
  --theme light \
  --outputs waybar
```

---

## Creating Pre-Hook Scripts

For themes you use frequently, create a pre-hook script:

### Example: Catppuccin Mocha Script

**`~/.config/tinct/hooks/catppuccin-mocha.sh`**
```bash
#!/bin/bash
# Fetch Catppuccin Mocha and save to temp file
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.mocha.colors" \
  --remote-json.map base=background,text=foreground,red=danger,green=success,yellow=warning,blue=info \
  --save-palette /tmp/catppuccin-mocha.txt \
  --dry-run > /dev/null 2>&1
```

**Usage:**
```bash
chmod +x ~/.config/tinct/hooks/catppuccin-mocha.sh

tinct generate \
  --pre-hook ~/.config/tinct/hooks/catppuccin-mocha.sh \
  --input file \
  --file.path /tmp/catppuccin-mocha.txt \
  --theme dark \
  --outputs waybar
```

---

## Tips & Tricks

### Apply Multiple Outputs at Once

```bash
tinct generate --input file \
  --file.path ~/.config/tinct/catppuccin-mocha.txt \
  --theme dark \
  --outputs hyprland,kitty,waybar
```

### Preview Before Applying

```bash
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.colors.mocha" \
  --remote-json.map base=background,text=foreground \
  --theme dark \
  --preview \
  --dry-run
```

### Save All Your Favorite Themes

Create a themes directory:
```bash
mkdir -p ~/.config/tinct/themes
```

Save themes:
```bash
# Catppuccin Mocha
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.mocha.colors" \
  --remote-json.map base=background,text=foreground \
  --save-palette ~/.config/tinct/themes/catppuccin-mocha.txt

# Catppuccin Latte
tinct generate --input remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "$.latte.colors" \
  --remote-json.map base=background,text=foreground \
  --save-palette ~/.config/tinct/themes/catppuccin-latte.txt
```

Switch themes easily:
```bash
# Day theme
tinct generate --input file \
  --file.path ~/.config/tinct/themes/catppuccin-latte.txt \
  --theme light \
  --outputs waybar

# Night theme
tinct generate --input file \
  --file.path ~/.config/tinct/themes/catppuccin-mocha.txt \
  --theme dark \
  --outputs waybar
```

### Create Your Own Theme Switcher

**`~/.local/bin/theme-switch`**
```bash
#!/bin/bash
THEME_DIR="$HOME/.config/tinct/themes"

case "$1" in
  dark|night)
    tinct generate --input file \
      --file.path "$THEME_DIR/catppuccin-mocha.txt" \
      --theme dark \
      --outputs hyprland,kitty,waybar
    ;;
  light|day)
    tinct generate --input file \
      --file.path "$THEME_DIR/catppuccin-latte.txt" \
      --theme light \
      --outputs hyprland,kitty,waybar
    ;;
  *)
    echo "Usage: theme-switch [dark|light]"
    exit 1
    ;;
esac
```

---

## Color Extraction from Wallpapers

Combine with image extraction for dynamic themes:

```bash
# Extract from wallpaper
tinct extract --image.path ~/wallpaper.jpg \
  --format palette \
  --output ~/.config/tinct/themes/wallpaper.txt

# Apply extracted theme
tinct generate --input file \
  --file.path ~/.config/tinct/themes/wallpaper.txt \
  --theme auto \
  --outputs waybar
```

---

## Troubleshooting

### Colors Look Wrong

- Ensure you're using the correct `--theme` flag (dark/light/auto)
- Check that your terminal supports true color
- Verify output plugin is writing to the correct config location

### Remote Fetch Fails

- Check your internet connection
- Verify the URL is accessible
- Some URLs may require specific user-agent headers (not currently supported)

### Role Mappings Don't Match

- Check the source palette structure (use `--preview` to see colors)
- Verify color names match the source exactly (case-sensitive)
- Some themes use different naming conventions

---

## Contributing

Have a theme to add? Submit a PR with:
1. Theme name and description
2. Working `tinct generate` command
3. Color preview (background, foreground, accents)
4. Links to official theme sources

---

## Resources

- [Catppuccin](https://github.com/catppuccin/catppuccin)
- [Dracula](https://draculatheme.com/)
- [Tokyo Night](https://github.com/enkia/tokyo-night-vscode-theme)
- [Nord](https://www.nordtheme.com/)
- [Gruvbox](https://github.com/morhetz/gruvbox)
- [Solarized](https://ethanschoonover.com/solarized/)

---

**Happy theming! **