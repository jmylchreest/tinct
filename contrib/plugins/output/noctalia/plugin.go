package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/jmylchreest/tinct/pkg/colour"
	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
	tincttemplate "github.com/jmylchreest/tinct/pkg/template"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// paletteName is the Noctalia custom-palette name; the file is written as
// <paletteName>.json and selected in config via `custom_palette = "tinct"`.
const paletteName = "tinct"

// Plugin implements the tinct OutputPlugin interface for the Noctalia shell.
//
// Noctalia (v5) ships its own Material You generator, but reads a fixed
// palette file when `[theme] source = "custom"` is set. This plugin writes
// that palette file so Noctalia's colours match the rest of a tinct-themed
// desktop. The schema is a JSON object with `dark` and `light` variants,
// each carrying 16 Material-3 roles plus a terminal colour block.
type Plugin struct {
	version   string
	commit    string
	date      string
	outputDir string
}

// noctaliaConfigDir returns ~/.config/noctalia (honouring NOCTALIA_CONFIG_HOME
// then XDG_CONFIG_HOME), and whether the directory currently exists.
func noctaliaConfigDir() (dir string, exists bool) {
	if c := os.Getenv("NOCTALIA_CONFIG_HOME"); c != "" {
		dir = c
	} else if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		dir = filepath.Join(x, "noctalia")
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		dir = filepath.Join(home, ".config", "noctalia")
	}
	_, err := os.Stat(dir)
	return dir, err == nil
}

// resolveOutputDir returns the palettes directory to write into.
func (p *Plugin) resolveOutputDir() (string, error) {
	if p.outputDir != "" {
		return p.outputDir, nil
	}
	dir, _ := noctaliaConfigDir()
	if dir == "" {
		return "", fmt.Errorf("failed to resolve noctalia config directory")
	}
	return filepath.Join(dir, "palettes"), nil
}

// Generate writes the Noctalia custom palette file from the theme data.
func (p *Plugin) Generate(_ context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
	tmpl, err := p.loadTemplate(palette.Verbose)
	if err != nil {
		return nil, err
	}

	// Render the primary variant.
	primaryObj, err := renderVariant(tmpl, convertPaletteDataToThemeData(palette))
	if err != nil {
		return nil, fmt.Errorf("failed to render primary variant: %w", err)
	}

	// Assign variants to Noctalia's fixed dark/light keys regardless of
	// which one is primary.
	var darkObj, lightObj map[string]any
	if palette.ThemeType == "light" {
		lightObj = primaryObj
	} else {
		darkObj = primaryObj
	}

	if palette.AlternateTheme != nil {
		if palette.Verbose {
			fmt.Fprintf(os.Stderr, "   Generating dual-theme (dark + light)\n")
		}
		altPalette := tinctplugin.PaletteData{
			Colours:    palette.AlternateTheme.Colours,
			AllColours: palette.AlternateTheme.AllColours,
			ThemeType:  palette.AlternateTheme.ThemeType,
			Verbose:    palette.Verbose,
		}
		altObj, err := renderVariant(tmpl, convertPaletteDataToThemeData(altPalette))
		if err != nil {
			return nil, fmt.Errorf("failed to render alternate variant: %w", err)
		}
		if palette.AlternateTheme.ThemeType == "light" {
			lightObj = altObj
		} else {
			darkObj = altObj
		}
	}

	// Noctalia requires `dark`; `light` is optional and falls back to dark.
	// If only a light variant exists, use it for dark too so the file is valid.
	if darkObj == nil {
		darkObj = lightObj
	}

	out := map[string]any{"dark": darkObj}
	if lightObj != nil {
		out["light"] = lightObj
	}

	content, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal palette: %w", err)
	}
	content = append(content, '\n')

	outputDir, err := p.resolveOutputDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create palettes directory %q: %w", outputDir, err)
	}

	files := map[string][]byte{
		filepath.Join(outputDir, paletteName+".json"): content,
	}
	return files, nil
}

// loadTemplate loads and parses the palette variant template.
func (p *Plugin) loadTemplate(verbose bool) (*template.Template, error) {
	loader := tincttemplate.New("noctalia", templatesFS)
	if verbose {
		loader.WithVerbose(true, &stdLogger{})
	}

	tmplContent, fromCustom, err := loader.Load("templates/palette.json.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}
	if verbose {
		if fromCustom {
			fmt.Fprintf(os.Stderr, "   Template source: custom (%s)\n", loader.CustomPath("templates/palette.json.tmpl"))
		} else {
			fmt.Fprintf(os.Stderr, "   Template source: embedded\n")
		}
	}

	tmpl, err := template.New("palette").Funcs(tincttemplate.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return tmpl, nil
}

// renderVariant executes the template for one theme variant and returns the
// parsed JSON object (the 16 Material-3 roles + terminal block).
func renderVariant(tmpl *template.Template, themeData *colour.ThemeData) (map[string]any, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(buf.Bytes(), &obj); err != nil {
		return nil, fmt.Errorf("template produced invalid JSON: %w", err)
	}
	return obj, nil
}

// PreExecute skips the plugin when Noctalia is not installed (no config dir).
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	if p.outputDir != "" {
		return false, "", nil
	}
	if _, exists := noctaliaConfigDir(); !exists {
		return true, "Noctalia not installed (config directory does not exist)", nil
	}
	return false, "", nil
}

// PostExecute is a no-op: Noctalia v5 watches its config/palette files and
// hot-reloads automatically when the palette changes.
func (p *Plugin) PostExecute(_ context.Context, _ []string) error {
	return nil
}

// Templates exposes the embedded template so `tinct plugins templates dump`
// works for this external plugin.
func (p *Plugin) Templates() map[string][]byte {
	out := map[string][]byte{}
	_ = fs.WalkDir(templatesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, readErr := templatesFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		out[path] = b
		return nil
	})
	return out
}

// GetMetadata returns plugin metadata.
func (p *Plugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "noctalia",
		Type:            "output",
		Version:         p.version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate a custom colour palette for the Noctalia shell",
		PluginProtocol:  string(tinctplugin.PluginTypeGoPlugin),
		Metadata: &tinctplugin.Metadata{
			// Noctalia detection is by config-directory presence
			// (~/.config/noctalia); no binary on PATH is required.
			DefaultOutputDir: "~/.config/noctalia/palettes",
			GeneratedFiles:   []string{"tinct.json"},
			Pattern:          "single-file",
			Reload: &tinctplugin.ReloadMetadata{
				// Noctalia v5 watches its config/palette files and reloads
				// the active palette live when the file changes.
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
			Default:     "~/.config/noctalia/palettes",
			Description: "Output directory for the Noctalia palette file",
			Required:    false,
		},
	}
}

// stdLogger implements the Logger interface for template loading.
type stdLogger struct{}

func (l *stdLogger) Printf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// convertPaletteDataToThemeData converts the protocol PaletteData to colour.ThemeData.
func convertPaletteDataToThemeData(palette tinctplugin.PaletteData) *colour.ThemeData {
	colors := make(map[colour.Role]colour.ColorValue)
	indexed := make([]colour.ColorValue, 0, len(palette.AllColours))

	for roleName, cc := range palette.Colours {
		role := colour.Role(roleName)
		rgba := colour.RGBA{R: cc.RGB.R, G: cc.RGB.G, B: cc.RGB.B, A: 255}
		colors[role] = colour.NewColorValue(rgba, role, cc.Index)
	}
	for _, cc := range palette.AllColours {
		role := colour.Role(cc.Role)
		rgba := colour.RGBA{R: cc.RGB.R, G: cc.RGB.G, B: cc.RGB.B, A: 255}
		indexed = append(indexed, colour.NewColorValue(rgba, role, cc.Index))
	}

	themeType := colour.ThemeLight
	if palette.ThemeType == "dark" {
		themeType = colour.ThemeDark
	}

	helper := colour.NewPaletteHelper(themeType, colors, indexed)
	return &colour.ThemeData{
		PaletteHelper: helper,
		WallpaperPath: "",
		ThemeName:     "Tinct",
	}
}
