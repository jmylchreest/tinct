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
	tincttemplate "github.com/jmylchreest/tinct/pkg/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// Plugin implements the tinct OutputPlugin interface for Spicetify themes.
//
// Spicetify is invasive — `spicetify apply` mutates Spotify's installed
// app directory and requires the Spotify client to be closed. We deliberately
// do NOT auto-apply; instead, after writing color.ini we surface (in verbose
// mode) the exact commands the user must run themselves with Spotify closed.
type Plugin struct {
	version   string
	commit    string
	date      string
	outputDir string
}

// defaultOutputDir returns the spicetify theme directory under the user's
// home: ~/.config/spicetify/Themes/tinct. Spicetify expects color.ini to
// live inside a per-theme subdirectory under Themes/.
func defaultOutputDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "spicetify", "Themes", "tinct"), nil
}

// Generate creates a Spicetify color.ini file from the palette data.
func (p *Plugin) Generate(_ context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
	tmpl, err := p.loadTemplate(palette.Verbose)
	if err != nil {
		return nil, err
	}

	themeData := convertPaletteDataToThemeData(palette)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	outputDir := p.outputDir
	if outputDir == "" {
		def, err := defaultOutputDir()
		if err != nil {
			return nil, err
		}
		outputDir = def
	}

	// Spicetify expects Themes/<name>/ to exist with color.ini inside.
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create theme directory %q: %w", outputDir, err)
	}

	return map[string][]byte{
		filepath.Join(outputDir, "color.ini"): buf.Bytes(),
	}, nil
}

// loadTemplate loads and parses the Spicetify color.ini template.
func (p *Plugin) loadTemplate(verbose bool) (*template.Template, error) {
	loader := tincttemplate.New("spicetify", templatesFS)
	if verbose {
		loader.WithVerbose(true, &stdLogger{})
	}

	tmplContent, fromCustom, err := loader.Load("templates/color.ini.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	if verbose {
		if fromCustom {
			fmt.Fprintf(os.Stderr, "   Template source: custom (%s)\n", loader.CustomPath("templates/color.ini.tmpl"))
		} else {
			fmt.Fprintf(os.Stderr, "   Template source: embedded\n")
		}
	}

	tmpl, err := template.New("color.ini").
		Funcs(tincttemplate.TemplateFuncs()).
		Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// PreExecute verifies that the spicetify CLI is installed.
//
// Without spicetify, the theme files we'd write are useless — there's no
// way to apply them. Skip with a clear pointer to the install docs.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	if _, lookErr := exec.LookPath("spicetify"); lookErr != nil {
		return true, "spicetify CLI not installed; install from https://spicetify.app", nil
	}
	return false, "", nil
}

// PostExecute prints the manual apply instructions when verbose.
//
// We intentionally do NOT auto-run `spicetify apply`:
//   - apply patches Spotify's installed app directory (invasive).
//   - it requires the Spotify client to be closed; running it under a
//     live Spotify produces a partially-patched client.
//   - it can fail in user-visible ways that the user needs to see (e.g.
//     permission errors on the Spotify dir on first run).
//
// The user is the right actor to choose when to do this.
func (p *Plugin) PostExecute(_ context.Context, files []string) error {
	if len(files) == 0 {
		return nil
	}

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "   spicetify: theme written but NOT applied. To activate:")
	fmt.Fprintln(os.Stderr, "     1. Close Spotify completely.")
	fmt.Fprintln(os.Stderr, "     2. Run:")
	fmt.Fprintln(os.Stderr, "          spicetify config current_theme tinct color_scheme tinct")
	fmt.Fprintln(os.Stderr, "          spicetify apply")
	fmt.Fprintln(os.Stderr, "")

	return nil
}

// stdLogger implements the Logger interface for template loading.
type stdLogger struct{}

func (l *stdLogger) Printf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// GetMetadata returns plugin metadata.
//
// The Metadata block lets tinct-check-readmes diff the README frontmatter
// against runtime reality. Keep these fields in sync with what the plugin
// actually does.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "spicetify",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate Spicetify color scheme for Spotify theming",
		PluginProtocol:  string(tinctplugin.PluginTypeGoPlugin),
		Metadata: &tinctplugin.Metadata{
			// spicetify CLI is mandatory — without it the generated
			// files cannot be applied to Spotify.
			RequiredBinaries: []string{"spicetify"},
			DefaultOutputDir: "~/.config/spicetify/Themes/tinct",
			GeneratedFiles:   []string{"color.ini"},
			Pattern:          "single-file",
			Reload: &tinctplugin.ReloadMetadata{
				// No automatic reload: applying the theme requires
				// Spotify to be closed and mutates the installed app
				// directory. The user must run `spicetify apply`.
				Method:             "none",
				UserActionRequired: true,
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
			Default:     "~/.config/spicetify/Themes/tinct",
			Description: "Output directory for Spicetify theme files (must be a directory under spicetify's Themes/)",
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
		cv := colour.NewColorValue(rgba, role, cc.Index)
		colors[role] = cv
	}

	for _, cc := range palette.AllColours {
		role := colour.Role(cc.Role)
		rgba := colour.RGBA{R: cc.RGB.R, G: cc.RGB.G, B: cc.RGB.B, A: 255}
		cv := colour.NewColorValue(rgba, role, cc.Index)
		indexed = append(indexed, cv)
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
