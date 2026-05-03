// Package walker provides an output plugin for Walker application launcher colour themes.
package walker

import (
	"bytes"
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
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Walker.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Walker output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "walker"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Walker application launcher theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "walker.output-dir", "", "Output directory (default: ~/.config/walker/themes)")
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
		{Name: "walker.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/walker/themes)", Required: false},
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
		return ".config/walker/themes"
	}
	return filepath.Join(home, ".config", "walker", "themes")
}

// Generate creates the theme files.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate tinct-colours.css file - goes to themes/tinct-colours.css.
	colorsContent, err := p.generateColors(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate colors: %w", err)
	}
	files["tinct-colours.css"] = colorsContent

	// Generate tinct.css file - goes to themes/tinct/style.css.
	styleContent, err := p.generateStyle(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate style: %w", err)
	}
	files["tinct/style.css"] = styleContent

	return files, nil
}

// generateColors creates the tinct-colours.css file with color definitions.
func (p *Plugin) generateColors(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("walker", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colours.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read colors template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colours.css.tmpl\n")
	}

	tmpl, err := template.New("colors").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse colors template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute colors template: %w", err)
	}

	return buf.Bytes(), nil
}

// generateStyle creates the tinct.css file that imports tinct-colours.css.
func (p *Plugin) generateStyle(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("walker", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read style template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.css.tmpl\n")
	}

	tmpl, err := template.New("style").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse style template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute style template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute checks if walker is available and config directory exists.
// Implements the output.PreExecuteHook interface.
// Hooks declares walker's pre-execute behaviour. Walker has no
// hot-reload mechanism so the spec is binary check + dir create only.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredBinaries: []string{"walker"},
		AutoCreateDir:    true,
	}
}
