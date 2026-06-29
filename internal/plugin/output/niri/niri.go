// Package niri provides an output plugin for the niri scrollable-tiling
// Wayland compositor.
package niri

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

// colourFileName is the merge file niri includes; it carries the actual themed
// config sections. stubFileName is the example wiring file.
const (
	colourFileName = "tinct-colours.kdl"
	stubFileName   = "tinct.kdl"
)

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for niri.
type Plugin struct {
	outputDir    string
	generateStub bool
	stubPath     string
	verbose      bool
}

// New creates a new niri output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir:    "",
		generateStub: true,
		stubPath:     "",
		verbose:      false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "niri"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "niri compositor theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "niri.output-dir", "", "Output directory (default: ~/.config/niri)")
	cmd.Flags().BoolVar(&p.generateStub, "niri.generate-stub", true, "Generate example config stub")
	cmd.Flags().StringVar(&p.stubPath, "niri.stub-path", "", "Custom path for stub file")
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
		{Name: "niri.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/niri)", Required: false},
		{Name: "niri.generate-stub", Type: "bool", Default: "true", Description: "Generate example config stub", Required: false},
		{Name: "niri.stub-path", Type: "string", Default: "", Description: "Custom path for stub file", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	// Nothing to validate - all fields have defaults.
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("niri", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "niri"))
}

// Hooks declares niri's pre-execute behaviour. The niri config dir must exist
// (signal that niri is configured); the niri binary is optional. No reload hook
// is needed: niri watches its config and every included file, live-reloading
// automatically when the generated colour file changes.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredDirs:     []string{"~/.config/niri"},
		OptionalBinaries: []string{"niri"},
		AutoCreateDir:    true,
	}
}

// Generate creates the theme file and optional stub configuration.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate output directory and colour file name in theme data for templates.
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = colourFileName

	files := make(map[string][]byte)

	// Generate main theme (colour merge) file.
	themeContent, err := p.render("tinct-colours.kdl.tmpl", themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}
	files[colourFileName] = themeContent

	// Generate stub if requested.
	if p.generateStub {
		stubContent, err := p.render("tinct.kdl.tmpl", themeData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate stub: %w", err)
		}

		stubFile := stubFileName
		if p.stubPath != "" {
			stubFile = filepath.Base(p.stubPath)
		}
		files[stubFile] = stubContent
	}

	return files, nil
}

// render loads and executes a template, honouring custom user overrides.
func (p *Plugin) render(templateName string, themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("niri", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	tmplContent, fromCustom, err := loader.Load(templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", templateName, err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for %s\n", templateName)
	}

	tmpl, err := template.New(templateName).Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.Bytes(), nil
}
