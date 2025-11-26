// Package gtk4 provides an output plugin for GTK4 application theming.
package gtk4

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for GTK4 themes.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new GTK4 plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "gtk4"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "GTK4 application theme (traditional GTK4 apps)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers command-line flags for this plugin.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "gtk4.output-dir", "",
		"Output directory for GTK4 theme files (default: ~/.config/gtk-4.0)")
}

// SetVerbose enables verbose output.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetEmbeddedFS returns the embedded filesystem containing templates.
func (p *Plugin) GetEmbeddedFS() any {
	return templates
}

// GetFlagHelp returns help text for plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{
			Name:        "gtk4.output-dir",
			Type:        "string",
			Default:     "~/.config/gtk-4.0",
			Description: "Output directory for GTK4 theme files",
			Required:    false,
		},
	}
}

// Validate validates the plugin configuration.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/gtk-4.0"
	}
	return filepath.Join(home, ".config", "gtk-4.0")
}

// PreExecute checks if the GTK4 config directory exists.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return true, fmt.Sprintf("GTK4 config directory does not exist (%s). Install a GTK4 application first.", configDir), nil
	}

	return false, "", nil
}

// Generate creates the GTK4 theme files.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate the GTK4 CSS file
	cssContent, err := p.generateCSS(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GTK4 CSS: %w", err)
	}
	files["tinct-gtk4.css"] = cssContent

	return files, nil
}

// generateCSS generates the GTK4 CSS content.
func (p *Plugin) generateCSS(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("gtk4", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}

	tmplContent, _, err := loader.Load("tinct-gtk4.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tmpl, err := template.New("gtk4").
		Funcs(common.TemplateFuncs()).
		Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// PostExecute provides instructions for configuring GTK4.
func (p *Plugin) PostExecute(_ context.Context, execCtx output.ExecutionContext, _ []string) error {
	if execCtx.DryRun {
		return nil
	}

	gtkCSSPath := filepath.Join(p.DefaultOutputDir(), "gtk.css")

	// Check if gtk.css has our import
	content, err := os.ReadFile(gtkCSSPath)
	alreadyConfigured := false
	if err == nil {
		alreadyConfigured = strings.Contains(string(content), `@import url("tinct-gtk4.css")`)
	}

	if !alreadyConfigured {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "  GTK4 Theme Configuration Required\n")
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Add this line to %s:\n", gtkCSSPath)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "    @import url(\"tinct-gtk4.css\");\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Changes will apply to newly opened GTK4 applications.\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	return nil
}
