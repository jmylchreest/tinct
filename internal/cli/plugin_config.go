// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"os"
	"strings"

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

	// Handle "all" pseudo-plugin.
	if pluginName == pluginTypeAll {
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

	if verbose {
		fmt.Fprintf(os.Stderr, "DEBUG: parsed type='%s', name='%s'\n", parsedType, parsedName)
	}

	// Restrict to output plugins only.
	if parsedType != "" && parsedType != "output" {
		return fmt.Errorf("only output plugins can be enabled/disabled (input plugins are always on-demand)")
	}

	// Default to output type if not specified.
	if parsedType == "" {
		parsedType = "output"
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
	lock, lockPath := loadOrCreatePluginLock()

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

	// Restrict to output plugins only.
	if parsedType != "" && parsedType != "output" {
		return fmt.Errorf("only output plugins can be enabled/disabled (input plugins are always on-demand)")
	}

	// Default to output type if not specified.
	if parsedType == "" {
		parsedType = "output"
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

// parsePluginName parses a plugin name into type and name.
func parsePluginName(name string) (pluginType, pluginName string) {
	parts := strings.Split(name, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", name
}
