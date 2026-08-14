# Plugin README Standard

Canonical structure for end-user documentation of every tinct plugin — both **input** plugins (palette sources) and **output** plugins (configuration generators).

The README that lives next to a plugin's source code is the **single source of truth** for that plugin's user-facing docs. The Docusaurus site at `docs/plugins/{input,output}/...` is built from these READMEs, not maintained separately. Keeping them next to the code is how we stop the kind of drift that left the old `dunst/README.md` describing a plugin behaviour that hadn't existed for several releases.

A small warn-only CI tool (`cmd/tinct-check-readmes`) parses the frontmatter and compares it against the plugin's actual runtime behaviour. Mismatches surface as warnings; they do not fail the build.

## Input vs output plugins

The schema branches on a mandatory `plugin.type` discriminator:

- **`type: output`** — generates configuration files for an external app (terminals, desktops, bars, editors, …). Has integration paths, reload behaviour, an uninstall flow.
- **`type: input`** — produces a colour palette from a source (image, file, URL, API). Stateless; no config files to write; no app to reload.

Most fields are type-specific. Both types are subject to the same drift detection and the same docs-build skip rule.

## Table of contents

1. [Where READMEs live](#where-readmes-live)
2. [Frontmatter schema — output plugins](#frontmatter-schema--output-plugins)
3. [Frontmatter schema — input plugins](#frontmatter-schema--input-plugins)
4. [Body sections — output plugins](#body-sections--output-plugins)
5. [Body sections — input plugins](#body-sections--input-plugins)
6. [The uninstall / revert section](#the-uninstall--revert-section)
7. [Docs site build behaviour](#docs-site-build-behaviour)
8. [Drift detection](#drift-detection)
9. [Worked examples](#worked-examples)

---

## Where READMEs live

| Plugin source | README location |
|---|---|
| Built-in output plugins | `internal/plugin/output/<name>/README.md` |
| Built-in input plugins | `internal/plugin/input/<name>/README.md` |
| External Go plugins (in this repo) | `contrib/plugins/{input,output}/<name>/README.md` |
| External script plugins (in this repo) | `contrib/plugins/{input,output}/<basename>.md` alongside the script |
| Third-party external plugins | Author's repo (encouraged to follow this schema for consistency) |

The Docusaurus build collects all of the above and routes each by its frontmatter:
- `type: output` → `docs/plugins/output/<category>/<name>.md`
- `type: input` → `docs/plugins/input/<name>.md` (flat, no sub-categorisation)

---

## Frontmatter schema — output plugins

YAML frontmatter at the top of the README. Two key blocks:

- **Docusaurus block** (`title`, `sidebar_position`) — present so the same file works unmodified as a docs site page.
- **`plugin:` block** — tinct-specific metadata consumed by `tinct-check-readmes` and the docs build.

### Required fields

```yaml
---
title: <plugin name>                  # used by Docusaurus as the page title
sidebar_position: <int>               # docs site ordering within its category

plugin:
  type: output                        # MANDATORY discriminator
  name: <string>                      # MUST match Plugin.Name()
  category: <enum>                    # see categories.json — the single source of truth shared by the docs build, sidebars.ts, and tinct-check-readmes
  source: <enum>                      # builtin | external
  app: <string>                       # human-readable app name, e.g. "Dunst"
  app_url: <url>                      # upstream project URL
  requires: [<binary>, ...]           # binaries that MUST be on PATH
  optional: [<binary>, ...]           # binaries used opportunistically (e.g. reload helpers)
  pattern: <enum>                     # two-file | single-file | drop-in | flavor-pack — optional; omit for plugins that don't write files (e.g. notification-only plugins)
  default_output_dir: <path>          # ~ left unexpanded
  generated_files: [<name>, ...]      # filenames inside default_output_dir
  reload:
    method: <enum>                    # ipc | signal | watch | wallpaper-apply | none
    command: <string>                 # descriptive; the actual reload action
    user_action_required: <bool>      # true if user must reload manually
---
```

The `wallpaper-apply` method is for plugins whose "reload" is the act of pushing a new wallpaper (awww, wbg, hyprpaper). These declare `SupportsWallpaper: true` in their `hooks.Spec` and the check tool derives the method from that.

### External-only fields

When `source: external`, additionally:

```yaml
plugin:
  version: <semver>                   # plugin's own version, independent of tinct
  protocol_version: <semver>          # e.g. "0.2.0"
  repository: <url>                   # plugin source repo
  install: <command>                  # exact install command, e.g. "tinct plugins install opencode"
```

### Field semantics

| Field | Why it matters | How CI checks it |
|---|---|---|
| `name` | must match the plugin's runtime identity | builtin: `Plugin.Name()`. external: `--plugin-info`.`name` |
| `category` | drives docs site routing | enum validation |
| `source` | discriminator for everything below | enum validation |
| `requires` / `optional` | sets user expectations about prerequisites | builtin: `Hooks().RequiredBinaries`/`OptionalBinaries`. external: `--plugin-info`.`metadata.required_binaries`/`optional_binaries` |
| `pattern` | tells readers how files relate | external: `--plugin-info`.`metadata.pattern` (builtin is documentation-only) |
| `default_output_dir` | where files land | builtin: `Plugin.DefaultOutputDir()` (with `$HOME` re-folded to `~`). external: `--plugin-info`.`metadata.default_output_dir` |
| `generated_files` | what the user can `rm` to uninstall | external: `--plugin-info`.`metadata.generated_files` (builtin is documentation-only) |
| `reload.method` | does the user need to do anything? | builtin: derived from `Hooks().Reload` (`VerbSignal` → `signal`, `VerbExec` → `ipc`), `ReloadFn` → `ipc`, `SupportsWallpaper: true` → `wallpaper-apply`, none of the above → `none`. external: `--plugin-info`.`metadata.reload.method` |
| `version` / `protocol_version` | external plugin identity | `--plugin-info`.`version` / `.protocol_version` |

Fields that the CI tool cannot mechanically verify against runtime are documentation-only — they're not checked but are still required so the docs site renders consistently.

---

## Frontmatter schema — input plugins

Input plugins produce palettes; they don't write app configs. The schema reflects that.

### Required fields

```yaml
---
title: <plugin name>
sidebar_position: <int>

plugin:
  type: input                         # MANDATORY discriminator
  name: <string>                      # MUST match Plugin.Name()
  source: <enum>                      # builtin | external
  source_type: <enum>                 # image | file | url | api | generated
  description: <string>               # one-line summary of what palette source this produces
---
```

### Optional fields

```yaml
plugin:
  # For api / url source_type — name of the upstream service
  service: <string>                   # e.g. "Google Gemini", "OpenRouter", "remote JSON"
  service_url: <url>                  # upstream URL

  # Local prerequisites (rare for input plugins)
  requires: [<binary>, ...]
  optional: [<binary>, ...]

  # Runtime behaviour
  requires_network: <bool>            # true for url / api source_types
  requires_credentials: [<ENV_VAR>, ...]  # env vars expected (e.g. GOOGLE_API_KEY); empty if none
  produces_wallpaper: <bool>          # true if the plugin also produces a wallpaper image alongside the palette
```

### External-only fields

When `source: external`, additionally:

```yaml
plugin:
  version: <semver>
  protocol_version: <semver>
  repository: <url>
  install: <command>
```

### Field semantics (input plugins)

| Field | Why it matters | How CI checks it |
|---|---|---|
| `type` | discriminator; routes to `docs/plugins/input/` | enum validation |
| `name` | runtime identity | builtin: `Plugin.Name()`. external: `--plugin-info`.`name` |
| `source` | discriminator for everything below | enum validation |
| `source_type` | tells readers what input to provide | enum validation |
| `service` / `service_url` | makes external services discoverable | not machine-verifiable |
| `requires_network` | warns offline users | not machine-verifiable |
| `requires_credentials` | tells users which env vars to set | not machine-verifiable |
| `produces_wallpaper` | enables `--set-wallpaper` flag in output plugins | builtin: derived from `WallpaperProvider` interface |

What input plugins **don't have**: `category` (they're flat), `app`/`app_url` (they're not an integration with one app), `default_output_dir`/`generated_files`/`pattern` (they don't write user config), `reload` (they're stateless).

---

## Body sections — output plugins

Sections appear in this order. Empty sections may be omitted only if explicitly marked optional below.

### 1. What this plugin themes

One paragraph. What the plugin generates, for which app, with a link to the upstream project. No CLI invocations or config blocks — just orientation.

### 2. Installation (external plugins only)

Built-in plugins are skipped. For `source: external`:

- One-line install via `tinct plugins install <name>` if the plugin is in the official repository.
- Build-from-source instructions for users who want a specific version or are developing the plugin.
- How to verify installation (`which tinct-plugin-<name>`, `tinct-plugin-<name> --plugin-info`).

### 3. Quick start

One `tinct generate ...` invocation showing the most common use of this plugin. No flag forest — show the simplest working command.

### 4. Generated files

For each file produced: filename, full default path, one-line description of its role.

For two-file plugins, explicitly label which is the **palette** (importable into user configs) and which is the **example/main config**.

### 5. Integration

How to wire the plugin's output into the target app's existing config. Branches:

- **drop-in pattern**: state explicitly that no manual config is required, and which directory the app auto-loads from.
- **include-based**: show the exact line the user adds to their config (`include=...`, `source=...`, `@import "..."`, `inherits = "..."`).
- **single-file / atomic theme**: show the exact config key that selects the theme (`config.color_scheme = "tinct"`, `[flavor] dark = "tinct"`).
- **no include support**: state this explicitly, then show how to copy the palette section by hand or how the user can install the theme as a whole-file replacement.

### 6. Reload behaviour

Two sub-headings, in this order:

- **Automatic** — what `PostExecute` does (or "none — no automatic reload available"). Be specific: name the verb and command (e.g. `dunstctl reload`, `pkill -USR1 -x foot`, `tmux source-file`).
- **Manual fallback** — what the user does if automatic reload didn't happen (e.g. the app wasn't running at generate time, or the helper binary is missing).

### 7. Uninstall / revert

See the dedicated section [below](#the-uninstall--revert-section). This section is **mandatory** for every plugin — there is no scenario in which a user shouldn't be able to find out how to undo a plugin's effects.

### 8. Flags

Table of `--<plugin>.*` flags with default values and one-line descriptions. Match the table format used elsewhere in tinct docs.

### 9. Colour role mapping

Table of `app's colour key → tinct semantic role`. For apps with many keys, group by section (e.g. tab bar, status line, diagnostics).

### 10. Customising the template

One-liner showing `tinct plugins templates dump -o <plugin> -l ~/.config/tinct/templates/<plugin>`. Link to the templating reference.

### 11. Troubleshooting

Three to five common failure modes. Each as a sub-heading with the symptom, then the fix.

### 12. Related plugins (optional)

Links to plugins for related apps (alternatives, companion plugins, plugins users often configure together).

---

## Body sections — input plugins

Input plugins use a different (shorter) section set. Several output-only sections (Integration, Reload behaviour, Uninstall, Colour role mapping, Customising the template) don't apply.

### 1. What this plugin does

One paragraph: what palette source does it pull from, what kind of palette does it produce. For api/url plugins, name the service.

### 2. Installation

`source: builtin` → one line: "Built into tinct — nothing to install separately."

`source: external` → install command + build-from-source + verification.

### 3. Quick start

One `tinct generate -i <plugin> ...` invocation.

### 4. Configuration / credentials

For api plugins: which env vars are required (`requires_credentials`), where to get credentials (linked), and one-liner examples of fetching them from `secret-tool`, `bw`, `pass`, etc.

For file/image plugins: any prerequisites for the input source.

For url plugins: example URLs, JSONPath/CSS-selector configuration.

### 5. Flags

Table of `--<plugin>.*` flags.

### 6. Output

What kind of palette this produces — colour count, dark/light hint, whether `produces_wallpaper` is set (and where the wallpaper file lands).

### 7. Costs & rate limits (api plugins only, optional)

For Google Gemini, OpenRouter, etc.: free-tier limits, paid pricing tier links, retry/backoff behaviour.

### 8. Troubleshooting

2–4 common failure modes (auth errors, network failures, source-not-found, malformed responses).

### 9. Related plugins (optional)

Links to other input plugins that solve similar use cases.

---

## The uninstall / revert section

Standardised because most plugin READMEs historically skipped it, and the trap cases (state outside `~/.config`, drop-in dirs, flavor packs as directories) are exactly where users hit problems.

Every revert section must answer four questions, in this order:

1. **What config line do I remove?** Exact text the user grep-and-deletes from their app config. For drop-in patterns, state that no config line was added (so this step is skipped).
2. **What files do I delete?** Exact `rm` (or `rm -rf` for directory-based outputs like yazi flavor packs) commands. Use `generated_files` from frontmatter; the docs build can render this from the metadata.
3. **How do I reload to drop the styling?** Same verbs as the reload section, run by the user after deletion.
4. **Did the plugin mutate anything outside `~/.config`?** If yes (spicetify, OpenRGB, system-discovered GTK/Qt themes), the specific commands to undo that mutation. If no, say so explicitly — "this plugin only writes to its output directory; no further cleanup is needed."

Example template:

```markdown
## Uninstall / revert

1. **Remove the include line** from your `<app>.conf`:

       grep -v "tinct" ~/.config/<app>/<app>.conf

2. **Delete the generated files**:

       rm ~/.config/<app>/tinct-colours.conf ~/.config/<app>/tinct.conf

3. **Reload to drop the theme**:

       <reload command>

4. **External state**: this plugin only writes to `~/.config/<app>/`. No further cleanup is required.
```

---

## Docs site build behaviour

The Docusaurus build consumes these READMEs as the canonical source for per-plugin pages. Routing is by `plugin.type`:

- `type: output` → `/docs/plugins/output/<category>/<name>`
- `type: input` → `/docs/plugins/input/<name>`

Two rules govern what makes it onto the site:

1. **Mandatory fields**:
   - **Output**: `plugin.type`, `plugin.name`, `plugin.category`, `plugin.source`, `plugin.app` must all be present and non-empty.
   - **Input**: `plugin.type`, `plugin.name`, `plugin.source`, `plugin.source_type`, `plugin.description` must all be present and non-empty.

   A README missing any mandatory field is **skipped** from the docs site with a build-time warning — the page simply doesn't appear. The runtime plugin is unaffected.

2. **Inconsistent or unparseable frontmatter**: a README whose frontmatter fails YAML parsing is **skipped** with a warning. Same outcome as above.

Skipping is preferable to shipping a half-formed page that misleads users. The warning surfaces in CI for someone to fix.

Non-mandatory fields that are missing (e.g. `generated_files`, `reload.command`, `service_url`) reduce page completeness but don't trigger a skip — the page renders with those sections omitted or sparse.

## Drift detection

`cmd/tinct-check-readmes` is the warn-only tool that diffs frontmatter against runtime behaviour. Two modes:

```bash
tinct-check-readmes builtin --root internal/plugin/output
tinct-check-readmes external --binary contrib/plugins/output/opencode/tinct-plugin-opencode
```

Both modes always exit 0 and write findings to stderr in the form:

```
WARN dunst: frontmatter default_output_dir=~/.config/dunst differs from runtime ~/.config/dunst/dunstrc.d
```

The Makefile target `make check-readmes` invokes the builtin mode across the repo. CI runs it as a non-blocking job. The output is intentionally informational; it surfaces drift without gating merges.

---

## Worked examples

- **Built-in, drop-in pattern**: [`dunst/README.md`](dunst/README.md) — writes to `dunstrc.d/`, `dunstctl reload` post-exec.
- **External, single-file**: [`../../../contrib/plugins/output/opencode/README.md`](../../../contrib/plugins/output/opencode/README.md) — Go-plugin RPC, emits the `metadata` block via `--plugin-info`.

---

## See also

- [Built-in plugins](../README.md) — the coding standard for in-tree Go plugins
- [Plugin protocols](../../../docs/docs/plugin-development/protocols.md) — the `--plugin-info` schema including the optional `metadata` block
