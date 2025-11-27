// Package qt6 provides an output plugin for Qt6 application theming via qt6ct.
package qt6

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
	kdedbus "github.com/jmylchreest/tinct/internal/plugin/output/shared/dbus_kde"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/util/appdetect"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for Qt6 themes via qt6ct.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Qt6 plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "qt6"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Qt6 application theme (via qt6ct configuration tool)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers command-line flags for this plugin.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "qt6.output-dir", "",
		"Output directory for Qt6 color scheme files (default: ~/.config/qt6ct/colors)")
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
			Name:        "qt6.output-dir",
			Type:        "string",
			Default:     "~/.config/qt6ct/colors",
			Description: "Output directory for Qt6 color scheme files",
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
		return ".config/qt6ct/colors"
	}
	return filepath.Join(home, ".config", "qt6ct", "colors")
}

// PreExecute checks if the qt6ct config directory exists.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return true, "cannot determine home directory", nil
	}

	qt6ctConfigDir := filepath.Join(home, ".config", "qt6ct")

	// Check if qt6ct config directory exists
	if !appdetect.IsPresentAll(nil, []string{qt6ctConfigDir}) {
		return true, fmt.Sprintf("qt6ct config directory does not exist (%s). Install qt6ct first:\n"+
			"  Arch/CachyOS: sudo pacman -S qt6ct\n"+
			"  Then set: export QT_QPA_PLATFORMTHEME=qt6ct", qt6ctConfigDir), nil
	}

	return false, "", nil
}

// Generate creates the Qt6 color scheme file.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate the Qt6 color scheme file
	confContent, err := p.generateColorScheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Qt6 color scheme: %w", err)
	}
	files["tinct.conf"] = confContent

	return files, nil
}

// generateColorScheme generates the Qt6 color scheme content.
func (p *Plugin) generateColorScheme(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("qt6", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	tmplContent, _, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tmpl, err := template.New("qt6").
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

// PostExecute provides instructions for configuring Qt6 and attempts live reload via D-Bus.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, _ []string) error {
	if execCtx.DryRun {
		return nil
	}

	// Try to trigger live Qt theme reload via KDE D-Bus (if on KDE Plasma)
	reloaded, err := kdedbus.ReloadTheme(ctx)
	if err != nil && p.verbose {
		fmt.Fprintf(os.Stderr, "   Note: D-Bus theme reload failed: %v\n", err)
	} else if reloaded && p.verbose {
		fmt.Fprintf(os.Stderr, "   Qt6 theme reload triggered via D-Bus (KDE Plasma running apps should update)\n")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	qt6ctConfigPath := filepath.Join(home, ".config", "qt6ct", "qt6ct.conf")

	// Check if qt6ct.conf exists and uses the tinct color scheme
	content, err := os.ReadFile(qt6ctConfigPath)
	alreadyConfigured := false
	if err == nil {
		// Check if custom palette is set and points to tinct
		alreadyConfigured = bytes.Contains(content, []byte("color_scheme_path=")) &&
			bytes.Contains(content, []byte("tinct.conf"))
	}

	if !alreadyConfigured {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "  Qt6 Theme Configuration Required\n")
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "To apply the tinct color scheme to Qt6 applications:\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "1. Ensure QT_QPA_PLATFORMTHEME is set:\n")
		fmt.Fprintf(os.Stderr, "   export QT_QPA_PLATFORMTHEME=qt6ct\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "2. Run qt6ct and select the 'tinct' color scheme:\n")
		fmt.Fprintf(os.Stderr, "   qt6ct\n")
		fmt.Fprintf(os.Stderr, "   → Appearance → Color scheme → Custom → tinct\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Changes will apply to Qt6 applications after restart.\n")
		fmt.Fprintf(os.Stderr, "\n")
	} else if p.verbose {
		fmt.Fprintf(os.Stderr, "   Qt6 color scheme: OK (tinct.conf configured)\n")
	}

	return nil
}
