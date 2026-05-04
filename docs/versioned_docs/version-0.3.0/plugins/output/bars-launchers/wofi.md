---
sidebar_position: 6
---

# wofi

Generate colour themes for [Wofi](https://hg.sr.ht/~scoopta/wofi) application launcher.

## Description

Wofi is a GTK-based launcher/menu for wlroots-based compositors. The plugin generates CSS files that style the launcher interface.

## Output paths

```
~/.config/wofi/tinct-colors
~/.config/wofi/style.css
```

The plugin generates:
- `tinct-colors` - Colour variable definitions
- `style.css` - Complete styling

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o wofi
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--wofi.output-dir` | `~/.config/wofi` | Output directory for theme files |

## Configuration

Wofi automatically loads `style.css` from its config directory. The generated files will be used on the next Wofi invocation.

If you have custom styling, import just the colours:

```css
@import "tinct-colors";

/* Your custom styling using the tinct variables */
window {
    background-color: @background;
}
```

## Themed elements

The theme affects:

- Window background
- Input field styling
- Entry list colours
- Selection highlighting
- Scrollbar colours
- Image icons

## Generated format

### tinct-colors

```css
@define-color background #1e1e2e;
@define-color foreground #cdd6f4;
@define-color surface #313244;
@define-color accent1 #89b4fa;
@define-color accent2 #f5c2e7;
@define-color danger #f38ba8;
@define-color success #a6e3a1;
@define-color warning #f9e2af;
```

### style.css

```css
@import "tinct-colors";

window {
    margin: 0px;
    background-color: @background;
    border-radius: 8px;
    border: 2px solid @accent1;
}

#input {
    margin: 5px;
    background-color: @surface;
    color: @foreground;
    border: none;
    border-radius: 4px;
}

#entry:selected {
    background-color: @accent1;
    color: @background;
}
```

## See also

- [Wofi documentation](https://hg.sr.ht/~scoopta/wofi)
- [fuzzel](/docs/plugins/output/bars-launchers/fuzzel)
- [walker](/docs/plugins/output/bars-launchers/walker)
