package manager

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/output"
)

// Mock input plugin for testing.
type mockInputPlugin struct {
	name        string
	description string
	version     string
}

func (m *mockInputPlugin) Name() string        { return m.name }
func (m *mockInputPlugin) Description() string { return m.description }
func (m *mockInputPlugin) Version() string     { return m.version }
func (m *mockInputPlugin) Generate(_ context.Context, _ input.GenerateOptions) (*colour.Palette, error) {
	return nil, nil
}
func (m *mockInputPlugin) RegisterFlags(_ *cobra.Command) {}
func (m *mockInputPlugin) Validate() error                { return nil }
func (m *mockInputPlugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{}
}

// Mock output plugin for testing.
type mockOutputPlugin struct {
	name        string
	description string
	version     string
}

func (m *mockOutputPlugin) Name() string        { return m.name }
func (m *mockOutputPlugin) Description() string { return m.description }
func (m *mockOutputPlugin) Version() string     { return m.version }
func (m *mockOutputPlugin) Generate(_ *colour.ThemeData) (map[string][]byte, error) {
	return nil, nil
}
func (m *mockOutputPlugin) DefaultOutputDir() string       { return "" }
func (m *mockOutputPlugin) RegisterFlags(_ *cobra.Command) {}
func (m *mockOutputPlugin) Validate() error                { return nil }
func (m *mockOutputPlugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{}
}

// TestNewBuilder tests the builder constructor.
func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("NewBuilder returned nil")
	}
	if builder.inputRegistry == nil {
		t.Error("inputRegistry not initialized")
	}
	if builder.outputRegistry == nil {
		t.Error("outputRegistry not initialized")
	}
	if builder.useEnv {
		t.Error("useEnv should default to false")
	}
}

// TestBuilderWithConfig tests the WithConfig builder method.
func TestBuilderWithConfig(t *testing.T) {
	config := Config{
		DisabledPlugins: []string{"output:test"},
		EnabledPlugins:  []string{"input:image"},
	}

	builder := NewBuilder().WithConfig(config)
	if len(builder.config.DisabledPlugins) != 1 {
		t.Errorf("Expected 1 disabled plugin, got %d", len(builder.config.DisabledPlugins))
	}
	if len(builder.config.EnabledPlugins) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(builder.config.EnabledPlugins))
	}
}

// TestBuilderWithEnvConfig tests environment variable configuration.
func TestBuilderWithEnvConfig(t *testing.T) {
	// Set test environment variables.
	os.Setenv("TINCT_DISABLED_PLUGINS", "output:kitty,output:waybar")
	os.Setenv("TINCT_ENABLED_PLUGINS", "input:image")
	defer func() {
		os.Unsetenv("TINCT_DISABLED_PLUGINS")
		os.Unsetenv("TINCT_ENABLED_PLUGINS")
	}()

	manager := NewBuilder().WithEnvConfig().Build()

	if len(manager.config.DisabledPlugins) != 2 {
		t.Errorf("Expected 2 disabled plugins, got %d", len(manager.config.DisabledPlugins))
	}
	if len(manager.config.EnabledPlugins) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(manager.config.EnabledPlugins))
	}
}

// TestBuilderWithCustomRegistries tests custom registry injection.
func TestBuilderWithCustomRegistries(t *testing.T) {
	inputReg := input.NewRegistry()
	outputReg := output.NewRegistry()

	mock := &mockInputPlugin{name: "test", description: "Test plugin", version: "1.0.0"}
	inputReg.Register(mock)

	builder := NewBuilder().WithCustomRegistries(inputReg, outputReg)
	manager := builder.Build()

	// Note: Build() also registers built-in plugins, so we check for our test plugin specifically.
	plugin, ok := manager.GetInputPlugin("test")
	if !ok {
		t.Error("Failed to get test plugin")
	}
	if plugin.Name() != "test" {
		t.Errorf("Expected plugin name 'test', got '%s'", plugin.Name())
	}
}

// TestBuildRegistersBuiltinPlugins tests that Build registers built-in plugins.
func TestBuildRegistersBuiltinPlugins(t *testing.T) {
	manager := NewBuilder().Build()

	// Check that built-in input plugins are registered.
	inputPlugins := manager.AllInputPlugins()
	if len(inputPlugins) == 0 {
		t.Error("No input plugins registered")
	}

	// Check for specific built-in plugins.
	expectedInputs := []string{"image", "file", "remote-json", "remote-css", "google-genai"}
	for _, name := range expectedInputs {
		if _, ok := manager.GetInputPlugin(name); !ok {
			t.Errorf("Built-in input plugin '%s' not registered", name)
		}
	}

	// Check that built-in output plugins are registered.
	outputPlugins := manager.AllOutputPlugins()
	if len(outputPlugins) == 0 {
		t.Error("No output plugins registered")
	}

	// Check for specific built-in plugins.
	expectedOutputs := []string{"kitty", "waybar", "fuzzel", "hyprland"}
	for _, name := range expectedOutputs {
		if _, ok := manager.GetOutputPlugin(name); !ok {
			t.Errorf("Built-in output plugin '%s' not registered", name)
		}
	}
}

// TestIsEnabledDefault tests the default behavior (all plugins disabled).
func TestIsEnabledDefault(t *testing.T) {
	manager := NewBuilder().Build()

	// Create mock plugins.
	inputPlugin := &mockInputPlugin{name: "test-input"}
	outputPlugin := &mockOutputPlugin{name: "test-output"}

	// By default, all plugins are disabled.
	if manager.IsInputEnabled(inputPlugin) {
		t.Error("Input plugin should be disabled by default")
	}
	if manager.IsOutputEnabled(outputPlugin) {
		t.Error("Output plugin should be disabled by default")
	}
}

// TestIsEnabledWhitelist tests whitelist mode (EnabledPlugins set).
func TestIsEnabledWhitelist(t *testing.T) {
	config := Config{
		EnabledPlugins: []string{"input:test-input", "output:test-output"},
	}
	manager := NewBuilder().WithConfig(config).Build()

	// Create mock plugins.
	enabledInput := &mockInputPlugin{name: "test-input"}
	disabledInput := &mockInputPlugin{name: "other-input"}
	enabledOutput := &mockOutputPlugin{name: "test-output"}
	disabledOutput := &mockOutputPlugin{name: "other-output"}

	// Check enabled plugins.
	if !manager.IsInputEnabled(enabledInput) {
		t.Error("Explicitly enabled input plugin should be enabled")
	}
	if !manager.IsOutputEnabled(enabledOutput) {
		t.Error("Explicitly enabled output plugin should be enabled")
	}

	// Check disabled plugins.
	if manager.IsInputEnabled(disabledInput) {
		t.Error("Non-whitelisted input plugin should be disabled")
	}
	if manager.IsOutputEnabled(disabledOutput) {
		t.Error("Non-whitelisted output plugin should be disabled")
	}
}

// TestIsEnabledDisabledList tests explicit disable list.
func TestIsEnabledDisabledList(t *testing.T) {
	config := Config{
		EnabledPlugins:  []string{"all"},
		DisabledPlugins: []string{"output:test-output"},
	}
	manager := NewBuilder().WithConfig(config).Build()

	// Create mock plugins.
	enabledInput := &mockInputPlugin{name: "test-input"}
	disabledOutput := &mockOutputPlugin{name: "test-output"}
	enabledOutput := &mockOutputPlugin{name: "other-output"}

	// Check that "all" enables everything.
	if !manager.IsInputEnabled(enabledInput) {
		t.Error("Input plugin should be enabled with 'all'")
	}

	// Check that explicit disable overrides "all".
	if manager.IsOutputEnabled(disabledOutput) {
		t.Error("Explicitly disabled output plugin should be disabled")
	}
	if !manager.IsOutputEnabled(enabledOutput) {
		t.Error("Non-disabled output plugin should be enabled with 'all'")
	}
}

// TestIsEnabledDisableAll tests that "all" in DisabledPlugins disables everything.
func TestIsEnabledDisableAll(t *testing.T) {
	config := Config{
		DisabledPlugins: []string{"all"},
		EnabledPlugins:  []string{"input:test-input"}, // Should be ignored
	}
	manager := NewBuilder().WithConfig(config).Build()

	// Create mock plugins.
	inputPlugin := &mockInputPlugin{name: "test-input"}
	outputPlugin := &mockOutputPlugin{name: "test-output"}

	// "all" in DisabledPlugins should disable everything.
	if manager.IsInputEnabled(inputPlugin) {
		t.Error("Input plugin should be disabled with 'all' in DisabledPlugins")
	}
	if manager.IsOutputEnabled(outputPlugin) {
		t.Error("Output plugin should be disabled with 'all' in DisabledPlugins")
	}
}

// TestFilterInputPlugins tests filtering enabled input plugins.
func TestFilterInputPlugins(t *testing.T) {
	inputReg := input.NewRegistry()
	outputReg := output.NewRegistry()

	// Register mock plugins.
	inputReg.Register(&mockInputPlugin{name: "test1"})
	inputReg.Register(&mockInputPlugin{name: "test2"})

	config := Config{
		EnabledPlugins: []string{"input:test1"},
	}

	manager := NewBuilder().
		WithConfig(config).
		WithCustomRegistries(inputReg, outputReg).
		Build()

	enabled := manager.FilterInputPlugins()

	if len(enabled) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(enabled))
	}
	if _, ok := enabled["test1"]; !ok {
		t.Error("test1 should be in enabled plugins")
	}
	if _, ok := enabled["test2"]; ok {
		t.Error("test2 should not be in enabled plugins")
	}
}

// TestFilterOutputPlugins tests filtering enabled output plugins.
func TestFilterOutputPlugins(t *testing.T) {
	inputReg := input.NewRegistry()
	outputReg := output.NewRegistry()

	// Register mock plugins.
	outputReg.Register(&mockOutputPlugin{name: "test1"})
	outputReg.Register(&mockOutputPlugin{name: "test2"})

	config := Config{
		EnabledPlugins: []string{"output:test1"},
	}

	manager := NewBuilder().
		WithConfig(config).
		WithCustomRegistries(inputReg, outputReg).
		Build()

	enabled := manager.FilterOutputPlugins()

	if len(enabled) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(enabled))
	}
	if _, ok := enabled["test1"]; !ok {
		t.Error("test1 should be in enabled plugins")
	}
	if _, ok := enabled["test2"]; ok {
		t.Error("test2 should not be in enabled plugins")
	}
}

// TestListInputPlugins tests listing enabled input plugin names.
func TestListInputPlugins(t *testing.T) {
	inputReg := input.NewRegistry()
	outputReg := output.NewRegistry()

	inputReg.Register(&mockInputPlugin{name: "test1"})
	inputReg.Register(&mockInputPlugin{name: "test2"})

	config := Config{
		EnabledPlugins: []string{"input:test1"},
	}

	manager := NewBuilder().
		WithConfig(config).
		WithCustomRegistries(inputReg, outputReg).
		Build()

	names := manager.ListInputPlugins()

	if len(names) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(names))
	}
	if names[0] != "test1" {
		t.Errorf("Expected 'test1', got '%s'", names[0])
	}
}

// TestListOutputPlugins tests listing enabled output plugin names.
func TestListOutputPlugins(t *testing.T) {
	inputReg := input.NewRegistry()
	outputReg := output.NewRegistry()

	outputReg.Register(&mockOutputPlugin{name: "test1"})
	outputReg.Register(&mockOutputPlugin{name: "test2"})

	config := Config{
		EnabledPlugins: []string{"output:test1"},
	}

	manager := NewBuilder().
		WithConfig(config).
		WithCustomRegistries(inputReg, outputReg).
		Build()

	names := manager.ListOutputPlugins()

	if len(names) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(names))
	}
	if names[0] != "test1" {
		t.Errorf("Expected 'test1', got '%s'", names[0])
	}
}

// TestUpdateConfig tests updating configuration after creation.
func TestUpdateConfig(t *testing.T) {
	manager := NewBuilder().Build()

	// Initially no plugins enabled.
	inputPlugin := &mockInputPlugin{name: "test"}
	if manager.IsInputEnabled(inputPlugin) {
		t.Error("Plugin should be disabled initially")
	}

	// Update config to enable the plugin.
	newConfig := Config{
		EnabledPlugins: []string{"input:test"},
	}
	manager.UpdateConfig(newConfig)

	if !manager.IsInputEnabled(inputPlugin) {
		t.Error("Plugin should be enabled after config update")
	}
}

// TestSetEnabled tests enabling a plugin.
func TestSetEnabled(t *testing.T) {
	manager := NewBuilder().Build()

	manager.SetEnabled("input", "test")

	if len(manager.config.EnabledPlugins) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(manager.config.EnabledPlugins))
	}
	if manager.config.EnabledPlugins[0] != "input:test" {
		t.Errorf("Expected 'input:test', got '%s'", manager.config.EnabledPlugins[0])
	}
}

// TestSetDisabled tests disabling a plugin.
func TestSetDisabled(t *testing.T) {
	config := Config{
		EnabledPlugins: []string{"input:test"},
	}
	manager := NewBuilder().WithConfig(config).Build()

	manager.SetDisabled("input", "test")

	if len(manager.config.DisabledPlugins) != 1 {
		t.Errorf("Expected 1 disabled plugin, got %d", len(manager.config.DisabledPlugins))
	}
	if len(manager.config.EnabledPlugins) != 0 {
		t.Errorf("Expected 0 enabled plugins, got %d", len(manager.config.EnabledPlugins))
	}
}

// TestSetEnabledRemovesFromDisabled tests that enabling removes from disabled list.
func TestSetEnabledRemovesFromDisabled(t *testing.T) {
	config := Config{
		DisabledPlugins: []string{"input:test"},
	}
	manager := NewBuilder().WithConfig(config).Build()

	manager.SetEnabled("input", "test")

	if len(manager.config.DisabledPlugins) != 0 {
		t.Errorf("Expected 0 disabled plugins, got %d", len(manager.config.DisabledPlugins))
	}
	if len(manager.config.EnabledPlugins) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(manager.config.EnabledPlugins))
	}
}

// TestRegisterExternalPluginInvalidPath tests registering with invalid path.
func TestRegisterExternalPluginInvalidPath(t *testing.T) {
	manager := NewBuilder().Build()

	err := manager.RegisterExternalPlugin("test", "output", "relative/path", "Test plugin")
	if err == nil {
		t.Error("Expected error for relative path")
	}
}

// TestRegisterExternalPluginNonExistentPath tests registering with non-existent path.
func TestRegisterExternalPluginNonExistentPath(t *testing.T) {
	manager := NewBuilder().Build()

	err := manager.RegisterExternalPlugin("test", "output", "/nonexistent/path", "Test plugin")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

// TestRegisterExternalPluginDirectory tests registering a directory instead of file.
func TestRegisterExternalPluginDirectory(t *testing.T) {
	manager := NewBuilder().Build()
	tmpDir := t.TempDir()

	err := manager.RegisterExternalPlugin("test", "output", tmpDir, "Test plugin")
	if err == nil {
		t.Error("Expected error for directory path")
	}
}

// TestParsePluginList tests parsing comma-separated plugin lists.
func TestParsePluginList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single plugin",
			input:    "kitty",
			expected: []string{"kitty"},
		},
		{
			name:     "Multiple plugins",
			input:    "kitty,waybar,fuzzel",
			expected: []string{"kitty", "waybar", "fuzzel"},
		},
		{
			name:     "With type prefix",
			input:    "output:kitty,input:image",
			expected: []string{"output:kitty", "input:image"},
		},
		{
			name:     "With spaces",
			input:    " kitty , waybar , fuzzel ",
			expected: []string{"kitty", "waybar", "fuzzel"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Trailing comma",
			input:    "kitty,waybar,",
			expected: []string{"kitty", "waybar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePluginList(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected '%s' at index %d, got '%s'", expected, i, result[i])
				}
			}
		})
	}
}

// TestExternalInputPluginBasics tests basic external input plugin methods.
func TestExternalInputPluginBasics(t *testing.T) {
	plugin := NewExternalInputPlugin("test", "Test plugin", "/path/to/plugin")

	if plugin.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", plugin.Name())
	}
	if plugin.Description() != "Test plugin" {
		t.Errorf("Expected description 'Test plugin', got '%s'", plugin.Description())
	}

	// Test args.
	args := map[string]any{"key": "value"}
	plugin.SetArgs(args)
	retrievedArgs := plugin.GetArgs()
	if retrievedArgs["key"] != "value" {
		t.Error("Failed to set/get args")
	}

	// Test dry-run.
	plugin.SetDryRun(true)
	if !plugin.GetDryRun() {
		t.Error("Failed to set/get dry-run")
	}
}

// TestExternalOutputPluginBasics tests basic external output plugin methods.
func TestExternalOutputPluginBasics(t *testing.T) {
	plugin := NewExternalOutputPlugin("test", "Test plugin", "/path/to/plugin")

	if plugin.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", plugin.Name())
	}
	if plugin.Description() != "Test plugin" {
		t.Errorf("Expected description 'Test plugin', got '%s'", plugin.Description())
	}

	// Test args.
	args := map[string]any{"key": "value"}
	plugin.SetArgs(args)
	retrievedArgs := plugin.GetArgs()
	if retrievedArgs["key"] != "value" {
		t.Error("Failed to set/get args")
	}

	// Test dry-run.
	plugin.SetDryRun(true)
	if !plugin.GetDryRun() {
		t.Error("Failed to set/get dry-run")
	}

	// Test verbose.
	plugin.SetVerbose(true)
	if !plugin.GetVerbose() {
		t.Error("Failed to set/get verbose")
	}

	// Test DefaultOutputDir.
	if plugin.DefaultOutputDir() != "" {
		t.Error("External plugin should return empty DefaultOutputDir")
	}
}

// TestConvertCategorisedPaletteToProtocol tests palette conversion.
func TestConvertCategorisedPaletteToProtocol(t *testing.T) {
	palette := &colour.CategorisedPalette{
		Colours: map[colour.Role]colour.CategorisedColour{
			colour.RoleBackground: {
				RGB:        colour.RGB{R: 30, G: 30, B: 46},
				Hex:        "#1e1e2e",
				Role:       colour.RoleBackground,
				Luminance:  0.1,
				IsLight:    false,
				Hue:        240.0,
				Saturation: 0.2,
				Index:      0,
			},
		},
		AllColours: []colour.CategorisedColour{
			{
				RGB:        colour.RGB{R: 30, G: 30, B: 46},
				Hex:        "#1e1e2e",
				Role:       colour.RoleBackground,
				Luminance:  0.1,
				IsLight:    false,
				Hue:        240.0,
				Saturation: 0.2,
				Index:      0,
			},
		},
		ThemeType: colour.ThemeDark,
	}

	args := map[string]any{"test": "value"}
	result := convertCategorisedPaletteToProtocol(palette, args, true)

	if result.ThemeType != "dark" {
		t.Errorf("Expected theme type 'dark', got '%s'", result.ThemeType)
	}
	if !result.DryRun {
		t.Error("Expected DryRun to be true")
	}
	if len(result.Colours) != 1 {
		t.Errorf("Expected 1 colour, got %d", len(result.Colours))
	}
	if len(result.AllColours) != 1 {
		t.Errorf("Expected 1 colour in AllColours, got %d", len(result.AllColours))
	}
	if result.PluginArgs["test"] != "value" {
		t.Error("Plugin args not preserved")
	}

	// Check colour conversion.
	bgColour, ok := result.Colours["background"]
	if !ok {
		t.Fatal("Background colour not found")
	}
	if bgColour.Hex != "#1e1e2e" {
		t.Errorf("Expected hex '#1e1e2e', got '%s'", bgColour.Hex)
	}
	if bgColour.RGB.R != 30 || bgColour.RGB.G != 30 || bgColour.RGB.B != 46 {
		t.Error("RGB values not preserved")
	}
}

// TestIsOutputDisabled tests checking if an output plugin is disabled.
func TestIsOutputDisabled(t *testing.T) {
	config := Config{
		DisabledPlugins: []string{"output:test"},
	}
	manager := NewBuilder().WithConfig(config).Build()

	disabledPlugin := &mockOutputPlugin{name: "test"}
	enabledPlugin := &mockOutputPlugin{name: "other"}

	if !manager.IsOutputDisabled(disabledPlugin) {
		t.Error("Plugin should be disabled")
	}
	if manager.IsOutputDisabled(enabledPlugin) {
		t.Error("Plugin should not be disabled")
	}
}

// TestGetConfig tests retrieving the current configuration.
func TestGetConfig(t *testing.T) {
	config := Config{
		DisabledPlugins: []string{"output:test"},
		EnabledPlugins:  []string{"input:image"},
	}
	manager := NewBuilder().WithConfig(config).Build()

	retrievedConfig := manager.GetConfig()
	if len(retrievedConfig.DisabledPlugins) != 1 {
		t.Errorf("Expected 1 disabled plugin, got %d", len(retrievedConfig.DisabledPlugins))
	}
	if len(retrievedConfig.EnabledPlugins) != 1 {
		t.Errorf("Expected 1 enabled plugin, got %d", len(retrievedConfig.EnabledPlugins))
	}
}

// TestAllPlugins tests retrieving all registered plugins.
func TestAllPlugins(t *testing.T) {
	inputReg := input.NewRegistry()
	outputReg := output.NewRegistry()

	inputReg.Register(&mockInputPlugin{name: "test1"})
	inputReg.Register(&mockInputPlugin{name: "test2"})
	outputReg.Register(&mockOutputPlugin{name: "test3"})

	manager := NewBuilder().WithCustomRegistries(inputReg, outputReg).Build()

	allInput := manager.AllInputPlugins()
	allOutput := manager.AllOutputPlugins()

	// Note: Build() also registers built-in plugins, so we check for our test plugins specifically.
	if _, ok := allInput["test1"]; !ok {
		t.Error("test1 plugin not found")
	}
	if _, ok := allInput["test2"]; !ok {
		t.Error("test2 plugin not found")
	}
	if _, ok := allOutput["test3"]; !ok {
		t.Error("test3 plugin not found")
	}
}

// TestInputOutputRegistry tests accessing registries directly.
func TestInputOutputRegistry(t *testing.T) {
	manager := NewBuilder().Build()

	if manager.InputRegistry() == nil {
		t.Error("InputRegistry should not be nil")
	}
	if manager.OutputRegistry() == nil {
		t.Error("OutputRegistry should not be nil")
	}
}

// TestGetPluginMethods tests GetInputPlugin and GetOutputPlugin.
func TestGetPluginMethods(t *testing.T) {
	inputReg := input.NewRegistry()
	outputReg := output.NewRegistry()

	inputReg.Register(&mockInputPlugin{name: "test-input"})
	outputReg.Register(&mockOutputPlugin{name: "test-output"})

	manager := NewBuilder().WithCustomRegistries(inputReg, outputReg).Build()

	// Test GetInputPlugin.
	inputPlugin, ok := manager.GetInputPlugin("test-input")
	if !ok {
		t.Error("Failed to get input plugin")
	}
	if inputPlugin.Name() != "test-input" {
		t.Errorf("Expected 'test-input', got '%s'", inputPlugin.Name())
	}

	_, ok = manager.GetInputPlugin("nonexistent")
	if ok {
		t.Error("Should not find nonexistent plugin")
	}

	// Test GetOutputPlugin.
	outputPlugin, ok := manager.GetOutputPlugin("test-output")
	if !ok {
		t.Error("Failed to get output plugin")
	}
	if outputPlugin.Name() != "test-output" {
		t.Errorf("Expected 'test-output', got '%s'", outputPlugin.Name())
	}

	_, ok = manager.GetOutputPlugin("nonexistent")
	if ok {
		t.Error("Should not find nonexistent plugin")
	}
}

// TestExternalPluginValidate tests external plugin validation.
func TestExternalPluginValidate(t *testing.T) {
	inputPlugin := NewExternalInputPlugin("test", "Test", "/path/to/plugin")
	outputPlugin := NewExternalOutputPlugin("test", "Test", "/path/to/plugin")

	// Validate is currently a no-op, should not error.
	if err := inputPlugin.Validate(); err != nil {
		t.Errorf("Validate should not error: %v", err)
	}
	if err := outputPlugin.Validate(); err != nil {
		t.Errorf("Validate should not error: %v", err)
	}
}

// TestRegisterExternalPluginUnknownType tests registering with unknown plugin type.
func TestRegisterExternalPluginUnknownType(t *testing.T) {
	manager := NewBuilder().Build()

	// Create a dummy file for testing.
	tmpFile := filepath.Join(t.TempDir(), "plugin.sh")
	if err := os.WriteFile(tmpFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	// Note: This will fail at queryPluginInfo stage, not at type check.
	// To test type check specifically, we'd need to mock queryPluginInfo.
	err := manager.RegisterExternalPlugin("test", "unknown", tmpFile, "Test")
	if err == nil {
		t.Error("Expected error for unknown plugin type or failed query")
	}
}
