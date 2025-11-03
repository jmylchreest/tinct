// Package hyprpaper provides an output plugin for Hyprpaper wallpaper manager configuration.
package hyprpaper

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

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/common"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/spf13/cobra"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Hyprpaper.
type Plugin struct {
	outputDir string
	verbose   bool
}

// New creates a new Hyprpaper output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "hyprpaper"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Generate Hyprpaper wallpaper manager configuration"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "hyprpaper.output-dir", "", "Output directory (default: ~/.config/hypr)")
}

// SetVerbose enables or disables verbose logging for the plugin.
// Implements the output.VerbosePlugin interface.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetEmbeddedFS returns the embedded template filesystem.
// Implements the output.TemplateProvider interface.
func (p *Plugin) GetEmbeddedFS() interface{} {
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
		return ".config/hypr"
	}
	return filepath.Join(home, ".config", "hypr")
}

// Generate creates the configuration file.
// Returns map of filename -> content
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate config file
	configContent, err := p.generateConfig(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate config: %w", err)
	}

	files["tinct-hyprpaper.conf"] = configContent

	return files, nil
}

// generateConfig creates the configuration file.
func (p *Plugin) generateConfig(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support
	loader := tmplloader.New("hyprpaper", templates)
	if p.verbose {
		loader.WithVerbose(true, common.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read config template: %w", err)
	}

	// Log if using custom template
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.conf.tmpl\n")
	}

	tmpl, err := template.New("config").Funcs(common.TemplateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, themeData); err != nil {
		return nil, fmt.Errorf("failed to execute config template: %w", err)
	}

	return buf.Bytes(), nil
}

// PreExecute checks if the config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Check if config directory exists (create it if not)
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return true, fmt.Sprintf("hypr config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}

// PostExecute applies the wallpaper using hyprpaper after files are written.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, writtenFiles []string) error {
	// If we have a wallpaper path, try to apply it
	if execCtx.WallpaperPath != "" {
		// Check if hyprpaper is running before trying to set wallpaper
		cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "listloaded")
		if err := cmd.Run(); err != nil {
			// hyprpaper not running - skip wallpaper application
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Skipping wallpaper application (hyprpaper not running)\n")
			}
			return nil
		}

		if err := p.setWallpaper(ctx, execCtx.WallpaperPath); err != nil {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Failed to set wallpaper: %v\n", err)
			}
			// Don't return error - wallpaper setting is optional
			return nil
		}
	}

	return nil
}

// setWallpaper applies the wallpaper using hyprpaper
func (p *Plugin) setWallpaper(ctx context.Context, wallpaperPath string) error {
	// Make the path absolute
	absPath, err := filepath.Abs(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Get current wallpaper assignments (monitors and/or wildcard)
	assignments, err := p.getActiveWallpaperAssignments(ctx)
	if err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: failed to get active assignments, using wildcard: %v\n", err)
		}
		// Fallback to wildcard if we can't get assignments
		assignments = []string{""}
	}

	// If no assignments found, use wildcard
	if len(assignments) == 0 {
		assignments = []string{""}
	}

	// First, unload all existing wallpapers to clear any cached images
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "unload", "all")
	cmd.Run() // Ignore errors - wallpapers might not be loaded

	// Preload the new wallpaper
	cmd = exec.CommandContext(ctx, "hyprctl", "hyprpaper", "preload", absPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to preload wallpaper: %w (output: %s)", err, string(output))
	}

	// Set the wallpaper using the same assignments (monitors or wildcard)
	successCount := 0
	for _, monitor := range assignments {
		cmd = exec.CommandContext(ctx, "hyprctl", "hyprpaper", "wallpaper", monitor+","+absPath)
		if err := cmd.Run(); err != nil {
			if p.verbose {
				monitorName := monitor
				if monitorName == "" {
					monitorName = "wildcard"
				}
				fmt.Fprintf(os.Stderr, "   Warning: failed to set wallpaper on %s: %v\n", monitorName, err)
			}
			continue
		}
		successCount++
		if p.verbose {
			monitorName := monitor
			if monitorName == "" {
				monitorName = "wildcard"
			}
			fmt.Fprintf(os.Stderr, "   Set wallpaper on %s\n", monitorName)
		}
	}

	if p.verbose {
		if len(assignments) == 1 && assignments[0] == "" {
			fmt.Fprintf(os.Stderr, "   Set wallpaper using hyprpaper (wildcard): %s\n", absPath)
		} else {
			fmt.Fprintf(os.Stderr, "   Set wallpaper using hyprpaper on %d monitor(s): %s\n", len(assignments), absPath)
		}
	}

	return nil
}

// getActiveWallpaperAssignments retrieves the current monitor assignments from listactive
// Returns a list of monitor names (or empty string for wildcard)
func (p *Plugin) getActiveWallpaperAssignments(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "listactive")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query active wallpapers: %w", err)
	}

	// Parse the output line by line
	// Format: "MONITOR = /path/to/wallpaper"
	// Wildcard format: " = /path/to/wallpaper" (empty monitor name)
	lines := strings.Split(string(output), "\n")
	assignments := make([]string, 0, len(lines))

	for _, line := range lines {
		// Don't trim yet - we need to preserve leading space for wildcard detection
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split on " = " to get monitor name
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) != 2 {
			continue
		}

		// Now trim the monitor name (will be empty string for wildcard " = /path")
		monitor := strings.TrimSpace(parts[0])
		assignments = append(assignments, monitor)
	}

	return assignments, nil
}
