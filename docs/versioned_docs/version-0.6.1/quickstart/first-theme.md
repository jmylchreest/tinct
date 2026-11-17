---
sidebar_position: 2
---

# Create your first theme

A step-by-step guide to generating your first theme with tinct.

## Step 1: Choose your source

You need a colour source. The most common is an image:

```bash
# From a local image
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all

# From a URL (auto-cached)
tinct generate -i image -p "https://example.com/wallpaper.jpg" -o all
```

## Step 2: Preview the palette

Before applying, preview what colours tinct extracted:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all --preview
```

This shows the colour palette with role assignments without writing any files.

## Step 3: Check what will be written

Use dry-run mode to see the files that would be created:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all --dry-run
```

## Step 4: Generate and apply

When you're satisfied, run without `--preview` or `--dry-run`:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all
```

Tinct will:
- Write configuration files to each application's expected location
- Track files in `~/.config/tinct/.tinct-manifest.json`
- Trigger reload hooks for supported applications

## Step 5: Reload applications

Most applications need to be reloaded or restarted to pick up changes:

```bash
# Kitty (tinct signals it for you; this is the manual equivalent)
kill -SIGUSR1 $(pgrep kitty)

# Waybar
killall -SIGUSR2 waybar

# Hyprland (reads colour variables automatically)
# No reload needed
```

Several plugins reload the target application themselves after writing, so the
step above is usually unnecessary. A few make it opt-in via a flag:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o hyprland --hyprland.reload
```

## Step 6: Save your theme

Save the theme as a portable markdown file:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg \
  -o markdown \
  --markdown.name "My Theme" \
  --markdown.include-thumbnail
```

This creates `~/.config/tinct/themes/my-theme.md` with:
- Full colour palette
- Embedded wallpaper (base64)
- Optional thumbnail preview

## Step 7: Restore later

Restore from the saved markdown file:

```bash
tinct generate -i markdown \
  --markdown.path ~/.config/tinct/themes/my-theme.md \
  -o all
```

## Force theme type

Override automatic theme detection:

```bash
# Force dark theme
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all -t dark

# Force light theme
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all -t light
```

## Verbose output

See detailed information about what tinct is doing:

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o all -v
```

## Next steps

- [Common workflows](./workflows.md) - Popular theme recipes
- [Colour extraction](../concepts/color-extraction.md) - How palette extraction works
- [Output plugins](../plugins/output/index.md) - All supported applications
