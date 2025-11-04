// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
	"github.com/spf13/cobra"
)

var (
	repoManager    *repository.Manager
	repoPriority   int
	repoConfigPath string
	repoCachePath  string
)

// pluginRepoCmd represents the plugins repo command.
var pluginRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage plugin repositories",
	Long: `Manage plugin repositories for discovering and installing plugins.

Repositories are manifests that list available plugins with their versions
and download links. This allows for easy plugin discovery and installation
without storing binaries in the repository itself.`,
}

// pluginRepoAddCmd adds a new repository.
var pluginRepoAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a plugin repository",
	Long: `Add a new plugin repository by providing a name and URL to the repository manifest.

The manifest URL should point to a JSON file containing plugin metadata.

Example:
  tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/main/repository.json`,
	Args: cobra.ExactArgs(2),
	RunE: runPluginRepoAdd,
}

// pluginRepoListCmd lists all repositories.
var pluginRepoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured repositories",
	Long:  `List all configured plugin repositories with their status.`,
	RunE:  runPluginRepoList,
}

// pluginRepoRemoveCmd removes a repository.
var pluginRepoRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a plugin repository",
	Long:    `Remove a configured plugin repository by name.`,
	Args:    cobra.ExactArgs(1),
	RunE:    runPluginRepoRemove,
}

// pluginRepoUpdateCmd updates repository manifests.
var pluginRepoUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update repository manifests",
	Long: `Refresh plugin repository manifests from their source URLs.

If no repository name is provided, all enabled repositories will be updated.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPluginRepoUpdate,
}

// pluginRepoInfoCmd shows repository details.
var pluginRepoInfoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show repository information",
	Long:  `Display detailed information about a specific repository.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPluginRepoInfo,
}

func init() {
	// Add repo subcommands.
	pluginRepoCmd.AddCommand(pluginRepoAddCmd)
	pluginRepoCmd.AddCommand(pluginRepoListCmd)
	pluginRepoCmd.AddCommand(pluginRepoRemoveCmd)
	pluginRepoCmd.AddCommand(pluginRepoUpdateCmd)
	pluginRepoCmd.AddCommand(pluginRepoInfoCmd)

	// Add repo command to plugins.
	pluginsCmd.AddCommand(pluginRepoCmd)

	// Flags.
	pluginRepoAddCmd.Flags().IntVar(&repoPriority, "priority", 10, "Repository priority (lower = higher priority)")
}

func getRepoManager() (*repository.Manager, error) {
	if repoManager != nil {
		return repoManager, nil
	}

	// Default paths.
	if repoConfigPath == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get config directory: %w", err)
		}
		repoConfigPath = filepath.Join(configDir, "tinct", ".tinct-repos.json")
	}

	if repoCachePath == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get cache directory: %w", err)
		}
		repoCachePath = filepath.Join(cacheDir, "tinct", "repo-cache")
	}

	var err error
	repoManager, err = repository.NewManager(repoConfigPath, repoCachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise repository manager: %w", err)
	}

	return repoManager, nil
}

func runPluginRepoAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	url := args[1]

	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	fmt.Printf("Adding repository %q...\n", name)
	fmt.Printf("  URL: %s\n", url)

	if err := mgr.AddRepository(name, url, repoPriority); err != nil {
		return fmt.Errorf("failed to add repository: %w", err)
	}

	fmt.Printf(" Repository %q added successfully\n", name)
	fmt.Printf("  Priority: %d\n", repoPriority)
	fmt.Printf("\nRun 'tinct plugins search' to discover available plugins.\n")

	return nil
}

func runPluginRepoList(cmd *cobra.Command, args []string) error {
	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	repos := mgr.ListRepositories()

	if len(repos) == 0 {
		fmt.Println("No repositories configured.")
		fmt.Println("\nAdd a repository with:")
		fmt.Println("  tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/main/repository.json")
		return nil
	}

	fmt.Println("Configured Repositories:")
	fmt.Println()

	table := NewTable([]string{"NAME", "STATUS", "PRIORITY", "URL"})

	for _, repo := range repos {
		status := "enabled"
		if !repo.Enabled {
			status = "disabled"
		}

		table.AddRow([]string{
			repo.Name,
			status,
			strconv.Itoa(repo.Priority),
			repo.URL,
		})
	}

	fmt.Print(table.Render())
	return nil
}

func runPluginRepoRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	if err := mgr.RemoveRepository(name); err != nil {
		return fmt.Errorf("failed to remove repository: %w", err)
	}

	fmt.Printf(" Repository %q removed successfully\n", name)
	return nil
}

func runPluginRepoUpdate(cmd *cobra.Command, args []string) error {
	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		// Update all repositories.
		fmt.Println("Updating all repositories...")

		repos := mgr.ListRepositories()
		if len(repos) == 0 {
			fmt.Println("No repositories configured.")
			return nil
		}

		for _, repo := range repos {
			if !repo.Enabled {
				continue
			}

			fmt.Printf("\n%s...\n", repo.Name)
			if err := mgr.UpdateRepository(repo.Name); err != nil {
				fmt.Printf("   Failed: %v\n", err)
			} else {
				fmt.Printf("   Updated\n")
			}
		}

		fmt.Println("\n Repository update complete")
	} else {
		// Update specific repository.
		name := args[0]
		fmt.Printf("Updating repository %q...\n", name)

		if err := mgr.UpdateRepository(name); err != nil {
			return fmt.Errorf("failed to update repository: %w", err)
		}

		fmt.Printf(" Repository %q updated successfully\n", name)
	}

	return nil
}

func runPluginRepoInfo(cmd *cobra.Command, args []string) error {
	name := args[0]

	mgr, err := getRepoManager()
	if err != nil {
		return err
	}

	repo, err := mgr.GetRepository(name)
	if err != nil {
		return err
	}

	// Load manifest if not already loaded.
	if repo.Manifest == nil {
		fmt.Printf("Loading manifest for %q...\n", name)
		if err := mgr.UpdateRepository(name); err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}
		repo, err = mgr.GetRepository(name)
		if err != nil {
			return fmt.Errorf("failed to get repository: %w", err)
		}
	}

	fmt.Printf("Repository: %s\n", repo.Name)
	fmt.Printf("URL: %s\n", repo.URL)
	fmt.Printf("Status: %s\n", map[bool]string{true: "enabled", false: "disabled"}[repo.Enabled])
	fmt.Printf("Priority: %d\n", repo.Priority)

	if repo.Manifest != nil {
		fmt.Printf("\nManifest:\n")
		fmt.Printf("  Name: %s\n", repo.Manifest.Name)
		fmt.Printf("  Description: %s\n", repo.Manifest.Description)
		if repo.Manifest.MaintainedBy != "" {
			fmt.Printf("  Maintained By: %s\n", repo.Manifest.MaintainedBy)
		}
		if !repo.Manifest.LastUpdated.IsZero() {
			fmt.Printf("  Last Updated: %s\n", repo.Manifest.LastUpdated.Format(time.RFC3339))
		}
		fmt.Printf("  Plugins: %d\n", len(repo.Manifest.Plugins))

		// List plugin names.
		if len(repo.Manifest.Plugins) > 0 {
			fmt.Printf("\nAvailable Plugins:\n")
			for name, plugin := range repo.Manifest.Plugins {
				versionCount := len(plugin.Versions)
				latestVersion := ""
				if versionCount > 0 {
					latestVersion = plugin.Versions[0].Version
				}
				fmt.Printf("  - %s (%s) - %d version(s), latest: %s\n",
					name,
					plugin.Type,
					versionCount,
					latestVersion,
				)
			}
		}
	}

	return nil
}
