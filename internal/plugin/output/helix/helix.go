// Package helix provides an output plugin for the Helix text editor.
//
// Helix loads themes from ~/.config/helix/themes/<name>.toml. Themes
// may declare a palette via [palette] and reference colour names by
// key, and one theme can `inherits = "<other>"` another to pick up
// its palette. We exploit that to produce two files:
//
//   - tinct-palette.toml — just [palette] entries (the named colours).
//   - tinct.toml          — `inherits = "tinct-palette"` plus all the
//     UI / syntax key bindings.
//
// This keeps the user-facing file small enough to fork without
// losing the colour values, while letting tinct overwrite the
// palette on every run.
package helix

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

// Plugin implements the output.Plugin interface for Helix.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Helix output plugin with default settings.
func New() *Plugin { return &Plugin{} }

// Name returns the plugin name.
func (p *Plugin) Name() string { return "helix" }

// Description returns the plugin description.
func (p *Plugin) Description() string { return "Helix editor theme (TOML, two-file)" }

// Version returns the plugin version.
func (p *Plugin) Version() string { return version.Version }

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "helix.output-dir", "", "Output directory (default: ~/.config/helix/themes)")
}

// SetVerbose enables or disables verbose logging for the plugin.
// Implements the output.VerbosePlugin interface.
func (p *Plugin) SetVerbose(verbose bool) { p.verbose = verbose }

// GetEmbeddedFS returns the embedded template filesystem.
// Implements the output.TemplateProvider interface.
func (p *Plugin) GetEmbeddedFS() any { return templates }

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "helix.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/helix/themes)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error { return nil }

// DefaultOutputDir returns the resolved output directory.
// Resolution chain (highest priority first): --helix.output-dir,
// TINCT_PLUGIN_HELIX_OUTPUT_DIR, [plugin.helix] output_dir in
// tinct.toml, then the platform-resolved default. Helix uses
// $XDG_CONFIG_HOME on every supported platform.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("helix", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "helix", "themes"))
}

// Hooks declares helix's pre/post-execute behaviour. Helix has no IPC
// channel for theme reload, so we mark `hx` as required (it's the
// editor binary) but don't wire any post-execute action — the
// instructions block tells the user the two commands they need to run
// inside Helix itself.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredBinaries: []string{"hx"},
		AutoCreateDir:    true,
		Instructions: `   Helix theme generated successfully!

   To activate, either set this in your ~/.config/helix/config.toml:

     theme = "tinct"

   Or, inside a running Helix session, run:

     :config-reload
     :theme tinct

   Helix has no external reload mechanism, so existing sessions need
   the :theme command above. New sessions pick up the config-file
   value automatically.`,
	}
}

// Generate creates the theme files.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate template metadata fields so the (sparse) template
	// header comments can mention where the file lives.
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = "tinct-palette.toml"

	files := make(map[string][]byte)

	paletteContent, err := p.renderTemplate("tinct-palette.toml.tmpl", themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate palette: %w", err)
	}
	files["tinct-palette.toml"] = paletteContent

	themeContent, err := p.renderTemplate("tinct.toml.tmpl", themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}
	files["tinct.toml"] = themeContent

	return files, nil
}

// renderTemplate loads (with user-override support), parses, and
// executes a single embedded template. The custom-template loader is
// shared by every output plugin so users can `tinct plugins templates
// dump -o helix` and edit either file independently.
func (p *Plugin) renderTemplate(name string, themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("helix", templates)
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

	tmpl, err := template.New("helix").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}
