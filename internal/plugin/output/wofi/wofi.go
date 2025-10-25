// Package wofi provides an output plugin for Wofi application launcher colour themes.
package wofi

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

//go:embed *.tmpl tinct-colors
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Wofi.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Wofi output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "wofi"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Wofi application launcher colour theme configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "wofi.output-dir", "", "Output directory (default: ~/.config/wofi)")
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
		return ".config/wofi"
	}
	return filepath.Join(home, ".config", "wofi")
}

// Generate creates the theme files.
// Returns map of filename -> content
func (p *Plugin) Generate(palette *colour.CategorisedPalette) (map[string][]byte, error) {
	if palette == nil {
		return nil, fmt.Errorf("palette cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate colors file
	colorsContent, err := p.generateColors(palette)
	if err != nil {
		return nil, fmt.Errorf("failed to generate colors: %w", err)
	}
	files["tinct-colors"] = colorsContent

	// Generate style file
	styleContent, err := p.generateStyle(palette)
	if err != nil {
		return nil, fmt.Errorf("failed to generate style: %w", err)
	}
	files["style.css"] = styleContent

	return files, nil
}

// generateColors creates the colors file.
func (p *Plugin) generateColors(palette *colour.CategorisedPalette) ([]byte, error) {
	loader := tmplloader.New("wofi", templates)
	if p.verbose {
		loader.WithVerbose(true, &verboseLogger{out: os.Stderr})
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colors")
	if err != nil {
		return nil, fmt.Errorf("failed to read colors template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colors\n")
	}

	tmpl, err := template.New("colors").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse colors template: %w", err)
	}

	data := p.prepareThemeData(palette)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute colors template: %w", err)
	}

	return buf.Bytes(), nil
}

// generateStyle creates the style CSS file.
func (p *Plugin) generateStyle(palette *colour.CategorisedPalette) ([]byte, error) {
	loader := tmplloader.New("wofi", templates)
	if p.verbose {
		loader.WithVerbose(true, &verboseLogger{out: os.Stderr})
	}
	tmplContent, fromCustom, err := loader.Load("style.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read style template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for style.css.tmpl\n")
	}

	tmpl, err := template.New("style").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse style template: %w", err)
	}

	data := p.prepareThemeData(palette)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute style template: %w", err)
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

// prepareThemeData converts a categorised palette to Wofi theme data.
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

// getColour retrieves a colour by role with a fallback.
func (p *Plugin) getColour(palette *colour.CategorisedPalette, role colour.ColourRole, fallback string) string {
	if c, ok := palette.Get(role); ok {
		return c.Hex
	}
	return fallback
}

// PreExecute checks if wofi is available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if wofi executable exists on PATH
	_, err = exec.LookPath("wofi")
	if err != nil {
		return true, "wofi executable not found on $PATH", nil
	}

	// Check if config directory exists (create it if not)
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return true, fmt.Sprintf("wofi config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}
