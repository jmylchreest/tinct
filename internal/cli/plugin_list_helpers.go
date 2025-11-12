// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"sort"

	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/jmylchreest/tinct/internal/plugin/protocol"
)

// pluginInfo holds information about a plugin for display.
type pluginInfo struct {
	pluginType      string // input or output
	name            string // plugin name without type prefix
	status          string
	version         string
	protocolVersion string // plugin protocol version
	description     string
	isExternal      bool
	source          string
}

// pluginCollector collects and organizes plugin information.
type pluginCollector struct {
	mgr         *manager.Manager
	lock        *PluginLock
	plugins     []pluginInfo
	seenPlugins map[string]bool
}

// newPluginCollector creates a new plugin collector.
func newPluginCollector(mgr *manager.Manager, lock *PluginLock) *pluginCollector {
	inputCount := len(mgr.AllInputPlugins())
	outputCount := len(mgr.AllOutputPlugins())

	return &pluginCollector{
		mgr:         mgr,
		lock:        lock,
		plugins:     make([]pluginInfo, 0, inputCount+outputCount),
		seenPlugins: make(map[string]bool),
	}
}

// addInputPlugins adds all input plugins to the collection.
func (c *pluginCollector) addInputPlugins() {
	for name, plugin := range c.mgr.AllInputPlugins() {
		protocolVersion := c.getPluginProtocolVersion(name, "input")
		info := c.buildPluginInfo("input", name, plugin.Version(), plugin.Description(), protocolVersion)
		c.plugins = append(c.plugins, info)
		c.seenPlugins[name] = true
	}
}

// addOutputPlugins adds all output plugins to the collection.
func (c *pluginCollector) addOutputPlugins() {
	for name, plugin := range c.mgr.AllOutputPlugins() {
		protocolVersion := c.getPluginProtocolVersion(name, "output")
		info := c.buildPluginInfo("output", name, plugin.Version(), plugin.Description(), protocolVersion)
		c.plugins = append(c.plugins, info)
		c.seenPlugins[name] = true
	}
}

// buildPluginInfo builds plugin information from a managed plugin.
func (c *pluginCollector) buildPluginInfo(pluginType, name, version, description, protocolVersion string) pluginInfo {
	status := c.determinePluginStatus(pluginType, name)
	isExternal := c.isExternalPlugin(name, pluginType)
	source := c.getPluginPath(name, pluginType)

	return pluginInfo{
		pluginType:      pluginType,
		name:            name,
		status:          status,
		version:         version,
		protocolVersion: protocolVersion,
		description:     description,
		isExternal:      isExternal,
		source:          source,
	}
}

// determinePluginStatus determines the status of a plugin (enabled/disabled/on-demand).
func (c *pluginCollector) determinePluginStatus(_ /* pluginType */, pluginName string) string {
	if c.lock == nil {
		return "O" // on-demand
	}

	// Check disabled list.
	if c.isInList(c.lock.DisabledPlugins, pluginName) {
		return "D" // disabled
	}

	// Check enabled list.
	if len(c.lock.EnabledPlugins) > 0 {
		if c.isInList(c.lock.EnabledPlugins, pluginName) {
			return "E" // enabled
		}
		return "O" // on-demand
	}

	return "O" // on-demand
}

// isInList checks if a plugin name is in a list.
func (c *pluginCollector) isInList(list []string, name string) bool {
	for _, item := range list {
		if item == name || item == pluginTypeAll {
			return true
		}
	}
	return false
}

// isExternalPlugin checks if a plugin is an external plugin.
func (c *pluginCollector) isExternalPlugin(name, pluginType string) bool {
	if c.lock == nil || c.lock.ExternalPlugins == nil {
		return false
	}

	for _, meta := range c.lock.ExternalPlugins {
		if meta.Name == name && meta.Type == pluginType {
			return true
		}
	}
	return false
}

// getPluginPath retrieves the actual path for an external plugin.
func (c *pluginCollector) getPluginPath(name, pluginType string) string {
	if c.lock == nil || c.lock.ExternalPlugins == nil {
		return ""
	}

	for _, meta := range c.lock.ExternalPlugins {
		if meta.Name == name && meta.Type == pluginType {
			// Return the actual plugin path being used, not the original source
			return meta.Path
		}
	}
	return ""
}

// getPluginProtocolVersion retrieves the protocol version for a plugin.
func (c *pluginCollector) getPluginProtocolVersion(name, pluginType string) string {
	// Check if it's an external plugin and query it directly
	if c.lock != nil && c.lock.ExternalPlugins != nil {
		for _, meta := range c.lock.ExternalPlugins {
			if meta.Name == name && meta.Type == pluginType {
				// Query the plugin directly for its protocol version
				_, _, _, _, protocolVersion := queryPluginMetadata(meta.Path)
				if protocolVersion != "" {
					return protocolVersion
				}
				// If query failed, print warning and return "unknown"
				fmt.Printf("Warning: Failed to query protocol version from external plugin '%s' (%s) at %s\n", name, pluginType, meta.Path)
				return "unknown"
			}
		}
	}

	// For built-in plugins, they all use the current protocol version
	return protocol.ProtocolVersion
}

// addExternalOnlyPlugins adds plugins that are only in the lock file (not in manager).
func (c *pluginCollector) addExternalOnlyPlugins() {
	if c.lock == nil || c.lock.ExternalPlugins == nil {
		return
	}

	for lockKey, meta := range c.lock.ExternalPlugins {
		pluginName := meta.Name
		if pluginName == "" {
			pluginName = lockKey
		}

		if c.seenPlugins[pluginName] {
			continue // Already added from manager
		}

		info := c.buildExternalOnlyInfo(pluginName, meta)
		c.plugins = append(c.plugins, info)
	}
}

// buildExternalOnlyInfo builds plugin info for external-only plugins.
func (c *pluginCollector) buildExternalOnlyInfo(name string, meta *ExternalPluginMeta) pluginInfo {
	status := c.determinePluginStatus(meta.Type, name)

	// Query the plugin directly for current metadata
	_, queryDescription, _, queryVersion, queryProtocolVersion := queryPluginMetadata(meta.Path)

	description := queryDescription
	if description == "" {
		description = meta.Description
	}
	if description == "" {
		sourceStr := formatPluginSourceString(meta.Source)
		if sourceStr == "" {
			sourceStr = meta.SourceLegacy
		}
		description = fmt.Sprintf("External plugin (source: %s)", sourceStr)
	}

	version := queryVersion
	if version == "" {
		version = meta.Version
	}
	if version == "" {
		version = "unknown"
	}

	protocolVersion := queryProtocolVersion
	if protocolVersion == "" {
		fmt.Printf("Warning: Failed to query protocol version from external plugin '%s' (%s) at %s\n", name, meta.Type, meta.Path)
		protocolVersion = "unknown"
	}

	return pluginInfo{
		pluginType:      meta.Type,
		name:            name,
		status:          status,
		version:         version,
		protocolVersion: protocolVersion,
		description:     description,
		isExternal:      true,
		source:          meta.Path, // Show actual plugin path, not original source
	}
}

// getSortedPlugins returns all collected plugins sorted by type then name.
func (c *pluginCollector) getSortedPlugins() []pluginInfo {
	sort.Slice(c.plugins, func(i, j int) bool {
		if c.plugins[i].pluginType != c.plugins[j].pluginType {
			return c.plugins[i].pluginType < c.plugins[j].pluginType
		}
		return c.plugins[i].name < c.plugins[j].name
	})
	return c.plugins
}

// collectAllPlugins collects all plugin information.
func collectAllPlugins(mgr *manager.Manager, lock *PluginLock) []pluginInfo {
	collector := newPluginCollector(mgr, lock)

	collector.addInputPlugins()
	collector.addOutputPlugins()
	collector.addExternalOnlyPlugins()

	return collector.getSortedPlugins()
}

// displayPluginTable displays plugins in a formatted table.
func displayPluginTable(plugins []pluginInfo, showPath bool) {
	var headers []string
	if showPath {
		headers = []string{"", "S", "TYPE", "PLUGIN", "VERSION", "C", "PATH"}
	} else {
		headers = []string{"", "S", "TYPE", "PLUGIN", "VERSION", "C", "DESCRIPTION"}
	}

	tbl := NewTable(headers)

	// Enable terminal-aware column sizing
	// Last column (description or path) will automatically size to fit terminal width
	tbl.EnableTerminalAwareWidth(6, 40) // Min width of 40 chars

	for _, p := range plugins {
		addPluginToTable(tbl, p, showPath)
	}

	fmt.Print(tbl.Render())

	// Print legends
	fmt.Println()
	if hasExternalPlugins(plugins) {
		pluginDir, err := getPluginDir()
		if err == nil {
			fmt.Printf("* = external plugin (default location: %s)\n", pluginDir)
		} else {
			fmt.Println("* = external plugin")
		}
	}
	fmt.Println("S = Status: O (on-demand), E (enabled), D (disabled)")
	fmt.Println("C = Compatible with current tinct (Y/N)")
}

// addPluginToTable adds a single plugin to the table.
func addPluginToTable(tbl *Table, p pluginInfo, showPath bool) {
	marker := ""
	if p.isExternal {
		marker = "*"
	}

	// Check protocol compatibility
	compatible := "Y"
	if p.protocolVersion == "" || p.protocolVersion == "unknown" {
		// Plugin query failed or protocol version unavailable
		compatible = "N"
	} else {
		isCompat, _ := protocol.IsCompatible(p.protocolVersion)
		if !isCompat {
			compatible = "N"
		}
	}

	var lastColumn string
	if showPath {
		lastColumn = p.source
		if lastColumn == "" {
			lastColumn = "(built-in)"
		}
	} else {
		lastColumn = p.description
	}

	tbl.AddRow([]string{marker, p.status, p.pluginType, p.name, p.version, compatible, lastColumn})
}

// hasExternalPlugins checks if any plugins in the list are external.
func hasExternalPlugins(plugins []pluginInfo) bool {
	for _, p := range plugins {
		if p.isExternal {
			return true
		}
	}
	return false
}
