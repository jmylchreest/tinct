// Package waybar provides an output plugin for Waybar status bar colour themes.
package waybar

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/mitchellh/go-ps"
	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
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

// Plugin implements the output.Plugin interface for Waybar.
type Plugin struct {
	outputDir    string
	generateStub bool
	reloadConfig bool
	verbose      bool
}

// New creates a new Waybar output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir:    "",
		generateStub: true,
		verbose:      false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "waybar"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Waybar status bar color theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
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
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
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

// findProcessByName finds all PIDs of processes with the given name.
// Uses go-ps library for cross-platform process discovery.
func findProcessByName(name string) ([]int, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %w", err)
	}

	var pids []int
	for _, p := range processes {
		if p.Executable() == name {
			pids = append(pids, p.Pid())
		}
	}

	return pids, nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	// Expand ~ to home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/waybar/themes"
	}
	return filepath.Join(home, ".config", "waybar", "themes")
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
	loader := tmplloader.New("waybar", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colours.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read colors template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colours.css.tmpl\n")
	}

	tmpl, err := template.New("colors").Funcs(common.TemplateFuncs()).Parse(string(tmplContent))
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
	loader := tmplloader.New("waybar", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read stub template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.css.tmpl\n")
	}

	tmpl, err := template.New("stub").Funcs(common.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse stub template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute stub template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute checks if waybar is available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if waybar executable exists on PATH.
	_, err = exec.LookPath("waybar")
	if err != nil {
		return true, "waybar executable not found on $PATH", nil
	}

	// Check if config directory exists.
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return true, fmt.Sprintf("waybar config directory not found: %s", configDir), nil
	}

	return false, "", nil
}
