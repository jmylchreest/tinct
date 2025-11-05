// Package swayosd provides an output plugin for SwayOSD on-screen display colour themes.
package swayosd

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for SwayOSD.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new SwayOSD output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "swayosd"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate SwayOSD on-screen display colour theme configuration"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "swayosd.output-dir", "", "Output directory (default: ~/.config/swayosd)")
}

// SetVerbose enables or disables verbose logging for the plugin.
// Implements the output.VerbosePlugin interface.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetEmbeddedFS returns the embedded template filesystem.
// Implements the output.TemplateProvider interface.
func (p *Plugin) GetEmbeddedFS() interface{} {
	return templates
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/swayosd"
	}
	return filepath.Join(home, ".config", "swayosd")
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

	files["style.css"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("swayosd", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("style.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for style.css.tmpl\n")
	}

	tmpl, err := template.New("theme").Funcs(common.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute checks if swayosd is available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if swayosd-client executable exists on PATH.
	_, err = exec.LookPath("swayosd-client")
	if err != nil {
		return true, "swayosd-client executable not found on $PATH", nil
	}

	// Check if config directory exists (create it if not).
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// For swayosd, we can create the directory since it's straightforward.
		if err := os.MkdirAll(configDir, 0o755); err != nil { // #nosec G301 - Config directory needs standard permissions
			return true, fmt.Sprintf("swayosd config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}
