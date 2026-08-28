---
sidebar_position: 100
sidebar_label: Changelog
toc_max_heading_level: 2
---

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

## [Unreleased]

## [0.5.0]
*2026-08-28*

### Breaking changes

- Input plugins must now speak protocol `0.2.0` or newer. `WallpaperRawPath` was added after the `0.1.0` bump but before `0.2.0`, so a `0.1.0` input plugin may not implement it — and the host calls it unconditionally, where an absent method blocked the run indefinitely rather than erroring. Older input plugins are refused at registration with a message naming the gap; run `tinct plugins update <name>`. Output plugins are unaffected and keep the `0.0.1` floor: every RPC the host calls on them unconditionally predates `0.1.0`.

### Added

- The `noctalia` and `spicetify` plugins are now built and published with releases. Both existed only as source before, so `tinct plugins install noctalia` could never succeed.
- `hooks.Spec.RequiredAny` — a satisfied-by-any detection group (`AnyOf{Binaries, Dirs, Reason}`). Groups are ANDed, entries within a group ORed, so plugins whose app has more than one valid install shape no longer need a hand-rolled `PreExecute`. Paths are checked with `os.Stat`, so plain files (`kdeglobals`, `config.toml`) count as markers; binaries still resolve through appdetect (PATH/Flatpak/AppImage). An optional per-group `Reason` replaces the generated skip message.
- Multi-line skip reasons and `Instructions` are formatted by the hook runner. `hooks.Indent` dedents continuation lines and re-indents them under the header, preserving relative nesting so a command indented under a numbered step stays indented.
- Optional `Configurable` plugin interface and `Plugin.Configure` RPC. The host pushes args, dry-run and verbose into a plugin before the pre-execute lifecycle, so `PreExecute` and `Hooks` can honour flags instead of assuming defaults. Plugins that do not implement it are unaffected — unknown-method errors fall back like `Validate` and `GetHooks`.

### Changed

- External plugins keep one subprocess for a whole run instead of spawning a fresh one per RPC. A full generate went from roughly 25 plugin launches to 3 (5 to 1 for a single-plugin run), which also all but removes the stray `yamux: Failed to write header` lines those teardowns produced.
- The noctalia plugin skips the shell reload when the regenerated palette is byte-identical to the one already on disk. A reload makes Noctalia re-read its whole config tree, which is wasted work when the colours have not moved — regenerating from the same wallpaper now costs nothing.
- kde-plasma, rosec, qt5, qt6 and spicetify move their installation detection (and, for spicetify, the manual-apply message) off hand-rolled `PreExecute`/`PostExecute` onto the declarative `hooks.Spec`.
- CI now builds and tests the contrib plugin modules, and the lint, `go vet` and formatting steps can actually fail. They carried `continue-on-error`, so the Lint job reported success while printing `##[error]issues found`.

### Fixed

- The noctalia output plugin now reloads the running shell after writing the palette. Noctalia does not watch `~/.config/noctalia/palettes/` — its inotify watch is non-recursive and only reacts to `*.toml` — so a regenerated palette never reached the running shell and the colours stayed as they were at startup. The plugin runs `noctalia msg config-reload`, with `noctalia` as an optional binary.
- Plugin RPCs are now bounded by a deadline. `net/rpc` has no timeouts, so a plugin that never answers hung tinct indefinitely with no output and nothing to diagnose. Metadata-style calls get 30s; `Generate` gets 10 minutes, since an input plugin may legitimately be waiting on remote image generation.
- The protocol floor is enforced rather than nominal. `MinCompatibleVersion` was `0.0.1` while the only other rule was "major version must match", so every `0.x` plugin ever built passed and the gate never fired.
- An outdated plugin is reported without `--verbose`. It was rejected silently, leaving `unknown plugin: <name>` as the only symptom, which points at entirely the wrong problem.
- `tinct plugins update <name>` now updates only the named plugin. The command declared no positional arguments and never read them, so cobra accepted the name, discarded it, and updated every plugin in the manifest. An unknown name is now an error listing what is installed, rather than a silent full update.
- `tinct plugins update` can update plugins added from a local path. Local sources record their origin in `source.original_path`, which the update path never consulted — so every plugin installed with `tinct plugins add ./binary` reported "No source information" and could never be updated.
- `--noctalia.output-dir` was accepted and silently ignored — the plugin never read `PaletteData.PluginArgs` and its output directory was never assigned. It now arrives via `Configure` before `Hooks`/`PreExecute`, expands `~`, and suppresses the config-directory check when set.
- External input plugins no longer lose their debug output under `--verbose`. Verbose is set on the plugin before `Validate`, which is the first RPC and therefore the call that starts the subprocess whose logger is fixed for its lifetime.

### Breaking changes

- fix(plugin)!: bound RPC calls and enforce the protocol floor (#11) (5ae7cd0)

### Added

- feat(hooks): add any-of detection groups and message formatting (#15) (805022c)

### Fixed

- fix(release): build and publish the noctalia and spicetify plugins (#24) (9c9d5ee)
- fix: clear the pre-existing lint backlog and make lint a real gate (#23) (8e6ed6e)
- fix: resolve lint regressions introduced this cycle (#22) (3442aeb)
- fix(spicetify): update tests to the declarative detection contract (#17) (023413c)
- fix(noctalia): reload the shell after writing the palette (#13) (621ec16)
- fix(cli): honour the plugin name argument in `tinct plugins update` (#12) (53ae210)
- fix(plugin)!: bound RPC calls and enforce the protocol floor (#11) (5ae7cd0)

### Changed

- perf(noctalia): skip the reload when the palette is unchanged (#16) (da850e8)
- perf(plugins): reuse one plugin subprocess per run (#14) (9832096)

## [0.4.2]
*2026-08-14*

### Added

- feat(noctalia): add Noctalia shell external output plugin (144bddc)

### Fixed

- fix(google-genai): name generated images by actual format, not hardcoded .png (8b40397)

## [0.4.1]
*2026-06-30*

### Added

- feat(niri): add niri compositor output plugin (5a1ee49)

## [0.4.0]
*2026-06-20*

### Added

- feat(minimax): add MiniMax (Hailuo) image input plugin (5cf1b8f)
- feat(google-genai): refresh image model catalogue with Gemini 3 models (cd63965)
- feat(input): trim letterbox borders from AI-generated images (d311e26)

### Fixed

- fix(deps): bump golang.org/x/net for GO-2026-5026 and refresh dependencies (9266d16)
- fix(docs): rewrite archived absolute template links to relative paths (5f8b1d0)
- fix(scripts): promote-unreleased exits 0 when both bodies are empty (2fcb8fe)

## [0.3.3]
*2026-05-25*

### Added

- New built-in output plugins: **helix** (Helix editor themes, `inherits = "tinct-palette"` directive), **foot** (foot terminal, SIGUSR1 reload) (3b1622e)
- New contrib output plugin: **spicetify** (Spotify theming via Spicetify CLI, manual apply) (3b1622e)
- Contrib plugins migrated to the canonical README schema with populated `--plugin-info` Metadata: awob, dunstify, keylightd-tray, ptyxis, templater, wob, zed (output); paletty, random (input) (3b1622e)
- External plugins now accept the same `--<plugin>.<flag>` syntax as built-ins (e.g. `--paletty.palette=…`); `--plugin-args` JSON remains as the escape hatch and for json-stdio plugins (3b1622e)
- External plugin flag-help cached in the install manifest at install/update time, eliminating the per-invocation `--plugin-info` spawn on cold start (3b1622e)
- `tinct-check-readmes` CLI: warn-only drift detector that diffs each plugin's README frontmatter against runtime behaviour (built-in via package import, external via `--plugin-info`) (3b1622e)
- `PLUGIN-README-STANDARD.md`: canonical README schema for every plugin (input + output, builtin + external), enforced by `tinct-check-readmes` and consumed by the docs pipeline (3b1622e)
- Auto-generated Docusaurus plugin pages from in-tree READMEs via `docs/scripts/sync-plugin-readmes.mjs`, including auto-regenerated landing-page tables and `autogenerated` sidebar entries (3b1622e)
- `internal/plugin/output/categories.json` as the single source of truth for output plugin categories, consumed by the sync script, sidebar, and check tool (3b1622e)
- Changelog automation: `scripts/{changelog-sections,promote-unreleased,backfill-changelog}.sh` plus release-time integration; 11 historical empty version sections backfilled from conventional-commits history (9d34abc)

### Fixed

- waybar: `New()` now sets `reloadConfig: true` to match the `--waybar.reload` default; previously the constructor and flag disagreed (3b1622e)
- gnome-shell: `--help` text now reflects the actual `DefaultOutputDir()` (`~/.local/share/themes`); previously advertised a `/tinct/gnome-shell` subdirectory the runtime never used (3b1622e)
- hyprpaper: migrated to declarative `Hooks()` (`SupportsWallpaper`, `OptionalBinaries=[hyprctl]`, `AutoCreateDir`) so `reload.method: wallpaper-apply` is machine-verifiable (3b1622e)
- Docs build: MDX-incompatible content in wob, templater, and `plugins/overview.md` after the contrib migration (3b1622e)
- `scripts/promote-unreleased.sh`: trailing `[ -n "$var" ] && echo` lines became the script's exit status under `set -e`, failing release docs when both human and auto bodies were empty (2fcb8fe)
- Archived versioned docs: absolute `/docs/plugins/output/special/template` URLs rewritten to relative paths so each version links to its own snapshot's `template.md` after the rename to `templater` in current docs (5f8b1d0)

### Changed

- tmux & rofi promoted to the two-file pattern: palette file (`tinct-palette.conf`, `tinct-colours.rasi`) plus main config that references it. tmux gains a `ReloadFn` that runs `tmux source-file` only when at least one session exists (3b1622e)
- `tinct generate --help`: plugin-flags section split into builtin vs external with working examples for both `--<plugin>.<flag>` and `--plugin-args` JSON (3b1622e)
- Lint sweep + cross-module dead-code triage (9ce39aa)
- Contrib plugin binaries removed from git (~55 MB) and gitignored; `.gitignore` extended for `tinct-check-readmes` and `contrib/**/__pycache__/` (3b1622e)

## [0.3.2]
*2026-05-06*

### Fixed

- fix(plugin): stop external plugins from leaking across CLI invocations (f87c41b)

## [0.3.1]
*2026-05-06*

### Added

- feat(plugin/warp): add Warp terminal output plugin (99ba4d1)

### Fixed

- fix: ship paletty in releases, fix docs build, flush plugin telemetry (11c5686)

### Changed

- refactor(cli/telemetry): centralise plugin telemetry on cobra.OnFinalize (55573e4)

## [0.3.0]
*2026-05-04*

### Added

- feat(plugin): expose external plugin templates via TemplateLister RPC (84eafa3)
- feat(telemetry): emit plugins_command event for each subcommand (582d35e)
- feat(plugin/random): adopt Validator for plugin-args (bb6fb04)
- feat(cli/plugins): auto-prune missing-binary entries from manifest views (0174610)
- feat(plugin/paletty): add external paletty.dev input plugin (ca0e991)
- feat(plugin): protocol 0.3.0 — role hints, theme hints, validator, hooks-provider (cb636ff)

### Fixed

- fix(plugin/awob): drop unconditional PostExecute banner (0509d45)

### Changed

- refactor(cli/plugins): rename "lock file" → "manifest" (1e015a0)
- refactor(plugin/awob,ptyxis): migrate static instructions to HooksProvider (c145d6b)

## [0.2.4]
*2026-05-04*

### Added

- New external plugin **awob** for the [animated Wayland Overlay Bar](https://github.com/jmylchreest/awob).
  Writes `~/.config/awob/themes/_palettes/tinct.kdl` (a shared palette +
  named-style block) plus a `tinct/` theme directory containing
  `scene.kdl` and `manifest.toml`. Other awob themes can adopt the
  tinct palette by adding `import "../_palettes/tinct.kdl"` to their
  scene; the palette refreshes in place on every `tinct generate` and
  awob-daemon hot-reloads. Install with `tinct plugins install awob`.

### Removed

- **BREAKING**: AUR per-plugin packages (`tinct-plugin-random-bin`,
  `tinct-plugin-wob-bin`, `tinct-plugin-templater-bin`,
  `tinct-plugin-dunstify-bin`, `tinct-plugin-zed-bin`,
  `tinct-plugin-opencode-bin`, `tinct-plugin-keylightd-tray-bin`) are no
  longer built. Plugins should be installed via the plugin manager
  (`tinct plugins install <name>`).

  This is the canonical install path — it pins versions, sandboxes
  installs to `~/.local/share/tinct/plugins/`, and uninstalls cleanly.
  The `tinct-bin` AUR package for tinct itself remains supported. The
  goreleaser config keeps a commented-out `tinct-plugin-templater-bin`
  block as a copy-paste reference for anyone re-enabling AUR
  distribution for a specific plugin.

### Changed

- **BREAKING**: `ptyxis` is no longer a built-in output plugin. It has
  been migrated to an external plugin so its bespoke Flatpak detection,
  dual-mode `[Light]`/`[Dark]` palette generation, and live-reload
  install hint can iterate independently of tinct releases. Users who
  previously relied on the built-in plugin can install the external
  one with:

      tinct plugin install ptyxis

  Behaviour and output paths are unchanged once installed. The
  `--ptyxis.output-dir` flag becomes `--ptyxis.output-dir` on the
  external plugin (same name, same semantics) and continues to honour
  the `TINCT_PLUGIN_PTYXIS_OUTPUT_DIR` environment variable.

## [0.2.3]
*2026-05-03*

### Fixed

- fix(ptyxis): write to watched dir and emit valid Light/Dark palettes (a891dd7)

### Changed

- refactor: extract shared helpers to eliminate code clones (9e0965c)

## [0.2.2]
*2026-04-10*

### Added

- feat: add rosec-prompt output plugin and update dependencies (0800d0c)
- feat: add standalone lualine theme generation to neovim plugin (bd89c2d)

### Fixed

- fix: resolve all golangci-lint issues (aad4050)

## [0.2.1]
*2026-03-18*

### Fixed

- fix: use secrets with vars fallback for telemetry credentials in CI (a6ae7a8)

## [0.2.0]
*2026-03-18*

### Fixed

- fix: implement JSON-stdio structured response protocol (v0.2.0) (c65bbb3)

### Changed

- refactor: replace aptabase telemetry with statsfactory SDK (ab6d98a)

## [0.1.28]
*2026-03-14*

### Fixed

- fix: restore wallpaper format detection for URLs without file extensions (60810ea)

### Changed

- refactor: remove security.ValidateFilePath in favour of plugin-managed output paths (52770fb)
- refactor: eliminate 3 remaining clone groups (RPC client, path utils, sync finalization) (0104c9a)
- refactor: extract shared two-pass tar extraction into extractFromTar (aab71bb)
- refactor: eliminate code duplication across 6 clone groups (da18f5a)

## [0.1.27]
*2026-03-12*

## [0.1.26]
*2026-02-20*

### Added

- feat: centralise config to tinct.toml, async batched telemetry, remove plugin enable/disable (6b11cf8)

### Fixed

- fix: resolve lint warnings (errcheck, errname, gocognit) (0a8151c)

## [0.1.25]
*2026-02-19*

### Added

- feat: add anonymous usage telemetry via Aptabase (b3c137b)
- feat: add wbg (simple Wayland wallpaper) output plugin (7127b75)
- feat: add awww (animated Wayland wallpaper daemon) output plugin (e8049b8)
- feat: add Midnight Commander (mc) skin plugin and xterm256 color conversion (46bed02)
- feat: add OpenCode theme output plugin (0044489)

### Fixed

- fix: deploy docs from latest main after versioning (b7d0c7b)

## [0.1.24]
*2026-02-10*

### Added

- feat: add Midnight Commander (mc) skin plugin and xterm256 color conversion (3d3541a)

## [0.1.23]
*2026-02-10*

### Added

- feat: add OpenCode theme output plugin (ba85ac0)

### Fixed

- fix: changelog TOC and release workflow (64f4995)

## [0.1.22]
*2026-01-02*

### Added
- Docusaurus documentation site at [jmylchreest.github.io/tinct](https://jmylchreest.github.io/tinct/)
- Versioned template system for breaking config changes
- Plugin development documentation with protocols, hooks, and publishing guides

### Changed
- Simplified README with links to documentation site

### Fixed
- hyprpaper v0.8.0 support with versioned templates

## [0.1.21]
*2025-12-31*

### Added
- histui output plugin for notification daemon theming

### Fixed
- Security annotations for config file reads (G304)

## [0.1.20]
*2025-12-31*

### Changed
- Tightened file and directory permissions for security

## [0.1.18]
*2025-12-24*

### Added
- ptyxis terminal output plugin

## [0.1.17]
*2025-12-14*

### Added
- konsole terminal output plugin

## [0.1.16]
*2025-12-14*

### Added
- zed editor output plugin (contrib)
- templater output plugin (contrib)
- keylightd-tray output plugin (contrib)
- wob bar output plugin (contrib)

## [0.1.15]
*2025-12-14*

### Added
- External plugin repository system
- Plugin installation from git repositories

## [0.1.14]
*2025-12-14*

### Added
- JSON-stdio protocol for external plugins
- go-plugin RPC protocol support

## [0.1.11]
*2025-12-08*

### Added
- markdown input/output plugins for theme portability
- Theme save and restore functionality

## [0.1.10]
*2025-11-26*

### Added
- openrouter input plugin for AI-generated themes
- google-genai input plugin for Gemini integration

## [0.1.8]
*2025-11-25*

### Added
- remote-json input plugin
- remote-css input plugin

## [0.1.5]
*2025-11-19*

### Added
- Ambient edge/corner colour extraction for LED sync
- Position-based colour roles

## [0.1.0]
*2025-11-16*

### Added
- Initial release
- Image colour extraction with K-means clustering
- Material Design 3 semantic colour roles
- Output plugins: alacritty, kitty, ghostty, hyprland, hyprlock, hyprpaper, waybar, dunst, gtk3, gtk4, qt5, qt6, gnome-shell, kde-plasma, neovim, zellij, fuzzel, walker, wofi, swayosd, libadwaita
- Template system with colour format conversion
- File tracking and management
