// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/hyprland"
	"github.com/jmylchreest/tinct/internal/plugin/output/kitty"
	"github.com/jmylchreest/tinct/internal/plugin/output/template"
	"github.com/spf13/cobra"
)

var (
	templateOutputPlugins []string
	templateForce         bool
	templateVerbose       bool
)

// pluginTemplatesCmd represents the plugins templates command
var pluginTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage output plugin templates",
	Long: `Manage output plugin templates including listing and dumping embedded templates.

Templates can be customized by extracting them to ~/.config/tinct/templates/{plugin-name}/
and modifying them. Custom templates will be used instead of embedded ones.

Examples:
  tinct plugins templates list
  tinct plugins templates dump -o hyprland,kitty
  tinct plugins templates dump -o hyprland --force`,
}

// pluginTemplatesListCmd lists available templates
var pluginTemplatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available plugin templates",
	Long: `List all available templates from output plugins.

Shows which templates are embedded and which have custom overrides.`,
	RunE: runPluginTemplatesList,
}

// pluginTemplatesDumpCmd dumps embedded templates to files
var pluginTemplatesDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump embedded templates to files",
	Long: `Extract embedded plugin templates to ~/.config/tinct/templates/{plugin-name}/

This allows you to customize the templates used for generating output files.
By default, all available plugins will have their templates dumped.

Use -o/--output-plugins to specify which plugins to dump.

Examples:
  tinct plugins templates dump
  tinct plugins templates dump -o hyprland,kitty
  tinct plugins templates dump -o hyprland --force`,
	RunE: runPluginTemplatesDump,
}

func init() {
	// Add template subcommands
	pluginTemplatesCmd.AddCommand(pluginTemplatesListCmd)
	pluginTemplatesCmd.AddCommand(pluginTemplatesDumpCmd)

	// Add templates command to plugins
	pluginsCmd.AddCommand(pluginTemplatesCmd)

	// Dump flags
	pluginTemplatesDumpCmd.Flags().StringSliceVarP(&templateOutputPlugins, "output-plugins", "o", []string{}, "comma-separated list of output plugins (default: all)")
	pluginTemplatesDumpCmd.Flags().BoolVarP(&templateForce, "force", "f", false, "overwrite existing custom templates")
	pluginTemplatesDumpCmd.Flags().BoolVarP(&templateVerbose, "verbose", "v", false, "show verbose output")

	// List flags
	pluginTemplatesListCmd.Flags().StringSliceVarP(&templateOutputPlugins, "output-plugins", "o", []string{}, "comma-separated list of output plugins to list (default: all)")
}

func runPluginTemplatesList(cmd *cobra.Command, args []string) error {
	// Get available output plugins
	plugins := getAvailableOutputPlugins()
	if len(plugins) == 0 {
		fmt.Println("No output plugins available")
		return nil
	}

	// Filter plugins if specified
	if len(templateOutputPlugins) > 0 {
		filtered := make(map[string]output.Plugin)
		for _, name := range templateOutputPlugins {
			if plugin, ok := plugins[name]; ok {
				filtered[name] = plugin
			} else {
				fmt.Fprintf(os.Stderr, "Warning: plugin %q not found\n", name)
			}
		}
		plugins = filtered
	}

	if len(plugins) == 0 {
		fmt.Println("No matching plugins found")
		return nil
	}

	fmt.Println("Available plugin templates:")
	fmt.Println()

	hasCustomTemplates := false
	for pluginName, plugin := range plugins {
		loader, _ := getPluginTemplateLoader(pluginName, plugin)
		if loader == nil {
			fmt.Printf("Plugin: %s (no templates)\n", pluginName)
			fmt.Println()
			continue
		}

		templates, err := loader.ListEmbeddedTemplates()
		if err != nil {
			return fmt.Errorf("failed to list templates for %s: %w", pluginName, err)
		}

		fmt.Printf("Plugin: %s\n", pluginName)
		fmt.Printf("  Custom template directory: %s\n", loader.CustomDir())
		fmt.Printf("  Templates:\n")

		for _, tmpl := range templates {
			info := loader.GetInfo(tmpl)
			if info.CustomExists {
				fmt.Printf("    - %s*\n", tmpl)
				hasCustomTemplates = true
			} else {
				fmt.Printf("    - %s\n", tmpl)
			}
		}
		fmt.Println()
	}

	fmt.Println("To customize a template, use: tinct plugins templates dump -o <plugin-name>")
	if hasCustomTemplates {
		fmt.Println("Templates with active overrides are shown with an asterisk (*).")
	}

	return nil
}

func runPluginTemplatesDump(cmd *cobra.Command, args []string) error {
	// Get available output plugins
	plugins := getAvailableOutputPlugins()
	if len(plugins) == 0 {
		return fmt.Errorf("no output plugins available")
	}

	// Filter plugins if specified
	if len(templateOutputPlugins) > 0 {
		filtered := make(map[string]output.Plugin)
		for _, name := range templateOutputPlugins {
			if plugin, ok := plugins[name]; ok {
				filtered[name] = plugin
			} else {
				return fmt.Errorf("plugin %q not found", name)
			}
		}
		plugins = filtered
	}

	if len(plugins) == 0 {
		return fmt.Errorf("no matching plugins found")
	}

	totalDumped := 0
	for pluginName, plugin := range plugins {
		loader, _ := getPluginTemplateLoader(pluginName, plugin)
		if loader == nil {
			if templateVerbose {
				fmt.Printf("Skipping %s: no templates\n", pluginName)
			}
			continue
		}

		templates, err := loader.ListEmbeddedTemplates()
		if err != nil {
			return fmt.Errorf("failed to list templates for %s: %w", pluginName, err)
		}

		if len(templates) == 0 {
			if templateVerbose {
				fmt.Printf("Skipping %s: no templates\n", pluginName)
			}
			continue
		}

		fmt.Printf("Dumping templates for %s...\n", pluginName)

		dumped, err := loader.DumpAllTemplates(templateForce)

		// Show what was successfully dumped
		for _, path := range dumped {
			fmt.Printf("  ✓ %s\n", path)
			totalDumped++
		}

		// Handle errors (likely "already exists" messages)
		if err != nil {
			// Check if it's an "already exists" error
			if !templateForce && strings.Contains(err.Error(), "already exists") {
				// Split multiple errors and show each skipped file
				errorParts := strings.Split(err.Error(), "; ")
				for _, errPart := range errorParts {
					if strings.Contains(errPart, "already exists") {
						// Extract just the filename from the path
						if idx := strings.Index(errPart, "already exists: "); idx != -1 {
							path := strings.TrimPrefix(errPart[idx:], "already exists: ")
							path = strings.TrimSuffix(path, " (use --force to overwrite)")
							fmt.Printf("  ⊘ %s (already exists)\n", path)
						}
					}
				}
				if len(dumped) == 0 {
					fmt.Fprintf(os.Stderr, "  Use --force to overwrite existing templates\n")
				}
			} else {
				// Other errors should fail
				return fmt.Errorf("failed to dump templates for %s: %w", pluginName, err)
			}
		}
	}

	if totalDumped == 0 {
		if templateForce {
			fmt.Println("No templates were dumped.")
		} else {
			fmt.Println("No templates were dumped. Custom templates may already exist.")
			fmt.Println("Use --force to overwrite existing templates.")
		}
		return nil
	}

	fmt.Println()
	fmt.Printf("Successfully dumped %d template(s)\n", totalDumped)
	fmt.Println()
	fmt.Println("You can now customize these templates. They will be used instead of the embedded versions.")

	return nil
}

// getAvailableOutputPlugins returns a map of all available output plugins
func getAvailableOutputPlugins() map[string]output.Plugin {
	plugins := make(map[string]output.Plugin)

	// Register built-in output plugins
	hyprlandPlugin := hyprland.New()
	kittyPlugin := kitty.New()

	plugins[hyprlandPlugin.Name()] = hyprlandPlugin
	plugins[kittyPlugin.Name()] = kittyPlugin

	// TODO: Add support for external plugins when that system is implemented

	return plugins
}

// getPluginTemplateLoader returns a template loader for the given plugin
// Returns nil if the plugin doesn't support templates
func getPluginTemplateLoader(pluginName string, plugin output.Plugin) (*template.Loader, interface{}) {
	// We need to access the embedded FS from each plugin
	// This is a bit of a workaround since the FS is not exposed in the Plugin interface

	switch pluginName {
	case "hyprland":
		// Access hyprland's embedded templates
		if hp, ok := plugin.(*hyprland.Plugin); ok {
			loader := template.New(pluginName, hyprland.GetEmbeddedTemplates())
			return loader, hp
		}
	case "kitty":
		// Access kitty's embedded templates
		if kp, ok := plugin.(*kitty.Plugin); ok {
			loader := template.New(pluginName, kitty.GetEmbeddedTemplates())
			return loader, kp
		}
	}

	return nil, nil
}
