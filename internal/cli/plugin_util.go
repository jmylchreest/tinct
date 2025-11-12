// Package cli provides the command-line interface for Tinct.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
)

// getPluginDir returns the plugin directory path.
func getPluginDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "tinct", "plugins"), nil
}

// queryPluginMetadata queries a plugin for its name, description, type, and version.
func queryPluginMetadata(pluginPath string) (name, description, pluginType, version, protocolVersion string) {
	cmd := exec.Command(pluginPath, "--plugin-info")
	output, err := cmd.Output()
	if err != nil {
		return "", "", "", "", ""
	}

	var info struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Type            string `json:"type"`
		Version         string `json:"version"`
		ProtocolVersion string `json:"protocol_version"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return "", "", "", "", ""
	}

	return info.Name, info.Description, info.Type, info.Version, info.ProtocolVersion
}

// formatPluginSourceString converts a PluginSource struct to a display string.
func formatPluginSourceString(source *repository.PluginSource) string {
	if source == nil {
		return ""
	}
	switch source.Type {
	case sourceTypeRepository:
		return fmt.Sprintf("repo:%s/%s@%s", source.Repository, source.Plugin, source.Version)
	case sourceTypeHTTP:
		return source.URL
	case sourceTypeLocal:
		return source.OriginalPath
	default:
		return source.Type
	}
}
