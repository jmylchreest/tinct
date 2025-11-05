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
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/plugin/output"
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
func runGenerate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Reload shared plugin manager config from lock file if available (overrides env).
	lock, _, err := loadPluginLock()
	if err == nil && lock != nil {
		config := manager.Config{
			EnabledPlugins:  lock.EnabledPlugins,
			DisabledPlugins: lock.DisabledPlugins,
		}
		sharedPluginManager.UpdateConfig(config)

		// Register external plugins from lock file.
		if lock.ExternalPlugins != nil {
			for _, meta := range lock.ExternalPlugins {
				// Use the plugin's actual name from metadata.
				pluginName := meta.Name
				if pluginName == "" {
					// Fallback: query the plugin if name is missing.
					pluginName, _, _, _ = queryPluginMetadata(meta.Path)
				}

				// Use plugin's description if available.
				desc := meta.Description
				if desc == "" {
					desc = fmt.Sprintf("External plugin (source: %s)", meta.Source)
				}

				// Convert relative paths to absolute.
				pluginPath := meta.Path
				if !filepath.IsAbs(pluginPath) {
					absPath, err := filepath.Abs(pluginPath)
					if err != nil {
						if generateVerbose {
							fmt.Fprintf(os.Stderr, " Failed to resolve absolute path for plugin '%s': %v\n", pluginName, err)
						}
						continue
					}
					pluginPath = absPath
				}

				if err := sharedPluginManager.RegisterExternalPlugin(pluginName, meta.Type, pluginPath, desc); err != nil {
					if generateVerbose {
						fmt.Fprintf(os.Stderr, " Failed to register external plugin '%s': %v\n", pluginName, err)
					}
				} else {
					// Set dry-run mode on external plugin.
					if err := setPluginDryRun(sharedPluginManager, pluginName, meta.Type, generateDryRun); err != nil {
						if generateVerbose {
							fmt.Fprintf(os.Stderr, " Failed to set dry-run for plugin '%s': %v\n", pluginName, err)
						}
					}

					// Set plugin args if provided.
					if argsJSON, ok := generatePluginArgs[pluginName]; ok {
						if err := setPluginArgs(sharedPluginManager, pluginName, meta.Type, argsJSON); err != nil {
							if generateVerbose {
								fmt.Fprintf(os.Stderr, " Failed to set args for plugin '%s': %v\n", pluginName, err)
							}
						}
					}
				}
			}
		}
	}

	// Get input plugin.
	plugin, ok := sharedPluginManager.GetInputPlugin(generateInputPlugin)
	if !ok {
		availablePlugins := make([]string, 0)
		for pluginName := range sharedPluginManager.AllInputPlugins() {
			availablePlugins = append(availablePlugins, pluginName)
		}
		return fmt.Errorf("unknown input plugin: %s (available: %s)", generateInputPlugin, strings.Join(availablePlugins, ", "))
	}

	// Input plugin specified on CLI is enabled for this execution regardless of config.
	inputPlugin := plugin

	// Validate input plugin.
	if err := inputPlugin.Validate(); err != nil {
		return fmt.Errorf("input plugin validation failed: %w", err)
	}

	// Generate palette.
	if generateVerbose {
		fmt.Fprintf(os.Stderr, " Input plugin: %s\n", inputPlugin.Name())
		fmt.Fprintf(os.Stderr, "   %s\n", inputPlugin.Description())
	}

	// Prepare options for input plugin.
	inputOpts := input.GenerateOptions{
		Verbose:         generateVerbose,
		DryRun:          generateDryRun,
		ColourOverrides: []string{}, // Could be added as a flag if needed
		PluginArgs:      make(map[string]any),
	}

	// Extract plugin-specific args if provided.
	if argsJSON, ok := generatePluginArgs[generateInputPlugin]; ok {
		var args map[string]any
		if err := json.Unmarshal([]byte(argsJSON), &args); err == nil {
			inputOpts.PluginArgs = args
			if generateVerbose {
				fmt.Fprintf(os.Stderr, "   Plugin args: %v\n", args)
			}
		} else if generateVerbose {
			fmt.Fprintf(os.Stderr, "   Failed to parse plugin args: %v\n", err)
		}
	}

	// Generate raw palette from input plugin.
	rawPalette, err := inputPlugin.Generate(ctx, inputOpts)
	if err != nil {
		return fmt.Errorf("failed to generate palette: %w", err)
	}

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "   Generated raw palette (%d colours)\n", len(rawPalette.Colors))
	}

	// Check if input plugin provides a wallpaper source.
	var wallpaperPath string
	if provider, ok := inputPlugin.(input.WallpaperProvider); ok {
		wallpaperPath = provider.WallpaperPath()
		if generateVerbose && wallpaperPath != "" {
			fmt.Fprintf(os.Stderr, "   Wallpaper source: %s\n", wallpaperPath)
		}
	}

	// Determine theme type from global flag.
	themeType := colour.ThemeAuto
	switch globalTheme {
	case "dark":
		themeType = colour.ThemeDark
	case "light":
		themeType = colour.ThemeLight
	case "auto":
		// Check if plugin provides a theme hint (optional).
		if hinter, ok := inputPlugin.(input.ThemeHinter); ok {
			hint := hinter.ThemeHint()
			if generateVerbose && hint != "" && hint != "auto" {
				fmt.Fprintf(os.Stderr, "   Plugin suggests theme: %s\n", hint)
			}
			// Plugin hints are advisory only - we let the categorizer decide.
			// based on weighted color distribution.
		}
		themeType = colour.ThemeAuto
	}

	// Categorize the palette (auto-detection uses weighted color distribution).
	config := colour.DefaultCategorisationConfig()
	config.ThemeType = themeType
	palette := colour.Categorise(rawPalette, config)

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "   Categorized palette (%d colours, %s theme)\n",
			len(palette.AllColours), palette.ThemeType.String())
		fmt.Fprintf(os.Stderr, "   Plugin execution complete.\n")
	}

	// Show preview if requested.
	if generatePreview {
		fmt.Println()
		fmt.Println(palette.StringWithPreview(true))
		fmt.Println()
	}

	// Save palette if requested.
	if generateSavePalette != "" {
		if err := savePalette(palette, generateSavePalette); err != nil {
			return fmt.Errorf("failed to save palette: %w", err)
		}
		if generateVerbose {
			fmt.Fprintf(os.Stderr, " Saved palette to: %s\n", generateSavePalette)
		}
	}

	// Determine which output plugins to run.
	var outputPlugins []output.Plugin
	if len(generateOutputs) == 1 && generateOutputs[0] == pluginTypeAll {
		// Run all enabled plugins (filtered by manager).
		for _, plugin := range sharedPluginManager.FilterOutputPlugins() {
			outputPlugins = append(outputPlugins, plugin)
		}
	} else {
		// Run specific plugins - when plugins are specified on CLI, they are enabled for this execution.
		for _, name := range generateOutputs {
			plugin, ok := sharedPluginManager.GetOutputPlugin(name)
			if !ok {
				availablePlugins := make([]string, 0)
				for pluginName := range sharedPluginManager.AllOutputPlugins() {
					availablePlugins = append(availablePlugins, pluginName)
				}
				return fmt.Errorf("unknown output plugin: %s (available: %s)", name, strings.Join(availablePlugins, ", "))
			}
			// Plugins specified on CLI are enabled for this execution regardless of config.
			outputPlugins = append(outputPlugins, plugin)
		}
	}

	if len(outputPlugins) == 0 {
		return fmt.Errorf("no output plugins selected")
	}

	// Run global pre-hook script if it exists.
	if err := runGlobalHookScript(ctx, "pre-generate", generateVerbose, generateDryRun); err != nil {
		if generateVerbose {
			fmt.Fprintf(os.Stderr, " Global pre-hook failed: %v\n", err)
		}
	}

	// Phase 1: Run all pre-execute hooks and validate plugins.
	type pluginExecution struct {
		plugin       output.Plugin
		skip         bool
		skipReason   string
		writtenFiles []string
	}
	executions := make([]pluginExecution, 0, len(outputPlugins))

	for _, plugin := range outputPlugins {
		exec := pluginExecution{plugin: plugin}

		// Set verbose mode if plugin supports it.
		if verbosePlugin, ok := plugin.(output.VerbosePlugin); ok {
			verbosePlugin.SetVerbose(generateVerbose)
		}

		// Validate plugin.
		if err := plugin.Validate(); err != nil {
			if generateVerbose {
				fmt.Fprintf(os.Stderr, " Skipping %s: %v\n", plugin.Name(), err)
			}
			exec.skip = true
			exec.skipReason = fmt.Sprintf("validation failed: %v", err)
			executions = append(executions, exec)
			continue
		}

		// Run PreExecute hook if plugin implements it.
		if preHook, ok := plugin.(output.PreExecuteHook); ok {
			hookCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			skip, reason, err := preHook.PreExecute(hookCtx)
			cancel()

			if err != nil {
				fmt.Fprintf(os.Stderr, " %s pre-execution check failed: %v\n", plugin.Name(), err)
				exec.skip = true
				exec.skipReason = fmt.Sprintf("pre-hook error: %v", err)
				executions = append(executions, exec)
				continue
			}

			if skip {
				if generateVerbose {
					fmt.Fprintf(os.Stderr, "⊘ Skipping %s: %s\n", plugin.Name(), reason)
				}
				exec.skip = true
				exec.skipReason = reason
				executions = append(executions, exec)
				continue
			}
		}

		executions = append(executions, exec)
	}

	// Phase 2: Generate and write files for non-skipped plugins.
	successCount := 0
	firstOutputPlugin := true
	for i := range executions {
		exec := &executions[i]
		if exec.skip {
			continue
		}

		plugin := exec.plugin

		if generateVerbose {
			// Add blank line before first output plugin for separation from preview.
			if firstOutputPlugin {
				fmt.Fprintf(os.Stderr, "→ Running output plugins...\n")
				firstOutputPlugin = false
			}
			fmt.Fprintf(os.Stderr, " Output plugin: %s\n", plugin.Name())
			fmt.Fprintf(os.Stderr, "   %s\n", plugin.Description())
		}

		// Create theme data with wallpaper context.
		themeData := colour.NewThemeData(palette, wallpaperPath, "")

		// Generate files.
		files, err := plugin.Generate(themeData)
		if err != nil {
			fmt.Fprintf(os.Stderr, " %s failed: %v\n", plugin.Name(), err)
			exec.skip = true
			exec.skipReason = fmt.Sprintf("generation failed: %v", err)
			continue
		}

		// Write files.
		outputDir := plugin.DefaultOutputDir()
		exec.writtenFiles = make([]string, 0, len(files))

		// Convert files map to slice for proper ordering.
		fileList := make([]struct {
			name    string
			content []byte
		}, 0, len(files))
		for filename, content := range files {
			fileList = append(fileList, struct {
				name    string
				content []byte
			}{filename, content})
		}

		for _, file := range fileList {
			filename := file.name
			content := file.content

			// Check if this is external plugin output (virtual file).
			if strings.HasSuffix(filename, "-output.txt") && outputDir == "" {
				// External plugin - display output directly.
				if len(content) > 0 {
					fmt.Println(string(content))
				}
			} else {
				// Regular plugin - write to file.
				fullPath := filepath.Join(outputDir, filename)

				if generateDryRun {
					fmt.Printf("   Would write: %s (%d bytes)\n", fullPath, len(content))
				} else {
					if err := writeFile(fullPath, content, generateVerbose); err != nil {
						fmt.Fprintf(os.Stderr, " Failed to write %s: %v\n", fullPath, err)
						exec.skip = true
						exec.skipReason = fmt.Sprintf("write failed: %v", err)
						continue
					}
					fmt.Printf("   %s (%d bytes)\n", fullPath, len(content))
					exec.writtenFiles = append(exec.writtenFiles, fullPath)
				}
			}
		}

		if !exec.skip {
			successCount++
			if generateVerbose {
				fmt.Fprintf(os.Stderr, "   Plugin execution complete.\n")
			}
		}
	}

	// Phase 3: Run all post-execute hooks for successful plugins.
	if !generateDryRun {
		// Run post-hooks for each plugin.
		for _, exec := range executions {
			if exec.skip || len(exec.writtenFiles) == 0 {
				continue
			}

			plugin := exec.plugin
			if postHook, ok := plugin.(output.PostExecuteHook); ok {
				// Build execution context for the hook.
				execContext := output.ExecutionContext{
					DryRun:        generateDryRun,
					Verbose:       generateVerbose,
					OutputDir:     plugin.DefaultOutputDir(),
					WallpaperPath: wallpaperPath,
				}

				if generateVerbose {
					fmt.Fprintf(os.Stderr, "→ Running %s post-hook...\n", plugin.Name())
				}

				hookCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				err := postHook.PostExecute(hookCtx, execContext, exec.writtenFiles)
				cancel()

				if err != nil {
					fmt.Fprintf(os.Stderr, "   %s post-hook failed: %v\n", plugin.Name(), err)
				}
			}
		}

		// Run global post-hook script if it exists.
		if err := runGlobalHookScript(ctx, "post-generate", generateVerbose, generateDryRun); err != nil {
			if generateVerbose {
				fmt.Fprintf(os.Stderr, " Global post-hook failed: %v\n", err)
			}
		}
	}

	// Summary.
	if !generateDryRun {
		fmt.Println()
		if successCount > 0 {
			fmt.Printf(" Done! Generated %d output plugin(s)\n", successCount)
		} else {
			return fmt.Errorf("no output plugins succeeded")
		}
	}

	return nil
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
