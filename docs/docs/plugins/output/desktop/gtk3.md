---
sidebar_position: 3
---

# gtk3

Generate colour themes for [GTK3](https://docs.gtk.org/gtk3/) applications.

## Description

GTK3 is the older version of the GTK toolkit, still used by many applications like GIMP, Inkscape (older versions), and various legacy GNOME apps. The plugin generates a CSS file that overrides application colours.

## Output path

```
~/.config/gtk-3.0/tinct-gtk3.css
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o gtk3
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--gtk3.output-dir` | `~/.config/gtk-3.0` | Output directory for GTK3 theme files |

## Configuration

After generation, import the theme in your GTK3 configuration.

### Step 1: Import the CSS

Add this line to `~/.config/gtk-3.0/gtk.css`:

```css
@import url("tinct-gtk3.css");
```

### Step 2: Use a compatible theme

For tinct colours to apply properly, use the `adw-gtk3` theme which respects custom colour definitions:

```bash
# Arch/CachyOS
sudo pacman -S adw-gtk-theme

# Or install adw-gtk3 on other distributions
```

Then set in `~/.config/gtk-3.0/settings.ini`:

```ini
[Settings]
gtk-theme-name=adw-gtk3-dark
```

## Why adw-gtk-theme?

Standard GTK3 themes have hardcoded colours that override CSS variables. The `adw-gtk-theme` (also known as `adw-gtk3`) is designed to respect CSS colour definitions, making it compatible with tinct's generated colours.

## Themed elements

The GTK3 theme affects:

- Window backgrounds
- Button colours
- Selection highlights
- Text colours
- Scrollbars
- Menu colours

## See also

- [GTK3 CSS reference](https://docs.gtk.org/gtk3/css-overview.html)
- [gtk4](./gtk4.md)
- [libadwaita](./libadwaita.md)
