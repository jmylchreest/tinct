// Package repocli provides CLI commands for the repository manager tool.
package repocli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jmylchreest/tinct/internal/repomanager"
)

const noChangesMessage = "No changes"

// ChangeLog tracks all changes made during a sync operation.
type ChangeLog struct {
	// PluginsAdded tracks entirely new plugins (didn't exist before)
	PluginsAdded []PluginAddition

	// VersionsAdded tracks new versions of existing plugins
	VersionsAdded []VersionAddition

	// PlatformsAdded tracks new platforms added to existing versions
	PlatformsAdded []PlatformAddition

	// EntriesRemoved tracks entries that were pruned/removed
	EntriesRemoved []EntryRemoval

	// MetadataUpdated tracks plugins whose metadata was updated
	MetadataUpdated []MetadataUpdate
}

// PluginChanges represents all changes for a single plugin.
type PluginChanges struct {
	Plugin          string
	VersionsAdded   []VersionChange
	VersionsRemoved []VersionChange
}

// VersionChange represents changes to a specific version.
type VersionChange struct {
	Version          string
	PlatformsAdded   []string
	PlatformsRemoved []string
	Reason           string // For removals: "old version pruned", etc.
}

// PluginAddition represents a newly added plugin.
type PluginAddition struct {
	Name      string
	Version   string
	Platforms []string
}

// VersionAddition represents a new version of an existing plugin.
type VersionAddition struct {
	Plugin    string
	Version   string
	Platforms []string
}

// PlatformAddition represents a new platform for an existing version.
type PlatformAddition struct {
	Plugin   string
	Version  string
	Platform string
}

// EntryRemoval represents a removed plugin version.
type EntryRemoval struct {
	Plugin  string
	Version string
	Reason  string // e.g., "unavailable for 30+ days", "incompatible protocol", "old version"
}

// MetadataUpdate represents a plugin whose metadata changed.
type MetadataUpdate struct {
	Plugin string
	Fields []string // e.g., "description", "author", "tags"
}

// NewChangeLog creates a new empty changelog.
func NewChangeLog() *ChangeLog {
	return &ChangeLog{
		PluginsAdded:    []PluginAddition{},
		VersionsAdded:   []VersionAddition{},
		PlatformsAdded:  []PlatformAddition{},
		EntriesRemoved:  []EntryRemoval{},
		MetadataUpdated: []MetadataUpdate{},
	}
}

// AddPlugin records a new plugin addition.
func (c *ChangeLog) AddPlugin(name, version string, platforms []string) {
	c.PluginsAdded = append(c.PluginsAdded, PluginAddition{
		Name:      name,
		Version:   version,
		Platforms: platforms,
	})
}

// AddVersion records a new version of an existing plugin.
func (c *ChangeLog) AddVersion(plugin, version string, platforms []string) {
	c.VersionsAdded = append(c.VersionsAdded, VersionAddition{
		Plugin:    plugin,
		Version:   version,
		Platforms: platforms,
	})
}

// AddPlatform records a new platform for an existing version.
func (c *ChangeLog) AddPlatform(plugin, version, platform string) {
	c.PlatformsAdded = append(c.PlatformsAdded, PlatformAddition{
		Plugin:   plugin,
		Version:  version,
		Platform: platform,
	})
}

// RemoveEntry records a removed plugin version.
func (c *ChangeLog) RemoveEntry(plugin, version, reason string) {
	c.EntriesRemoved = append(c.EntriesRemoved, EntryRemoval{
		Plugin:  plugin,
		Version: version,
		Reason:  reason,
	})
}

// UpdateMetadata records a metadata update.
func (c *ChangeLog) UpdateMetadata(plugin string, fields []string) {
	c.MetadataUpdated = append(c.MetadataUpdated, MetadataUpdate{
		Plugin: plugin,
		Fields: fields,
	})
}

// IsEmpty returns true if no changes were recorded.
func (c *ChangeLog) IsEmpty() bool {
	return len(c.PluginsAdded) == 0 &&
		len(c.VersionsAdded) == 0 &&
		len(c.PlatformsAdded) == 0 &&
		len(c.EntriesRemoved) == 0 &&
		len(c.MetadataUpdated) == 0
}

// HasMaterialChanges returns true if there are changes beyond just metadata updates.
// Platform additions alone don't count as material changes to avoid noise from
// adding platforms to old versions that immediately get pruned.
func (c *ChangeLog) HasMaterialChanges() bool {
	return len(c.PluginsAdded) > 0 ||
		len(c.VersionsAdded) > 0 ||
		len(c.EntriesRemoved) > 0
}

// String formats the changelog as a human-readable string.
func (c *ChangeLog) String() string { //nolint:gocyclo,gocognit // formatted output with sections, categories, and markdown
	if c.IsEmpty() {
		return noChangesMessage
	}

	// Consolidate changes by plugin
	pluginChanges := c.consolidateByPlugin()

	var sb strings.Builder

	// New plugins
	if len(c.PluginsAdded) > 0 {
		sb.WriteString("\nNew plugins:\n")
		sort.Slice(c.PluginsAdded, func(i, j int) bool {
			return c.PluginsAdded[i].Name < c.PluginsAdded[j].Name
		})
		for _, p := range c.PluginsAdded {
			platforms := strings.Join(p.Platforms, ", ")
			fmt.Fprintf(&sb, "  - %s v%s (%s)\n", p.Name, p.Version, platforms)
		}
	}

	// Updated plugins (consolidated view)
	updatedPlugins := []string{}
	for plugin := range pluginChanges {
		// Skip if it's a new plugin
		isNew := false
		for _, p := range c.PluginsAdded {
			if p.Name == plugin {
				isNew = true
				break
			}
		}
		if !isNew {
			updatedPlugins = append(updatedPlugins, plugin)
		}
	}

	if len(updatedPlugins) > 0 {
		sort.Strings(updatedPlugins)
		sb.WriteString("\nUpdated plugins:\n")

		for _, plugin := range updatedPlugins {
			changes := pluginChanges[plugin]
			fmt.Fprintf(&sb, "\n  - %s:\n", plugin)

			// Show added versions with their platforms
			if len(changes.VersionsAdded) > 0 {
				sort.Slice(changes.VersionsAdded, func(i, j int) bool {
					return changes.VersionsAdded[i].Version > changes.VersionsAdded[j].Version
				})
				for _, v := range changes.VersionsAdded {
					if len(v.PlatformsAdded) > 0 {
						platforms := strings.Join(v.PlatformsAdded, ", ")
						fmt.Fprintf(&sb, "    - Added: v%s (%s)\n", v.Version, platforms)
					}
				}
			}

			// Show removed versions with their reason
			if len(changes.VersionsRemoved) > 0 {
				sort.Slice(changes.VersionsRemoved, func(i, j int) bool {
					return changes.VersionsRemoved[i].Version > changes.VersionsRemoved[j].Version
				})
				for _, v := range changes.VersionsRemoved {
					fmt.Fprintf(&sb, "    - Pruned: v%s (%s)\n", v.Version, v.Reason)
				}
			}
		}
	}

	// Metadata updates
	if len(c.MetadataUpdated) > 0 {
		sb.WriteString("\nMetadata updates:\n")
		sort.Slice(c.MetadataUpdated, func(i, j int) bool {
			return c.MetadataUpdated[i].Plugin < c.MetadataUpdated[j].Plugin
		})
		for _, m := range c.MetadataUpdated {
			fields := strings.Join(m.Fields, ", ")
			fmt.Fprintf(&sb, "  - %s: updated %s\n", m.Plugin, fields)
		}
	}

	// Summary
	totalAdded := len(c.PluginsAdded) + len(c.VersionsAdded)
	totalPlatforms := len(c.PlatformsAdded)
	totalRemoved := len(c.EntriesRemoved)

	sb.WriteString("\nSummary:")
	if totalAdded > 0 {
		fmt.Fprintf(&sb, " +%d versions", totalAdded)
	}
	if totalPlatforms > 0 {
		fmt.Fprintf(&sb, " +%d platforms", totalPlatforms)
	}
	if totalRemoved > 0 {
		fmt.Fprintf(&sb, " -%d removed", totalRemoved)
	}
	if len(c.MetadataUpdated) > 0 {
		fmt.Fprintf(&sb, " ~%d metadata", len(c.MetadataUpdated))
	}
	sb.WriteString("\n")

	return sb.String()
}

// BuildFromManifestDiff creates a changelog from a manifest diff.
// This is the proper way to build a changelog based on actual changes to the repository.
func BuildFromManifestDiff(diff *repomanager.ManifestDiff) *ChangeLog {
	if diff == nil {
		return NewChangeLog()
	}

	changelog := NewChangeLog()

	// Process newly added plugins
	for name, plugin := range diff.PluginsAdded {
		if len(plugin.Versions) == 0 {
			continue
		}
		version := plugin.Versions[0]
		platforms := []string{}
		if version.Downloads != nil {
			for platform := range version.Downloads {
				platforms = append(platforms, platform)
			}
		}
		sort.Strings(platforms)
		changelog.AddPlugin(name, version.Version, platforms)
	}

	// Process changed plugins
	for name, versionDiff := range diff.PluginsChanged {
		// Process added versions
		for version, ver := range versionDiff.VersionsAdded {
			platforms := []string{}
			if ver.Downloads != nil {
				for platform := range ver.Downloads {
					platforms = append(platforms, platform)
				}
			}
			sort.Strings(platforms)
			changelog.AddVersion(name, version, platforms)
		}

		// Process platform changes
		for version, platformDiff := range versionDiff.PlatformsChanged {
			for _, platform := range platformDiff.Added {
				changelog.AddPlatform(name, version, platform)
			}
		}

		// Process removed versions
		for version := range versionDiff.VersionsRemoved {
			changelog.RemoveEntry(name, version, "old version pruned")
		}
	}

	// Process removed plugins (all versions)
	for name, plugin := range diff.PluginsRemoved {
		for _, version := range plugin.Versions {
			changelog.RemoveEntry(name, version.Version, "plugin removed")
		}
	}

	return changelog
}

// consolidateByPlugin groups all changes by plugin name for easier diff-style display.
func (c *ChangeLog) consolidateByPlugin() map[string]*PluginChanges {
	result := make(map[string]*PluginChanges)

	// Process version additions
	for _, v := range c.VersionsAdded {
		if _, exists := result[v.Plugin]; !exists {
			result[v.Plugin] = &PluginChanges{
				Plugin:          v.Plugin,
				VersionsAdded:   []VersionChange{},
				VersionsRemoved: []VersionChange{},
			}
		}
		result[v.Plugin].VersionsAdded = append(result[v.Plugin].VersionsAdded, VersionChange{
			Version:        v.Version,
			PlatformsAdded: v.Platforms,
		})
	}

	// Process platform additions (add to existing versions or create new entries)
	platformsByVersion := make(map[string]map[string][]string) // plugin -> version -> platforms
	for _, p := range c.PlatformsAdded {
		if _, exists := platformsByVersion[p.Plugin]; !exists {
			platformsByVersion[p.Plugin] = make(map[string][]string)
		}
		platformsByVersion[p.Plugin][p.Version] = append(platformsByVersion[p.Plugin][p.Version], p.Platform)
	}

	for plugin, versions := range platformsByVersion {
		if _, exists := result[plugin]; !exists {
			result[plugin] = &PluginChanges{
				Plugin:          plugin,
				VersionsAdded:   []VersionChange{},
				VersionsRemoved: []VersionChange{},
			}
		}

		for version, platforms := range versions {
			// Check if this version already exists in VersionsAdded
			found := false
			for i, v := range result[plugin].VersionsAdded {
				if v.Version == version {
					// Merge platforms
					result[plugin].VersionsAdded[i].PlatformsAdded = append(
						result[plugin].VersionsAdded[i].PlatformsAdded,
						platforms...,
					)
					found = true
					break
				}
			}

			if !found {
				// This is a platform addition to an existing version
				result[plugin].VersionsAdded = append(result[plugin].VersionsAdded, VersionChange{
					Version:        version,
					PlatformsAdded: platforms,
				})
			}
		}
	}

	// Process removals
	for _, r := range c.EntriesRemoved {
		if _, exists := result[r.Plugin]; !exists {
			result[r.Plugin] = &PluginChanges{
				Plugin:          r.Plugin,
				VersionsAdded:   []VersionChange{},
				VersionsRemoved: []VersionChange{},
			}
		}
		result[r.Plugin].VersionsRemoved = append(result[r.Plugin].VersionsRemoved, VersionChange{
			Version: r.Version,
			Reason:  r.Reason,
		})
	}

	return result
}

// Format formats the changelog in a specific format.
func (c *ChangeLog) Format(format string) string {
	switch format {
	case "text", "":
		return c.String()
	case "short":
		return c.formatShort()
	case "json":
		return c.formatJSON()
	default:
		return c.String()
	}
}

// formatShort returns a one-line summary.
func (c *ChangeLog) formatShort() string {
	if c.IsEmpty() {
		return noChangesMessage
	}

	parts := []string{}

	totalAdded := len(c.PluginsAdded) + len(c.VersionsAdded)
	if totalAdded > 0 {
		parts = append(parts, fmt.Sprintf("+%d versions", totalAdded))
	}

	if len(c.PlatformsAdded) > 0 {
		parts = append(parts, fmt.Sprintf("+%d platforms", len(c.PlatformsAdded)))
	}

	if len(c.EntriesRemoved) > 0 {
		parts = append(parts, fmt.Sprintf("-%d removed", len(c.EntriesRemoved)))
	}

	if len(c.MetadataUpdated) > 0 {
		parts = append(parts, fmt.Sprintf("~%d metadata", len(c.MetadataUpdated)))
	}

	return strings.Join(parts, ", ")
}

// formatJSON returns a JSON representation (simple implementation).
func (c *ChangeLog) formatJSON() string {
	// For a proper JSON implementation, you'd want to use encoding/json
	// This is a simple string-based version for now
	var sb strings.Builder
	sb.WriteString("{\n")
	fmt.Fprintf(&sb, "  \"plugins_added\": %d,\n", len(c.PluginsAdded))
	fmt.Fprintf(&sb, "  \"versions_added\": %d,\n", len(c.VersionsAdded))
	fmt.Fprintf(&sb, "  \"platforms_added\": %d,\n", len(c.PlatformsAdded))
	fmt.Fprintf(&sb, "  \"entries_removed\": %d,\n", len(c.EntriesRemoved))
	fmt.Fprintf(&sb, "  \"metadata_updated\": %d\n", len(c.MetadataUpdated))
	sb.WriteString("}\n")
	return sb.String()
}
