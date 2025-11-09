// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"sort"

	"github.com/jmylchreest/tinct/internal/plugin/manager"
)

const (
	pluginStatusOnDemand = "on-demand"
)

// pluginInfo holds information about a plugin for display.
type pluginInfo struct {
	fullName    string
	status      string
	version     string
	description string
	isExternal  bool
	source      string
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
		info := c.buildPluginInfo("input", name, plugin.Version(), plugin.Description())
		c.plugins = append(c.plugins, info)
		c.seenPlugins[info.fullName] = true
	}
}

// addOutputPlugins adds all output plugins to the collection.
func (c *pluginCollector) addOutputPlugins() {
	for name, plugin := range c.mgr.AllOutputPlugins() {
		info := c.buildPluginInfo("output", name, plugin.Version(), plugin.Description())
		c.plugins = append(c.plugins, info)
		c.seenPlugins[info.fullName] = true
	}
}

// buildPluginInfo builds plugin information from a managed plugin.
func (c *pluginCollector) buildPluginInfo(pluginType, name, version, description string) pluginInfo {
	fullName := fmt.Sprintf("%s:%s", pluginType, name)
	status := c.determinePluginStatus(pluginType, name)
	isExternal := c.isExternalPlugin(name, pluginType)
	source := c.getPluginPath(name, pluginType)

	return pluginInfo{
		fullName:    fullName,
		status:      status,
		version:     version,
		description: description,
		isExternal:  isExternal,
		source:      source,
	}
}

// determinePluginStatus determines the status of a plugin (enabled/disabled/on-demand).
func (c *pluginCollector) determinePluginStatus(pluginType, pluginName string) string {
	if c.lock == nil {
		return pluginStatusOnDemand
	}

	fullName := fmt.Sprintf("%s:%s", pluginType, pluginName)

	// Check disabled list.
	if c.isInList(c.lock.DisabledPlugins, pluginName, fullName) {
		return "disabled"
	}

	// Check enabled list.
	if len(c.lock.EnabledPlugins) > 0 {
		if c.isInList(c.lock.EnabledPlugins, pluginName, fullName) {
			return "enabled"
		}
		return pluginStatusOnDemand
	}

	return pluginStatusOnDemand
}

// isInList checks if a plugin name or full name is in a list.
func (c *pluginCollector) isInList(list []string, name, fullName string) bool {
	for _, item := range list {
		if item == fullName || item == name || item == pluginTypeAll {
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

		fullName := fmt.Sprintf("%s:%s", meta.Type, pluginName)
		if c.seenPlugins[fullName] {
			continue // Already added from manager
		}

		info := c.buildExternalOnlyInfo(pluginName, meta)
		c.plugins = append(c.plugins, info)
	}
}

// buildExternalOnlyInfo builds plugin info for external-only plugins.
func (c *pluginCollector) buildExternalOnlyInfo(name string, meta *ExternalPluginMeta) pluginInfo {
	fullName := fmt.Sprintf("%s:%s", meta.Type, name)
	status := c.determinePluginStatus(meta.Type, name)

	description := meta.Description
	if description == "" {
		sourceStr := formatPluginSourceString(meta.Source)
		if sourceStr == "" {
			sourceStr = meta.SourceLegacy
		}
		description = fmt.Sprintf("External plugin (source: %s)", sourceStr)
	}

	version := meta.Version
	if version == "" {
		version = "unknown"
	}

	return pluginInfo{
		fullName:    fullName,
		status:      status,
		version:     version,
		description: description,
		isExternal:  true,
		source:      meta.Path, // Show actual plugin path, not original source
	}
}

// getSortedPlugins returns all collected plugins sorted by name.
func (c *pluginCollector) getSortedPlugins() []pluginInfo {
	sort.Slice(c.plugins, func(i, j int) bool {
		return c.plugins[i].fullName < c.plugins[j].fullName
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
func displayPluginTable(plugins []pluginInfo) {
	tbl := NewTable([]string{"", "PLUGIN", "STATUS", "VERSION", "DESCRIPTION"})

	// Set max width of 80 characters for the description column (index 4).
	tbl.SetColumnMaxWidth(4, 80)

	for _, p := range plugins {
		addPluginToTable(tbl, p)
	}

	fmt.Print(tbl.Render())

	if hasExternalPlugins(plugins) {
		fmt.Println()
		fmt.Println("* = external plugin")
	}
}

// addPluginToTable adds a single plugin to the table.
func addPluginToTable(tbl *Table, p pluginInfo) {
	marker := ""
	if p.isExternal {
		marker = "*"
	}

	tbl.AddRow([]string{marker, p.fullName, p.status, p.version, p.description})

	if p.isExternal && p.source != "" {
		tbl.AddRow([]string{"", "", "", "", "  path: " + p.source})
	}
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
