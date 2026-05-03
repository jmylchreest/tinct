# tinct-plugin-ptyxis

External output plugin for [Tinct](https://github.com/jmylchreest/tinct)
that generates [Ptyxis](https://gitlab.gnome.org/chergert/ptyxis) terminal
colour palettes.

## What it does

Writes a `tinct.palette` file directly into Ptyxis's watched palette
directory. PtyxisUserPalettes detects the file via `GFileMonitor` and
the new palette appears under **Preferences → Appearance → Palette → Tinct**
without requiring an import step.

When tinct generates with both primary and alternate themes, the plugin
emits a dual-mode palette with `[Light]` and `[Dark]` sections. Ptyxis
switches between them based on the system colour preference. Otherwise
a single-section palette is written; Ptyxis derives its mode from the
background luminance.

## Install location detection

The plugin auto-detects whether Ptyxis is installed natively or via
Flatpak and writes to the corresponding palette directory:

- **Native**: `$XDG_DATA_HOME/org.gnome.Ptyxis/palettes` (default
  `~/.local/share/org.gnome.Ptyxis/palettes`).
- **Flatpak**: `~/.var/app/app.devsuite.Ptyxis/data/org.gnome.Ptyxis/palettes`.

Override with `--ptyxis.output-dir=/path` or
`TINCT_PLUGIN_PTYXIS_OUTPUT_DIR=/path`.

## Install

```sh
tinct plugin install ptyxis
```

## License

Apache 2.0 — see the parent project.
