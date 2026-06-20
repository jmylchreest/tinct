---
title: yazi
sidebar_position: 1
plugin:
  type: output
  name: yazi
  category: special
  source: builtin
  app: Yazi
  app_url: 'https://yazi-rs.github.io/'
  requires:
    - yazi
  optional: []
  pattern: flavor-pack
  default_output_dir: ~/.config/yazi/flavors/tinct.yazi
  generated_files:
    - flavor.toml
  reload:
    method: none
    user_action_required: true
---

# yazi

Generates a [Yazi](https://yazi-rs.github.io/) **flavor pack** — Yazi's directory-based theme format. The output is a directory named `<name>.yazi/` containing `flavor.toml`. Yazi loads flavor packs by name from `~/.config/yazi/flavors/` when its `theme.toml` references them.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o yazi` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o yazi
```

The first time, you also need to point Yazi at the flavor — see [Integration](#integration).

## Generated files

| File | Path | Role |
|------|------|------|
| `flavor.toml` | `~/.config/yazi/flavors/tinct.yazi/flavor.toml` | Yazi flavor pack manifest. Defines colours for the manager pane, status bar, modes, selection, input/confirm/pick prompts, which-key, help, tasks, notifications, and per-filetype rules. |

The flavor pack's directory name (`tinct.yazi`) **must** end in `.yazi` — that's part of Yazi's flavor pack naming convention. The directory name without the `.yazi` suffix is the value you put in `theme.toml`.

The flavor name defaults to `tinct`. With `--yazi.theme-name=mything` the output dir becomes `~/.config/yazi/flavors/mything.yazi/` and you reference it as `"mything"`.

## Integration

In `~/.config/yazi/theme.toml`, activate the flavor:

```toml
[flavor]
dark = "tinct"
# light = "tinct"   # optional: use the same flavor in light mode
```

Yazi resolves `"tinct"` to `~/.config/yazi/flavors/tinct.yazi/`. Restart any running Yazi instances to pick up the change.

## Reload behaviour

### Automatic

None — `Hooks().Reload` is nil. Yazi only reads its theme on startup; it does not watch the flavors directory for changes.

### Manual fallback

Quit and relaunch Yazi (`q`, then `yazi` again). Running instances will continue using the previously loaded flavor until restarted.

## Uninstall / revert

1. **Remove the flavor reference** from `~/.config/yazi/theme.toml`: delete the `dark = "tinct"` (and/or `light = "tinct"`) line, or change it to another flavor name (e.g. `dark = "default-dark"`).

2. **Delete the generated flavor pack**:

   ```bash
   rm -rf ~/.config/yazi/flavors/tinct.yazi
   ```

   Note: the output **directory** is the flavor pack, not just `flavor.toml`. Use `rm -rf` against the whole `.yazi` directory.

3. **Reload to drop the theme**: quit and relaunch Yazi.

4. **External state**: this plugin only writes to `~/.config/yazi/flavors/<name>.yazi/`. No further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--yazi.output-dir` | `~/.config/yazi/flavors/tinct.yazi` | Override the output directory (must end in `.yazi`) |
| `--yazi.theme-name` | `tinct` | Flavor name; output directory becomes `<name>.yazi` and you reference it as `<name>` in `theme.toml` |

## Colour role mapping

| Yazi key | Tinct role | Notes |
|---|---|---|
| `[mgr] cwd.fg` | `foreground` | Current working directory text |
| `[mgr] find_keyword`, `find_position` | `warning` | Search highlights |
| `[mgr] marker_copied` | `success` | Yanked-to-copy marker |
| `[mgr] marker_cut` | `danger` | Yanked-to-cut marker |
| `[mgr] marker_marked` | `info` | Multi-selection marker |
| `[mgr] marker_selected` | `accent1` | Active selection marker |
| `[mgr] count_copied/cut/selected` | `onSuccess`/`success`, `onDanger`/`danger`, `onAccent1`/`accent1` | Selection counters |
| `[mgr] border_style` | `outline` | Pane borders |
| `[mgr] hovered`, `tab_active` | `onAccent1` / `accent1` | Hover and active tab |
| `[mgr] tab_inactive` | `foregroundMuted` / `surfaceContainer` | Inactive tabs |
| `[mgr.folder_offset]`, `[mgr.symlink_offset]` | `accent1`, `accent2` | Folder / symlink size column |
| `[status] separator_style` | `surfaceContainerHigh` / `surfaceContainer` | Powerline separators |
| `[status] mode_normal` | `background` / `accent1` | Normal mode badge |
| `[status] mode_select` | `background` / `warning` | Select mode badge |
| `[status] mode_unset` | `background` / `danger` | Unset mode badge |
| `[status] progress_normal` | `accent1` / `surfaceContainer` | Progress bar (in-flight) |
| `[status] progress_error` | `danger` / `surfaceContainer` | Progress bar (failed) |
| `[status] permissions_t/r/w/x/s` | `accent1`, `warning`, `danger`, `success`, `foregroundMuted` | File-permission column |
| `[select]`, `[input]`, `[confirm]`, `[pick]` borders/titles | `accent1` (+ `accent2` for pick active) | Prompt UI |
| `[which] mask` | `surfaceContainerLow` | Which-key overlay backdrop |
| `[which] cand`, `desc`, `rest`, `separator_style` | `accent2`, `accent1`, `foregroundMuted`, `outline` | Which-key text |
| `[help]` on/exec/desc/hovered/footer | `accent1`, `accent2`, `foregroundMuted`, `surfaceContainerHigh`, `background`/`foreground` | Help screen |
| `[tasks] border` | `accent1` | Task manager border |
| `[notify] title_info/warn/error` | `info`, `warning`, `danger` | Notification titles |
| `[filetype] rules` | accent2/3/4 + warning + accent1 + foreground | image/video/audio MIME, archives, dirs, fallback |

## Customising the template

Dump the default template to override it:

```bash
tinct plugins templates dump -o yazi -l ~/.config/tinct/templates/yazi
```

This writes `~/.config/tinct/templates/yazi/flavor.toml.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions.

## Troubleshooting

### Yazi shows the default theme

Two checks:

```bash
grep -A2 '^\[flavor\]' ~/.config/yazi/theme.toml
ls ~/.config/yazi/flavors/tinct.yazi/flavor.toml
```

If the theme.toml block is missing or points at a different name, fix it. If the flavor pack directory or `flavor.toml` is missing, re-run `tinct generate -o yazi`.

### Restarting Yazi doesn't change anything

Yazi reads `theme.toml` and the flavor pack once at startup. If you have Yazi running across tmux panes, restart each instance. Multiplexers like tmux do not detach Yazi's already-loaded theme.

### Wrong directory used

If you set `--yazi.output-dir` to a path that does **not** end in `.yazi`, Yazi will still load `flavor.toml` from it, but the activation key in `theme.toml` is the directory name minus `.yazi`. Keep the suffix; it's load-bearing.

### Custom flavor name not found by Yazi

When using `--yazi.theme-name=foo`, the generated dir is `foo.yazi` and your `theme.toml` must say `dark = "foo"`, not `"tinct"`.

## Related plugins

- [neovim](../editors/neovim.md) — editor theming for users who pair Yazi with Neovim as a file manager workflow.
- [ghostty](../terminals/ghostty.md), [kitty](../terminals/kitty.md), [wezterm](../terminals/wezterm.md) — terminal themes for the host terminal Yazi runs inside.
