// Package cli provides the command-line interface for Tinct.
package cli

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/template"
)

var (
	templateOutputPlugins []string
	templateForce         bool
	templateVerbose       bool
	templateLocation      string
)

// pluginTemplatesCmd represents the plugins templates command.
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

// pluginTemplatesListCmd lists available templates.
var pluginTemplatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available plugin templates",
	Long: `List all available templates from output plugins.

Shows which templates are embedded and which have custom overrides.`,
	RunE: runPluginTemplatesList,
}

// pluginTemplatesDumpCmd dumps embedded templates to files.
var pluginTemplatesDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump embedded templates to files",
	Long: `Extract embedded plugin templates to ~/.config/tinct/templates/{plugin-name}/

This allows you to customize the templates used for generating output files.
By default, all available plugins will have their templates dumped.

Use -o/--output-plugins to specify which plugins to dump.
Use -l/--location to specify a custom output directory.

Examples:
  tinct plugins templates dump
  tinct plugins templates dump -o hyprland,kitty
  tinct plugins templates dump -o hyprland --force
  tinct plugins templates dump -l ./templates
  tinct plugins templates dump -l ~/my-themes/templates -o hyprland,kitty`,
	RunE: runPluginTemplatesDump,
}

func init() {
	// Add template subcommands.
	pluginTemplatesCmd.AddCommand(pluginTemplatesListCmd)
	pluginTemplatesCmd.AddCommand(pluginTemplatesDumpCmd)

	// Add templates command to plugins.
	pluginsCmd.AddCommand(pluginTemplatesCmd)

	// Dump flags.
	pluginTemplatesDumpCmd.Flags().StringSliceVarP(&templateOutputPlugins, "output-plugins", "o", []string{}, "comma-separated list of output plugins (default: all)")
	pluginTemplatesDumpCmd.Flags().BoolVarP(&templateForce, "force", "f", false, "overwrite existing custom templates")
	pluginTemplatesDumpCmd.Flags().BoolVarP(&templateVerbose, "verbose", "v", false, "show verbose output")
	pluginTemplatesDumpCmd.Flags().StringVarP(&templateLocation, "location", "l", "", "custom location to dump templates (default: ~/.config/tinct/templates)")

	// List flags.
	pluginTemplatesListCmd.Flags().StringSliceVarP(&templateOutputPlugins, "output-plugins", "o", []string{}, "comma-separated list of output plugins to list (default: all)")
}

func runPluginTemplatesList(_ *cobra.Command, _ []string) (err error) {
	items := 0
	defer func() { sendPluginSubcommandResult("templates", "list", items, err) }()

	// Auto-register external plugins from the manifest so this command
	// also reports templates from contrib plugins (paletty, awob, …),
	// not just the built-in set.
	_ = loadAndApplyPluginManifest()

	plugins := sharedPluginManager.AllOutputPlugins()
	if len(plugins) == 0 {
		fmt.Println("No output plugins available")
		return nil
	}
	items = len(plugins)

	// Filter plugins if specified.
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
		loader := getPluginTemplateLoader(pluginName, plugin)
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

func runPluginTemplatesDump(_ *cobra.Command, _ []string) (err error) { //nolint:gocyclo,gocognit // template extraction for multiple plugins with filtering
	items := 0
	defer func() { sendPluginSubcommandResult("templates", "dump", items, err) }()

	// Auto-register external plugins so dump also reaches contrib
	// plugins (paletty, awob, …), not just the built-in set.
	_ = loadAndApplyPluginManifest()

	plugins := sharedPluginManager.AllOutputPlugins()
	if len(plugins) == 0 {
		return fmt.Errorf("no output plugins available")
	}
	items = len(plugins)

	// Filter plugins if specified.
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

	// Expand the custom location if provided.
	customBase := templateLocation
	if customBase != "" {
		// Expand tilde to home directory if present.
		if strings.HasPrefix(customBase, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			customBase = filepath.Join(home, customBase[2:])
		}
		fmt.Printf("Dumping templates to custom location: %s\n", customBase)
		fmt.Println()
	}

	totalDumped := 0
	for pluginName, plugin := range plugins {
		loader := getPluginTemplateLoaderWithBase(pluginName, plugin, customBase)
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

		// Show what was successfully dumped.
		for _, path := range dumped {
			fmt.Printf("   %s\n", path)
			totalDumped++
		}

		// Handle errors (likely "already exists" messages).
		if err == nil {
			continue
		}

		// Check if it's an "already exists" error.
		if templateForce || !strings.Contains(err.Error(), "already exists") {
			// Other errors should fail.
			return fmt.Errorf("failed to dump templates for %s: %w", pluginName, err)
		}

		// Split multiple errors and show each skipped file.
		errorParts := strings.SplitSeq(err.Error(), "; ")
		for errPart := range errorParts {
			if !strings.Contains(errPart, "already exists") {
				continue
			}
			// Extract just the filename from the path.
			idx := strings.Index(errPart, "already exists: ")
			if idx == -1 {
				continue
			}
			path := strings.TrimPrefix(errPart[idx:], "already exists: ")
			path = strings.TrimSuffix(path, " (use --force to overwrite)")
			fmt.Printf("  ⊘ %s (already exists)\n", path)
		}
		if len(dumped) == 0 {
			fmt.Fprintf(os.Stderr, "  Use --force to overwrite existing templates\n")
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

// getPluginTemplateLoader returns a template loader for the given plugin.
// Returns nil if the plugin doesn't support templates (i.e., doesn't implement TemplateProvider).
func getPluginTemplateLoader(pluginName string, plugin output.Plugin) *template.Loader {
	return getPluginTemplateLoaderWithBase(pluginName, plugin, "")
}

// getPluginTemplateLoaderWithBase returns a template loader for the
// given plugin with an optional custom base directory. Returns nil
// when the plugin doesn't expose any templates.
//
// Resolution:
//  1. In-tree plugins satisfy output.TemplateProvider → loader uses
//     the plugin's embed.FS directly.
//  2. External plugins satisfy externalTemplateLister (Templates()
//     returns map[string][]byte) → loader is built over a synthetic
//     fstest.MapFS populated from the RPC response.
//  3. Plugins that satisfy neither return nil and the templates
//     commands silently skip them.
func getPluginTemplateLoaderWithBase(pluginName string, plugin output.Plugin, customBase string) *template.Loader {
	var loader *template.Loader

	if tp, ok := plugin.(output.TemplateProvider); ok {
		if embedFS, ok := tp.GetEmbeddedFS().(embed.FS); ok {
			loader = template.New(pluginName, embedFS)
		}
	}

	if loader == nil {
		if etl, ok := plugin.(externalTemplateLister); ok {
			if tmpls := etl.Templates(); len(tmpls) > 0 {
				mapFS := fstest.MapFS{}
				for name, content := range tmpls {
					mapFS[name] = &fstest.MapFile{Data: content}
				}
				loader = template.NewFromFS(pluginName, mapFS)
			}
		}
	}

	if loader == nil {
		return nil
	}

	if customBase != "" {
		loader = loader.WithCustomBase(customBase)
	}

	return loader
}

// externalTemplateLister mirrors the SDK's TemplateLister contract on
// the host side without dragging the SDK package into this CLI file.
// ExternalOutputPlugin in the manager satisfies this shape; in-tree
// plugins use TemplateProvider above and never hit this path.
type externalTemplateLister interface {
	Templates() map[string][]byte
}
