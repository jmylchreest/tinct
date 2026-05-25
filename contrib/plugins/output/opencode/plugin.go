package main

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"text/template"

	"github.com/jmylchreest/tinct/pkg/colour"
	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
	tincttemplate "github.com/jmylchreest/tinct/pkg/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

const (
	themeTypeDark  = "dark"
	themeTypeLight = "light"
)

// Plugin implements the tinct OutputPlugin interface for OpenCode themes.
type Plugin struct {
	version   string
	commit    string
	date      string
	outputDir string
}

// detectOpenCodePaths finds all installed OpenCode theme directories.
//
// Looks for both the resolved XDG config dir (paths.XDGConfigDir, which
// honours $XDG_CONFIG_HOME and falls back to platform defaults) and
// the conventional ~/.config/opencode in case the user has that
// configured but XDG_CONFIG_HOME points elsewhere. Duplicates are
// elided so we don't write the same theme twice.
func detectOpenCodePaths() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	var found []string
	add := func(dir string) {
		themesPath := filepath.Join(dir, "themes")
		if slices.Contains(found, themesPath) {
			return
		}
		if _, statErr := os.Stat(dir); statErr == nil {
			found = append(found, themesPath)
		}
	}

	add(filepath.Join(paths.XDGConfigDir(), "opencode"))
	add(filepath.Join(home, ".config", "opencode"))

	return found, nil
}

// Generate creates an OpenCode theme JSON file from the palette data.
func (p *Plugin) Generate(_ context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
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
		altSuffix := themeTypeLight
		if palette.AlternateTheme.ThemeType == themeTypeDark {
			altSuffix = themeTypeDark
		}

		altFiles, err := p.writeThemeFilesWithName(altBuf.Bytes(), fmt.Sprintf("tinct-%s.json", altSuffix))
		if err != nil {
			return nil, err
		}

		maps.Copy(files, altFiles)

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
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
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
// The optional Metadata block lets tinct-check-readmes diff the README
// frontmatter against runtime reality. Keep these fields in sync with
// what the plugin actually does — they're the machine-checkable
// counterpart to the README's `plugin:` block.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "opencode",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate OpenCode theme files",
		PluginProtocol:  string(tinctplugin.PluginTypeGoPlugin),
		Metadata: &tinctplugin.Metadata{
			// OpenCode detection is by config-directory presence, not
			// by a binary on PATH; no required_binaries to declare.
			DefaultOutputDir: "~/.config/opencode/themes",
			GeneratedFiles:   []string{"tinct.json"},
			Pattern:          "single-file",
			Reload: &tinctplugin.ReloadMetadata{
				Method:             "none",
				UserActionRequired: true,
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
