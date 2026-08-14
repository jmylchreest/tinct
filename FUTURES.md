# Futures / TODO

Ideas and follow-ups that are too large or too speculative for the current
work cycle. Each entry should be self-contained — anyone (including future
us) should be able to pick one up cold.

## Style presets (vibrancy / mute / pastel)

**Status**: not started · **Estimated effort**: half a day + tuning · **Risk**: medium (visual judgement required)

### Why

Matugen exposes nine Material You scheme algorithms (Content, Expressive,
Fidelity, FruitSalad, Monochrome, Neutral, Rainbow, TonalSpot, Vibrant).
That breadth is matugen's wrapper around Google's Material Color Utilities
spec — we don't have an equivalent upstream to port, so cloning all nine
would mean inventing nine. But users do reasonably ask "can I make this
more vibrant / more muted?" and tinct has no answer today beyond the
existing `--no-desaturation` toggle for alternate themes. A small, well-
chosen preset system gives the control without the complexity.

### What

A single `--style` persistent flag (mirroring `--theme`) accepting:

- `standard` — current default. Zero behavioural change.
- `vibrant` — boosted saturation, narrower hue tolerance, higher
  semantic-colour boost. Punchy. For users who want their theme to read.
- `muted` — reduced saturation, lower semantic boost. Calm. For users who
  prefer the colours to recede.
- `pastel` — reduced saturation + raised luminance. Soft. Mostly useful
  on light themes; on dark themes it tends toward indistinct.

Default stays `standard`. Existing behaviour is byte-identical for users
who don't pass the flag.

### How

`internal/colour/categorisation.go::CategorisationConfig` already exposes
the relevant knobs:

```go
type CategorisationConfig struct {
    ThemeType             ThemeType
    MinContrastRatio      float64
    RequireAAA            bool
    MutedLuminanceAdjust  float64
    EnhanceSemanticColors bool
    SemanticBoostAmount   float64
}
```

Add a `Style` field plus a derived `SaturationFactor` and
`LuminanceShift`:

```go
type Style string

const (
    StyleStandard Style = "standard"
    StyleVibrant  Style = "vibrant"
    StyleMuted    Style = "muted"
    StylePastel   Style = "pastel"
)

type CategorisationConfig struct {
    // existing fields ...
    Style              Style
    SaturationFactor   float64 // 1.0 default; preset-derived
    LuminanceShift     float64 // 0.0 default; preset-derived (positive = lighter)
}

func (s Style) ApplyTo(cfg *CategorisationConfig) {
    switch s {
    case StyleVibrant:
        cfg.SaturationFactor = 1.3
        cfg.SemanticBoostAmount = 0.5
    case StyleMuted:
        cfg.SaturationFactor = 0.7
        cfg.SemanticBoostAmount = 0.15
    case StylePastel:
        cfg.SaturationFactor = 0.6
        cfg.LuminanceShift = 0.10
    case StyleStandard, "":
        // no-op — keep DefaultCategorisationConfig values
    }
}
```

Apply `SaturationFactor` and `LuminanceShift` inside:
- `internal/colour/accents.go::sortAccentsForTheme` — when scoring accent
  saturation.
- `internal/colour/semantic.go::enhanceSemanticColour` — when boosting
  semantic colour saturation.
- `internal/colour/muted.go` — when computing muted variants.

### Wiring

- `internal/cli/root.go` — add `--style` persistent flag, default
  `"standard"`, validated against the `Style` enum.
- `internal/cli/generate_helpers.go::categorizePalette` — pass the style
  through to `CategorisationConfig` before calling `Categorise`.
- Surface roles in `internal/colour/surface.go` already use parameterised
  helpers (`generateLuminanceOffset`, `generateFixedLuminance`); their
  `satFactor` becomes a function of `cfg.SaturationFactor`.

### Testing

Snapshot tests against a fixed input image, locking in the expected
palette output per style. Place under
`internal/colour/style_snapshot_test.go`. The point isn't byte-stability
(tuning will require revisiting these) but regression detection: if a
future change inadvertently shifts the standard preset, the diff will
flag it.

### What to resist

- **More than four presets.** Matugen has nine; users meaningfully use
  two. Adding `synthwave`, `solarized`, etc. is a slippery slope and a
  tuning rabbit hole. Anyone who wants those can already get them via
  the `remote-json` or `markdown` input plugins with a curated palette.
- **Per-output-plugin style override.** The pipeline categorises once
  and fans out. Per-output style means re-categorising N times and
  stitching together. Wait until someone asks.
- **Exposing the underlying knobs on the CLI** (`--saturation-factor=...`).
  Keep them implementation detail; the user-facing promise is "vibrant /
  muted / pastel", not "saturation 1.3, semantic boost 0.5".

### Acceptance criteria

- `--style standard` (or omitted) produces byte-identical output to
  current behaviour. Verified by replaying existing snapshot tests.
- `--style vibrant`, `--style muted`, `--style pastel` produce visibly
  distinct palettes from the same input image, each consistent with its
  name on at least three test images (sunset, forest, cityscape).
- Help text on the flag reads as a one-liner: "palette style preset
  (standard, vibrant, muted, pastel)".
- Documentation updated: a single short page under
  `docs/docs/concepts/styles.md` with the four presets, an example image
  per preset, and the "no per-output override" note.

---

## Auto-upgrade external plugins on execution

**Status**: not started · **Estimated effort**: half a day · **Risk**: medium (tinct executing remote code without explicit user action)

### Why

External plugins are installed via `tinct plugins install <repo>` and
pinned in `plugins.manifest.json`. Today users have to run `tinct plugins update`
manually to pick up new releases — easy to forget, leaving themes stale
or worse, broken against a tinct version they no longer match. Matugen
sidesteps this by being template-driven (no plugins). For tinct's
plugin model, an opt-in auto-upgrade-on-run gives the same
"just works" feel.

### What

A persistent setting plus a corresponding flag/env override:

```toml
[plugins]
# Check installed external plugins for upgrades on every `tinct generate`.
# Default: false (explicit upgrade only).
auto_upgrade = false

# How often to re-check upstream when auto_upgrade is enabled.
# Values: "always", "daily", "weekly", "never". Default: "daily".
auto_upgrade_interval = "daily"
```

CLI override: `--plugins.auto-upgrade=true|false`.
Env override: `TINCT_PLUGINS_AUTO_UPGRADE=true`.

Behaviour when enabled:
1. Before resolving plugins, walk `plugins.manifest.json` external entries.
2. For each, check the timestamp of the last upgrade probe (cached in
   `$XDG_STATE_HOME/tinct/upgrade-cache.json`).
3. If older than `auto_upgrade_interval`, query the source (GitHub
   releases / git refs depending on plugin protocol).
4. If a newer compatible version is available, upgrade in place
   (download + replace + update lockfile entry).
5. Continue with generation.

Failures (network down, repo unreachable, version incompatible) are
non-fatal: log a warning, fall through to the pinned version.

### How

Resolver flow already lives in `internal/pluginconfig`; add a
`Resolve` call site for `plugins.auto_upgrade` next to the existing
per-plugin keys. The upgrade machinery is in
`internal/plugin/repository/` — add a `CheckAndUpgrade(ctx, lock,
interval)` method that returns a list of upgraded entries plus errors,
called from `loadAndConfigurePlugins` in `internal/cli/generate_helpers.go`
before plugin registration.

For the cache, a tiny JSON file under `$XDG_STATE_HOME/tinct/` keyed by
`plugin name → {last_check: timestamp, last_upgrade: timestamp}`. No
schema migration concerns because the file is regenerated on miss.

### What to resist

- **Auto-upgrade by default.** Surprise upgrades break user trust. Stay
  opt-in. Document loudly in the changelog and config reference.
- **Auto-upgrading to incompatible majors.** If a plugin's
  protocol-version requirement diverges from tinct's, skip with a
  warning. Don't break running setups silently.
- **Hammering remotes on every run.** `auto_upgrade_interval` exists
  precisely so a user with `auto_upgrade=true` doesn't issue 50 GitHub
  API calls on every `tinct generate`.
- **Auto-installing new plugins.** Out of scope. Auto-upgrade only
  touches plugins already in `plugins.manifest.json`.

### Acceptance criteria

- `auto_upgrade = false` (default) preserves current behaviour exactly.
- With `auto_upgrade = true`, running `tinct generate` against a
  lockfile with stale plugins triggers per-plugin upgrade checks and
  applies any compatible upgrades before generating.
- Network failures during the upgrade check produce a single warning
  line each and never abort generation.
- The upgrade cache prevents repeated probes within the configured
  interval; a manual `tinct plugins update` always forces a check.
- Documented in `docs/docs/configuration.md` alongside the existing
  `[plugin.<name>]` and `[telemetry]` tables.

---

## (placeholder for future entries)

Add new entries above this line. Convention: each future has a status,
estimated effort, risk, and clear acceptance criteria. If an idea can't
articulate acceptance criteria yet, it belongs in a chat or a sketch
file, not here.
