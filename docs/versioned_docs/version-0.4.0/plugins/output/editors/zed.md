---
title: zed
sidebar_position: 6
plugin:
  type: output
  name: zed
  category: editors
  source: external
  app: Zed
  app_url: 'https://zed.dev'
  version: 0.1.0
  protocol_version: 0.3.0
  repository: 'https://github.com/jmylchreest/tinct'
  install: tinct plugins install zed
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/zed/themes
  generated_files:
    - tinct.json
  reload:
    method: watch
    user_action_required: false
---

# zed

Generates a complete theme file for the [Zed editor](https://zed.dev) — a fast Rust-based collaborative code editor. The output is a single JSON file conforming to Zed's v0.2.0 theme schema, mapping tinct's semantic palette onto Zed's UI tokens, syntax highlighting, terminal ANSI block, and diff/diagnostic colours. Zed auto-discovers themes by directory and reloads the active theme live when its file changes on disk.

When tinct is run with both primary and alternate themes (`--dual-theme`), the plugin emits a single JSON file containing both variants inside the `themes` array. Zed displays them as separate entries in the theme picker.

## Installation

### Via the official plugin repository

```bash
tinct plugins install zed
```

### Build from source

```bash
cd contrib/plugins/output/zed
go build -ldflags "-X main.Version=0.1.0" -o tinct-plugin-zed
install -m 0755 tinct-plugin-zed ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-zed
tinct-plugin-zed --plugin-info | jq .
```

The plugin uses tinct's go-plugin RPC protocol and is discovered automatically once it's on `$PATH`.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o zed
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.json` | `~/.config/zed/themes/tinct.json` | Complete Zed theme — UI tokens, syntax highlighting, terminal ANSI block, diff/diagnostic colours. Wrapped in Zed's v0.2.0 `themes` array. |

The plugin auto-detects Zed's config directory at both `~/.config/zed/themes` (native) and `~/.var/app/dev.zed.Zed/config/zed/themes` (Flatpak), writing to whichever exists. If both exist, both get the theme.

## Integration

**No manual configuration required to make the theme available.** Zed scans `~/.config/zed/themes/` on startup and watches it for changes. Generating the theme is enough to make it appear in Zed's theme picker.

To activate it, choose one:

### Via the command palette (recommended)

Press `Cmd+Shift+P` (macOS) / `Ctrl+Shift+P` (Linux), then run:

```
theme selector: toggle
```

then select **Tinct** from the list.

### Via Zed's settings file

```bash
$EDITOR ~/.config/zed/settings.json
```

```json
{
  "theme": "Tinct"
}
```

For dual-theme mode, point both light and dark slots:

```json
{
  "theme": {
    "mode": "system",
    "light": "Tinct LIGHT",
    "dark": "Tinct DARK"
  }
}
```

## Reload behaviour

### Automatic

Zed watches its `themes` directory and reloads the active theme live when its file changes. After the initial selection, regenerating with tinct propagates immediately to any running Zed window — no restart, no re-selection.

### Manual fallback

If a regeneration doesn't appear:

- Confirm Zed has the theme selected (`grep theme ~/.config/zed/settings.json`).
- Reload the theme manually via the command palette: `theme selector: toggle` → select **Tinct** again.
- Restart Zed if all else fails.

## Uninstall / revert

1. **Remove the config line** from `~/.config/zed/settings.json` if you set `"theme": "Tinct"` — pick a different theme, or remove the key to fall back to Zed's default:

   ```bash
   $EDITOR ~/.config/zed/settings.json
   ```

2. **Delete the generated file**:

   ```bash
   rm ~/.config/zed/themes/tinct.json
   # also remove from Flatpak install if present:
   rm -f ~/.var/app/dev.zed.Zed/config/zed/themes/tinct.json
   ```

3. **Reload**: Zed picks up the file deletion live via its themes-dir watcher. If the active theme was Tinct, Zed will switch to its default — re-select your preferred theme via the command palette.

4. **External state**: this plugin only writes to Zed's `themes/` directory. To also remove the plugin binary:

   ```bash
   tinct plugins uninstall zed
   # or, for a source build:
   rm ~/.local/bin/tinct-plugin-zed
   ```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--zed.output-dir` / `-o` | auto-detected (`~/.config/zed/themes`) | Override the output directory |

## Colour role mapping

Zed's v0.2.0 theme schema is large — ~100 tokens covering editor chrome, syntax, terminal, diagnostics, diffs, and Git status. Selected mappings:

### UI chrome

| Zed token | Tinct role |
|---|---|
| `background` | `background` |
| `surface.background` / `elevated_surface.background` | `surface` |
| `element.background` | `surfaceContainer` |
| `element.hover` | `surfaceContainerHigh` |
| `element.active` / `element.selected` | `surfaceContainerHighest` |
| `border` | `outline` |
| `border.variant` | `outlineVariant` |
| `border.focused` | `accent1` |
| `border.disabled` | `borderMuted` |
| `text` | `foreground` |
| `text.muted` | `foregroundMuted` |
| `text.accent` | `accent1` |
| `panel.background` / `tab_bar.background` | `surfaceContainerLowest` |
| `tab.active_background` / `toolbar.background` | `surfaceContainerLow` |

### Diagnostics & status

| Zed token | Tinct role |
|---|---|
| `error` | `danger` |
| `warning` | `warning` |
| `success` | `success` |
| `info` / `hint` | `info` |

### Terminal ANSI

All 16 ANSI colours (`terminal.ansi.black`–`white` and `bright_*`) use tinct's perceptual ANSI matcher so terminal output sits close to standard expectations regardless of palette source.

### Syntax highlighting

| Zed token | Tinct role |
|---|---|
| `comment` | `foregroundMuted` (italic) |
| `keyword` | `accent4` |
| `function` | `accent1` |
| `string` | `success` |
| `type` | `accent2` |
| `constant` | `accent3` |
| `operator` | `onSurface` |
| `variable` | `foreground` |

## Customising the template

```bash
mkdir -p ~/.config/tinct/templates/zed
tinct plugins templates dump -o zed -l ~/.config/tinct/templates/zed
$EDITOR ~/.config/tinct/templates/zed/theme.json.tmpl
```

Tinct prefers your version over the embedded default. See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### Plugin skipped: "Zed not installed (config directory does not exist)"

```bash
ls ~/.config/zed 2>/dev/null
ls ~/.var/app/dev.zed.Zed/config/zed 2>/dev/null
```

The plugin requires either the native or Flatpak Zed config directory to exist. If you've installed Zed in a non-standard location, pass `--zed.output-dir` explicitly.

### Theme not appearing in Zed's picker

```bash
ls -l ~/.config/zed/themes/tinct.json
jq '.themes[].name' ~/.config/zed/themes/tinct.json
```

If the file exists and is valid JSON, the picker scans on each toggle — open `theme selector: toggle` again. If the file is missing, the plugin was skipped: run with `--verbose` to see detection logs.

### JSON parse errors in Zed

```bash
jq . ~/.config/zed/themes/tinct.json
```

A template syntax error from a custom override is the usual cause. Remove the override or fix it; if you didn't customise the template, file an issue with the verbose output.

## Related plugins

- [opencode](./opencode.md) — AI-assisted coding agent themed the same way.
- [neovim](./neovim.md), [helix](./helix.md) — other editors tinct themes.
