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
	dirty    bool                 // Tracks if manifest has been modified
	snapshot *repository.Manifest // Snapshot of manifest before sync for diff calculation
}

// AddResult describes what happened during an add/update operation.
type AddResult struct {
	PluginAdded    bool     // True if this is a new plugin
	VersionAdded   bool     // True if this is a new version
	PlatformsAdded []string // List of platforms that were added
	Updated        bool     // True if anything was updated
}

// LoadManifest loads a manifest from disk or creates a new one if it doesn't exist.
func LoadManifest(path string) (*ManifestManager, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- Manifest path from repo manager
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
		snapshot: nil, // Snapshot is created on-demand
	}, nil
}

// Save writes the manifest to disk.
func (m *ManifestManager) Save() error {
	// Skip saving if no changes were made
	if !m.dirty {
		return nil
	}

	// Only update LastUpdated timestamp if we have material changes
	// This prevents the timestamp from causing spurious commits when nothing actually changed
	if m.snapshot != nil {
		diff := m.ComputeDiff()
		if diff != nil && (len(diff.PluginsAdded) > 0 || len(diff.PluginsRemoved) > 0 || len(diff.PluginsChanged) > 0) {
			m.manifest.LastUpdated = time.Now()
		}
	} else {
		// No snapshot means we can't compute diff, so assume material changes
		m.manifest.LastUpdated = time.Now()
	}

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
// Returns detailed information about what was added/updated.
func (m *ManifestManager) AddOrUpdatePluginVersion(pluginName string, version *repository.Version) (*AddResult, error) { //nolint:gocognit // plugin version management with platform handling
	result := &AddResult{
		PlatformsAdded: []string{},
	}

	plugin, exists := m.manifest.Plugins[pluginName]
	if !exists {
		// Create new plugin entry
		plugin = &repository.Plugin{
			Name:     pluginName,
			Versions: []repository.Version{},
		}
		m.manifest.Plugins[pluginName] = plugin
		m.dirty = true // Mark as modified
		result.PluginAdded = true
		result.Updated = true
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
					result.Updated = true
					// All platforms are new
					for platform := range version.Downloads {
						result.PlatformsAdded = append(result.PlatformsAdded, platform)
					}
				} else {
					// Merge platform-specific downloads
					for platform, download := range version.Downloads {
						existing, exists := v.Downloads[platform]
						// Only mark dirty if platform is new or download details changed
						if !exists {
							v.Downloads[platform] = download
							m.dirty = true
							result.Updated = true
							result.PlatformsAdded = append(result.PlatformsAdded, platform)
						} else if existing.URL != download.URL || existing.Checksum != download.Checksum {
							v.Downloads[platform] = download
							m.dirty = true
							result.Updated = true
						}
					}
					plugin.Versions[i].Downloads = v.Downloads
				}
			}

			// Update other fields if provided
			if version.Compatibility != "" && plugin.Versions[i].Compatibility != version.Compatibility {
				plugin.Versions[i].Compatibility = version.Compatibility
				m.dirty = true
				result.Updated = true
			}
			if version.ChangelogURL != "" && plugin.Versions[i].ChangelogURL != version.ChangelogURL {
				plugin.Versions[i].ChangelogURL = version.ChangelogURL
				m.dirty = true
				result.Updated = true
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
		result.VersionAdded = true
		result.Updated = true
		// All platforms are new
		for platform := range version.Downloads {
			result.PlatformsAdded = append(result.PlatformsAdded, platform)
		}
	}

	return result, nil
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
// Returns list of fields that were changed.
func (m *ManifestManager) SetPluginMetadata(pluginName string, metadata *PluginMetadata) []string {
	changedFields := []string{}

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
		changedFields = append(changedFields, "type")
	}
	if metadata.Description != "" && plugin.Description != metadata.Description {
		plugin.Description = metadata.Description
		m.dirty = true
		changedFields = append(changedFields, "description")
	}
	if metadata.Author != "" && plugin.Author != metadata.Author {
		plugin.Author = metadata.Author
		m.dirty = true
		changedFields = append(changedFields, "author")
	}
	if metadata.Repository != "" && plugin.Repository != metadata.Repository {
		plugin.Repository = metadata.Repository
		m.dirty = true
		changedFields = append(changedFields, "repository")
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
			changedFields = append(changedFields, "tags")
		}
	}
	if metadata.License != "" && plugin.License != metadata.License {
		plugin.License = metadata.License
		m.dirty = true
		changedFields = append(changedFields, "license")
	}

	return changedFields
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

// CreateSnapshot creates a deep copy of the current manifest state.
// This should be called before starting a sync operation.
func (m *ManifestManager) CreateSnapshot() {
	// Create a deep copy by marshaling and unmarshaling
	data, err := json.Marshal(m.manifest)
	if err != nil {
		// If we can't create a snapshot, just skip it
		return
	}

	var snapshot repository.Manifest
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return
	}

	m.snapshot = &snapshot
}

// ManifestDiff represents the differences between two manifest states.
type ManifestDiff struct {
	PluginsAdded   map[string]*repository.Plugin // Plugins that didn't exist before
	PluginsRemoved map[string]*repository.Plugin // Plugins that were removed
	PluginsChanged map[string]*PluginVersionDiff // Plugins with version changes
}

// PluginVersionDiff represents version changes for a single plugin.
type PluginVersionDiff struct {
	VersionsAdded    map[string]*repository.Version // version -> Version
	VersionsRemoved  map[string]*repository.Version // version -> Version
	PlatformsChanged map[string]*PlatformDiff       // version -> platform changes
}

// PlatformDiff represents platform changes for a version.
type PlatformDiff struct {
	Added   []string // platforms added
	Removed []string // platforms removed
}

// ComputeDiff compares the current manifest with the snapshot and returns differences.
// Returns nil if no snapshot was created.
func (m *ManifestManager) ComputeDiff() *ManifestDiff {
	if m.snapshot == nil {
		return nil
	}

	diff := &ManifestDiff{
		PluginsAdded:   make(map[string]*repository.Plugin),
		PluginsRemoved: make(map[string]*repository.Plugin),
		PluginsChanged: make(map[string]*PluginVersionDiff),
	}

	// Find added and changed plugins
	for name, currentPlugin := range m.manifest.Plugins {
		snapshotPlugin, existed := m.snapshot.Plugins[name]

		if !existed {
			// Entire plugin is new
			diff.PluginsAdded[name] = currentPlugin
		} else {
			// Check for version differences
			versionDiff := computePluginVersionDiff(snapshotPlugin, currentPlugin)
			if versionDiff != nil {
				diff.PluginsChanged[name] = versionDiff
			}
		}
	}

	// Find removed plugins
	for name, snapshotPlugin := range m.snapshot.Plugins {
		if _, exists := m.manifest.Plugins[name]; !exists {
			diff.PluginsRemoved[name] = snapshotPlugin
		}
	}

	return diff
}

// computePluginVersionDiff compares versions between old and new plugin states.
func computePluginVersionDiff(oldPlugin, newPlugin *repository.Plugin) *PluginVersionDiff {
	// Build maps for easier comparison
	oldVersions := make(map[string]*repository.Version)
	for i := range oldPlugin.Versions {
		v := &oldPlugin.Versions[i]
		oldVersions[v.Version] = v
	}

	newVersions := make(map[string]*repository.Version)
	for i := range newPlugin.Versions {
		v := &newPlugin.Versions[i]
		newVersions[v.Version] = v
	}

	diff := &PluginVersionDiff{
		VersionsAdded:    make(map[string]*repository.Version),
		VersionsRemoved:  make(map[string]*repository.Version),
		PlatformsChanged: make(map[string]*PlatformDiff),
	}

	hasChanges := false

	// Find added versions and platform changes
	for ver, newVer := range newVersions {
		oldVer, existed := oldVersions[ver]

		if !existed {
			// Entire version is new
			diff.VersionsAdded[ver] = newVer
			hasChanges = true
		} else {
			// Check for platform differences
			platformDiff := computePlatformDiff(oldVer, newVer)
			if platformDiff != nil {
				diff.PlatformsChanged[ver] = platformDiff
				hasChanges = true
			}
		}
	}

	// Find removed versions
	for ver, oldVer := range oldVersions {
		if _, exists := newVersions[ver]; !exists {
			diff.VersionsRemoved[ver] = oldVer
			hasChanges = true
		}
	}

	if !hasChanges {
		return nil
	}

	return diff
}

// computePlatformDiff compares platforms between old and new versions.
func computePlatformDiff(oldVer, newVer *repository.Version) *PlatformDiff {
	if oldVer.Downloads == nil && newVer.Downloads == nil {
		return nil
	}

	oldPlatforms := make(map[string]bool)
	if oldVer.Downloads != nil {
		for platform := range oldVer.Downloads {
			oldPlatforms[platform] = true
		}
	}

	newPlatforms := make(map[string]bool)
	if newVer.Downloads != nil {
		for platform := range newVer.Downloads {
			newPlatforms[platform] = true
		}
	}

	diff := &PlatformDiff{
		Added:   []string{},
		Removed: []string{},
	}

	// Find added platforms
	for platform := range newPlatforms {
		if !oldPlatforms[platform] {
			diff.Added = append(diff.Added, platform)
		}
	}

	// Find removed platforms
	for platform := range oldPlatforms {
		if !newPlatforms[platform] {
			diff.Removed = append(diff.Removed, platform)
		}
	}

	if len(diff.Added) == 0 && len(diff.Removed) == 0 {
		return nil
	}

	// Sort for consistent output
	sort.Strings(diff.Added)
	sort.Strings(diff.Removed)

	return diff
}
