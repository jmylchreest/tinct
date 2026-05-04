---
sidebar_position: 1
---

# Quick start

Get started with tinct in minutes. This section covers the basics of generating themes and common workflows.

## Prerequisites

- Tinct [installed](/docs/installation)
- An image to extract colours from (wallpaper recommended)

## Your first theme

Generate a theme from your wallpaper:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all
```

This extracts colours from the image and applies them to all enabled output plugins.

## Preview before applying

Use `--preview` to see the extracted palette without writing files:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all --preview
```

## Dry run mode

Use `--dry-run` to see what files would be written:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all --dry-run
```

## Target specific applications

Instead of `all`, specify the plugins you want:

```bash
# Just terminal and window manager
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o kitty,hyprland,waybar

# Just desktop environment
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o gnome-shell,libadwaita,gtk4
```

## What happens next

After generation, tinct:

1. Writes configuration files to application-specific locations
2. Tracks all generated files in `.tinct-manifest.json`
3. Optionally triggers reload hooks (e.g., reloading Kitty, applying Hyprpaper)

See the individual [plugin documentation](/docs/plugins/overview) for reload behaviour and output paths.

## Quick start guides

- [Create your first theme](/docs/quickstart/first-theme) - Step-by-step walkthrough
- [Common workflows](/docs/quickstart/workflows) - Popular theme recipes
