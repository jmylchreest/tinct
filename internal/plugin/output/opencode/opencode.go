// Package opencode provides an output plugin for Opencode CLI agent themes.
package opencode

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

// Plugin implements the output.Plugin interface for Opencode.
type Plugin struct {
	outputDir string
	themeName string
	verbose   bool
}

// New creates a new Opencode output plugin.
func New() *Plugin { return &Plugin{themeName: "tinct"} }

// Name returns the plugin name.
func (p *Plugin) Name() string { return "opencode" }

// Description returns the plugin description.
func (p *Plugin) Description() string { return "Opencode CLI agent theme (JSON)" }

// Version returns the plugin version.
func (p *Plugin) Version() string { return version.Version }

// RegisterFlags registers plugin-specific flags.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "opencode.output-dir", "", "Output directory (default: ~/.config/opencode/themes)")
	cmd.Flags().StringVar(&p.themeName, "opencode.theme-name", "tinct", "Theme name (filename: <name>.json)")
}

// SetVerbose enables or disables verbose logging.
func (p *Plugin) SetVerbose(verbose bool) { p.verbose = verbose }

// GetEmbeddedFS returns the embedded template filesystem.
func (p *Plugin) GetEmbeddedFS() any { return templates }

// GetFlagHelp returns help info for plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "opencode.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/opencode/themes)"},
		{Name: "opencode.theme-name", Type: "string", Default: "tinct", Description: "Theme name (filename: <name>.json)"},
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
// Resolution: --opencode.output-dir → TINCT_PLUGIN_OPENCODE_OUTPUT_DIR →
// [plugin.opencode] output_dir → platform default.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("opencode", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "opencode", "themes"))
}

// Hooks declares Opencode's pre/post-execute behaviour. Opencode reads
// the theme file fresh on each invocation, so no reload is needed.
//
// The opencode CLI is distributed under multiple binary names and via
// curl-installer scripts that don't always land on $PATH; we treat the
// binary as optional rather than required so users who write themes
// before installing the agent aren't blocked.
func (p *Plugin) Hooks() hooks.Spec {
	themeName := p.themeName
	return hooks.Spec{
		OptionalBinaries: []string{"opencode"},
		AutoCreateDir:    true,
		InstructionsFn: func(_ hooks.Context) string {
			return fmt.Sprintf("   Set 'theme: %q' in your opencode config (~/.config/opencode/opencode.json or opencode.jsonc).", themeName)
		},
	}
}

// Generate creates the theme file.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	loader := tmplloader.New("opencode", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, _, err := loader.Load("tinct.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("opencode").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return map[string][]byte{p.themeName + ".json": buf.Bytes()}, nil
}
