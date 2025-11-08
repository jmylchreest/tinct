// Package cli provides the command-line interface for Tinct.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/output"
)

// loadAndConfigurePlugins loads the plugin lock file and configures plugins.
func loadAndConfigurePlugins() error {
	if err := loadAndApplyPluginLock(); err != nil {
		return nil // No lock file is OK
	}

	lock, _, err := loadPluginLock()
	if err != nil || lock == nil {
		return nil
	}

	// Register external plugins with absolute path resolution.
	registerExternalPluginsFromLock(lock, true, generateVerbose)

	// Configure external plugins with dry-run and args.
	if lock.ExternalPlugins != nil {
		for _, meta := range lock.ExternalPlugins {
			pluginName := meta.Name
			if pluginName == "" {
				pluginName, _, _, _ = queryPluginMetadata(meta.Path)
			}
			if pluginName != "" {
				configureExternalPlugin(pluginName, meta.Type, generateDryRun, generatePluginArgs, generateVerbose)
			}
		}
	}

	return nil
}

// getAndValidateInputPlugin retrieves and validates the input plugin.
func getAndValidateInputPlugin() (input.Plugin, error) {
	plugin, ok := sharedPluginManager.GetInputPlugin(generateInputPlugin)
	if !ok {
		availablePlugins := make([]string, 0)
		for pluginName := range sharedPluginManager.AllInputPlugins() {
			availablePlugins = append(availablePlugins, pluginName)
		}
		return nil, fmt.Errorf("unknown input plugin: %s (available: %s)", generateInputPlugin, strings.Join(availablePlugins, ", "))
	}

	if err := plugin.Validate(); err != nil {
		return nil, fmt.Errorf("input plugin validation failed: %w", err)
	}

	return plugin, nil
}

// generateInputPalette generates a raw palette from the input plugin.
func generateInputPalette(ctx context.Context, inputPlugin input.Plugin) (*colour.Palette, string, error) {
	if generateVerbose {
		fmt.Fprintf(os.Stderr, " Input plugin: %s\n", inputPlugin.Name())
		fmt.Fprintf(os.Stderr, "   %s\n", inputPlugin.Description())
	}

	// Prepare options for input plugin.
	inputOpts := buildInputOptions()

	// Generate raw palette from input plugin.
	rawPalette, err := inputPlugin.Generate(ctx, inputOpts)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate palette: %w", err)
	}

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "   Generated raw palette (%d colours)\n", len(rawPalette.Colors))
	}

	// Extract wallpaper path if available.
	wallpaperPath := extractWallpaperPath(inputPlugin)
	if generateVerbose && wallpaperPath != "" {
		fmt.Fprintf(os.Stderr, "   Wallpaper source: %s\n", wallpaperPath)
	}

	return rawPalette, wallpaperPath, nil
}

// buildInputOptions creates input plugin options.
func buildInputOptions() input.GenerateOptions {
	inputOpts := input.GenerateOptions{
		Verbose:         generateVerbose,
		DryRun:          generateDryRun,
		ColourOverrides: []string{},
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

	return inputOpts
}

// extractWallpaperPath extracts wallpaper path from input plugin if it provides one.
func extractWallpaperPath(inputPlugin input.Plugin) string {
	if provider, ok := inputPlugin.(input.WallpaperProvider); ok {
		return provider.WallpaperPath()
	}
	return ""
}

// categorizePalette categorizes a raw palette based on theme settings.
func categorizePalette(rawPalette *colour.Palette, inputPlugin input.Plugin) *colour.CategorisedPalette {
	themeType := determineThemeType(inputPlugin)

	config := colour.DefaultCategorisationConfig()
	config.ThemeType = themeType
	palette := colour.Categorise(rawPalette, config)

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "   Categorized palette (%d colours, %s theme)\n",
			len(palette.AllColours), palette.ThemeType.String())
		fmt.Fprintf(os.Stderr, "   Plugin execution complete.\n")
	}

	return palette
}

// determineThemeType determines the theme type from global flag and plugin hints.
func determineThemeType(inputPlugin input.Plugin) colour.ThemeType {
	switch globalTheme {
	case "dark":
		return colour.ThemeDark
	case "light":
		return colour.ThemeLight
	case "auto":
		// Check if plugin provides a theme hint (optional).
		if hinter, ok := inputPlugin.(input.ThemeHinter); ok {
			hint := hinter.ThemeHint()
			if generateVerbose && hint != "" && hint != "auto" {
				fmt.Fprintf(os.Stderr, "   Plugin suggests theme: %s\n", hint)
			}
		}
		return colour.ThemeAuto
	default:
		return colour.ThemeAuto
	}
}

// handlePaletteOutput shows preview and saves palette if requested.
func handlePaletteOutput(palette *colour.CategorisedPalette) error {
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

	return nil
}

// selectOutputPlugins determines which output plugins to run.
func selectOutputPlugins() ([]output.Plugin, error) {
	if len(generateOutputs) == 1 && generateOutputs[0] == pluginTypeAll {
		// Run all non-disabled plugins (on-demand plugins are included).
		allPlugins := sharedPluginManager.AllOutputPlugins()
		plugins := make([]output.Plugin, 0, len(allPlugins))
		for _, plugin := range allPlugins {
			// Skip explicitly disabled plugins.
			if sharedPluginManager.IsOutputDisabled(plugin) {
				if generateVerbose {
					fmt.Fprintf(os.Stderr, "⊘ Skipping %s: plugin is disabled\n", plugin.Name())
				}
				continue
			}
			plugins = append(plugins, plugin)
		}
		if len(plugins) == 0 {
			return nil, fmt.Errorf("no output plugins available (all plugins are disabled)")
		}
		return plugins, nil
	}

	// Run specific plugins - check that they're not disabled.
	plugins := make([]output.Plugin, 0, len(generateOutputs))
	for _, name := range generateOutputs {
		plugin, ok := sharedPluginManager.GetOutputPlugin(name)
		if !ok {
			availablePlugins := make([]string, 0)
			for pluginName := range sharedPluginManager.AllOutputPlugins() {
				availablePlugins = append(availablePlugins, pluginName)
			}
			return nil, fmt.Errorf("unknown output plugin: %s (available: %s)", name, strings.Join(availablePlugins, ", "))
		}
		// Check if plugin is explicitly disabled.
		if sharedPluginManager.IsOutputDisabled(plugin) {
			return nil, fmt.Errorf("plugin %s is disabled (check TINCT_DISABLED_PLUGINS)", name)
		}
		plugins = append(plugins, plugin)
	}

	if len(plugins) == 0 {
		return nil, fmt.Errorf("no output plugins selected")
	}

	return plugins, nil
}

// pluginExecution tracks the execution state of an output plugin.
type pluginExecution struct {
	plugin       output.Plugin
	skip         bool
	skipReason   string
	writtenFiles []string
}

// preparePluginExecutions validates plugins and runs pre-execute hooks.
func preparePluginExecutions(ctx context.Context, plugins []output.Plugin) []pluginExecution {
	executions := make([]pluginExecution, 0, len(plugins))

	for _, plugin := range plugins {
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
		if shouldSkipFromPreHook(ctx, plugin, &exec) {
			executions = append(executions, exec)
			continue
		}

		executions = append(executions, exec)
	}

	return executions
}

// shouldSkipFromPreHook runs the pre-execute hook and determines if plugin should be skipped.
func shouldSkipFromPreHook(ctx context.Context, plugin output.Plugin, exec *pluginExecution) bool {
	preHook, ok := plugin.(output.PreExecuteHook)
	if !ok {
		return false
	}

	hookCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	skip, reason, err := preHook.PreExecute(hookCtx)
	cancel()

	if err != nil {
		fmt.Fprintf(os.Stderr, " %s pre-execution check failed: %v\n", plugin.Name(), err)
		exec.skip = true
		exec.skipReason = fmt.Sprintf("pre-hook error: %v", err)
		return true
	}

	if skip {
		if generateVerbose {
			fmt.Fprintf(os.Stderr, "⊘ Skipping %s: %s\n", plugin.Name(), reason)
		}
		exec.skip = true
		exec.skipReason = reason
		return true
	}

	return false
}

// generateAndWriteFiles generates files from plugins and writes them to disk.
func generateAndWriteFiles(executions []pluginExecution, palette *colour.CategorisedPalette, wallpaperPath string) int {
	successCount := 0
	firstOutputPlugin := true

	for i := range executions {
		exec := &executions[i]
		if exec.skip {
			continue
		}

		if generateVerbose && firstOutputPlugin {
			fmt.Fprintf(os.Stderr, "→ Running output plugins...\n")
			firstOutputPlugin = false
		}

		if processPluginGeneration(exec, palette, wallpaperPath) {
			successCount++
		}
	}

	return successCount
}

// processPluginGeneration generates and writes files for a single plugin.
func processPluginGeneration(exec *pluginExecution, palette *colour.CategorisedPalette, wallpaperPath string) bool {
	plugin := exec.plugin

	if generateVerbose {
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
		return false
	}

	// Write files.
	if !writePluginFiles(exec, plugin, files) {
		return false
	}

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "   Plugin execution complete.\n")
	}

	return true
}

// writePluginFiles writes generated files to disk.
func writePluginFiles(exec *pluginExecution, plugin output.Plugin, files map[string][]byte) bool {
	outputDir := plugin.DefaultOutputDir()
	exec.writtenFiles = make([]string, 0, len(files))

	for filename, content := range files {
		// Check if this is external plugin output (virtual file).
		if strings.HasSuffix(filename, "-output.txt") && outputDir == "" {
			// External plugin - display output directly.
			if len(content) > 0 {
				fmt.Println(string(content))
			}
			continue
		}

		// Regular plugin - write to file.
		fullPath := filepath.Join(outputDir, filename)

		if generateDryRun {
			fmt.Printf("   Would write: %s (%d bytes)\n", fullPath, len(content))
		} else {
			if err := writeFile(fullPath, content, generateVerbose); err != nil {
				fmt.Fprintf(os.Stderr, " Failed to write %s: %v\n", fullPath, err)
				exec.skip = true
				exec.skipReason = fmt.Sprintf("write failed: %v", err)
				return false
			}
			fmt.Printf("   %s (%d bytes)\n", fullPath, len(content))
			exec.writtenFiles = append(exec.writtenFiles, fullPath)
		}
	}

	return true
}

// runPostExecutionHooks runs post-execute hooks for successful plugins.
func runPostExecutionHooks(ctx context.Context, executions []pluginExecution, wallpaperPath string) {
	for _, exec := range executions {
		if exec.skip || len(exec.writtenFiles) == 0 {
			continue
		}

		plugin := exec.plugin
		postHook, ok := plugin.(output.PostExecuteHook)
		if !ok {
			continue
		}

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

// printGenerationSummary prints the final summary of the generation process.
func printGenerationSummary(successCount int) error {
	if generateDryRun {
		return nil
	}

	fmt.Println()
	if successCount > 0 {
		fmt.Printf(" Done! Generated %d output plugin(s)\n", successCount)
		return nil
	}

	return fmt.Errorf("no output plugins succeeded")
}
