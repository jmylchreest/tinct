---
sidebar_position: 4
---

# gtk4

Generate colour themes for [GTK4](https://docs.gtk.org/gtk4/) applications.

## Description

GTK4 is the current version of the GTK toolkit, used by traditional (non-libadwaita) GTK4 applications. The plugin generates a CSS file that overrides application colours.

For modern GNOME applications that use libadwaita, see the [libadwaita](/docs/plugins/output/desktop/libadwaita) plugin instead.

## Output path

```
~/.config/gtk-4.0/tinct-gtk4.css
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o gtk4
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--gtk4.output-dir` | `~/.config/gtk-4.0` | Output directory for GTK4 theme files |

## Configuration

After generation, import the theme in your GTK4 configuration.

Add this line to `~/.config/gtk-4.0/gtk.css`:

```css
@import url("tinct-gtk4.css");
```

Changes will apply to newly opened GTK4 applications.

## Themed elements

The GTK4 theme affects:

- Window backgrounds
- Button colours
- Selection highlights
- Text colours
- Scrollbars
- Menu colours
- Header bars

## GTK4 vs libadwaita

| Aspect | GTK4 | libadwaita |
|--------|------|------------|
| Target apps | Traditional GTK4 apps | Modern GNOME apps |
| Styling | Custom CSS | CSS with named colours |
| Examples | Some IDEs, utilities | Files, Text Editor, Console |

Most modern GNOME applications use libadwaita rather than plain GTK4. If you're targeting GNOME apps, you likely want both plugins.

## See also

- [GTK4 documentation](https://docs.gtk.org/gtk4/)
- [libadwaita](/docs/plugins/output/desktop/libadwaita)
- [gtk3](/docs/plugins/output/desktop/gtk3)
