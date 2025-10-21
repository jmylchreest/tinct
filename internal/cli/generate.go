// Package cli provides the command-line interface for Tinct.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/spf13/cobra"
)

var (
	// Global plugin manager
	pluginManager *manager.Manager

	// Generate command flags
	generateInputPlugin string
	generateOutputs     []string
	generateDryRun      bool
	generatePreview     bool
	generateSavePalette string
	generateVerbose     bool
	generatePluginArgs  map[string]string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate configuration files from a colour palette",
	Long: `Generate configuration files for various applications from a colour palette.

The palette can be created from an image, loaded from a file, or built from
command-line colour specifications. Generated files are written to their
default locations or a custom output directory.

Input Plugins:
  image  - Extract colours from an image file
  file   - Load palette from file or build from colour specifications

Output Plugins:
  (dynamically listed based on enabled plugins)

Examples:
  # From image - generate all outputs
  tinct generate --input image -p wallpaper.jpg

  # From palette file - specific output
  tinct generate --input file -p theme.json --outputs hyprland

  # Pure manual - build from colour specs
  tinct generate --input file \
    --colour background=#1e1e2e \
    --colour foreground=#cdd6f4 \
    --colour accent1=#f38ba8

  # File with overrides
  tinct generate --input file -p base.json \
    --colour accent1=#ff0000

  # With custom output options
  tinct generate --input image -p wall.jpg \
    --outputs hyprland \
    --hyprland.output-dir ~/.config/hypr/themes

  # Preview before generating
  tinct generate --input image -p wall.jpg --preview --dry-run

  # With theme preference
  tinct generate --input image -p wall.jpg --theme dark

  # Pass args to specific plugins (JSON format)
  tinct generate --input image -p wall.jpg \
    --plugin-args myinput='{"seed":42,"count":20}' \
    --plugin-args myplugin='{"format":"extended"}'

Note: All role names use camelCase (e.g., backgroundMuted, accent1)`,
	RunE: runGenerate,
}

func init() {
	// Initialize plugin manager from environment (will be overridden by lock file at runtime)
	pluginManager = manager.NewFromEnv()

	// Input plugin selection (required)
	generateCmd.Flags().StringVarP(&generateInputPlugin, "input", "i", "", "Input plugin (required: image, file)")
	generateCmd.MarkFlagRequired("input")

	// Output plugin selection
	generateCmd.Flags().StringSliceVarP(&generateOutputs, "outputs", "o", []string{"all"}, "Output plugins (comma-separated or 'all')")

	// General options
	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false, "Preview without writing files")
	generateCmd.Flags().BoolVar(&generatePreview, "preview", false, "Show colour palette preview")
	generateCmd.Flags().StringVar(&generateSavePalette, "save-palette", "", "Save palette to file (JSON)")
	generateCmd.Flags().BoolVarP(&generateVerbose, "verbose", "v", false, "Verbose output")
	generateCmd.Flags().StringToStringVar(&generatePluginArgs, "plugin-args", nil, "Plugin-specific arguments (key=value format, repeatable for multiple plugins)")

	// Register plugin flags
	for _, plugin := range pluginManager.AllInputPlugins() {
		plugin.RegisterFlags(generateCmd)
	}
	for _, plugin := range pluginManager.AllOutputPlugins() {
		plugin.RegisterFlags(generateCmd)
	}

	// Update help text with enabled plugins
	generateCmd.Long = buildGenerateHelp()
}

// runGenerate executes the generate command.
func runGenerate(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Reload plugin manager config from lock file if available (overrides env)
	// Don't recreate the manager to preserve flag bindings
	lock, _, err := loadPluginLock()
	if err == nil && lock != nil {
		config := manager.Config{
			EnabledPlugins:  lock.EnabledPlugins,
			DisabledPlugins: lock.DisabledPlugins,
		}
		pluginManager.UpdateConfig(config)

		// Register external plugins from lock file
		if lock.ExternalPlugins != nil {
			for _, meta := range lock.ExternalPlugins {
				// Use the plugin's actual name from metadata
				pluginName := meta.Name
				if pluginName == "" {
					// Fallback: query the plugin if name is missing
					pluginName, _, _, _ = queryPluginMetadata(meta.Path)
				}

				// Use plugin's description if available
				desc := meta.Description
				if desc == "" {
					desc = fmt.Sprintf("External plugin (source: %s)", meta.Source)
				}

				if err := pluginManager.RegisterExternalPlugin(pluginName, meta.Type, meta.Path, desc); err != nil {
					if generateVerbose {
						fmt.Fprintf(os.Stderr, "⚠ Failed to register external plugin '%s': %v\n", pluginName, err)
					}
				} else {
					// Set dry-run mode on external plugin
					if err := setPluginDryRun(pluginManager, pluginName, meta.Type, generateDryRun); err != nil {
						if generateVerbose {
							fmt.Fprintf(os.Stderr, "⚠ Failed to set dry-run for plugin '%s': %v\n", pluginName, err)
						}
					}

					// Set plugin args if provided
					if argsJSON, ok := generatePluginArgs[pluginName]; ok {
						if err := setPluginArgs(pluginManager, pluginName, meta.Type, argsJSON); err != nil {
							if generateVerbose {
								fmt.Fprintf(os.Stderr, "⚠ Failed to set args for plugin '%s': %v\n", pluginName, err)
							}
						}
					}
				}
			}
		}
	}

	// Get input plugin
	plugin, ok := pluginManager.GetInputPlugin(generateInputPlugin)
	if !ok {
		availablePlugins := make([]string, 0)
		for pluginName := range pluginManager.AllInputPlugins() {
			availablePlugins = append(availablePlugins, pluginName)
		}
		return fmt.Errorf("unknown input plugin: %s (available: %s)", generateInputPlugin, strings.Join(availablePlugins, ", "))
	}

	// Input plugin specified on CLI is enabled for this execution regardless of config
	inputPlugin := plugin

	// Validate input plugin
	if err := inputPlugin.Validate(); err != nil {
		return fmt.Errorf("input plugin validation failed: %w", err)
	}

	// Generate palette
	if generateVerbose {
		fmt.Fprintf(os.Stderr, "✓ Input plugin: %s\n", inputPlugin.Name())
		fmt.Fprintf(os.Stderr, "  └─ %s\n", inputPlugin.Description())
	}

	// Prepare options for input plugin
	inputOpts := input.GenerateOptions{
		Verbose:         generateVerbose,
		DryRun:          generateDryRun,
		ColourOverrides: []string{}, // Could be added as a flag if needed
		PluginArgs:      make(map[string]any),
	}

	// Extract plugin-specific args if provided
	if argsJSON, ok := generatePluginArgs[generateInputPlugin]; ok {
		var args map[string]any
		if err := json.Unmarshal([]byte(argsJSON), &args); err == nil {
			inputOpts.PluginArgs = args
			if generateVerbose {
				fmt.Fprintf(os.Stderr, "  └─ Plugin args: %v\n", args)
			}
		} else if generateVerbose {
			fmt.Fprintf(os.Stderr, "  └─ Failed to parse plugin args: %v\n", err)
		}
	}

	// Generate raw palette from input plugin
	rawPalette, err := inputPlugin.Generate(ctx, inputOpts)
	if err != nil {
		return fmt.Errorf("failed to generate palette: %w", err)
	}

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "  └─ Generated raw palette (%d colours)\n", len(rawPalette.Colors))
	}

	// Determine theme type from global flag
	themeType := colour.ThemeAuto
	switch globalTheme {
	case "dark":
		themeType = colour.ThemeDark
	case "light":
		themeType = colour.ThemeLight
	case "auto":
		// Check if plugin provides a theme hint (optional)
		if hinter, ok := inputPlugin.(input.ThemeHinter); ok {
			hint := hinter.ThemeHint()
			if generateVerbose && hint != "" && hint != "auto" {
				fmt.Fprintf(os.Stderr, "  └─ Plugin suggests theme: %s\n", hint)
			}
			// Plugin hints are advisory only - we let the categorizer decide
			// based on weighted color distribution
		}
		themeType = colour.ThemeAuto
	}

	// Categorize the palette (auto-detection uses weighted color distribution)
	config := colour.DefaultCategorisationConfig()
	config.ThemeType = themeType
	palette := colour.Categorise(rawPalette, config)

	if generateVerbose {
		fmt.Fprintf(os.Stderr, "  └─ Categorized palette (%d colours, %s theme)\n",
			len(palette.AllColours), palette.ThemeType.String())
	}

	// Show preview if requested
	if generatePreview {
		fmt.Println()
		fmt.Println(palette.StringWithPreview(true))
		fmt.Println()
	}

	// Save palette if requested
	if generateSavePalette != "" {
		if err := savePalette(palette, generateSavePalette); err != nil {
			return fmt.Errorf("failed to save palette: %w", err)
		}
		if generateVerbose {
			fmt.Fprintf(os.Stderr, "✓ Saved palette to: %s\n", generateSavePalette)
		}
	}

	// Determine which output plugins to run
	var outputPlugins []output.Plugin
	if len(generateOutputs) == 1 && generateOutputs[0] == "all" {
		// Run all enabled plugins (filtered by manager)
		for _, plugin := range pluginManager.FilterOutputPlugins() {
			outputPlugins = append(outputPlugins, plugin)
		}
	} else {
		// Run specific plugins - when plugins are specified on CLI, they are enabled for this execution
		for _, name := range generateOutputs {
			plugin, ok := pluginManager.GetOutputPlugin(name)
			if !ok {
				availablePlugins := make([]string, 0)
				for pluginName := range pluginManager.AllOutputPlugins() {
					availablePlugins = append(availablePlugins, pluginName)
				}
				return fmt.Errorf("unknown output plugin: %s (available: %s)", name, strings.Join(availablePlugins, ", "))
			}
			// Plugins specified on CLI are enabled for this execution regardless of config
			outputPlugins = append(outputPlugins, plugin)
		}
	}

	if len(outputPlugins) == 0 {
		return fmt.Errorf("no output plugins selected")
	}

	// Execute output plugins
	successCount := 0
	for _, plugin := range outputPlugins {
		if err := plugin.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Skipping %s: %v\n", plugin.Name(), err)
			continue
		}

		if generateVerbose {
			fmt.Fprintf(os.Stderr, "\n✓ Output plugin: %s\n", plugin.Name())
			fmt.Fprintf(os.Stderr, "  └─ %s\n", plugin.Description())
		}

		// Generate files
		files, err := plugin.Generate(palette)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s failed: %v\n", plugin.Name(), err)
			continue
		}

		// Write files
		outputDir := plugin.DefaultOutputDir()
		for filename, content := range files {
			// Check if this is external plugin output (virtual file)
			if strings.HasSuffix(filename, "-output.txt") && outputDir == "" {
				// External plugin - display output directly
				if len(content) > 0 {
					fmt.Println(string(content))
				}
			} else {
				// Regular plugin - write to file
				fullPath := filepath.Join(outputDir, filename)

				if generateDryRun {
					fmt.Printf("  Would write: %s (%d bytes)\n", fullPath, len(content))
				} else {
					if err := writeFile(fullPath, content); err != nil {
						return fmt.Errorf("failed to write %s: %w", fullPath, err)
					}
					fmt.Printf("  ├─ %s (%d bytes)\n", fullPath, len(content))
				}
			}
		}

		successCount++
	}

	// Summary
	if !generateDryRun {
		fmt.Println()
		if successCount > 0 {
			fmt.Printf("✓ Done! Generated %d output plugin(s)\n", successCount)
		} else {
			return fmt.Errorf("no output plugins succeeded")
		}
	}

	return nil
}

// buildGenerateHelp dynamically builds the help text with enabled plugins.
func buildGenerateHelp() string {
	help := `Generate configuration files for various applications from a colour palette.

The palette can be created from an image, loaded from a file, or built from
command-line colour specifications. Generated files are written to their
default locations or a custom output directory.

Input Plugins:
  image  - Extract colours from an image file
  file   - Load palette from file or build from colour specifications

Output Plugins:
`

	// List enabled plugins (filtered by manager)
	enabledPlugins := []string{}
	for _, plugin := range pluginManager.FilterOutputPlugins() {
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

	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// setPluginArgs sets custom arguments for a plugin
func setPluginArgs(mgr *manager.Manager, pluginName, pluginType, argsJSON string) error {
	// Parse JSON args
	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Get the plugin and set args based on type
	if pluginType == "output" {
		plugin, ok := mgr.GetOutputPlugin(pluginName)
		if !ok {
			return fmt.Errorf("plugin not found")
		}

		// Check if it's an external plugin that supports args
		if extPlugin, ok := plugin.(*manager.ExternalOutputPlugin); ok {
			extPlugin.SetArgs(args)
		}
	}
	// Could add input plugin support here in the future

	return nil
}

// setPluginDryRun sets dry-run mode for a plugin
func setPluginDryRun(mgr *manager.Manager, pluginName, pluginType string, dryRun bool) error {
	// Get the plugin and set dry-run based on type
	if pluginType == "output" {
		plugin, ok := mgr.GetOutputPlugin(pluginName)
		if !ok {
			return fmt.Errorf("plugin not found")
		}

		// Check if it's an external plugin that supports dry-run
		if extPlugin, ok := plugin.(*manager.ExternalOutputPlugin); ok {
			extPlugin.SetDryRun(dryRun)
		}
	}
	// Could add input plugin support here in the future

	return nil
}

// writeFile writes content to a file, creating directories as needed.
func writeFile(path string, content []byte) error {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists and create backup
	if _, err := os.Stat(path); err == nil {
		backupPath := path + ".backup"
		if err := os.Rename(path, backupPath); err != nil {
			// If backup fails, continue anyway
			fmt.Fprintf(os.Stderr, "  ⚠ Could not create backup: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "  ℹ Created backup: %s\n", backupPath)
		}
	}

	// Write the file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
