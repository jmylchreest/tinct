---
title: swaync
sidebar_position: 10
plugin:
  type: output
  name: swaync
  category: bars-launchers
  source: builtin
  app: SwayNotificationCenter
  app_url: 'https://github.com/ErikReider/SwayNotificationCenter'
  requires:
    - swaync-client
  optional: []
  pattern: include-file
  default_output_dir: ~/.config/swaync
  generated_files:
    - tinct.css
  reload:
    method: ipc
    command: org.erikreider.swaync.cc.ReloadCss
    user_action_required: false
---

# swaync

Generates a colour overlay for the GTK4 version of [SwayNotificationCenter](https://github.com/ErikReider/SwayNotificationCenter). Uses its default stylesheet's CSS custom properties to theme notifications and the control centre while preserving layout. Older GTK3 releases using `@define-color` need a custom template.

## Installation

Built into tinct. Install SwayNotificationCenter separately.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o swaync
```

## Generated files

`tinct.css` is written to `~/.config/swaync/`. Your `style.css` and `config.json` are preserved.

## Integration

Add this import after other colour rules in your `~/.config/swaync/style.css`:

```css
@import "tinct.css";
```

Keep the default SwayNC stylesheet or your existing layout rules. The generated file supplies colour overrides, not a complete layout. Later rules can override these colours. If using a non-default stylesheet location, adjust the import path.

## Reload behaviour

After generation tinct calls SwayNC’s `ReloadCss` D-Bus method directly with a three-second deadline and service activation disabled. A stopped daemon is left stopped. This avoids the client’s connection wait loop, which can occur even with `--skip-wait`.

When SwayNC is running, use `swaync-client --reload-css --skip-wait` to reload manually after adding the import. Changing its stylesheet location requires restarting it.

## Flags

| Flag | Default | Description |
|---|---|---|
| `--swaync.output-dir` | `~/.config/swaync` | Directory for the colour overlay |

The output directory also supports `[plugin.swaync] output_dir` in `tinct.toml`.

## Colour role mapping

| SwayNC property | Tinct role |
|---|---|
| Control centre background | background |
| Notification background | surface |
| Hover/focus background | surfaceVariant |
| Text / disabled text | foreground / foregroundMuted |
| Selection | accent1 |
| Critical notification border | danger |

## Customising the template

```bash
tinct plugins templates dump -o swaync -l ~/.config/tinct/templates/swaync
```

Edit `tinct.css.tmpl` there to override the embedded template.

## Uninstall / revert

Remove the import from `style.css`, delete the generated `tinct.css`, and run `swaync-client --reload-css --skip-wait`.

## Troubleshooting

If colours do not change, check the import path and whether later rules override the variables. GTK3 releases cannot use this GTK4 template. Custom stylesheets that do not consume SwayNC's standard variables need corresponding custom rules.
