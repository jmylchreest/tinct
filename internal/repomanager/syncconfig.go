package repomanager

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// SyncSourceType represents the type of sync source.
type SyncSourceType string

const (
	SyncSourceGitHub SyncSourceType = "github"
	SyncSourceURL    SyncSourceType = "url"
)

// SyncSource represents a single source to sync plugins from.
type SyncSource struct {
	// Type of source: "github" or "url"
	Type SyncSourceType `json:"type"`

	// GitHub source fields
	Repo    string   `json:"repo,omitempty"`    // GitHub repo (owner/repo)
	Version string   `json:"version,omitempty"` // "latest", "all", or specific tag, or "-" for auto-detect
	Filter  []string `json:"filter,omitempty"`  // Patterns to include
	Exclude []string `json:"exclude,omitempty"` // Patterns to exclude

	// URL source fields
	URL        string `json:"url,omitempty"`         // Direct download URL
	Plugin     string `json:"plugin,omitempty"`      // Plugin name
	PluginType string `json:"plugin_type,omitempty"` // "input" or "output"
	Platform   string `json:"platform,omitempty"`    // Platform identifier
	Runtime    string `json:"runtime,omitempty"`     // Runtime (e.g., "bash", "python3")
}

// SyncConfig represents the complete sync configuration.
type SyncConfig struct {
	Sources []SyncSource
}

// LoadSyncConfig loads a sync configuration from a JSONL file.
func LoadSyncConfig(path string) (*SyncConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &SyncConfig{
		Sources: []SyncSource{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Skip comment lines (starting with #)
		if line[0] == '#' {
			continue
		}

		var source SyncSource
		if err := json.Unmarshal([]byte(line), &source); err != nil {
			return nil, fmt.Errorf("line %d: failed to parse JSON: %w", lineNum, err)
		}

		// Validate source
		if err := validateSyncSource(&source); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		config.Sources = append(config.Sources, source)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(config.Sources) == 0 {
		return nil, fmt.Errorf("no sync sources found in config file")
	}

	return config, nil
}

// validateSyncSource validates a sync source configuration.
func validateSyncSource(source *SyncSource) error {
	switch source.Type {
	case SyncSourceGitHub:
		if source.Repo == "" {
			return fmt.Errorf("github source requires 'repo' field")
		}
		if source.Version == "" {
			return fmt.Errorf("github source requires 'version' field")
		}
		if len(source.Filter) == 0 {
			return fmt.Errorf("github source requires 'filter' field")
		}

	case SyncSourceURL:
		if source.URL == "" {
			return fmt.Errorf("url source requires 'url' field")
		}
		if source.Plugin == "" {
			return fmt.Errorf("url source requires 'plugin' field")
		}
		if source.PluginType == "" {
			return fmt.Errorf("url source requires 'plugin_type' field")
		}
		if source.Version == "" {
			return fmt.Errorf("url source requires 'version' field")
		}
		if source.Platform == "" {
			return fmt.Errorf("url source requires 'platform' field")
		}

	default:
		return fmt.Errorf("unknown source type: %s (must be 'github' or 'url')", source.Type)
	}

	return nil
}

// CheckProtocolVersion checks if a plugin's protocol version meets the minimum requirement.
// Returns true if the plugin version >= minimum version, false otherwise.
func CheckProtocolVersion(pluginVersion, minVersion string) bool {
	if minVersion == "" {
		return true // No minimum requirement
	}

	if pluginVersion == "" {
		return false // Plugin has no version
	}

	// Use existing CompareVersions function from manifest.go
	return CompareVersions(pluginVersion, minVersion) >= 0
}
