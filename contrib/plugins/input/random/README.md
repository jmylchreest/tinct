---
title: random
sidebar_position: 9

plugin:
  type: input
  name: random
  source: external
  source_type: generated
  description: Generate random colour palettes with configurable seed and colour count
  version: 0.1.0
  protocol_version: 0.3.0
  repository: https://github.com/jmylchreest/tinct
  install: tinct plugins install random
  requires: []
  optional: []
  requires_network: false
  requires_credentials: []
  produces_wallpaper: false
---

# random

Generates a colour palette algorithmically — no image, no file, no network. Each colour is drawn from a deterministic PRNG (ChaCha8) seeded either explicitly via `--random.seed` or, by default, from `crypto/rand`. The same seed always produces the same palette, which makes this plugin useful for reproducible experimentation, stress-testing the categoriser with hostile inputs, and bootstrapping a theme when you have no image or palette in mind.

## Installation

### Via the official plugin repository

```bash
tinct plugins install random
```

### Build from source

```bash
cd contrib/plugins/input/random
go build -ldflags "-X main.Version=0.1.0" -o tinct-plugin-random
install -m 0755 tinct-plugin-random ~/.local/bin/
```

### Verify

```bash
which tinct-plugin-random
tinct-plugin-random --plugin-info | jq .
```

The plugin uses tinct's go-plugin RPC protocol and is discovered automatically once it's on `$PATH`.

## Quick start

```bash
# Default — 32 random colours from a fresh crypto/rand seed
tinct generate -i random -o kitty

# Reproducible — same seed, same palette every run
tinct generate -i random -o kitty --random.seed 12345

# Smaller palette
tinct generate -i random -o kitty --random.count 16

# Inspect a palette without writing any theme files
tinct extract -i random --random.seed 12345 --random.count 8
```

## Configuration / credentials

No credentials, no network access, no files. The only inputs are the two
plugin-args (`count` and `seed`), and both are per-invocation — external plugin
arguments are not read from `tinct.toml`, so there is nothing to persist. Wrap
the call in a shell alias if you want a fixed seed:

```bash
alias tinct-random='tinct generate -i random --random.seed 12345 -o'
```

If `seed` is omitted, the plugin reads 8 bytes from `crypto/rand` and uses them as the ChaCha8 seed — so every run produces a different palette. Pass `seed` to lock generation to a known result. The flag is registered as a signed 64-bit integer, so use the `--plugin-args` JSON form if you need a value above 2^63-1.

## Flags

Once the plugin is installed, tinct registers its arguments as regular
`--random.<arg>` flags. They can also be passed as JSON via
`--plugin-args random='{"seed":12345}'`.

| Flag | Type | Default | Description |
|----------|------|---------|-------------|
| `--random.count` | integer | `32` | Number of colours to generate (0–4096) |
| `--random.seed` | integer | _(crypto/rand)_ | ChaCha8 seed for reproducible generation |

## Output

A palette of `count` colours (default 32), each drawn uniformly at random from the full RGB cube. No role hints are emitted — tinct's categoriser is responsible for picking `background`, `foreground`, accents, etc., from the generated set. No theme hint is emitted either, so light/dark is inferred from the average luminance of the result.

This plugin **does not provide a wallpaper**.

## Troubleshooting

### `plugin-arg 'count' out of range`

`count` is capped at 4096 to prevent accidental misuse (a million colours doesn't yield a more interesting palette than a hundred). Pass a value in `0..4096`.

### `plugin-arg 'count' must be a number`

Plugin-args arrive as JSON, where numbers decode as `float64`. The `--random.count` flag always sends a number, so this only bites with the raw JSON form: write `--plugin-args random='{"count":32}'`, not `'{"count":"32"}'` — quoting turns it into a string and the validator rejects it.

### Same seed produces a different palette than last time

A change to the colour-generation function (e.g. tinct upgraded the PRNG, or this plugin's version changed) will produce different output for the same seed. Pin both the tinct version and the `tinct-plugin-random` version if exact reproducibility across machines matters.

### Palette looks washed-out / cluttered

Uniform random sampling of RGB doesn't produce aesthetically grouped palettes — that's the point. If you wanted a palette with semantic structure, use `image`, `paletty`, or one of the AI-generation inputs instead.

## Related plugins

- [`image`](../../../../internal/plugin/input/image/README.md) — extract a coherent palette from a real image.
- [`paletty`](../paletty/README.md) — fetch a curated palette from paletty.dev.
- [`google-genai`](../../../../internal/plugin/input/googlegenai/README.md), [`openrouter`](../../../../internal/plugin/input/openrouter/README.md) — generate a source image with an AI model.
