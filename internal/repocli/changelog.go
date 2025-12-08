// Package repocli provides CLI commands for the repository manager tool.
package repocli

import (
	"fmt"
	"sort"
	"strings"
)

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
func (c *ChangeLog) HasMaterialChanges() bool {
	return len(c.PluginsAdded) > 0 ||
		len(c.VersionsAdded) > 0 ||
		len(c.PlatformsAdded) > 0 ||
		len(c.EntriesRemoved) > 0
}

// String formats the changelog as a human-readable string.
func (c *ChangeLog) String() string {
	if c.IsEmpty() {
		return "No changes"
	}

	var sb strings.Builder

	// New plugins
	if len(c.PluginsAdded) > 0 {
		sb.WriteString("\nNew plugins:\n")
		sort.Slice(c.PluginsAdded, func(i, j int) bool {
			return c.PluginsAdded[i].Name < c.PluginsAdded[j].Name
		})
		for _, p := range c.PluginsAdded {
			platforms := strings.Join(p.Platforms, ", ")
			sb.WriteString(fmt.Sprintf("  - %s v%s (%s)\n", p.Name, p.Version, platforms))
		}
	}

	// New versions
	if len(c.VersionsAdded) > 0 {
		sb.WriteString("\nNew versions:\n")
		sort.Slice(c.VersionsAdded, func(i, j int) bool {
			if c.VersionsAdded[i].Plugin == c.VersionsAdded[j].Plugin {
				return c.VersionsAdded[i].Version > c.VersionsAdded[j].Version
			}
			return c.VersionsAdded[i].Plugin < c.VersionsAdded[j].Plugin
		})
		for _, v := range c.VersionsAdded {
			platforms := strings.Join(v.Platforms, ", ")
			sb.WriteString(fmt.Sprintf("  - %s v%s (%s)\n", v.Plugin, v.Version, platforms))
		}
	}

	// New platforms
	if len(c.PlatformsAdded) > 0 {
		sb.WriteString("\nPlatform additions:\n")
		sort.Slice(c.PlatformsAdded, func(i, j int) bool {
			if c.PlatformsAdded[i].Plugin == c.PlatformsAdded[j].Plugin {
				if c.PlatformsAdded[i].Version == c.PlatformsAdded[j].Version {
					return c.PlatformsAdded[i].Platform < c.PlatformsAdded[j].Platform
				}
				return c.PlatformsAdded[i].Version > c.PlatformsAdded[j].Version
			}
			return c.PlatformsAdded[i].Plugin < c.PlatformsAdded[j].Plugin
		})
		for _, p := range c.PlatformsAdded {
			sb.WriteString(fmt.Sprintf("  - %s v%s: added %s\n", p.Plugin, p.Version, p.Platform))
		}
	}

	// Removed entries
	if len(c.EntriesRemoved) > 0 {
		sb.WriteString("\nRemoved:\n")
		sort.Slice(c.EntriesRemoved, func(i, j int) bool {
			if c.EntriesRemoved[i].Plugin == c.EntriesRemoved[j].Plugin {
				return c.EntriesRemoved[i].Version > c.EntriesRemoved[j].Version
			}
			return c.EntriesRemoved[i].Plugin < c.EntriesRemoved[j].Plugin
		})
		for _, r := range c.EntriesRemoved {
			sb.WriteString(fmt.Sprintf("  - %s v%s (%s)\n", r.Plugin, r.Version, r.Reason))
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
			sb.WriteString(fmt.Sprintf("  - %s: updated %s\n", m.Plugin, fields))
		}
	}

	// Summary
	totalAdded := len(c.PluginsAdded) + len(c.VersionsAdded)
	totalPlatforms := len(c.PlatformsAdded)
	totalRemoved := len(c.EntriesRemoved)

	sb.WriteString("\nSummary:")
	if totalAdded > 0 {
		sb.WriteString(fmt.Sprintf(" +%d versions", totalAdded))
	}
	if totalPlatforms > 0 {
		sb.WriteString(fmt.Sprintf(" +%d platforms", totalPlatforms))
	}
	if totalRemoved > 0 {
		sb.WriteString(fmt.Sprintf(" -%d removed", totalRemoved))
	}
	if len(c.MetadataUpdated) > 0 {
		sb.WriteString(fmt.Sprintf(" ~%d metadata", len(c.MetadataUpdated)))
	}
	sb.WriteString("\n")

	return sb.String()
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
		return "No changes"
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
	sb.WriteString(fmt.Sprintf("  \"plugins_added\": %d,\n", len(c.PluginsAdded)))
	sb.WriteString(fmt.Sprintf("  \"versions_added\": %d,\n", len(c.VersionsAdded)))
	sb.WriteString(fmt.Sprintf("  \"platforms_added\": %d,\n", len(c.PlatformsAdded)))
	sb.WriteString(fmt.Sprintf("  \"entries_removed\": %d,\n", len(c.EntriesRemoved)))
	sb.WriteString(fmt.Sprintf("  \"metadata_updated\": %d\n", len(c.MetadataUpdated)))
	sb.WriteString("}\n")
	return sb.String()
}
