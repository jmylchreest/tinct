// Package ptyxis provides an output plugin for Ptyxis terminal color palettes.
package ptyxis

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/util/appdetect"
)

//go:embed *.tmpl
var templates embed.FS

const (
	// ptyxisAppID matches APP_ID in Ptyxis's meson build — the GNOME-style
	// reverse-DNS application id used as the user-data subdirectory name.
	// Ref: ptyxis_get_user_palettes_dir() in src/ptyxis-palette.c.
	ptyxisAppID = "org.gnome.Ptyxis"

	// ptyxisFlatpakID is the Flathub application id for Ptyxis. Inside the
	// sandbox XDG_DATA_HOME maps to ~/.var/app/<flatpak-id>/data, so the
	// host-visible palette dir is ~/.var/app/<flatpak-id>/data/<app-id>/palettes.
	ptyxisFlatpakID = "app.devsuite.Ptyxis"

	// paletteFilename is the filename written to the Ptyxis palette dir. The
	// basename (minus .palette) becomes the palette id Ptyxis exposes.
	paletteFilename = "tinct.palette"
)

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Ptyxis terminal.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Ptyxis output plugin with default settings.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "ptyxis"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Ptyxis terminal color palette"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "ptyxis.output-dir", "",
		"Output directory (default: auto-detect native or Flatpak palette dir)")
}

// SetVerbose enables or disables verbose logging for the plugin.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetEmbeddedFS returns the embedded template filesystem.
func (p *Plugin) GetEmbeddedFS() any {
	return templates
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{
			Name:        "ptyxis.output-dir",
			Type:        "string",
			Default:     "",
			Description: "Output directory (default: auto-detect native or Flatpak palette dir)",
			Required:    false,
		},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// nativeDataDir returns Ptyxis's user data dir for a native install:
// $XDG_DATA_HOME/org.gnome.Ptyxis (default $XDG_DATA_HOME = ~/.local/share).
func nativeDataDir(home string) string {
	xdg := os.Getenv("XDG_DATA_HOME")
	if xdg == "" {
		xdg = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(xdg, ptyxisAppID)
}

// nativePalettesDir returns the palettes subdirectory for a native install.
func nativePalettesDir(home string) string {
	return filepath.Join(nativeDataDir(home), "palettes")
}

// flatpakDataDir returns Ptyxis's user data dir as visible from the host
// when Ptyxis is installed via Flatpak.
func flatpakDataDir(home string) string {
	return filepath.Join(home, ".var", "app", ptyxisFlatpakID, "data", ptyxisAppID)
}

// flatpakPalettesDir returns the palettes subdirectory inside the Flatpak
// per-app data dir.
func flatpakPalettesDir(home string) string {
	return filepath.Join(flatpakDataDir(home), "palettes")
}

// DefaultOutputDir returns the default output directory for this plugin.
// Picks the Flatpak host-visible path if Ptyxis is only installed via Flatpak,
// otherwise falls back to the native XDG_DATA_HOME path.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".local", "share", ptyxisAppID, "palettes")
	}
	if appdetect.GetInstallationType([]string{"ptyxis"}, nil) == "flatpak" {
		return flatpakPalettesDir(home)
	}
	return nativePalettesDir(home)
}

// PreExecute checks if ptyxis is installed before generating the palette.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return true, "cannot determine home directory", nil
	}
	if !appdetect.IsPresentAny([]string{"ptyxis"}, []string{nativeDataDir(home), flatpakDataDir(home)}) {
		return true, "Ptyxis terminal is not installed", nil
	}
	return false, "", nil
}

// Generate creates a single-mode Ptyxis palette file. All colour keys are
// written directly under [Palette]; Ptyxis derives whether the palette is
// light or dark from the background luminance (is_dark() in ptyxis-palette.c).
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}
	face, err := p.renderFace(themeData)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("# Ptyxis colour palette generated by Tinct\n")
	buf.WriteString("[Palette]\n")
	buf.WriteString("Name=Tinct\n")
	buf.Write(face)

	return map[string][]byte{paletteFilename: buf.Bytes()}, nil
}

// GenerateDualTheme creates a dual-mode Ptyxis palette file with separate
// [Light] and [Dark] sections holding the actual light/dark variants. Ptyxis
// switches between them automatically based on the system colour preference.
// Implements output.DualThemePlugin.
func (p *Plugin) GenerateDualTheme(primaryTheme, alternateTheme *colour.ThemeData) (map[string][]byte, error) {
	if primaryTheme == nil {
		return nil, fmt.Errorf("primary theme data cannot be nil")
	}
	if alternateTheme == nil {
		return nil, fmt.Errorf("alternate theme data cannot be nil")
	}

	light, dark := splitByMode(primaryTheme, alternateTheme)

	lightFace, err := p.renderFace(light)
	if err != nil {
		return nil, fmt.Errorf("failed to render light face: %w", err)
	}
	darkFace, err := p.renderFace(dark)
	if err != nil {
		return nil, fmt.Errorf("failed to render dark face: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("# Ptyxis colour palette generated by Tinct\n")
	buf.WriteString("[Palette]\n")
	buf.WriteString("Name=Tinct\n\n")
	buf.WriteString("[Light]\n")
	buf.Write(lightFace)
	buf.WriteString("\n[Dark]\n")
	buf.Write(darkFace)

	return map[string][]byte{paletteFilename: buf.Bytes()}, nil
}

// splitByMode returns (light, dark) given two themes of opposite mode. If both
// themes report the same mode (shouldn't happen via the dual-theme pipeline),
// the first argument is treated as light.
func splitByMode(a, b *colour.ThemeData) (light, dark *colour.ThemeData) {
	if a.ThemeType() == colour.ThemeLight {
		return a, b
	}
	return b, a
}

// renderFace renders the face body (the colour keys for one variant) using
// the embedded template. The output has no [Palette]/[Light]/[Dark] header;
// the caller assembles the surrounding sections.
func (p *Plugin) renderFace(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("ptyxis", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	tmplContent, fromCustom, err := loader.Load("tinct.palette.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.palette.tmpl\n")
	}

	tmpl, err := template.New("palette").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// PostExecute prints installation guidance after writing the palette. Because
// the file is written directly into Ptyxis's watched palette dir,
// PtyxisUserPalettes picks it up live via GFileMonitor — no import step.
func (p *Plugin) PostExecute(_ context.Context, execCtx output.ExecutionContext, writtenFiles []string) error {
	if execCtx.DryRun || !p.verbose || len(writtenFiles) == 0 {
		return nil
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "  Ptyxis Palette Installation\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Wrote palette to: %s\n", writtenFiles[0])
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Ptyxis watches its palette directory and picks up new\n")
	fmt.Fprintf(os.Stderr, "palettes immediately. Select 'Tinct' in:\n")
	fmt.Fprintf(os.Stderr, "  Preferences → Appearance → Palette → Tinct\n")
	fmt.Fprintf(os.Stderr, "\n")

	return nil
}
