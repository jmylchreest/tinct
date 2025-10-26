// Package hyprland provides an output plugin for Hyprland window manager colour themes.
package hyprland

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

// Plugin implements the output.Plugin interface for Hyprland.
type Plugin struct {
	outputDir    string
	generateStub bool
	stubPath     string
	reloadConfig bool
	verbose      bool
}

// New creates a new Hyprland output plugin with default settings.
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
	return "hyprland"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Hyprland colour theme configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "hyprland.output-dir", "", "Output directory (default: ~/.config/hypr)")
	cmd.Flags().BoolVar(&p.generateStub, "hyprland.generate-stub", true, "Generate example config stub")
	cmd.Flags().StringVar(&p.stubPath, "hyprland.stub-path", "", "Custom path for stub file")
	cmd.Flags().BoolVar(&p.reloadConfig, "hyprland.reload", false, "Reload hyprland config after generation (runs hyprctl reload)")
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
	// Nothing to validate - all fields have defaults
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	// Expand ~ to home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/hypr"
	}
	return filepath.Join(home, ".config", "hypr")
}

// Generate creates the theme file and optional stub configuration.
// Returns map of filename -> content
func (p *Plugin) Generate(palette *colour.CategorisedPalette) (map[string][]byte, error) {
	if palette == nil {
		return nil, fmt.Errorf("palette cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate main theme file
	themeContent, err := p.generateTheme(palette)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	themeFile := "tinct-colours.conf"
	files[themeFile] = themeContent

	// Generate stub if requested
	if p.generateStub {
		stubContent, err := p.generateStubConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to generate stub: %w", err)
		}

		stubFile := "tinct.conf"
		if p.stubPath != "" {
			stubFile = filepath.Base(p.stubPath)
		}
		files[stubFile] = stubContent
	}

	return files, nil
}

// generateTheme creates the main theme configuration file with colour variables.
func (p *Plugin) generateTheme(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("hyprland", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colours.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colours.conf.tmpl\n")
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

// generateStubConfig creates an example configuration file showing how to use the theme.
func (p *Plugin) generateStubConfig() ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("hyprland", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read example template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.conf.tmpl\n")
	}

	tmpl, err := template.New("example").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse example template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("failed to execute example template: %w", err)
	}

	return buf.Bytes(), nil
}

// prepareThemeData converts a categorised palette to PaletteHelper for template access.
func (p *Plugin) prepareThemeData(palette *colour.CategorisedPalette) *colour.PaletteHelper {
	return colour.NewPaletteHelper(palette)
}

// PreExecute checks if the config directory exists before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if config directory exists - create it if needed
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Try to create the directory
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return true, fmt.Sprintf("hyprland config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}

// PostExecute reloads hyprland configuration if requested.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, writtenFiles []string) error {
	if !p.reloadConfig {
		return nil
	}

	// Reload hyprland configuration using hyprctl
	cmd := exec.CommandContext(ctx, "hyprctl", "reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload hyprland config: %w (output: %s)", err, string(output))
	}

	return nil
}
