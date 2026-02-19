package telemetry

import (
	"strings"
	"testing"
)

// TestNewEvent verifies basic event creation.
func TestNewEvent(t *testing.T) {
	e := NewEvent("test_event")

	if e.Name != "test_event" {
		t.Errorf("Name = %q, want %q", e.Name, "test_event")
	}
	if e.Props == nil {
		t.Error("Props should not be nil")
	}
	if len(e.Props) != 0 {
		t.Errorf("Props should be empty, got %d items", len(e.Props))
	}
}

// TestEventSet verifies chainable property setting.
func TestEventSet(t *testing.T) {
	e := NewEvent("test").
		Set("string", "hello").
		Set("int", 42).
		Set("bool", true).
		Set("float", 3.14)

	if e.Props["string"] != "hello" {
		t.Errorf("Props[string] = %v, want hello", e.Props["string"])
	}
	if e.Props["int"] != 42 {
		t.Errorf("Props[int] = %v, want 42", e.Props["int"])
	}
	if e.Props["bool"] != true {
		t.Errorf("Props[bool] = %v, want true", e.Props["bool"])
	}
	if e.Props["float"] != 3.14 {
		t.Errorf("Props[float] = %v, want 3.14", e.Props["float"])
	}
}

// TestNewGenerateEvent verifies the generate event builder with all fields.
func TestNewGenerateEvent(t *testing.T) {
	params := GenerateEventParams{
		InputPlugin:         "image",
		OutputPlugins:       []string{"hyprland", "kitty", "waybar"},
		SuccessCount:        3,
		ThemeType:           "dark",
		SeedMode:            "content",
		Backend:             "kmeans",
		ExtractAmbience:     true,
		DryRun:              false,
		DualTheme:           true,
		ExternalPluginCount: 1,
	}

	e := NewGenerateEvent(params)

	if e.Name != "generate" {
		t.Errorf("Name = %q, want %q", e.Name, "generate")
	}

	// Check required props exist.
	requiredProps := []string{
		"arch", "version", "input_plugin", "output_plugins",
		"output_plugin_count", "success_count", "theme_type",
		"seed_mode", "backend", "extract_ambience", "dry_run",
		"dual_theme", "external_plugin_count", "ai_input",
	}

	for _, key := range requiredProps {
		if _, ok := e.Props[key]; !ok {
			t.Errorf("missing required prop: %s", key)
		}
	}

	// Check specific values.
	if e.Props["input_plugin"] != "image" {
		t.Errorf("input_plugin = %v, want image", e.Props["input_plugin"])
	}

	outputPlugins, ok := e.Props["output_plugins"].(string)
	if !ok {
		t.Fatalf("output_plugins should be string, got %T", e.Props["output_plugins"])
	}
	if outputPlugins != "hyprland,kitty,waybar" {
		t.Errorf("output_plugins = %q, want %q", outputPlugins, "hyprland,kitty,waybar")
	}

	if e.Props["output_plugin_count"] != 3 {
		t.Errorf("output_plugin_count = %v, want 3", e.Props["output_plugin_count"])
	}
	if e.Props["success_count"] != 3 {
		t.Errorf("success_count = %v, want 3", e.Props["success_count"])
	}
	if e.Props["theme_type"] != "dark" {
		t.Errorf("theme_type = %v, want dark", e.Props["theme_type"])
	}
	if e.Props["seed_mode"] != "content" {
		t.Errorf("seed_mode = %v, want content", e.Props["seed_mode"])
	}
	if e.Props["backend"] != "kmeans" {
		t.Errorf("backend = %v, want kmeans", e.Props["backend"])
	}
	if e.Props["extract_ambience"] != true {
		t.Errorf("extract_ambience = %v, want true", e.Props["extract_ambience"])
	}
	if e.Props["dry_run"] != false {
		t.Errorf("dry_run = %v, want false", e.Props["dry_run"])
	}
	if e.Props["dual_theme"] != true {
		t.Errorf("dual_theme = %v, want true", e.Props["dual_theme"])
	}
	if e.Props["external_plugin_count"] != 1 {
		t.Errorf("external_plugin_count = %v, want 1", e.Props["external_plugin_count"])
	}
	if e.Props["ai_input"] != false {
		t.Errorf("ai_input = %v, want false (image is not AI)", e.Props["ai_input"])
	}
}

// TestNewGenerateEventAIInput verifies AI input detection.
func TestNewGenerateEventAIInput(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"image", false},
		{"file", false},
		{"markdown", false},
		{"remote-json", false},
		{"google-genai", true},
		{"openrouter", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			e := NewGenerateEvent(GenerateEventParams{InputPlugin: tt.input})
			if e.Props["ai_input"] != tt.expected {
				t.Errorf("ai_input for %q = %v, want %v", tt.input, e.Props["ai_input"], tt.expected)
			}
		})
	}
}

// TestNewGenerateEventMinimalParams verifies the event works with zero-value params.
func TestNewGenerateEventMinimalParams(t *testing.T) {
	e := NewGenerateEvent(GenerateEventParams{})

	if e.Name != "generate" {
		t.Errorf("Name = %q, want %q", e.Name, "generate")
	}

	// Should still have system props.
	if e.Props["arch"] == nil || e.Props["arch"] == "" {
		t.Error("arch should be set even with minimal params")
	}
	if e.Props["version"] == nil {
		t.Error("version should be set even with minimal params")
	}

	// Optional string fields should be absent when empty.
	if _, ok := e.Props["input_plugin"]; ok {
		t.Error("input_plugin should not be set when empty")
	}
	if _, ok := e.Props["theme_type"]; ok {
		t.Error("theme_type should not be set when empty")
	}

	// Boolean/int fields should still be present.
	if _, ok := e.Props["dry_run"]; !ok {
		t.Error("dry_run should be set even when false")
	}
	if _, ok := e.Props["success_count"]; !ok {
		t.Error("success_count should be set even when 0")
	}
}

// TestNewGenerateEventOutputPluginsJoined verifies plugin names are comma-joined.
func TestNewGenerateEventOutputPluginsJoined(t *testing.T) {
	e := NewGenerateEvent(GenerateEventParams{
		OutputPlugins: []string{"a", "b", "c"},
	})

	plugins := e.Props["output_plugins"].(string)
	parts := strings.Split(plugins, ",")
	if len(parts) != 3 {
		t.Errorf("expected 3 plugins in joined string, got %d: %q", len(parts), plugins)
	}
}

// TestNewPluginUsedEvent verifies the plugin_used event builder.
func TestNewPluginUsedEvent(t *testing.T) {
	e := NewPluginUsedEvent("kitty", false, true)

	if e.Name != "plugin_used" {
		t.Errorf("Name = %q, want %q", e.Name, "plugin_used")
	}
	if e.Props["plugin_name"] != "kitty" {
		t.Errorf("plugin_name = %v, want kitty", e.Props["plugin_name"])
	}
	if e.Props["is_external"] != false {
		t.Errorf("is_external = %v, want false", e.Props["is_external"])
	}
	if e.Props["succeeded"] != true {
		t.Errorf("succeeded = %v, want true", e.Props["succeeded"])
	}
}

// TestNewPluginUsedEventExternal verifies external plugin detection.
func TestNewPluginUsedEventExternal(t *testing.T) {
	e := NewPluginUsedEvent("my-custom-plugin", true, false)

	if e.Props["plugin_name"] != "my-custom-plugin" {
		t.Errorf("plugin_name = %v, want my-custom-plugin", e.Props["plugin_name"])
	}
	if e.Props["is_external"] != true {
		t.Errorf("is_external = %v, want true", e.Props["is_external"])
	}
	if e.Props["succeeded"] != false {
		t.Errorf("succeeded = %v, want false", e.Props["succeeded"])
	}
}

// TestIsAIInput verifies the AI input detection helper.
func TestIsAIInput(t *testing.T) {
	if !isAIInput("google-genai") {
		t.Error("google-genai should be AI input")
	}
	if !isAIInput("openrouter") {
		t.Error("openrouter should be AI input")
	}
	if isAIInput("image") {
		t.Error("image should not be AI input")
	}
	if isAIInput("") {
		t.Error("empty string should not be AI input")
	}
}
