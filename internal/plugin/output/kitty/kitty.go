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
	outputDir    string
	reloadConfig bool
	verbose      bool
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

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "kitty.output-dir", "", "Output directory (default: ~/.config/kitty)")
	cmd.Flags().BoolVar(&p.reloadConfig, "kitty.reload", false, "Reload kitty config after generation (sends SIGUSR1)")
}

// SetVerbose enables or disables verbose logging for the plugin.
// Implements the output.VerbosePlugin interface.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
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
		return ".config/kitty"
	}
	return filepath.Join(home, ".config", "kitty")
}

// Generate creates the theme file.
// Returns map of filename -> content
func (p *Plugin) Generate(palette *colour.CategorisedPalette) (map[string][]byte, error) {
	if palette == nil {
		return nil, fmt.Errorf("palette cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate theme file
	themeContent, err := p.generateTheme(palette)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	files["tinct.conf"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("kitty", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.conf.tmpl\n")
	}

	tmpl, err := template.New("theme").Funcs(common.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	data := p.prepareThemeData(palette)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	return buf.Bytes(), nil
}

// prepareThemeData converts a categorised palette to a PaletteHelper for templates.
// This provides DRY access to all colors with multiple format options.
func (p *Plugin) prepareThemeData(palette *colour.CategorisedPalette) *colour.PaletteHelper {
	return colour.NewPaletteHelper(palette)
}

// PreExecute checks if kitty is available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if kitty executable exists on PATH
	_, err = exec.LookPath("kitty")
	if err != nil {
		return true, "kitty executable not found on $PATH", nil
	}

	// Check if config directory exists
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return true, fmt.Sprintf("kitty config directory not found: %s", configDir), nil
	}

	return false, "", nil
}

// PostExecute reloads kitty configuration if requested.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, writtenFiles []string) error {
	if !p.reloadConfig {
		return nil
	}

	// Send SIGUSR1 to all running kitty instances to reload config
	// This is the standard way to reload kitty configuration
	cmd := exec.CommandContext(ctx, "killall", "-SIGUSR1", "kitty")
	if err := cmd.Run(); err != nil {
		// Check if the error is because no kitty process was found
		if exitErr, ok := err.(*exec.ExitError); ok {
			// killall returns exit code 1 if no processes matched
			if exitErr.ExitCode() == 1 {
				return fmt.Errorf("no running kitty instances found to reload")
			}
		}
		return fmt.Errorf("failed to reload kitty config: %w", err)
	}

	return nil
}
