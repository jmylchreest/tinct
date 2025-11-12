// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
)

// runPluginEnable enables a plugin.
func runPluginEnable(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Load or create plugin lock.
	lock, lockPath := loadOrCreatePluginLock()

	if verbose {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n", lockPath)
	}

	// Check if plugin is an input plugin (only allow output plugins to be enabled/disabled).
	if pluginName != pluginTypeAll {
		if pluginType := getPluginType(lock, pluginName); pluginType == "input" {
			return fmt.Errorf("only output plugins can be enabled/disabled (input plugins are always on-demand)")
		}
	}

	// Handle "all" pseudo-plugin.
	if pluginName == pluginTypeAll {
		if pluginClear {
			// Just remove "all" from disabled list.
			lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, "all")
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

	// Use plugin name directly (no type prefix).
	if pluginClear {
		// Just remove from disabled list.
		lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, pluginName)
	} else {
		// Remove from disabled list.
		lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, pluginName)

		// Add to enabled list if not already there.
		if !containsPlugin(lock.EnabledPlugins, pluginName) {
			lock.EnabledPlugins = append(lock.EnabledPlugins, pluginName)
		}
	}

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	if pluginClear {
		fmt.Printf("Cleared '%s' from disabled list\n", pluginName)
	} else {
		fmt.Printf("Plugin '%s' enabled\n", pluginName)
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
	lock, lockPath := loadOrCreatePluginLock()

	if verbose {
		fmt.Fprintf(os.Stderr, "Using lock file: %s\n", lockPath)
	}

	// Check if plugin is an input plugin (only allow output plugins to be enabled/disabled).
	if pluginName != pluginTypeAll && pluginName != "all" {
		if pluginType := getPluginType(lock, pluginName); pluginType == "input" {
			return fmt.Errorf("only output plugins can be enabled/disabled (input plugins are always on-demand)")
		}
	}

	// Handle "all" pseudo-plugin.
	if pluginName == "all" {
		if pluginClear {
			// Just remove "all" from enabled list.
			lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, "all")
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

	// Use plugin name directly (no type prefix).
	if pluginClear {
		// Just remove from enabled list.
		lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, pluginName)
	} else {
		// Remove from enabled list.
		lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, pluginName)

		// Add to disabled list if not already there.
		if !containsPlugin(lock.DisabledPlugins, pluginName) {
			lock.DisabledPlugins = append(lock.DisabledPlugins, pluginName)
		}
	}

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	if pluginClear {
		fmt.Printf("Cleared '%s' from enabled list\n", pluginName)
	} else {
		fmt.Printf("Plugin '%s' disabled\n", pluginName)
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
	lock, lockPath := loadOrCreatePluginLock()

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
		lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, "all")
		lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, "all")

		if err := savePluginLock(lockPath, lock); err != nil {
			return fmt.Errorf("failed to save plugin lock: %w", err)
		}

		fmt.Println("Cleared 'all' configuration")
		return nil
	}

	// Use plugin name directly (no type prefix).
	// Remove from both lists.
	lock.EnabledPlugins = removeFromList(lock.EnabledPlugins, pluginName)
	lock.DisabledPlugins = removeFromList(lock.DisabledPlugins, pluginName)

	// Save lock file.
	if err := savePluginLock(lockPath, lock); err != nil {
		return fmt.Errorf("failed to save plugin lock: %w", err)
	}

	fmt.Printf("Cleared configuration for '%s'\n", pluginName)
	return nil
}

// containsPlugin checks if a plugin is in a list.
func containsPlugin(list []string, name string) bool {
	return slices.Contains(list, name)
}

// removeFromList removes a plugin from a list.
func removeFromList(list []string, name string) []string {
	result := make([]string, 0, len(list))
	for _, item := range list {
		if item != name {
			result = append(result, item)
		}
	}
	return result
}

// getPluginType returns the type of a plugin by checking the lock file's external plugins.
// Returns "input", "output", or empty string if not found.
func getPluginType(lock *PluginLock, pluginName string) string {
	// Check external plugins first.
	if lock != nil && lock.ExternalPlugins != nil {
		if meta, exists := lock.ExternalPlugins[pluginName]; exists {
			return meta.Type
		}
	}

	// Check built-in plugins via the shared manager.
	mgr := createManagerFromLock(lock)

	// Try to find in input plugins.
	if _, exists := mgr.GetInputPlugin(pluginName); exists {
		return "input"
	}

	// Try to find in output plugins.
	if _, exists := mgr.GetOutputPlugin(pluginName); exists {
		return "output"
	}

	return ""
}
