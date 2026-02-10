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

// Plugin implements the tinct OutputPlugin interface for OpenCode themes.
type Plugin struct {
	version   string
	commit    string
	date      string
	outputDir string
}

// detectOpenCodePaths finds all installed OpenCode theme directories.
func detectOpenCodePaths() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	var paths []string

	// Check for XDG_CONFIG_HOME first.
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		opencodePath := filepath.Join(xdgConfig, "opencode")
		if _, err := os.Stat(opencodePath); err == nil {
			paths = append(paths, filepath.Join(opencodePath, "themes"))
		}
	}

	// Check for standard OpenCode config directory.
	standardPath := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(standardPath); err == nil {
		themesPath := filepath.Join(standardPath, "themes")
		// Avoid adding duplicate if XDG_CONFIG_HOME resolved to the same path.
		if len(paths) == 0 || paths[0] != themesPath {
			paths = append(paths, themesPath)
		}
	}

	return paths, nil
}

// Generate creates an OpenCode theme JSON file from the palette data.
func (p *Plugin) Generate(ctx context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
	// Load and parse template.
	tmpl, err := p.loadTemplate(palette.Verbose)
	if err != nil {
		return nil, err
	}

	// Generate theme data for the primary palette.
	themeData := convertPaletteDataToThemeData(palette)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	content := buf.Bytes()

	// If we have dual-theme support (alternate light/dark), generate a second
	// theme file so OpenCode can auto-detect based on terminal background.
	if palette.AlternateTheme != nil {
		if palette.Verbose {
			fmt.Fprintf(os.Stderr, "   Generating dual-theme (light + dark)\n")
		}

		alternatePaletteData := tinctplugin.PaletteData{
			Colours:    palette.AlternateTheme.Colours,
			AllColours: palette.AlternateTheme.AllColours,
			ThemeType:  palette.AlternateTheme.ThemeType,
			Verbose:    palette.Verbose,
		}
		altThemeData := convertPaletteDataToThemeData(alternatePaletteData)

		var altBuf bytes.Buffer
		if err := tmpl.Execute(&altBuf, altThemeData); err != nil {
			return nil, fmt.Errorf("failed to generate alternate theme: %w", err)
		}

		// Write both theme files.
		files, err := p.writeThemeFiles(content)
		if err != nil {
			return nil, err
		}

		// Determine alternate filename based on theme type.
		altSuffix := "light"
		if palette.AlternateTheme.ThemeType == "dark" {
			altSuffix = "dark"
		}

		altFiles, err := p.writeThemeFilesWithName(altBuf.Bytes(), fmt.Sprintf("tinct-%s.json", altSuffix))
		if err != nil {
			return nil, err
		}

		for k, v := range altFiles {
			files[k] = v
		}

		return files, nil
	}

	return p.writeThemeFiles(content)
}

// loadTemplate loads and parses the OpenCode theme template.
func (p *Plugin) loadTemplate(verbose bool) (*template.Template, error) {
	loader := tincttemplate.New("opencode", templatesFS)
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

// writeThemeFiles writes theme content to all appropriate output directories.
func (p *Plugin) writeThemeFiles(content []byte) (map[string][]byte, error) {
	return p.writeThemeFilesWithName(content, "tinct.json")
}

// writeThemeFilesWithName writes theme content with a specific filename.
func (p *Plugin) writeThemeFilesWithName(content []byte, filename string) (map[string][]byte, error) {
	// Determine output directories.
	var outputDirs []string
	if p.outputDir != "" {
		// User specified a custom output directory - use only that.
		outputDirs = []string{p.outputDir}
	} else {
		// No custom directory specified - detect OpenCode installation paths.
		detectedPaths, err := detectOpenCodePaths()
		if err != nil {
			return nil, err
		}

		if len(detectedPaths) == 0 {
			// Fallback to default if no installations detected.
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			outputDirs = []string{filepath.Join(home, ".config", "opencode", "themes")}
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
		fullPath := filepath.Join(outputDir, filename)
		files[fullPath] = content
	}

	return files, nil
}

// PreExecute checks if OpenCode config directory exists.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Detect OpenCode installations.
	detectedPaths, err := detectOpenCodePaths()
	if err != nil {
		return false, "", err
	}

	// If no installations found, skip this plugin.
	if len(detectedPaths) == 0 {
		return true, "OpenCode not installed (config directory does not exist)", nil
	}

	return false, "", nil
}

// PostExecute runs after theme files are written.
// OpenCode does not currently support hot-reloading themes, so this is a no-op.
func (p *Plugin) PostExecute(ctx context.Context, files []string) error {
	return nil
}

// stdLogger implements the Logger interface for template loading.
type stdLogger struct{}

func (l *stdLogger) Printf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// GetMetadata returns plugin metadata.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "opencode",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate OpenCode theme files",
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
			Description: "Output directory for OpenCode theme files (default: auto-detects ~/.config/opencode/themes)",
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
