# tinct-plugin-awob

External output plugin for [Tinct](https://github.com/jmylchreest/tinct)
that generates a theme + palette for [awob](https://github.com/jmylchreest/awob),
the animated Wayland Overlay Bar.

## What it writes

Three files under `~/.config/awob/themes/`:

```
themes/
├── _palettes/
│   └── tinct.kdl       ← shared palette + named-style block
└── tinct/
    ├── manifest.toml   ← theme manifest awob-daemon reads
    └── scene.kdl       ← scene that imports the palette and uses
                          every awob primitive (icon, label, track,
                          bar with $lastValue tweening, fade phases).
```

## Use the tinct palette in your own theme

Drop a single line at the top of any theme's `scene.kdl`:

```kdl
import "../_palettes/tinct.kdl"
```

After that, `$bg`, `$fg`, `$track`, `$accent` (and `$low` / `$normal` /
`$warn` / `$crit` / `$muted`) resolve from the imported palette. Tinct
overwrites the palette in place on every `tinct generate`; awob-daemon
hot-reloads.

## Activate

```sh
awob-daemon --theme tinct
# or set 'theme = "tinct"' in ~/.config/awob/awob.toml
```

## Override the output dir

```sh
tinct generate -o awob --awob.output-dir=/path/to/themes
# or
TINCT_PLUGIN_AWOB_OUTPUT_DIR=/path/to/themes tinct generate -o awob
```

## Install

```sh
tinct plugin install awob
```

## License

Apache 2.0 — see the parent project.
