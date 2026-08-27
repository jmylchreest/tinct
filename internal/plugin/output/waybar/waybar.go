// Package waybar provides an output plugin for Waybar status bar colour themes.
package waybar

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

// pluginName is this plugin's identifier: the Name() result, the
// pluginconfig key, its XDG directory and the binary it looks for are
// all the same string.
const pluginName = "waybar"

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Waybar.
type Plugin struct {
	outputDir    string
	generateStub bool
	reloadConfig bool
	verbose      bool
}

// New creates a new Waybar output plugin with default settings. The
// reloadConfig default matches the --waybar.reload flag default so that
// constructing the plugin without flag parsing (e.g. from the
// check-readmes tool) sees the same reload behaviour users get on the
// CLI.
func New() *Plugin {
	return &Plugin{
		outputDir:    "",
		generateStub: true,
		reloadConfig: true,
		verbose:      false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return pluginName
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Waybar status bar theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "waybar.output-dir", "", "Output directory (default: ~/.config/waybar/themes)")
	cmd.Flags().BoolVar(&p.generateStub, "waybar.generate-stub", true, "Generate example CSS stub")
	cmd.Flags().BoolVar(&p.reloadConfig, "waybar.reload", true, "Reload waybar after generation (sends SIGUSR2)")
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
		{Name: "waybar.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/waybar/themes)", Required: false},
		{Name: "waybar.generate-stub", Type: "bool", Default: "true", Description: "Generate example CSS stub", Required: false},
		{Name: "waybar.reload", Type: "bool", Default: "true", Description: "Reload waybar after generation", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	// Nothing to validate - all fields have defaults.
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve(pluginName, "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), pluginName, "themes"))
}

// Generate creates the theme files.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate output directory and color file name in theme data for templates.
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = "tinct.css"

	files := make(map[string][]byte)

	// Generate colors file - goes to themes/tinct.css
	colorsContent, err := p.generateColors(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate colors: %w", err)
	}
	files["tinct.css"] = colorsContent

	// Generate example CSS if requested - goes to parent dir (waybar/) as tinct-style.css
	if p.generateStub {
		stubContent, err := p.generateStubCSS(themeData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate stub: %w", err)
		}
		files["../tinct-style.css"] = stubContent
	}

	return files, nil
}

// generateColors creates the color definitions CSS file.
func (p *Plugin) generateColors(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New(pluginName, templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colours.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read colors template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colours.css.tmpl\n")
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

// generateStubCSS creates an example CSS file showing how to use the colors.
func (p *Plugin) generateStubCSS(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New(pluginName, templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read stub template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.css.tmpl\n")
	}

	tmpl, err := template.New("stub").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse stub template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute stub template: %w", err)
	}

	return buf.Bytes(), nil
}

// Hooks declares waybar's pre/post-execute behaviour. The themes
// subdirectory is required (waybar must already be configured); reload
// sends SIGUSR2 to all running waybar instances when --waybar.reload is
// set (default). The previous implementation declared the SIGUSR2 flag
// but never wired the signal — this connects it via the shared signal
// verb runner.
func (p *Plugin) Hooks() hooks.Spec {
	spec := hooks.Spec{
		RequiredBinaries: []string{pluginName},
		RequiredDirs:     []string{p.DefaultOutputDir()},
	}
	if p.reloadConfig {
		spec.Reload = &hooks.ReloadSpec{
			Verb: hooks.VerbSignal,
			Args: []string{pluginName, "SIGUSR2"},
		}
	}
	return spec
}
