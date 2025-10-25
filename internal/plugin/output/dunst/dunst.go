// Package dunst provides an output plugin for Dunst notification daemon colour themes.
package dunst

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/jmylchreest/tinct/internal/colour"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/spf13/cobra"
)

// verboseLogger implements the template.Logger interface for verbose output.
type verboseLogger struct {
	out io.Writer
}

func (l *verboseLogger) Printf(format string, v ...any) {
	fmt.Fprintf(l.out, format+"\n", v...)
}

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Dunst.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Dunst output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "dunst"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Dunst notification daemon colour theme configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "dunst.output-dir", "", "Output directory (default: ~/.config/dunst)")
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
		return ".config/dunst"
	}
	return filepath.Join(home, ".config", "dunst")
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

	files["tinct.dunstrc"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("dunst", templates)
	if p.verbose {
		loader.WithVerbose(true, &verboseLogger{out: os.Stderr})
	}
	tmplContent, fromCustom, err := loader.Load("tinct.dunstrc.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.dunstrc.tmpl\n")
	}

	tmpl, err := template.New("theme").Parse(string(tmplContent))
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

// prepareThemeData converts a categorised palette to Dunst theme data.
func (p *Plugin) prepareThemeData(palette *colour.CategorisedPalette) ThemeData {
	return ThemeData{
		SourceTheme:     palette.ThemeType.String(),
		Background:      p.getColour(palette, colour.RoleBackground, "#1a1b26"),
		BackgroundMuted: p.getColour(palette, colour.RoleBackgroundMuted, "#16161e"),
		Foreground:      p.getColour(palette, colour.RoleForeground, "#c0caf5"),
		ForegroundMuted: p.getColour(palette, colour.RoleForegroundMuted, "#a9b1d6"),
		Accent1:         p.getColour(palette, colour.RoleAccent1, "#7aa2f7"),
		Accent2:         p.getColour(palette, colour.RoleAccent2, "#bb9af7"),
		Accent3:         p.getColour(palette, colour.RoleAccent3, "#7dcfff"),
		Accent4:         p.getColour(palette, colour.RoleAccent4, "#9ece6a"),
		Danger:          p.getColour(palette, colour.RoleDanger, "#f7768e"),
		Warning:         p.getColour(palette, colour.RoleWarning, "#e0af68"),
		Success:         p.getColour(palette, colour.RoleSuccess, "#9ece6a"),
		Info:            p.getColour(palette, colour.RoleInfo, "#7aa2f7"),
	}
}

// Helper methods for ThemeData to add alpha channel

// DangerWithAlpha returns danger color with specified alpha
func (td ThemeData) DangerWithAlpha(alpha string) string {
	return hexWithAlpha(td.Danger, alpha)
}

// WarningWithAlpha returns warning color with specified alpha
func (td ThemeData) WarningWithAlpha(alpha string) string {
	return hexWithAlpha(td.Warning, alpha)
}

// SuccessWithAlpha returns success color with specified alpha
func (td ThemeData) SuccessWithAlpha(alpha string) string {
	return hexWithAlpha(td.Success, alpha)
}

// InfoWithAlpha returns info color with specified alpha
func (td ThemeData) InfoWithAlpha(alpha string) string {
	return hexWithAlpha(td.Info, alpha)
}

// BackgroundWithAlpha returns background color with specified alpha
func (td ThemeData) BackgroundWithAlpha(alpha string) string {
	return hexWithAlpha(td.Background, alpha)
}

// ForegroundWithAlpha returns foreground color with specified alpha
func (td ThemeData) ForegroundWithAlpha(alpha string) string {
	return hexWithAlpha(td.Foreground, alpha)
}

// Accent1WithAlpha returns accent1 color with specified alpha
func (td ThemeData) Accent1WithAlpha(alpha string) string {
	return hexWithAlpha(td.Accent1, alpha)
}

// Accent2WithAlpha returns accent2 color with specified alpha
func (td ThemeData) Accent2WithAlpha(alpha string) string {
	return hexWithAlpha(td.Accent2, alpha)
}

// Accent3WithAlpha returns accent3 color with specified alpha
func (td ThemeData) Accent3WithAlpha(alpha string) string {
	return hexWithAlpha(td.Accent3, alpha)
}

// Accent4WithAlpha returns accent4 color with specified alpha
func (td ThemeData) Accent4WithAlpha(alpha string) string {
	return hexWithAlpha(td.Accent4, alpha)
}

// hexWithAlpha adds alpha channel to hex color
// Dunst uses #RRGGBBAA format
func hexWithAlpha(hex string, alpha string) string {
	if len(hex) == 7 && hex[0] == '#' {
		return hex + alpha
	}
	if len(hex) == 6 {
		return "#" + hex + alpha
	}
	// If already has alpha or other format, return as-is
	return hex
}

// getColour retrieves a colour by role with a fallback.
func (p *Plugin) getColour(palette *colour.CategorisedPalette, role colour.ColourRole, fallback string) string {
	if c, ok := palette.Get(role); ok {
		return c.Hex
	}
	return fallback
}

// PreExecute checks if dunst is available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if dunst executable exists on PATH
	_, err = exec.LookPath("dunst")
	if err != nil {
		return true, "dunst executable not found on $PATH", nil
	}

	// Check if config directory exists (create it if not)
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// For dunst, we can create the directory since it's straightforward
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return true, fmt.Sprintf("dunst config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}
