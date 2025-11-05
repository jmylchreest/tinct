// Package manager provides plugin management with configuration support.
package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/executor"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/file"
	"github.com/jmylchreest/tinct/internal/plugin/input/image"
	"github.com/jmylchreest/tinct/internal/plugin/input/remotecss"
	"github.com/jmylchreest/tinct/internal/plugin/input/remotejson"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/output/dunst"
	"github.com/jmylchreest/tinct/internal/plugin/output/fuzzel"
	"github.com/jmylchreest/tinct/internal/plugin/output/hyprland"
	"github.com/jmylchreest/tinct/internal/plugin/output/hyprlock"
	"github.com/jmylchreest/tinct/internal/plugin/output/hyprpaper"
	"github.com/jmylchreest/tinct/internal/plugin/output/kitty"
	"github.com/jmylchreest/tinct/internal/plugin/output/neovim"
	"github.com/jmylchreest/tinct/internal/plugin/output/swayosd"
	"github.com/jmylchreest/tinct/internal/plugin/output/waybar"
	"github.com/jmylchreest/tinct/internal/plugin/output/wofi"
	"github.com/jmylchreest/tinct/internal/plugin/output/zellij"
	"github.com/jmylchreest/tinct/internal/plugin/protocol"
)

const (
	versionUnknown = "unknown"
)

// Config holds plugin configuration.
type Config struct {
	// DisabledPlugins is a list of plugin names to disable.
	// Format: "plugin_type:plugin_name" (e.g., "output:tailwind", "input:image").
	DisabledPlugins []string

	// EnabledPlugins is a list of plugin names to explicitly enable.
	// If set, only these plugins are enabled (whitelist mode).
	EnabledPlugins []string
}

// Builder provides a fluent interface for constructing a Manager with configuration.
type Builder struct {
	config         Config
	inputRegistry  *input.Registry
	outputRegistry *output.Registry
	lockFilePath   string
	useEnv         bool
}

// NewBuilder creates a new Manager builder with default settings.
func NewBuilder() *Builder {
	return &Builder{
		config:         Config{},
		inputRegistry:  input.NewRegistry(),
		outputRegistry: output.NewRegistry(),
		useEnv:         false,
	}
}

// WithConfig sets the configuration for the manager.
func (b *Builder) WithConfig(config Config) *Builder {
	b.config = config
	return b
}

// WithEnvConfig loads configuration from environment variables.
// Reads TINCT_DISABLED_PLUGINS and TINCT_ENABLED_PLUGINS.
func (b *Builder) WithEnvConfig() *Builder {
	b.useEnv = true
	return b
}

// WithLockFile sets the path to a plugin lock file to load configuration from.
func (b *Builder) WithLockFile(path string) *Builder {
	b.lockFilePath = path
	return b
}

// WithCustomRegistries allows providing custom plugin registries (useful for testing).
func (b *Builder) WithCustomRegistries(inputReg *input.Registry, outputReg *output.Registry) *Builder {
	b.inputRegistry = inputReg
	b.outputRegistry = outputReg
	return b
}

// Build constructs the Manager with the configured settings.
// If both env and lock file are specified, lock file takes precedence.
func (b *Builder) Build() *Manager {
	// Start with base config.
	config := b.config

	// Apply env config if requested.
	if b.useEnv {
		if disabled := os.Getenv("TINCT_DISABLED_PLUGINS"); disabled != "" {
			config.DisabledPlugins = parsePluginList(disabled)
		}
		if enabled := os.Getenv("TINCT_ENABLED_PLUGINS"); enabled != "" {
			config.EnabledPlugins = parsePluginList(enabled)
		}
	}

	// Apply lock file config if specified (overrides env).
	// Note: Lock file loading is handled externally and updated via UpdateConfig.
	// The lockFilePath in the builder just signals that a lock file path was provided.

	m := &Manager{
		config:         config,
		inputRegistry:  b.inputRegistry,
		outputRegistry: b.outputRegistry,
	}

	// Register built-in plugins.
	m.registerBuiltinPlugins()

	return m
}

// Manager manages plugin enable/disable state and owns plugin registries.
type Manager struct {
	config         Config
	inputRegistry  *input.Registry
	outputRegistry *output.Registry
}

// registerBuiltinPlugins registers all built-in plugins.
func (m *Manager) registerBuiltinPlugins() {
	// Register input plugins.
	m.inputRegistry.Register(image.New())
	m.inputRegistry.Register(file.New())
	m.inputRegistry.Register(remotejson.New())
	m.inputRegistry.Register(remotecss.New())

	// Register output plugins.
	m.outputRegistry.Register(dunst.New())
	m.outputRegistry.Register(fuzzel.New())
	m.outputRegistry.Register(hyprland.New())
	m.outputRegistry.Register(hyprlock.New())
	m.outputRegistry.Register(hyprpaper.New())
	m.outputRegistry.Register(kitty.New())
	m.outputRegistry.Register(neovim.New())
	m.outputRegistry.Register(swayosd.New())
	m.outputRegistry.Register(waybar.New())
	m.outputRegistry.Register(wofi.New())
	m.outputRegistry.Register(zellij.New())
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

	// Check if "all" is explicitly disabled (takes precedence over everything).
	if slices.Contains(m.config.DisabledPlugins, "all") {
		return false
	}

	// Check if explicitly disabled.
	for _, disabled := range m.config.DisabledPlugins {
		if disabled == fullName || disabled == name {
			return false
		}
	}

	// Check if "all" is enabled (enables all plugins).
	if slices.Contains(m.config.EnabledPlugins, "all") {
		return true
	}

	// If whitelist mode (EnabledPlugins set), only listed plugins are enabled.
	if len(m.config.EnabledPlugins) > 0 {
		for _, enabled := range m.config.EnabledPlugins {
			if enabled == fullName || enabled == name {
				return true
			}
		}
		return false
	}

	// When no config is present (no enabled/disabled lists), all plugins are disabled by default.
	// This makes the plugin's Enabled() method irrelevant for both internal and external plugins.
	// Plugins must be explicitly enabled via config, CLI flags, or environment variables.
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
	// Validate plugin path - must be absolute and should exist.
	if !filepath.IsAbs(path) {
		return fmt.Errorf("plugin path must be absolute: %s", path)
	}

	// Check if the plugin file exists and is executable.
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("plugin not found or not accessible: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("plugin path is a directory, not a file: %s", path)
	}

	// Query plugin info to check protocol version.
	pluginInfo, err := queryPluginInfo(path)
	if err != nil {
		return fmt.Errorf("failed to query plugin info: %w", err)
	}

	// Check protocol version compatibility.
	if pluginInfo.ProtocolVersion != "" {
		compatible, err := protocol.IsCompatible(pluginInfo.ProtocolVersion)
		if err != nil || !compatible {
			errMsg := "unknown error"
			if err != nil {
				errMsg = err.Error()
			}
			return fmt.Errorf(
				"plugin '%s' protocol version %s is incompatible with tinct %s: %s",
				name,
				pluginInfo.ProtocolVersion,
				protocol.ProtocolVersion,
				errMsg,
			)
		}
	}
	// Note: If protocol_version is missing, we allow the plugin (backward compatibility)
	// but this should be warned about in verbose mode

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

// PluginInfo holds metadata returned by a plugin's --plugin-info command.
type PluginInfo struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	Version         string `json:"version"`
	ProtocolVersion string `json:"protocol_version"`
	Description     string `json:"description"`
}

// queryPluginInfo queries a plugin for its metadata.
func queryPluginInfo(pluginPath string) (PluginInfo, error) {
	cmd := exec.Command(pluginPath, "--plugin-info")
	output, err := cmd.Output()
	if err != nil {
		return PluginInfo{}, fmt.Errorf("failed to execute plugin: %w", err)
	}

	var info PluginInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return PluginInfo{}, fmt.Errorf("failed to parse plugin info: %w", err)
	}

	return info, nil
}

// parsePluginList parses a comma-separated list of plugin names.
// Handles formats like "tailwind", "output:tailwind", "input:image,output:tailwind".
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

	// Remove from enabled list if present.
	for i, enabled := range m.config.EnabledPlugins {
		if enabled == fullName || enabled == name {
			m.config.EnabledPlugins = append(m.config.EnabledPlugins[:i], m.config.EnabledPlugins[i+1:]...)
			break
		}
	}

	// Add to disabled list if not already there.
	if slices.Contains(m.config.DisabledPlugins, fullName) {
		return
	}
	m.config.DisabledPlugins = append(m.config.DisabledPlugins, fullName)
}

// SetEnabled adds a plugin to the enabled list (whitelist mode).
func (m *Manager) SetEnabled(pluginType, name string) {
	fullName := fmt.Sprintf("%s:%s", pluginType, name)

	// Remove from disabled list if present.
	for i, disabled := range m.config.DisabledPlugins {
		if disabled == fullName || disabled == name {
			m.config.DisabledPlugins = append(m.config.DisabledPlugins[:i], m.config.DisabledPlugins[i+1:]...)
			break
		}
	}

	// Add to enabled list if not already there.
	if slices.Contains(m.config.EnabledPlugins, fullName) {
		return
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

// Version returns the plugin's version.
// For external plugins, this queries the plugin executable.
func (p *ExternalInputPlugin) Version() string {
	info, err := queryPluginInfo(p.path)
	if err != nil {
		return versionUnknown
	}
	if info.Version == "" {
		return versionUnknown
	}
	return info.Version
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
// Uses the hybrid executor which automatically detects and uses the appropriate
// protocol (go-plugin RPC or JSON-stdio).
func (p *ExternalInputPlugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	// Create executor (detects protocol automatically).
	exec, err := executor.NewWithVerbose(p.path, opts.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin executor: %w", err)
	}
	defer exec.Close()

	// Merge plugin args from opts with plugin's own args.
	mergedArgs := make(map[string]any)
	maps.Copy(mergedArgs, p.args)
	maps.Copy(mergedArgs, opts.PluginArgs)

	// Convert to protocol format.
	protocolOpts := protocol.InputOptions{
		Verbose:         opts.Verbose,
		DryRun:          opts.DryRun || p.dryRun,
		ColourOverrides: opts.ColourOverrides,
		PluginArgs:      mergedArgs,
	}

	// Debug: show what's being sent to plugin.
	if opts.Verbose {
		optsJSON, _ := json.Marshal(protocolOpts)
		fmt.Fprintf(os.Stderr, "   Sending to plugin: %s\n", string(optsJSON))
	}

	// Execute input plugin.
	colors, err := exec.ExecuteInput(ctx, protocolOpts)
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	return colour.NewPalette(colors), nil
}

// RegisterFlags is a no-op for external plugins (they don't have flags).
func (p *ExternalInputPlugin) RegisterFlags(_ *cobra.Command) {
	// External plugins don't register flags in Tinct.
	// They handle their own arguments if needed.
}

// Validate checks if the plugin is valid.
func (p *ExternalInputPlugin) Validate() error {
	// Check if plugin file exists and is executable.
	// This is a basic check - the plugin might fail at runtime.
	return nil
}

// ExternalOutputPlugin wraps an external executable as an output plugin.
type ExternalOutputPlugin struct {
	name        string
	description string
	path        string
	args        map[string]any
	dryRun      bool
	verbose     bool
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

// Version returns the plugin's version.
// For external plugins, this queries the plugin executable.
func (p *ExternalOutputPlugin) Version() string {
	info, err := queryPluginInfo(p.path)
	if err != nil {
		return versionUnknown
	}
	if info.Version == "" {
		return versionUnknown
	}
	return info.Version
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

// SetVerbose sets the verbose flag for this plugin.
func (p *ExternalOutputPlugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetVerbose returns the verbose setting for this plugin.
func (p *ExternalOutputPlugin) GetVerbose() bool {
	return p.verbose
}

// Generate executes the external plugin and returns its output.
func (p *ExternalOutputPlugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	// Create executor (detects protocol automatically).
	exec, err := executor.NewWithVerbose(p.path, p.verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin executor: %w", err)
	}
	defer exec.Close()

	// Extract palette from themeData.
	palette := themeData.Palette()

	// Convert to protocol format.
	paletteData := convertCategorisedPaletteToProtocol(palette, p.args, p.dryRun)

	// Execute output plugin.
	files, err := exec.ExecuteOutput(context.Background(), paletteData)
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// If no files generated, return empty map.
	if len(files) == 0 {
		return make(map[string][]byte), nil
	}

	return files, nil
}

// RegisterFlags is a no-op for external plugins (they don't have flags).
func (p *ExternalOutputPlugin) RegisterFlags(_ *cobra.Command) {
	// External plugins don't register flags in Tinct.
	// They handle their own arguments if needed.
}

// Validate checks if the plugin is valid.
func (p *ExternalOutputPlugin) Validate() error {
	// Check if plugin file exists and is executable.
	// This is a basic check - the plugin might fail at runtime.
	return nil
}

// DefaultOutputDir returns the default output directory (not used for external plugins).
func (p *ExternalOutputPlugin) DefaultOutputDir() string {
	return "" // External plugins handle their own output
}

// PreExecute calls the external plugin's pre-execute hook.
// Implements the output.PreExecuteHook interface.
func (p *ExternalOutputPlugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Create executor (detects protocol automatically).
	exec, err := executor.NewWithVerbose(p.path, p.verbose)
	if err != nil {
		return false, "", fmt.Errorf("failed to create plugin executor: %w", err)
	}
	defer exec.Close()

	// Execute pre-execute hook.
	return exec.PreExecute(ctx)
}

// PostExecute calls the external plugin's post-execute hook.
// Implements the output.PostExecuteHook interface.
func (p *ExternalOutputPlugin) PostExecute(ctx context.Context, writtenFiles []string) error {
	// Create executor (detects protocol automatically).
	exec, err := executor.NewWithVerbose(p.path, p.verbose)
	if err != nil {
		return fmt.Errorf("failed to create plugin executor: %w", err)
	}
	defer exec.Close()

	// Execute post-execute hook.
	return exec.PostExecute(ctx, writtenFiles)
}

// GetExternalPluginInfo queries an external plugin for its metadata.
func GetExternalPluginInfo(path string) (name, description string, err error) {
	// #nosec G204 -- Path comes from validated plugin installation or user's explicit plugin add command
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
// Uses the hybrid executor which automatically detects and uses the appropriate
// protocol (go-plugin RPC or JSON-stdio).
// This is a utility function that may be used by external code.
func ExecuteExternalPlugin(ctx context.Context, path string, palette *colour.CategorisedPalette, pluginArgs map[string]any, dryRun bool) ([]byte, error) {
	// Create executor (detects protocol automatically).
	exec, err := executor.New(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin executor: %w", err)
	}
	defer exec.Close()

	// Convert palette to protocol format.
	paletteData := convertCategorisedPaletteToProtocol(palette, pluginArgs, dryRun)

	// Execute output plugin.
	files, err := exec.ExecuteOutput(ctx, paletteData)
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w", err)
	}

	// For backward compatibility, if the plugin generated files,
	// return them as JSON. If no files (like notification plugins),
	// return empty JSON object.
	if len(files) == 0 {
		return []byte("{}"), nil
	}

	// Return files as JSON map.
	result, err := json.Marshal(files)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plugin output: %w", err)
	}

	return result, nil
}

// convertCategorisedPaletteToProtocol converts a CategorisedPalette to protocol.PaletteData.
func convertCategorisedPaletteToProtocol(palette *colour.CategorisedPalette, pluginArgs map[string]any, dryRun bool) protocol.PaletteData {
	colours := make(map[string]protocol.CategorisedColour)
	for role, colour := range palette.Colours {
		colours[string(role)] = protocol.CategorisedColour{
			RGB: protocol.RGBColour{
				R: colour.RGB.R,
				G: colour.RGB.G,
				B: colour.RGB.B,
			},
			Hex:        colour.Hex,
			Role:       string(colour.Role),
			Luminance:  colour.Luminance,
			IsLight:    colour.IsLight,
			Hue:        colour.Hue,
			Saturation: colour.Saturation,
			Index:      colour.Index,
		}
	}

	allColours := make([]protocol.CategorisedColour, len(palette.AllColours))
	for i, colour := range palette.AllColours {
		allColours[i] = protocol.CategorisedColour{
			RGB: protocol.RGBColour{
				R: colour.RGB.R,
				G: colour.RGB.G,
				B: colour.RGB.B,
			},
			Hex:        colour.Hex,
			Role:       string(colour.Role),
			Luminance:  colour.Luminance,
			IsLight:    colour.IsLight,
			Hue:        colour.Hue,
			Saturation: colour.Saturation,
			Index:      colour.Index,
		}
	}

	themeType := "dark"
	if palette.ThemeType == colour.ThemeLight {
		themeType = "light"
	}

	return protocol.PaletteData{
		Colours:    colours,
		AllColours: allColours,
		ThemeType:  themeType,
		PluginArgs: pluginArgs,
		DryRun:     dryRun,
	}
}
