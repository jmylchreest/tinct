package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/jmylchreest/tinct/pkg/colour"
	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
	"github.com/jmylchreest/tinct/pkg/plugin/config"
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

// PostExecute prints follow-up guidance.
func (p *Plugin) PostExecute(_ context.Context, files []string) error {
	if len(files) == 0 {
		return nil
	}
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "  awob theme installed\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Wrote %d files. Activate with one of:\n", len(files))
	fmt.Fprintf(os.Stderr, "  awob-daemon --theme tinct\n")
	fmt.Fprintf(os.Stderr, "  # or set 'theme = \"tinct\"' in ~/.config/awob/awob.toml\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Already-installed themes can adopt the tinct palette by\n")
	fmt.Fprintf(os.Stderr, "adding `import \"../_palettes/tinct.kdl\"` to their scene.kdl\n")
	fmt.Fprintf(os.Stderr, "and referencing $bg / $fg / $track / $accent / $low / $normal\n")
	fmt.Fprintf(os.Stderr, "/ $warn / $crit / $muted. The palette refreshes in place on\n")
	fmt.Fprintf(os.Stderr, "every tinct generate; awob-daemon hot-reloads.\n")
	fmt.Fprintf(os.Stderr, "\n")
	return nil
}

// stdLogger implements the Logger interface for template loading.
type stdLogger struct{}

func (l *stdLogger) Printf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// GetMetadata returns plugin metadata.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "awob",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate awob (animated Wayland Overlay Bar) palette + theme",
		PluginProtocol:  string(tinctplugin.PluginTypeGoPlugin),
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
