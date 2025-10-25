// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"strings"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
	"github.com/spf13/cobra"
)

var (
	searchType   string
	searchTags   []string
	searchAuthor string
	searchRepo   string
)

// pluginSearchCmd searches for plugins across repositories
var pluginSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for plugins in repositories",
	Long: `Search for plugins across all configured repositories.

You can search by plugin name or description, and filter by type, tags, or author.

Examples:
  tinct plugins search random           # Search for "random"
  tinct plugins search --type input     # List all input plugins
  tinct plugins search --tag notification # Find notification plugins
  tinct plugins search --author "John"  # Find plugins by author`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPluginSearch,
}

// pluginBrowseCmd lists all available plugins
var pluginBrowseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse all available plugins",
	Long: `List all plugins available in configured repositories.

This shows a comprehensive list of all plugins with their details.`,
	RunE: runPluginBrowse,
}

// pluginInfoCmd shows detailed plugin information
var pluginInfoCmd = &cobra.Command{
	Use:   "info <plugin-name>",
	Short: "Show detailed plugin information",
	Long: `Display detailed information about a specific plugin from repositories.

This includes all available versions, download information, and metadata.`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginInfo,
}

func init() {
	// Add search commands to plugins
	pluginsCmd.AddCommand(pluginSearchCmd)
	pluginsCmd.AddCommand(pluginBrowseCmd)
	pluginsCmd.AddCommand(pluginInfoCmd)

	// Search flags
	pluginSearchCmd.Flags().StringVar(&searchType, "type", "", "Filter by plugin type (input/output)")
	pluginSearchCmd.Flags().StringSliceVar(&searchTags, "tag", []string{}, "Filter by tags")
	pluginSearchCmd.Flags().StringVar(&searchAuthor, "author", "", "Filter by author")
	pluginSearchCmd.Flags().StringVar(&searchRepo, "repo", "", "Search only in specific repository")

	// Browse flags
	pluginBrowseCmd.Flags().StringVar(&searchType, "type", "", "Filter by plugin type (input/output)")
	pluginBrowseCmd.Flags().StringVar(&searchRepo, "repo", "", "Browse only specific repository")
}

func runPluginSearch(cmd *cobra.Command, args []string) error {
	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	// Check if any repositories are configured
	repos := mgr.ListRepositories()
	if len(repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println("\nAdd a repository with:")
		fmt.Println("  tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/main/repository.json")
		return nil
	}

	// Build search filter
	filter := repository.SearchFilter{
		Type:       searchType,
		Tags:       searchTags,
		Author:     searchAuthor,
		Repository: searchRepo,
	}

	if len(args) > 0 {
		filter.Query = args[0]
	}

	// Search for plugins
	results, err := mgr.Search(filter)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No plugins found matching your criteria.")
		return nil
	}

	// Display results
	fmt.Printf("Found %d plugin(s):\n\n", len(results))

	table := NewTable([]string{"NAME", "TYPE", "VERSION", "REPO", "DESCRIPTION"})

	for _, result := range results {
		version := ""
		if result.Version != nil {
			version = result.Version.Version
		}

		description := result.Plugin.Description
		if len(description) > 60 {
			description = description[:57] + "..."
		}

		table.AddRow([]string{
			result.Plugin.Name,
			result.Plugin.Type,
			version,
			result.Repository,
			description,
		})
	}

	fmt.Print(table.Render())

	fmt.Println("\nUse 'tinct plugins info <name>' for more details.")
	fmt.Println("Install with 'tinct plugins install <name>'")

	return nil
}

func runPluginBrowse(cmd *cobra.Command, args []string) error {
	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	// Check if any repositories are configured
	repos := mgr.ListRepositories()
	if len(repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println("\nAdd a repository with:")
		fmt.Println("  tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/main/repository.json")
		return nil
	}

	// Build filter
	filter := repository.SearchFilter{
		Type:       searchType,
		Repository: searchRepo,
	}

	// Search for all plugins
	results, err := mgr.Search(filter)
	if err != nil {
		return fmt.Errorf("failed to browse plugins: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No plugins available.")
		return nil
	}

	// Group by repository
	byRepo := make(map[string][]*repository.SearchResult)
	for _, result := range results {
		byRepo[result.Repository] = append(byRepo[result.Repository], result)
	}

	fmt.Printf("Available Plugins (%d total):\n\n", len(results))

	for repoName, repoResults := range byRepo {
		fmt.Printf("Repository: %s (%d plugin(s))\n", repoName, len(repoResults))
		fmt.Println(strings.Repeat("-", 80))

		for _, result := range repoResults {
			version := ""
			if result.Version != nil {
				version = result.Version.Version
			}

			fmt.Printf("  %s (%s) - v%s\n", result.Plugin.Name, result.Plugin.Type, version)
			fmt.Printf("    %s\n", result.Plugin.Description)

			if len(result.Plugin.Tags) > 0 {
				fmt.Printf("    Tags: %s\n", strings.Join(result.Plugin.Tags, ", "))
			}

			fmt.Println()
		}
	}

	fmt.Println("Use 'tinct plugins info <name>' for more details.")
	fmt.Println("Install with 'tinct plugins install <name>'")

	return nil
}

func runPluginInfo(cmd *cobra.Command, args []string) error {
	pluginName := args[0]

	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	// Find plugin
	result, err := mgr.FindPlugin(pluginName, "")
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	plugin := result.Plugin

	// Display plugin information
	fmt.Printf("Plugin: %s\n", plugin.Name)
	fmt.Printf("Type: %s\n", plugin.Type)
	fmt.Printf("Description: %s\n", plugin.Description)
	fmt.Printf("Repository: %s\n", result.Repository)

	if plugin.Author != "" {
		fmt.Printf("Author: %s\n", plugin.Author)
	}

	if plugin.License != "" {
		fmt.Printf("License: %s\n", plugin.License)
	}

	if len(plugin.Tags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(plugin.Tags, ", "))
	}

	if plugin.Repository != "" {
		fmt.Printf("Source: %s\n", plugin.Repository)
	}

	// Display versions
	if len(plugin.Versions) > 0 {
		fmt.Printf("\nAvailable Versions:\n")

		for i, v := range plugin.Versions {
			marker := ""
			if i == 0 {
				marker = " (latest)"
			}

			fmt.Printf("  %s%s", v.Version, marker)

			if v.Compatibility != "" {
				fmt.Printf(" - Requires tinct %s", v.Compatibility)
			}

			if !v.Released.IsZero() {
				fmt.Printf(" - Released %s", v.Released.Format("2006-01-02"))
			}

			fmt.Println()

			// Show platforms
			if len(v.Downloads) > 0 {
				platforms := make([]string, 0, len(v.Downloads))
				for platform := range v.Downloads {
					platforms = append(platforms, platform)
				}
				fmt.Printf("    Platforms: %s\n", strings.Join(platforms, ", "))
			}

			if v.ChangelogURL != "" {
				fmt.Printf("    Changelog: %s\n", v.ChangelogURL)
			}
		}
	}

	fmt.Println("\nInstall with:")
	fmt.Printf("  tinct plugins install %s\n", pluginName)

	if len(plugin.Versions) > 1 {
		fmt.Println("\nInstall specific version:")
		fmt.Printf("  tinct plugins install %s@%s\n", pluginName, plugin.Versions[1].Version)
	}

	return nil
}
