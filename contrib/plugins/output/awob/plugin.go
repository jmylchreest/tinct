package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/jmylchreest/tinct/pkg/colour"
	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
	"github.com/jmylchreest/tinct/pkg/plugin/config"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
	tincttemplate "github.com/jmylchreest/tinct/pkg/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

const (
	// themesDirName is the awob themes subdirectory under the user's
	// awob config dir. awob-daemon reads themes from this location.
	themesDirName = "themes"

	// palettesDirName is the relative subdirectory within themes/ where
	// importable palette files live. The daemon doesn't load these
	// directly — they're picked up via `import` statements from each
	// theme's scene.kdl. Mirroring awob's stock convention.
	palettesDirName = "_palettes"

	// paletteBasename is the palette filename written under
	// themes/_palettes/. Plugin name choice — calling it tinct.kdl makes
	// the import statement self-documenting in any user-authored theme.
	paletteBasename = "tinct.kdl"

	// themeName is the directory name under themes/ that holds the
	// generated scene.kdl + manifest.toml. Users can rename this on disk
	// and tinct will overwrite the same path on the next run.
	themeName = "tinct"
)

// Plugin implements the tinct OutputPlugin interface for awob themes.
type Plugin struct {
	version   string
	commit    string
	date      string
	outputDir string
}

// resolveThemesDir picks the awob themes/ directory.
//
// Resolution order: --awob.output-dir flag → TINCT_PLUGIN_AWOB_OUTPUT_DIR
// env var → $XDG_CONFIG_HOME/awob/themes (the awob daemon default).
func (p *Plugin) resolveThemesDir() string {
	fallback := filepath.Join(paths.XDGConfigDir(), "awob", themesDirName)
	return config.Resolve("awob", "output_dir", p.outputDir, fallback)
}

// Generate produces three files:
//
//   - <themes>/_palettes/tinct.kdl  — shared palette + named-style block.
//     Other awob themes can `import "../_palettes/tinct.kdl"` to inherit
//     these colours without copying them. This is the file tinct
//     refreshes on each run.
//
//   - <themes>/tinct/scene.kdl     — a worked-example theme that imports
//     the palette and uses every awob scene primitive (icon, label,
//     track, bar, fade phases). Modelled on awob's bundled "default"
//     theme so users have a feature-complete starting point.
//
//   - <themes>/tinct/manifest.toml — the theme manifest awob-daemon
//     reads to enumerate icons, fonts, etc.
//
// Other themes (user-authored or bundled) can adopt the tinct palette
// by adding `import "../_palettes/tinct.kdl"` to their scene.kdl —
// tinct overwrites the palette in place and awob-daemon hot-reloads.
func (p *Plugin) Generate(_ context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
	themesDir := p.resolveThemesDir()
	themeData := convertPaletteDataToThemeData(palette)

	if err := os.MkdirAll(filepath.Join(themesDir, palettesDirName), 0o750); err != nil {
		return nil, fmt.Errorf("failed to create palettes dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(themesDir, themeName), 0o750); err != nil {
		return nil, fmt.Errorf("failed to create theme dir: %w", err)
	}

	paletteContent, err := renderTemplate("templates/palette.kdl.tmpl", themeData, palette.Verbose)
	if err != nil {
		return nil, fmt.Errorf("palette: %w", err)
	}
	sceneContent, err := renderTemplate("templates/scene.kdl.tmpl", themeData, palette.Verbose)
	if err != nil {
		return nil, fmt.Errorf("scene: %w", err)
	}
	manifestContent, err := renderTemplate("templates/manifest.toml.tmpl", themeData, palette.Verbose)
	if err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}

	return map[string][]byte{
		filepath.Join(themesDir, palettesDirName, paletteBasename): paletteContent,
		filepath.Join(themesDir, themeName, "scene.kdl"):           sceneContent,
		filepath.Join(themesDir, themeName, "manifest.toml"):       manifestContent,
	}, nil
}

// renderTemplate loads + parses + executes a single embedded template.
func renderTemplate(path string, themeData *colour.ThemeData, verbose bool) ([]byte, error) {
	loader := tincttemplate.New("awob", templatesFS)
	if verbose {
		loader.WithVerbose(true, &stdLogger{})
	}

	tmplContent, fromCustom, err := loader.Load(path)
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	if verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Template source: custom (%s)\n", loader.CustomPath(path))
	}

	tmpl, err := template.New(filepath.Base(path)).
		Funcs(tincttemplate.TemplateFuncs()).
		Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("execute %s: %w", path, err)
	}
	return buf.Bytes(), nil
}

// PreExecute checks if awob is installed (binary on PATH) or has a
// config dir. Either is sufficient evidence the user runs awob.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	awobConfigDir := filepath.Join(paths.XDGConfigDir(), "awob")
	if _, statErr := os.Stat(awobConfigDir); statErr == nil {
		return false, "", nil
	}
	if _, lookErr := exec.LookPath("awob-daemon"); lookErr == nil {
		return false, "", nil
	}
	if _, lookErr := exec.LookPath("awob"); lookErr == nil {
		return false, "", nil
	}
	return true, "awob not installed (no awob/awob-daemon on PATH and no ~/.config/awob)", nil
}

// PostExecute installs the freshly-written palette as awob's runtime
// force-palette overlay so the new colours apply immediately,
// regardless of which theme the daemon currently has active.
//
// `awob force-palette set <path>` is exactly the right hook here: the
// daemon overlays the palette + styles after the active theme's own
// palette resolution, so tinct's colours win on every theme without
// editing anyone's scene.kdl. The daemon also adds the overlay to its
// hot-reload watch list, so future tinct runs that rewrite the same
// file refresh in place.
//
// PostExecute runs silently. The activation guide (when the user wants
// to wire the palette via `import` instead of force-palette) lives in
// Hooks().Instructions, which the host runner gates behind --verbose.
// We can't gate per-run reporting from this method directly: external
// plugins re-spawn between Generate and PostExecute, so any verbose
// flag captured during Generate is gone here. If the awob CLI is
// missing or the daemon isn't running, force-palette is a silent no-op
// and the host's verbose-mode Instructions tell the user what to do.
func (p *Plugin) PostExecute(ctx context.Context, files []string) error {
	if len(files) == 0 {
		return nil
	}

	palettePath := findPalettePath(files)
	if palettePath == "" {
		return nil
	}

	if _, err := exec.LookPath("awob"); err != nil {
		return nil
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = exec.CommandContext(cctx, "awob", "force-palette", "set", palettePath).Run()
	// Failure is non-fatal — the daemon may simply not be running yet.
	return nil
}

// Hooks declares the static post-execute guidance — the daemon-free
// `import` instructions are the same every run, so they live here
// rather than being re-printed by PostExecute. The dynamic banner
// (force-palette success vs manual fallback) stays in PostExecute.
// Implements tinctplugin.HooksProvider (protocol 0.3.0+).
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		Instructions: "   Themes can also adopt the palette via `import` (no daemon needed): add `import \"../_palettes/tinct.kdl\"` to scene.kdl and reference $bg / $fg / $track / $accent / $low / $normal / $warn / $crit / $muted / $overflow_bg / $overflow_accent.",
	}
}

// Templates exposes awob's bundled templates so `tinct plugins
// templates list` and `tinct plugins templates dump -o awob` work
// across the RPC boundary. Implements tinctplugin.TemplateLister
// (protocol 0.3.0+).
func (p *Plugin) Templates() map[string][]byte {
	out := map[string][]byte{}
	_ = fs.WalkDir(templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, rerr := templatesFS.ReadFile(path)
		if rerr != nil {
			return nil
		}
		out[path] = b
		return nil
	})
	return out
}

// findPalettePath returns the path of the palette file from a list of
// written files (the `_palettes/tinct.kdl` member). Used by PostExecute
// to drive `awob force-palette set`.
func findPalettePath(files []string) string {
	for _, f := range files {
		if filepath.Base(f) == paletteBasename && filepath.Base(filepath.Dir(f)) == palettesDirName {
			return f
		}
	}
	return ""
}

// stdLogger implements the Logger interface for template loading.
type stdLogger struct{}

func (l *stdLogger) Printf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// GetMetadata returns plugin metadata.
//
// The optional Metadata block lets tinct-check-readmes diff the README
// frontmatter against runtime reality. Keep these fields in sync with
// what the plugin actually does — they're the machine-checkable
// counterpart to the README's `plugin:` block.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "awob",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate awob (animated Wayland Overlay Bar) palette + theme",
		PluginProtocol:  string(tinctplugin.PluginTypeGoPlugin),
		Metadata: &tinctplugin.Metadata{
			// awob detection accepts any of: awob-daemon on PATH, awob
			// CLI on PATH, or ~/.config/awob present. None is strictly
			// required at install time — declare the CLI as optional so
			// users see what enables the automatic force-palette reload.
			OptionalBinaries: []string{"awob", "awob-daemon"},
			DefaultOutputDir: "~/.config/awob/themes",
			GeneratedFiles: []string{
				"_palettes/tinct.kdl",
				"tinct/scene.kdl",
				"tinct/manifest.toml",
			},
			Pattern: "two-file",
			Reload: &tinctplugin.ReloadMetadata{
				// PostExecute runs `awob force-palette set <palette>`
				// when the awob CLI is on PATH. The daemon hot-reloads
				// the overlay, so no user action is required.
				Method:             "ipc",
				Command:            "awob force-palette set <palette>",
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
			Default:     "auto-detected",
			Description: "awob themes directory (default: ~/.config/awob/themes)",
			Required:    false,
		},
	}
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

	var themeType colour.ThemeType
	if palette.ThemeType == "dark" {
		themeType = colour.ThemeDark
	} else {
		themeType = colour.ThemeLight
	}

	helper := colour.NewPaletteHelper(themeType, colors, indexed)
	return &colour.ThemeData{
		PaletteHelper: helper,
		WallpaperPath: "",
		ThemeName:     "Tinct",
	}
}
