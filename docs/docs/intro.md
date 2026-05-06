---
sidebar_position: 1
slug: /
---

# Introduction

Tinct is an extensible colour palette generator and theme manager that brings unified theming to your entire Linux environment. Extract colours from images, generate them with AI, or fetch from remote sources, then apply them consistently across 25+ applications.

## What tinct does

1. **Extracts colours** from images using K-means clustering, or generates them via AI (Google Gemini, OpenRouter)
2. **Categorises them** into 49+ semantic roles following Material Design 3 guidelines
3. **Generates configurations** for terminals, desktop environments, window managers, and more
4. **Manages themes** with portable markdown files that include embedded wallpapers

## Key features

- **Multiple input sources**: Images, AI generation, remote JSON/CSS themes, or manual specification
- **Smart categorisation**: Automatic background/foreground detection, accent assignment, and semantic role generation
- **WCAG compliance**: Contrast checking ensures accessibility standards are met
- **Theme-aware**: Automatic light/dark theme detection with appropriate colour adjustments
- **Extensible**: Plugin architecture supports both built-in Go plugins and external plugins in any language

## Supported applications

Tinct generates configurations for a wide range of applications:

| Category | Applications |
|----------|-------------|
| **Terminals** | Alacritty, Ghostty, Kitty, Konsole, Ptyxis |
| **Desktop Environments** | GNOME Shell, KDE Plasma, GTK3/GTK4, Libadwaita, Qt5/Qt6 |
| **Window Managers** | Hyprland, Hyprpaper, Hyprlock |
| **Bars & Notifications** | Waybar, Dunst, SwayOSD |
| **Launchers** | Fuzzel, Walker, Wofi |
| **Editors & Multiplexers** | Neovim, Zellij |

## How it works

```bash
# Extract colours from a wallpaper and apply to all plugins
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all
```

This single command:

1. Loads the image and extracts dominant colours using K-means clustering
2. Determines if the image suggests a dark or light theme
3. Assigns colours to semantic roles (background, foreground, accents, semantic colours)
4. Generates appropriate values for surface and container colours
5. Writes configuration files for each enabled output plugin
6. Optionally triggers reload hooks to apply changes immediately

## Next steps

- [Installation](./installation.md) - Get tinct installed on your system
- [Quick start](./quickstart/index.md) - Generate your first theme
- [Concepts](./concepts/color-extraction.md) - Understand how colour extraction works
