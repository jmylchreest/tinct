---
title: awob
sidebar_position: 7
plugin:
  type: output
  name: awob
  category: bars-launchers
  source: external
  app: awob
  app_url: 'https://github.com/jmylchreest/awob'
  version: 0.1.0
  protocol_version: 0.3.0
  repository: 'https://github.com/jmylchreest/tinct'
  install: tinct plugins install awob
  requires: []
  optional:
    - awob
    - awob-daemon
  pattern: two-file
  default_output_dir: ~/.config/awob/themes
  generated_files:
    - _palettes/tinct.kdl
    - tinct/scene.kdl
    - tinct/manifest.toml
  reload:
    method: ipc
    command: awob force-palette set <palette>
    user_action_required: false
---

# awob

Generates a shared palette plus a worked-example theme for [awob](https://github.com/jmylchreest/awob), an animated Wayland Overlay Bar. The palette lives at `themes/_palettes/tinct.kdl` and is the single source of truth for tinct colours; the worked-example theme at `themes/tinct/` imports it and exercises every awob scene primitive (icon, label, track, bar with `$lastValue` tweening, fade phases). Other awob themes (user-authored or bundled) can pick up the same palette by adding one `import` line to their `scene.kdl`.

## Installation

### Via the official plugin repository

```bash
tinct plugins install awob
```

### Build from source

```bash
cd contrib/plugins/output/awob
go build -ldflags "-X main.Version=0.1.0" -o tinct-plugin-awob
install -m 0755 tinct-plugin-awob ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-awob
tinct-plugin-awob --plugin-info | jq .
```

The plugin uses tinct's go-plugin RPC protocol and is discovered automatically once it's on `$PATH`. It skips cleanly if awob is not installed (no `awob`/`awob-daemon` binary on `$PATH` and no `~/.config/awob` directory).

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o awob
```

## Generated files

| File | Path | Role |
|------|------|------|
| `tinct.kdl` | `~/.config/awob/themes/_palettes/tinct.kdl` | **Palette** — defines `$bg`, `$fg`, `$track`, `$low`, `$normal`, `$warn`, `$crit`, `$muted`, `$overflow_bg`, `$overflow_accent` plus a `styles {}` block keying off severity. Tinct overwrites this in place on each run. |
| `scene.kdl` | `~/.config/awob/themes/tinct/scene.kdl` | **Example theme** — imports the palette and uses every awob scene primitive (icon, label, track, bar, fade phases). Modelled on awob's bundled `default` theme. |
| `manifest.toml` | `~/.config/awob/themes/tinct/manifest.toml` | Theme manifest awob-daemon reads to enumerate icons, fonts, etc. |

## Integration

awob themes adopt tinct colours by importing the palette. Two paths are available:

### Activate the bundled tinct theme

```bash
awob-daemon --theme tinct
# or set the active theme in awob's config:
$EDITOR ~/.config/awob/awob.toml
```

```toml
theme = "tinct"
```

### Use tinct's palette in another theme

Drop a single `import` line at the top of any theme's `scene.kdl`:

```kdl
import "../_palettes/tinct.kdl"
```

After this, the theme can use `$bg`, `$fg`, `$track`, `$accent`, `$low`, `$normal`, `$warn`, `$crit`, `$muted`, `$overflow_bg`, `$overflow_accent` and the `style "low"|"normal"|"warn"|"critical"|"muted"|"overflow"` accent blocks. The palette file is rewritten by tinct on every `tinct generate`; awob-daemon hot-reloads.

## Reload behaviour

### Automatic

If the `awob` CLI is on `$PATH`, `PostExecute` runs:

```bash
awob force-palette set ~/.config/awob/themes/_palettes/tinct.kdl
```

awob-daemon overlays this palette on top of whatever theme is currently active, so tinct colours take effect immediately on every theme without editing its scene.kdl. The daemon also adds the file to its hot-reload watch list, so subsequent tinct runs that rewrite the same path refresh in place.

The call has a 5-second timeout and silently no-ops if the daemon isn't running or the CLI is missing — failure is non-fatal.

### Manual fallback

If the `awob` CLI is missing or the daemon was not running, either:

```bash
awob force-palette set ~/.config/awob/themes/_palettes/tinct.kdl
```

or restart awob-daemon to pick up the regenerated theme:

```bash
systemctl --user restart awob-daemon   # if you run it as a user service
# otherwise: kill awob-daemon and re-launch it via your usual mechanism
```

## Uninstall / revert

1. **Remove the include line** from any theme's `scene.kdl` that you wired tinct into:

       grep -rl 'import "../_palettes/tinct.kdl"' ~/.config/awob/themes/

   Edit each file and drop the import line. If you set `theme = "tinct"` in `~/.config/awob/awob.toml`, switch it back to your preferred theme (or remove the line to fall back to awob's default).

2. **Drop the force-palette overlay** so the daemon stops applying tinct colours:

       awob force-palette clear

3. **Delete the generated files**:

       rm -rf ~/.config/awob/themes/tinct ~/.config/awob/themes/_palettes/tinct.kdl
       rmdir --ignore-fail-on-non-empty ~/.config/awob/themes/_palettes

4. **Reload to drop the theme**: re-run `awob force-palette clear` (no-op if already cleared) and restart the daemon if you removed the active theme it was using:

       systemctl --user restart awob-daemon

5. **External state**: this plugin only writes to `~/.config/awob/themes/`. To also remove the plugin binary:

       tinct plugins delete awob
       # or, for a source build:
       rm ~/.local/bin/tinct-plugin-awob

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--awob.output-dir` | auto-detected (`~/.config/awob/themes`) | awob themes directory. May also be set via `TINCT_PLUGIN_AWOB_OUTPUT_DIR`. |

## Colour role mapping

awob's palette uses short single-letter-or-word identifiers that are referenced as `$bg`, `$fg`, etc. from scene primitives.

### Palette (`palette {}`)

| awob identifier | Tinct role | Notes |
|---|---|---|
| `bg` | `background` | Surface, baked at 0.85 alpha |
| `fg` | `foreground` | Label and icon colour |
| `track` | `onSurface` | Bar track, baked at 0.08 alpha |
| `low` | `success` | Accent for low-severity events |
| `normal` | `accent1` | Accent for normal events |
| `warn` | `warning` | Accent for warning events |
| `crit` | `danger` | Accent for critical events |
| `muted` | `foregroundMuted` | Reserved for muted/disabled UI |
| `overflow_bg` | `background` (at 0.85 alpha) | Bar background when `value > max` |
| `overflow_accent` | `danger` | Bar/border accent when `value > max` |

### Style blocks (`styles {}`)

| awob style | Maps to |
|---|---|
| `low` | `accent = $low` |
| `normal` | `accent = $normal` |
| `warn` | `accent = $warn` |
| `critical` | `accent = $crit` |
| `muted` | `accent = $crit` at 0.6 alpha |
| `overflow` | `bg = $overflow_bg`, `accent = $overflow_accent` (auto-applied when `value > max`) |

## Customising the template

```bash
tinct plugins templates dump -o awob -l ~/.config/tinct/templates/awob
$EDITOR ~/.config/tinct/templates/awob/palette.kdl.tmpl
```

Tinct prefers your version over the embedded default. See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions. Three templates are dumped: `palette.kdl.tmpl`, `scene.kdl.tmpl`, `manifest.toml.tmpl`.

## Troubleshooting

### Plugin not found by tinct

```bash
which tinct-plugin-awob
tinct plugins list | grep awob
```

If the binary isn't on `$PATH`, reinstall via `tinct plugins install awob` or symlink your source build into `~/.local/bin/`.

### "awob not installed" — plugin skipped

The plugin requires evidence that you actually run awob: either `awob` or `awob-daemon` on `$PATH`, or a `~/.config/awob` directory. Create the config directory to convince the plugin to write the theme even if you haven't installed the daemon yet:

```bash
mkdir -p ~/.config/awob
tinct generate -o awob --verbose
```

### Colours don't update after regenerating

Confirm awob-daemon picked up the force-palette overlay:

```bash
awob force-palette get   # should print the path tinct just wrote
```

If empty, the `awob` CLI was not on `$PATH` at generate time. Install awob and either re-run `tinct generate` or apply manually:

```bash
awob force-palette set ~/.config/awob/themes/_palettes/tinct.kdl
```

### Theme not appearing in awob-daemon's theme list

awob-daemon enumerates themes from the directories listed in its config. Confirm the themes dir tinct wrote to matches:

```bash
ls ~/.config/awob/themes/
awob-daemon --print-config | grep -i themes
```

If you ran tinct with `--awob.output-dir` pointing somewhere else, either move the files or update awob-daemon's config to scan the new location.

## Related plugins

- [swayosd](./swayosd.md) — alternative on-screen-display server with a different theming model.
- [dunst](./dunst.md) — notification daemon companion.
