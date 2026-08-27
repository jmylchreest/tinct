// Package foot provides an output plugin for the foot Wayland terminal
// emulator. The plugin emits a two-file palette: a `tinct-colors.ini`
// fragment containing just the `[colors]` block, and a `tinct.ini`
// example main config that `include=`s the palette plus reasonable
// font/scrollback defaults so users without an existing `foot.ini` can
// rename it directly.
package foot

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
const pluginName = "foot"

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for the foot terminal.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new foot output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return pluginName
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "foot terminal colour theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "foot.output-dir", "", "Output directory (default: ~/.config/foot)")
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
		{Name: "foot.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/foot)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
// Resolution: --foot.output-dir → TINCT_PLUGIN_FOOT_OUTPUT_DIR →
// [plugin.foot] output_dir → platform default. foot is Wayland-only
// (Linux); paths.XDGConfigDir resolves the standard XDG path.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve(pluginName, "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), pluginName))
}

// Hooks declares foot's pre/post-execute behaviour. foot reloads its
// colour config in-place on SIGUSR1, so we signal-broadcast to every
// running foot instance. The user-facing InstructionsFn only prints the
// include-line reminder; the signal happens silently except in
// verbose mode (the runner emits its own diagnostics when verbose).
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredBinaries: []string{pluginName},
		AutoCreateDir:    true,
		Reload: &hooks.ReloadSpec{
			Verb: hooks.VerbSignal,
			Args: []string{pluginName, "SIGUSR1"},
		},
		InstructionsFn: func(_ hooks.Context) string {
			return "   Add `include=~/.config/foot/tinct-colors.ini` to your foot.ini (or rename the generated tinct.ini to foot.ini if you don't have one yet)."
		},
	}
}

// Generate creates the palette fragment and the example main config.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate template metadata fields so templates can reference the
	// resolved output path / palette filename.
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = "tinct-colors.ini"

	files := make(map[string][]byte)

	// Palette fragment — just the [colors] block, importable via include=.
	colourContent, err := p.renderTemplate("tinct-colors.ini.tmpl", themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}
	files["tinct-colors.ini"] = colourContent

	// Example main config — sensible font/scrollback defaults plus the
	// include= line. Users without an existing foot.ini can rename this
	// to foot.ini directly.
	mainContent, err := p.renderTemplate("tinct.ini.tmpl", themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate example config: %w", err)
	}
	files["tinct.ini"] = mainContent

	return files, nil
}

// renderTemplate loads (honouring user overrides), parses and executes
// the named embedded template.
func (p *Plugin) renderTemplate(name string, themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New(pluginName, templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", name, err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for %s\n", name)
	}

	tmpl, err := template.New(name).Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.Bytes(), nil
}
