package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
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
func (p *Plugin) Generate(_ context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
	// Load and parse template once
	tmpl, err := p.loadTemplate(palette.Verbose)
	if err != nil {
		return nil, err
	}

	// Generate theme data for primary palette
	primaryThemeData := convertPaletteDataToThemeData(palette)
	primaryThemeObj, err := p.executeTemplate(tmpl, primaryThemeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate primary theme: %w", err)
	}

	// Check if we have dual-theme support
	if palette.AlternateTheme != nil {
		if palette.Verbose {
			fmt.Fprintf(os.Stderr, "   Generating dual-theme (light + dark)\n")
		}

		// Generate theme data for alternate palette
		alternatePaletteData := tinctplugin.PaletteData{
			Colours:    palette.AlternateTheme.Colours,
			AllColours: palette.AlternateTheme.AllColours,
			ThemeType:  palette.AlternateTheme.ThemeType,
			Verbose:    palette.Verbose,
		}
		alternateThemeData := convertPaletteDataToThemeData(alternatePaletteData)
		alternateThemeObj, err := p.executeTemplate(tmpl, alternateThemeData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate alternate theme: %w", err)
		}

		// Combine both themes
		combined := map[string]any{
			"$schema": "https://zed.dev/schema/themes/v0.2.0.json",
			"name":    "Tinct",
			"author":  "Tinct Color Generator",
			"themes":  []any{primaryThemeObj, alternateThemeObj},
		}

		combinedJSON, err := json.MarshalIndent(combined, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal combined theme: %w", err)
		}

		return p.writeThemeFiles(combinedJSON)
	}

	// Single theme mode - marshal the full theme file
	singleTheme := map[string]any{
		"$schema": "https://zed.dev/schema/themes/v0.2.0.json",
		"name":    "Tinct",
		"author":  "Tinct Color Generator",
		"themes":  []any{primaryThemeObj},
	}

	singleJSON, err := json.MarshalIndent(singleTheme, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal theme: %w", err)
	}

	return p.writeThemeFiles(singleJSON)
}

// loadTemplate loads and parses the Zed theme template.
func (p *Plugin) loadTemplate(verbose bool) (*template.Template, error) {
	loader := tincttemplate.New("zed", templatesFS)
	if verbose {
		loader.WithVerbose(true, &stdLogger{})
	}

	tmplContent, fromCustom, err := loader.Load("templates/theme.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	if verbose {
		if fromCustom {
			fmt.Fprintf(os.Stderr, "   Template source: custom (%s)\n", loader.CustomPath("templates/theme.json.tmpl"))
		} else {
			fmt.Fprintf(os.Stderr, "   Template source: embedded\n")
		}
	}

	tmpl, err := template.New("theme").
		Funcs(tincttemplate.TemplateFuncs()).
		Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// executeTemplate executes the template and extracts the theme object.
func (p *Plugin) executeTemplate(tmpl *template.Template, themeData *colour.ThemeData) (map[string]any, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Parse the generated JSON
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse generated JSON: %w", err)
	}

	// Extract the theme object from the themes array
	themes, ok := result["themes"].([]any)
	if !ok || len(themes) == 0 {
		return nil, fmt.Errorf("invalid theme structure: missing themes array")
	}

	themeObj, ok := themes[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid theme object structure")
	}

	return themeObj, nil
}

// writeThemeFiles writes theme content to all appropriate output directories.
func (p *Plugin) writeThemeFiles(content []byte) (map[string][]byte, error) {
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
		if err := os.MkdirAll(outputDir, 0o750); err != nil {
			return nil, fmt.Errorf("failed to create themes directory %q: %w", outputDir, err)
		}

		// Add file with full path.
		fullPath := filepath.Join(outputDir, "tinct.json")
		files[fullPath] = content
	}

	return files, nil
}

// PreExecute checks if Zed is installed (config directory exists).
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
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
func (p *Plugin) PostExecute(_ context.Context, _ []string) error {
	return nil
}

// stdLogger implements the Logger interface for template loading.
type stdLogger struct{}

func (l *stdLogger) Printf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// GetMetadata returns plugin metadata.
//
// The Metadata block lets tinct-check-readmes diff the README frontmatter
// against runtime reality. Keep these fields in sync with what the plugin
// actually does.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "zed",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate Zed editor theme files",
		PluginProtocol:  string(tinctplugin.PluginTypeGoPlugin),
		Metadata: &tinctplugin.Metadata{
			// Zed detection is by config-directory presence (native at
			// ~/.config/zed and Flatpak at
			// ~/.var/app/dev.zed.Zed/config/zed); no binary on PATH is
			// required.
			DefaultOutputDir: "~/.config/zed/themes",
			GeneratedFiles:   []string{"tinct.json"},
			Pattern:          "single-file",
			Reload: &tinctplugin.ReloadMetadata{
				// Zed watches its themes directory and reloads the
				// active theme live when its file changes.
				Method:             "watch",
				UserActionRequired: false,
			},
		},
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
