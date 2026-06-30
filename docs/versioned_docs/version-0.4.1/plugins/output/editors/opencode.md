---
title: opencode
sidebar_position: 5
plugin:
  type: output
  name: opencode
  category: editors
  source: external
  app: OpenCode
  app_url: 'https://opencode.ai'
  version: 0.1.0
  protocol_version: 0.3.0
  repository: 'https://github.com/jmylchreest/tinct'
  install: tinct plugins install opencode
  requires: []
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/opencode/themes
  generated_files:
    - tinct.json
  reload:
    method: none
    user_action_required: true
---

# opencode

Generates a complete theme file for [OpenCode](https://opencode.ai), the AI-assisted coding agent. The output is a single JSON file mapping tinct's semantic palette onto OpenCode's ~50 theme tokens (UI, syntax highlighting, diff colours, markdown rendering). OpenCode auto-detects installed themes by directory, so the file appears in `/theme` as soon as it's written.

## Installation

### Via the official plugin repository

```bash
tinct plugins install opencode
```

### Build from source

```bash
cd contrib/plugins/output/opencode
go build -ldflags "-X main.Version=0.1.0" -o tinct-plugin-opencode
install -m 0755 tinct-plugin-opencode ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-opencode
tinct-plugin-opencode --plugin-info | jq .
```

The plugin uses tinct's go-plugin RPC protocol and is discovered automatically once it's on `$PATH`.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o opencode
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.json` | `~/.config/opencode/themes/tinct.json` | Complete OpenCode theme — colour `defs`, UI tokens, syntax highlighting, diff/markdown rendering |

If you're running tinct in dual-theme mode (`--dual-theme`), a second file is generated:

| File | Path | Role |
|------|------|------|
| `tinct-dark.json` *or* `tinct-light.json` | same directory | The alternate theme variant matched to the primary |

The plugin auto-detects OpenCode's config directory at both `$XDG_CONFIG_HOME/opencode` and `~/.config/opencode`, writing to whichever exists. If both exist, both get the theme.

## Integration

**No manual configuration required to make the theme available.** OpenCode scans `~/.config/opencode/themes/` on startup and lists installed themes by filename. Generating the theme is enough to make it appear in OpenCode's theme picker.

To activate it, choose one:

### Via OpenCode's command palette (recommended)

```
/theme
```

then select **tinct** from the list.

### Via OpenCode's config file

```bash
$EDITOR ~/.config/opencode/opencode.json
```

```json
{
  "$schema": "https://opencode.ai/config.json",
  "theme": "tinct"
}
```

## Reload behaviour

### Automatic

None. OpenCode does not currently support theme hot-reload. After regenerating the theme, the file on disk is up to date but a running OpenCode session will keep using the previously loaded copy. See the [upstream feature request](https://github.com/anomalyco/opencode/issues/815).

### Manual fallback

Re-select the theme in OpenCode to pick up changes:

```
/theme
```

then choose **tinct** again. Or restart OpenCode.

**Tip:** If you also theme your terminal with tinct, set OpenCode's theme to `system` (built-in) instead — it adapts to your terminal's ANSI colours in real time, so terminal regenerations propagate automatically.

## Uninstall / revert

1. **Remove the config line** from `~/.config/opencode/opencode.json` if you set `"theme": "tinct"`. OpenCode falls back to its default theme.

2. **Delete the generated file(s)**:

   ```bash
   rm ~/.config/opencode/themes/tinct.json
   rm -f ~/.config/opencode/themes/tinct-dark.json ~/.config/opencode/themes/tinct-light.json
   ```

3. **Reload to drop the theme**: restart OpenCode, or use `/theme` to select a different theme.

4. **External state**: this plugin only writes to `~/.config/opencode/themes/`. To also remove the plugin binary:

   ```bash
   tinct plugins uninstall opencode
   # or, for a source build:
   rm ~/.local/bin/tinct-plugin-opencode
   ```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--opencode.output-dir` | auto-detected (`~/.config/opencode/themes`) | Override the output directory |

## Colour role mapping

OpenCode's theme uses a `defs` section for reusable colour references and a `theme` section that maps UI tokens to those defs.

### Colour definitions (`defs`)

| Definition | Tinct role |
|---|---|
| `background` | `background` |
| `foreground` | `foreground` |
| `surface` | `surface` |
| `accent1`–`accent4` | `accent1`–`accent4` |
| `danger` | `danger` |
| `warning` | `warning` |
| `success` | `success` |
| `info` | `info` |

### UI tokens

| OpenCode token | Tinct role |
|---|---|
| `primary` | `accent1` |
| `secondary` | `accent2` |
| `accent` | `accent3` |
| `error` | `danger` |
| `warning` | `warning` |
| `success` | `success` |
| `info` | `info` |
| `text` | `foreground` |
| `textMuted` | `foregroundMuted` |
| `background` | `background` |
| `backgroundPanel` | `surfaceContainerLow` |
| `backgroundElement` | `surfaceContainer` |
| `border` | `outline` |
| `borderActive` | `accent1` |
| `borderSubtle` | `outlineVariant` |

### Diff colours

| OpenCode token | Tinct role |
|---|---|
| `diffAdded` | `success` |
| `diffRemoved` | `danger` |
| `diffContext` | `onSurfaceVariant` |
| `diffHunkHeader` | `accent2` |
| `diffAddedBg` | `success` @ 15% alpha |
| `diffRemovedBg` | `danger` @ 15% alpha |

### Markdown rendering

| OpenCode token | Tinct role |
|---|---|
| `markdownHeading` | `accent1` |
| `markdownLink` | `accent2` |
| `markdownCode` | `success` |
| `markdownEmph` | `accent4` |
| `markdownStrong` | `accent1` |

### Syntax highlighting

| OpenCode token | Tinct role |
|---|---|
| `syntaxComment` | `foregroundMuted` |
| `syntaxKeyword` | `accent4` |
| `syntaxFunction` | `accent1` |
| `syntaxString` | `success` |
| `syntaxNumber` | `accent3` |
| `syntaxType` | `accent2` |
| `syntaxVariable` | `foreground` |
| `syntaxOperator` | `onSurface` |
| `syntaxPunctuation` | `onSurfaceVariant` |

## Customising the template

```bash
mkdir -p ~/.config/tinct/templates/opencode
cp templates/theme.json.tmpl ~/.config/tinct/templates/opencode/
$EDITOR ~/.config/tinct/templates/opencode/theme.json.tmpl
```

Tinct prefers your version over the embedded default. See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### Plugin not found by tinct

```bash
which tinct-plugin-opencode
tinct plugins list | grep opencode
```

If the binary isn't on `$PATH`, either reinstall via `tinct plugins install opencode` or symlink the source build into `~/.local/bin/`.

### Theme not appearing in OpenCode's `/theme` picker

```bash
ls -l ~/.config/opencode/themes/tinct.json
jq . ~/.config/opencode/themes/tinct.json | head
```

If the file exists and is valid JSON, restart OpenCode — the theme list is scanned at startup. If the file is missing, the plugin was skipped: this usually means tinct couldn't find an OpenCode config directory. Run with `--verbose` to see the detection log:

```bash
tinct generate -i image -p ~/wallpaper.jpg -o opencode --verbose
```

### JSON parse errors in OpenCode

Validate the generated file:

```bash
jq . ~/.config/opencode/themes/tinct.json
```

A template syntax error from a custom override is the usual cause. Remove or fix the override; if you didn't customise the template, file an issue with the verbose output.

## Related plugins

- [neovim](./neovim.md) — Neovim colourscheme.
- [ghostty](../terminals/ghostty.md) — terminal theme that pairs well with OpenCode's `system` theme.
