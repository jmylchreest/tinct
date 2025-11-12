// Package cli provides the command-line interface for Tinct.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/plugin/protocol"
)

// pluginAction represents the type of action being performed on a plugin.
type pluginAction string

const (
	pluginActionAdd       pluginAction = "add"
	pluginActionUpgrade   pluginAction = "upgrade"
	pluginActionDowngrade pluginAction = "downgrade"
	pluginActionOverwrite pluginAction = "overwrite"
)

// pluginMetadata holds comprehensive metadata about a plugin.
type pluginMetadata struct {
	Name            string
	Description     string
	Type            string
	Version         string
	ProtocolVersion string
}

// resolvePluginSource resolves a plugin source path and determines if it's already installed.
// Returns: (sourcePath, isAlreadyInstalled, error).
func resolvePluginSource(source, pluginDir, forcedSourceType string, verbose bool) (string, bool, error) {
	// For local files, resolve to absolute path
	if !strings.HasPrefix(source, "http://") && !strings.HasPrefix(source, "https://") && !strings.HasSuffix(source, ".git") {
		absSource, err := filepath.Abs(source)
		if err != nil {
			return "", false, fmt.Errorf("failed to resolve source path: %w", err)
		}

		// Check if file exists
		if _, err := os.Stat(absSource); err != nil {
			return "", false, fmt.Errorf("source file not found: %w", err)
		}

		// Check if source is already in plugin directory
		absPluginDir, err := filepath.Abs(pluginDir)
		if err != nil {
			return "", false, fmt.Errorf("failed to resolve plugin directory: %w", err)
		}

		// Check if source is within or equal to plugin directory
		sourceDir := filepath.Dir(absSource)
		if sourceDir == absPluginDir {
			return absSource, true, nil
		}

		return absSource, false, nil
	}

	// For remote sources, we need to download first to a temp location
	// to query metadata before installing
	tmpDir, err := os.MkdirTemp("", "tinct-plugin-*")
	if err != nil {
		return "", false, fmt.Errorf("failed to create temp directory: %w", err)
	}
	// Note: Caller is responsible for cleanup if needed

	tmpPath, err := installPluginFromSource(source, "", tmpDir, forcedSourceType, verbose)
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", false, err
	}

	return tmpPath, false, nil
}

// queryFullPluginMetadata queries all metadata from a plugin including protocol version.
func queryFullPluginMetadata(pluginPath string) (*pluginMetadata, error) {
	cmd := exec.Command(pluginPath, "--plugin-info")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute plugin --plugin-info: %w", err)
	}

	var info struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Type            string `json:"type"`
		Version         string `json:"version"`
		ProtocolVersion string `json:"protocol_version"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse plugin info: %w", err)
	}

	if info.Name == "" {
		return nil, fmt.Errorf("plugin did not return a name")
	}

	// Set defaults
	if info.Type == "" {
		info.Type = "output"
	}
	if info.ProtocolVersion == "" {
		info.ProtocolVersion = "1.0.0" // Assume earliest version if not specified
	}

	return &pluginMetadata{
		Name:            info.Name,
		Description:     info.Description,
		Type:            info.Type,
		Version:         info.Version,
		ProtocolVersion: info.ProtocolVersion,
	}, nil
}

// checkProtocolCompatibility verifies the plugin's protocol version is compatible with tinct.
func checkProtocolCompatibility(protocolVersion string, verbose bool) error {
	compatible, err := protocol.IsCompatible(protocolVersion)
	if err != nil {
		return fmt.Errorf("protocol compatibility check failed: %w", err)
	}

	if !compatible {
		return fmt.Errorf("plugin protocol version %s is not compatible with tinct (requires %s, min %s)",
			protocolVersion, protocol.ProtocolVersion, protocol.MinCompatibleVersion)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Protocol version %s is compatible\n", protocolVersion)
	}

	return nil
}

// determinePluginAction determines what action to take based on existing plugin state.
func determinePluginAction(lock *PluginLock, pluginInfo *pluginMetadata, force bool) (pluginAction, *ExternalPluginMeta, error) {
	existingMeta, exists := lock.ExternalPlugins[pluginInfo.Name]
	if !exists {
		return pluginActionAdd, nil, nil
	}

	// Plugin already exists - determine version relationship
	if existingMeta.Version == "" || pluginInfo.Version == "" {
		// Can't determine version relationship - require --force
		if !force {
			return "", existingMeta, fmt.Errorf("plugin '%s' already exists (use --force to overwrite)", pluginInfo.Name)
		}
		return pluginActionOverwrite, existingMeta, nil
	}

	cmp, err := compareVersions(pluginInfo.Version, existingMeta.Version)
	if err != nil {
		// Invalid version format - require --force
		if !force {
			return "", existingMeta, fmt.Errorf("plugin '%s' already exists with unparseable version (use --force to overwrite)", pluginInfo.Name)
		}
		return pluginActionOverwrite, existingMeta, nil
	}

	switch {
	case cmp > 0:
		// Upgrade: newer version, allow without --force
		return pluginActionUpgrade, existingMeta, nil
	case cmp < 0:
		// Downgrade: older version, require --force (potentially dangerous)
		if !force {
			return "", existingMeta, fmt.Errorf("plugin '%s' downgrade detected (%s → %s), use --force to downgrade",
				pluginInfo.Name, existingMeta.Version, pluginInfo.Version)
		}
		return pluginActionDowngrade, existingMeta, nil
	default:
		// Same version: require --force (unclear intent)
		if !force {
			return "", existingMeta, fmt.Errorf("plugin '%s' version %s is already installed (use --force to reinstall)",
				pluginInfo.Name, pluginInfo.Version)
		}
		return pluginActionOverwrite, existingMeta, nil
	}
}

// compareVersions compares two semantic version strings.
// Returns: >0 if v1 > v2, <0 if v1 < v2, 0 if equal.
func compareVersions(v1, v2 string) (int, error) {
	ver1, err := protocol.Parse(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid version %s: %w", v1, err)
	}

	ver2, err := protocol.Parse(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid version %s: %w", v2, err)
	}

	if ver1.Major != ver2.Major {
		return ver1.Major - ver2.Major, nil
	}
	if ver1.Minor != ver2.Minor {
		return ver1.Minor - ver2.Minor, nil
	}
	return ver1.Patch - ver2.Patch, nil
}

// installPlugin copies a plugin file to its final location.
func installPlugin(sourcePath, destPath string, verbose bool) error {
	if verbose {
		fmt.Fprintf(os.Stderr, "Installing plugin to %s\n", destPath)
	}

	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("failed to copy plugin: %w", err)
	}

	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return fmt.Errorf("failed to make plugin executable: %w", err)
	}

	return nil
}

// printPluginAddSuccess prints a success message based on the action performed.
func printPluginAddSuccess(pluginInfo *pluginMetadata, action pluginAction, existingMeta *ExternalPluginMeta, finalPath string) {
	switch action {
	case pluginActionAdd:
		fmt.Printf("Plugin '%s' added successfully\n", pluginInfo.Name)
	case pluginActionUpgrade:
		fmt.Printf("Plugin '%s' upgraded from %s to %s\n", pluginInfo.Name, existingMeta.Version, pluginInfo.Version)
	case pluginActionDowngrade:
		fmt.Printf("Plugin '%s' downgraded from %s to %s\n", pluginInfo.Name, existingMeta.Version, pluginInfo.Version)
	case pluginActionOverwrite:
		fmt.Printf("Plugin '%s' overwritten\n", pluginInfo.Name)
	}

	if pluginInfo.Description != "" {
		fmt.Printf("Description: %s\n", pluginInfo.Description)
	}
	fmt.Printf("Type: %s\n", pluginInfo.Type)
	if pluginInfo.Version != "" {
		fmt.Printf("Version: %s\n", pluginInfo.Version)
	}
	fmt.Printf("Protocol: %s\n", pluginInfo.ProtocolVersion)
	fmt.Printf("Path: %s\n", finalPath)
}
