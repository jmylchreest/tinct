// Package hyprland provides an output plugin for Hyprland window manager colour themes.
package hyprland

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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

//go:embed *.tmpl templates
var templates embed.FS

// versionRegex matches the Hyprland version from "hyprland --version" or "hyprctl version" output.
// Example: "Hyprland 0.53.0 built from branch...".
var versionRegex = regexp.MustCompile(`Hyprland\s+(\d+\.\d+(?:\.\d+)?)`)

// Cached version to avoid repeated exec calls.
var (
	cachedVersion     string
	cachedVersionOnce sync.Once
)

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Hyprland.
type Plugin struct {
	outputDir    string
	generateStub bool
	stubPath     string
	reloadConfig bool
	verbose      bool
}

// New creates a new Hyprland output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir:    "",
		generateStub: true,
		stubPath:     "",
		verbose:      false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "hyprland"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Hyprland compositor theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "hyprland.output-dir", "", "Output directory (default: ~/.config/hypr/themes)")
	cmd.Flags().BoolVar(&p.generateStub, "hyprland.generate-stub", true, "Generate example config stub")
	cmd.Flags().StringVar(&p.stubPath, "hyprland.stub-path", "", "Custom path for stub file")
	cmd.Flags().BoolVar(&p.reloadConfig, "hyprland.reload", false, "Reload hyprland config after generation (runs hyprctl reload)")
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

// GetTargetVersion returns the installed Hyprland version.
// Implements the output.VersionedPlugin interface.
// This enables version-specific templates for Hyprland configuration changes.
func (p *Plugin) GetTargetVersion() string {
	cachedVersionOnce.Do(func() {
		cachedVersion = detectHyprlandVersion(p.verbose)
	})
	return cachedVersion
}

// detectHyprlandVersion runs "hyprctl version" or "hyprland --version" to get the version.
func detectHyprlandVersion(_ bool) string {
	// Try hyprctl version first (works when Hyprland is running)
	cmdOutput, err := exec.Command("hyprctl", "version").Output()
	if err == nil {
		if v := parseHyprlandVersion(string(cmdOutput)); v != "" {
			return v
		}
	}

	// Fall back to hyprland --version
	cmdOutput, err = exec.Command("hyprland", "--version").Output()
	if err == nil {
		if v := parseHyprlandVersion(string(cmdOutput)); v != "" {
			return v
		}
	}

	return ""
}

// parseHyprlandVersion extracts the version number from Hyprland version output.
func parseHyprlandVersion(versionOutput string) string {
	matches := versionRegex.FindStringSubmatch(versionOutput)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "hyprland.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/hypr/themes)", Required: false},
		{Name: "hyprland.generate-stub", Type: "bool", Default: "true", Description: "Generate example config stub", Required: false},
		{Name: "hyprland.stub-path", Type: "string", Default: "", Description: "Custom path for stub file", Required: false},
		{Name: "hyprland.reload", Type: "bool", Default: "false", Description: "Reload hyprland config after generation", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	// Nothing to validate - all fields have defaults.
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("hyprland", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "hypr", "themes"))
}

// Hooks declares hyprland's pre/post-execute behaviour. The Hyprland
// config dir must exist (signal that Hyprland is installed); hyprctl is
// optional. When --hyprland.reload is set, runs hyprctl reload via the
// shared exec verb runner.
func (p *Plugin) Hooks() hooks.Spec {
	spec := hooks.Spec{
		RequiredDirs:     []string{"~/.config/hypr"},
		OptionalBinaries: []string{"hyprctl"},
		AutoCreateDir:    true,
	}
	if p.reloadConfig {
		spec.Reload = &hooks.ReloadSpec{
			Verb: hooks.VerbExec,
			Args: []string{"hyprctl", "reload"},
		}
	}
	return spec
}

// Generate creates the theme file and optional stub configuration.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	// Populate output directory and color file name in theme data for templates.
	themeData.OutputDir = p.DefaultOutputDir()
	themeData.ColorFileName = "tinct-colours.conf"

	files := make(map[string][]byte)

	// Generate main theme file.
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}

	themeFile := "tinct-colours.conf"
	files[themeFile] = themeContent

	// Generate stub if requested.
	if p.generateStub {
		stubContent, err := p.generateStubConfig(themeData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate stub: %w", err)
		}

		stubFile := "tinct.conf"
		if p.stubPath != "" {
			stubFile = filepath.Base(p.stubPath)
		}
		files[stubFile] = stubContent
	}

	return files, nil
}

// generateTheme creates the main theme configuration file with colour variables.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override and versioned template support.
	loader := tmplloader.New("hyprland", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	// Enable versioned template selection based on installed Hyprland version.
	targetVersion := p.GetTargetVersion()
	if targetVersion != "" {
		loader.WithTargetVersion(targetVersion)
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Detected Hyprland version: %s\n", targetVersion)
		}
	} else if p.verbose {
		fmt.Fprintf(os.Stderr, "   Could not detect Hyprland version, using default templates\n")
	}

	tmplContent, fromCustom, err := loader.Load("tinct-colours.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-colours.conf.tmpl\n")
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

// generateStubConfig creates an example configuration file showing how to use the theme.
func (p *Plugin) generateStubConfig(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override and versioned template support.
	loader := tmplloader.New("hyprland", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	// Enable versioned template selection based on installed Hyprland version.
	targetVersion := p.GetTargetVersion()
	if targetVersion != "" {
		loader.WithTargetVersion(targetVersion)
	}

	tmplContent, fromCustom, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read example template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.conf.tmpl\n")
	}

	tmpl, err := template.New("example").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse example template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute example template: %w", err)
	}

	return buf.Bytes(), nil
}
