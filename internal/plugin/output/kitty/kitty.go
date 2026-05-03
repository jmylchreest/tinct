// Package kitty provides an output plugin for Kitty terminal colour themes.
package kitty

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
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Kitty terminal.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Kitty output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "kitty"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Kitty terminal colour theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "kitty.output-dir", "", "Output directory (default: ~/.config/kitty/themes)")
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
		{Name: "kitty.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/kitty/themes)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
// Kitty honours $XDG_CONFIG_HOME on macOS too, so XDGConfigDir works on
// every platform kitty supports.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}
	return filepath.Join(paths.XDGConfigDir(), "kitty", "themes")
}

// Hooks declares kitty's pre/post-execute behaviour. The conflict
// detection in PostExecute (custom message inspecting kitty.conf
// content) stays imperative because it's not expressible as a spec
// field — the spec covers the routine binary check, dir auto-create,
// and SIGUSR1 reload broadcast.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		RequiredBinaries: []string{"kitty"},
		OptionalBinaries: []string{"kitten"},
		AutoCreateDir:    true,
		Reload: &hooks.ReloadSpec{
			Verb: hooks.VerbSignal,
			Args: []string{"kitty", "SIGUSR1"},
		},
	}
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

	files["tinct.conf"] = themeContent

	return files, nil
}

// generateTheme creates the theme configuration file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("kitty", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.conf.tmpl\n")
	}

	tmpl, err := template.New("theme").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse theme template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute theme template: %w", err)
	}

	return buf.Bytes(), nil
}

// PostExecute keeps the conflict-detection warning. The reload itself is
// handled declaratively by the Hooks() spec (signal broadcast).
func (p *Plugin) PostExecute(_ context.Context, _ output.ExecutionContext, _ []string) error {
	p.checkForConflictingTheme()
	return nil
}

// checkForConflictingTheme checks if there's a current-theme.conf that might conflict.
func (p *Plugin) checkForConflictingTheme() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	kittyConf := filepath.Join(home, ".config", "kitty", "kitty.conf")
	content, err := os.ReadFile(kittyConf) // #nosec G304 -- Standard config path
	if err != nil {
		return // File doesn't exist or can't be read
	}

	confStr := string(content)
	hasTinctInclude := false
	hasCurrentTheme := false
	tinctAfterCurrent := false

	lines := strings.Split(confStr, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "include") && strings.Contains(trimmed, "tinct.conf") {
			hasTinctInclude = true
			// Check if current-theme.conf appears before this line
			for j := range i {
				if strings.Contains(lines[j], "current-theme.conf") {
					tinctAfterCurrent = true
					break
				}
			}
		}
		if strings.Contains(trimmed, "current-theme.conf") {
			hasCurrentTheme = true
		}
	}

	if hasTinctInclude && hasCurrentTheme && !tinctAfterCurrent {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: Conflicting theme configuration detected!\n")
		fmt.Fprintf(os.Stderr, "   Your kitty.conf includes both 'current-theme.conf' and 'themes/tinct.conf'\n")
		fmt.Fprintf(os.Stderr, "   The 'current-theme.conf' (from 'kitten themes') appears AFTER tinct.conf\n")
		fmt.Fprintf(os.Stderr, "   and will override tinct's colors.\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   Recommended fix:\n")
		fmt.Fprintf(os.Stderr, "   1. Remove the kitten themes block from your kitty.conf:\n")
		fmt.Fprintf(os.Stderr, "      # BEGIN_KITTY_THEME\n")
		fmt.Fprintf(os.Stderr, "      ...\n")
		fmt.Fprintf(os.Stderr, "      # END_KITTY_THEME\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "   OR move 'include themes/tinct.conf' to appear AFTER 'include current-theme.conf'\n")
		fmt.Fprintf(os.Stderr, "\n")
	}
}
