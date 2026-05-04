// Package yazi provides an output plugin for Yazi file manager themes (flavor pack).
package yazi

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

// Plugin implements the output.Plugin interface for Yazi.
//
// Yazi consumes themes as "flavor packs" — directories named
// <name>.yazi/ under ~/.config/yazi/flavors/ containing flavor.toml.
// We write flavor.toml into that directory; the user activates it via
// `[flavor] dark = "<name>"` in their theme.toml.
type Plugin struct {
	outputDir string
	themeName string
	verbose   bool
}

// New creates a new Yazi output plugin.
func New() *Plugin { return &Plugin{themeName: "tinct"} }

// Name returns the plugin name.
func (p *Plugin) Name() string { return "yazi" }

// Description returns the plugin description.
func (p *Plugin) Description() string { return "Yazi file manager flavor pack (flavor.toml)" }

// Version returns the plugin version.
func (p *Plugin) Version() string { return version.Version }

// RegisterFlags registers plugin-specific flags.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "yazi.output-dir", "", "Output directory (default: ~/.config/yazi/flavors/tinct.yazi)")
	cmd.Flags().StringVar(&p.themeName, "yazi.theme-name", "tinct", "Flavor name (output directory becomes <name>.yazi)")
}

// SetVerbose enables or disables verbose logging.
func (p *Plugin) SetVerbose(verbose bool) { p.verbose = verbose }

// GetEmbeddedFS returns the embedded template filesystem.
func (p *Plugin) GetEmbeddedFS() any { return templates }

// GetFlagHelp returns help info for plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "yazi.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/yazi/flavors/tinct.yazi)"},
		{Name: "yazi.theme-name", Type: "string", Default: "tinct", Description: "Flavor name (output directory becomes <name>.yazi)"},
	}
}

// Validate checks plugin configuration.
func (p *Plugin) Validate() error {
	if p.themeName == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	return nil
}

// DefaultOutputDir returns the resolved output directory. The flavor
// pack lives at ~/.config/yazi/flavors/<name>.yazi.
//
// Resolution: --yazi.output-dir → TINCT_PLUGIN_YAZI_OUTPUT_DIR →
// [plugin.yazi] output_dir → platform default.
func (p *Plugin) DefaultOutputDir() string {
	fallback := filepath.Join(paths.XDGConfigDir(), "yazi", "flavors", p.themeName+".yazi")
	return pluginconfig.Resolve("yazi", "output_dir", p.outputDir, fallback)
}

// Hooks declares Yazi's pre/post-execute behaviour. Yazi rereads the
// flavor pack on next launch; running instances need restarting.
func (p *Plugin) Hooks() hooks.Spec {
	themeName := p.themeName
	return hooks.Spec{
		RequiredBinaries: []string{"yazi"},
		AutoCreateDir:    true,
		InstructionsFn: func(_ hooks.Context) string {
			return fmt.Sprintf("   Add `[flavor]\\ndark = %q` (and/or `light = %q`) to ~/.config/yazi/theme.toml. Restart yazi to apply.", themeName, themeName)
		},
	}
}

// Generate creates the flavor.toml file.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	loader := tmplloader.New("yazi", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, _, err := loader.Load("flavor.toml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("yazi").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return map[string][]byte{"flavor.toml": buf.Bytes()}, nil
}
