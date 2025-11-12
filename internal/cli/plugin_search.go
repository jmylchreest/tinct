// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
)

var (
	searchType   string
	searchTags   []string
	searchAuthor string
	searchRepo   string
)

// pluginSearchCmd searches for plugins across repositories.
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

func init() {
	// Add search command to plugins.
	pluginsCmd.AddCommand(pluginSearchCmd)

	// Search flags.
	pluginSearchCmd.Flags().StringVar(&searchType, "type", "", "Filter by plugin type (input/output)")
	pluginSearchCmd.Flags().StringSliceVar(&searchTags, "tag", []string{}, "Filter by tags")
	pluginSearchCmd.Flags().StringVar(&searchAuthor, "author", "", "Filter by author")
	pluginSearchCmd.Flags().StringVar(&searchRepo, "repo", "", "Search only in specific repository")
}

func runPluginSearch(_ *cobra.Command, args []string) error {
	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	// Check if any repositories are configured.
	repos := mgr.ListRepositories()
	if len(repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println("\nAdd a repository with:")
		fmt.Printf("  tinct plugins repo add official %s\n", repository.OfficialRepoURL)
		return nil
	}

	// Build search filter.
	filter := repository.SearchFilter{
		Type:       searchType,
		Tags:       searchTags,
		Author:     searchAuthor,
		Repository: searchRepo,
	}

	if len(args) > 0 {
		filter.Query = args[0]
	}

	// Search for plugins.
	results, err := mgr.Search(filter)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No plugins found matching your criteria.")
		return nil
	}

	// Display results in table format matching plugins list
	table := NewTable([]string{"TYPE", "PLUGIN", "VERSION", "REPO", "DESCRIPTION"})

	// Enable terminal-aware column sizing for description
	table.EnableTerminalAwareWidth(4, 40) // Min width of 40 chars for description

	for _, result := range results {
		version := ""
		if result.Version != nil {
			version = result.Version.Version
		}

		table.AddRow([]string{
			result.Plugin.Type,
			result.Plugin.Name,
			version,
			result.Repository,
			result.Plugin.Description,
		})
	}

	fmt.Print(table.Render())

	fmt.Printf("\nFound %d plugin(s) in repositories\n", len(results))
	fmt.Println("\nInstall with: tinct plugins install <plugin-name>")
	fmt.Println("Example: tinct plugins install", results[0].Plugin.Name)

	return nil
}
