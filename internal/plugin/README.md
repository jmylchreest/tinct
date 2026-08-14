# Built-in plugins

Built-in plugins are Go packages compiled into the `tinct` binary. They live in
`internal/plugin/input/<name>/` and `internal/plugin/output/<name>/`.

This document covers **only what is specific to built-in plugins**. Everything
else — colour roles, template functions, palette structure, CLI behaviour — is
in the [docs site](../../docs/docs/). If you are writing an *external* plugin
(a standalone executable speaking the JSON-stdio or go-plugin protocol), you
want [plugin-development](../../docs/docs/plugin-development/) instead; none of
the interfaces below apply to you.

Prefer external plugins for anything niche or independently released. Built-ins
are for broadly useful targets that should ship in the box.

## Interfaces

Output plugins implement `output.Plugin` ([plugin.go](output/plugin.go)):

```go
Name() string
Description() string
Version() string
Generate(themeData *colour.ThemeData) (map[string][]byte, error)
RegisterFlags(cmd *cobra.Command)
Validate() error
DefaultOutputDir() string
GetFlagHelp() []output.FlagHelp
```

Input plugins implement `input.Plugin` ([plugin.go](input/plugin.go)) — the same
shape, but `Generate(ctx, opts input.GenerateOptions) (*colour.Palette, error)`
and no `DefaultOutputDir`.

`Generate` returns **filename → content**, not paths. The runtime decides where
files land.

### Optional interfaces

Implement these only when you need them; the runtime type-asserts for each.

| Interface | Purpose |
|---|---|
| `output.VerbosePlugin` | Receive `--verbose` |
| `output.TemplateProvider` | Expose your `embed.FS` to the template-management commands |
| `output.VersionedPlugin` | Select templates by target app version (`templates/0.53/…`) |
| `output.DualThemePlugin` | Emit light *and* dark variants in one run |
| `output.PreExecuteHook` | Skip cleanly when the app isn't installed/configured |
| `output.PostExecuteHook` | Reload the app after files are written |
| `input.ThemeHinter` | Hint light/dark from the source |
| `input.WallpaperProvider` | Expose a wallpaper path for downstream plugins |

## Registration

New plugins must be registered in **two** places, or they will silently not
exist:

1. `internal/plugin/manager/manager.go` — the real registry
2. `cmd/tinct-check-readmes/main.go` — the README validator's plugin list

## Templates

Embed templates in the package and render through the shared loader, which
applies user overrides from `~/.config/tinct/templates/<plugin>/`:

```go
//go:embed *.tmpl
var templates embed.FS

loader := tmplloader.New("myplugin", templates)
content, fromCustom, err := loader.Load("tinct.conf.tmpl")
```

Use `utils.TemplateFuncs()` when parsing so plugins share one function set.

### Generated file naming

- App supports variables/includes → `tinct-colours.{ext}` (definitions) plus
  `tinct.{ext}` (example config that sources it)
- App does not → `tinct.{ext}` alone, colours inlined

## Flags and configuration

Namespace every flag with your plugin name: `--myplugin.output-dir`,
`--myplugin.generate-stub`. Register at minimum an `output-dir` flag, and
resolve paths through `pluginconfig.Resolve`, which applies the precedence
flag → env → TOML → fallback:

```go
func (p *Plugin) DefaultOutputDir() string {
    return pluginconfig.Resolve("myplugin", "output_dir", p.outputDir,
        filepath.Join(paths.XDGConfigDir(), "myplugin"))
}
```

## Testing

Use the shared suite in `internal/plugin/output/shared/testing` — it covers the
interface contract, generation, flags, and hooks:

```go
plugintesting.RunAllTests(t, New(), plugintesting.TestConfig{
    ExpectedName:  "myplugin",
    ExpectedFiles: []string{"tinct-colours.conf", "tinct.conf"},
})
```

**Gotcha:** the template loader prefers user overrides, so any test asserting on
rendered content fails on a machine that has customised your plugin's template.
Pin `HOME` and `XDG_CONFIG_HOME` to a temp dir in `TestMain` — see
[niri_test.go](output/niri/niri_test.go).

## Plugin README

Every plugin directory needs a `README.md`. These are not incidental: the docs
site is generated from them by `docs/scripts/sync-plugin-readmes.mjs`, and
`cmd/tinct-check-readmes` validates them.

The frontmatter schema and required body sections are specified in
[PLUGIN-README-STANDARD.md](output/PLUGIN-README-STANDARD.md).

For a new output plugin, also add its category to
[categories.json](output/categories.json) and an entry to `docs/sidebars.ts`.
