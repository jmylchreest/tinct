---
sidebar_position: 5
---

# walker

Generate colour themes for [Walker](https://github.com/abenz1267/walker) application launcher.

## Description

Walker is a modern, GTK-based application launcher for Wayland. The plugin generates CSS files that style the launcher interface.

## Output paths

```
~/.config/walker/themes/tinct-colours.css
~/.config/walker/themes/tinct/style.css
```

The plugin generates:
- `tinct-colours.css` - CSS colour variable definitions
- `tinct/style.css` - Complete theme styling

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o walker
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--walker.output-dir` | `~/.config/walker/themes` | Output directory for theme files |

## Configuration

Set the theme in your Walker configuration.

Edit `~/.config/walker/config.toml`:

```toml
[ui]
theme = "tinct"
```

Or to use just the colours with your own styling:

```toml
[ui]
css = "themes/tinct-colours.css"
```

## Themed elements

The theme affects:

- Window background
- Search input styling
- Result list colours
- Selection highlighting
- Category headers
- Scrollbar colours

## Generated format

### tinct-colours.css

```css
:root {
    --background: #1e1e2e;
    --foreground: #cdd6f4;
    --surface: #313244;
    --accent1: #89b4fa;
    --accent2: #f5c2e7;
    /* ... additional colours */
}
```

### tinct/style.css

```css
@import url("../tinct-colours.css");

window {
    background-color: var(--background);
    border-radius: 12px;
}

#search {
    background-color: var(--surface);
    color: var(--foreground);
}

.item:selected {
    background-color: var(--accent1);
}
```

## See also

- [Walker documentation](https://github.com/abenz1267/walker)
- [fuzzel](./fuzzel.md)
- [wofi](./wofi.md)
