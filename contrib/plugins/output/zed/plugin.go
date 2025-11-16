package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/jmylchreest/tinct/pkg/colour"
	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
	tincttemplate "github.com/jmylchreest/tinct/pkg/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// Plugin implements the tinct OutputPlugin interface for Zed editor themes.
type Plugin struct {
	version   string
	commit    string
	date      string
	outputDir string
}

// Generate creates a Zed theme JSON file from the palette data.
func (p *Plugin) Generate(ctx context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
	// Convert PaletteData to ThemeData for template rendering.
	themeData := convertPaletteDataToThemeData(palette)

	// Load template.
	loader := tincttemplate.New("zed", templatesFS)
	tmplContent, _, err := loader.Load("templates/theme.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Parse template with helper functions.
	tmpl, err := template.New("theme").
		Funcs(tincttemplate.TemplateFuncs()).
		Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Return generated file.
	files := make(map[string][]byte)
	files["tinct.json"] = buf.Bytes()

	return files, nil
}

// PreExecute checks if the output directory exists and creates it if needed.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Determine output directory.
	outputDir := p.outputDir
	if outputDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false, "", fmt.Errorf("failed to get home directory: %w", err)
		}
		outputDir = filepath.Join(home, ".config", "zed", "themes")
	}

	// Create directory if it doesn't exist.
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return false, "", fmt.Errorf("failed to create output directory: %w", err)
	}

	return false, "", nil
}

// PostExecute is a no-op for Zed (themes are loaded automatically).
func (p *Plugin) PostExecute(ctx context.Context, files []string) error {
	return nil
}

// GetMetadata returns plugin metadata.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "zed",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate Zed editor theme files",
		PluginProtocol:  string(tinctplugin.PluginTypeGoPlugin),
	}
}

// GetFlagHelp returns help for plugin-specific flags.
func (p *Plugin) GetFlagHelp() []tinctplugin.FlagHelp {
	return []tinctplugin.FlagHelp{
		{
			Name:        "output-dir",
			Shorthand:   "o",
			Type:        "string",
			Default:     "~/.config/zed/themes",
			Description: "Output directory for Zed theme files",
			Required:    false,
		},
	}
}

// convertPaletteDataToThemeData converts the protocol PaletteData to colour.ThemeData.
func convertPaletteDataToThemeData(palette tinctplugin.PaletteData) *colour.ThemeData {
	// Build color map and indexed array.
	colors := make(map[colour.Role]colour.ColorValue)
	indexed := make([]colour.ColorValue, 0, len(palette.AllColours))

	for roleName, cc := range palette.Colours {
		role := colour.Role(roleName)
		rgba := colour.RGBA{R: cc.RGB.R, G: cc.RGB.G, B: cc.RGB.B, A: 255}
		cv := colour.NewColorValue(rgba, role, cc.Index)
		colors[role] = cv
	}

	for _, cc := range palette.AllColours {
		role := colour.Role(cc.Role)
		rgba := colour.RGBA{R: cc.RGB.R, G: cc.RGB.G, B: cc.RGB.B, A: 255}
		cv := colour.NewColorValue(rgba, role, cc.Index)
		indexed = append(indexed, cv)
	}

	// Determine theme type.
	var themeType colour.ThemeType
	if palette.ThemeType == "dark" {
		themeType = colour.ThemeDark
	} else {
		themeType = colour.ThemeLight
	}

	// Create palette helper.
	helper := colour.NewPaletteHelper(themeType, colors, indexed)

	// Wrap in ThemeData.
	return &colour.ThemeData{
		PaletteHelper: helper,
		WallpaperPath: "",
		ThemeName:     "Tinct",
	}
}
