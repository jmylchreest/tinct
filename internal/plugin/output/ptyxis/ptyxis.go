// Package ptyxis provides an output plugin for Ptyxis terminal color palettes.
package ptyxis

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
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
	return "0.0.1"
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
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
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
	_, err = exec.LookPath("ptyxis")
	if err != nil {
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
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}

	tmplContent, fromCustom, err := loader.Load("tinct.palette.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.palette.tmpl\n")
	}

	tmpl, err := template.New("palette").Funcs(common.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}
