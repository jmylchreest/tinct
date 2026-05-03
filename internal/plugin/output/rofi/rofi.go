// Package rofi provides an output plugin for rofi launcher themes.
package rofi

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

// Plugin implements the output.Plugin interface for rofi.
type Plugin struct {
	outputDir string
	themeName string
	verbose   bool
}

// New creates a new rofi output plugin.
func New() *Plugin { return &Plugin{themeName: "tinct"} }

// Name returns the plugin name.
func (p *Plugin) Name() string { return "rofi" }

// Description returns the plugin description.
func (p *Plugin) Description() string { return "rofi launcher theme (rasi colour variables)" }

// Version returns the plugin version.
func (p *Plugin) Version() string { return version.Version }

// RegisterFlags registers plugin-specific flags.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "rofi.output-dir", "", "Output directory (default: ~/.config/rofi/themes)")
	cmd.Flags().StringVar(&p.themeName, "rofi.theme-name", "tinct", "Theme name (filename: <name>.rasi)")
}

// SetVerbose enables or disables verbose logging.
func (p *Plugin) SetVerbose(verbose bool) { p.verbose = verbose }

// GetEmbeddedFS returns the embedded template filesystem.
func (p *Plugin) GetEmbeddedFS() any { return templates }

// GetFlagHelp returns help info for plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "rofi.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/rofi/themes)"},
		{Name: "rofi.theme-name", Type: "string", Default: "tinct", Description: "Theme name (filename: <name>.rasi)"},
	}
}

// Validate checks plugin configuration.
func (p *Plugin) Validate() error {
	if p.themeName == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	return nil
}

// DefaultOutputDir returns the resolved output directory.
// Resolution: --rofi.output-dir → TINCT_PLUGIN_ROFI_OUTPUT_DIR →
// [plugin.rofi] output_dir → platform default.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("rofi", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "rofi", "themes"))
}

// Hooks declares rofi's pre/post-execute behaviour. rofi reads the theme
// fresh on each launch, so no reload action is required.
func (p *Plugin) Hooks() hooks.Spec {
	themeName := p.themeName
	return hooks.Spec{
		RequiredBinaries: []string{"rofi"},
		AutoCreateDir:    true,
		InstructionsFn: func(_ hooks.Context) string {
			return fmt.Sprintf("   Add `@theme %q` to your rofi config, or run `rofi -theme %s` directly.", themeName, themeName)
		},
	}
}

// Generate creates the theme file.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	loader := tmplloader.New("rofi", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, _, err := loader.Load("tinct.rasi.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("rofi").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return map[string][]byte{p.themeName + ".rasi": buf.Bytes()}, nil
}
