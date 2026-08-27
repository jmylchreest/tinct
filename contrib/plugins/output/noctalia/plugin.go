package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jmylchreest/tinct/pkg/colour"
	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	tincttemplate "github.com/jmylchreest/tinct/pkg/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// paletteName is the Noctalia custom-palette name; the file is written as
// <paletteName>.json and selected in config via `custom_palette = "tinct"`.
const paletteName = "tinct"

// Plugin implements the tinct OutputPlugin interface for the Noctalia shell.
//
// Noctalia (v5) ships its own Material You generator, but reads a fixed
// palette file when `[theme] source = "custom"` is set. This plugin writes
// that palette file so Noctalia's colours match the rest of a tinct-themed
// desktop. The schema is a JSON object with `dark` and `light` variants,
// each carrying 16 Material-3 roles plus a terminal colour block.
type Plugin struct {
	version   string
	commit    string
	date      string
	outputDir string
}

// noctaliaConfigDir resolves Noctalia's config directory, honouring
// NOCTALIA_CONFIG_HOME then XDG_CONFIG_HOME before falling back to
// ~/.config/noctalia. It returns the path whether or not it exists —
// existence is the hook runner's business (see Hooks) — or "" when even
// the home directory cannot be determined.
//
// This stays a local helper rather than a hooks.Spec literal because
// RequiredDirs paths are only ~-expanded by the runner; it has no
// env-var expansion, so the XDG chain has to be resolved here.
func noctaliaConfigDir() string {
	if c := os.Getenv("NOCTALIA_CONFIG_HOME"); c != "" {
		return c
	}
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "noctalia")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "noctalia")
}

// Configure records the host's per-run configuration. It is what makes
// --noctalia.output-dir load-bearing: the flag arrives here before
// PreExecute and Hooks run, so the declared RequiredDirs can follow the
// override instead of always pointing at the default config directory.
//
// Implements tinctplugin.Configurable (protocol 0.3.0+).
func (p *Plugin) Configure(req tinctplugin.ConfigureRequest) error {
	if dir, ok := req.Args["output-dir"].(string); ok && dir != "" {
		p.outputDir = expandTilde(dir)
	}
	return nil
}

// expandTilde expands a leading ~ so a flag value can be written in the
// shell-style form users expect. The host does not expand it for us —
// plugin args arrive as literal strings.
func expandTilde(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}

// resolveOutputDir returns the palettes directory to write into.
func (p *Plugin) resolveOutputDir() (string, error) {
	if p.outputDir != "" {
		return p.outputDir, nil
	}
	dir := noctaliaConfigDir()
	if dir == "" {
		return "", fmt.Errorf("failed to resolve noctalia config directory")
	}
	return filepath.Join(dir, "palettes"), nil
}

// Generate writes the Noctalia custom palette file from the theme data.
func (p *Plugin) Generate(_ context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
	tmpl, err := p.loadTemplate(palette.Verbose)
	if err != nil {
		return nil, err
	}

	// Render the primary variant.
	primaryObj, err := renderVariant(tmpl, convertPaletteDataToThemeData(palette))
	if err != nil {
		return nil, fmt.Errorf("failed to render primary variant: %w", err)
	}

	// Assign variants to Noctalia's fixed dark/light keys regardless of
	// which one is primary.
	var darkObj, lightObj map[string]any
	if palette.ThemeType == "light" {
		lightObj = primaryObj
	} else {
		darkObj = primaryObj
	}

	if palette.AlternateTheme != nil {
		if palette.Verbose {
			fmt.Fprintf(os.Stderr, "   Generating dual-theme (dark + light)\n")
		}
		altPalette := tinctplugin.PaletteData{
			Colours:    palette.AlternateTheme.Colours,
			AllColours: palette.AlternateTheme.AllColours,
			ThemeType:  palette.AlternateTheme.ThemeType,
			Verbose:    palette.Verbose,
		}
		altObj, err := renderVariant(tmpl, convertPaletteDataToThemeData(altPalette))
		if err != nil {
			return nil, fmt.Errorf("failed to render alternate variant: %w", err)
		}
		if palette.AlternateTheme.ThemeType == "light" {
			lightObj = altObj
		} else {
			darkObj = altObj
		}
	}

	// Noctalia requires `dark`; `light` is optional and falls back to dark.
	// If only a light variant exists, use it for dark too so the file is valid.
	if darkObj == nil {
		darkObj = lightObj
	}

	out := map[string]any{"dark": darkObj}
	if lightObj != nil {
		out["light"] = lightObj
	}

	content, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal palette: %w", err)
	}
	content = append(content, '\n')

	outputDir, err := p.resolveOutputDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create palettes directory %q: %w", outputDir, err)
	}

	files := map[string][]byte{
		filepath.Join(outputDir, paletteName+".json"): content,
	}
	return files, nil
}

// loadTemplate loads and parses the palette variant template.
func (p *Plugin) loadTemplate(verbose bool) (*template.Template, error) {
	loader := tincttemplate.New("noctalia", templatesFS)
	if verbose {
		loader.WithVerbose(true, &stdLogger{})
	}

	tmplContent, fromCustom, err := loader.Load("templates/palette.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}
	if verbose {
		if fromCustom {
			fmt.Fprintf(os.Stderr, "   Template source: custom (%s)\n", loader.CustomPath("templates/palette.json.tmpl"))
		} else {
			fmt.Fprintf(os.Stderr, "   Template source: embedded\n")
		}
	}

	tmpl, err := template.New("palette").Funcs(tincttemplate.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return tmpl, nil
}

// renderVariant executes the template for one theme variant and returns the
// parsed JSON object (the 16 Material-3 roles + terminal block).
func renderVariant(tmpl *template.Template, themeData *colour.ThemeData) (map[string]any, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(buf.Bytes(), &obj); err != nil {
		return nil, fmt.Errorf("template produced invalid JSON: %w", err)
	}
	return obj, nil
}

// PreExecute is a no-op — the "is Noctalia installed?" check is declared
// in Hooks() as a RequiredDirs entry, which the host's shared runner
// evaluates before this method is called.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	return false, "", nil
}

// PostExecute is a no-op — the reload is declared statically in Hooks()
// so the host's shared runner owns the binary check and the timeout.
func (p *Plugin) PostExecute(_ context.Context, _ []string) error {
	return nil
}

// Hooks declares Noctalia's static post-execute behaviour.
//
// Noctalia v5 does NOT pick up palette changes on its own. It inotifies
// ~/.config/noctalia (and ~/.local/state/noctalia) non-recursively and
// only treats a changed *.toml as a config change. The palette misses on
// both counts: palettes/ is a subdirectory, and tinct.json is not TOML.
// So a regenerated palette produces no event and the running shell keeps
// the colours it resolved at startup.
//
// `noctalia msg config-reload` forces a config reload, which re-runs the
// theme resolution and — for a custom palette source — re-reads the JSON
// from disk. That is the whole fix, so it is the reload verb here.
//
// The binary is optional rather than required: the plugin detects
// Noctalia by config directory (see PreExecute), which lets users
// generate into a config tree for a machine where the binary is absent.
// The host runner warns about the missing binary in verbose mode and the
// exec verb is a non-fatal no-op when it cannot be found.
//
// `color-scheme-set custom <name>` also re-reads the file, but persists a
// theme override into ~/.local/state/noctalia/settings.toml that then
// shadows the user's own config — config-reload has no such side effect.
//
// Detection is by config directory, declared as RequiredDirs so the
// shared runner skips the plugin (with a consistent reason string) when
// Noctalia is not installed. The path is resolved here rather than
// written as a "~/.config/noctalia" literal because the runner does not
// expand environment variables — see noctaliaConfigDir.
//
// Implements tinctplugin.HooksProvider (protocol 0.3.0+).
func (p *Plugin) Hooks() hooks.Spec {
	spec := hooks.Spec{
		OptionalBinaries: []string{"noctalia"},
		Reload: &hooks.ReloadSpec{
			Verb: hooks.VerbExec,
			Args: []string{"noctalia", "msg", "config-reload"},
		},
	}
	// With an explicit --noctalia.output-dir the user has said where to
	// write, so the config-directory check would be wrong — that is the
	// documented way to generate for a machine that has no Noctalia
	// install of its own.
	if p.outputDir != "" {
		return spec
	}

	// An unresolvable config dir is left to Generate, which fails with a
	// specific error; declaring RequiredDirs{""} would skip the plugin
	// with a truncated, confusing reason instead.
	if dir := noctaliaConfigDir(); dir != "" {
		spec.RequiredDirs = []string{dir}
	}
	return spec
}

// Templates exposes the embedded template so `tinct plugins templates dump`
// works for this external plugin.
func (p *Plugin) Templates() map[string][]byte {
	out := map[string][]byte{}
	_ = fs.WalkDir(templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, readErr := templatesFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		out[path] = b
		return nil
	})
	return out
}

// GetMetadata returns plugin metadata.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "noctalia",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate a custom colour palette for the Noctalia shell",
		PluginProtocol:  string(tinctplugin.PluginTypeGoPlugin),
		Metadata: &tinctplugin.Metadata{
			// Noctalia detection is by config-directory presence
			// (~/.config/noctalia); no binary on PATH is required.
			OptionalBinaries: []string{"noctalia"},
			DefaultOutputDir: "~/.config/noctalia/palettes",
			GeneratedFiles:   []string{"tinct.json"},
			Pattern:          "single-file",
			Reload: &tinctplugin.ReloadMetadata{
				// Noctalia does not watch the palettes directory, so the
				// declarative hook in Hooks() nudges the running shell to
				// re-resolve the palette from disk.
				Method:             "ipc",
				Command:            "noctalia msg config-reload",
				UserActionRequired: false,
			},
		},
	}
}

// GetFlagHelp returns help for plugin-specific flags.
func (p *Plugin) GetFlagHelp() []tinctplugin.FlagHelp {
	return []tinctplugin.FlagHelp{
		{
			Name:        "output-dir",
			Shorthand:   "o",
			Type:        "string",
			Default:     "~/.config/noctalia/palettes",
			Description: "Output directory for the Noctalia palette file",
			Required:    false,
		},
	}
}

// stdLogger implements the Logger interface for template loading.
type stdLogger struct{}

func (l *stdLogger) Printf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// convertPaletteDataToThemeData converts the protocol PaletteData to colour.ThemeData.
func convertPaletteDataToThemeData(palette tinctplugin.PaletteData) *colour.ThemeData {
	colors := make(map[colour.Role]colour.ColorValue)
	indexed := make([]colour.ColorValue, 0, len(palette.AllColours))

	for roleName, cc := range palette.Colours {
		role := colour.Role(roleName)
		rgba := colour.RGBA{R: cc.RGB.R, G: cc.RGB.G, B: cc.RGB.B, A: 255}
		colors[role] = colour.NewColorValue(rgba, role, cc.Index)
	}
	for _, cc := range palette.AllColours {
		role := colour.Role(cc.Role)
		rgba := colour.RGBA{R: cc.RGB.R, G: cc.RGB.G, B: cc.RGB.B, A: 255}
		indexed = append(indexed, colour.NewColorValue(rgba, role, cc.Index))
	}

	themeType := colour.ThemeLight
	if palette.ThemeType == "dark" {
		themeType = colour.ThemeDark
	}

	helper := colour.NewPaletteHelper(themeType, colors, indexed)
	return &colour.ThemeData{
		PaletteHelper: helper,
		WallpaperPath: "",
		ThemeName:     "Tinct",
	}
}
