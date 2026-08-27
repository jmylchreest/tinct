// Package rosec provides an output plugin for rosec-prompt theme configuration.
package rosec

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
	"github.com/jmylchreest/tinct/internal/pluginconfig"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for rosec.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new rosec output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "rosec"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Rosec prompt theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "rosec.output-dir", "", "Output directory (default: ~/.config/rosec)")
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
		{Name: "rosec.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/rosec)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("rosec", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "rosec"))
}

// Generate creates the theme files.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}
	files["tinct-theme.toml"] = themeContent

	return files, nil
}

// generateTheme creates the rosec theme file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("rosec", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct-theme.toml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-theme.toml.tmpl\n")
	}

	tmpl, err := template.New("theme").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute is a no-op — detection is declared in Hooks().
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	return false, "", nil
}

// Hooks declares rosec's detection. rosec counts as installed if EITHER
// its config file exists OR the rosec-prompt binary is on PATH — an
// any-of group, which RequiredDirs / RequiredBinaries cannot express
// because both are all-of.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredAny: []hooks.AnyOf{{
			Binaries: []string{"rosec-prompt"},
			Dirs:     []string{filepath.Join(p.DefaultOutputDir(), "config.toml")},
			Reason: "rosec does not appear to be installed.\n" +
				"Neither ~/.config/rosec/config.toml nor the rosec-prompt binary was found.\n" +
				"Install rosec-prompt, or create the config, to enable this plugin.",
		}},
	}
}
