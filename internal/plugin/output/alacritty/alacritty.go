// Package alacritty provides an output plugin for Alacritty terminal colour themes.
package alacritty

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
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/output"
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

// Plugin implements the output.Plugin interface for Alacritty terminal.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Alacritty output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "alacritty"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Alacritty terminal colour theme configuration"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "alacritty.output-dir", "", "Output directory (default: ~/.config/alacritty)")
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
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		{Name: "alacritty.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/alacritty)", Required: false},
	}
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
		return ".config/alacritty"
	}
	return filepath.Join(home, ".config", "alacritty")
}

// Generate creates the theme file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate template metadata fields.
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = "tinct-colors.toml"

	files := make(map[string][]byte)

	// Generate theme file.
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	files["tinct-colors.toml"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("alacritty", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colors.toml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colors.toml.tmpl\n")
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

// PreExecute checks if alacritty is available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if alacritty executable exists on PATH.
	_, err = exec.LookPath("alacritty")
	if err != nil {
		return true, "alacritty executable not found on $PATH", nil
	}

	// Check if config directory exists, create if it doesn't.
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Try to create the config directory.
		if err := os.MkdirAll(configDir, 0o755); err != nil { // #nosec G301 - Config directory needs standard permissions
			return true, fmt.Sprintf("failed to create alacritty config directory: %s", configDir), nil
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Created alacritty config directory: %s\n", configDir)
		}
	}

	return false, "", nil
}

// PostExecute provides usage instructions for applying the theme.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(_ context.Context, _ output.ExecutionContext, generatedFiles []string) error {
	if p.verbose && len(generatedFiles) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Alacritty theme generated successfully!\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   To use this theme, add to your alacritty.toml:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   import = [\n")
		fmt.Fprintf(os.Stderr, "     \"%s\"\n", filepath.Join(p.DefaultOutputDir(), "tinct-colors.toml"))
		fmt.Fprintf(os.Stderr, "   ]\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Note: Alacritty automatically reloads config when files change.\n")
		fmt.Fprintf(os.Stderr, "   New colors will apply immediately to all open terminals.\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	return nil
}
