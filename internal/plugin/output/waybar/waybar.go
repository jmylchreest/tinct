// Package waybar provides an output plugin for Waybar status bar color themes.
package waybar

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "waybar.output-dir", "", "Output directory (default: ~/.config/waybar)")
	cmd.Flags().BoolVar(&p.generateStub, "waybar.generate-stub", true, "Generate example CSS stub")
	cmd.Flags().BoolVar(&p.reloadConfig, "waybar.reload", false, "Reload waybar after generation (sends SIGUSR2)")
}

// SetVerbose enables or disables verbose logging for the plugin.
// Implements the output.VerbosePlugin interface.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
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
		return ".config/waybar"
	}
	return filepath.Join(home, ".config", "waybar")
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
	files["tinct-colours.css"] = colorsContent

	// Generate example CSS if requested
	if p.generateStub {
		stubContent, err := p.generateStubCSS()
		if err != nil {
			return nil, fmt.Errorf("failed to generate stub: %w", err)
		}
		files["tinct.css"] = stubContent
	}

	return files, nil
}

// generateColors creates the color definitions CSS file.
func (p *Plugin) generateColors(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("waybar", templates)
	if p.verbose {
		loader.WithVerbose(true, &verboseLogger{out: os.Stderr})
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colours.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read colors template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colours.css.tmpl\n")
	}

	tmpl, err := template.New("colors").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse colors template: %w", err)
	}

	data := p.prepareColorsData(palette)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute colors template: %w", err)
	}

	return buf.Bytes(), nil
}

// generateStubCSS creates an example CSS file showing how to use the colors.
func (p *Plugin) generateStubCSS() ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("waybar", templates)
	if p.verbose {
		loader.WithVerbose(true, &verboseLogger{out: os.Stderr})
	}
	tmplContent, fromCustom, err := loader.Load("tinct.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read stub template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.css.tmpl\n")
	}

	tmpl, err := template.New("stub").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse stub template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("failed to execute stub template: %w", err)
	}

	return buf.Bytes(), nil
}

// ColorsData holds data for the colors template.
type ColorsData struct {
	SourceTheme string            // "dark" or "light"
	Colors      []ColorDefinition // Semantic theme colors
}

// ColorDefinition represents a color definition in GTK CSS format.
type ColorDefinition struct {
	Name string // e.g., "background", "accent1"
	Hex  string // e.g., "#1a1b26"
}

// prepareColorsData converts a categorised palette to Waybar colors data.
func (p *Plugin) prepareColorsData(palette *colour.CategorisedPalette) ColorsData {
	data := ColorsData{
		SourceTheme: palette.ThemeType.String(),
		Colors:      []ColorDefinition{},
	}

	// Map semantic roles to CSS variable names
	roleMapping := map[colour.ColourRole]string{
		colour.RoleBackground:      "background",
		colour.RoleBackgroundMuted: "background-muted",
		colour.RoleForeground:      "foreground",
		colour.RoleForegroundMuted: "foreground-muted",
		colour.RoleAccent1:         "accent1",
		colour.RoleAccent1Muted:    "accent1-muted",
		colour.RoleAccent2:         "accent2",
		colour.RoleAccent2Muted:    "accent2-muted",
		colour.RoleAccent3:         "accent3",
		colour.RoleAccent3Muted:    "accent3-muted",
		colour.RoleAccent4:         "accent4",
		colour.RoleAccent4Muted:    "accent4-muted",
		colour.RoleDanger:          "danger",
		colour.RoleWarning:         "warning",
		colour.RoleSuccess:         "success",
		colour.RoleInfo:            "info",
		colour.RoleNotification:    "notification",
	}

	// Generate semantic color definitions
	for role, name := range roleMapping {
		if color, ok := palette.Get(role); ok {
			data.Colors = append(data.Colors, ColorDefinition{
				Name: name,
				Hex:  color.Hex,
			})
		}
	}

	return data
}

// PreExecute checks if waybar is available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if waybar executable exists on PATH
	_, err = exec.LookPath("waybar")
	if err != nil {
		return true, "waybar executable not found on $PATH", nil
	}

	// Check if config directory exists
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return true, fmt.Sprintf("waybar config directory not found: %s", configDir), nil
	}

	return false, "", nil
}

// PostExecute reloads waybar configuration if requested.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, writtenFiles []string) error {
	if !p.reloadConfig {
		return nil
	}

	// Get waybar process PIDs
	cmd := exec.CommandContext(ctx, "pgrep", "-x", "waybar")
	output, err := cmd.Output()
	if err != nil {
		// Check if the error is because no process was found
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return fmt.Errorf("no running waybar instances found to reload")
			}
		}
		return fmt.Errorf("failed to find waybar processes: %w", err)
	}

	pids := strings.Fields(strings.TrimSpace(string(output)))
	if len(pids) == 0 {
		return fmt.Errorf("no running waybar instances found")
	}

	// Send SIGUSR2 to all waybar instances to reload config
	for _, pid := range pids {
		killCmd := exec.CommandContext(ctx, "kill", "-SIGUSR2", pid)
		if err := killCmd.Run(); err != nil {
			return fmt.Errorf("failed to send reload signal to waybar (PID %s): %w", pid, err)
		}
	}

	return nil
}
