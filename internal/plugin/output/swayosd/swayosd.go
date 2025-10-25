// Package swayosd provides an output plugin for SwayOSD on-screen display colour themes.
package swayosd

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// Plugin implements the output.Plugin interface for SwayOSD.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new SwayOSD output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "swayosd"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate SwayOSD on-screen display colour theme configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "swayosd.output-dir", "", "Output directory (default: ~/.config/swayosd)")
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
		return ".config/swayosd"
	}
	return filepath.Join(home, ".config", "swayosd")
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

	files["style.css"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("swayosd", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("style.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for style.css.tmpl\n")
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

// prepareThemeData converts a categorised palette to SwayOSD theme data.
func (p *Plugin) prepareThemeData(palette *colour.CategorisedPalette) ThemeData {
	return ThemeData{
		SourceTheme:     palette.ThemeType.String(),
		Background:      util.GetColour(palette, colour.RoleBackground, "#1e1e2e"),
		BackgroundMuted: util.GetColour(palette, colour.RoleBackgroundMuted, "#313244"),
		Foreground:      util.GetColour(palette, colour.RoleForeground, "#cdd6f4"),
		ForegroundMuted: util.GetColour(palette, colour.RoleForegroundMuted, "#a6adc8"),
		Accent1:         util.GetColour(palette, colour.RoleAccent1, "#89b4fa"),
		Accent2:         util.GetColour(palette, colour.RoleAccent2, "#f38ba8"),
		Accent3:         util.GetColour(palette, colour.RoleAccent3, "#a6e3a1"),
		Accent4:         util.GetColour(palette, colour.RoleAccent4, "#f9e2af"),
		Danger:          util.GetColour(palette, colour.RoleDanger, "#f38ba8"),
		Warning:         util.GetColour(palette, colour.RoleWarning, "#f9e2af"),
		Success:         util.GetColour(palette, colour.RoleSuccess, "#a6e3a1"),
		Info:            util.GetColour(palette, colour.RoleInfo, "#89b4fa"),
		Notification:    util.GetColour(palette, colour.RoleNotification, "#cba6f7"),
	}
}

// Helper methods for ThemeData to convert colors to CSS RGB/RGBA format

// BackgroundRGB returns background in CSS rgb() format
func (td ThemeData) BackgroundRGB() string {
	return hexToRGB(td.Background)
}

// BackgroundRGBA returns background in CSS rgba() format with alpha
func (td ThemeData) BackgroundRGBA(alpha string) string {
	return hexToRGBA(td.Background, alpha)
}

// BackgroundMutedRGB returns background muted in CSS rgb() format
func (td ThemeData) BackgroundMutedRGB() string {
	return hexToRGB(td.BackgroundMuted)
}

// BackgroundMutedRGBA returns background muted in CSS rgba() format with alpha
func (td ThemeData) BackgroundMutedRGBA(alpha string) string {
	return hexToRGBA(td.BackgroundMuted, alpha)
}

// ForegroundRGB returns foreground in CSS rgb() format
func (td ThemeData) ForegroundRGB() string {
	return hexToRGB(td.Foreground)
}

// ForegroundRGBA returns foreground in CSS rgba() format with alpha
func (td ThemeData) ForegroundRGBA(alpha string) string {
	return hexToRGBA(td.Foreground, alpha)
}

// ForegroundMutedRGB returns foreground muted in CSS rgb() format
func (td ThemeData) ForegroundMutedRGB() string {
	return hexToRGB(td.ForegroundMuted)
}

// ForegroundMutedRGBA returns foreground muted in CSS rgba() format with alpha
func (td ThemeData) ForegroundMutedRGBA(alpha string) string {
	return hexToRGBA(td.ForegroundMuted, alpha)
}

// Accent1RGB returns accent1 in CSS rgb() format
func (td ThemeData) Accent1RGB() string {
	return hexToRGB(td.Accent1)
}

// Accent1RGBA returns accent1 in CSS rgba() format with alpha
func (td ThemeData) Accent1RGBA(alpha string) string {
	return hexToRGBA(td.Accent1, alpha)
}

// Accent2RGB returns accent2 in CSS rgb() format
func (td ThemeData) Accent2RGB() string {
	return hexToRGB(td.Accent2)
}

// Accent2RGBA returns accent2 in CSS rgba() format with alpha
func (td ThemeData) Accent2RGBA(alpha string) string {
	return hexToRGBA(td.Accent2, alpha)
}

// Accent3RGB returns accent3 in CSS rgb() format
func (td ThemeData) Accent3RGB() string {
	return hexToRGB(td.Accent3)
}

// Accent3RGBA returns accent3 in CSS rgba() format with alpha
func (td ThemeData) Accent3RGBA(alpha string) string {
	return hexToRGBA(td.Accent3, alpha)
}

// Accent4RGB returns accent4 in CSS rgb() format
func (td ThemeData) Accent4RGB() string {
	return hexToRGB(td.Accent4)
}

// Accent4RGBA returns accent4 in CSS rgba() format with alpha
func (td ThemeData) Accent4RGBA(alpha string) string {
	return hexToRGBA(td.Accent4, alpha)
}

// DangerRGB returns danger in CSS rgb() format
func (td ThemeData) DangerRGB() string {
	return hexToRGB(td.Danger)
}

// DangerRGBA returns danger in CSS rgba() format with alpha
func (td ThemeData) DangerRGBA(alpha string) string {
	return hexToRGBA(td.Danger, alpha)
}

// WarningRGB returns warning in CSS rgb() format
func (td ThemeData) WarningRGB() string {
	return hexToRGB(td.Warning)
}

// WarningRGBA returns warning in CSS rgba() format with alpha
func (td ThemeData) WarningRGBA(alpha string) string {
	return hexToRGBA(td.Warning, alpha)
}

// SuccessRGB returns success in CSS rgb() format
func (td ThemeData) SuccessRGB() string {
	return hexToRGB(td.Success)
}

// SuccessRGBA returns success in CSS rgba() format with alpha
func (td ThemeData) SuccessRGBA(alpha string) string {
	return hexToRGBA(td.Success, alpha)
}

// InfoRGB returns info in CSS rgb() format
func (td ThemeData) InfoRGB() string {
	return hexToRGB(td.Info)
}

// InfoRGBA returns info in CSS rgba() format with alpha
func (td ThemeData) InfoRGBA(alpha string) string {
	return hexToRGBA(td.Info, alpha)
}

// hexToRGB converts #RRGGBB to CSS rgb(r, g, b) format
func hexToRGB(hex string) string {
	if len(hex) == 7 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return hex // Return as-is if invalid
	}

	r := hexToDec(hex[0:2])
	g := hexToDec(hex[2:4])
	b := hexToDec(hex[4:6])

	return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
}

// hexToRGBA converts #RRGGBB to CSS rgba(r, g, b, a) format
// alpha is a float string like "0.85" or "1.0"
func hexToRGBA(hex string, alpha string) string {
	if len(hex) == 7 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return hex // Return as-is if invalid
	}

	r := hexToDec(hex[0:2])
	g := hexToDec(hex[2:4])
	b := hexToDec(hex[4:6])

	// Validate alpha is a valid float
	alphaFloat, err := strconv.ParseFloat(alpha, 64)
	if err != nil {
		alphaFloat = 1.0
	}

	return fmt.Sprintf("rgba(%d, %d, %d, %.2f)", r, g, b, alphaFloat)
}

// hexToDec converts a 2-character hex string to decimal
func hexToDec(hex string) int {
	var result int
	fmt.Sscanf(hex, "%x", &result)
	return result
}

// PreExecute checks if swayosd is available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if swayosd-server executable exists on PATH
	_, err = exec.LookPath("swayosd-server")
	if err != nil {
		return true, "swayosd-server executable not found on $PATH", nil
	}

	// Check if config directory exists (create it if not)
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// For swayosd, we can create the directory since it's straightforward
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return true, fmt.Sprintf("swayosd config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}
