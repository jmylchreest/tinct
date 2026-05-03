// Package btop provides an output plugin for btop resource monitor themes.
package btop

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
	"github.com/jmylchreest/tinct/internal/pluginconfig"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for btop.
type Plugin struct {
	outputDir string
	themeName string
	verbose   bool
}

// New creates a new btop output plugin with default settings.
func New() *Plugin {
	return &Plugin{themeName: "tinct"}
}

// Name returns the plugin name.
func (p *Plugin) Name() string { return "btop" }

// Description returns the plugin description.
func (p *Plugin) Description() string { return "btop resource monitor theme" }

// Version returns the plugin version.
func (p *Plugin) Version() string { return version.Version }

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "btop.output-dir", "", "Output directory (default: ~/.config/btop/themes)")
	cmd.Flags().StringVar(&p.themeName, "btop.theme-name", "tinct", "Theme name (filename: <name>.theme)")
}

// SetVerbose enables or disables verbose logging.
func (p *Plugin) SetVerbose(verbose bool) { p.verbose = verbose }

// GetEmbeddedFS returns the embedded template filesystem.
func (p *Plugin) GetEmbeddedFS() any { return templates }

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "btop.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/btop/themes)"},
		{Name: "btop.theme-name", Type: "string", Default: "tinct", Description: "Theme name (filename: <name>.theme)"},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	if p.themeName == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	return nil
}

// DefaultOutputDir returns the resolved output directory.
// Resolution: --btop.output-dir → TINCT_PLUGIN_BTOP_OUTPUT_DIR →
// [plugin.btop] output_dir → platform default.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("btop", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "btop", "themes"))
}

// Hooks declares btop's pre/post-execute behaviour. btop reloads themes
// only when the user picks one from its menu, so we just print the
// follow-up steps.
func (p *Plugin) Hooks() hooks.Spec {
	themeName := p.themeName
	return hooks.Spec{
		RequiredBinaries: []string{"btop"},
		AutoCreateDir:    true,
		InstructionsFn: func(_ hooks.Context) string {
			return fmt.Sprintf("   Set 'color_theme = %q' in ~/.config/btop/btop.conf, or pick %s from btop's options menu (press 'p').", themeName, themeName)
		},
	}
}

// Generate creates the theme file.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	loader := tmplloader.New("btop", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, _, err := loader.Load("tinct.theme.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	tmpl, err := template.New("btop").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	return map[string][]byte{p.themeName + ".theme": buf.Bytes()}, nil
}
