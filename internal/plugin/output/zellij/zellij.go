// Package zellij provides an output plugin for Zellij terminal multiplexer colour themes.
package zellij

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

// Plugin implements the output.Plugin interface for Zellij.
type Plugin struct {
	outputDir string
	themeName string
	verbose   bool
}

// New creates a new Zellij output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		themeName: "tinct",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "zellij"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Zellij terminal multiplexer theme (KDL)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "zellij.output-dir", "", "Output directory (default: ~/.config/zellij/themes)")
	cmd.Flags().StringVar(&p.themeName, "zellij.theme-name", "tinct", "Theme name for the theme file")
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
		{Name: "zellij.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/zellij/themes)", Required: false},
		{Name: "zellij.theme-name", Type: "string", Default: "tinct", Description: "Theme name for the theme file", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	if p.themeName == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
// DefaultOutputDir returns the default output directory for this plugin.
// Resolution: --zellij.output-dir → TINCT_PLUGIN_ZELLIJ_OUTPUT_DIR →
// [plugin.zellij] output_dir → platform default.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("zellij", "output_dir", p.outputDir, p.platformDefault())
}

func (p *Plugin) platformDefault() string {
	if os.Getenv("XDG_CONFIG_HOME") == "" && runtime.GOOS == "darwin" {
		return filepath.Join(paths.MacOSAppSupport("org.Zellij-Contributors.Zellij"), "themes")
	}
	return filepath.Join(paths.XDGConfigDir(), "zellij", "themes")
}

// Hooks declares zellij's pre/post-execute behaviour. Zellij doesn't
// expose a hot-reload signal, so the post message just nudges the user
// to set the theme in their config.
func (p *Plugin) Hooks() hooks.Spec {
	themeName := p.themeName
	return hooks.Spec{
		RequiredBinaries: []string{"zellij"},
		AutoCreateDir:    true,
		InstructionsFn: func(_ hooks.Context) string {
			return fmt.Sprintf("   To use this theme in Zellij, add 'theme %q' to your config.kdl or temporarily by executing: zellij options --theme %s", themeName, themeName)
		},
	}
}

// Generate creates the theme file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate theme file.
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	filename := fmt.Sprintf("%s.kdl", p.themeName)
	files[filename] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("zellij", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("theme.kdl.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for theme.kdl.tmpl\n")
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

// PreExecute checks if zellij config directory exists before generating the theme.
// Implements the output.PreExecuteHook interface.

// PostExecute provides instructions for enabling the theme.
// Implements the output.PostExecuteHook interface.
