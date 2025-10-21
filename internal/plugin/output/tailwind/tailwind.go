// Package tailwind provides a Tailwind CSS / shadcn/ui output plugin.
package tailwind

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/spf13/cobra"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for Tailwind CSS.
type Plugin struct {
	format    string // "css" or "config"
	outputDir string
}

// New creates a new Tailwind CSS output plugin.
func New() *Plugin {
	return &Plugin{
		format:    "css",
		outputDir: "",
	}
}

// NewWithFormat creates a new Tailwind CSS output plugin with a specific format.
func NewWithFormat(format string) *Plugin {
	return &Plugin{
		format:    format,
		outputDir: "",
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "tailwind"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Tailwind CSS / shadcn/ui theme configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.format, "tailwind.format", "css", "Output format (css or config)")
	cmd.Flags().StringVar(&p.outputDir, "tailwind.output-dir", "", "Output directory (default: current directory)")
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	if p.format != "css" && p.format != "config" {
		return fmt.Errorf("invalid format: %s (must be 'css' or 'config')", p.format)
	}
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	// Default paths based on format
	if p.format == "config" {
		return "."
	}

	// For CSS, try to detect if we're in a Next.js project
	if _, err := os.Stat("app"); err == nil {
		return "app"
	}
	if _, err := os.Stat("src"); err == nil {
		return filepath.Join("src", "app")
	}

	return "."
}

// Generate creates the Tailwind CSS configuration from the palette.
func (p *Plugin) Generate(palette *colour.CategorisedPalette) (map[string][]byte, error) {
	if palette == nil {
		return nil, fmt.Errorf("palette cannot be nil")
	}

	files := make(map[string][]byte)

	if p.format == "config" {
		content, err := p.generateConfig(palette)
		if err != nil {
			return nil, err
		}
		files["tailwind.config.js"] = content
	} else {
		content, err := p.generateCSS(palette)
		if err != nil {
			return nil, err
		}
		files["globals.css"] = content
	}

	return files, nil
}

// generateCSS creates a CSS file with shadcn/ui CSS variables.
func (p *Plugin) generateCSS(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template from embedded filesystem
	tmplContent, err := templates.ReadFile("globals.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read CSS template: %w", err)
	}

	tmpl, err := template.New("globals.css").Funcs(templateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSS template: %w", err)
	}

	data := prepareCSSData(palette)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute CSS template: %w", err)
	}

	return buf.Bytes(), nil
}

// generateConfig creates a tailwind.config.js file.
func (p *Plugin) generateConfig(palette *colour.CategorisedPalette) ([]byte, error) {
	// Load template from embedded filesystem
	tmplContent, err := templates.ReadFile("tailwind.config.js.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read config template: %w", err)
	}

	tmpl, err := template.New("tailwind.config.js").Funcs(templateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config template: %w", err)
	}

	data := prepareConfigData(palette)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute config template: %w", err)
	}

	return buf.Bytes(), nil
}

// CSSData holds data for CSS template.
type CSSData struct {
	ThemeType string
	Light     CSSTheme
	Dark      CSSTheme
	HasCustom bool
	Custom    []CSSCustomColour
}

// CSSTheme holds CSS variables for a theme.
type CSSTheme struct {
	Background            string
	Foreground            string
	Card                  string
	CardForeground        string
	Popover               string
	PopoverForeground     string
	Primary               string
	PrimaryForeground     string
	Secondary             string
	SecondaryForeground   string
	Muted                 string
	MutedForeground       string
	Accent                string
	AccentForeground      string
	Destructive           string
	DestructiveForeground string
	Border                string
	Input                 string
	Ring                  string
}

// CSSCustomColour represents a custom colour variable.
type CSSCustomColour struct {
	Name  string
	Value string
}

// ConfigData holds data for config template.
type ConfigData struct {
	Colors map[string]any
}

// prepareCSSData converts a categorised palette to CSS template data.
func prepareCSSData(palette *colour.CategorisedPalette) CSSData {
	data := CSSData{
		ThemeType: palette.ThemeType.String(),
		HasCustom: true,
	}

	// Get colors from palette
	bg, hasBg := palette.Get(colour.RoleBackground)
	fg, hasFg := palette.Get(colour.RoleForeground)
	bgMuted, hasBgMuted := palette.Get(colour.RoleBackgroundMuted)
	fgMuted, hasFgMuted := palette.Get(colour.RoleForegroundMuted)
	primary, hasPrimary := palette.Get(colour.RoleAccent1)
	secondary, hasSecondary := palette.Get(colour.RoleAccent2)
	accent, hasAccent := palette.Get(colour.RoleAccent3)
	destructive, hasDestructive := palette.Get(colour.RoleDanger)

	// Determine if we're starting with a dark or light theme
	isDark := palette.ThemeType == colour.ThemeDark

	if isDark {
		// Dark theme as primary
		data.Dark = buildCSSTheme(bg, fg, bgMuted, fgMuted, primary, secondary, accent, destructive,
			hasBg, hasFg, hasBgMuted, hasFgMuted, hasPrimary, hasSecondary, hasAccent, hasDestructive, true)
		data.Light = buildCSSTheme(bg, fg, bgMuted, fgMuted, primary, secondary, accent, destructive,
			hasBg, hasFg, hasBgMuted, hasFgMuted, hasPrimary, hasSecondary, hasAccent, hasDestructive, false)
	} else {
		// Light theme as primary
		data.Light = buildCSSTheme(bg, fg, bgMuted, fgMuted, primary, secondary, accent, destructive,
			hasBg, hasFg, hasBgMuted, hasFgMuted, hasPrimary, hasSecondary, hasAccent, hasDestructive, true)
		data.Dark = buildCSSTheme(bg, fg, bgMuted, fgMuted, primary, secondary, accent, destructive,
			hasBg, hasFg, hasBgMuted, hasFgMuted, hasPrimary, hasSecondary, hasAccent, hasDestructive, false)
	}

	// Add custom semantic colours
	data.Custom = []CSSCustomColour{}
	if warning, ok := palette.Get(colour.RoleWarning); ok {
		data.Custom = append(data.Custom, CSSCustomColour{
			Name:  "warning",
			Value: toHSL(warning),
		})
	}
	if success, ok := palette.Get(colour.RoleSuccess); ok {
		data.Custom = append(data.Custom, CSSCustomColour{
			Name:  "success",
			Value: toHSL(success),
		})
	}
	if info, ok := palette.Get(colour.RoleInfo); ok {
		data.Custom = append(data.Custom, CSSCustomColour{
			Name:  "info",
			Value: toHSL(info),
		})
	}
	if notification, ok := palette.Get(colour.RoleNotification); ok {
		data.Custom = append(data.Custom, CSSCustomColour{
			Name:  "notification",
			Value: toHSL(notification),
		})
	}

	return data
}

// buildCSSTheme creates a CSS theme from colors.
func buildCSSTheme(
	bg, fg, bgMuted, fgMuted, primary, secondary, accent, destructive colour.CategorisedColour,
	hasBg, hasFg, hasBgMuted, hasFgMuted, hasPrimary, hasSecondary, hasAccent, hasDestructive bool,
	isPrimaryTheme bool,
) CSSTheme {
	theme := CSSTheme{}

	if hasBg {
		theme.Background = toHSL(bg)
		theme.Card = toHSL(bg)
		theme.Popover = toHSL(bg)
	}

	if hasFg {
		theme.Foreground = toHSL(fg)
		theme.CardForeground = toHSL(fg)
		theme.PopoverForeground = toHSL(fg)
	}

	if hasBgMuted {
		theme.Muted = toHSL(bgMuted)
		theme.Border = toHSL(bgMuted)
		theme.Input = toHSL(bgMuted)
	}

	if hasFgMuted {
		theme.MutedForeground = toHSL(fgMuted)
	}

	if hasPrimary {
		theme.Primary = toHSL(primary)
		if hasFg {
			theme.PrimaryForeground = toHSL(fg)
		}
		theme.Ring = toHSL(primary)
	}

	if hasSecondary {
		theme.Secondary = toHSL(secondary)
		if hasFg {
			theme.SecondaryForeground = toHSL(fg)
		}
	}

	if hasAccent {
		theme.Accent = toHSL(accent)
		if hasFg {
			theme.AccentForeground = toHSL(fg)
		}
	}

	if hasDestructive {
		theme.Destructive = toHSL(destructive)
		if hasFg {
			theme.DestructiveForeground = toHSL(fg)
		}
	}

	return theme
}

// prepareConfigData converts a categorised palette to config template data.
func prepareConfigData(palette *colour.CategorisedPalette) ConfigData {
	colors := make(map[string]any)

	// Add semantic colours
	if danger, ok := palette.Get(colour.RoleDanger); ok {
		colors["destructive"] = danger.Hex
	}
	if warning, ok := palette.Get(colour.RoleWarning); ok {
		colors["warning"] = warning.Hex
	}
	if success, ok := palette.Get(colour.RoleSuccess); ok {
		colors["success"] = success.Hex
	}
	if info, ok := palette.Get(colour.RoleInfo); ok {
		colors["info"] = info.Hex
	}

	// Add accent colours
	accents := make(map[string]string)
	if accent1, ok := palette.Get(colour.RoleAccent1); ok {
		accents["1"] = accent1.Hex
	}
	if accent2, ok := palette.Get(colour.RoleAccent2); ok {
		accents["2"] = accent2.Hex
	}
	if accent3, ok := palette.Get(colour.RoleAccent3); ok {
		accents["3"] = accent3.Hex
	}
	if accent4, ok := palette.Get(colour.RoleAccent4); ok {
		accents["4"] = accent4.Hex
	}
	if len(accents) > 0 {
		colors["accent"] = accents
	}

	return ConfigData{Colors: colors}
}

// toHSL converts a categorised colour to HSL format for CSS variables.
// Format: "hue saturation% lightness%" (e.g., "222.2 47.4% 11.2%")
func toHSL(cc colour.CategorisedColour) string {
	return fmt.Sprintf("%.1f %.1f%% %.1f%%", cc.Hue, cc.Saturation*100, cc.Luminance*100)
}

// templateFuncs returns template functions.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"toHSL": toHSL,
	}
}
