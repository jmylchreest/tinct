// Package awww provides an output plugin for awww (An Answer to your Wayland Wallpaper Woes) wallpaper daemon.
package awww

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
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
)

//go:embed *.tmpl
var templates embed.FS

// Default transition values for awww.
const (
	defaultTransitionType = "simple"
	defaultTransitionStep = "2"
	defaultTransitionFPS  = "30"
)

// Plugin implements the output.Plugin interface for awww.
type Plugin struct {
	outputDir      string
	verbose        bool
	transitionType string
	transitionStep string
	transitionFPS  string
	resizeMode     string
	fillColor      string
	outputs        string
}

// New creates a new awww output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir:      "",
		verbose:        false,
		transitionType: defaultTransitionType,
		transitionStep: defaultTransitionStep,
		transitionFPS:  defaultTransitionFPS,
		resizeMode:     "",
		fillColor:      "",
		outputs:        "",
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "awww"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "awww animated Wayland wallpaper daemon config"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "awww.output-dir", "", "Output directory (default: ~/.config/awww)")
	cmd.Flags().StringVar(&p.transitionType, "awww.transition-type", defaultTransitionType, "Transition type (simple, left, right, top, bottom, wipe, grow, center, outer, any, random)")
	cmd.Flags().StringVar(&p.transitionStep, "awww.transition-step", defaultTransitionStep, "Transition step (1-255, smaller = smoother)")
	cmd.Flags().StringVar(&p.transitionFPS, "awww.transition-fps", defaultTransitionFPS, "Transition FPS (1-65535)")
	cmd.Flags().StringVar(&p.resizeMode, "awww.resize", "", "Resize mode (crop, fit, no)")
	cmd.Flags().StringVar(&p.fillColor, "awww.fill-color", "", "Fill colour for fit mode (rrggbb, uses background colour if empty)")
	cmd.Flags().StringVar(&p.outputs, "awww.outputs", "", "Comma-separated list of output names to set wallpaper on (default: all)")
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
		{Name: "awww.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/awww)", Required: false},
		{Name: "awww.transition-type", Type: "string", Default: defaultTransitionType, Description: "Transition type (simple, left, right, top, bottom, wipe, grow, center, outer, any, random)", Required: false},
		{Name: "awww.transition-step", Type: "string", Default: defaultTransitionStep, Description: "Transition step (1-255, smaller = smoother)", Required: false},
		{Name: "awww.transition-fps", Type: "string", Default: defaultTransitionFPS, Description: "Transition FPS (1-65535)", Required: false},
		{Name: "awww.resize", Type: "string", Default: "", Description: "Resize mode (crop, fit, no)", Required: false},
		{Name: "awww.fill-color", Type: "string", Default: "", Description: "Fill colour for fit mode (rrggbb, uses background colour if empty)", Required: false},
		{Name: "awww.outputs", Type: "string", Default: "", Description: "Comma-separated list of output names (default: all)", Required: false},
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
		return ".config/awww"
	}
	return filepath.Join(home, ".config", "awww")
}

// Generate creates the shell script for awww wallpaper persistence.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate shell script.
	configContent, err := p.generateConfig(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate config: %w", err)
	}

	files["tinct-awww.sh"] = configContent

	return files, nil
}

// templateData extends ThemeData with awww-specific fields.
type templateData struct {
	*colour.ThemeData
	TransitionType string
	TransitionStep string
	TransitionFPS  string
	ResizeMode     string
	FillColor      string
}

// generateConfig creates the configuration file.
func (p *Plugin) generateConfig(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("awww", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	tmplContent, fromCustom, err := loader.Load("tinct-awww.sh.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read config template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-awww.sh.tmpl\n")
	}

	tmpl, err := template.New("config").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config template: %w", err)
	}

	// Determine fill colour: use configured value, or derive from background.
	fillColor := p.fillColor
	if fillColor == "" {
		// Extract background hex (without #) for awww fill colour.
		fillColor = "" // Will be handled in template.
	}

	data := &templateData{
		ThemeData:      themeData,
		TransitionType: p.transitionType,
		TransitionStep: p.transitionStep,
		TransitionFPS:  p.transitionFPS,
		ResizeMode:     p.resizeMode,
		FillColor:      fillColor,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute config template: %w", err)
	}

	return buf.Bytes(), nil
}

// Hooks declares awww's pre/post-execute behaviour. awww is optional
// (we still write the helper script without the daemon); chmod +x and
// wallpaper application use the spec's MakeExecutable and Wallpaper
// callback. The bespoke setWallpaper logic (transition opts,
// daemon-running probe, fill-color, resize-mode) is exposed via
// Wallpaper.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		OptionalBinaries:  []string{"awww"},
		AutoCreateDir:     true,
		MakeExecutable:    []string{"tinct-awww.sh"},
		SupportsWallpaper: true,
		Wallpaper:         p.setWallpaper,
	}
}

// setWallpaper applies the wallpaper using awww.
func (p *Plugin) setWallpaper(ctx context.Context, wallpaperPath string) error {
	// Make the path absolute.
	absPath, err := filepath.Abs(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Validate path to prevent command injection.
	if !utils.IsValidPath(absPath) {
		return fmt.Errorf("invalid wallpaper path: contains suspicious characters")
	}

	// Check if awww-daemon is running by attempting a query.
	checkCmd := exec.CommandContext(ctx, "awww", "query")
	if err := checkCmd.Run(); err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Skipping wallpaper application (awww-daemon not running)\n")
		}
		return nil
	}

	// Build the awww img command with transition options.
	args := []string{"img"}

	// Add output targets if specified.
	if p.outputs != "" {
		args = append(args, "-o", p.outputs)
	}

	// Add transition options.
	args = append(args,
		"--transition-type", p.transitionType,
		"--transition-step", p.transitionStep,
		"--transition-fps", p.transitionFPS,
	)

	// Add resize mode if specified.
	if p.resizeMode != "" {
		args = append(args, "--resize", p.resizeMode)
	}

	// Add fill colour if specified or derive from theme background.
	if p.fillColor != "" {
		args = append(args, "--fill-color", p.fillColor)
	}

	// Add the image path last.
	args = append(args, absPath)

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Running: awww %s\n", strings.Join(args, " "))
	}

	// #nosec G204 -- awww is a system command with validated absolute path for wallpaper.
	cmd := exec.CommandContext(ctx, "awww", args...)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set wallpaper: %w (output: %s)", err, strings.TrimSpace(string(cmdOutput)))
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Set wallpaper using awww: %s\n", absPath)
	}

	return nil
}

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}
