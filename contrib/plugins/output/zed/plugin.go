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

// detectZedPaths finds all installed Zed configurations (standard and Flatpak).
func detectZedPaths() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	var paths []string

	// Check for standard Zed installation.
	standardPath := filepath.Join(home, ".config", "zed", "themes")
	if _, err := os.Stat(filepath.Join(home, ".config", "zed")); err == nil {
		paths = append(paths, standardPath)
	}

	// Check for Flatpak Zed installation.
	flatpakPath := filepath.Join(home, ".var", "app", "dev.zed.Zed", "config", "zed", "themes")
	if _, err := os.Stat(filepath.Join(home, ".var", "app", "dev.zed.Zed", "config", "zed")); err == nil {
		paths = append(paths, flatpakPath)
	}

	return paths, nil
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

	// Determine output directories.
	var outputDirs []string
	if p.outputDir != "" {
		// User specified a custom output directory - use only that.
		outputDirs = []string{p.outputDir}
	} else {
		// No custom directory specified - detect Zed installation paths.
		detectedPaths, err := detectZedPaths()
		if err != nil {
			return nil, err
		}

		if len(detectedPaths) == 0 {
			// Fallback to default if no installations detected.
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			outputDirs = []string{filepath.Join(home, ".config", "zed", "themes")}
		} else {
			outputDirs = detectedPaths
		}
	}

	// Generate files for all output directories.
	files := make(map[string][]byte)
	for _, outputDir := range outputDirs {
		// Create themes directory if it doesn't exist.
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create themes directory %q: %w", outputDir, err)
		}

		// Add file with full path.
		fullPath := filepath.Join(outputDir, "tinct.json")
		files[fullPath] = buf.Bytes()
	}

	return files, nil
}

// PreExecute checks if Zed is installed (config directory exists).
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Detect Zed installations.
	detectedPaths, err := detectZedPaths()
	if err != nil {
		return false, "", err
	}

	// If no installations found, skip this plugin.
	if len(detectedPaths) == 0 {
		return true, "Zed not installed (config directory does not exist)", nil
	}

	return false, "", nil
}

// PostExecute is no longer needed - file watcher compatibility is now handled
// by disabling backups in the file write logic.
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
			Default:     "auto-detected",
			Description: "Output directory for Zed theme files (default: auto-detects standard and Flatpak installations)",
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
