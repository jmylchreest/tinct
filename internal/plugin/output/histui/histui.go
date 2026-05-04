// Package histui provides an output plugin for histui notification daemon CSS themes.
package histui

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
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/pluginconfig"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for histui.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new histui output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "histui"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "histui notification daemon CSS colours"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "histui.output-dir", "", "Output directory (default: ~/.config/histui/themes)")
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
		{Name: "histui.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/histui/themes)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("histui", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "histui", "themes"))
}

// Hooks declares the routine bits of histui's lifecycle. The "binary OR
// config-dir present" gate stays imperative below — it's an OR check the
// data-only spec can't represent.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		AutoCreateDir: true,
		Instructions: `   tinct-colors.css written to histui themes directory
   Touch your theme.css or restart histuid to apply changes`,
	}
}

// Generate creates the theme file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate tinct-colors.css file.
	colorsContent, err := p.generateColors(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate colors: %w", err)
	}

	files["tinct-colors.css"] = colorsContent

	return files, nil
}

// generateColors creates the CSS colours file.
func (p *Plugin) generateColors(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("histui", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colors.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read colors template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colors.css.tmpl\n")
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

// PreExecute keeps histui's bespoke "histuid on PATH OR config dir
// present" gate. The data-only spec can't express OR semantics, so this
// stays imperative; the spec covers AutoCreateDir for the themes
// subdirectory and the post-execute message.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	histuiConfigDir := filepath.Join(paths.XDGConfigDir(), "histui")
	if _, lookErr := exec.LookPath("histuid"); lookErr != nil {
		if _, statErr := os.Stat(histuiConfigDir); os.IsNotExist(statErr) {
			return true, "histuid executable not found on $PATH and " + histuiConfigDir + " does not exist", nil
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Note: histuid not found on $PATH but %s exists - continuing\n", histuiConfigDir)
		}
	}
	return false, "", nil
}
