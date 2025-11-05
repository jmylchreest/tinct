// Package output provides the interface and base types for output plugins.
package output

import (
	"context"
	"maps"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
)

// Plugin represents an output plugin that can generate configuration files.
// from a categorised colour palette.
type Plugin interface {
	// Name returns the plugin's name (e.g., "tailwind", "hyprland").
	Name() string

	// Description returns a human-readable description of the plugin.
	Description() string

	// Version returns the plugin version (e.g., "1.0.0").
	Version() string

	// Generate creates output file(s) from the given theme data.
	// Returns map of filename -> content to support plugins that generate multiple files.
	Generate(themeData *colour.ThemeData) (map[string][]byte, error)

	// RegisterFlags registers plugin-specific flags with cobra command.
	RegisterFlags(cmd *cobra.Command)

	// Validate checks if the plugin configuration is valid.
	Validate() error

	// DefaultOutputDir returns the default output directory for this plugin.
	DefaultOutputDir() string
}

// PreExecuteHook is an optional interface that plugins can implement to perform.
// checks before generation. If the hook returns an error or skip=true, the plugin
// will be skipped without generating output.
//
// Common use cases:.
//   - Check if required executables exist on $PATH
//   - Verify configuration directories exist
//   - Validate environment prerequisites
type PreExecuteHook interface {
	// PreExecute runs before Generate(). Returns:.
	//   - skip: if true, plugin is skipped (not an error, just bypassed)
	//   - reason: human-readable explanation for skipping
	//   - error: actual error that should stop execution
	PreExecute(ctx context.Context) (skip bool, reason string, err error)
}

// PostExecuteHook is an optional interface that plugins can implement to perform.
// actions after successful file generation.
//
// Common use cases:.
//   - Reload application configuration
//   - Send signals to running processes
//   - Restart services
//   - Notify users of changes
//   - Set wallpaper (when --set-wallpaper flag is used)
type PostExecuteHook interface {
	// PostExecute runs after successful Generate() and file writing.
	// The execCtx contains execution context including wallpaper path if available.
	// The writtenFiles contains the paths that were written.
	// Errors are logged but don't fail the overall operation.
	PostExecute(ctx context.Context, execCtx ExecutionContext, writtenFiles []string) error
}

// VerbosePlugin is an optional interface that plugins can implement to receive.
// verbose logging settings from the CLI.
type VerbosePlugin interface {
	// SetVerbose enables or disables verbose logging for the plugin.
	SetVerbose(verbose bool)
}

// TemplateProvider is an optional interface that plugins can implement to expose.
// their embedded template filesystem for template management commands.
type TemplateProvider interface {
	// GetEmbeddedFS returns the embedded filesystem containing template files.
	// This should return the same embed.FS used by the plugin's generateTheme method.
	GetEmbeddedFS() any
}

// ExecutionContext provides context for hook execution.
type ExecutionContext struct {
	DryRun        bool   // Whether this is a dry-run
	Verbose       bool   // Whether verbose output is enabled
	OutputDir     string // The output directory being used
	WallpaperPath string // Optional path to source wallpaper (from input plugin)
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
	// Return a copy to prevent external modification.
	plugins := make(map[string]Plugin, len(r.plugins))
	maps.Copy(plugins, r.plugins)
	return plugins
}
