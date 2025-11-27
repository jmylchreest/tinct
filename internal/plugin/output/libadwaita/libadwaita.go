// Package libadwaita provides an output plugin for libadwaita/GNOME application theming.
package libadwaita

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
	gtkdbus "github.com/jmylchreest/tinct/internal/plugin/output/shared/dbus_gtk"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/util/appdetect"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for libadwaita themes.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new libadwaita plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "libadwaita"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Libadwaita application theme (modern GNOME apps)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers command-line flags for this plugin.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "libadwaita.output-dir", "",
		"Output directory for libadwaita theme files (default: ~/.config/gtk-4.0)")
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
			Name:        "libadwaita.output-dir",
			Type:        "string",
			Default:     "~/.config/gtk-4.0",
			Description: "Output directory for libadwaita theme files",
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
	home, err := os.UserHomeDir()
	if err != nil {
		return true, "cannot determine home directory", nil
	}

	gtk4ConfigDir := filepath.Join(home, ".config", "gtk-4.0")

	if !appdetect.IsPresentAll(nil, []string{gtk4ConfigDir}) {
		return true, fmt.Sprintf("GTK4 config directory does not exist (%s). Install a GTK4 application first.", gtk4ConfigDir), nil
	}

	return false, "", nil
}

// Generate creates the libadwaita theme files.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate the libadwaita CSS file
	cssContent, err := p.generateCSS(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate libadwaita CSS: %w", err)
	}
	files["tinct-libadwaita.css"] = cssContent

	return files, nil
}

// generateCSS generates the libadwaita CSS content.
func (p *Plugin) generateCSS(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("libadwaita", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	tmplContent, _, err := loader.Load("tinct-libadwaita.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tmpl, err := template.New("libadwaita").
		Funcs(utils.TemplateFuncs()).
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

// PostExecute provides instructions for configuring libadwaita and attempts live reload via D-Bus.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, _ []string) error {
	if execCtx.DryRun {
		return nil
	}

	// Try to trigger live GTK theme reload via GNOME Settings Daemon D-Bus
	reloaded, err := gtkdbus.ReloadTheme(ctx)
	if err != nil && p.verbose {
		fmt.Fprintf(os.Stderr, "   Note: D-Bus theme reload failed: %v\n", err)
	} else if reloaded && p.verbose {
		fmt.Fprintf(os.Stderr, "   Libadwaita theme reload triggered via D-Bus (running apps should update)\n")
	}

	gtkCSSPath := filepath.Join(p.DefaultOutputDir(), "gtk.css")

	// Check if gtk.css has our import
	content, err := os.ReadFile(gtkCSSPath)
	alreadyConfigured := false
	if err == nil {
		alreadyConfigured = strings.Contains(string(content), `@import url("tinct-libadwaita.css")`)
	}

	if !alreadyConfigured {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "  Libadwaita Theme Configuration Required\n")
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Add this line to %s:\n", gtkCSSPath)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "    @import url(\"tinct-libadwaita.css\");\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Changes will apply to GNOME/libadwaita applications.\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	return nil
}
