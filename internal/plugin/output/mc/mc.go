// Package mc provides an output plugin for Midnight Commander (mc) skin themes.
//
// MC skins use INI format with color specifications in the form:
//
//	foreground;background;attributes
//
// where colors use xterm-256 color indices (colorN for N=0-255).
// This plugin uses the xterm256 template function from pkg/colour to convert
// tinct's hex RGB colors to the nearest xterm-256 index, skipping the first
// 16 system colors (0-15) for terminal independence.
package mc

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
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Midnight Commander.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new MC output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "mc"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Midnight Commander (mc) skin theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "mc.output-dir", "", "Output directory (default: ~/.local/share/mc/skins)")
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
		{Name: "mc.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.local/share/mc/skins)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
// MC uses XDG_DATA_HOME everywhere, which paths.XDGDataDir resolves
// consistently across Linux, macOS, and Windows.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}
	return filepath.Join(paths.XDGDataDir(), "mc", "skins")
}

// Hooks declares mc's pre/post-execute behaviour. MC does not support
// hot-reload of skins, so the post-execute message just instructs the
// user how to apply it manually.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredBinaries: []string{"mc"},
		AutoCreateDir:    true,
		Instructions: `   MC skin generated. To apply:
     mc -S tinct
   Or set in ~/.config/mc/ini:
     skin=tinct
   Note: MC does not support hot-reloading skins.
   Restart mc or select the skin via Options > Appearance.`,
	}
}

// Generate creates the skin file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate skin file.
	skinContent, err := p.generateSkin(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate skin: %w", err)
	}
	files["tinct.ini"] = skinContent

	return files, nil
}

// generateSkin creates the MC skin INI file.
func (p *Plugin) generateSkin(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("mc", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.ini.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read skin template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.ini.tmpl\n")
	}

	// Use standard template functions which include xterm256.
	tmpl, err := template.New("skin").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse skin template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute skin template: %w", err)
	}

	return buf.Bytes(), nil
}
