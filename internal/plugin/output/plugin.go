// Package output provides the interface and base types for output plugins.
package output

import (
	"github.com/jmylchreest/tinct/internal/colour"
)

// Plugin represents an output plugin that can generate configuration files
// from a categorised colour palette.
type Plugin interface {
	// Name returns the plugin's name (e.g., "tailwind", "alacritty").
	Name() string

	// Description returns a human-readable description of the plugin.
	Description() string

	// Generate creates the output file content from the given palette.
	Generate(palette *colour.CategorisedPalette) ([]byte, error)

	// FileExtension returns the file extension for the output (e.g., "css", "toml").
	FileExtension() string

	// DefaultPath returns the default output path for the generated file.
	// This is relative to the user's home directory or config directory.
	DefaultPath() string
}

// GenerateOptions holds options for output generation.
type GenerateOptions struct {
	// OutputPath is the path where the file should be written.
	// If empty, stdout is used.
	OutputPath string

	// Format specifies the output format variant (if plugin supports multiple).
	Format string

	// Overwrite determines whether to overwrite existing files.
	Overwrite bool

	// DryRun generates output without writing to disk.
	DryRun bool

	// BackupExisting creates a backup of existing files before overwriting.
	BackupExisting bool
}

// DefaultGenerateOptions returns the default generation options.
func DefaultGenerateOptions() GenerateOptions {
	return GenerateOptions{
		OutputPath:     "",
		Format:         "default",
		Overwrite:      false,
		DryRun:         false,
		BackupExisting: true,
	}
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

// All returns all registered plugins.
func (r *Registry) All() map[string]Plugin {
	// Return a copy to prevent external modification
	plugins := make(map[string]Plugin, len(r.plugins))
	for name, plugin := range r.plugins {
		plugins[name] = plugin
	}
	return plugins
}
