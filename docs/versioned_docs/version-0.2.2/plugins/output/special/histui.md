---
sidebar_position: 3
---

# histui

Generate colour themes for [histui](https://github.com/jmylchreest/histui) notification daemon.

## Description

histui is a GTK-based notification daemon with CSS theming support. The plugin generates a CSS file with colour variable definitions that can be imported into your histui theme.

## Output path

```
~/.config/histui/themes/tinct-colors.css
```

## Usage

```bash
tinct generate -i image -p ~/wallpaper.jpg -o histui
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--histui.output-dir` | `~/.config/histui/themes` | Output directory for theme files |

## Configuration

Import the generated colours in your histui theme.

Add to your `~/.config/histui/themes/theme.css`:

```css
@import "tinct-colors.css";

/* Use the tinct colour variables */
window.notification {
    background-color: var(--background);
    border: 2px solid var(--accent1);
}

.notification-title {
    color: var(--foreground);
}

.notification-body {
    color: var(--foregroundMuted);
}

.notification-urgency-low {
    border-color: var(--info);
}

.notification-urgency-critical {
    border-color: var(--danger);
}
```

## Hot reload

histui watches theme files via inotify for hot-reload. However, imported CSS files are inlined at load time. To apply colour changes:

1. Touch your main theme file: `touch ~/.config/histui/themes/theme.css`
2. Or restart histuid: `systemctl --user restart histuid`

## CSS variables

The generated `tinct-colors.css` defines:

```css
:root {
    --background: #1e1e2e;
    --foreground: #cdd6f4;
    --foregroundMuted: #a6adc8;
    --surface: #313244;
    --surfaceContainer: #45475a;
    --accent1: #89b4fa;
    --accent2: #f5c2e7;
    --danger: #f38ba8;
    --success: #a6e3a1;
    --warning: #f9e2af;
    --info: #94e2d5;
    /* ... additional semantic colours */
}
```

## Themed elements

The theme affects:

- Notification window background
- Title and body text colours
- Border colours by urgency level
- Progress bar colours
- Action button styling

## See also

- [histui documentation](https://jmylchreest.github.io/histui/)
- [dunst](/docs/plugins/output/bars-launchers/dunst)
- [libadwaita](/docs/plugins/output/desktop/libadwaita)
