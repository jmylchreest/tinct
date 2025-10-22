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
		stubContent, err := p.generateStubConfig()
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
func (p *Plugin) generateStubConfig() ([]byte, error) {
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
	Colours     []ColourVariable // Semantic theme colours
	Indexed     []ColourVariable // Indexed colours (colour0, colour1, etc.)
}

// ColourVariable represents a colour variable in Hyprland format.
type ColourVariable struct {
	Name     string // e.g., "background", "accent1"
	RGBHex   string // e.g., "f5e0dc" for rgb() function
	RGBDecim string // e.g., "245,224,220" for rgba() function
}

// prepareThemeData converts a categorised palette to Hyprland theme data.
func (p *Plugin) prepareThemeData(palette *colour.CategorisedPalette) ThemeData {
	data := ThemeData{
		SourceTheme: palette.ThemeType.String(),
		Colours:     []ColourVariable{},
		Indexed:     []ColourVariable{},
	}

	// Map semantic roles to variable names
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

	// Generate semantic colour variables
	for role, name := range roleMapping {
		if color, ok := palette.Get(role); ok {
			data.Colours = append(data.Colours, ColourVariable{
				Name:     name,
				RGBHex:   stripHash(color.Hex),
				RGBDecim: fmt.Sprintf("%d,%d,%d", color.RGB.R, color.RGB.G, color.RGB.B),
			})
		}
	}

	// Add indexed colour variables (colour0, colour1, etc.)
	for _, color := range palette.AllColours {
		data.Indexed = append(data.Indexed, ColourVariable{
			Name:     fmt.Sprintf("colour%d", color.Index),
			RGBHex:   stripHash(color.Hex),
			RGBDecim: fmt.Sprintf("%d,%d,%d", color.RGB.R, color.RGB.G, color.RGB.B),
		})
	}

	return data
}

// stripHash removes the # prefix from a hex colour string.
func stripHash(hex string) string {
	return strings.TrimPrefix(hex, "#")
}
