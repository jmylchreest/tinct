---
sidebar_position: 5
---

# libadwaita

Generate colour themes for [libadwaita](https://gnome.pages.gitlab.gnome.org/libadwaita/) (modern GNOME) applications.

## Description

Libadwaita is the GNOME-specific library that provides the modern GNOME application look and feel. Most current GNOME applications use libadwaita, including Files (Nautilus), Text Editor, Console, Settings, and many others.

The plugin generates CSS that overrides libadwaita's named colour variables, providing consistent theming across all GNOME applications.

## Output path

```
~/.config/gtk-4.0/tinct-libadwaita.css
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o libadwaita
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--libadwaita.output-dir` | `~/.config/gtk-4.0` | Output directory for libadwaita theme files |

## Configuration

After generation, import the theme in your GTK4 configuration.

Add this line to `~/.config/gtk-4.0/gtk.css`:

```css
@import url("tinct-libadwaita.css");
```

Changes will apply to GNOME/libadwaita applications.

## Themed elements

The libadwaita theme affects:

- Application backgrounds
- Header bars and title bars
- Buttons and controls
- Accent colours
- Selection highlights
- Sidebar colours
- Card backgrounds
- Dialog styling

## Named colour variables

Libadwaita uses named colour variables that tinct maps to semantic roles:

| libadwaita Variable | tinct Role |
|---------------------|------------|
| `--accent-color` | accent1 |
| `--accent-bg-color` | accent1 |
| `--destructive-color` | danger |
| `--success-color` | success |
| `--warning-color` | warning |
| `--window-bg-color` | background |
| `--view-bg-color` | surface |
| `--card-bg-color` | surfaceContainer |
| `--headerbar-bg-color` | surfaceContainerHigh |
| `--sidebar-bg-color` | surfaceContainerLow |

## Complete GNOME theming

For comprehensive GNOME desktop theming, combine multiple plugins:

```bash
tinct generate -i image -p ~/wallpaper.jpg \
  -o libadwaita \
  -o gnome-shell \
  -o gtk3
```

This themes:
- Modern GNOME apps (libadwaita)
- GNOME Shell (panel, overview)
- Legacy GTK3 apps (GIMP, etc.)

## See also

- [libadwaita documentation](https://gnome.pages.gitlab.gnome.org/libadwaita/)
- [GNOME Human Interface Guidelines](https://developer.gnome.org/hig/)
- [gnome-shell](./gnome-shell.md)
- [gtk4](./gtk4.md)
