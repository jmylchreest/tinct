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
	"github.com/spf13/pflag"

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
	generateBackend     string
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
	generateCmd.Flags().StringVar(&generateBackend, "backend", "kmeans", "Colour extraction backend (kmeans)")
	generateCmd.Flags().BoolVarP(&generateVerbose, "verbose", "v", false, "Verbose output")
	generateCmd.Flags().StringToStringVar(&generatePluginArgs, "plugin-args", nil, "Plugin-specific arguments (key=value format, repeatable for multiple plugins)")

	// Override Help method to generate dynamic help text with filtered flags.
	generateCmd.SetHelpFunc(customGenerateHelp)
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

// customGenerateHelp provides dynamic help that filters flags based on selected plugins.
func customGenerateHelp(cmd *cobra.Command, args []string) {

	// Parse which input and output plugins are requested
	inputPlugin, _ := cmd.Flags().GetString("input")
	outputPlugins, _ := cmd.Flags().GetStringSlice("outputs")

	// If specific plugins are selected, show plugin-specific help and filter flags
	if inputPlugin != "" || (len(outputPlugins) > 0 && outputPlugins[0] != pluginTypeAll) {
		// Build plugin-specific help text
		cmd.Long = buildPluginSpecificHelp(inputPlugin, outputPlugins)

		// Create a filtered flag set
		filteredFlags := make(map[string]bool)

		// Build a set of ALL plugin flags from ALL plugins
		allPluginFlags := make(map[string]bool)
		for _, plugin := range sharedPluginManager.AllInputPlugins() {
			for _, fh := range plugin.GetFlagHelp() {
				allPluginFlags[fh.Name] = true
				if fh.Shorthand != "" {
					allPluginFlags[fh.Shorthand] = true
				}
			}
		}
		for _, plugin := range sharedPluginManager.AllOutputPlugins() {
			for _, fh := range plugin.GetFlagHelp() {
				allPluginFlags[fh.Name] = true
				if fh.Shorthand != "" {
					allPluginFlags[fh.Shorthand] = true
				}
			}
		}

		// Add core command flags (any flag that's not a plugin flag)
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if !allPluginFlags[flag.Name] {
				filteredFlags[flag.Name] = true
			}
		})

		// Add selected input plugin flags
		if inputPlugin != "" {
			if plugin, ok := sharedPluginManager.GetInputPlugin(inputPlugin); ok {
				flagHelp := plugin.GetFlagHelp()
				for _, fh := range flagHelp {
					filteredFlags[fh.Name] = true
					if fh.Shorthand != "" {
						filteredFlags[fh.Shorthand] = true
					}
				}
			}
		}

		// Add selected output plugin flags
		if len(outputPlugins) > 0 && outputPlugins[0] != pluginTypeAll {
			for _, outputName := range outputPlugins {
				if plugin, ok := sharedPluginManager.GetOutputPlugin(outputName); ok {
					flagHelp := plugin.GetFlagHelp()
					for _, fh := range flagHelp {
						filteredFlags[fh.Name] = true
						if fh.Shorthand != "" {
							filteredFlags[fh.Shorthand] = true
						}
					}
				}
			}
		}

		// Temporarily hide flags that aren't in our filtered set
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			if !filteredFlags[flag.Name] {
				flag.Hidden = true
			}
		})

		// Call default help with filtered flags
		cmd.Parent().HelpFunc()(cmd, args)

		// Restore all flags to visible (for next help invocation)
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			flag.Hidden = false
		})

		return
	}

	// No specific plugins selected - show general help with plugin lists and examples
	cmd.Long = buildGeneralHelp()
	cmd.Parent().HelpFunc()(cmd, args)
}

// buildPluginSpecificHelp builds help text for specific plugin(s).
func buildPluginSpecificHelp(inputPlugin string, outputPlugins []string) string {
	var help strings.Builder

	// Show input plugin description if specified
	if inputPlugin != "" {
		if plugin, ok := sharedPluginManager.GetInputPlugin(inputPlugin); ok {
			help.WriteString(fmt.Sprintf("Input Plugin: %s\n", plugin.Name()))
			help.WriteString(fmt.Sprintf("%s\n\n", plugin.Description()))
		}
	}

	// Show output plugin descriptions if specified
	if len(outputPlugins) > 0 && outputPlugins[0] != pluginTypeAll {
		help.WriteString("Output Plugins:\n")
		for _, outputName := range outputPlugins {
			if plugin, ok := sharedPluginManager.GetOutputPlugin(outputName); ok {
				help.WriteString(fmt.Sprintf("  %s - %s\n", plugin.Name(), plugin.Description()))
			}
		}
		help.WriteString("\n")
	}

	return help.String()
}

// buildGeneralHelp builds the general help text with plugin lists and examples.
func buildGeneralHelp() string {
	help := `Generate configuration files for various applications from a colour palette.

The palette can be created from an image, loaded from a file, or built from
command-line colour specifications. Generated files are written to their
default locations or a custom output directory.

Input Plugins:
`

	// List available input plugins
	for _, pluginName := range sharedPluginManager.ListInputPlugins() {
		if plugin, ok := sharedPluginManager.GetInputPlugin(pluginName); ok {
			help += fmt.Sprintf("  %-12s - %s\n", plugin.Name(), plugin.Description())
		}
	}

	help += "\nOutput Plugins:\n"

	// List enabled output plugins
	enabledPlugins := sharedPluginManager.FilterOutputPlugins()
	if len(enabledPlugins) > 0 {
		for _, plugin := range enabledPlugins {
			help += fmt.Sprintf("  %-12s - %s\n", plugin.Name(), plugin.Description())
		}
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
    --colour background=#1e1e2e \
    --colour foreground=#cdd6f4 \
    --colour accent1=#89b4fa \
    --outputs hyprland

  # Generate image with Google Imagen and extract colors
  tinct generate -i google-genai --prompt "sunset over mountains"

  # Extract from image with custom colour count
  tinct generate -i image -p wallpaper.jpg -c 32 --preview

Use 'tinct generate -i <plugin> --help' to see plugin-specific options.`

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
