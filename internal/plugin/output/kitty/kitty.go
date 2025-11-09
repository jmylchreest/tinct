// Package kitty provides an output plugin for Kitty terminal colour themes.
package kitty

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
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
	return "Generate Kitty terminal colour theme configuration"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
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
		return ".config/kitty/themes"
	}
	return filepath.Join(home, ".config", "kitty", "themes")
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
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.conf.tmpl\n")
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

// PreExecute checks if kitty and kitten are available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if kitty executable exists on PATH.
	_, err = exec.LookPath("kitty")
	if err != nil {
		return true, "kitty executable not found on $PATH", nil
	}

	// Check if kitten executable exists on PATH (needed for reload).
	_, err = exec.LookPath("kitten")
	if err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: kitten not found - config reload will not be available\n")
		}
	}

	// Check if themes directory exists, create if it doesn't.
	themesDir := p.DefaultOutputDir()
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		// Try to create the themes directory.
		if err := os.MkdirAll(themesDir, 0o755); err != nil { // #nosec G301 - Config directory needs standard permissions
			return true, fmt.Sprintf("failed to create kitty themes directory: %s", themesDir), nil
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Created kitty themes directory: %s\n", themesDir)
		}
	}

	return false, "", nil
}

// PostExecute applies the theme to all running kitty instances.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, _ output.ExecutionContext, _ []string) error {
	// Check for conflicting current-theme.conf that might override tinct theme.
	p.checkForConflictingTheme()

	// Reload all kitty instances by sending SIGUSR1 signal.
	// This works on Unix-like systems (Linux, macOS, BSD).
	if err := p.reloadAllKittyInstances(); err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Note: Could not reload kitty instances: %v\n", err)
		}
		return nil
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Kitty config reloaded in all instances\n")
	}

	return nil
}

// checkForConflictingTheme checks if there's a current-theme.conf that might conflict.
func (p *Plugin) checkForConflictingTheme() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	kittyConf := filepath.Join(home, ".config", "kitty", "kitty.conf")
	content, err := os.ReadFile(kittyConf)
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
