// Package neovim provides an output plugin for Neovim colour themes.
package neovim

import (
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	"github.com/jmylchreest/tinct/internal/version"
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
	return "Neovim colour scheme (Lua)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
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
func (p *Plugin) GetEmbeddedFS() any {
	return templates
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "neovim.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/nvim/colors)", Required: false},
		{Name: "neovim.theme-name", Type: "string", Default: "tinct", Description: "Theme name for the colorscheme", Required: false},
	}
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

// lualineThemePath returns the relative path (from DefaultOutputDir) for the
// standalone lualine theme file. Since DefaultOutputDir is ~/.config/nvim/colors
// and lualine expects themes at ~/.config/nvim/lua/lualine/themes/, we traverse
// up one level from colors/ into the nvim config root.
// This follows the same pattern used by the waybar plugin for stub files.
func (p *Plugin) lualineThemePath() string {
	return filepath.Join("..", "lua", "lualine", "themes", fmt.Sprintf("%s.lua", p.themeName))
}

// Generate creates the theme files.
// Returns map of filename -> content, where filenames are relative to DefaultOutputDir().
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate template metadata fields.
	filename := fmt.Sprintf("%s.lua", p.themeName)
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = filename

	files := make(map[string][]byte)

	// Generate main colorscheme file (written to ~/.config/nvim/colors/<theme>.lua).
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}
	files[filename] = themeContent

	// Generate standalone lualine theme file (written to ~/.config/nvim/lua/lualine/themes/<theme>.lua).
	// Lualine discovers themes via nvim_get_runtime_file("lua/lualine/themes/<name>.lua"),
	// which searches &runtimepath. Since ~/.config/nvim is on runtimepath, placing the
	// file there ensures lualine finds it regardless of plugin load order.
	lualineContent, err := p.generateLualineTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate lualine theme: %w", err)
	}
	files[p.lualineThemePath()] = lualineContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	themeData.ThemeName = p.themeName
	out, fromCustom, err := utils.RenderTemplate("neovim", "theme.lua.tmpl", templates, themeData, p.verbose)
	if err != nil {
		return nil, err
	}
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for theme.lua.tmpl\n")
	}
	return out, nil
}

// generateLualineTheme creates a standalone lualine theme file.
func (p *Plugin) generateLualineTheme(themeData *colour.ThemeData) ([]byte, error) {
	themeData.ThemeName = p.themeName
	out, fromCustom, err := utils.RenderTemplate("neovim", "lualine_theme.lua.tmpl", templates, themeData, p.verbose)
	if err != nil {
		return nil, err
	}
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for lualine_theme.lua.tmpl\n")
	}
	return out, nil
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
		if err := os.MkdirAll(configDir, 0o750); err != nil {
			return true, fmt.Sprintf("neovim colors directory not found and could not be created: %s", configDir), nil
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Created neovim colors directory: %s\n", configDir)
		}
	}

	return false, "", nil
}

// PostExecute provides usage instructions.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(_ context.Context, _ output.ExecutionContext, writtenFiles []string) error {
	if p.verbose && len(writtenFiles) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Neovim colorscheme generated successfully!\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   To use this theme, add to your init.lua:\n")
		fmt.Fprintf(os.Stderr, "   vim.cmd('colorscheme %s')\n", p.themeName)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Or in init.vim:\n")
		fmt.Fprintf(os.Stderr, "   colorscheme %s\n", p.themeName)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Lualine: theme = 'auto' works out of the box.\n")
		fmt.Fprintf(os.Stderr, "   For the dedicated theme: theme = '%s'\n", p.themeName)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Note: The colorscheme will auto-reload when tinct updates it.\n")
		fmt.Fprintf(os.Stderr, "   This works via a file system watcher built into the theme.\n")
	}
	return nil
}
