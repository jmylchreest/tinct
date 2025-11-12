// Package cli provides the command-line interface for Tinct.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/plugin/repository"
)

const (
	// PluginLockFile is the name of the plugin lock file.
	PluginLockFile = ".tinct-plugins.json"
)

// PluginLock represents the plugin lock file structure.
type PluginLock struct {
	// Version of the lock file format.
	Version string `json:"version,omitempty"`

	// EnabledPlugins is a list of explicitly enabled plugins.
	EnabledPlugins []string `json:"enabled_plugins,omitempty"`

	// DisabledPlugins is a list of explicitly disabled plugins.
	DisabledPlugins []string `json:"disabled_plugins,omitempty"`

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

	// InstalledAt is the timestamp when the plugin was installed.
	InstalledAt string `json:"installed_at,omitempty"`

	// Config holds plugin-specific configuration (optional).
	Config map[string]any `json:"config,omitempty"`
}

var (
	// Plugin command flags.
	pluginLockPath   string
	pluginType       string
	pluginForce      bool
	pluginClear      bool
	pluginYes        bool
	pluginSourceType string
	pluginNoCopy     bool
	pluginShowPath   bool
)

// pluginsCmd represents the plugins command.
var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Manage plugins",
	Long: `Manage Tinct plugins including listing, enabling, disabling, and managing external plugins.

Plugins can be controlled via:
  1. Plugin lock file (.tinct-plugins.json)
  2. Environment variables (TINCT_ENABLED_PLUGINS, TINCT_DISABLED_PLUGINS)
  3. Default plugin settings

Priority order: lock file > environment variables > plugin defaults

When TINCT_ENABLED_PLUGINS is set, only those plugins are enabled (whitelist mode).
When TINCT_DISABLED_PLUGINS is set, those plugins are disabled (blacklist mode).

Commands that modify the lock file:
  - add: Adds external plugin and updates lock file
  - delete: Removes external plugin and updates lock file
  - enable: Updates lock file to enable plugin
  - disable: Updates lock file to disable plugin
  - clear: Updates lock file to clear plugin configuration`,
}

// pluginListCmd lists all available plugins.
var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available plugins",
	Long: `List all available plugins including their enabled/disabled state.

Shows both built-in and external plugins with their type and description.`,
	RunE: runPluginList,
}

// pluginEnableCmd enables a plugin.
var pluginEnableCmd = &cobra.Command{
	Use:   "enable <plugin-name>",
	Short: "Enable a plugin",
	Long: `Enable a plugin by adding it to the plugin lock file.

Examples:
  tinct plugins enable hyprland
  tinct plugins enable waybar
  tinct plugins enable image
  tinct plugins enable all
  tinct plugins enable hyprland --clear  # Remove from disabled list only`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginEnable,
}

// pluginDisableCmd disables a plugin.
var pluginDisableCmd = &cobra.Command{
	Use:   "disable <plugin-name>",
	Short: "Disable a plugin",
	Long: `Disable a plugin by adding it to the disabled list in the plugin lock file.

Examples:
  tinct plugins disable hyprland
  tinct plugins disable waybar
  tinct plugins disable image
  tinct plugins disable all
  tinct plugins disable hyprland --clear  # Remove from enabled list only`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginDisable,
}

// pluginClearCmd clears plugin configuration.
var pluginClearCmd = &cobra.Command{
	Use:   "clear [plugin-name]",
	Short: "Clear plugin configuration",
	Long: `Clear plugin enabled/disabled status, returning it to default behavior.

If a plugin name is provided, clears that plugin's configuration.
If no plugin name is provided, clears all plugin configuration.

Examples:
  tinct plugins clear hyprland  # Clear hyprland config
  tinct plugins clear waybar    # Clear waybar config
  tinct plugins clear           # Clear all plugin config`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPluginClear,
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
  tinct plugins update --lock-file /path/to/.tinct-plugins.json`,
	RunE: runPluginUpdate,
}

func init() {
	// Add plugins command flags.
	pluginsCmd.PersistentFlags().StringVar(&pluginLockPath, "lock-file", "", "path to plugin lock file (default: .tinct-plugins.json in current or home directory)")

	// Add type flag to relevant commands (no shorthand to avoid conflict with global -t theme flag).
	pluginEnableCmd.Flags().StringVar(&pluginType, "type", "", "plugin type (input or output)")
	pluginDisableCmd.Flags().StringVar(&pluginType, "type", "", "plugin type (input or output)")
	pluginListCmd.Flags().BoolVar(&pluginShowPath, "show-path", false, "show the actual file path used when loading each plugin")
	pluginAddCmd.Flags().StringVar(&pluginType, "type", "output", "plugin type (input or output)")
	pluginAddCmd.Flags().BoolVarP(&pluginForce, "force", "f", false, "force overwrite if plugin already exists")
	pluginAddCmd.Flags().StringVar(&pluginSourceType, "source-type", "", "force source type (local, http, git) - auto-detected if not specified")
	pluginAddCmd.Flags().BoolVar(&pluginNoCopy, "no-copy", false, "register plugin at its current location without copying (useful for system packages)")
	pluginDeleteCmd.Flags().BoolVarP(&pluginForce, "force", "f", false, "force deletion without confirmation")

	// Add subcommands.
	pluginsCmd.AddCommand(pluginListCmd)
	pluginsCmd.AddCommand(pluginEnableCmd)
	pluginsCmd.AddCommand(pluginDisableCmd)
	pluginsCmd.AddCommand(pluginClearCmd)
	pluginsCmd.AddCommand(pluginAddCmd)
	pluginsCmd.AddCommand(pluginDeleteCmd)
	pluginsCmd.AddCommand(pluginUpdateCmd)

	// Add flags.
	pluginEnableCmd.Flags().BoolVarP(&pluginClear, "clear", "c", false, "Only remove from disabled list (don't add to enabled)")
	pluginDisableCmd.Flags().BoolVarP(&pluginClear, "clear", "c", false, "Only remove from enabled list (don't add to disabled)")
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

	// Ensure plugin directory exists.
	if err := os.MkdirAll(pluginDir, 0o755); err != nil { // #nosec G301 - Plugin directory needs standard permissions
		return fmt.Errorf("failed to create plugin directory: %w", err)
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
	lock.ExternalPlugins[pluginInfo.Name] = &ExternalPluginMeta{
		Name:         pluginInfo.Name,
		Path:         finalPath,
		Type:         pluginInfo.Type,
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

	// Parse plugin name to support both "name" and "type:name" formats
	_, parsedName := parsePluginName(pluginName)
	if parsedName == "" {
		parsedName = pluginName
	}

	// Load plugin lock.
	lock, lockPath, err := loadPluginLock()
	if err != nil {
		return fmt.Errorf("failed to load plugin lock: %w", err)
	}

	if lock == nil || lock.ExternalPlugins == nil {
		return fmt.Errorf("no external plugins found")
	}

	// Check if plugin exists (using parsed name)
	meta, exists := lock.ExternalPlugins[parsedName]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", parsedName)
	}

	// Confirm deletion if not forced.
	if !pluginForce {
		fmt.Printf("Are you sure you want to delete plugin '%s'? (y/N): ", parsedName)
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
	delete(lock.ExternalPlugins, parsedName)

	// Remove from enabled/disabled lists.
	fullName := fmt.Sprintf("%s:%s", meta.Type, parsedName)
	lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, parsedName, fullName)
	lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, parsedName, fullName)

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	fmt.Printf("Plugin '%s' deleted successfully\n", parsedName)
	return nil
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
		fmt.Fprintf(os.Stderr, "Plugin directory: %s\n\n", pluginDir)
	}

	// Ensure plugin directory exists.
	if err := os.MkdirAll(pluginDir, 0o755); err != nil { // #nosec G301 - Plugin directory needs standard permissions
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Update each external plugin.
	successCount := 0
	failCount := 0

	pluginNames := make([]string, 0, len(lock.ExternalPlugins))
	for name := range lock.ExternalPlugins {
		pluginNames = append(pluginNames, name)
	}
	sort.Strings(pluginNames)

	for _, name := range pluginNames {
		meta := lock.ExternalPlugins[name]
		sourceStr := ""
		if meta.Source != nil {
			sourceStr = formatPluginSourceString(meta.Source)
		} else if meta.SourceLegacy != "" {
			sourceStr = meta.SourceLegacy
		}
		fmt.Printf("Updating plugin '%s' from %s...\n", name, sourceStr)

		// Install plugin from source.
		pluginPath, err := installPluginFromSource(sourceStr, name, pluginDir, "", verbose)
		if err != nil {
			fmt.Printf("   %v\n", err)
			failCount++
			continue
		}

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

		fmt.Printf("   Updated: %s\n", pluginPath)
		successCount++
	}

	// Save updated lock file.
	if successCount > 0 {
		if err := savePluginLock(lockPath, lock); err != nil {
			return fmt.Errorf("failed to save plugin lock: %w", err)
		}
	}

	// Summary.
	fmt.Printf("\nUpdate complete: %d succeeded, %d failed\n", successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("some plugins failed to update")
	}

	return nil
}

// loadPluginLock loads the plugin lock file.
func loadPluginLock() (*PluginLock, string, error) {
	lockPath := pluginLockPath

	if lockPath == "" {
		// Try current directory first.
		lockPath = PluginLockFile
		if _, err := os.Stat(lockPath); os.IsNotExist(err) {
			// Try home directory.
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, "", fmt.Errorf("no plugin lock file found")
			}

			homeLockPath := filepath.Join(home, PluginLockFile)
			if _, err := os.Stat(homeLockPath); err != nil {
				return nil, "", fmt.Errorf("no plugin lock file found")
			}
			lockPath = homeLockPath
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

	// Create new lock file.
	lockPath = pluginLockPath
	if lockPath == "" {
		lockPath = PluginLockFile
	}

	lock = &PluginLock{
		EnabledPlugins:  []string{},
		DisabledPlugins: []string{},
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
	if lock == nil {
		return manager.NewBuilder().WithEnvConfig().Build()
	}

	config := manager.Config{
		EnabledPlugins:  lock.EnabledPlugins,
		DisabledPlugins: lock.DisabledPlugins,
	}

	mgr := manager.NewBuilder().WithConfig(config).Build()

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

// loadAndApplyPluginLock loads the plugin lock file and applies configuration
// to the shared plugin manager. This is used by commands that need to respect
// plugin enable/disable settings without creating a new manager.
func loadAndApplyPluginLock() error {
	lock, _, err := loadPluginLock()
	if err != nil {
		return err
	}

	if lock != nil {
		config := manager.Config{
			EnabledPlugins:  lock.EnabledPlugins,
			DisabledPlugins: lock.DisabledPlugins,
		}
		sharedPluginManager.UpdateConfig(config)
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
