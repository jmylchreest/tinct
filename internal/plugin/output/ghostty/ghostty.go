// Package ghostty provides an output plugin for Ghostty terminal emulator colour themes.
package ghostty

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

// pluginName is this plugin's identifier: the Name() result, the
// pluginconfig key, its XDG directory and the binary it looks for are
// all the same string.
const pluginName = "ghostty"

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Ghostty.
type Plugin struct {
	outputDir string
	reload    bool
	verbose   bool
}

// New creates a new Ghostty output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		reload:    true,
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return pluginName
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Ghostty terminal emulator theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "ghostty.output-dir", "", "Output directory (default: ~/.config/ghostty/themes)")
	cmd.Flags().BoolVar(&p.reload, "ghostty.reload", true, "Reload ghostty after generation (sends SIGUSR2)")
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
		{Name: "ghostty.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/ghostty/themes)", Required: false},
		{Name: "ghostty.reload", Type: "bool", Default: "true", Description: "Reload ghostty after generation", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
// Resolution: --ghostty.output-dir → TINCT_PLUGIN_GHOSTTY_OUTPUT_DIR →
// [plugin.ghostty] output_dir → platform default. Defaults are:
//   - Linux: $XDG_CONFIG_HOME/ghostty/themes
//   - macOS: $XDG_CONFIG_HOME/ghostty/themes if set, else
//     ~/Library/Application Support/com.mitchellh.ghostty/themes
//   - Windows: %APPDATA%/ghostty/themes
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve(pluginName, "output_dir", p.outputDir, p.platformDefault())
}

func (p *Plugin) platformDefault() string {
	if os.Getenv("XDG_CONFIG_HOME") == "" && runtime.GOOS == "darwin" {
		return filepath.Join(paths.MacOSAppSupport("com.mitchellh.ghostty"), "themes")
	}
	return filepath.Join(paths.XDGConfigDir(), pluginName, "themes")
}

// Hooks declares ghostty's pre/post-execute behaviour. The SIGUSR2
// reload broadcast is expressed via VerbSignal — the shared hooks
// runner handles process discovery and signal delivery, with a Windows
// stub that no-ops with a warning.
func (p *Plugin) Hooks() hooks.Spec {
	spec := hooks.Spec{
		RequiredBinaries: []string{pluginName},
		AutoCreateDir:    true,
	}
	if p.reload {
		spec.Reload = &hooks.ReloadSpec{
			Verb: hooks.VerbSignal,
			Args: []string{pluginName, "SIGUSR2"},
		}
	}
	return spec
}

// Generate creates the theme files.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate tinct theme file.
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}
	files["tinct.conf"] = themeContent

	return files, nil
}

// generateTheme creates the Ghostty theme file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New(pluginName, templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.tmpl\n")
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
