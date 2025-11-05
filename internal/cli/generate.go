// Package cli provides the command-line interface for Tinct.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/version"
)

const (
	pluginTypeAll    = "all"
	pluginTypeOutput = "output"
)

// isValidPath checks if a path is safe to use in commands.
func isValidPath(path string) bool {
	// Reject paths with suspicious characters
	if strings.Contains(path, "..") || strings.ContainsAny(path, "|&;`$()") {
		return false
	}
	// Clean the path and ensure it matches
	cleaned := filepath.Clean(path)
	return cleaned == path
}

var (
	// Generate command flags.
	generateInputPlugin string
	generateOutputs     []string
	generateDryRun      bool
	generatePreview     bool
	generateSavePalette string
	generateVerbose     bool
	generatePluginArgs  map[string]string
)

// generateCmd represents the generate command.
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate configuration files from a colour palette",
	Long:  "", // Set dynamically in Help()
	RunE:  runGenerate,
}

func init() {
	// Note: Plugin manager is initialised in root.go and flags are registered there.

	// Input plugin selection (required).
	generateCmd.Flags().StringVarP(&generateInputPlugin, "input", "i", "", "Input plugin (required: image, file)")
	_ = generateCmd.MarkFlagRequired("input") // Error only occurs if flag doesn't exist, which is impossible here

	// Output plugin selection.
	generateCmd.Flags().StringSliceVarP(&generateOutputs, "outputs", "o", []string{pluginTypeAll}, "Output plugins (comma-separated or 'all')")

	// General options.
	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "Preview without writing files")
	generateCmd.Flags().BoolVar(&generatePreview, "preview", false, "Show colour palette preview")
	generateCmd.Flags().StringVar(&generateSavePalette, "save-palette", "", "Save palette to file (JSON)")
	generateCmd.Flags().BoolVarP(&generateVerbose, "verbose", "v", false, "Verbose output")
	generateCmd.Flags().StringToStringVar(&generatePluginArgs, "plugin-args", nil, "Plugin-specific arguments (key=value format, repeatable for multiple plugins)")

	// Override Help method to generate dynamic help text.
	generateCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		cmd.Long = buildGenerateHelp()
		cmd.Parent().HelpFunc()(cmd, args)
	})
}

// runGenerate executes the generate command.
func runGenerate(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	// Phase 1: Load and configure plugins.
	if err := loadAndConfigurePlugins(); err != nil {
		return err
	}

	// Phase 2: Get and validate input plugin.
	inputPlugin, err := getAndValidateInputPlugin()
	if err != nil {
		return err
	}

	// Phase 3: Generate input palette.
	rawPalette, wallpaperPath, err := generateInputPalette(ctx, inputPlugin)
	if err != nil {
		return err
	}

	// Phase 4: Categorize the palette.
	palette := categorizePalette(rawPalette, inputPlugin)

	// Phase 5: Handle palette output (preview/save).
	if err := handlePaletteOutput(palette); err != nil {
		return err
	}

	// Phase 6: Select output plugins.
	outputPlugins, err := selectOutputPlugins()
	if err != nil {
		return err
	}

	// Phase 7: Run global pre-hook.
	if err := runGlobalHookScript(ctx, "pre-generate", generateVerbose, generateDryRun); err != nil {
		if generateVerbose {
			fmt.Fprintf(os.Stderr, " Global pre-hook failed: %v\n", err)
		}
	}

	// Phase 8: Validate plugins and run pre-execute hooks.
	executions := preparePluginExecutions(ctx, outputPlugins)

	// Phase 9: Generate and write files.
	successCount := generateAndWriteFiles(executions, palette, wallpaperPath)

	// Phase 10: Run post-execute hooks.
	if !generateDryRun {
		runPostExecutionHooks(ctx, executions, wallpaperPath)

		// Run global post-hook.
		if err := runGlobalHookScript(ctx, "post-generate", generateVerbose, generateDryRun); err != nil {
			if generateVerbose {
				fmt.Fprintf(os.Stderr, " Global post-hook failed: %v\n", err)
			}
		}
	}

	// Phase 11: Print summary.
	return printGenerationSummary(successCount)
}

// runGlobalHookScript executes a global hook script if it exists.
// Looks for scripts at ~/.config/tinct/hooks/{hook-name}.sh.
func runGlobalHookScript(ctx context.Context, hookName string, verbose, dryRun bool) error {
	if dryRun {
		return nil // Don't run hooks in dry-run mode
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil // Silently skip if we can't get config dir
	}

	hookPath := filepath.Join(configDir, "tinct", "hooks", fmt.Sprintf("%s.sh", hookName))

	// Check if hook exists and is executable.
	info, err := os.Stat(hookPath)
	if os.IsNotExist(err) {
		return nil // No hook, that's fine
	}
	if err != nil {
		return err
	}

	// Check if executable.
	if info.Mode()&0o111 == 0 {
		if verbose {
			fmt.Fprintf(os.Stderr, " Hook script exists but is not executable: %s\n", hookPath)
		}
		return nil
	}

	// Validate path to prevent command injection.
	if !isValidPath(hookPath) {
		return fmt.Errorf("invalid hook path: contains suspicious characters")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "→ Running global %s hook: %s\n", hookName, hookPath)
	}

	// Execute the script.
	hookCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// #nosec G204 -- hookPath is validated to be a safe file path
	cmd := exec.CommandContext(hookCtx, hookPath)

	// Set environment variables for the hook script.
	env := append(os.Environ(),
		"TINCT_HOOK="+hookName,
		"TINCT_VERSION="+getVersion(),
	)

	// Only set TINCT_VERBOSE if verbose is true (script can check if var exists).
	if verbose {
		env = append(env, "TINCT_VERBOSE=true")
	}

	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("%w: %s", err, string(output))
		}
		return err
	}

	// Always show script output - the script itself decides what to output based on TINCT_VERBOSE.
	if len(output) > 0 {
		fmt.Fprintf(os.Stderr, "  %s\n", string(output))
	}

	return nil
}

// getVersion returns the current version.
func getVersion() string {
	return version.Short()
}

// buildGenerateHelp dynamically builds the help text with enabled plugins.
func buildGenerateHelp() string {
	help := `Generate configuration files for various applications from a colour palette.

The palette can be created from an image, loaded from a file, or built from
command-line colour specifications. Generated files are written to their
default locations or a custom output directory.

Input Plugins:
  image        - Extract colours from an image file
  file         - Load palette from file or build from colour specifications
  remote-json  - Fetch colour palette from remote JSON source with optional JSONPath queries
  remote-css   - Fetch colour palette from remote CSS source (extracts CSS variables and color values)

Output Plugins:
`

	// List enabled plugins (filtered by shared manager).
	enabledPlugins := []string{}
	for _, plugin := range sharedPluginManager.FilterOutputPlugins() {
		enabledPlugins = append(enabledPlugins, fmt.Sprintf("  %-12s - %s", plugin.Name(), plugin.Description()))
	}

	if len(enabledPlugins) > 0 {
		help += strings.Join(enabledPlugins, "\n") + "\n"
	} else {
		help += "  (no enabled plugins)\n"
	}

	help += `  all          - Run all available output plugins (default)

Examples:
  # From image - generate all outputs
  tinct generate --input image -p wallpaper.jpg

  # From palette file - specific output
  tinct generate --input file -p theme.json --outputs hyprland

  # From colours - generate for specific app
  tinct generate --input file \
    --file.background '#1e1e2e' \
    --file.foreground '#cdd6f4' \
    --file.accent1 '#89b4fa' \
    --outputs hyprland

  # Extract from image with custom colour count
  tinct generate -i image -p wallpaper.jpg -c 32 --preview`

	return help
}

// savePalette saves a categorised palette to a JSON file.
func savePalette(palette *colour.CategorisedPalette, path string) error {
	data, err := palette.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal palette: %w", err)
	}

	// Ensure directory exists.
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil { // #nosec G301 - Output directory needs standard permissions
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// setPluginArgs sets custom arguments for a plugin.
func setPluginArgs(mgr *manager.Manager, pluginName, pluginType, argsJSON string) error {
	// Parse JSON args.
	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Get the plugin and set args based on type.
	if pluginType == pluginTypeOutput {
		plugin, ok := mgr.GetOutputPlugin(pluginName)
		if !ok {
			return fmt.Errorf("plugin not found")
		}

		// Check if it's an external plugin that supports args.
		if extPlugin, ok := plugin.(*manager.ExternalOutputPlugin); ok {
			extPlugin.SetArgs(args)
		}
	}
	// Could add input plugin support here in the future.

	return nil
}

// setPluginDryRun sets dry-run mode for a plugin.
func setPluginDryRun(mgr *manager.Manager, pluginName, pluginType string, dryRun bool) error {
	// Get the plugin and set dry-run based on type.
	if pluginType == pluginTypeOutput {
		plugin, ok := mgr.GetOutputPlugin(pluginName)
		if !ok {
			return fmt.Errorf("plugin not found")
		}

		// Check if it's an external plugin that supports dry-run.
		if extPlugin, ok := plugin.(*manager.ExternalOutputPlugin); ok {
			extPlugin.SetDryRun(dryRun)
		}
	}
	// Could add input plugin support here in the future.

	return nil
}

// writeFile writes content to a file, creating directories as needed.
func writeFile(path string, content []byte, verbose bool) error {
	// Expand ~ to home directory.
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Ensure directory exists.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil { // #nosec G301 - Output directory needs standard permissions
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists and create backup.
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".backup"
		if err := os.Rename(path, backupPath); err != nil {
			// If backup fails, continue anyway.
			if verbose {
				fmt.Fprintf(os.Stderr, "    Could not create backup: %v\n", err)
			}
		} else if verbose {
			fmt.Fprintf(os.Stderr, "   Created backup: %s\n", backupPath)
		}
	}

	// Write the file.
	if err := os.WriteFile(path, content, 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
