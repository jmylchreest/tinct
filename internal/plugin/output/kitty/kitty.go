// Package kitty provides an output plugin for Kitty terminal colour themes.
package kitty

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/spf13/cobra"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Kitty terminal.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Kitty output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "kitty"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Kitty terminal colour theme configuration"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "kitty.output-dir", "", "Output directory (default: ~/.config/kitty/themes)")
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
		return ".config/kitty/themes"
	}
	return filepath.Join(home, ".config", "kitty", "themes")
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

	files["tinct.conf"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("kitty", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.conf.tmpl\n")
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

// PreExecute checks if kitty is available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if kitty executable exists on PATH.
	_, err = exec.LookPath("kitty")
	if err != nil {
		return true, "kitty executable not found on $PATH", nil
	}

	// Check if themes directory exists, create if it doesn't.
	themesDir := p.DefaultOutputDir()
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		// Try to create the themes directory.
		if err := os.MkdirAll(themesDir, 0o755); err != nil { // #nosec G301 - Config directory needs standard permissions
			return true, fmt.Sprintf("failed to create kitty themes directory: %s", themesDir), nil
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Created kitty themes directory: %s\n", themesDir)
		}
	}

	return false, "", nil
}

// PostExecute applies the theme to all running kitty instances.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, writtenFiles []string) error {
	// Apply the theme to all running kitty instances using kitten themes command.
	// This is the recommended way per https://sw.kovidgoyal.net/kitty/kittens/themes/.
	cmd := exec.CommandContext(ctx, "kitten", "themes", "--reload-in=all", "tinct")
	if err := cmd.Run(); err != nil {
		// Don't treat this as a fatal error - theme file was still generated successfully.
		// This can fail if no kitty instances are running or kitten is not in PATH.
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Note: Could not auto-apply theme (no running kitty instances or kitten not in $PATH)\n")
		}
		return nil
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Theme applied to all kitty instances\n")
	}

	return nil
}
