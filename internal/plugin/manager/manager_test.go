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
	result := convertCategorisedPaletteToProtocol(palette, args, true, false)

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

// TestExternalPluginValidate tests that Validate calls the plugin's
// optional Validator RPC. With a bogus binary path the executor cannot
// even detect the protocol, so Validate surfaces a fail-fast error
// rather than silently succeeding (the pre-0.3.0 no-op behaviour).
func TestExternalPluginValidate(t *testing.T) {
	inputPlugin := NewExternalInputPlugin("test", "Test", "/path/to/plugin")
	outputPlugin := NewExternalOutputPlugin("test", "Test", "/path/to/plugin")

	if err := inputPlugin.Validate(); err == nil {
		t.Error("Validate should error for non-existent input plugin binary")
	}
	if err := outputPlugin.Validate(); err == nil {
		t.Error("Validate should error for non-existent output plugin binary")
	}
}

// TestRegisterExternalPluginUnknownType tests registering with unknown plugin type.
func TestRegisterExternalPluginUnknownType(t *testing.T) {
	manager := NewBuilder().Build()

	// Create a dummy file for testing.
	tmpFile := filepath.Join(t.TempDir(), "plugin.sh")
	if err := os.WriteFile(tmpFile, []byte("#!/bin/sh\necho test"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Note: This will fail at queryPluginInfo stage, not at type check.
	err := manager.RegisterExternalPlugin("test", "unknown", tmpFile, "Test")
	if err == nil {
		t.Error("Expected error for unknown plugin type or failed query")
	}
}
