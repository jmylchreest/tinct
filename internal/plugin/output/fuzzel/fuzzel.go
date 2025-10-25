// Package fuzzel provides an output plugin for Fuzzel application launcher colour themes.
package fuzzel

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
	"github.com/jmylchreest/tinct/internal/util"
	"github.com/spf13/cobra"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Fuzzel.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Fuzzel output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "fuzzel"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Fuzzel application launcher colour theme configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "fuzzel.output-dir", "", "Output directory (default: ~/.config/fuzzel)")
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
		return ".config/fuzzel"
	}
	return filepath.Join(home, ".config", "fuzzel")
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

	files["tinct.ini"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("fuzzel", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.ini.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.ini.tmpl\n")
	}

	// Create template with helper functions
	tmpl, err := template.New("theme").Funcs(template.FuncMap{
		"RGBA": func(hex string, alpha string) string {
			// Convert #RRGGBB to RRGGBBAA (without #)
			if len(hex) == 7 && hex[0] == '#' {
				return hex[1:] + alpha
			}
			return hex + alpha
		},
	}).Parse(string(tmplContent))
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

// ThemeData holds data for the theme template.
// All colours are exposed directly with their semantic names.
type ThemeData struct {
	SourceTheme     string
	Background      string
	BackgroundMuted string
	Foreground      string
	ForegroundMuted string
	Accent1         string
	Accent2         string
	Accent3         string
	Accent4         string
	Danger          string
	Warning         string
	Success         string
	Info            string
}

// prepareThemeData converts a categorised palette to Fuzzel theme data.
// All colours are exposed directly with their semantic names for clarity.
func (p *Plugin) prepareThemeData(palette *colour.CategorisedPalette) ThemeData {
	data := ThemeData{
		SourceTheme:     palette.ThemeType.String(),
		Background:      util.GetColour(palette, colour.RoleBackground, "#1a1b26"),
		BackgroundMuted: util.GetColour(palette, colour.RoleBackgroundMuted, "#16161e"),
		Foreground:      util.GetColour(palette, colour.RoleForeground, "#c0caf5"),
		ForegroundMuted: util.GetColour(palette, colour.RoleForegroundMuted, "#a9b1d6"),
		Accent1:         util.GetColour(palette, colour.RoleAccent1, "#7aa2f7"),
		Accent2:         util.GetColour(palette, colour.RoleAccent2, "#bb9af7"),
		Accent3:         util.GetColour(palette, colour.RoleAccent3, "#7dcfff"),
		Accent4:         util.GetColour(palette, colour.RoleAccent4, "#9ece6a"),
		Danger:          util.GetColour(palette, colour.RoleDanger, "#f7768e"),
		Warning:         util.GetColour(palette, colour.RoleWarning, "#e0af68"),
		Success:         util.GetColour(palette, colour.RoleSuccess, "#9ece6a"),
		Info:            util.GetColour(palette, colour.RoleInfo, "#7aa2f7"),
	}

	// Add helper methods to data for RGBA conversion
	return data
}

// Helper methods for ThemeData to convert colors to RGBA format
func (td ThemeData) BackgroundRGBA(alpha string) string {
	return hexToRGBA(td.Background, alpha)
}

func (td ThemeData) BackgroundMutedRGBA(alpha string) string {
	return hexToRGBA(td.BackgroundMuted, alpha)
}

func (td ThemeData) ForegroundRGBA(alpha string) string {
	return hexToRGBA(td.Foreground, alpha)
}

func (td ThemeData) ForegroundMutedRGBA(alpha string) string {
	return hexToRGBA(td.ForegroundMuted, alpha)
}

func (td ThemeData) Accent1RGBA(alpha string) string {
	return hexToRGBA(td.Accent1, alpha)
}

func (td ThemeData) Accent2RGBA(alpha string) string {
	return hexToRGBA(td.Accent2, alpha)
}

func (td ThemeData) Accent3RGBA(alpha string) string {
	return hexToRGBA(td.Accent3, alpha)
}

func (td ThemeData) Accent4RGBA(alpha string) string {
	return hexToRGBA(td.Accent4, alpha)
}

func (td ThemeData) DangerRGBA(alpha string) string {
	return hexToRGBA(td.Danger, alpha)
}

func (td ThemeData) WarningRGBA(alpha string) string {
	return hexToRGBA(td.Warning, alpha)
}

func (td ThemeData) SuccessRGBA(alpha string) string {
	return hexToRGBA(td.Success, alpha)
}

func (td ThemeData) InfoRGBA(alpha string) string {
	return hexToRGBA(td.Info, alpha)
}

// hexToRGBA converts #RRGGBB to RRGGBBAA format (without #)
func hexToRGBA(hex string, alpha string) string {
	if len(hex) == 7 && hex[0] == '#' {
		return hex[1:] + alpha
	}
	// If already without #, just append alpha
	if len(hex) == 6 {
		return hex + alpha
	}
	return hex + alpha
}

// PreExecute checks if fuzzel is available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if fuzzel executable exists on PATH
	_, err = exec.LookPath("fuzzel")
	if err != nil {
		return true, "fuzzel executable not found on $PATH", nil
	}

	// Check if config directory exists (create it if not, as it's simple)
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// For fuzzel, we can create the directory since it's straightforward
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return true, fmt.Sprintf("fuzzel config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}
