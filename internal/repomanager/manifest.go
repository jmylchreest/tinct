// Package repomanager provides repository manifest management functionality.
package repomanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
)

// ManifestManager handles repository manifest operations.
type ManifestManager struct {
	manifest *repository.Manifest
	path     string
	dirty    bool // Tracks if manifest has been modified
}

// LoadManifest loads a manifest from disk or creates a new one if it doesn't exist.
func LoadManifest(path string) (*ManifestManager, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new manifest with placeholder metadata
			mgr := &ManifestManager{
				manifest: &repository.Manifest{
					Version:     "1.0",
					Name:        "New Plugin Repository",
					Description: "A new Tinct plugin repository",
					URL:         "https://example.com/repository.json",
					Plugins:     make(map[string]*repository.Plugin),
					LastUpdated: time.Now(),
				},
				path:  path,
				dirty: true,
			}
			// Save the new manifest immediately
			if err := mgr.Save(); err != nil {
				return nil, fmt.Errorf("failed to save new manifest: %w", err)
			}
			return mgr, nil
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest repository.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Ensure Plugins map is initialized
	if manifest.Plugins == nil {
		manifest.Plugins = make(map[string]*repository.Plugin)
	}

	return &ManifestManager{
		manifest: &manifest,
		path:     path,
	}, nil
}

// Save writes the manifest to disk.
func (m *ManifestManager) Save() error {
	// Skip saving if no changes were made
	if !m.dirty {
		return nil
	}

	// Update LastUpdated timestamp
	m.manifest.LastUpdated = time.Now()

	// Use json.Encoder with SetEscapeHTML(false) to prevent escaping <, >, & in compatibility strings
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(m.manifest); err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(m.path, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Reset dirty flag after successful save
	m.dirty = false

	return nil
}

// AddOrUpdatePluginVersion adds or updates a plugin version.
func (m *ManifestManager) AddOrUpdatePluginVersion(pluginName string, version *repository.Version) error {
	plugin, exists := m.manifest.Plugins[pluginName]
	if !exists {
		// Create new plugin entry
		plugin = &repository.Plugin{
			Name:     pluginName,
			Versions: []repository.Version{},
		}
		m.manifest.Plugins[pluginName] = plugin
		m.dirty = true // Mark as modified
	}

	// Ensure Versions is initialized
	if plugin.Versions == nil {
		plugin.Versions = []repository.Version{}
	}

	// Check if version already exists
	versionExists := false
	for i, v := range plugin.Versions {
		if v.Version == version.Version {
			// Merge downloads from existing and new version
			if version.Downloads != nil {
				if v.Downloads == nil {
					plugin.Versions[i].Downloads = version.Downloads
					m.dirty = true
				} else {
					// Merge platform-specific downloads
					for platform, download := range version.Downloads {
						existing, exists := v.Downloads[platform]
						// Only mark dirty if platform is new or download details changed
						if !exists || existing.URL != download.URL || existing.Checksum != download.Checksum {
							v.Downloads[platform] = download
							m.dirty = true
						}
					}
					plugin.Versions[i].Downloads = v.Downloads
				}
			}

			// Update other fields if provided
			if version.Compatibility != "" && plugin.Versions[i].Compatibility != version.Compatibility {
				plugin.Versions[i].Compatibility = version.Compatibility
				m.dirty = true
			}
			if version.ChangelogURL != "" && plugin.Versions[i].ChangelogURL != version.ChangelogURL {
				plugin.Versions[i].ChangelogURL = version.ChangelogURL
				m.dirty = true
			}

			versionExists = true
			break
		}
	}

	if !versionExists {
		// Ensure Downloads map is initialized
		if version.Downloads == nil {
			version.Downloads = make(map[string]*repository.Download)
		}
		// Add new version and sort (newest first)
		plugin.Versions = append(plugin.Versions, *version)
		sortVersionsNewestFirst(plugin.Versions)
		m.dirty = true
	}

	return nil
}

// sortVersionsNewestFirst sorts versions by release date, newest first.
func sortVersionsNewestFirst(versions []repository.Version) {
	sort.Slice(versions, func(i, j int) bool {
		// If dates are equal, compare version strings
		if versions[i].Released.Equal(versions[j].Released) {
			return CompareVersions(versions[i].Version, versions[j].Version) > 0
		}
		// Sort by date, newest first
		return versions[i].Released.After(versions[j].Released)
	})
}

// CompareVersions compares semantic versions (returns 1 if a > b, -1 if a < b, 0 if equal).
// Exported for use in protocol version cascade filtering.
func CompareVersions(a, b string) int {
	// Simple version comparison - split by dots and compare numerically
	aParts := strings.Split(strings.TrimPrefix(a, "v"), ".")
	bParts := strings.Split(strings.TrimPrefix(b, "v"), ".")

	maxLen := max(len(bParts), len(aParts))

	for i := range maxLen {
		var aNum, bNum int

		if i < len(aParts) {
			// Ignore error - invalid numbers default to 0
			//nolint:errcheck // Intentionally ignoring error - invalid numbers default to 0
			fmt.Sscanf(aParts[i], "%d", &aNum)
		}
		if i < len(bParts) {
			// Ignore error - invalid numbers default to 0
			//nolint:errcheck // Intentionally ignoring error - invalid numbers default to 0
			fmt.Sscanf(bParts[i], "%d", &bNum)
		}

		if aNum > bNum {
			return 1
		} else if aNum < bNum {
			return -1
		}
	}

	return 0
}

// RemovePluginVersion removes a specific version of a plugin.
func (m *ManifestManager) RemovePluginVersion(pluginName, version string) error {
	plugin, exists := m.manifest.Plugins[pluginName]
	if !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	// Find and remove version
	found := false
	for i, v := range plugin.Versions {
		if v.Version == version {
			plugin.Versions = append(plugin.Versions[:i], plugin.Versions[i+1:]...)
			found = true
			m.dirty = true
			break
		}
	}

	if !found {
		return fmt.Errorf("version '%s' not found for plugin '%s'", version, pluginName)
	}

	// Remove plugin entirely if no versions left
	if len(plugin.Versions) == 0 {
		delete(m.manifest.Plugins, pluginName)
		m.dirty = true
	}

	return nil
}

// RemovePlugin removes a plugin and all its versions.
func (m *ManifestManager) RemovePlugin(pluginName string) error {
	if _, exists := m.manifest.Plugins[pluginName]; !exists {
		return fmt.Errorf("plugin '%s' not found", pluginName)
	}

	delete(m.manifest.Plugins, pluginName)
	m.dirty = true
	return nil
}

// GetManifest returns the underlying manifest.
func (m *ManifestManager) GetManifest() *repository.Manifest {
	return m.manifest
}

// MarkDirty marks the manifest as modified.
// This should be called when external code directly modifies the manifest.
func (m *ManifestManager) MarkDirty() {
	m.dirty = true
}

// SetPluginMetadata updates plugin metadata (description, author, etc.).
func (m *ManifestManager) SetPluginMetadata(pluginName string, metadata *PluginMetadata) {
	plugin, exists := m.manifest.Plugins[pluginName]
	if !exists {
		plugin = &repository.Plugin{
			Name:     pluginName,
			Versions: []repository.Version{},
		}
		m.manifest.Plugins[pluginName] = plugin
		m.dirty = true
	}

	if metadata.Type != "" && plugin.Type != metadata.Type {
		plugin.Type = metadata.Type
		m.dirty = true
	}
	if metadata.Description != "" && plugin.Description != metadata.Description {
		plugin.Description = metadata.Description
		m.dirty = true
	}
	if metadata.Author != "" && plugin.Author != metadata.Author {
		plugin.Author = metadata.Author
		m.dirty = true
	}
	if metadata.Repository != "" && plugin.Repository != metadata.Repository {
		plugin.Repository = metadata.Repository
		m.dirty = true
	}
	if len(metadata.Tags) > 0 {
		// Check if tags actually changed
		tagsChanged := len(plugin.Tags) != len(metadata.Tags)
		if !tagsChanged {
			for i, tag := range metadata.Tags {
				if i >= len(plugin.Tags) || plugin.Tags[i] != tag {
					tagsChanged = true
					break
				}
			}
		}
		if tagsChanged {
			plugin.Tags = metadata.Tags
			m.dirty = true
		}
	}
	if metadata.License != "" && plugin.License != metadata.License {
		plugin.License = metadata.License
		m.dirty = true
	}
}

// SetManifestMetadata updates top-level manifest metadata.
func (m *ManifestManager) SetManifestMetadata(name, description, url, maintainedBy string) {
	if name != "" && m.manifest.Name != name {
		m.manifest.Name = name
		m.dirty = true
	}
	if description != "" && m.manifest.Description != description {
		m.manifest.Description = description
		m.dirty = true
	}
	if url != "" && m.manifest.URL != url {
		m.manifest.URL = url
		m.dirty = true
	}
	if maintainedBy != "" && m.manifest.MaintainedBy != maintainedBy {
		m.manifest.MaintainedBy = maintainedBy
		m.dirty = true
	}
}
