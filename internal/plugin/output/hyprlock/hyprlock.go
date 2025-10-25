// Package hyprlock provides an output plugin for Hyprlock screen lock colour themes.
package hyprlock

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

// Plugin implements the output.Plugin interface for Hyprlock.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Hyprlock output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "hyprlock"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Hyprlock screen lock colour theme configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "hyprlock.output-dir", "", "Output directory (default: ~/.config/hypr)")
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
		return ".config/hypr"
	}
	return filepath.Join(home, ".config", "hypr")
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

	files["tinct-hyprlock.conf"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("hyprlock", templates)
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
	Notification    string
}

// prepareThemeData converts a categorised palette to Hyprlock theme data.
func (p *Plugin) prepareThemeData(palette *colour.CategorisedPalette) ThemeData {
	return ThemeData{
		SourceTheme:     palette.ThemeType.String(),
		Background:      util.GetColour(palette, colour.RoleBackground, "#11121d"),
		BackgroundMuted: util.GetColour(palette, colour.RoleBackgroundMuted, "#1a1b26"),
		Foreground:      util.GetColour(palette, colour.RoleForeground, "#dddddd"),
		ForegroundMuted: util.GetColour(palette, colour.RoleForegroundMuted, "#999999"),
		Accent1:         util.GetColour(palette, colour.RoleAccent1, "#9fa8cd"),
		Accent2:         util.GetColour(palette, colour.RoleAccent2, "#7aa2f7"),
		Accent3:         util.GetColour(palette, colour.RoleAccent3, "#bb9af7"),
		Accent4:         util.GetColour(palette, colour.RoleAccent4, "#7dcfff"),
		Danger:          util.GetColour(palette, colour.RoleDanger, "#f7768e"),
		Warning:         util.GetColour(palette, colour.RoleWarning, "#e0af68"),
		Success:         util.GetColour(palette, colour.RoleSuccess, "#9ece6a"),
		Info:            util.GetColour(palette, colour.RoleInfo, "#7aa2f7"),
		Notification:    util.GetColour(palette, colour.RoleNotification, "#bb9af7"),
	}
}

// Helper methods for ThemeData to convert colors to RGB/RGBA format

// BackgroundRGB returns background in RGB format (without #)
func (td ThemeData) BackgroundRGB() string {
	return hexToRGB(td.Background)
}

func (td ThemeData) BackgroundMutedRGB() string {
	return hexToRGB(td.BackgroundMuted)
}

func (td ThemeData) ForegroundRGB() string {
	return hexToRGB(td.Foreground)
}

func (td ThemeData) ForegroundMutedRGB() string {
	return hexToRGB(td.ForegroundMuted)
}

func (td ThemeData) Accent1RGB() string {
	return hexToRGB(td.Accent1)
}

func (td ThemeData) Accent2RGB() string {
	return hexToRGB(td.Accent2)
}

func (td ThemeData) Accent3RGB() string {
	return hexToRGB(td.Accent3)
}

func (td ThemeData) Accent4RGB() string {
	return hexToRGB(td.Accent4)
}

func (td ThemeData) DangerRGB() string {
	return hexToRGB(td.Danger)
}

func (td ThemeData) WarningRGB() string {
	return hexToRGB(td.Warning)
}

func (td ThemeData) SuccessRGB() string {
	return hexToRGB(td.Success)
}

func (td ThemeData) InfoRGB() string {
	return hexToRGB(td.Info)
}

// RGBA variants with alpha channel

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

// hexToRGB converts #RRGGBB to RRGGBB format (without #)
func hexToRGB(hex string) string {
	if len(hex) == 7 && hex[0] == '#' {
		return hex[1:]
	}
	if len(hex) == 6 {
		return hex
	}
	return hex
}

// hexToRGBA converts #RRGGBB to RR, GG, BB, AA format for Hyprland rgba()
func hexToRGBA(hex string, alpha string) string {
	rgb := hexToRGB(hex)
	if len(rgb) != 6 {
		return rgb + ", 255" // fallback
	}

	// Parse hex to decimal
	r := hexToDec(rgb[0:2])
	g := hexToDec(rgb[2:4])
	b := hexToDec(rgb[4:6])

	// Convert alpha from hex (00-ff) to decimal (0-255)
	a := hexToDec(alpha)

	// Hyprland uses 0-1 float for alpha, so convert 0-255 to float
	aFloat := float64(a) / 255.0

	return fmt.Sprintf("%d, %d, %d, %.2f", r, g, b, aFloat)
}

// hexToDec converts a 2-character hex string to decimal
func hexToDec(hex string) int {
	var result int
	fmt.Sscanf(hex, "%x", &result)
	return result
}

// PreExecute checks if hyprlock is available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if hyprlock executable exists on PATH
	_, err = exec.LookPath("hyprlock")
	if err != nil {
		return true, "hyprlock executable not found on $PATH", nil
	}

	// Check if config directory exists (create it if not, as it's straightforward)
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// For hyprlock, we can create the directory since it's straightforward
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return true, fmt.Sprintf("hypr config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}
