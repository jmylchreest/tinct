---
sidebar_position: 3
---

# Common workflows

Ready-to-use commands for popular theming scenarios.

## Use popular themes

### Catppuccin Mocha

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.mocha" \
  -o all
```

### Catppuccin Latte (light)

```bash
tinct generate -i remote-json \
  --remote-json.url "https://raw.githubusercontent.com/catppuccin/palette/main/palette.json" \
  --remote-json.query "colors.latte" \
  -o all -t light
```

### Dracula

```bash
tinct generate -i file \
  --file.colors "#282a36,#f8f8f2,#bd93f9,#ff79c6,#50fa7b,#f1fa8c,#8be9fd,#ff5555" \
  --file.hints "background=0,foreground=1,accent1=2,accent2=3,success=4,warning=5,info=6,danger=7" \
  -o all -t dark
```

### Nord

```bash
tinct generate -i file \
  --file.colors "#2e3440,#eceff4,#88c0d0,#81a1c1,#a3be8c,#ebcb8b,#5e81ac,#bf616a" \
  --file.hints "background=0,foreground=1,accent1=2,accent2=3,success=4,warning=5,info=6,danger=7" \
  -o all -t dark
```

### Gruvbox Dark

```bash
tinct generate -i file \
  --file.colors "#282828,#ebdbb2,#458588,#b16286,#98971a,#d79921,#689d6a,#cc241d" \
  --file.hints "background=0,foreground=1,accent1=2,accent2=3,success=4,warning=5,info=6,danger=7" \
  -o all -t dark
```

## AI-generated themes

### Google Gemini

Generate an image and extract colours:

```bash
export GOOGLE_API_KEY="your-api-key"

tinct generate -i google-genai \
  --ai.prompt "sunset over rolling hills of tuscany" \
  -o all
```

Get an API key from [Google AI Studio](https://aistudio.google.com/apikey). See the [google-genai plugin docs](/docs/plugins/input/google-genai) for storing your key in a secret manager.

### OpenRouter

Use various AI models (auto-selects cheapest image-capable model):

```bash
export OPENROUTER_API_KEY="your-api-key"

tinct generate -i openrouter \
  --ai.prompt "cyberpunk city at night with neon lights" \
  -o all
```

Get an API key from [OpenRouter](https://openrouter.ai/keys).

## Ambient lighting

Extract edge colours for LED strip synchronisation:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg \
  --extract-ambience \
  --regions 8 \
  -o all
```

This extracts colours from 8 positions around the image edges (top, top-right, right, etc.) for monitor bias lighting.

## Save and share themes

### Save to markdown

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg \
  -o markdown \
  --markdown.name "Mountain Sunset" \
  --markdown.include-thumbnail
```

### Restore from markdown

```bash
tinct generate -i markdown \
  --markdown.path ~/.config/tinct/themes/mountain-sunset.md \
  -o all
```

## Hyprland ecosystem

Apply theme to full Hyprland setup:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg \
  -o hyprland,hyprpaper,hyprlock,waybar,dunst,fuzzel
```

## GNOME ecosystem

Apply theme to GNOME desktop:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg \
  -o gnome-shell,gtk4,libadwaita
```

Note: Requires the User Themes extension for GNOME Shell.

## KDE Plasma ecosystem

Apply theme to KDE Plasma:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg \
  -o kde-plasma,qt5,qt6
```

After generation:

```bash
# Apply the color scheme
plasma-apply-colorscheme TinctDark  # or TinctLight
```

## Terminal-focused

Just terminals and multiplexers:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg \
  -o kitty,alacritty,ghostty,zellij
```

## Extract without generating

Just see the colour palette:

```bash
tinct extract -i image -p ~/Pictures/wallpaper.jpg --preview
```

Save palette to JSON:

```bash
tinct extract -i image -p ~/Pictures/wallpaper.jpg -o palette.json --format json
```

## Next steps

- [Concepts](/docs/concepts/color-extraction) - Understand colour extraction
- [Commands](/docs/commands) - Full command reference
- [Plugins](/docs/plugins/overview) - All plugin documentation
