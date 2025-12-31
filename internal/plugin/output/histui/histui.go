// Package histui provides an output plugin for histui notification daemon CSS themes.
package histui

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for histui.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new histui output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "histui"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "histui notification daemon CSS colours"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "histui.output-dir", "", "Output directory (default: ~/.config/histui/themes)")
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
		{Name: "histui.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/histui/themes)", Required: false},
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
		return ".config/histui/themes"
	}
	return filepath.Join(home, ".config", "histui", "themes")
}

// Generate creates the theme file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate tinct-colors.css file.
	colorsContent, err := p.generateColors(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate colors: %w", err)
	}

	files["tinct-colors.css"] = colorsContent

	return files, nil
}

// generateColors creates the CSS colours file.
func (p *Plugin) generateColors(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("histui", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct-colors.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read colors template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colors.css.tmpl\n")
	}

	tmpl, err := template.New("colors").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse colors template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute colors template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute checks if histui is available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Get the histui config base directory (~/.config/histui)
	home, err := os.UserHomeDir()
	if err != nil {
		return true, "cannot determine home directory", nil
	}
	histuiConfigDir := filepath.Join(home, ".config", "histui")

	// Check if histuid executable exists on PATH.
	_, lookErr := exec.LookPath("histuid")
	if lookErr != nil {
		// histuid not found - check if config directory exists anyway
		// (user may have histui installed but not in PATH, or want to pre-generate themes)
		if _, statErr := os.Stat(histuiConfigDir); os.IsNotExist(statErr) {
			return true, "histuid executable not found on $PATH and ~/.config/histui does not exist", nil
		}
		// Config directory exists, continue with generation
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Note: histuid not found on $PATH but ~/.config/histui exists - continuing\n")
		}
	}

	// Check if themes directory exists (create it if not).
	themesDir := p.DefaultOutputDir()
	if _, statErr := os.Stat(themesDir); os.IsNotExist(statErr) {
		// #nosec G301 -- Theme directory needs to be readable by histui daemon
		if mkErr := os.MkdirAll(themesDir, 0o750); mkErr != nil {
			return true, fmt.Sprintf("histui themes directory does not exist and cannot be created: %s", themesDir), nil
		}
	}

	return false, "", nil
}

// PostExecute touches the theme file to trigger histui hot-reload.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(_ context.Context, _ output.ExecutionContext, writtenFiles []string) error {
	// histui watches theme files via inotify - imports are inlined at load time,
	// so changing tinct-colors.css alone won't trigger a reload.
	// Users should configure their tinct theme.css to be watched, or restart histuid.

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   tinct-colors.css written to histui themes directory\n")
		fmt.Fprintf(os.Stderr, "   Touch your theme.css or restart histuid to apply changes\n")
	}

	return nil
}
