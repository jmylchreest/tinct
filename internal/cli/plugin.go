// Package cli provides the command-line interface for Tinct.
package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/plugin/repository"
	"github.com/jmylchreest/tinct/internal/security"
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
	ExternalPlugins map[string]ExternalPluginMeta `json:"external_plugins,omitempty"`
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
}

var (
	// Plugin command flags.
	pluginLockPath string
	pluginType     string
	pluginForce    bool
	pluginClear    bool
	pluginVerify   bool
	pluginYes      bool
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
When TINCT_DISABLED_PLUGINS is set, those plugins are disabled (blacklist mode).`,
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

Format: plugin-name or type:plugin-name
Examples:
  tinct plugins enable tailwind
  tinct plugins enable output:tailwind
  tinct plugins enable input:image
  tinct plugins enable all
  tinct plugins enable tailwind --clear  # Remove from disabled list only`,
	Args: cobra.ExactArgs(1),
	RunE: runPluginEnable,
}

// pluginDisableCmd disables a plugin.
var pluginDisableCmd = &cobra.Command{
	Use:   "disable <plugin-name>",
	Short: "Disable a plugin",
	Long: `Disable a plugin by adding it to the disabled list in the plugin lock file.

Format: plugin-name or type:plugin-name
Examples:
  tinct plugins disable tailwind
  tinct plugins disable output:tailwind
  tinct plugins disable input:image
  tinct plugins disable all
  tinct plugins disable tailwind --clear  # Remove from enabled list only`,
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
  tinct plugins clear tailwind        # Clear tailwind config
  tinct plugins clear output:hyprland # Clear hyprland config
  tinct plugins clear                 # Clear all plugin config`,
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

Examples:
  tinct plugins add ./contrib/notify-send.py
  tinct plugins add https://example.com/plugins/theme.sh
  tinct plugins add https://github.com/user/plugin.git
  tinct plugins add https://github.com/user/plugin.git:path/to/plugin.sh`,
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
	pluginAddCmd.Flags().StringVar(&pluginType, "type", "output", "plugin type (input or output)")
	pluginAddCmd.Flags().BoolVarP(&pluginForce, "force", "f", false, "force overwrite if plugin already exists")
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
func runPluginList(cmd *cobra.Command, args []string) error {
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

	// Build list of all plugins with type prefix.
	type pluginInfo struct {
		fullName    string
		status      string
		version     string
		description string
		isExternal  bool
		source      string
	}

	var allPlugins []pluginInfo
	seenPlugins := make(map[string]bool)

	// Helper to determine plugin status
	getPluginStatus := func(pluginType, pluginName string) string {
		fullName := fmt.Sprintf("%s:%s", pluginType, pluginName)

		// Check if explicitly disabled
		if lock != nil {
			for _, disabled := range lock.DisabledPlugins {
				if disabled == fullName || disabled == pluginName || disabled == "all" {
					return "disabled"
				}
			}
		}

		// Check if explicitly enabled
		if lock != nil && len(lock.EnabledPlugins) > 0 {
			for _, enabled := range lock.EnabledPlugins {
				if enabled == fullName || enabled == pluginName || enabled == "all" {
					return "enabled"
				}
			}
			// If enabled list exists but plugin not in it, it's on-demand
			return "on-demand"
		}

		// No config = on-demand
		return "on-demand"
	}

	// Add input plugins.
	inputPlugins := mgr.AllInputPlugins()
	for name, plugin := range inputPlugins {
		status := getPluginStatus("input", name)
		fullName := fmt.Sprintf("input:%s", name)
		// Check if this is an external plugin by comparing names.
		isExternal := false
		if lock != nil && lock.ExternalPlugins != nil {
			for _, meta := range lock.ExternalPlugins {
				if meta.Name == name && meta.Type == "input" {
					isExternal = true
					break
				}
			}
		}
		// Get source for external plugins.
		pluginSource := ""
		if isExternal && lock != nil && lock.ExternalPlugins != nil {
			for _, meta := range lock.ExternalPlugins {
				if meta.Name == name && meta.Type == "input" {
					if meta.Source != nil {
						pluginSource = formatPluginSourceString(meta.Source)
					} else if meta.SourceLegacy != "" {
						pluginSource = meta.SourceLegacy
					}
					break
				}
			}
		}

		allPlugins = append(allPlugins, pluginInfo{
			fullName:    fullName,
			status:      status,
			version:     plugin.Version(),
			description: plugin.Description(),
			isExternal:  isExternal,
			source:      pluginSource,
		})
		seenPlugins[fullName] = true
	}

	// Add output plugins.
	outputPlugins := mgr.AllOutputPlugins()
	for name, plugin := range outputPlugins {
		status := getPluginStatus("output", name)
		fullName := fmt.Sprintf("output:%s", name)
		// Check if this is an external plugin by comparing names.
		isExternal := false
		if lock != nil && lock.ExternalPlugins != nil {
			for _, meta := range lock.ExternalPlugins {
				if meta.Name == name && meta.Type == "output" {
					isExternal = true
					break
				}
			}
		}
		// Get source for external plugins.
		pluginSource := ""
		if isExternal && lock != nil && lock.ExternalPlugins != nil {
			for _, meta := range lock.ExternalPlugins {
				if meta.Name == name && meta.Type == "output" {
					if meta.Source != nil {
						pluginSource = formatPluginSourceString(meta.Source)
					} else if meta.SourceLegacy != "" {
						pluginSource = meta.SourceLegacy
					}
					break
				}
			}
		}

		allPlugins = append(allPlugins, pluginInfo{
			fullName:    fullName,
			status:      status,
			version:     plugin.Version(),
			description: plugin.Description(),
			isExternal:  isExternal,
			source:      pluginSource,
		})
		seenPlugins[fullName] = true
	}

	// Add external plugins that aren't in the manager.
	if lock != nil && lock.ExternalPlugins != nil {
		for lockKey, meta := range lock.ExternalPlugins {
			// Use the plugin's actual name, not the lock file key.
			pluginName := meta.Name
			if pluginName == "" {
				pluginName = lockKey // Fallback to lock key if name is missing
			}

			fullName := fmt.Sprintf("%s:%s", meta.Type, pluginName)
			if seenPlugins[fullName] {
				continue // Already added from manager
			}

			// Determine status for external-only plugins.
			status := getPluginStatus(meta.Type, pluginName)

			// Use plugin's description if available, otherwise show source.
			description := meta.Description
			if description == "" {
				sourceStr := ""
				if meta.Source != nil {
					sourceStr = formatPluginSourceString(meta.Source)
				} else if meta.SourceLegacy != "" {
					sourceStr = meta.SourceLegacy
				}
				description = fmt.Sprintf("External plugin (source: %s)", sourceStr)
			}

			// Get version from meta if available
			version := meta.Version
			if version == "" {
				version = "unknown"
			}

			allPlugins = append(allPlugins, pluginInfo{
				fullName:    fullName,
				status:      status,
				version:     version,
				description: description,
				isExternal:  true,
				source:      formatPluginSourceString(meta.Source),
			})
		}
	}

	// Sort by full name.
	sort.Slice(allPlugins, func(i, j int) bool {
		return allPlugins[i].fullName < allPlugins[j].fullName
	})

	// Display plugins using table formatter.
	tbl := NewTable([]string{"", "PLUGIN", "STATUS", "VERSION", "DESCRIPTION"})

	for _, p := range allPlugins {
		marker := ""
		if p.isExternal {
			marker = "*"
		}
		tbl.AddRow([]string{marker, p.fullName, p.status, p.version, p.description})
		// Show source for external plugins.
		if p.isExternal && p.source != "" {
			tbl.AddRow([]string{"", "", "", "", "  src: " + p.source})
		}
	}

	fmt.Print(tbl.Render())

	// Add legend if there are external plugins.
	hasExternal := false
	for _, p := range allPlugins {
		if p.isExternal {
			hasExternal = true
			break
		}
	}
	if hasExternal {
		fmt.Println()
		fmt.Println("* = external plugin")
	}

	return nil
}

// runPluginEnable enables a plugin.
func runPluginEnable(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load or create plugin lock.
	lock, lockPath, err := loadOrCreatePluginLock()
	if err != nil {
		return fmt.Errorf("failed to load plugin lock: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n", lockPath)
	}

	// Handle "all" pseudo-plugin.
	if pluginName == "all" {
		if pluginClear {
			// Just remove "all" from disabled list.
			lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, "all", "all")
		} else {
			// Clear disabled list.
			lock.DisabledPlugins = []string{}
			// Add "all" to enabled list.
			lock.EnabledPlugins = []string{"all"}
		}

		if err := savePluginLock(lockPath, lock); err != nil {
			return fmt.Errorf("failed to save plugin lock: %w", err)
		}

		if pluginClear {
			fmt.Println("Cleared 'all' from disabled list")
		} else {
			fmt.Println("All plugins enabled")
		}
		return nil
	}

	// Parse plugin name.
	parsedType, parsedName := parsePluginName(pluginName)
	if pluginType != "" {
		parsedType = pluginType
	}

	// Format full plugin name.
	fullName := fmt.Sprintf("%s:%s", parsedType, parsedName)

	if pluginClear {
		// Just remove from disabled list.
		lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, parsedName, fullName)
	} else {
		// Remove from disabled list.
		lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, parsedName, fullName)

		// Add to enabled list if not already there.
		if !containsPlugin(lock.EnabledPlugins, parsedName, fullName) {
			lock.EnabledPlugins = append(lock.EnabledPlugins, fullName)
		}
	}

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	if pluginClear {
		fmt.Printf("Cleared '%s' from disabled list\n", fullName)
	} else {
		fmt.Printf("Plugin '%s' enabled\n", fullName)
	}
	return nil
}

// runPluginDisable disables a plugin.
func runPluginDisable(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load or create plugin lock.
	lock, lockPath, err := loadOrCreatePluginLock()
	if err != nil {
		return fmt.Errorf("failed to load plugin lock: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n", lockPath)
	}

	// Handle "all" pseudo-plugin.
	if pluginName == "all" {
		if pluginClear {
			// Just remove "all" from enabled list.
			lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, "all", "all")
		} else {
			// Clear enabled list.
			lock.EnabledPlugins = []string{}
			// Add "all" to disabled list.
			lock.DisabledPlugins = []string{"all"}
		}

		if err := savePluginLock(lockPath, lock); err != nil {
			return fmt.Errorf("failed to save plugin lock: %w", err)
		}

		if pluginClear {
			fmt.Println("Cleared 'all' from enabled list")
		} else {
			fmt.Println("All plugins disabled")
		}
		return nil
	}

	// Parse plugin name.
	parsedType, parsedName := parsePluginName(pluginName)
	if pluginType != "" {
		parsedType = pluginType
	}

	// Format full plugin name.
	fullName := fmt.Sprintf("%s:%s", parsedType, parsedName)

	if pluginClear {
		// Just remove from enabled list.
		lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, parsedName, fullName)
	} else {
		// Remove from enabled list.
		lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, parsedName, fullName)

		// Add to disabled list if not already there.
		if !containsPlugin(lock.DisabledPlugins, parsedName, fullName) {
			lock.DisabledPlugins = append(lock.DisabledPlugins, fullName)
		}
	}

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	if pluginClear {
		fmt.Printf("Cleared '%s' from enabled list\n", fullName)
	} else {
		fmt.Printf("Plugin '%s' disabled\n", fullName)
	}
	return nil
}

// runPluginClear clears plugin configuration.
func runPluginClear(cmd *cobra.Command, args []string) error {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load or create plugin lock.
	lock, lockPath, err := loadOrCreatePluginLock()
	if err != nil {
		return fmt.Errorf("failed to load plugin lock: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n", lockPath)
	}

	// If no plugin name provided, clear all.
	if len(args) == 0 {
		lock.EnabledPlugins = []string{}
		lock.DisabledPlugins = []string{}

		if err := savePluginLock(lockPath, lock); err != nil {
			return fmt.Errorf("failed to save plugin lock: %w", err)
		}

		fmt.Println("Cleared all plugin configuration")
		return nil
	}

	pluginName := args[0]

	// Handle "all" pseudo-plugin.
	if pluginName == "all" {
		lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, "all", "all")
		lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, "all", "all")

		if err := savePluginLock(lockPath, lock); err != nil {
			return fmt.Errorf("failed to save plugin lock: %w", err)
		}

		fmt.Println("Cleared 'all' configuration")
		return nil
	}

	// Parse plugin name.
	parsedType, parsedName := parsePluginName(pluginName)
	if pluginType != "" {
		parsedType = pluginType
	}

	// Format full plugin name.
	fullName := fmt.Sprintf("%s:%s", parsedType, parsedName)

	// Remove from both lists.
	lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, parsedName, fullName)
	lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, parsedName, fullName)

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	fmt.Printf("Cleared configuration for '%s'\n", fullName)
	return nil
}

// runPluginAdd adds an external plugin.
func runPluginAdd(cmd *cobra.Command, args []string) error {
	source := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load or create plugin lock.
	lock, lockPath, err := loadOrCreatePluginLock()
	if err != nil {
		return fmt.Errorf("failed to load plugin lock: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n", lockPath)
	}

	// Initialize external plugins map if needed.
	if lock.ExternalPlugins == nil {
		lock.ExternalPlugins = make(map[string]ExternalPluginMeta)
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

	// Determine plugin path based on source type.
	pluginPath, err := installPluginFromSource(source, "", pluginDir, verbose)
	if err != nil {
		return err
	}

	// Query plugin for its actual name and metadata.
	pluginName, pluginDescription, pluginType, version := queryPluginMetadata(pluginPath)
	if pluginName == "" {
		return fmt.Errorf("failed to query plugin name from --plugin-info")
	}
	if pluginType == "" {
		pluginType = "output" // default to output if not specified
	}

	// Check if plugin already exists.
	if _, exists := lock.ExternalPlugins[pluginName]; exists && !pluginForce {
		return fmt.Errorf("plugin '%s' already exists (use --force to overwrite)", pluginName)
	}

	// Add to lock file.
	lock.ExternalPlugins[pluginName] = ExternalPluginMeta{
		Name:         pluginName,
		Path:         pluginPath,
		Type:         pluginType,
		SourceLegacy: source,
		Version:      version,
		Description:  pluginDescription,
	}

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	fmt.Printf("Plugin '%s' added successfully\n", pluginName)
	if pluginDescription != "" {
		fmt.Printf("Description: %s\n", pluginDescription)
	}
	fmt.Printf("Type: %s\n", pluginType)
	if version != "" {
		fmt.Printf("Version: %s\n", version)
	}
	fmt.Printf("Path: %s\n", pluginPath)
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

	// Delete plugin file.
	if err := os.Remove(meta.Path); err != nil && !os.IsNotExist(err) {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete plugin file: %v\n", err)
		}
	} else if verbose {
		fmt.Fprintf(os.Stderr, "Deleted plugin file: %s\n", meta.Path)
	}

	// Remove from lock file.
	delete(lock.ExternalPlugins, pluginName)

	// Remove from enabled/disabled lists.
	fullName := fmt.Sprintf("%s:%s", meta.Type, pluginName)
	lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, pluginName, fullName)
	lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, pluginName, fullName)

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	fmt.Printf("Plugin '%s' deleted successfully\n", pluginName)
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
		pluginPath, err := installPluginFromSource(sourceStr, name, pluginDir, verbose)
		if err != nil {
			fmt.Printf("   %v\n", err)
			failCount++
			continue
		}

		// Query plugin for updated metadata.
		actualName, pluginDescription, pluginType, version := queryPluginMetadata(pluginPath)
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
		lock.ExternalPlugins[name] = ExternalPluginMeta{
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
			if home, err := os.UserHomeDir(); err == nil {
				homeLockPath := filepath.Join(home, PluginLockFile)
				if _, err := os.Stat(homeLockPath); err == nil {
					lockPath = homeLockPath
				} else {
					return nil, "", fmt.Errorf("no plugin lock file found")
				}
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
func loadOrCreatePluginLock() (*PluginLock, string, error) {
	lock, lockPath, err := loadPluginLock()
	if err == nil {
		return lock, lockPath, nil
	}

	// Create new lock file.
	lockPath = pluginLockPath
	if lockPath == "" {
		lockPath = PluginLockFile
	}

	lock = &PluginLock{
		EnabledPlugins:  []string{},
		DisabledPlugins: []string{},
		ExternalPlugins: make(map[string]ExternalPluginMeta),
	}

	return lock, lockPath, nil
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
				pluginName, _, _, _ = queryPluginMetadata(meta.Path)
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

// queryPluginMetadata queries a plugin for its name, description, type, and version.
func queryPluginMetadata(pluginPath string) (name, description, pluginType, version string) {
	cmd := exec.Command(pluginPath, "--plugin-info")
	output, err := cmd.Output()
	if err != nil {
		return "", "", "", ""
	}

	var info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Version     string `json:"version"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return "", "", "", ""
	}

	return info.Name, info.Description, info.Type, info.Version
}

// parsePluginName parses a plugin name into type and name.
func parsePluginName(name string) (pluginType, pluginName string) {
	parts := strings.Split(name, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", name
}

// containsPlugin checks if a plugin is in a list.
func containsPlugin(list []string, name, fullName string) bool {
	for _, item := range list {
		if item == name || item == fullName {
			return true
		}
	}
	return false
}

// removeFromList removes a plugin from a list.
func removeFromList(list []string, name, fullName string) []string {
	result := make([]string, 0, len(list))
	for _, item := range list {
		if item != name && item != fullName {
			result = append(result, item)
		}
	}
	return result
}

// getPluginDir returns the plugin directory path.
func getPluginDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "tinct", "plugins"), nil
}

// formatPluginSourceString converts a PluginSource struct to a display string.
func formatPluginSourceString(source *repository.PluginSource) string {
	if source == nil {
		return ""
	}
	switch source.Type {
	case "repository":
		return fmt.Sprintf("repo:%s/%s@%s", source.Repository, source.Plugin, source.Version)
	case "http":
		return source.URL
	case "local":
		return source.OriginalPath
	default:
		return source.Type
	}
}

// installPluginFromSource installs a plugin from various source types.
func installPluginFromSource(source, pluginName, pluginDir string, verbose bool) (string, error) {
	// Parse source to determine type.
	sourceType, sourceInfo := parsePluginSource(source)

	switch sourceType {
	case "local":
		return installFromLocal(sourceInfo, pluginDir, verbose)
	case "http":
		return installFromHTTP(sourceInfo, pluginName, pluginDir, verbose)
	case "git":
		return installFromGit(sourceInfo, pluginName, pluginDir, verbose)
	default:
		return "", fmt.Errorf("unsupported source type: %s", source)
	}
}

// PluginSourceType represents the type of plugin source.
type PluginSourceInfo struct {
	URL      string
	FilePath string // For git repos, path to file within repo
	Ref      string // For git repos, branch/tag/commit
}

// parsePluginSource determines the source type and extracts relevant info.
func parsePluginSource(source string) (string, PluginSourceInfo) {
	info := PluginSourceInfo{}

	// Git repository (https://github.com/user/repo.git or git@github.com:user/repo.git).
	if strings.HasSuffix(source, ".git") || strings.Contains(source, "github.com") || strings.Contains(source, "gitlab.com") || strings.Contains(source, "bitbucket.org") {
		// Check for file path specification: repo.git:path/to/file.sh.
		if idx := strings.LastIndex(source, ":"); idx > 0 && !strings.HasPrefix(source, "git@") {
			// Make sure it's not the : in git@github.com.
			if idx > 6 && source[idx-1] != 'm' { // Not ending in ".com:"
				info.URL = source[:idx]
				info.FilePath = source[idx+1:]
			} else {
				info.URL = source
			}
		} else {
			info.URL = source
		}
		return "git", info
	}

	// HTTP/HTTPS URL.
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Check for file path specification: url.tar.gz:path/to/plugin.
		if idx := strings.LastIndex(source, ":"); idx > 0 {
			// Check if it's part of the protocol (http:// or https://).
			if idx > 7 && source[idx-2:idx] != "tp" && source[idx-3:idx] != "tps" {
				info.URL = source[:idx]
				info.FilePath = source[idx+1:]
			} else {
				info.URL = source
			}
		} else {
			info.URL = source
		}
		return "http", info
	}

	// Local file.
	info.FilePath = source
	return "local", info
}

// installFromLocal installs a plugin from a local file.
func installFromLocal(info PluginSourceInfo, pluginDir string, verbose bool) (string, error) {
	absSource, err := filepath.Abs(info.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve plugin path: %w", err)
	}

	// Verify plugin exists.
	if _, err := os.Stat(absSource); err != nil {
		return "", fmt.Errorf("plugin file not found: %w", err)
	}

	// Copy plugin to plugin directory.
	destPath := filepath.Join(pluginDir, filepath.Base(absSource))
	if err := copyFile(absSource, destPath); err != nil {
		return "", fmt.Errorf("failed to copy plugin: %w", err)
	}

	// Make it executable.
	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Copied plugin to: %s\n", destPath)
	}

	return destPath, nil
}

// installFromHTTP downloads a plugin from an HTTP/HTTPS URL.
func installFromHTTP(info PluginSourceInfo, pluginName, pluginDir string, verbose bool) (string, error) {
	if verbose {
		fmt.Fprintf(os.Stderr, "Downloading from %s...\n", info.URL)
	}

	// Download the file.
	resp, err := http.Get(info.URL)
	if err != nil {
		return "", fmt.Errorf("failed to download plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download plugin: HTTP %d", resp.StatusCode)
	}

	// Read the entire response into memory for archive detection.
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read download: %w", err)
	}

	// Determine filename from URL.
	filename := filepath.Base(info.URL)
	if filename == "" || filename == "." {
		filename = pluginName
	}

	// Check if it's an archive.
	if strings.HasSuffix(info.URL, ".tar.gz") || strings.HasSuffix(info.URL, ".tgz") {
		// Extract from tar.gz archive.
		return extractFromTarGz(data, info.FilePath, pluginDir, verbose)
	} else if strings.HasSuffix(info.URL, ".zip") {
		// Extract from zip archive.
		return extractFromZip(data, info.FilePath, pluginDir, verbose)
	}

	// Not an archive - treat as direct plugin file.
	destPath := filepath.Join(pluginDir, filename)

	// Write file.
	// #nosec G306 -- Plugin executable needs exec permissions.
	if err := os.WriteFile(destPath, data, 0o755); err != nil {
		return "", fmt.Errorf("failed to write plugin file: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Downloaded plugin to: %s\n", destPath)
	}

	return destPath, nil
}

// extractFromTarGz extracts a plugin from a tar.gz archive.
func extractFromTarGz(data []byte, targetFile, pluginDir string, verbose bool) (string, error) {
	// Create gzip reader.
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader.
	tr := tar.NewReader(gzr)

	// If no specific file requested, find the first executable or use first file.
	var targetPath string
	foundFiles := []string{}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar archive: %w", err)
		}

		// Skip directories.
		if header.Typeflag == tar.TypeDir {
			continue
		}

		foundFiles = append(foundFiles, header.Name)

		// Check if this is the file we want.
		if targetFile != "" {
			if header.Name == targetFile || strings.HasSuffix(header.Name, "/"+targetFile) {
				targetPath = header.Name
				break
			}
		} else {
			// Auto-detect: prefer executable files.
			if header.FileInfo().Mode()&0o111 != 0 {
				targetPath = header.Name
				break
			}
		}
	}

	// If we didn't find the target, reset and look for any match.
	if targetPath == "" && targetFile != "" {
		return "", fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	}

	// If still no target and we have files, use the first one.
	if targetPath == "" && len(foundFiles) > 0 {
		targetPath = foundFiles[0]
	}

	if targetPath == "" {
		return "", fmt.Errorf("no files found in archive")
	}

	// Reset readers to extract the target file.
	gzr, err = gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr = tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return "", fmt.Errorf("file not found in archive")
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Name == targetPath {
			// Extract the file.
			destPath := filepath.Join(pluginDir, filepath.Base(targetPath))

			out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
			if err != nil {
				return "", fmt.Errorf("failed to create plugin file: %w", err)
			}
			defer out.Close()

			// Limit decompression size to prevent zip bombs (100MB limit for plugins).
			limitedReader := security.NewLimitedReader(tr, 100*1024*1024)
			if _, err := io.Copy(out, limitedReader); err != nil {
				return "", fmt.Errorf("failed to extract plugin: %w", err)
			}

			// Make executable.
			if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
				return "", fmt.Errorf("failed to make plugin executable: %w", err)
			}

			if verbose {
				fmt.Fprintf(os.Stderr, "Extracted plugin to: %s\n", destPath)
			}

			return destPath, nil
		}
	}
}

// extractFromZip extracts a plugin from a zip archive.
func extractFromZip(data []byte, targetFile, pluginDir string, verbose bool) (string, error) {
	// Create zip reader.
	reader := bytes.NewReader(data)
	zr, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to create zip reader: %w", err)
	}

	// If no specific file requested, find the first executable or use first file.
	var targetZipFile *zip.File
	foundFiles := []string{}

	for _, f := range zr.File {
		// Skip directories.
		if f.FileInfo().IsDir() {
			continue
		}

		foundFiles = append(foundFiles, f.Name)

		// Check if this is the file we want.
		if targetFile != "" {
			if f.Name == targetFile || strings.HasSuffix(f.Name, "/"+targetFile) {
				targetZipFile = f
				break
			}
		} else {
			// Auto-detect: prefer executable files.
			if f.FileInfo().Mode()&0o111 != 0 {
				targetZipFile = f
				break
			}
		}
	}

	// If we didn't find the target, check if any file matches.
	if targetZipFile == nil && targetFile != "" {
		return "", fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	}

	// If still no target and we have files, use the first one.
	if targetZipFile == nil && len(foundFiles) > 0 {
		targetZipFile = zr.File[0]
	}

	if targetZipFile == nil {
		return "", fmt.Errorf("no files found in archive")
	}

	// Extract the file.
	destPath := filepath.Join(pluginDir, filepath.Base(targetZipFile.Name))

	rc, err := targetZipFile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file in archive: %w", err)
	}
	defer rc.Close()

	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return "", fmt.Errorf("failed to create plugin file: %w", err)
	}
	defer out.Close()

	// Limit decompression size to prevent zip bombs (100MB limit for plugins).
	limitedReader := security.NewLimitedReader(rc, 100*1024*1024)
	if _, err := io.Copy(out, limitedReader); err != nil {
		return "", fmt.Errorf("failed to extract plugin: %w", err)
	}

	// Make executable.
	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Extracted plugin to: %s\n", destPath)
	}

	return destPath, nil
}

// installFromGit clones a git repository and extracts the plugin.
func installFromGit(info PluginSourceInfo, pluginName, pluginDir string, verbose bool) (string, error) {
	// Check if git is available.
	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("git is not installed or not in PATH")
	}

	// Validate the git URL for security.
	if err := security.ValidateGitURL(info.URL); err != nil {
		return "", fmt.Errorf("invalid git URL: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Cloning from %s...\n", info.URL)
	}

	// Create temporary directory for cloning.
	tmpDir, err := os.MkdirTemp("", "tinct-plugin-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repository using -- to separate options from arguments (prevents command injection).
	// #nosec G204 -- URL is validated via security.ValidateGitURL above.
	cloneCmd := exec.Command("git", "clone", "--depth", "1", "--", info.URL, tmpDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, string(output))
	}

	// Determine which file to copy.
	var sourceFile string
	if info.FilePath != "" {
		// Specific file path provided.
		sourceFile = filepath.Join(tmpDir, info.FilePath)
		if _, err := os.Stat(sourceFile); err != nil {
			return "", fmt.Errorf("plugin file not found in repository: %s", info.FilePath)
		}
	} else {
		// Try to find a plugin file automatically.
		// Look for common patterns: plugin.sh, *.py, *.sh in root or bin/.
		candidates := []string{
			filepath.Join(tmpDir, pluginName),
			filepath.Join(tmpDir, pluginName+".sh"),
			filepath.Join(tmpDir, pluginName+".py"),
			filepath.Join(tmpDir, "plugin.sh"),
			filepath.Join(tmpDir, "plugin.py"),
			filepath.Join(tmpDir, "bin", pluginName),
			filepath.Join(tmpDir, "bin", pluginName+".sh"),
			filepath.Join(tmpDir, "bin", pluginName+".py"),
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				sourceFile = candidate
				break
			}
		}

		if sourceFile == "" {
			return "", fmt.Errorf("could not find plugin file in repository, please specify path: repo.git:path/to/plugin.sh")
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Found plugin file: %s\n", filepath.Base(sourceFile))
	}

	// Copy plugin to plugin directory.
	destPath := filepath.Join(pluginDir, filepath.Base(sourceFile))
	if err := copyFile(sourceFile, destPath); err != nil {
		return "", fmt.Errorf("failed to copy plugin: %w", err)
	}

	// Make it executable.
	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Installed plugin to: %s\n", destPath)
	}

	return destPath, nil
}
