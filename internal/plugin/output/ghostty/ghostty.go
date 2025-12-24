// Package ghostty provides an output plugin for Ghostty terminal emulator colour themes.
package ghostty

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"text/template"

	"github.com/mitchellh/go-ps"
	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/shared/utils"
	tmplloader "github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/jmylchreest/tinct/internal/version"
)

//go:embed *.tmpl
var templates embed.FS

// GetEmbeddedTemplates returns the embedded template filesystem.
// This is used by the template management commands.
func GetEmbeddedTemplates() embed.FS {
	return templates
}

// Plugin implements the output.Plugin interface for Ghostty.
type Plugin struct {
	outputDir string
	reload    bool
	verbose   bool
}

// New creates a new Ghostty output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		outputDir: "",
		reload:    true,
		verbose:   false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "ghostty"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Ghostty terminal emulator theme"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.outputDir, "ghostty.output-dir", "", "Output directory (default: ~/.config/ghostty/themes)")
	cmd.Flags().BoolVar(&p.reload, "ghostty.reload", true, "Reload ghostty after generation (sends SIGUSR2)")
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
		{Name: "ghostty.output-dir", Type: "string", Default: "", Description: "Output directory (default: ~/.config/ghostty/themes)", Required: false},
		{Name: "ghostty.reload", Type: "bool", Default: "true", Description: "Reload ghostty after generation", Required: false},
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
		return ".config/ghostty/themes"
	}
	return filepath.Join(home, ".config", "ghostty", "themes")
}

// Generate creates the theme files.
// Returns map of filename -> content.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	if themeData == nil {
		return nil, fmt.Errorf("theme data cannot be nil")
	}

	files := make(map[string][]byte)

	// Generate tinct theme file.
	themeContent, err := p.generateTheme(themeData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate theme: %w", err)
	}
	files["tinct.conf"] = themeContent

	return files, nil
}

// generateTheme creates the Ghostty theme file.
func (p *Plugin) generateTheme(themeData *colour.ThemeData) ([]byte, error) {
	// Load template with custom override support.
	loader := tmplloader.New("ghostty", templates)
	if p.verbose {
		loader.WithVerbose(true, utils.NewVerboseLogger(os.Stderr))
	}
	tmplContent, fromCustom, err := loader.Load("tinct.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read theme template: %w", err)
	}

	if p.verbose && fromCustom {
		fmt.Fprintf(os.Stderr, "   Using custom template for tinct.tmpl\n")
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

// PreExecute checks if ghostty is available before generating the theme.
// Implements the output.PreExecuteHook interface.
func (p *Plugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	// Check if ghostty executable exists on PATH.
	_, err = exec.LookPath("ghostty")
	if err != nil {
		return true, "ghostty executable not found on $PATH", nil
	}

	// Check if config directory exists (create it if not).
	configDir := p.DefaultOutputDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0o750); err != nil {
			return true, fmt.Sprintf("ghostty config directory does not exist and cannot be created: %s", configDir), nil
		}
	}

	return false, "", nil
}

// PostExecute reloads ghostty configuration after successful theme generation.
// Implements the output.PostExecuteHook interface.
func (p *Plugin) PostExecute(_ context.Context, execCtx output.ExecutionContext, _ []string) error {
	// Skip reload in dry-run mode or if reload is disabled.
	if execCtx.DryRun || !p.reload {
		return nil
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Reloading ghostty configuration...\n")
	}

	// Find all ghostty processes.
	pids, err := findProcessByName("ghostty")
	if err != nil {
		return fmt.Errorf("failed to find ghostty processes: %w", err)
	}

	if len(pids) == 0 {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   No running ghostty instances found\n")
		}
		return nil
	}

	// Send SIGUSR2 to all ghostty processes to reload configuration.
	successCount := 0
	for _, pid := range pids {
		process, err := os.FindProcess(pid)
		if err != nil {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Warning: Failed to find process %d: %v\n", pid, err)
			}
			continue
		}

		if err := process.Signal(syscall.SIGUSR2); err != nil {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Warning: Failed to send SIGUSR2 to ghostty process %d: %v\n", pid, err)
			}
			continue
		}

		successCount++
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Sent SIGUSR2 to ghostty process %d\n", pid)
		}
	}

	if p.verbose && successCount > 0 {
		fmt.Fprintf(os.Stderr, "   Successfully reloaded %d ghostty instance(s)\n", successCount)
	}

	return nil
}

// findProcessByName finds all PIDs of processes with the given name.
// Uses go-ps library for cross-platform process discovery.
func findProcessByName(name string) ([]int, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %w", err)
	}

	var pids []int
	for _, p := range processes {
		if p.Executable() == name {
			pids = append(pids, p.Pid())
		}
	}

	return pids, nil
}
