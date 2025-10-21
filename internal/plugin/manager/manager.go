// Package manager provides plugin management with configuration support.
package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"strings"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/file"
	"github.com/jmylchreest/tinct/internal/plugin/input/image"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/hyprland"
	"github.com/jmylchreest/tinct/internal/plugin/output/tailwind"
	"github.com/spf13/cobra"
)

// Config holds plugin configuration.
type Config struct {
	// DisabledPlugins is a list of plugin names to disable.
	// Format: "plugin_type:plugin_name" (e.g., "output:tailwind", "input:image")
	DisabledPlugins []string

	// EnabledPlugins is a list of plugin names to explicitly enable.
	// If set, only these plugins are enabled (whitelist mode).
	EnabledPlugins []string
}

// Manager manages plugin enable/disable state and owns plugin registries.
type Manager struct {
	config         Config
	inputRegistry  *input.Registry
	outputRegistry *output.Registry
}

// New creates a new plugin manager with the given configuration.
// Automatically registers built-in plugins.
func New(config Config) *Manager {
	m := &Manager{
		config:         config,
		inputRegistry:  input.NewRegistry(),
		outputRegistry: output.NewRegistry(),
	}
	m.registerBuiltinPlugins()
	return m
}

// NewWithRegistries creates a plugin manager with custom registries (for testing).
func NewWithRegistries(config Config, inputReg *input.Registry, outputReg *output.Registry) *Manager {
	return &Manager{
		config:         config,
		inputRegistry:  inputReg,
		outputRegistry: outputReg,
	}
}

// NewFromEnv creates a plugin manager from environment variables.
// Reads TINCT_DISABLED_PLUGINS and TINCT_ENABLED_PLUGINS.
// Format: comma-separated list like "output:tailwind,input:file"
func NewFromEnv() *Manager {
	config := Config{}

	// Read disabled plugins from env
	if disabled := os.Getenv("TINCT_DISABLED_PLUGINS"); disabled != "" {
		config.DisabledPlugins = parsePluginList(disabled)
	}

	// Read enabled plugins from env
	if enabled := os.Getenv("TINCT_ENABLED_PLUGINS"); enabled != "" {
		config.EnabledPlugins = parsePluginList(enabled)
	}

	m := New(config)
	return m
}

// registerBuiltinPlugins registers all built-in plugins.
func (m *Manager) registerBuiltinPlugins() {
	// Register input plugins
	m.inputRegistry.Register(image.New())
	m.inputRegistry.Register(file.New())

	// Register output plugins
	m.outputRegistry.Register(hyprland.New())
	m.outputRegistry.Register(tailwind.New())
}

// InputRegistry returns the input plugin registry.
func (m *Manager) InputRegistry() *input.Registry {
	return m.inputRegistry
}

// OutputRegistry returns the output plugin registry.
func (m *Manager) OutputRegistry() *output.Registry {
	return m.outputRegistry
}

// GetInputPlugin retrieves an input plugin by name.
func (m *Manager) GetInputPlugin(name string) (input.Plugin, bool) {
	plugin, ok := m.inputRegistry.Get(name)
	return plugin, ok
}

// GetOutputPlugin retrieves an output plugin by name.
func (m *Manager) GetOutputPlugin(name string) (output.Plugin, bool) {
	plugin, ok := m.outputRegistry.Get(name)
	return plugin, ok
}

// IsInputEnabled checks if an input plugin is enabled.
// All plugins are disabled by default and must be explicitly enabled.
func (m *Manager) IsInputEnabled(plugin input.Plugin) bool {
	return m.isEnabled("input", plugin.Name())
}

// IsOutputEnabled checks if an output plugin is enabled.
// All plugins are disabled by default and must be explicitly enabled.
func (m *Manager) IsOutputEnabled(plugin output.Plugin) bool {
	return m.isEnabled("output", plugin.Name())
}

// isEnabled determines if a plugin is enabled based on configuration.
func (m *Manager) isEnabled(pluginType, name string) bool {
	fullName := fmt.Sprintf("%s:%s", pluginType, name)

	// Check if "all" is explicitly disabled (takes precedence over everything)
	for _, disabled := range m.config.DisabledPlugins {
		if disabled == "all" {
			return false
		}
	}

	// Check if explicitly disabled
	for _, disabled := range m.config.DisabledPlugins {
		if disabled == fullName || disabled == name {
			return false
		}
	}

	// Check if "all" is enabled (enables all plugins)
	for _, enabled := range m.config.EnabledPlugins {
		if enabled == "all" {
			return true
		}
	}

	// If whitelist mode (EnabledPlugins set), only listed plugins are enabled
	if len(m.config.EnabledPlugins) > 0 {
		for _, enabled := range m.config.EnabledPlugins {
			if enabled == fullName || enabled == name {
				return true
			}
		}
		return false
	}

	// When no config is present (no enabled/disabled lists), all plugins are disabled by default
	// This makes the plugin's Enabled() method irrelevant for both internal and external plugins
	// Plugins must be explicitly enabled via config, CLI flags, or environment variables
	return false
}

// FilterInputPlugins returns only enabled input plugins.
func (m *Manager) FilterInputPlugins() map[string]input.Plugin {
	enabled := make(map[string]input.Plugin)
	for name, plugin := range m.inputRegistry.All() {
		if m.IsInputEnabled(plugin) {
			enabled[name] = plugin
		}
	}
	return enabled
}

// FilterOutputPlugins returns only enabled output plugins.
func (m *Manager) FilterOutputPlugins() map[string]output.Plugin {
	enabled := make(map[string]output.Plugin)
	for name, plugin := range m.outputRegistry.All() {
		if m.IsOutputEnabled(plugin) {
			enabled[name] = plugin
		}
	}
	return enabled
}

// ListInputPlugins returns names of enabled input plugins.
func (m *Manager) ListInputPlugins() []string {
	names := []string{}
	for name, plugin := range m.inputRegistry.All() {
		if m.IsInputEnabled(plugin) {
			names = append(names, name)
		}
	}
	return names
}

// ListOutputPlugins returns names of enabled output plugins.
func (m *Manager) ListOutputPlugins() []string {
	names := []string{}
	for name, plugin := range m.outputRegistry.All() {
		if m.IsOutputEnabled(plugin) {
			names = append(names, name)
		}
	}
	return names
}

// AllInputPlugins returns all registered input plugins (including disabled).
func (m *Manager) AllInputPlugins() map[string]input.Plugin {
	return m.inputRegistry.All()
}

// AllOutputPlugins returns all registered output plugins (including disabled).
func (m *Manager) AllOutputPlugins() map[string]output.Plugin {
	return m.outputRegistry.All()
}

// UpdateConfig updates the manager's configuration without recreating plugin instances.
// This preserves flag bindings and other plugin state.
func (m *Manager) UpdateConfig(config Config) {
	m.config = config
}

// RegisterExternalPlugin registers an external plugin with the manager.
func (m *Manager) RegisterExternalPlugin(name, pluginType, path, description string) error {
	switch pluginType {
	case "output":
		plugin := NewExternalOutputPlugin(name, description, path)
		m.outputRegistry.Register(plugin)
		return nil
	case "input":
		plugin := NewExternalInputPlugin(name, description, path)
		m.inputRegistry.Register(plugin)
		return nil
	default:
		return fmt.Errorf("unknown plugin type: %s", pluginType)
	}
}

// parsePluginList parses a comma-separated list of plugin names.
// Handles formats like "tailwind", "output:tailwind", "input:image,output:tailwind"
func parsePluginList(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// GetConfig returns the current configuration.
func (m *Manager) GetConfig() Config {
	return m.config
}

// SetDisabled adds a plugin to the disabled list.
func (m *Manager) SetDisabled(pluginType, name string) {
	fullName := fmt.Sprintf("%s:%s", pluginType, name)

	// Remove from enabled list if present
	for i, enabled := range m.config.EnabledPlugins {
		if enabled == fullName || enabled == name {
			m.config.EnabledPlugins = append(m.config.EnabledPlugins[:i], m.config.EnabledPlugins[i+1:]...)
			break
		}
	}

	// Add to disabled list if not already there
	for _, disabled := range m.config.DisabledPlugins {
		if disabled == fullName {
			return
		}
	}
	m.config.DisabledPlugins = append(m.config.DisabledPlugins, fullName)
}

// SetEnabled adds a plugin to the enabled list (whitelist mode).
func (m *Manager) SetEnabled(pluginType, name string) {
	fullName := fmt.Sprintf("%s:%s", pluginType, name)

	// Remove from disabled list if present
	for i, disabled := range m.config.DisabledPlugins {
		if disabled == fullName || disabled == name {
			m.config.DisabledPlugins = append(m.config.DisabledPlugins[:i], m.config.DisabledPlugins[i+1:]...)
			break
		}
	}

	// Add to enabled list if not already there
	for _, enabled := range m.config.EnabledPlugins {
		if enabled == fullName {
			return
		}
	}
	m.config.EnabledPlugins = append(m.config.EnabledPlugins, fullName)
}

// ExternalInputPlugin wraps an external executable as an input plugin.
type ExternalInputPlugin struct {
	name        string
	description string
	path        string
	args        map[string]any
	dryRun      bool
}

// NewExternalInputPlugin creates a new external input plugin wrapper.
func NewExternalInputPlugin(name, description, path string) *ExternalInputPlugin {
	return &ExternalInputPlugin{
		name:        name,
		description: description,
		path:        path,
	}
}

// Name returns the plugin's name.
func (p *ExternalInputPlugin) Name() string {
	return p.name
}

// Description returns the plugin's description.
func (p *ExternalInputPlugin) Description() string {
	return p.description
}

// SetArgs sets custom arguments for this plugin.
func (p *ExternalInputPlugin) SetArgs(args map[string]any) {
	p.args = args
}

// GetArgs returns custom arguments for this plugin.
func (p *ExternalInputPlugin) GetArgs() map[string]any {
	return p.args
}

// SetDryRun sets the dry-run mode for this plugin.
func (p *ExternalInputPlugin) SetDryRun(dryRun bool) {
	p.dryRun = dryRun
}

// GetDryRun returns the dry-run mode for this plugin.
func (p *ExternalInputPlugin) GetDryRun() bool {
	return p.dryRun
}

// Generate executes the external plugin and returns a palette.
func (p *ExternalInputPlugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	// Create extended payload with plugin args and dry-run flag
	type ExtendedInputOptions struct {
		Verbose         bool           `json:"verbose"`
		DryRun          bool           `json:"dry_run"`
		ColourOverrides []string       `json:"colour_overrides,omitempty"`
		PluginArgs      map[string]any `json:"plugin_args,omitempty"`
	}

	// Merge plugin args from opts with plugin's own args
	mergedArgs := make(map[string]any)
	for k, v := range p.args {
		mergedArgs[k] = v
	}
	for k, v := range opts.PluginArgs {
		mergedArgs[k] = v
	}

	extended := ExtendedInputOptions{
		Verbose:         opts.Verbose,
		DryRun:          opts.DryRun || p.dryRun,
		ColourOverrides: opts.ColourOverrides,
		PluginArgs:      mergedArgs,
	}

	// Convert to JSON
	optsJSON, err := json.Marshal(extended)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	// Debug: show what's being sent to plugin
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "  └─ Sending to plugin: %s\n", string(optsJSON))
	}

	// Execute external plugin
	cmd := exec.CommandContext(ctx, p.path)
	cmd.Stdin = bytes.NewReader(optsJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w\nStderr: %s", err, stderr.String())
	}

	// Parse the output - try simple color array first, then categorised format for backwards compatibility
	var rawColors []struct {
		R uint8 `json:"r"`
		G uint8 `json:"g"`
		B uint8 `json:"b"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &rawColors); err == nil {
		// Simple color array format
		colors := make([]color.Color, len(rawColors))
		for i, rgb := range rawColors {
			colors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
		}
		return colour.NewPalette(colors), nil
	}

	// Try categorised palette format (backwards compatibility)
	var categorised colour.CategorisedPalette
	if err := json.Unmarshal(stdout.Bytes(), &categorised); err == nil {
		// Extract colors from categorised palette
		colors := make([]color.Color, 0)

		// Get colors from AllColours if available
		if len(categorised.AllColours) > 0 {
			for _, cc := range categorised.AllColours {
				colors = append(colors, color.RGBA{
					R: cc.RGB.R,
					G: cc.RGB.G,
					B: cc.RGB.B,
					A: 255,
				})
			}
		} else {
			// Fallback to colors from Colours map
			for _, cc := range categorised.Colours {
				colors = append(colors, color.RGBA{
					R: cc.RGB.R,
					G: cc.RGB.G,
					B: cc.RGB.B,
					A: 255,
				})
			}
		}

		return colour.NewPalette(colors), nil
	}

	return nil, fmt.Errorf("failed to parse plugin output as color array or categorised palette\nOutput: %s", stdout.String())
}

// RegisterFlags is a no-op for external plugins (they don't have flags).
func (p *ExternalInputPlugin) RegisterFlags(cmd *cobra.Command) {
	// External plugins don't register flags in Tinct
	// They handle their own arguments if needed
}

// Validate checks if the plugin is valid.
func (p *ExternalInputPlugin) Validate() error {
	// Check if plugin file exists and is executable
	// This is a basic check - the plugin might fail at runtime
	return nil
}

// ExternalOutputPlugin wraps an external executable as an output plugin.
type ExternalOutputPlugin struct {
	name        string
	description string
	path        string
	args        map[string]any
	dryRun      bool
}

// NewExternalOutputPlugin creates a new external output plugin wrapper.
func NewExternalOutputPlugin(name, description, path string) *ExternalOutputPlugin {
	return &ExternalOutputPlugin{
		name:        name,
		description: description,
		path:        path,
	}
}

// Name returns the plugin's name.
func (p *ExternalOutputPlugin) Name() string {
	return p.name
}

// Description returns the plugin's description.
func (p *ExternalOutputPlugin) Description() string {
	return p.description
}

// SetArgs sets custom arguments for this plugin.
func (p *ExternalOutputPlugin) SetArgs(args map[string]any) {
	p.args = args
}

// GetArgs returns custom arguments for this plugin.
func (p *ExternalOutputPlugin) GetArgs() map[string]any {
	return p.args
}

// SetDryRun sets the dry-run mode for this plugin.
func (p *ExternalOutputPlugin) SetDryRun(dryRun bool) {
	p.dryRun = dryRun
}

// GetDryRun returns the dry-run mode for this plugin.
func (p *ExternalOutputPlugin) GetDryRun() bool {
	return p.dryRun
}

// Generate executes the external plugin and returns its output.
func (p *ExternalOutputPlugin) Generate(palette *colour.CategorisedPalette) (map[string][]byte, error) {
	// Create extended payload with plugin args and dry-run flag
	type ExtendedPalette struct {
		Colours    map[colour.ColourRole]colour.CategorisedColour `json:"colours"`
		AllColours []colour.CategorisedColour                     `json:"all_colours"`
		ThemeType  colour.ThemeType                               `json:"theme_type"`
		PluginArgs map[string]any                                 `json:"plugin_args,omitempty"`
		DryRun     bool                                           `json:"dry_run"`
	}

	extended := ExtendedPalette{
		Colours:    palette.Colours,
		AllColours: palette.AllColours,
		ThemeType:  palette.ThemeType,
		PluginArgs: p.args,
		DryRun:     p.dryRun,
	}

	// Convert to JSON
	paletteJSON, err := json.Marshal(extended)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal palette: %w", err)
	}

	// Execute external plugin
	cmd := exec.Command(p.path)
	cmd.Stdin = bytes.NewReader(paletteJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w\nStderr: %s", err, stderr.String())
	}

	// External plugins output to stdout, not files
	// Return stdout as a virtual file for display purposes
	result := make(map[string][]byte)
	if stdout.Len() > 0 {
		result[p.name+"-output.txt"] = stdout.Bytes()
	}

	return result, nil
}

// RegisterFlags is a no-op for external plugins (they don't have flags).
func (p *ExternalOutputPlugin) RegisterFlags(cmd *cobra.Command) {
	// External plugins don't register flags in Tinct
	// They handle their own arguments if needed
}

// Validate checks if the plugin is valid.
func (p *ExternalOutputPlugin) Validate() error {
	// Check if plugin file exists and is executable
	// This is a basic check - the plugin might fail at runtime
	return nil
}

// DefaultOutputDir returns the default output directory (not used for external plugins).
func (p *ExternalOutputPlugin) DefaultOutputDir() string {
	return "" // External plugins handle their own output
}

// GetExternalPluginInfo queries an external plugin for its metadata.
func GetExternalPluginInfo(path string) (name, description string, err error) {
	cmd := exec.Command(path, "--plugin-info")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get plugin info: %w", err)
	}

	var info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return "", "", fmt.Errorf("failed to parse plugin info: %w", err)
	}

	return info.Name, info.Description, nil
}

// ExecuteExternalPlugin runs an external plugin with the given palette.
func ExecuteExternalPlugin(ctx context.Context, path string, palette *colour.CategorisedPalette) ([]byte, error) {
	// Convert palette to JSON
	paletteJSON, err := palette.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal palette: %w", err)
	}

	// Execute external plugin
	cmd := exec.CommandContext(ctx, path)
	cmd.Stdin = bytes.NewReader(paletteJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w\nStderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
