// Package tmux provides an output plugin for tmux colour themes.
package tmux

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
func GetEmbeddedTemplates() embed.FS { return templates }

// Plugin implements the output.Plugin interface for tmux.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new tmux output plugin.
func New() *Plugin { return &Plugin{} }

// Name returns the plugin name.
func (p *Plugin) Name() string { return "tmux" }

// Description returns the plugin description.
func (p *Plugin) Description() string { return "tmux colour theme (status bar, panes, modes)" }

// Version returns the plugin version.
func (p *Plugin) Version() string { return version.Version }

// RegisterFlags registers plugin-specific flags.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "tmux.output-dir", "", "Output directory (default: ~/.config/tmux)")
}

// SetVerbose enables or disables verbose logging.
func (p *Plugin) SetVerbose(verbose bool) { p.verbose = verbose }

// GetEmbeddedFS returns the embedded template filesystem.
func (p *Plugin) GetEmbeddedFS() any { return templates }

// GetFlagHelp returns help info for plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "tmux.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/tmux)"},
	}
}

// Validate checks plugin configuration.
func (p *Plugin) Validate() error { return nil }

// DefaultOutputDir returns the resolved output directory.
// Resolution: --tmux.output-dir → TINCT_PLUGIN_TMUX_OUTPUT_DIR →
// [plugin.tmux] output_dir → platform default.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("tmux", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "tmux"))
}

// Hooks declares tmux's pre/post-execute behaviour. tmux doesn't watch
// config files, so reload requires the user to source the new file.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredBinaries: []string{"tmux"},
		AutoCreateDir:    true,
		InstructionsFn: func(hctx hooks.Context) string {
			path := filepath.Join(hctx.OutputDir, "tinct.conf")
			return fmt.Sprintf("   Add 'source-file %s' to your tmux.conf, then reload with 'tmux source-file ~/.config/tmux/tmux.conf' or prefix-r.", path)
		},
	}
}

// Generate creates the theme file.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	loader := tmplloader.New("tmux", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, _, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("tmux").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return map[string][]byte{"tinct.conf": buf.Bytes()}, nil
}
