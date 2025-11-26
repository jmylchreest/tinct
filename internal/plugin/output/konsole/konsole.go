// Package konsole provides an output plugin for Konsole terminal color themes.
package konsole

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/pkg/util/appdetect"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Konsole terminal.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Konsole output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "konsole"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Konsole terminal color theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "konsole.output-dir", "", "Output directory (default: ~/.local/share/konsole)")
}

// SetVerbose enables or disables verbose logging for the plugin.
// Implements the output.VerbosePlugin interface.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetEmbeddedFS returns the embedded template filesystem.
// Implements the output.TemplateProvider interface.
func (p *Plugin) GetEmbeddedFS() any {
	return templates
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "konsole.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.local/share/konsole)", Required: false},
	}
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
		return ".local/share/konsole"
	}
	return filepath.Join(home, ".local", "share", "konsole")
}

// Generate creates the theme file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate template metadata fields.
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = "Tinct.colorscheme"

	files := make(map[string][]byte)

	// Generate theme file.
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	// Use theme-specific filename (TinctDark or TinctLight).
	var fileName string
	if themeData.ThemeType() == colour.ThemeDark {
		fileName = "TinctDark.colorscheme"
	} else {
		fileName = "TinctLight.colorscheme"
	}

	files[fileName] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("konsole", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.colorscheme.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.colorscheme.tmpl\n")
	}

	tmpl, err := template.New("theme").Funcs(common.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute checks if konsole is available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if konsole executable exists (native, Flatpak, or AppImage).
	if !appdetect.IsPresentAny([]string{"konsole"}, nil) {
		return true, "konsole executable not found on $PATH", nil
	}

	// Check if config directory exists, create if it doesn't.
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Try to create the config directory.
		if err := os.MkdirAll(configDir, 0o755); err != nil { // #nosec G301 - Config directory needs standard permissions
			return true, fmt.Sprintf("failed to create konsole config directory: %s", configDir), nil
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Created konsole config directory: %s\n", configDir)
		}
	}

	return false, "", nil
}

// PostExecute provides usage instructions for applying the theme.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(_ context.Context, _ output.ExecutionContext, generatedFiles []string) error {
	if p.verbose && len(generatedFiles) > 0 {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Konsole color scheme generated successfully!\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   To use this theme:\n")
		fmt.Fprintf(os.Stderr, "   1. Open Konsole\n")
		fmt.Fprintf(os.Stderr, "   2. Go to Settings > Edit Current Profile > Appearance\n")
		fmt.Fprintf(os.Stderr, "   3. Select 'Tinct Dark' or 'Tinct Light' from the color scheme list\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Note: You may need to restart Konsole to see the new themes.\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	return nil
}
