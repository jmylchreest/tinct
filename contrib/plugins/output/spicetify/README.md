---
title: spicetify
sidebar_position: 6

plugin:
  type: output
  name: spicetify
  category: special
  source: external
  app: Spicetify
  app_url: https://spicetify.app
  version: 0.1.0
  protocol_version: 0.3.0
  repository: https://github.com/jmylchreest/tinct
  install: tinct plugins install spicetify
  requires: [spicetify]
  optional: []
  pattern: single-file
  default_output_dir: ~/.config/spicetify/Themes/tinct
  generated_files: [color.ini]
  reload:
    method: none
    user_action_required: true
---

# spicetify

Generates a [Spicetify](https://spicetify.app) `color.ini` colour scheme so the Spotify desktop client picks up your tinct palette. Spicetify is a community theming tool that patches Spotify's installed application files; the tinct plugin only writes the colour scheme — applying it to Spotify is left to the user, because `spicetify apply` is invasive (it modifies Spotify's app directory, and requires the Spotify client to be closed).

## Installation

> **Heads up.** This plugin requires the [`spicetify`](https://spicetify.app) CLI to be installed. Spicetify patches Spotify's installed app files, can break on Spotify auto-updates, and is not endorsed by Spotify. Read the upstream documentation before using it.

### Install spicetify itself

Follow the [official installation guide](https://spicetify.app/docs/getting-started). On most Linux/macOS systems:

```bash
curl -fsSL https://raw.githubusercontent.com/spicetify/cli/main/install.sh | sh
```

Run `spicetify` once after installing to let it set up its config directory and (on Linux) request the file-permission grant on Spotify's app directory:

```bash
spicetify
```

### Install the tinct plugin

Via the official plugin repository:

```bash
tinct plugins install spicetify
```

Or build from source:

```bash
cd contrib/plugins/output/spicetify
go build -ldflags "-X main.Version=0.1.0" -o tinct-plugin-spicetify
install -m 0755 tinct-plugin-spicetify ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-spicetify
tinct-plugin-spicetify --plugin-info | jq .
which spicetify
```

The plugin uses tinct's go-plugin RPC protocol and is discovered automatically once it's on `$PATH`. If `spicetify` itself is missing, the plugin will skip cleanly with a pointer to the install URL.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o spicetify
```

This writes `~/.config/spicetify/Themes/tinct/color.ini`. It does **not** touch Spotify — see [Integration](#integration) for the manual apply step.

## Generated files

| File | Path | Role |
|------|------|------|
| `color.ini` | `~/.config/spicetify/Themes/tinct/color.ini` | Spicetify colour scheme (the `[base]` section that maps tinct's palette onto Spicetify's UI tokens) |

The directory `~/.config/spicetify/Themes/tinct/` is treated by Spicetify as a *theme called `tinct`*, with a single colour scheme also called `tinct`. The plugin creates the directory if it doesn't exist.

## Integration

**This plugin never invokes `spicetify apply` automatically.** Spicetify's apply step mutates Spotify's installed app directory in-place; doing that behind the user's back would:

- run against a live Spotify process and leave the client half-patched,
- surface permission errors silently the first time it touches Spotify's directory,
- repeatedly re-apply on every `tinct generate`, which is invasive and slow.

After tinct writes `color.ini`, **close Spotify completely**, then run:

```bash
spicetify config current_theme tinct color_scheme tinct
spicetify apply
```

`spicetify apply` will start Spotify with the new colour scheme. From then on, you can re-run `tinct generate -o spicetify` whenever the palette changes — but the new colours will not appear until you close Spotify and run `spicetify apply` again.

## Reload behaviour

### Automatic

None. `method: none`, `user_action_required: true`. The plugin deliberately does not call `spicetify apply` for the reasons listed in [Integration](#integration).

When tinct finishes writing the file, the plugin prints a short reminder of the apply commands on stderr.

### Manual fallback

Close Spotify, then:

```bash
spicetify config current_theme tinct color_scheme tinct
spicetify apply
```

If Spotify is already configured to use the `tinct` theme, you only need `spicetify apply` on subsequent regenerations.

## Uninstall / revert

1. **Restore Spotify to factory state.** This undoes every patch Spicetify has applied to the Spotify install:

   ```bash
   spicetify restore
   ```

2. **Delete the generated files**:

   ```bash
   rm -rf ~/.config/spicetify/Themes/tinct
   ```

3. **Switch Spicetify's configured theme back** to something else (or to its default) so a future `spicetify apply` doesn't try to reload the deleted theme:

   ```bash
   spicetify config current_theme '' color_scheme ''
   ```

4. **External state**: this plugin only writes files under `~/.config/spicetify/Themes/tinct/`. However, **the `spicetify` tool itself mutates Spotify's installed app directory** when you ran `spicetify apply`. Step 1 above (`spicetify restore`) is what undoes that — running `rm` on the theme directory alone is **not** enough.

5. **Remove the tinct plugin binary**:

   ```bash
   tinct plugins delete spicetify
   # or, for a source build:
   rm ~/.local/bin/tinct-plugin-spicetify
   ```

> **Auto-update warning.** Spotify auto-updates itself silently. Each Spotify update **reverts Spicetify's patches** and can break Spicetify itself if the new Spotify version isn't yet supported upstream. Symptoms: the theme stops working, Spotify looks completely default, or `spicetify apply` fails. Fix: update spicetify (`spicetify update` or via your package manager), then re-run `spicetify apply`.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--spicetify.output-dir` | `~/.config/spicetify/Themes/tinct` | Output directory for the Spicetify theme. Must be a directory directly under `~/.config/spicetify/Themes/`; the directory name becomes the theme name in `spicetify config current_theme`. |

## Colour role mapping

Spicetify's `color.ini` only requires a `[base]` section. Tinct generates the following keys:

| Spicetify key | Tinct role | Used by Spotify for |
|---|---|---|
| `text` | `foreground` | primary text colour |
| `subtext` | `foregroundMuted` | secondary / descriptive text |
| `sidebar-text` | `foreground` | navigation sidebar text |
| `main` | `background` | main view background |
| `sidebar` | `backgroundMuted` | navigation sidebar background |
| `player` | `backgroundMuted` | now-playing bar background |
| `card` | `surface` | album / playlist card backgrounds |
| `shadow` | `backgroundMuted` | drop shadows under floating panels |
| `selected-row` | `surfaceContainerHigh` (falls back to `surface`) | highlighted row in lists / queue |
| `button` | `accent1` | primary action buttons (play, follow) |
| `button-active` | `accent2` | pressed / active button state |
| `button-disabled` | `foregroundMuted` | disabled button state |
| `tab-active` | `surfaceContainer` (falls back to `surface`) | active tab background |
| `notification` | `info` | informational toasts |
| `notification-error` | `danger` | error toasts |
| `misc` | `foregroundMuted` | miscellaneous UI accents |

Values are emitted in `RRGGBB` (no leading `#`) per Spicetify's format. `selected-row` and `tab-active` fall back to `surface` if the `surfaceContainer*` roles aren't present in the palette (some palette sources don't emit the full Material-style container hierarchy).

## Customising the template

```bash
tinct plugins templates dump -o spicetify -l ~/.config/tinct/templates/spicetify
$EDITOR ~/.config/tinct/templates/spicetify/color.ini.tmpl
```

Tinct prefers your override over the embedded default. See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for available functions — the spicetify template uses `get`, `has`, `hexNoHash`, and `themeType`.

## Troubleshooting

### "spicetify CLI not installed; install from https://spicetify.app"

The plugin skips at `PreExecute` if `spicetify` is not on `$PATH`. Install it from [spicetify.app/docs/getting-started](https://spicetify.app/docs/getting-started), then re-run `tinct generate`.

### Spotify still looks default after `spicetify apply`

Spotify was probably still running when you ran apply, or auto-updated since the last apply. Fully close Spotify (check your tray / dock; on Linux `pgrep -f spotify` should return nothing), then:

```bash
spicetify apply
```

If that doesn't help, check that `current_theme` and `color_scheme` are set to `tinct`:

```bash
spicetify config | grep -E 'current_theme|color_scheme'
```

### "Permission denied" applying the theme

Spotify's installed app directory is owned by the package manager. Spicetify needs write access. Run `spicetify` once on its own:

```bash
spicetify
```

It will print the exact `chmod` / `chown` it needs and (on many distros) offer to apply it for you. See the [Spicetify Linux setup notes](https://spicetify.app/docs/advanced-usage/installation#spotify-installed-from-flatpak) for Flatpak / Snap caveats.

### Spotify auto-updated and Spicetify broke

This happens every few weeks. Update Spicetify (`spicetify update`, or via your package manager), then re-apply:

```bash
spicetify update
spicetify backup apply
```

If `spicetify` itself doesn't yet support the new Spotify build, watch the [upstream releases](https://github.com/spicetify/cli/releases). There's nothing tinct can do until upstream catches up.

### Theme directory created but Spotify can't find it

Spicetify expects the theme to live in a directory directly under `~/.config/spicetify/Themes/`, and the directory name is the theme name. Don't move `color.ini` elsewhere; if you used `--spicetify.output-dir`, make sure the parent is `~/.config/spicetify/Themes/`.

## Related plugins

- [opencode](../opencode/README.md) — sister external plugin that writes a JSON theme for the OpenCode editor.
