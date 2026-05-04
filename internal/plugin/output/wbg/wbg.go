// Package wbg provides an output plugin for wbg, a simple Wayland wallpaper application.
package wbg

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
	"github.com/jmylchreest/tinct/internal/pluginconfig"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
)

//go:embed *.tmpl
var templates embed.FS

// Plugin implements the output.Plugin interface for wbg.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new wbg output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "wbg"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "wbg simple Wayland wallpaper application config"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "wbg.output-dir", "", "Output directory (default: ~/.config/wbg)")
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
		{Name: "wbg.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/wbg)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("wbg", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "wbg"))
}

// Generate creates the shell script for wbg wallpaper persistence.
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

	files["tinct-wbg.sh"] = configContent

	return files, nil
}

// generateConfig creates the configuration file.
func (p *Plugin) generateConfig(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("wbg", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	tmplContent, fromCustom, err := loader.Load("tinct-wbg.sh.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read config template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct-wbg.sh.tmpl\n")
	}

	tmpl, err := template.New("config").Funcs(utils.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute config template: %w", err)
	}

	return buf.Bytes(), nil
}

// Hooks declares wbg's pre/post-execute behaviour. wbg is optional;
// chmod +x and wallpaper application use the spec's MakeExecutable and
// Wallpaper callback. The bespoke setWallpaper logic (kill-and-restart
// with detached process release) is exposed via Wallpaper because wbg
// is a foreground process that must be restarted to change wallpaper.
func (p *Plugin) Hooks() hooks.Spec {
	return hooks.Spec{
		OptionalBinaries:  []string{"wbg"},
		AutoCreateDir:     true,
		MakeExecutable:    []string{"tinct-wbg.sh"},
		SupportsWallpaper: true,
		Wallpaper:         p.setWallpaper,
	}
}

// setWallpaper applies the wallpaper using wbg.
// Since wbg is a foreground process, this kills any existing instance
// and starts a new one in the background.
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

	// Check if wbg is available.
	wbgPath, err := exec.LookPath("wbg")
	if err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Skipping wallpaper application (wbg not found)\n")
		}
		return nil
	}

	// Kill any existing wbg process.
	// We use killall which is safe even if no wbg is running.
	killCmd := exec.CommandContext(ctx, "killall", "wbg")
	_ = killCmd.Run() //nolint:errcheck // Explicitly ignore - wbg may not be running

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Running: wbg %s\n", absPath)
	}

	// Start wbg in the background as a detached process.
	// wbg is a foreground application, so we need to start it detached.
	// #nosec G204 -- wbg is a system command with validated absolute path for wallpaper.
	cmd := exec.Command(wbgPath, absPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start wbg: %w", err)
	}

	// Release the process so it continues running after we exit.
	if err := cmd.Process.Release(); err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: failed to release wbg process: %v\n", err)
		}
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Set wallpaper using wbg: %s\n", absPath)
	}

	return nil
}
