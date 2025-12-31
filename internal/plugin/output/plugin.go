// Package output provides the interface and base types for output plugins.
package output

import (
	"context"
	"maps"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/pkg/plugin"
)

// FlagHelp is re-exported from pkg/plugin for convenience.
// This allows output plugins to use output.FlagHelp instead of importing protocol directly.
type FlagHelp = plugin.FlagHelp

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

	// GetFlagHelp returns help information for plugin-specific flags.
	// This allows dynamic help generation based on selected plugins.
	GetFlagHelp() []FlagHelp
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

// VersionedPlugin is an optional interface that plugins can implement to support
// version-specific templates. When implemented, the template loader will select
// the most appropriate template based on the target application's version.
//
// This is useful for applications that change their configuration format between
// versions (e.g., Hyprland's windowrule syntax changed in v0.53).
//
// When this interface is implemented:
//   - GetTargetVersion() is called to get the installed application version
//   - Templates can be organized in version subdirectories (e.g., templates/0.53/)
//   - The loader selects the highest version that doesn't exceed the target version
//   - Falls back to default templates if no suitable versioned template is found
//
// Example directory structure:
//
//	internal/plugin/output/hyprland/
//	├── tinct.conf.tmpl           # Default template (for older versions)
//	└── templates/
//	    └── 0.53/
//	        └── tinct.conf.tmpl   # Template for v0.53+
type VersionedPlugin interface {
	// GetTargetVersion returns the version of the target application.
	// Returns an empty string if the version cannot be determined.
	// The version string should be in semantic version format (e.g., "0.53.0").
	GetTargetVersion() string
}

// DualThemePlugin is an optional interface that plugins can implement to support
// generating both light and dark theme variants in a single execution.
//
// This is useful for desktop environments and applications that automatically
// switch themes based on system preference (e.g., KDE Plasma, Qt applications).
//
// When this interface is implemented:
//   - GenerateDualTheme() is called instead of Generate()
//   - The plugin receives both the primary theme and its alternate variant
//   - The plugin can generate configuration files for both themes at once
//
// Example: KDE Plasma generates both TinctDark.colors and TinctLight.colors,
// allowing seamless switching via system settings.
type DualThemePlugin interface {
	// GenerateDualTheme creates output files for both light and dark themes.
	// primaryTheme is the main theme (based on input source).
	// alternateTheme is the opposite variant (light if primary is dark, vice versa).
	// Returns map of filename -> content for all generated files.
	GenerateDualTheme(primaryTheme, alternateTheme *colour.ThemeData) (map[string][]byte, error)
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
