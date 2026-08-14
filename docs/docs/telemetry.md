---
sidebar_position: 3
---

# Telemetry

Since v0.1.25, tinct collects anonymous usage telemetry to help prioritise plugin development and maintenance. Telemetry is **opt-out** (enabled by default) and can be easily disabled.

## What is collected

When you run `tinct generate`, two types of events are sent:

### Generate event (one per invocation)

| Dimension | Example | Purpose |
|-----------|---------|---------|
| `app.version` | `0.1.28` | Tinct version |
| `input.plugin` | `image` | Which input source was used |
| `input.ai` | `false` | Whether an AI input plugin was used |
| `output.plugins` | `["kitty","waybar"]` | Which output plugins were requested (array) |
| `generate.theme_type` | `dark` | Detected theme type |
| `generate.seed_mode` | `content` | Seed mode used |
| `generate.backend` | `kmeans` | Colour extraction backend |
| `generate.ambience` | `false` | Whether ambient extraction was enabled |
| `generate.dry_run` | `false` | Whether this was a dry run |
| `generate.dual_theme` | `true` | Whether both themes were generated |

### Plugin used event (one per output plugin)

| Dimension | Example | Purpose |
|-----------|---------|---------|
| `plugin.name` | `kitty` | Plugin name |
| `plugin.version` | `0.1.28` | Plugin version |
| `plugin.external` | `false` | Whether it is an external plugin |
| `plugin.status` | `ok` | Outcome: `ok`, `failed`, or `skipped` |

### Automatic metadata

The statsfactory SDK automatically includes the following in every request via the `User-Agent` header:

| Property | Example | Purpose |
|----------|---------|---------|
| Client name | `tinct` | Application identifier |
| Client version | `0.1.28` | Tinct version |
| OS | `linux` | Operating system |
| Architecture | `amd64` | CPU architecture |

Each event also includes a `distinct_id` set to the SHA256-hashed installation identifier, used for unique installation counting.

### What is NOT collected

- No file paths, filenames, or directory structures
- No image data or colour values
- No AI prompts or API keys
- No personal information, usernames, or hostnames
- No IP-based geolocation

## How it works

Telemetry is sent via HTTPS to a self-hosted [statsfactory](https://github.com/jmylchreest/statsfactory) instance, a privacy-first analytics platform. Each installation generates a random identifier that is SHA256-hashed before transmission -- the raw value is never correlated with any personal information.

Events are queued in memory and dispatched asynchronously by the statsfactory SDK's background worker, which batches events (up to 25 per request) and flushes before the process exits. The HTTP request has a 5-second timeout and failures are silently ignored, so telemetry never blocks or affects CLI behaviour.

### Installation ID

The anonymous installation identifier is stored separately from the main config, as a plain 64-character hex string in `~/.config/tinct/telemetry.id`. It is auto-generated on first run and should not be edited manually.

## Disabling telemetry

### Environment variable (takes priority)

```bash
export TINCT_TELEMETRY=off
```

Any of `off`, `false`, `0`, or `no` (case-insensitive) will disable telemetry.

### Config file

Edit `~/.config/tinct/tinct.toml`:

```toml
[telemetry]
enabled = false
```

:::note
Prior to v0.1.26, telemetry settings were stored in `~/.local/share/tinct/telemetry.json`. On first run, tinct automatically migrates the installation ID to `~/.config/tinct/telemetry.id` and removes the old file.
:::

## Why we collect telemetry

Tinct ships more than 35 output plugins and a range of input sources. Telemetry helps us:

- **Prioritise maintenance** -- focus on the plugins people actually use
- **Track plugin adoption** -- understand which external plugins are popular
- **Detect issues** -- spot high failure or skip rates for specific plugins or platforms
- **Guide development** -- know which features (AI input, ambient extraction, dual themes) are used in practice
