// Package cli provides the command-line interface for Tinct.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/plugin/repository"
)

const (
	// PluginLockFile is the name of the plugin lock file in the config directory.
	PluginLockFile = "plugins.lock.json"

	// legacyLockFileName is the old lock file name used before migration.
	legacyLockFileName = ".tinct-plugins.json"
)

// PluginLock represents the plugin lock file structure.
// This file tracks installed external plugins and their sources.
type PluginLock struct {
	// ExternalPlugins maps plugin names to their metadata.
	ExternalPlugins map[string]*ExternalPluginMeta `json:"external_plugins,omitempty"`
}

// ExternalPluginMeta contains metadata about an external plugin.
type ExternalPluginMeta struct {
	// Name is the plugin's actual name (from --plugin-info).
	Name string `json:"name"`

	// Path is the absolute path to the plugin executable.
	Path string `json:"path"`

	// Type is the plugin type (input or output).
	Type string `json:"type"`

	// Version is the plugin version if available.
	Version string `json:"version,omitempty"`

	// Description is the plugin description if available.
	Description string `json:"description,omitempty"`

	// Source contains structured information about where the plugin came from.
	Source *repository.PluginSource `json:"source,omitempty"`

	// SourceLegacy is the old string-based source field for backward compatibility.
	SourceLegacy string `json:"source_legacy,omitempty"`
}

var (
	// Plugin command flags.
	pluginLockPath   string
	pluginType       string
	pluginForce      bool
	pluginYes        bool
	pluginSourceType string
	pluginNoCopy     bool
	pluginShowPath   bool
)

// pluginsCmd represents the plugins command.
var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Manage plugins",
	Long: `Manage Tinct plugins including listing and managing external plugins.

The plugin lock file (~/.config/tinct/plugins.lock.json) tracks installed
external plugins and their sources. Use 'tinct plugins sync' to reproduce
your plugin setup on another machine.

Commands that modify the lock file:
  - add: Adds external plugin and updates lock file
  - delete: Removes external plugin and updates lock file
  - update: Updates external plugins from their sources`,
}

// pluginListCmd lists all available plugins.
var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available plugins",
	Long: `List all available plugins.

Shows both built-in and external plugins with their type and description.`,
	RunE: runPluginList,
}

// pluginAddCmd adds an external plugin.
var pluginAddCmd = &cobra.Command{
	Use:   "add <source>",
	Short: "Add an external plugin",
	Long: `Add an external plugin from a local file, HTTP URL, or Git repository.

The plugin will be copied to the plugin directory and registered
in the plugin lock file. The plugin name is automatically detected from
the plugin's --plugin-info output.

WARNING: Only install plugins from trusted sources. Plugins execute with your
user permissions and can access your system. Review plugin source code before
installation to ensure it is safe.

The command will:
  1. Verify source and destination are not the same file
  2. Query plugin metadata (name, version, type, protocol)
  3. Check protocol compatibility
  4. Check for version conflicts (upgrades proceed automatically)
  5. Copy plugin to ~/.local/share/tinct/plugins/ (unless --no-copy is used)
  6. Register plugin in lock file

Plugin upgrades (newer versions) proceed automatically.
Use --force to downgrade, reinstall same version, or overwrite.
Use --no-copy to reference the plugin at its current location without copying
(useful for system-installed packages that manage their own updates).

Examples:
  tinct plugins add ./contrib/notify-send.py
  tinct plugins add https://example.com/plugins/theme.sh
  tinct plugins add https://github.com/user/plugin.git
  tinct plugins add https://github.com/user/plugin.git:path/to/plugin.sh
  tinct plugins add ./my-plugin.sh --force  # Force overwrite
  tinct plugins add /usr/bin/tinct-plugin-random --no-copy  # System package`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginAdd,
}

// pluginDeleteCmd removes an external plugin.
var pluginDeleteCmd = &cobra.Command{
	Use:   "delete <plugin-name>",
	Short: "Delete an external plugin",
	Long: `Delete an external plugin from the plugin directory and remove it from the lock file.

Built-in plugins cannot be deleted.

Examples:
  tinct plugins delete notify
  tinct plugins delete custom-theme`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginDelete,
}

// pluginUpdateCmd updates external plugins from lock file.
var pluginUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update external plugins from lock file sources",
	Long: `Update external plugins by re-downloading/copying from their source locations.

This reads the plugin lock file and updates all external plugins based on their
source field. Useful for keeping plugins in sync across machines or after pulling
changes to the lock file.

Examples:
  tinct plugins update
  tinct plugins update --lock-file /path/to/plugins.lock.json`,
	RunE: runPluginUpdate,
}

func init() {
	// Add plugins command flags.
	pluginsCmd.PersistentFlags().StringVar(&pluginLockPath, "lock-file", "", "path to plugin lock file (default: ~/.config/tinct/plugins.lock.json)")

	// Plugin list flags.
	pluginListCmd.Flags().BoolVar(&pluginShowPath, "show-path", false, "show the actual file path used when loading each plugin")

	// Plugin add flags.
	pluginAddCmd.Flags().StringVar(&pluginType, "type", "output", "plugin type (input or output)")
	pluginAddCmd.Flags().BoolVarP(&pluginForce, "force", "f", false, "force overwrite if plugin already exists")
	pluginAddCmd.Flags().StringVar(&pluginSourceType, "source-type", "", "force source type (local, http, git) - auto-detected if not specified")
	pluginAddCmd.Flags().BoolVar(&pluginNoCopy, "no-copy", false, "register plugin at its current location without copying (useful for system packages)")

	// Plugin delete flags.
	pluginDeleteCmd.Flags().BoolVarP(&pluginForce, "force", "f", false, "force deletion without confirmation")

	// Add subcommands.
	pluginsCmd.AddCommand(pluginListCmd)
	pluginsCmd.AddCommand(pluginAddCmd)
	pluginsCmd.AddCommand(pluginDeleteCmd)
	pluginsCmd.AddCommand(pluginUpdateCmd)
}

// runPluginList lists all available plugins.
func runPluginList(cmd *cobra.Command, _ []string) error {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load plugin lock and create manager.
	lock, lockPath, err := loadPluginLock()
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Note: %v\n", err)
	}

	mgr := createManagerFromLock(lock)

	if verbose && lockPath != "" {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n\n", lockPath)
	}

	// Collect all plugins.
	allPlugins := collectAllPlugins(mgr, lock)

	// Display plugins.
	displayPluginTable(allPlugins, pluginShowPath)

	return nil
}

// runPluginAdd adds an external plugin with comprehensive safety checks.
func runPluginAdd(cmd *cobra.Command, args []string) error {
	source := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
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

	// Get plugin directory.
	pluginDir, err := getPluginDir()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Plugin directory: %s\n", pluginDir)
	}

	// Stage 1: Resolve source path and check if it's already in the plugin directory
	sourcePath, isAlreadyInstalled, err := resolvePluginSource(source, pluginDir, pluginSourceType, verbose)
	if err != nil {
		return err
	}

	if isAlreadyInstalled {
		if verbose {
			fmt.Fprintf(os.Stderr, "Note: Plugin source is already in plugin directory\n")
		}
	}

	// Stage 2: Query plugin metadata (before copying to avoid executing untrusted code from final location)
	pluginInfo, err := queryFullPluginMetadata(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to query plugin metadata: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Plugin name: %s\n", pluginInfo.Name)
		fmt.Fprintf(os.Stderr, "Plugin type: %s\n", pluginInfo.Type)
		fmt.Fprintf(os.Stderr, "Plugin version: %s\n", pluginInfo.Version)
		fmt.Fprintf(os.Stderr, "Protocol version: %s\n", pluginInfo.ProtocolVersion)
	}

	// Stage 3: Check protocol compatibility
	if err := checkProtocolCompatibility(pluginInfo.ProtocolVersion, verbose); err != nil {
		return err
	}

	// Stage 4: Check for conflicts and version comparisons
	action, existingMeta, err := determinePluginAction(lock, pluginInfo, pluginForce)
	if err != nil {
		return err
	}

	// Stage 5: Install plugin to final location (if not already there)
	var finalPath string
	if pluginNoCopy {
		// Use the source path directly without copying
		finalPath = sourcePath
		if verbose {
			fmt.Fprintf(os.Stderr, "Using plugin at: %s (no-copy mode)\n", finalPath)
		}

		// If we're overwriting an existing plugin that was copied to the plugin dir, clean it up
		if existingMeta != nil && existingMeta.Path != "" {
			existingPathDir := filepath.Dir(existingMeta.Path)
			if existingPathDir == pluginDir {
				// Old plugin was in the plugin directory, remove it
				if err := os.Remove(existingMeta.Path); err != nil && !os.IsNotExist(err) {
					if verbose {
						fmt.Fprintf(os.Stderr, "Warning: failed to remove old plugin file: %v\n", err)
					}
				} else if verbose {
					fmt.Fprintf(os.Stderr, "Removed old plugin file: %s\n", existingMeta.Path)
				}
			}
		}
	} else {
		finalPath = filepath.Join(pluginDir, filepath.Base(sourcePath))
		if !isAlreadyInstalled {
			if err := installPlugin(sourcePath, finalPath, verbose); err != nil {
				return fmt.Errorf("failed to install plugin: %w", err)
			}
		}
	}

	// Stage 6: Update lock file
	// Build structured source metadata.
	pluginSource := buildPluginSource(source, pluginSourceType, finalPath)

	lock.ExternalPlugins[pluginInfo.Name] = &ExternalPluginMeta{
		Name:         pluginInfo.Name,
		Path:         finalPath,
		Type:         pluginInfo.Type,
		Source:       pluginSource,
		SourceLegacy: source,
		Version:      pluginInfo.Version,
		Description:  pluginInfo.Description,
	}

	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	// Stage 7: Display success message
	printPluginAddSuccess(pluginInfo, action, existingMeta, finalPath)
	return nil
}

// runPluginDelete removes an external plugin.
func runPluginDelete(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load plugin lock.
	lock, lockPath, err := loadPluginLock()
	if err != nil {
		return fmt.Errorf("failed to load plugin lock: %w", err)
	}

	if lock == nil || lock.ExternalPlugins == nil {
		return fmt.Errorf("no external plugins found")
	}

	// Check if plugin exists.
	meta, exists := lock.ExternalPlugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	// Confirm deletion if not forced.
	if !pluginForce {
		fmt.Printf("Are you sure you want to delete plugin '%s'? (y/N): ", pluginName)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if !strings.EqualFold(response, "y") {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete plugin file only if it's in the plugin directory.
	// Plugins added with --no-copy are not deleted from disk.
	pluginDir, err := getPluginDir()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	// Check if plugin is in the plugin directory.
	pluginPathDir := filepath.Dir(meta.Path)
	if pluginPathDir == pluginDir {
		// Plugin is in plugin directory, safe to delete.
		if err := os.Remove(meta.Path); err != nil && !os.IsNotExist(err) {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to delete plugin file: %v\n", err)
			}
		} else if verbose {
			fmt.Fprintf(os.Stderr, "Deleted plugin file: %s\n", meta.Path)
		}
	} else {
		// Plugin is outside plugin directory (--no-copy), don't delete.
		if verbose {
			fmt.Fprintf(os.Stderr, "Plugin file not deleted (outside plugin directory): %s\n", meta.Path)
		}
	}

	// Remove from lock file.
	delete(lock.ExternalPlugins, pluginName)

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	fmt.Printf("Plugin '%s' deleted successfully\n", pluginName)
	return nil
}

// updatePluginFromRepository updates a plugin from a repository source.
func updatePluginFromRepository(meta *ExternalPluginMeta, pluginDir string, verbose bool) (string, error) {
	// Get repository manager.
	mgr, err := getRepoManager()
	if err != nil {
		return "", fmt.Errorf("failed to get repository manager: %w", err)
	}

	// Find the plugin in the repository.
	// If Source.Version is empty, update to latest. Otherwise respect the pinned version.
	version := "latest"
	if meta.Source.Version != "" {
		version = meta.Source.Version
	}

	result, err := mgr.FindPluginInRepository(
		meta.Source.Repository,
		meta.Source.Plugin,
		version,
	)
	if err != nil {
		return "", fmt.Errorf("failed to find plugin in repository: %w", err)
	}

	// Determine current platform.
	platform := repository.NormalizePlatform(runtime.GOOS, runtime.GOARCH)

	// Find download for current platform.
	download, ok := result.Version.Downloads[platform]
	if !ok {
		return "", fmt.Errorf("plugin not available for platform %s", platform)
	}

	if !download.Available {
		reason := "unknown reason"
		if download.UnavailableReason != "" {
			reason = download.UnavailableReason
		}
		return "", fmt.Errorf("plugin unavailable: %s", reason)
	}

	// Check if we need to update by comparing repository versions.
	// Use the source's checksum to detect if we already have this exact version.
	if meta.Source.Checksum != "" && download.Checksum == meta.Source.Checksum {
		// Same checksum = same file, no update needed.
		return "", fmt.Errorf("already up to date (version %s)", result.Version.Version)
	}

	// If pinned to a specific version and it matches, no update needed.
	if meta.Source.Version != "" && result.Version.Version == meta.Source.Version {
		return "", fmt.Errorf("already up to date (version %s)", meta.Source.Version)
	}

	// Download and install from URL (quiet=true since caller shows table status).
	pluginPath, err := installPluginFromSource(download.URL, "", pluginDir, sourceTypeHTTP, verbose, true)
	if err != nil {
		return "", fmt.Errorf("failed to download plugin: %w", err)
	}

	// Update the source metadata with the new checksum.
	// This is critical so next update can detect we already have this version.
	meta.Source.Checksum = download.Checksum

	return pluginPath, nil
}

// runPluginUpdate updates external plugins from lock file sources.
func runPluginUpdate(cmd *cobra.Command, args []string) error {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load plugin lock.
	lock, lockPath, err := loadPluginLock()
	if err != nil {
		return fmt.Errorf("failed to load plugin lock: %w", err)
	}

	if lock == nil || lock.ExternalPlugins == nil || len(lock.ExternalPlugins) == 0 {
		fmt.Println("No external plugins found in lock file")
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n", lockPath)
	}

	// Get plugin directory.
	pluginDir, err := getPluginDir()
	if err != nil {
		return fmt.Errorf("failed to get plugin directory: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Plugin directory: %s\n", pluginDir)
	}

	// Create live table for updates
	fmt.Fprintln(os.Stderr, "\nUpdating plugins...")
	fmt.Fprintln(os.Stderr)
	table := NewTable([]string{"PLUGIN", "STATUS"}).SetLive(true).WithWriter(os.Stderr)

	// Get sorted plugin names
	pluginNames := make([]string, 0, len(lock.ExternalPlugins))
	for name := range lock.ExternalPlugins {
		pluginNames = append(pluginNames, name)
	}
	sort.Strings(pluginNames)

	// Initialize all rows with "Queued" status
	// In live mode, rendering is deferred until Finish() is called
	for _, name := range pluginNames {
		table.AddRowWithID(name, []string{name, "Queued"})
	}

	// Update each external plugin
	successCount := 0
	failCount := 0

	for _, name := range pluginNames {
		meta := lock.ExternalPlugins[name]
		oldVersion := meta.Version

		table.UpdateRow(name, map[string]string{"STATUS": "Checking..."})

		// Handle repository sources specially.
		var pluginPath string
		var err error
		var skipUpdate bool
		if meta.Source != nil && meta.Source.Type == sourceTypeRepository {
			table.UpdateRow(name, map[string]string{"STATUS": "Downloading..."})
			pluginPath, err = updatePluginFromRepository(meta, pluginDir, verbose)
			// Check if error is "already up to date"
			if err != nil && strings.Contains(err.Error(), "already up to date") {
				skipUpdate = true
				err = nil
			}
		} else {
			// Install plugin from other source types (HTTP, local, git).
			var sourceForInstall string
			if meta.Source != nil && meta.Source.URL != "" {
				sourceForInstall = meta.Source.URL
			} else if meta.SourceLegacy != "" {
				sourceForInstall = meta.SourceLegacy
			} else {
				table.UpdateRow(name, map[string]string{"STATUS": "✗ No source information"})
				failCount++
				continue
			}
			table.UpdateRow(name, map[string]string{"STATUS": "Downloading..."})
			pluginPath, err = installPluginFromSource(sourceForInstall, name, pluginDir, "", verbose, true)
		}

		if err != nil {
			table.UpdateRow(name, map[string]string{"STATUS": fmt.Sprintf("✗ %v", err)})
			failCount++
			continue
		}

		if skipUpdate {
			status := "Already up to date"
			if oldVersion != "" {
				status = fmt.Sprintf("Already up to date (%s)", oldVersion)
			}
			table.UpdateRow(name, map[string]string{"STATUS": status})
			successCount++
			continue
		}

		table.UpdateRow(name, map[string]string{"STATUS": "Installing..."})

		// Query plugin for updated metadata.
		actualName, pluginDescription, pluginType, version, _ := queryPluginMetadata(pluginPath)
		if actualName == "" {
			actualName = meta.Name // Keep existing name if query fails
			if actualName == "" {
				actualName = name // Fallback to lock file key
			}
		}
		if pluginType == "" {
			pluginType = meta.Type // Keep existing type if query fails
		}

		// Update metadata in lock file.
		lock.ExternalPlugins[name] = &ExternalPluginMeta{
			Name:        actualName,
			Path:        pluginPath,
			Type:        pluginType,
			Source:      meta.Source,
			Version:     version,
			Description: pluginDescription,
		}

		// Show version change or just new version
		var status string
		if oldVersion != "" && version != "" && oldVersion != version {
			status = fmt.Sprintf("%s → %s", oldVersion, version)
		} else if version != "" {
			status = fmt.Sprintf("Updated (%s)", version)
		} else {
			status = "Updated"
		}
		table.UpdateRow(name, map[string]string{"STATUS": status})
		successCount++
	}

	table.Finish()

	// Save updated lock file.
	if successCount > 0 {
		if err := savePluginLock(lockPath, lock); err != nil {
			return fmt.Errorf("failed to save plugin lock: %w", err)
		}
	}

	if failCount > 0 {
		return fmt.Errorf("some plugins failed to update")
	}

	return nil
}

// getDefaultLockFilePath returns the default lock file path at ~/.config/tinct/plugins.lock.json.
func getDefaultLockFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}
	return filepath.Join(configDir, "tinct", PluginLockFile), nil
}

// migrateLegacyLockFile checks for lock files at legacy locations and migrates
// them to the new default path. Returns the path if found and migrated, or empty string.
func migrateLegacyLockFile(newPath string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Check legacy locations in order of priority.
	legacyPaths := []string{
		filepath.Join(home, legacyLockFileName), // ~/.tinct-plugins.json
		legacyLockFileName,                      // ./.tinct-plugins.json (CWD)
	}

	for _, legacyPath := range legacyPaths {
		data, err := os.ReadFile(legacyPath) // #nosec G304 - Legacy lock file path
		if err != nil {
			continue
		}

		// Ensure the new directory exists.
		if err := os.MkdirAll(filepath.Dir(newPath), 0o750); err != nil {
			return ""
		}

		// Write to new location.
		if err := os.WriteFile(newPath, data, 0o600); err != nil {
			return ""
		}

		// Remove the old file (best effort).
		_ = os.Remove(legacyPath)

		fmt.Fprintf(os.Stderr, "Migrated plugin lock file: %s → %s\n", legacyPath, newPath)
		return newPath
	}

	return ""
}

// loadPluginLock loads the plugin lock file.
func loadPluginLock() (*PluginLock, string, error) {
	lockPath := pluginLockPath

	if lockPath == "" {
		// Use the new default path.
		defaultPath, err := getDefaultLockFilePath()
		if err != nil {
			return nil, "", fmt.Errorf("no plugin lock file found")
		}

		if _, err := os.Stat(defaultPath); err == nil {
			lockPath = defaultPath
		} else {
			// Try migrating from legacy locations.
			migratedPath := migrateLegacyLockFile(defaultPath)
			if migratedPath != "" {
				lockPath = migratedPath
			} else {
				return nil, "", fmt.Errorf("no plugin lock file found")
			}
		}
	}

	data, err := os.ReadFile(lockPath) // #nosec G304 - Lock file path controlled by application
	if err != nil {
		return nil, "", fmt.Errorf("failed to read plugin lock file: %w", err)
	}

	var lock PluginLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, "", fmt.Errorf("failed to parse plugin lock file: %w", err)
	}

	return &lock, lockPath, nil
}

// loadOrCreatePluginLock loads or creates a plugin lock file.
// Always succeeds by creating a new lock if one doesn't exist.
func loadOrCreatePluginLock() (lock *PluginLock, lockPath string) {
	lock, lockPath, err := loadPluginLock()
	if err == nil {
		return lock, lockPath
	}

	// Create new lock file at the default path.
	lockPath = pluginLockPath
	if lockPath == "" {
		if defaultPath, err := getDefaultLockFilePath(); err == nil {
			lockPath = defaultPath
		} else {
			lockPath = PluginLockFile
		}
	}

	lock = &PluginLock{
		ExternalPlugins: make(map[string]*ExternalPluginMeta),
	}

	return lock, lockPath
}

// savePluginLock saves the plugin lock file.
func savePluginLock(path string, lock *PluginLock) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugin lock: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to update lock file: %w", err)
	}

	return nil
}

// createManagerFromLock creates a plugin manager from a lock file.
func createManagerFromLock(lock *PluginLock) *manager.Manager {
	mgr := manager.NewBuilder().Build()

	if lock == nil {
		return mgr
	}

	// Register external plugins using their actual names.
	if lock.ExternalPlugins != nil {
		for _, meta := range lock.ExternalPlugins {
			// Use the plugin's actual name (from metadata) not the lock file key.
			pluginName := meta.Name
			if pluginName == "" {
				// Fallback: query the plugin if name is missing.
				pluginName, _, _, _, _ = queryPluginMetadata(meta.Path)
			}

			// Use plugin's description if available.
			desc := meta.Description
			if desc == "" {
				desc = fmt.Sprintf("External plugin (source: %s)", meta.Source)
			}

			if err := mgr.RegisterExternalPlugin(pluginName, meta.Type, meta.Path, desc); err != nil {
				// Silently ignore registration errors.
				continue
			}
		}
	}

	return mgr
}

// loadAndApplyPluginLock loads the plugin lock file and registers any external
// plugins into the shared plugin manager.
func loadAndApplyPluginLock() error {
	lock, _, err := loadPluginLock()
	if err != nil {
		return err
	}

	if lock != nil {
		registerExternalPluginsFromLock(lock, true, false)
	}

	return nil
}

// registerExternalPluginsFromLock registers all external plugins from the lock file
// into the shared plugin manager. Optionally resolves relative paths to absolute.
func registerExternalPluginsFromLock(lock *PluginLock, resolveAbsolutePaths, verbose bool) {
	if lock == nil || lock.ExternalPlugins == nil {
		return
	}

	for _, meta := range lock.ExternalPlugins {
		if err := registerExternalPlugin(meta, resolveAbsolutePaths, verbose); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, " Failed to register external plugin '%s': %v\n", meta.Name, err)
			}
			// Continue with other plugins on error
		}
	}
}

// registerExternalPlugin registers a single external plugin into the shared manager.
func registerExternalPlugin(meta *ExternalPluginMeta, resolveAbsolutePaths, _ bool) error {
	// Use the plugin's actual name from metadata.
	pluginName := meta.Name
	if pluginName == "" {
		// Fallback: query the plugin if name is missing.
		pluginName, _, _, _, _ = queryPluginMetadata(meta.Path)
		if pluginName == "" {
			return fmt.Errorf("unable to determine plugin name")
		}
	}

	// Use plugin's description if available.
	desc := meta.Description
	if desc == "" {
		desc = fmt.Sprintf("External plugin (source: %s)", meta.Source)
	}

	// Convert relative paths to absolute if requested.
	pluginPath := meta.Path
	if resolveAbsolutePaths && !filepath.IsAbs(pluginPath) {
		absPath, err := filepath.Abs(pluginPath)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path: %w", err)
		}
		pluginPath = absPath
	}

	// Register the plugin.
	return sharedPluginManager.RegisterExternalPlugin(pluginName, meta.Type, pluginPath, desc)
}

// configureExternalPlugin applies additional configuration to an external plugin
// (dry-run mode, plugin-specific arguments, etc.)
func configureExternalPlugin(pluginName, pluginType string, dryRun bool, pluginArgs map[string]string, verbose bool) {
	// Set dry-run mode if applicable.
	if err := setPluginDryRun(sharedPluginManager, pluginName, pluginType, dryRun); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, " Failed to set dry-run for plugin '%s': %v\n", pluginName, err)
		}
	}

	// Set plugin args if provided.
	if argsJSON, ok := pluginArgs[pluginName]; ok {
		if err := setPluginArgs(sharedPluginManager, pluginName, pluginType, argsJSON); err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, " Failed to set args for plugin '%s': %v\n", pluginName, err)
			}
		}
	}
}
