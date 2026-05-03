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
	"github.com/jmylchreest/tinct/pkg/plugin/paths"
	"github.com/jmylchreest/tinct/pkg/util/semver"
)

//go:embed *.tmpl templates
var templates embed.FS

// versionRegex matches the hyprpaper version from "hyprpaper -h" output.
// Example: "┏ hyprpaper v0.8.0".
var versionRegex = regexp.MustCompile(`hyprpaper\s+v(\d+\.\d+(?:\.\d+)?)`)

// Cached version to avoid repeated exec calls.
var (
	cachedVersion     string
	cachedVersionOnce sync.Once
)

// tinctConfigFilename is the name of the tinct-generated config file.
const tinctConfigFilename = "tinct-hyprpaper.conf"

// configPaths holds the resolved paths for hyprpaper configuration files.
type configPaths struct {
	tinctConfig   string // Path to tinct-hyprpaper.conf
	hyprpaperConf string // Path to hyprpaper.conf
	backup        string // Path to hyprpaper.conf.bak
}

// resolveConfigPaths finds the tinct config in written files and returns all related paths.
// Returns nil if tinct config was not found in written files.
func resolveConfigPaths(writtenFiles []string) *configPaths {
	var tinctConfigPath string
	for _, f := range writtenFiles {
		if filepath.Base(f) == tinctConfigFilename {
			tinctConfigPath = f
			break
		}
	}
	if tinctConfigPath == "" {
		return nil
	}

	configDir := filepath.Dir(tinctConfigPath)
	return &configPaths{
		tinctConfig:   tinctConfigPath,
		hyprpaperConf: filepath.Join(configDir, "hyprpaper.conf"),
		backup:        filepath.Join(configDir, "hyprpaper.conf.bak"),
	}
}

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
	return "Hyprpaper wallpaper manager config"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
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
func (p *Plugin) GetEmbeddedFS() any {
	return templates
}

// GetTargetVersion returns the installed hyprpaper version.
// Implements the output.VersionedPlugin interface.
// This enables version-specific templates for hyprpaper configuration changes.
func (p *Plugin) GetTargetVersion() string {
	cachedVersionOnce.Do(func() {
		cachedVersion = detectHyprpaperVersion(p.verbose)
	})
	return cachedVersion
}

// detectHyprpaperVersion runs "hyprpaper -h" to get the version.
func detectHyprpaperVersion(_ bool) string {
	// hyprpaper -h outputs version in the header: "┏ hyprpaper v0.8.0".
	cmdOutput, err := exec.Command("hyprpaper", "-h").Output()
	if err == nil {
		if v := parseHyprpaperVersion(string(cmdOutput)); v != "" {
			return v
		}
	}

	return ""
}

// parseHyprpaperVersion extracts the version number from hyprpaper help text.
func parseHyprpaperVersion(helpText string) string {
	matches := versionRegex.FindStringSubmatch(helpText)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

// isVersion08OrNewer checks if the detected version is 0.8.0 or newer.
func (p *Plugin) isVersion08OrNewer() bool {
	targetVersion := p.GetTargetVersion()
	if targetVersion == "" {
		return false
	}
	v := semver.Parse(targetVersion)
	v08 := semver.Parse("0.8.0")
	if v == nil || v08 == nil {
		return false
	}
	return v.GreaterThanOrEqual(v08)
}

// isVersion081OrNewer checks if the detected version is 0.8.1 or newer.
// Version 0.8.1 reintroduced support for source= includes.
func (p *Plugin) isVersion081OrNewer() bool {
	targetVersion := p.GetTargetVersion()
	if targetVersion == "" {
		return false
	}
	v := semver.Parse(targetVersion)
	v081 := semver.Parse("0.8.1")
	if v == nil || v081 == nil {
		return false
	}
	return v.GreaterThanOrEqual(v081)
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "hyprpaper.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/hypr)", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	return pluginconfig.Resolve("hyprpaper", "output_dir", p.outputDir,
		filepath.Join(paths.XDGConfigDir(), "hypr"))
}

// Generate creates the configuration file.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate config file.
	configContent, err := p.generateConfig(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate config: %w", err)
	}

	files[tinctConfigFilename] = configContent

	return files, nil
}

// generateConfig creates the configuration file.
func (p *Plugin) generateConfig(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override and versioned template support.
	loader := tmplloader.New("hyprpaper", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}

	// Enable versioned template selection based on installed hyprpaper version.
	targetVersion := p.GetTargetVersion()
	if targetVersion != "" {
		loader.WithTargetVersion(targetVersion)
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Detected hyprpaper version: %s\n", targetVersion)
		}
	} else if p.verbose {
		fmt.Fprintf(os.Stderr, "   Could not detect hyprpaper version, using default templates\n")
	}

	tmplContent, fromCustom, err := loader.Load("tinct.conf.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read config template: %w", err)
	}

	// Log if using custom template.
	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.conf.tmpl\n")
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

// PreExecute checks if hyprctl is available and config directory exists.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if hyprctl executable exists on PATH (needed for wallpaper control).
	_, err = exec.LookPath("hyprctl")
	if err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: hyprctl not found - wallpaper setting will not be available\n")
		}
	}

	// Check if config directory exists (create it if not).
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0o750); err != nil {
			return true, fmt.Sprintf("hypr config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}

// PostExecute applies the wallpaper using hyprpaper after files are written.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(ctx context.Context, execCtx output.ExecutionContext, writtenFiles []string) error {
	p.handleConfigManagement(writtenFiles)

	if execCtx.WallpaperPath != "" {
		p.applyWallpaper(ctx, execCtx.WallpaperPath)
	}

	return nil
}

// handleConfigManagement manages config files based on hyprpaper version.
func (p *Plugin) handleConfigManagement(writtenFiles []string) {
	switch {
	case p.isVersion081OrNewer():
		// v0.8.1+ reintroduced source= support.
		if err := p.restoreSourceConfig(writtenFiles); err != nil && p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: failed to restore source config: %v\n", err)
		}
	case p.isVersion08OrNewer():
		// v0.8.0 only - setup symlink since source= is not supported.
		if err := p.setupConfigSymlink(writtenFiles); err != nil && p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: failed to setup config symlink: %v\n", err)
		}
	}
}

// applyWallpaper attempts to set the wallpaper, logging any errors.
func (p *Plugin) applyWallpaper(ctx context.Context, wallpaperPath string) {
	// Check if hyprpaper is running before trying to set wallpaper.
	// Note: v0.8.0 removed listloaded/listactive, so we skip the check and just try.
	if !p.isVersion08OrNewer() {
		cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "listloaded")
		if err := cmd.Run(); err != nil {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Skipping wallpaper application (hyprpaper not running)\n")
			}
			return
		}
	}

	if err := p.setWallpaper(ctx, wallpaperPath); err != nil && p.verbose {
		fmt.Fprintf(os.Stderr, "   Failed to set wallpaper: %v\n", err)
	}
}

// setWallpaper applies the wallpaper using hyprpaper.
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

	// Check if we're using 0.8.0+ IPC syntax.
	if p.isVersion08OrNewer() {
		return p.setWallpaperV08(ctx, absPath)
	}

	return p.setWallpaperLegacy(ctx, absPath)
}

// setWallpaperV08 applies the wallpaper using hyprpaper 0.8.0+ IPC syntax.
// New format: hyprctl hyprpaper wallpaper '[monitor],[path],[fit_mode]'.
func (p *Plugin) setWallpaperV08(ctx context.Context, absPath string) error {
	// In 0.8.0+, the wallpaper command format is: "monitor,path,fit_mode"
	// Empty monitor = fallback/wildcard, fit_mode is optional.
	arg := "," + absPath + ","
	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Running: hyprctl hyprpaper wallpaper '%s'\n", arg)
	}

	// #nosec G204 -- hyprctl is a system command with validated absolute path for wallpaper.
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "wallpaper", arg)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set wallpaper: %w (output: %s)", err, strings.TrimSpace(string(cmdOutput)))
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Set wallpaper using hyprpaper 0.8+ IPC: %s (output: %s)\n", absPath, strings.TrimSpace(string(cmdOutput)))
	}

	return nil
}

// setWallpaperLegacy applies the wallpaper using pre-0.8.0 hyprpaper IPC syntax.
func (p *Plugin) setWallpaperLegacy(ctx context.Context, absPath string) error {
	assignments := p.getAssignmentsOrWildcard(ctx)

	// Unload all existing wallpapers to clear any cached images.
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "unload", "all")
	_ = cmd.Run() //nolint:errcheck // Explicitly ignore errors - wallpapers might not be loaded

	// Preload the new wallpaper.
	// #nosec G204 -- absPath is validated to be a safe file path.
	cmd = exec.CommandContext(ctx, "hyprctl", "hyprpaper", "preload", absPath)
	if cmdOutput, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to preload wallpaper: %w (output: %s)", err, string(cmdOutput))
	}

	// Set the wallpaper on each monitor.
	p.setWallpaperOnMonitors(ctx, assignments, absPath)

	p.logWallpaperResult(assignments, absPath)
	return nil
}

// getAssignmentsOrWildcard returns monitor assignments, falling back to wildcard.
func (p *Plugin) getAssignmentsOrWildcard(ctx context.Context) []string {
	assignments, err := p.getActiveWallpaperAssignments(ctx)
	if err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Warning: failed to get active assignments, using wildcard: %v\n", err)
		}
		return []string{""}
	}
	if len(assignments) == 0 {
		return []string{""}
	}
	return assignments
}

// monitorDisplayName returns a human-readable name for a monitor (or "wildcard" for empty).
func monitorDisplayName(monitor string) string {
	if monitor == "" {
		return "wildcard"
	}
	return monitor
}

// setWallpaperOnMonitors sets the wallpaper on each monitor assignment.
func (p *Plugin) setWallpaperOnMonitors(ctx context.Context, assignments []string, absPath string) {
	for _, monitor := range assignments {
		// #nosec G204 -- hyprctl is a system command with validated absolute path for wallpaper.
		cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "wallpaper", monitor+","+absPath)
		if err := cmd.Run(); err != nil {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Warning: failed to set wallpaper on %s: %v\n", monitorDisplayName(monitor), err)
			}
			continue
		}
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Set wallpaper on %s\n", monitorDisplayName(monitor))
		}
	}
}

// logWallpaperResult logs the final wallpaper setting result.
func (p *Plugin) logWallpaperResult(assignments []string, absPath string) {
	if !p.verbose {
		return
	}
	if len(assignments) == 1 && assignments[0] == "" {
		fmt.Fprintf(os.Stderr, "   Set wallpaper using hyprpaper (wildcard): %s\n", absPath)
	} else {
		fmt.Fprintf(os.Stderr, "   Set wallpaper using hyprpaper on %d monitor(s): %s\n", len(assignments), absPath)
	}
}

// getActiveWallpaperAssignments retrieves the current monitor assignments from listactive.
// Returns a list of monitor names (or empty string for wildcard).
func (p *Plugin) getActiveWallpaperAssignments(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "hyprctl", "hyprpaper", "listactive")
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to query active wallpapers: %w", err)
	}

	// Parse the cmdOutput line by line.
	// Format: "MONITOR = /path/to/wallpaper".
	// Wildcard format: " = /path/to/wallpaper" (empty monitor name).
	lines := strings.Split(string(cmdOutput), "\n")
	assignments := make([]string, 0, len(lines))

	for _, line := range lines {
		// Don't trim yet - we need to preserve leading space for wildcard detection.
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split on " = " to get monitor name.
		parts := strings.SplitN(line, " = ", 2)
		if len(parts) != 2 {
			continue
		}

		// Now trim the monitor name (will be empty string for wildcard " = /path").
		monitor := strings.TrimSpace(parts[0])
		assignments = append(assignments, monitor)
	}

	return assignments, nil
}
