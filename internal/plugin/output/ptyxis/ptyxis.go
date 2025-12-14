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
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
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
	cmd.Flags().StringVar(&p.outputDir, "ptyxis.output-dir", "", "Output directory (default: ~/.local/share/ptyxis/palettes)")
}

// SetVerbose enables or disables verbose logging for the plugin.
// Implements the output.VerbosePlugin interface.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetEmbeddedFS returns the embedded template filesystem.
// Implements the output.TemplateProvider interface.
func (p *Plugin) GetEmbeddedFS() any {
	return templates
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "ptyxis.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.local/share/ptyxis/palettes)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".local/share/ptyxis/palettes"
	}
	return filepath.Join(home, ".local", "share", "ptyxis", "palettes")
}

// PreExecute checks if ptyxis is installed before generating the palette.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return true, "cannot determine home directory", nil
	}

	ptyxisDataDir := filepath.Join(home, ".local", "share", "org.gnome.Ptyxis")

	// Check for ptyxis binary (native/PATH) and common Flatpak/config locations
	if !appdetect.IsPresentAny([]string{"ptyxis"}, []string{ptyxisDataDir}) {
		return true, "Ptyxis terminal is not installed", nil
	}
	return false, "", nil
}

// Generate creates the Ptyxis palette file.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate the palette file
	paletteContent, err := p.generatePalette(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}

	files["tinct.palette"] = paletteContent

	return files, nil
}

// generatePalette creates the Ptyxis .palette file content.
func (p *Plugin) generatePalette(themeData *colour.ThemeData) ([]byte, error) {
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

// PostExecute provides instructions for installing the palette in Ptyxis.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(_ context.Context, execCtx output.ExecutionContext, writtenFiles []string) error {
	if execCtx.DryRun || !p.verbose {
		return nil
	}

	// Only show instructions if we actually wrote a palette file
	if len(writtenFiles) == 0 {
		return nil
	}

	// Determine installation type to provide appropriate instructions
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	ptyxisDataDir := filepath.Join(home, ".local", "share", "org.gnome.Ptyxis")
	installType := appdetect.GetInstallationType([]string{"ptyxis"}, []string{ptyxisDataDir})

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "  Ptyxis Palette Installation\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "\n")

	if installType == "flatpak" {
		fmt.Fprintf(os.Stderr, "Detected Flatpak installation of Ptyxis.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "To install the palette:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  Method 1 (Ptyxis 49+):\n")
		fmt.Fprintf(os.Stderr, "    flatpak run app.devsuite.Ptyxis --import-palette %s\n", writtenFiles[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  Method 2 (All versions):\n")
		fmt.Fprintf(os.Stderr, "    Drag and drop the palette file into Ptyxis\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Then select the 'Tinct' palette in Ptyxis preferences.\n")
	} else {
		fmt.Fprintf(os.Stderr, "To install the palette:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  Method 1 (Ptyxis 49+):\n")
		fmt.Fprintf(os.Stderr, "    ptyxis --import-palette %s\n", writtenFiles[0])
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  Method 2 (All versions):\n")
		fmt.Fprintf(os.Stderr, "    Drag and drop the palette file into Ptyxis\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Then select the 'Tinct' palette in Ptyxis preferences:\n")
		fmt.Fprintf(os.Stderr, "  Preferences → Appearance → Palette → Tinct\n")
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "The palette includes window dressing support:\n")
	fmt.Fprintf(os.Stderr, "  • Bell notifications use warning color\n")
	fmt.Fprintf(os.Stderr, "  • Remote sessions use info color\n")
	fmt.Fprintf(os.Stderr, "  • Superuser sessions use danger color\n")
	fmt.Fprintf(os.Stderr, "\n")

	return nil
}
