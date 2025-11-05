// Package neovim provides an output plugin for Neovim colour themes.
package neovim

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

// Plugin implements the output.Plugin interface for Neovim.
type Plugin struct {
	outputDir string
	themeName string
	verbose   bool
}

// New creates a new Neovim output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		themeName: "tinct",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "neovim"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Neovim colour scheme (Lua format)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "neovim.output-dir", "", "Output directory (default: ~/.config/nvim/colors)")
	cmd.Flags().StringVar(&p.themeName, "neovim.theme-name", "tinct", "Theme name for the colorscheme")
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
	if p.themeName == "" {
		return fmt.Errorf("theme name cannot be empty")
	}
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/nvim/colors"
	}
	return filepath.Join(home, ".config", "nvim", "colors")
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

	filename := fmt.Sprintf("%s.lua", p.themeName)
	files[filename] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("neovim", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("theme.lua.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for theme.lua.tmpl\n")
	}

	tmpl, err := template.New("theme").Funcs(common.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	// Set plugin-specific themeName for template.
	themeData.ThemeName = p.themeName

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute checks if neovim config directory exists before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if nvim executable exists on PATH.
	_, err = exec.LookPath("nvim")
	if err != nil {
		return true, "nvim executable not found on $PATH", nil
	}

	// Check if config directory exists, create if it doesn't.
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Try to create the directory.
		if err := os.MkdirAll(configDir, 0o755); err != nil { // #nosec G301 - Config directory needs standard permissions
			return true, fmt.Sprintf("neovim colors directory not found and could not be created: %s", configDir), nil
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Created neovim colors directory: %s\n", configDir)
		}
	}

	return false, "", nil
}

// PostExecute provides instructions for enabling the theme.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(_ context.Context, _ output.ExecutionContext, writtenFiles []string) error {
	if p.verbose && len(writtenFiles) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   To use this theme in Neovim, add to your init.lua:\n")
		fmt.Fprintf(os.Stderr, "   vim.cmd('colorscheme %s')\n", p.themeName)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Or in init.vim:\n")
		fmt.Fprintf(os.Stderr, "   colorscheme %s\n", p.themeName)
	}
	return nil
}
