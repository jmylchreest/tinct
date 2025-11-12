// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
)

var (
	installRepository string
	installVersion    string
	installForce      bool
)

// pluginInstallCmd installs a plugin from repositories.
var pluginInstallCmd = &cobra.Command{
	Use:   "install <plugin-name>",
	Short: "Install a plugin from repositories",
	Long: `Install a plugin from configured repositories by name.

The command will:
  1. Search for the plugin in configured repositories
  2. Select the appropriate version and platform download
  3. Download and install the plugin
  4. Register it in the lock file with repository source tracking

Repository source tracking enables:
  - Version pinning and updates via 'tinct plugins update'
  - Verification of installed plugins
  - Tracking where plugins came from

By default, searches all configured repositories. Use --repository to specify
a particular repository to install from.

Examples:
  tinct plugins install random                         # Install latest from any repository
  tinct plugins install random --repository official   # Install from specific repo
  tinct plugins install random --version 0.0.2         # Install specific version
  tinct plugins install random --force                 # Force reinstall`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginInstall,
}

func init() {
	// Add install command to plugins.
	pluginsCmd.AddCommand(pluginInstallCmd)

	// Install flags.
	pluginInstallCmd.Flags().StringVar(&installRepository, "repository", "", "Install from specific repository")
	pluginInstallCmd.Flags().StringVar(&installVersion, "version", "latest", "Plugin version to install (default: latest)")
	pluginInstallCmd.Flags().BoolVarP(&installForce, "force", "f", false, "Force reinstall if already installed")
}

func runPluginInstall(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Get repository manager.
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
		return fmt.Errorf("no repositories configured")
	}

	// Find the plugin.
	var result *repository.SearchResult
	if installRepository != "" {
		// Search in specific repository.
		result, err = mgr.FindPluginInRepository(installRepository, pluginName, installVersion)
		if err != nil {
			return fmt.Errorf("plugin %q not found in repository %q: %w", pluginName, installRepository, err)
		}
	} else {
		// Search in all repositories.
		result, err = mgr.FindPlugin(pluginName, installVersion)
		if err != nil {
			return fmt.Errorf("plugin %q not found in any repository: %w", pluginName, err)
		}
	}

	// Determine current platform with normalization.
	platform := repository.NormalizePlatform(runtime.GOOS, runtime.GOARCH)

	// Find download for current platform.
	download, ok := result.Version.Downloads[platform]
	if !ok {
		return fmt.Errorf("plugin %q version %s is not available for platform %s",
			pluginName, result.Version.Version, platform)
	}

	if !download.Available {
		reason := "unknown reason"
		if download.UnavailableReason != "" {
			reason = download.UnavailableReason
		}
		return fmt.Errorf("plugin %q version %s for %s is unavailable: %s",
			pluginName, result.Version.Version, platform, reason)
	}

	// Load or create plugin lock.
	lock, lockPath := loadOrCreatePluginLock()

	if verbose {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n", lockPath)
	}

	// Initialize external plugins map if needed.
	if lock.ExternalPlugins == nil {
		lock.ExternalPlugins = make(map[string]*ExternalPluginMeta)
	}

	// Check if plugin is already installed.
	existingMeta, exists := lock.ExternalPlugins[result.Plugin.Name]
	if exists && !installForce {
		return fmt.Errorf("plugin %q is already installed (version %s). Use --force to reinstall",
			result.Plugin.Name, existingMeta.Version)
	}

	// Display installation info.
	fmt.Printf("Installing plugin: %s\n", result.Plugin.Name)
	if result.Plugin.Description != "" {
		fmt.Printf("  Description: %s\n", result.Plugin.Description)
	}
	fmt.Printf("  Type: %s\n", result.Plugin.Type)
	fmt.Printf("  Version: %s\n", result.Version.Version)
	fmt.Printf("  Repository: %s\n", result.Repository)
	fmt.Printf("  Platform: %s\n", platform)
	if verbose {
		fmt.Printf("  URL: %s\n", download.URL)
		fmt.Printf("  Checksum: %s\n", download.Checksum)
	}
	fmt.Println()

	// Get plugin directory.
	pluginDir, err := getPluginDir()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	// Ensure plugin directory exists.
	if err := os.MkdirAll(pluginDir, 0o755); err != nil { // #nosec G301 - Plugin directory needs standard permissions
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Download and install from URL.
	pluginPath, err := installPluginFromSource(download.URL, "", pluginDir, sourceTypeHTTP, verbose)
	if err != nil {
		return fmt.Errorf("failed to install plugin: %w", err)
	}

	// Query plugin metadata to verify.
	metadata, err := queryFullPluginMetadata(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to query plugin metadata: %w", err)
	}

	// Verify plugin name matches.
	if metadata.Name != result.Plugin.Name {
		return fmt.Errorf("plugin name mismatch: expected %q, got %q", result.Plugin.Name, metadata.Name)
	}

	// Verify plugin type matches.
	if metadata.Type != result.Plugin.Type {
		return fmt.Errorf("plugin type mismatch: expected %s, got %s", result.Plugin.Type, metadata.Type)
	}

	// Determine if version should be pinned in source.
	// Only pin if user explicitly specified a version (not "latest").
	var pinnedVersion string
	if installVersion != "" && installVersion != "latest" {
		pinnedVersion = installVersion
	}

	// Update lock file with repository source tracking.
	lock.ExternalPlugins[metadata.Name] = &ExternalPluginMeta{
		Name:        metadata.Name,
		Path:        pluginPath,
		Type:        metadata.Type,
		Version:     metadata.Version,
		Description: metadata.Description,
		Source: &repository.PluginSource{
			Type:       sourceTypeRepository,
			Repository: result.Repository,
			Plugin:     metadata.Name,
			Version:    pinnedVersion, // Empty if not pinned, allows updates to latest
			Checksum:   download.Checksum,
		},
		InstalledAt: time.Now().Format(time.RFC3339),
	}

	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	fmt.Printf(" Plugin %q installed successfully\n", metadata.Name)
	fmt.Printf("  Installed at: %s\n", pluginPath)
	fmt.Printf("  Source: repository/%s\n", result.Repository)

	return nil
}
