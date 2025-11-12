// Package dunst provides an output plugin for Dunst notification daemon colour themes.
package dunst

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
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Dunst.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Dunst output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "dunst"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Dunst notification daemon theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "dunst.output-dir", "", "Output directory (default: ~/.config/dunst/dunstrc.d)")
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
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		{Name: "dunst.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/dunst/dunstrc.d)", Required: false},
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
		return ".config/dunst/dunstrc.d"
	}
	return filepath.Join(home, ".config", "dunst", "dunstrc.d")
}

// Generate creates the theme file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate theme file.
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	files["60-tinct.conf"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("dunst", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.dunstrc.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.dunstrc.tmpl\n")
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

// PreExecute checks if dunst and dunstctl are available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if dunst executable exists on PATH.
	_, err = exec.LookPath("dunst")
	if err != nil {
		return true, "dunst executable not found on $PATH", nil
	}

	// Check if dunstctl executable exists on PATH (needed for reload).
	_, err = exec.LookPath("dunstctl")
	if err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: dunstctl not found - config reload will not be available\n")
		}
	}

	// Check if config directory exists (create it if not).
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// For dunst, we can create the directory since it's straightforward.
		if err := os.MkdirAll(configDir, 0o755); err != nil { // #nosec G301 - Config directory needs standard permissions
			return true, fmt.Sprintf("dunst config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}

// PostExecute reloads dunst configuration after theme generation.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, _ output.ExecutionContext, _ []string) error {
	// Reload dunst configuration using dunstctl.
	cmd := exec.CommandContext(ctx, "dunstctl", "reload")
	if err := cmd.Run(); err != nil {
		// If dunstctl reload fails, inform the user but don't treat it as an error.
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Note: Could not reload dunst config automatically\n")
			fmt.Fprintf(os.Stderr, "   Please restart dunst manually to apply changes\n")
		}
		return nil
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Dunst configuration reloaded\n")
	}

	return nil
}
