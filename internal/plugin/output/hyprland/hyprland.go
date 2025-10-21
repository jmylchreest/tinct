// Package hyprland provides an output plugin for Hyprland window manager colour themes.
package hyprland

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/spf13/cobra"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for Hyprland.
type Plugin struct {
	outputDir    string
	generateStub bool
	stubPath     string
}

// New creates a new Hyprland output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir:    "",
		generateStub: true,
		stubPath:     "",
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "hyprland"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Hyprland colour theme configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "hyprland.output-dir", "", "Output directory (default: ~/.config/hypr/themes)")
	cmd.Flags().BoolVar(&p.generateStub, "hyprland.generate-stub", true, "Generate example config stub")
	cmd.Flags().StringVar(&p.stubPath, "hyprland.stub-path", "", "Custom path for stub file")
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
		return ".config/hypr/themes"
	}
	return filepath.Join(home, ".config", "hypr", "themes")
}

// Generate creates the theme file and optional stub configuration.
// Returns map of filename -> content
func (p *Plugin) Generate(palette *colour.CategorisedPalette) (map[string][]byte, error) {
	if palette == nil {
		return nil, fmt.Errorf("palette cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate main theme file
	themeContent, err := p.generateTheme(palette)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	themeFile := "tinct-colours.conf"
	files[themeFile] = themeContent

	// Generate stub if requested
	if p.generateStub {
		stubContent, err := p.generateStubConfig(palette)
		if err != nil {
			return nil, fmt.Errorf("failed to generate stub: %w", err)
		}

		stubFile := "tinct.conf"
		if p.stubPath != "" {
			stubFile = filepath.Base(p.stubPath)
		}
		files[stubFile] = stubContent
	}

	return files, nil
}

// generateTheme creates the main theme configuration file with colour variables.
func (p *Plugin) generateTheme(palette *colour.CategorisedPalette) ([]byte, error) {
	tmplContent, err := templates.ReadFile("tinct-colours.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
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

// generateStubConfig creates an example configuration file showing how to use the theme.
func (p *Plugin) generateStubConfig(palette *colour.CategorisedPalette) ([]byte, error) {
	tmplContent, err := templates.ReadFile("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read example template: %w", err)
	}

	tmpl, err := template.New("example").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse example template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("failed to execute example template: %w", err)
	}

	return buf.Bytes(), nil
}

// ThemeData holds data for the theme template.
type ThemeData struct {
	SourceTheme string           // "dark" or "light" - the source theme from extraction
	Colours     []ColourVariable // Primary theme colours (semantic roles from detected theme)
	Indexed     []ColourVariable // Indexed colours (colour0, colour1, etc.)

	// Optional extended palettes for templates that want explicit light/dark access
	Light ThemeColours // Light theme palette (available but optional)
	Dark  ThemeColours // Dark theme palette (available but optional)
}

// ThemeColours holds colour variables for a specific theme.
type ThemeColours struct {
	Background      ColourVariable
	BackgroundMuted ColourVariable
	Foreground      ColourVariable
	ForegroundMuted ColourVariable
	Accent1         ColourVariable
	Accent2         ColourVariable
	Accent3         ColourVariable
	Accent4         ColourVariable
	Danger          ColourVariable
	Warning         ColourVariable
	Success         ColourVariable
	Info            ColourVariable
	Notification    ColourVariable
}

// ColourVariable represents a colour variable in Hyprland format.
type ColourVariable struct {
	Name     string // e.g., "background", "accent1"
	RGBHex   string // e.g., "f5e0dc" for rgb() function
	RGBDecim string // e.g., "245,224,220" for rgba() function
}

// prepareThemeData converts a categorised palette to Hyprland theme data.
// It generates both light and dark themes, creating the opposite theme if needed.
func (p *Plugin) prepareThemeData(palette *colour.CategorisedPalette) ThemeData {
	data := ThemeData{
		SourceTheme: palette.ThemeType.String(),
		Colours:     []ColourVariable{},
		Indexed:     []ColourVariable{},
	}

	// Map our roles to Hyprland variable names (using Tinct semantic names in camelCase)
	roleMapping := map[colour.ColourRole]string{
		colour.RoleBackground:      "background",
		colour.RoleBackgroundMuted: "backgroundMuted",
		colour.RoleForeground:      "foreground",
		colour.RoleForegroundMuted: "foregroundMuted",
		colour.RoleAccent1:         "accent1",
		colour.RoleAccent2:         "accent2",
		colour.RoleAccent3:         "accent3",
		colour.RoleAccent4:         "accent4",
		colour.RoleDanger:          "danger",
		colour.RoleWarning:         "warning",
		colour.RoleSuccess:         "success",
		colour.RoleInfo:            "info",
		colour.RoleNotification:    "notification",
	}

	// Generate primary theme colour variables (detected theme)
	for role, name := range roleMapping {
		if color, ok := palette.Get(role); ok {
			rgb := stripHash(color.Hex)
			data.Colours = append(data.Colours, ColourVariable{
				Name:     name,
				RGBHex:   rgb,
				RGBDecim: fmt.Sprintf("%d,%d,%d", color.RGB.R, color.RGB.G, color.RGB.B),
			})
		}
	}

	// Add indexed colour variables (colour0, colour1, etc.) for all colours in palette
	for _, color := range palette.AllColours {
		rgb := stripHash(color.Hex)
		data.Indexed = append(data.Indexed, ColourVariable{
			Name:     fmt.Sprintf("colour%d", color.Index),
			RGBHex:   rgb,
			RGBDecim: fmt.Sprintf("%d,%d,%d", color.RGB.R, color.RGB.G, color.RGB.B),
		})
	}

	// Generate extended light and dark palettes (optional for advanced templates)
	if palette.ThemeType == colour.ThemeDark {
		data.Dark = paletteToThemeColours(palette)
		data.Light = generateInvertedThemeColours(palette)
	} else {
		data.Light = paletteToThemeColours(palette)
		data.Dark = generateInvertedThemeColours(palette)
	}

	return data
}

// paletteToThemeColours converts a CategorisedPalette to ThemeColours.
func paletteToThemeColours(palette *colour.CategorisedPalette) ThemeColours {
	theme := ThemeColours{}

	if c, ok := palette.Get(colour.RoleBackground); ok {
		theme.Background = toColourVariable("background", c)
	}
	if c, ok := palette.Get(colour.RoleBackgroundMuted); ok {
		theme.BackgroundMuted = toColourVariable("backgroundMuted", c)
	}
	if c, ok := palette.Get(colour.RoleForeground); ok {
		theme.Foreground = toColourVariable("foreground", c)
	}
	if c, ok := palette.Get(colour.RoleForegroundMuted); ok {
		theme.ForegroundMuted = toColourVariable("foregroundMuted", c)
	}
	if c, ok := palette.Get(colour.RoleAccent1); ok {
		theme.Accent1 = toColourVariable("accent1", c)
	}
	if c, ok := palette.Get(colour.RoleAccent2); ok {
		theme.Accent2 = toColourVariable("accent2", c)
	}
	if c, ok := palette.Get(colour.RoleAccent3); ok {
		theme.Accent3 = toColourVariable("accent3", c)
	}
	if c, ok := palette.Get(colour.RoleAccent4); ok {
		theme.Accent4 = toColourVariable("accent4", c)
	}
	if c, ok := palette.Get(colour.RoleDanger); ok {
		theme.Danger = toColourVariable("danger", c)
	}
	if c, ok := palette.Get(colour.RoleWarning); ok {
		theme.Warning = toColourVariable("warning", c)
	}
	if c, ok := palette.Get(colour.RoleSuccess); ok {
		theme.Success = toColourVariable("success", c)
	}
	if c, ok := palette.Get(colour.RoleInfo); ok {
		theme.Info = toColourVariable("info", c)
	}
	if c, ok := palette.Get(colour.RoleNotification); ok {
		theme.Notification = toColourVariable("notification", c)
	}

	return theme
}

// generateInvertedThemeColours creates the opposite theme (light->dark or dark->light).
func generateInvertedThemeColours(palette *colour.CategorisedPalette) ThemeColours {
	theme := ThemeColours{}

	if c, ok := palette.Get(colour.RoleBackground); ok {
		theme.Background = toColourVariableInverted("background", c)
	}
	if c, ok := palette.Get(colour.RoleBackgroundMuted); ok {
		theme.BackgroundMuted = toColourVariableInverted("backgroundMuted", c)
	}
	if c, ok := palette.Get(colour.RoleForeground); ok {
		theme.Foreground = toColourVariableInverted("foreground", c)
	}
	if c, ok := palette.Get(colour.RoleForegroundMuted); ok {
		theme.ForegroundMuted = toColourVariableInverted("foregroundMuted", c)
	}
	if c, ok := palette.Get(colour.RoleAccent1); ok {
		theme.Accent1 = toColourVariableInvertedModerate("accent1", c)
	}
	if c, ok := palette.Get(colour.RoleAccent2); ok {
		theme.Accent2 = toColourVariableInvertedModerate("accent2", c)
	}
	if c, ok := palette.Get(colour.RoleAccent3); ok {
		theme.Accent3 = toColourVariableInvertedModerate("accent3", c)
	}
	if c, ok := palette.Get(colour.RoleAccent4); ok {
		theme.Accent4 = toColourVariableInvertedModerate("accent4", c)
	}
	if c, ok := palette.Get(colour.RoleDanger); ok {
		theme.Danger = toColourVariableInvertedModerate("danger", c)
	}
	if c, ok := palette.Get(colour.RoleWarning); ok {
		theme.Warning = toColourVariableInvertedModerate("warning", c)
	}
	if c, ok := palette.Get(colour.RoleSuccess); ok {
		theme.Success = toColourVariableInvertedModerate("success", c)
	}
	if c, ok := palette.Get(colour.RoleInfo); ok {
		theme.Info = toColourVariableInvertedModerate("info", c)
	}
	if c, ok := palette.Get(colour.RoleNotification); ok {
		theme.Notification = toColourVariableInvertedModerate("notification", c)
	}

	return theme
}

// toColourVariable converts a CategorisedColour to a ColourVariable.
func toColourVariable(name string, c colour.CategorisedColour) ColourVariable {
	return ColourVariable{
		Name:     name,
		RGBHex:   stripHash(c.Hex),
		RGBDecim: fmt.Sprintf("%d,%d,%d", c.RGB.R, c.RGB.G, c.RGB.B),
	}
}

// toColourVariableInverted converts a colour with inverted luminance.
func toColourVariableInverted(name string, c colour.CategorisedColour) ColourVariable {
	invertedLum := 1.0 - c.Luminance
	rgb := colour.HSLToRGB(c.Hue, c.Saturation, invertedLum)
	return ColourVariable{
		Name:     name,
		RGBHex:   stripHash(rgb.Hex()),
		RGBDecim: fmt.Sprintf("%d,%d,%d", rgb.R, rgb.G, rgb.B),
	}
}

// toColourVariableInvertedModerate inverts luminance but keeps it in a moderate range.
func toColourVariableInvertedModerate(name string, c colour.CategorisedColour) ColourVariable {
	var invertedLum float64
	if c.Luminance < 0.5 {
		// Dark color -> make it lighter (but not too light)
		invertedLum = 0.5 + (0.5-c.Luminance)*0.6
	} else {
		// Light color -> make it darker (but not too dark)
		invertedLum = 0.5 - (c.Luminance-0.5)*0.6
	}
	rgb := colour.HSLToRGB(c.Hue, c.Saturation, invertedLum)
	return ColourVariable{
		Name:     name,
		RGBHex:   stripHash(rgb.Hex()),
		RGBDecim: fmt.Sprintf("%d,%d,%d", rgb.R, rgb.G, rgb.B),
	}
}

// stripHash removes the # prefix from a hex colour string.
func stripHash(hex string) string {
	return strings.TrimPrefix(hex, "#")
}
