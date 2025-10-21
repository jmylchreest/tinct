// Package input provides the interface and base types for input plugins.
package input

import (
	"context"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/spf13/cobra"
)

// GenerateOptions holds options passed to input plugins during generation.
type GenerateOptions struct {
	// Verbose enables verbose output
	Verbose bool

	// DryRun generates output without side effects
	DryRun bool

	// ColourOverrides are manual colour specifications (role=hex)
	ColourOverrides []string

	// PluginArgs are custom arguments for this plugin
	PluginArgs map[string]any
}

// ThemeHinter is an optional interface that input plugins can implement
// to provide hints about theme detection to the categorizer.
// This is purely advisory - the categorizer makes the final decision.
type ThemeHinter interface {
	// ThemeHint returns a suggested theme type based on the plugin's analysis.
	// Returns one of: "auto", "dark", "light", or "" for no hint.
	ThemeHint() string
}

// Plugin represents an input plugin that generates a colour palette.
type Plugin interface {
	// Name returns the plugin's name (e.g., "image", "file").
	Name() string

	// Description returns a human-readable description of the plugin.
	Description() string

	// Generate creates a raw Palette from plugin-specific inputs.
	// opts contains flags and arguments passed from the CLI.
	// Returns a simple list of colors - categorization happens separately.
	Generate(ctx context.Context, opts GenerateOptions) (*colour.Palette, error)

	// RegisterFlags registers plugin-specific flags with cobra command.
	RegisterFlags(cmd *cobra.Command)

	// Validate checks if the plugin has all required inputs configured.
	Validate() error
}

// Registry holds all registered input plugins.
type Registry struct {
	plugins map[string]Plugin
}

// NewRegistry creates a new input plugin registry.
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
