---
title: kde-plasma
sidebar_position: 7

plugin:
  type: output
  name: kde-plasma
  category: desktop
  source: builtin
  app: KDE Plasma
  app_url: https://kde.org/plasma-desktop/
  requires: []
  optional: []
  pattern: flavor-pack
  default_output_dir: ~/.local/share/color-schemes
  generated_files: [TinctDark1.colors, TinctDark2.colors, TinctLight1.colors, TinctLight2.colors]
  reload:
    method: ipc
    command: plasma-apply-colorscheme TinctDark<N> (alternates 1↔2) plus KDE D-Bus reload
    user_action_required: false
---

# kde-plasma

Generates [KDE Plasma](https://kde.org/plasma-desktop/) `.colors` colour schemes for the Plasma desktop and KDE applications. The file format is the standard KDE `KConfig`-style ini consumed by `plasma-apply-colorscheme`, KWin, Dolphin, Kate, and every other KDE app. The plugin uses a two-variant trick (`TinctDark1` / `TinctDark2`, and the same for `Light`) to work around a long-standing Plasma quirk: `plasma-apply-colorscheme` silently no-ops if you re-apply a scheme that is already active. By emitting two near-identical schemes per theme type and alternating between them on every run, tinct guarantees a visible reload every time.

## Installation

Built into tinct — nothing to install separately. `tinct generate -o kde-plasma` works out of the box. KDE Plasma must already be installed (the plugin checks for `~/.config/kdeglobals` or `~/.config/plasmarc` and aborts otherwise).

For non-Plasma desktops running KDE / Qt apps, use the [`qt5`](../qt5/README.md) or [`qt6`](../qt6/README.md) plugins instead.

## Quick start

```bash
tinct generate -i image -p ~/Pictures/wallpaper.jpg -o kde-plasma
```

## Generated files

In single-variant mode (one theme type only):

| File | Path | Role |
|------|------|------|
| `TinctDark1.colors` / `TinctLight1.colors` | `~/.local/share/color-schemes/TinctDark1.colors` etc. | Variant 1 of the active theme type |
| `TinctDark2.colors` / `TinctLight2.colors` | `~/.local/share/color-schemes/TinctDark2.colors` etc. | Variant 2 of the active theme type (identical palette, different `Name=`/`ColorScheme=` keys) |

In dual-theme mode (`GenerateDualTheme`, when both a primary and an alternate theme are configured), all four files are written: `TinctDark1`, `TinctDark2`, `TinctLight1`, `TinctLight2`.

The two-variant trick: `plasma-apply-colorscheme TinctDark1` does nothing if `TinctDark1` is already the active scheme. By switching to `TinctDark2` on the next run (and back again on the run after that), Plasma sees a "different" scheme each time and reloads.

## Integration

KDE Plasma auto-discovers `.colors` files from two locations:

- `~/.local/share/color-schemes/` (per-user) — where this plugin writes.
- `/usr/share/color-schemes/` (system-wide) — distro packages.

No manual configuration is required. After the files are written, the plugin invokes `plasma-apply-colorscheme` to make Plasma activate one of them and writes the active scheme name into `~/.config/kdeglobals` under `[General]` → `ColorScheme`.

In dual-theme mode the plugin reads `~/.config/kdeglobals` `[General] ColorScheme` (and falls back to `[KDE] LookAndFeelPackage`) to decide whether the user is currently on a dark or light scheme, and applies the matching variant.

## Reload behaviour

### Automatic

After writing the files, `PostExecute` performs three steps:

1. **Pick the variant to apply**. Read `kdeglobals` `[General] ColorScheme` via `kreadconfig5`. If it's `TinctDark1`, switch to `TinctDark2` (or vice-versa); otherwise, start at variant 1. This is the workaround for `plasma-apply-colorscheme` not reloading a scheme that's already active.
2. **Apply via `plasma-apply-colorscheme <ThemeName>`**. This rewrites `kdeglobals` and pokes Plasma's configuration daemon. If `plasma-apply-colorscheme` is missing from `$PATH`, the step is skipped with a warning telling the user how to apply manually.
3. **Trigger a D-Bus reload** of KWin and Plasma Shell via the shared KDE D-Bus helper. This ensures the running shell and window manager pick up the new palette immediately rather than on next start.

If a wallpaper was supplied via `-p`, the plugin also runs `plasma-apply-wallpaperimage <path>`.

### Manual fallback

If `plasma-apply-colorscheme` is missing or the D-Bus reload fails:

```bash
plasma-apply-colorscheme TinctDark1   # or TinctLight1
```

If neither variant is currently active and you want a guaranteed reload, run the command twice with different variant numbers:

```bash
plasma-apply-colorscheme TinctDark2
plasma-apply-colorscheme TinctDark1
```

For a full-belt-and-braces refresh you can also restart Plasma:

```bash
kquitapp5 plasmashell && kstart5 plasmashell    # Plasma 5
kquitapp6 plasmashell && kstart6 plasmashell    # Plasma 6
```

## Uninstall / revert

1. **Switch to a stock Plasma colour scheme** before deleting the files (otherwise Plasma keeps the cached palette in `kdeglobals` until you next change it):

   ```bash
   plasma-apply-colorscheme BreezeDark   # or Breeze for light, or any other installed scheme
   ```

2. **Delete the generated files**:

   ```bash
   rm ~/.local/share/color-schemes/TinctDark1.colors \
      ~/.local/share/color-schemes/TinctDark2.colors \
      ~/.local/share/color-schemes/TinctLight1.colors \
      ~/.local/share/color-schemes/TinctLight2.colors
   ```

3. **Reload** — `plasma-apply-colorscheme` in step 1 already did the reload. If needed, restart Plasma as shown in the [Manual fallback](#manual-fallback) above.

4. **External state**: this plugin **does mutate state outside `~/.config/tinct`**. Specifically:
   - `~/.config/kdeglobals` is rewritten by `plasma-apply-colorscheme` with `[General] ColorScheme=TinctDark<N>` (and tens of palette entries under `[Colors:*]` sections). Re-applying a stock scheme in step 1 overwrites these.
   - If `-p <wallpaper>` was passed, the active Plasma containment's wallpaper config in `~/.config/plasma-org.kde.plasma.desktop-appletsrc` was updated by `plasma-apply-wallpaperimage`. Pick a new wallpaper in System Settings → Wallpaper to clear it.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--kde-plasma.output-dir` | `~/.local/share/color-schemes` | Override the output directory |

## Colour role mapping

KDE colour schemes have many sections; the table below lists the recurring keys and their tinct roles. Within each section the same row applies unless noted.

| Section(s) | Key | Tinct role |
|---|---|---|
| `Colors:Button`, `Colors:Header`, `Colors:Tooltip`, `Colors:Window` | `BackgroundNormal` | `surface` |
| `Colors:Button`, `Colors:Tooltip`, `Colors:Window` | `BackgroundAlternate` | `surfaceVariant` |
| `Colors:Header`, `Colors:Complementary` | `BackgroundAlternate` | `surface` / `background` |
| `Colors:View` | `BackgroundNormal` | `background` |
| `Colors:View` | `BackgroundAlternate` | `surface` |
| `Colors:Selection` | `BackgroundNormal` / `BackgroundAlternate` | `accent1` |
| (all `Colors:*`) | `DecorationFocus` / `DecorationHover` | `accent1` (`Colors:Selection`.`DecorationHover` = `accent2`; `Colors:View`.`DecorationHover` = `backgroundMuted`/`surfaceVariant`) |
| (all `Colors:*`) | `ForegroundNormal` | `foreground` (`Colors:Selection` = `onAccent1`) |
| (all `Colors:*`) | `ForegroundActive` | `foreground` (`Colors:Window` = `accent1`; `Colors:Selection` = `onAccent1`) |
| (all `Colors:*`) | `ForegroundInactive` | `outline` / `border` / `foregroundMuted` (whichever the palette defines first) |
| (all `Colors:*`) | `ForegroundLink` | `accent1` |
| (all `Colors:*`) | `ForegroundVisited` | `accent2` |
| (all `Colors:*`) | `ForegroundNegative` | `danger` (falls back to ANSI red) |
| (all `Colors:*`) | `ForegroundNeutral` | `warning` (falls back to ANSI yellow) |
| (all `Colors:*`) | `ForegroundPositive` | `success` (falls back to ANSI green) |
| `ColorEffects:Disabled`, `ColorEffects:Inactive` | `Color` | `surface` |
| `WM` | `activeBackground` | `surface` |
| `WM` | `inactiveBackground` | `surfaceVariant` |
| `WM` | `activeForeground` / `activeBlend` | `foreground` |
| `WM` | `inactiveForeground` / `inactiveBlend` | `outline`/`border` |
| `General` | `ColorScheme` / `Name` | literal `TinctDark<N>` / `TinctLight<N>` (set by `Generate`, suffixed in `generateColorSchemeVariant`) |

## Customising the template

```bash
tinct plugins templates dump -o kde-plasma -l ~/.config/tinct/templates/kde-plasma
```

This creates `~/.config/tinct/templates/kde-plasma/tinct.colors.tmpl`. Tinct uses your version in preference to the embedded default.

Note that the variant-suffix post-processing (`ColorScheme=TinctDark` → `TinctDark1`/`TinctDark2`) happens in Go code, not in the template. Your custom template only needs to emit the un-suffixed name; the plugin will append the variant digit.

See the [templating reference](https://jmylchreest.github.io/tinct/docs/templating) for the available functions and palette accessors.

## Troubleshooting

### `Plugin skipped: KDE Plasma does not appear to be installed`

`PreExecute` couldn't find either `~/.config/kdeglobals` or `~/.config/plasmarc`. Either Plasma genuinely isn't installed, or the user hasn't logged into a Plasma session yet (those files are created on first login). Log into Plasma once, log out, and retry.

### `plasma-apply-colorscheme not found, theme not auto-applied`

The `plasma-workspace` package didn't install `plasma-apply-colorscheme` (uncommon, but seen on stripped-down Plasma builds). The `.colors` files were still written successfully; apply them manually with `plasma-apply-colorscheme TinctDark1` or by opening **System Settings → Colors** and selecting "Tinct Dark 1".

### Theme changes don't appear until I open System Settings

This is the bug the two-variant trick exists to work around. If you're still seeing it (e.g. the variant alternation didn't fire because your `kdeglobals` lists a non-tinct scheme), force the issue by running:

```bash
plasma-apply-colorscheme TinctDark2
plasma-apply-colorscheme TinctDark1
```

### KWin titlebars or Konsole colour scheme didn't update

KWin reads decoration colours from `kdeglobals` `[WM]` on D-Bus reload. The plugin's D-Bus reload step should handle this; if it didn't fire (D-Bus unavailable, or `dbus.IsAvailable()` returned false in CI / nested compositors), log out and back into your session to pick up the change.

### Dual-theme mode applied the wrong variant

`detectSystemColorScheme` reads `kdeglobals` `[General] ColorScheme` and looks for the substring `dark` or `light`. Custom-named user schemes that don't contain those words confuse the detection — it defaults to `TinctDark`. Either rename your previous scheme to include `Dark`/`Light`, or override with `--kde-plasma.theme-type` (if exposed) or by manually invoking `plasma-apply-colorscheme TinctLight1`.

## Related plugins

- [qt5](../qt5/README.md), [qt6](../qt6/README.md) — Qt application theming for non-Plasma desktops, or for Qt apps that don't follow the Plasma colour scheme.
- [gtk3](../gtk3/README.md), [gtk4](../gtk4/README.md), [libadwaita](../libadwaita/README.md) — theme GTK / libadwaita apps running under Plasma so they match the Plasma colour scheme.
