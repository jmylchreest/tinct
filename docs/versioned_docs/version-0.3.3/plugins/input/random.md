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
  repository: 'https://github.com/jmylchreest/tinct'
  install: tinct plugins install random
  requires: []
  optional: []
  requires_network: false
  requires_credentials: []
  produces_wallpaper: false
---

# random

Generates a colour palette algorithmically — no image, no file, no network. Each colour is drawn from a deterministic PRNG (ChaCha8) seeded either explicitly via `--plugin-arg seed=…` or, by default, from `crypto/rand`. The same seed always produces the same palette, which makes this plugin useful for reproducible experimentation, stress-testing the categoriser with hostile inputs, and bootstrapping a theme when you have no image or palette in mind.

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
tinct generate -i random -o tailwind

# Reproducible — same seed, same palette every run
tinct generate -i random -o tailwind --plugin-arg seed=12345

# Smaller palette
tinct generate -i random -o tailwind --plugin-arg count=16
```

## Configuration / credentials

No credentials, no network access, no files. The only inputs are the two plugin-args (`count` and `seed`). Both can be persisted via `tinct plugins config random` so subsequent `tinct generate -i random …` invocations don't need to repeat them.

If `seed` is omitted, the plugin reads 8 bytes from `crypto/rand` and uses them as the ChaCha8 seed — so every run produces a different palette. Pass `seed` (any unsigned 64-bit integer) to lock generation to a known result.

## Flags

This plugin uses `--plugin-arg key=value` rather than dedicated CLI flags:

| Argument | Type | Default | Description |
|----------|------|---------|-------------|
| `count` | integer | `32` | Number of colours to generate (0–4096) |
| `seed` | integer | _(crypto/rand)_ | ChaCha8 seed for reproducible generation |

## Output

A palette of `count` colours (default 32), each drawn uniformly at random from the full RGB cube. No role hints are emitted — tinct's categoriser is responsible for picking `background`, `foreground`, accents, etc., from the generated set. No theme hint is emitted either, so light/dark is inferred from the average luminance of the result.

This plugin **does not provide a wallpaper**.

## Troubleshooting

### `plugin-arg 'count' out of range`

`count` is capped at 4096 to prevent accidental misuse (a million colours doesn't yield a more interesting palette than a hundred). Pass a value in `0..4096`.

### `plugin-arg 'count' must be a number`

Plugin-args arrive as JSON, where numbers decode as `float64`. Pass the value bare (`count=32`) — quotes turn it into a string and the validator will reject it.

### Same seed produces a different palette than last time

A change to the colour-generation function (e.g. tinct upgraded the PRNG, or this plugin's version changed) will produce different output for the same seed. Pin both the tinct version and the `tinct-plugin-random` version if exact reproducibility across machines matters.

### Palette looks washed-out / cluttered

Uniform random sampling of RGB doesn't produce aesthetically grouped palettes — that's the point. If you wanted a palette with semantic structure, use `image`, `paletty`, or one of the AI-generation inputs instead.

## Related plugins

- [`image`](./image.md) — extract a coherent palette from a real image.
- [`paletty`](./paletty.md) — fetch a curated palette from paletty.dev.
- [`google-genai`](./google-genai.md), [`openrouter`](./openrouter.md) — generate a source image with an AI model.
