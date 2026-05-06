// Package warp provides an output plugin for Warp terminal colour themes.
package warp

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

// Plugin implements the output.Plugin interface for Warp.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Warp output plugin with default settings.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "warp"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Warp terminal theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "warp.output-dir", "", "Output directory (default: platform-specific Warp themes dir)")
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
		{Name: "warp.output-dir", Type: "string", Default: "", Description: "Output directory (default: platform-specific Warp themes dir)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the resolved output directory.
//
// Resolution: --warp.output-dir → TINCT_PLUGIN_WARP_OUTPUT_DIR →
// [plugin.warp] output_dir → platform default. Defaults are:
//
//   - Linux: $XDG_DATA_HOME/warp-terminal/themes (default ~/.local/share/warp-terminal/themes)
//   - macOS: $HOME/.warp/themes (Warp's documented location, not Application Support)
//   - other: $XDG_DATA_HOME/warp-terminal/themes
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("warp", "output_dir", p.outputDir, p.platformDefault())
}

func (p *Plugin) platformDefault() string {
	if runtime.GOOS == "darwin" {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".warp", "themes")
		}
	}
	return filepath.Join(paths.XDGDataDir(), "warp-terminal", "themes")
}

// Hooks declares warp's static behaviour: ensure the themes directory
// exists, no reload signal needed (Warp's file watcher picks up new
// themes within seconds), and a one-line activation hint.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		AutoCreateDir: true,
		Instructions:  "   Select 'Tinct' in Warp: Settings → Appearance → Theme. Warp picks up new themes from disk automatically.",
	}
}

// Generate creates the theme files. Returns map of filename → content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	content, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	return map[string][]byte{"tinct.yaml": content}, nil
}

func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("warp", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.yaml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.yaml.tmpl\n")
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
