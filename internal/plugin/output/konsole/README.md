---
title: konsole
sidebar_position: 4

plugin:
  type: output
  name: konsole
  category: terminals
  source: builtin
  app: Konsole
  app_url: https://konsole.kde.org/
  requires: [konsole]
  optional: []
  pattern: single-file
  default_output_dir: ~/.local/share/konsole
  generated_files: [TinctDark.colorscheme, TinctLight.colorscheme]
  reload:
    method: none
    user_action_required: true
---

# konsole

Generates a `.colorscheme` file for [Konsole](https://konsole.kde.org/), the KDE terminal emulator. The file is dropped into Konsole's user data directory where it shows up in the colour-scheme picker (Settings → Edit Current Profile → Appearance) alongside the bundled schemes.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o konsole` works out of the box.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o konsole
```

## Generated files

The filename depends on the theme type detected from the source image / palette:

| File | Path | Role |
|------|------|------|
| `TinctDark.colorscheme` | `~/.local/share/konsole/TinctDark.colorscheme` | Single-file scheme written when the resolved palette is dark. |
| `TinctLight.colorscheme` | `~/.local/share/konsole/TinctLight.colorscheme` | Single-file scheme written when the resolved palette is light. |

Only one of the two is produced per run — whichever matches the resolved theme type.

## Generated format

Konsole uses an INI-style format with colour values as RGB triplets:

```ini
[Background]
Color=30,30,46

[BackgroundFaint]
Color=30,30,46

[BackgroundIntense]
Color=30,30,46

[Color0]
Color=69,71,90

[Color0Faint]
Color=69,71,90

[Color0Intense]
Color=88,91,112

# ... continues for colors 1-7

[Foreground]
Color=205,214,244

[ForegroundFaint]
Color=166,173,200

[ForegroundIntense]
Color=205,214,244

[General]
Description=Tinct Theme
Opacity=1
Wallpaper=
```

## Integration

Konsole auto-discovers `.colorscheme` files in `~/.local/share/konsole/` (plus the system-wide locations). To activate the scheme inside a profile:

1. Settings → Edit Current Profile → Appearance.
2. Pick **TinctDark** or **TinctLight** from the colour scheme list.
3. OK / Apply.

Or run from a shell *inside* the Konsole session you want themed:

```bash
konsoleprofile colors=TinctDark
```

## Reload behaviour

### Automatic

The plugin spec declares no signal- or IPC-style reload, so the README contract flags reload as `method: none`. In practice the plugin's `PostExecute` does try two things opportunistically, in order, and exits successfully if either succeeds:

1. **D-Bus** — enumerates every running Konsole session on the session bus and calls `org.kde.konsole.Session.setProfile(<schemeName>)` to swap the scheme live.
2. **`konsoleprofile colors=<schemeName>`** — only works when tinct is run from inside a Konsole session; sets the colour scheme on the current session only.

Neither path is guaranteed: D-Bus may not be available (non-Linux, no session bus), no Konsole windows may be open, or `konsoleprofile` may not be on PATH. The frontmatter therefore treats reload as user-driven.

### Manual fallback

If neither auto-reload path fired, change the scheme in **Edit Current Profile → Appearance**, or run inside the target window:

```bash
konsoleprofile colors=TinctDark   # or TinctLight
```

New Konsole windows opened with the same profile pick up the change automatically.

## Uninstall / revert

1. **Remove the config line**: none added. The scheme is selected through the Konsole profile UI; the plugin never edits any text config file. Reverting is a matter of picking a different scheme in **Edit Current Profile → Appearance**.

2. **Delete the generated files**:

   ```bash
   rm -f ~/.local/share/konsole/TinctDark.colorscheme \
         ~/.local/share/konsole/TinctLight.colorscheme
   ```

3. **Reload to drop the theme**:

   Switch the Konsole profile back to the scheme you used previously, or restart Konsole.

4. **External state**: this plugin only writes to `~/.local/share/konsole/`. The profile file (`~/.local/share/konsole/*.profile`) keeps a reference to the chosen scheme name — Konsole will fall back to its default scheme once the `.colorscheme` file is gone, so no further cleanup is required.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--konsole.output-dir` | `~/.local/share/konsole` | Override the output directory |

## Colour role mapping

Konsole's scheme file uses `[ColorN]` sections plus `Faint` / `Intense` variants for each of the 8 ANSI slots, then a separate `[Foreground]` / `[Background]` block.

| Konsole section | Tinct role |
|---|---|
| `[Background]`, `[BackgroundFaint]`, `[BackgroundIntense]`, `[Color0]`, `[Color0Faint]` | ANSI `black` |
| `[Color0Intense]` | ANSI `brightblack` |
| `[Color1]` / `[Color1Faint]` / `[Color1Intense]` | ANSI `red` / `red` / `brightred` |
| `[Color2]` / `[Color2Faint]` / `[Color2Intense]` | ANSI `green` / `green` / `brightgreen` |
| `[Color3]` / `[Color3Faint]` / `[Color3Intense]` | ANSI `yellow` / `yellow` / `brightyellow` |
| `[Color4]` / `[Color4Faint]` / `[Color4Intense]` | ANSI `blue` / `blue` / `brightblue` |
| `[Color5]` / `[Color5Faint]` / `[Color5Intense]` | ANSI `magenta` / `magenta` / `brightmagenta` |
| `[Color6]` / `[Color6Faint]` / `[Color6Intense]` | ANSI `cyan` / `cyan` / `brightcyan` |
| `[Color7]` / `[Color7Faint]`, `[Foreground]`, `[ForegroundFaint]` | ANSI `white` |
| `[Color7Intense]`, `[ForegroundIntense]` | ANSI `brightwhite` |

The `[General]` block writes `Description = Tinct Dark` or `Tinct Light` based on theme type, with `Blur=false`, `Opacity=1`, no wallpaper.

## Customising the template

Extract the default template to override it:

```bash
tinct plugins templates dump -o konsole -l ~/.config/tinct/templates/konsole
```

This writes `~/.config/tinct/templates/konsole/tinct.colorscheme.tmpl`. Tinct uses your version in preference to the embedded default.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### Scheme doesn't appear in the picker

Konsole reads the directory once at startup. Restart Konsole (or close and re-open the profile-edit dialog) after the first generation. Subsequent updates to the same `.colorscheme` filename are picked up live by D-Bus / `konsoleprofile`.

### D-Bus update failed

`tinct generate -v -o konsole` prints why the D-Bus call failed (no session bus, permission denied, no Konsole instances on the bus). Fall back to `konsoleprofile colors=TinctDark` inside the window you want to change, or pick the scheme from the profile UI.

### Wrong polarity (light scheme on a dark image / vice versa)

The plugin writes whichever file matches the resolved theme type. If you want both, generate twice with explicit `--mode dark` and `--mode light` (or whichever tinct mode flag your version uses), or override the palette so the resolved type matches your intent.

## Related plugins

- [kde-plasma](../kde-plasma/README.md) — KDE Plasma desktop colour scheme; pair with Konsole for a consistent KDE look.
- [qt5](../qt5/README.md), [qt6](../qt6/README.md) — Qt application palettes that align with Konsole's colours.
