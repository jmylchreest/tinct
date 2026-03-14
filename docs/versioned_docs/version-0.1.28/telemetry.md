---
sidebar_position: 3
---

# Telemetry

Since v0.1.25, tinct collects anonymous usage telemetry to help prioritise plugin development and maintenance. Telemetry is **opt-out** (enabled by default) and can be easily disabled.

## What is collected

When you run `tinct generate`, two types of events are sent:

### Generate event (one per invocation)

| Property | Example | Purpose |
|----------|---------|---------|
| `arch` | `amd64` | CPU architecture |
| `version` | `0.1.25` | Tinct version |
| `input_plugin` | `image` | Which input source was used |
| `output_plugins` | `kitty,waybar` | Which output plugins were requested |
| `theme_type` | `dark` | Detected theme type |
| `seed_mode` | `content` | Seed mode used |
| `backend` | `kmeans` | Colour extraction backend |
| `extract_ambience` | `false` | Whether ambient extraction was enabled |
| `dry_run` | `false` | Whether this was a dry run |
| `dual_theme` | `true` | Whether both themes were generated |
| `ai_input` | `false` | Whether an AI input plugin was used |

### Plugin used event (one per output plugin)

| Property | Example | Purpose |
|----------|---------|---------|
| `plugin_name` | `kitty` | Plugin name |
| `plugin_version` | `0.1.25` | Plugin version |
| `is_external` | `false` | Whether it is an external plugin |
| `status` | `ok` | Outcome: `ok`, `failed`, or `skipped` |

### System properties (automatic)

| Property | Example | Purpose |
|----------|---------|---------|
| `osName` | `linux` | Operating system |
| `appVersion` | `0.1.25` | Tinct version |
| `sdkVersion` | `tinct-telemetry@0.1.25` | SDK identifier |

### What is NOT collected

- No file paths, filenames, or directory structures
- No image data or colour values
- No AI prompts or API keys
- No personal information, usernames, or hostnames
- No IP-based geolocation beyond country (handled by Aptabase)

## How it works

Telemetry is sent via HTTPS to [Aptabase](https://aptabase.com), an open-source, privacy-first analytics platform hosted in the EU. Each installation generates a random identifier that is SHA256-hashed before storage — the raw value is never transmitted.

Events are queued in memory and dispatched asynchronously as a single batched HTTP request at the end of each command. The request has a 5-second timeout and failures are silently ignored, so telemetry never blocks or affects CLI behaviour.

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

Tinct supports 25+ output plugins and multiple input sources. Telemetry helps us:

- **Prioritise maintenance** -- focus on the plugins people actually use
- **Track plugin adoption** -- understand which external plugins are popular
- **Detect issues** -- spot high failure or skip rates for specific plugins or platforms
- **Guide development** -- know which features (AI input, ambient extraction, dual themes) are used in practice
