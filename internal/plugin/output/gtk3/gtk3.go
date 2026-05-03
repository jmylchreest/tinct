// Package gtk3 provides an output plugin for GTK3 application theming.
package gtk3

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
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/pluginconfig"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for GTK3 themes.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new GTK3 plugin instance.
func New() *Plugin {
	return &Plugin{}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "gtk3"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "GTK3 application theme (legacy GTK3 apps like GIMP)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers command-line flags for this plugin.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "gtk3.output-dir", "",
		"Output directory for GTK3 theme files (default: ~/.config/gtk-3.0)")
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
			Name:        "gtk3.output-dir",
			Type:        "string",
			Default:     "~/.config/gtk-3.0",
			Description: "Output directory for GTK3 theme files",
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
	return pluginconfig.Resolve("gtk3", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "gtk-3.0"))
}

// PreExecute checks if the GTK3 config directory exists.
// Hooks declares gtk3's pre-execute behaviour. The config dir must
// exist (signal that GTK3 is installed). PostExecute stays imperative
// because it does content-aware inspection of gtk.css and settings.ini
// to emit different install messages depending on what's already
// configured — that's not expressible as a static template.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredDirs: []string{p.DefaultOutputDir()},
	}
}

// Generate creates the GTK3 theme files.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate the GTK3 CSS file
	cssContent, err := p.generateCSS(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GTK3 CSS: %w", err)
	}
	files["tinct-gtk3.css"] = cssContent

	return files, nil
}

// generateCSS generates the GTK3 CSS content.
func (p *Plugin) generateCSS(themeData *colour.ThemeData) ([]byte, error) {
	loader := tmplloader.New("gtk3", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	tmplContent, _, err := loader.Load("tinct-gtk3.css.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	tmpl, err := template.New("gtk3").
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

// PostExecute provides instructions for configuring GTK3 and attempts live reload via D-Bus.
func (p *Plugin) PostExecute(_ context.Context, execCtx output.ExecutionContext, _ []string) error { //nolint:gocyclo
	if execCtx.DryRun {
		return nil
	}

	configDir := p.DefaultOutputDir()
	gtkCSSPath := filepath.Join(configDir, "gtk.css")
	settingsIniPath := filepath.Join(configDir, "settings.ini")

	// Check if gtk.css has our import
	content, err := os.ReadFile(gtkCSSPath) // #nosec G304 -- Config path from DefaultOutputDir
	alreadyConfigured := false
	if err == nil {
		alreadyConfigured = strings.Contains(string(content), `@import url("tinct-gtk3.css")`)
	}

	// Check if settings.ini uses adw-gtk3 theme
	settingsContent, err := os.ReadFile(settingsIniPath) // #nosec G304 -- Config path from XDG config dir
	usesAdwGtk3 := false
	currentTheme := ""
	if err == nil {
		lines := strings.SplitSeq(string(settingsContent), "\n")
		for line := range lines {
			line = strings.TrimSpace(line)
			if after, ok := strings.CutPrefix(line, "gtk-theme-name="); ok {
				currentTheme = after
				if strings.HasPrefix(currentTheme, "adw-gtk3") {
					usesAdwGtk3 = true
				}
				break
			}
		}
	}

	// Show warnings/instructions if needed
	needsConfig := !alreadyConfigured || !usesAdwGtk3

	if needsConfig {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "  GTK3 Theme Configuration Required\n")
		fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Fprintf(os.Stderr, "\n")

		if !alreadyConfigured {
			fmt.Fprintf(os.Stderr, "Add this line to %s:\n", gtkCSSPath)
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "    @import url(\"tinct-gtk3.css\");\n")
			fmt.Fprintf(os.Stderr, "\n")
		}

		if !usesAdwGtk3 {
			if currentTheme != "" && p.verbose {
				fmt.Fprintf(os.Stderr, "   Current GTK3 theme: %s\n", currentTheme)
			}
			fmt.Fprintf(os.Stderr, "For tinct colors to apply, install and use the adw-gtk-theme:\n")
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "    # Arch/CachyOS:\n")
			fmt.Fprintf(os.Stderr, "    sudo pacman -S adw-gtk-theme\n")
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "    # Other distros may use: adw-gtk3\n")
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "Then set in %s:\n", settingsIniPath)
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "    gtk-theme-name=adw-gtk3-dark\n")
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "The adw-gtk-theme is required to respect tinct color definitions.\n")
		} else if p.verbose {
			fmt.Fprintf(os.Stderr, "   GTK3 theme correctly set to: %s\n", currentTheme)
		}

		if needsConfig {
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Fprintf(os.Stderr, "Changes will apply to newly opened GTK3 applications.\n")
			fmt.Fprintf(os.Stderr, "\n")
		}
	} else if p.verbose {
		fmt.Fprintf(os.Stderr, "   GTK3 theme configuration: OK\n")
	}

	return nil
}
