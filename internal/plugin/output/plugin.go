// Package output provides the interface and base types for output plugins.
package output

import (
	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/spf13/cobra"
)

// Plugin represents an output plugin that can generate configuration files
// from a categorised colour palette.
type Plugin interface {
	// Name returns the plugin's name (e.g., "tailwind", "hyprland").
	Name() string

	// Description returns a human-readable description of the plugin.
	Description() string

	// Generate creates output file(s) from the given palette.
	// Returns map of filename -> content to support plugins that generate multiple files.
	Generate(palette *colour.CategorisedPalette) (map[string][]byte, error)

	// RegisterFlags registers plugin-specific flags with cobra command.
	RegisterFlags(cmd *cobra.Command)

	// Validate checks if the plugin configuration is valid.
	Validate() error

	// DefaultOutputDir returns the default output directory for this plugin.
	DefaultOutputDir() string
}

// Registry holds all registered output plugins.
type Registry struct {
	plugins map[string]Plugin
}

// NewRegistry creates a new plugin registry.
func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(plugin Plugin) {
	r.plugins[plugin.Name()] = plugin
}

// Get retrieves a plugin by name.
func (r *Registry) Get(name string) (Plugin, bool) {
	plugin, ok := r.plugins[name]
	return plugin, ok
}

// List returns all registered plugin names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// All returns all registered plugins (including disabled ones).
func (r *Registry) All() map[string]Plugin {
	// Return a copy to prevent external modification
	plugins := make(map[string]Plugin, len(r.plugins))
	for name, plugin := range r.plugins {
		plugins[name] = plugin
	}
	return plugins
}
