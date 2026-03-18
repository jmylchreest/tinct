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

	if len(e.Props) != 4 {
		t.Errorf("Props has %d items, want 4", len(e.Props))
	}
	if e.Props["string"] != "hello" {
		t.Errorf("string = %v, want hello", e.Props["string"])
	}
	if e.Props["int"] != 42 {
		t.Errorf("int = %v, want 42", e.Props["int"])
	}
	if e.Props["bool"] != true {
		t.Errorf("bool = %v, want true", e.Props["bool"])
	}
	if e.Props["float"] != 3.14 {
		t.Errorf("float = %v, want 3.14", e.Props["float"])
	}
}

// TestEventSetChaining verifies that Set returns the same event for chaining.
func TestEventSetChaining(t *testing.T) {
	e := NewEvent("test").Set("a", 1).Set("b", 2)
	if e.Name != "test" {
		t.Errorf("Name = %q after chaining, want test", e.Name)
	}
	if len(e.Props) != 2 {
		t.Errorf("Props has %d items after chaining, want 2", len(e.Props))
	}
}

// TestNewGenerateEvent verifies the generate event builder with dot-notation dims.
func TestNewGenerateEvent(t *testing.T) {
	e := NewGenerateEvent(GenerateEventParams{
		InputPlugin:     "image",
		OutputPlugins:   []string{"kitty", "waybar", "gtk4"},
		ThemeType:       "dark",
		SeedMode:        "random",
		Backend:         "chroma",
		ExtractAmbience: true,
		DryRun:          false,
		DualTheme:       true,
	})

	if e.Name != "generate" {
		t.Errorf("Name = %q, want generate", e.Name)
	}

	// App dimensions.
	if _, ok := e.Props["app.version"]; !ok {
		t.Error("missing app.version dimension")
	}

	// Input dimensions.
	if e.Props["input.plugin"] != "image" {
		t.Errorf("input.plugin = %v, want image", e.Props["input.plugin"])
	}
	if e.Props["input.ai"] != false {
		t.Errorf("input.ai = %v, want false for 'image'", e.Props["input.ai"])
	}

	// Output dimensions.
	outputPlugins, ok := e.Props["output.plugins"].([]string)
	if !ok {
		t.Fatalf("output.plugins is %T, want []string", e.Props["output.plugins"])
	}
	if len(outputPlugins) != 3 || outputPlugins[0] != "kitty" || outputPlugins[1] != "waybar" || outputPlugins[2] != "gtk4" {
		t.Errorf("output.plugins = %v, want [kitty waybar gtk4]", outputPlugins)
	}

	// Generate dimensions.
	if e.Props["generate.theme_type"] != "dark" {
		t.Errorf("generate.theme_type = %v, want dark", e.Props["generate.theme_type"])
	}
	if e.Props["generate.seed_mode"] != "random" {
		t.Errorf("generate.seed_mode = %v, want random", e.Props["generate.seed_mode"])
	}
	if e.Props["generate.backend"] != "chroma" {
		t.Errorf("generate.backend = %v, want chroma", e.Props["generate.backend"])
	}
	if e.Props["generate.ambience"] != true {
		t.Errorf("generate.ambience = %v, want true", e.Props["generate.ambience"])
	}
	if e.Props["generate.dry_run"] != false {
		t.Errorf("generate.dry_run = %v, want false", e.Props["generate.dry_run"])
	}
	if e.Props["generate.dual_theme"] != true {
		t.Errorf("generate.dual_theme = %v, want true", e.Props["generate.dual_theme"])
	}

	// Verify no arch dimension (handled by SDK User-Agent).
	if _, ok := e.Props["arch"]; ok {
		t.Error("unexpected 'arch' dimension; arch is provided by the SDK User-Agent")
	}
	if _, ok := e.Props["app.arch"]; ok {
		t.Error("unexpected 'app.arch' dimension; arch is provided by the SDK User-Agent")
	}
}

// TestNewGenerateEventAIInput verifies AI input detection.
func TestNewGenerateEventAIInput(t *testing.T) {
	tests := []struct {
		input  string
		wantAI bool
	}{
		{"google-genai", true},
		{"openrouter", true},
		{"image", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			e := NewGenerateEvent(GenerateEventParams{InputPlugin: tt.input})
			if e.Props["input.ai"] != tt.wantAI {
				t.Errorf("input.ai = %v for %q, want %v", e.Props["input.ai"], tt.input, tt.wantAI)
			}
		})
	}
}

// TestNewGenerateEventOptionalFields verifies that empty optional fields are omitted.
func TestNewGenerateEventOptionalFields(t *testing.T) {
	e := NewGenerateEvent(GenerateEventParams{})

	// Optional string fields should not be present.
	for _, key := range []string{"input.plugin", "output.plugins", "generate.theme_type", "generate.seed_mode", "generate.backend"} {
		if _, ok := e.Props[key]; ok {
			t.Errorf("unexpected dimension %q when param is empty", key)
		}
	}

	// Boolean fields should always be present.
	for _, key := range []string{"generate.ambience", "generate.dry_run", "generate.dual_theme", "input.ai"} {
		if _, ok := e.Props[key]; !ok {
			t.Errorf("missing expected dimension %q", key)
		}
	}
}

// TestNewPluginUsedEvent verifies the plugin_used event builder with dot-notation.
func TestNewPluginUsedEvent(t *testing.T) {
	e := NewPluginUsedEvent("kitty", "0.1.27", false, "ok")

	if e.Name != "plugin_used" {
		t.Errorf("Name = %q, want plugin_used", e.Name)
	}
	if e.Props["plugin.name"] != "kitty" {
		t.Errorf("plugin.name = %v, want kitty", e.Props["plugin.name"])
	}
	if e.Props["plugin.version"] != "0.1.27" {
		t.Errorf("plugin.version = %v, want 0.1.27", e.Props["plugin.version"])
	}
	if e.Props["plugin.external"] != false {
		t.Errorf("plugin.external = %v, want false", e.Props["plugin.external"])
	}
	if e.Props["plugin.status"] != "ok" {
		t.Errorf("plugin.status = %v, want ok", e.Props["plugin.status"])
	}
}

// TestNewPluginUsedEventExternal verifies external plugin detection.
func TestNewPluginUsedEventExternal(t *testing.T) {
	e := NewPluginUsedEvent("templater", "1.0.0", true, "ok")
	if e.Props["plugin.external"] != true {
		t.Errorf("plugin.external = %v, want true", e.Props["plugin.external"])
	}
}

// TestNewPluginUsedEventNoVersion verifies version omission.
func TestNewPluginUsedEventNoVersion(t *testing.T) {
	e := NewPluginUsedEvent("kitty", "", false, "ok")
	if _, ok := e.Props["plugin.version"]; ok {
		t.Error("unexpected plugin.version when version is empty")
	}
}

// TestNewPluginUsedEventStatuses verifies all status values.
func TestNewPluginUsedEventStatuses(t *testing.T) {
	for _, status := range []string{"ok", "failed", "skipped"} {
		t.Run(status, func(t *testing.T) {
			e := NewPluginUsedEvent("test", "", false, status)
			if e.Props["plugin.status"] != status {
				t.Errorf("plugin.status = %v, want %s", e.Props["plugin.status"], status)
			}
		})
	}
}

// TestDimensionsDotNotation verifies all dimensions use dot-notation grouping.
func TestDimensionsDotNotation(t *testing.T) {
	events := []Event{
		NewGenerateEvent(GenerateEventParams{
			InputPlugin:   "image",
			OutputPlugins: []string{"kitty"},
			ThemeType:     "dark",
			SeedMode:      "random",
			Backend:       "chroma",
		}),
		NewPluginUsedEvent("kitty", "0.1.27", false, "ok"),
	}

	for _, e := range events {
		for key := range e.Props {
			// All keys should either be dot-notated or be a known simple key.
			if !strings.Contains(key, ".") {
				t.Errorf("event %q: dimension %q does not use dot-notation", e.Name, key)
			}
		}
	}
}

// TestIsAIInput verifies the AI input helper.
func TestIsAIInput(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"google-genai", true},
		{"openrouter", true},
		{"image", false},
		{"random", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := isAIInput(tt.input); got != tt.want {
			t.Errorf("isAIInput(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
