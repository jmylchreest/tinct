package executor

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmylchreest/tinct/internal/plugin/protocol"
	"github.com/jmylchreest/tinct/pkg/plugin"
)

// TestNewWithVerbose tests creating a new executor.
func TestNewWithVerbose(t *testing.T) {
	pluginPath := copyTestScript(t, "basic-input.sh")

	// Test creating executor.
	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	if executor.path != pluginPath {
		t.Errorf("Expected path '%s', got '%s'", pluginPath, executor.path)
	}
	if executor.verbose {
		t.Error("Expected verbose to be false")
	}
	if executor.protocolType != protocol.PluginTypeJSON {
		t.Errorf("Expected protocol type JSON, got %s", executor.protocolType)
	}
}

// TestNewWithVerboseVerboseMode tests creating executor with verbose mode.
func TestNewWithVerboseVerboseMode(t *testing.T) {
	pluginPath := copyTestScript(t, "basic-input.sh")

	executor, err := NewWithVerbose(pluginPath, true)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	if !executor.verbose {
		t.Error("Expected verbose to be true")
	}
}

// TestNewWithVerboseInvalidPlugin tests creating executor with invalid plugin.
func TestNewWithVerboseInvalidPlugin(t *testing.T) {
	_, err := NewWithVerbose("/nonexistent/plugin", false)
	if err == nil {
		t.Error("Expected error for nonexistent plugin")
	}
}

// TestExecuteInputJSONSuccess tests executing a JSON stdio input plugin.
func TestExecuteInputJSONSuccess(t *testing.T) {
	pluginPath := copyTestScript(t, "input-with-colors.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	opts := plugin.InputOptions{
		Verbose: false,
		DryRun:  false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	colors, err := executor.ExecuteInput(ctx, opts)
	if err != nil {
		t.Fatalf("ExecuteInput failed: %v", err)
	}

	if len(colors) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(colors))
	}

	// Check wallpaper path is stored.
	wallpaperPath := executor.GetWallpaperPath()
	if wallpaperPath != "/tmp/test.jpg" {
		t.Errorf("Expected wallpaper path '/tmp/test.jpg', got '%s'", wallpaperPath)
	}
}

// TestExecuteInputJSONError tests handling JSON stdio input plugin errors.
func TestExecuteInputJSONError(t *testing.T) {
	pluginPath := copyTestScript(t, "input-error.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	opts := plugin.InputOptions{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = executor.ExecuteInput(ctx, opts)
	if err == nil {
		t.Error("Expected error from plugin")
	}
}

// TestExecuteOutputJSONSuccess tests executing a legacy JSON stdio output plugin (< 0.2.0).
// Legacy plugins return freeform text on stdout; executor returns empty map.
func TestExecuteOutputJSONSuccess(t *testing.T) {
	pluginPath := copyTestScript(t, "basic-output.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	paletteData := plugin.PaletteData{
		Colours:    make(map[string]plugin.CategorisedColour),
		AllColours: []plugin.CategorisedColour{},
		ThemeType:  "dark",
		DryRun:     false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	files, err := executor.ExecuteOutput(ctx, paletteData)
	if err != nil {
		t.Fatalf("ExecuteOutput failed: %v", err)
	}

	// Legacy plugins (protocol < 0.2.0) now return empty map.
	// Freeform text is printed to stderr, not stored as output.txt.
	if len(files) != 0 {
		t.Errorf("Expected 0 files from legacy plugin, got %d", len(files))
	}
}

// TestExecuteOutputJSONV2Success tests executing a protocol 0.2.0 output plugin.
func TestExecuteOutputJSONV2Success(t *testing.T) {
	pluginPath := copyTestScript(t, "output-v2.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	paletteData := plugin.PaletteData{
		Colours:    make(map[string]plugin.CategorisedColour),
		AllColours: []plugin.CategorisedColour{},
		ThemeType:  "dark",
		DryRun:     false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	files, err := executor.ExecuteOutput(ctx, paletteData)
	if err != nil {
		t.Fatalf("ExecuteOutput failed: %v", err)
	}

	// Protocol 0.2.0 plugin returns files_written paths as keys with nil content.
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Check that one file path was returned (we don't know exact temp path).
	for path, content := range files {
		if !strings.HasPrefix(path, "/tmp/tinct-test-") {
			t.Errorf("Expected file path starting with '/tmp/tinct-test-', got '%s'", path)
		}
		if content != nil {
			t.Errorf("Expected nil content for plugin-managed file, got %d bytes", len(content))
		}
	}
}

// TestExecuteOutputJSONV2Failure tests protocol 0.2.0 plugin reporting failure.
func TestExecuteOutputJSONV2Failure(t *testing.T) {
	pluginPath := copyTestScript(t, "output-v2-fail.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	paletteData := plugin.PaletteData{
		Colours:    make(map[string]plugin.CategorisedColour),
		AllColours: []plugin.CategorisedColour{},
		ThemeType:  "dark",
		DryRun:     false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = executor.ExecuteOutput(ctx, paletteData)
	if err == nil {
		t.Error("Expected error from plugin reporting failure")
	}
	if !strings.Contains(err.Error(), "configuration validation failed") {
		t.Errorf("Expected error to contain 'configuration validation failed', got: %v", err)
	}
}

// TestSupportsStructuredResponse tests version-gating logic.
func TestSupportsStructuredResponse(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{"", false},
		{"0.0.1", false},
		{"0.1.0", false},
		{"0.1.9", false},
		{"0.2.0", true},
		{"0.2.1", true},
		{"0.3.0", true},
		{"1.0.0", true},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			e := &PluginExecutor{pluginProtocolVersion: tt.version}
			if got := e.supportsStructuredResponse(); got != tt.expected {
				t.Errorf("supportsStructuredResponse(%q) = %v, want %v", tt.version, got, tt.expected)
			}
		})
	}
}

// TestParseStructuredResponse tests parsing of protocol 0.2.0 JSON responses.
func TestParseStructuredResponse(t *testing.T) {
	e := &PluginExecutor{}

	t.Run("success with files", func(t *testing.T) {
		resp := `{"success": true, "files_written": ["/tmp/a.conf", "/tmp/b.css"], "message": "Generated 2 files"}`
		files, err := e.parseStructuredResponse([]byte(resp))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}
		if _, ok := files["/tmp/a.conf"]; !ok {
			t.Error("Missing /tmp/a.conf in result")
		}
		if _, ok := files["/tmp/b.css"]; !ok {
			t.Error("Missing /tmp/b.css in result")
		}
		// Content should be nil (plugin already wrote the files)
		for path, content := range files {
			if content != nil {
				t.Errorf("Expected nil content for %s, got %d bytes", path, len(content))
			}
		}
	})

	t.Run("success with empty files", func(t *testing.T) {
		resp := `{"success": true, "files_written": [], "message": "No files needed"}`
		files, err := e.parseStructuredResponse([]byte(resp))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("Expected 0 files, got %d", len(files))
		}
	})

	t.Run("failure response", func(t *testing.T) {
		resp := `{"success": false, "files_written": [], "message": "disk full"}`
		_, err := e.parseStructuredResponse([]byte(resp))
		if err == nil {
			t.Error("Expected error for failure response")
		}
		if !strings.Contains(err.Error(), "disk full") {
			t.Errorf("Expected error to contain 'disk full', got: %v", err)
		}
	})

	t.Run("failure with no message", func(t *testing.T) {
		resp := `{"success": false}`
		_, err := e.parseStructuredResponse([]byte(resp))
		if err == nil {
			t.Error("Expected error for failure response")
		}
		if !strings.Contains(err.Error(), "plugin reported failure") {
			t.Errorf("Expected error to contain 'plugin reported failure', got: %v", err)
		}
	})

	t.Run("empty stdout", func(t *testing.T) {
		_, err := e.parseStructuredResponse([]byte{})
		if err == nil {
			t.Error("Expected error for empty response")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := e.parseStructuredResponse([]byte("not json"))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

// TestProtocolVersionStored tests that protocol version is stored from detection.
func TestProtocolVersionStored(t *testing.T) {
	pluginPath := copyTestScript(t, "basic-input.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	if executor.pluginProtocolVersion != "0.1.0" {
		t.Errorf("Expected pluginProtocolVersion '0.1.0', got '%s'", executor.pluginProtocolVersion)
	}
}

// TestPreExecuteJSONSkip tests PreExecute returning skip=true.
func TestPreExecuteJSONSkip(t *testing.T) {
	pluginPath := copyTestScript(t, "output-preexecute-skip.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	skip, reason, err := executor.PreExecute(ctx)
	if err != nil {
		t.Fatalf("PreExecute failed: %v", err)
	}

	if !skip {
		t.Error("Expected skip=true")
	}
	if reason != "test skip reason" {
		t.Errorf("Expected reason 'test skip reason', got '%s'", reason)
	}
}

// TestPostExecuteJSONSuccess tests PostExecute hook.
func TestPostExecuteJSONSuccess(t *testing.T) {
	pluginPath := copyTestScript(t, "output-postexecute.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	writtenFiles := []string{"/tmp/theme.conf", "/tmp/colors.css"}
	err = executor.PostExecute(ctx, writtenFiles)
	if err != nil {
		t.Errorf("PostExecute failed: %v", err)
	}
}

// TestGetFlagHelpJSON tests retrieving flag help from JSON stdio plugin.
func TestGetFlagHelpJSON(t *testing.T) {
	pluginPath := copyTestScript(t, "basic-input.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	flagHelp, err := executor.GetFlagHelp(ctx)
	if err != nil {
		t.Fatalf("GetFlagHelp failed: %v", err)
	}

	// JSON stdio plugins currently return empty flag help
	if len(flagHelp) != 0 {
		t.Errorf("Expected 0 flags for JSON stdio plugin, got %d", len(flagHelp))
	}
}

// TestClose tests closing the executor.
func TestClose(t *testing.T) {
	pluginPath := copyTestScript(t, "basic-input.sh")

	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Close should not panic.
	executor.Close()

	// Second close should also not panic.
	executor.Close()
}

// TestExecuteInputUnsupportedProtocol tests error handling for unsupported protocol.
func TestExecuteInputUnsupportedProtocol(t *testing.T) {
	executor := &PluginExecutor{
		path:         "/tmp/test",
		protocolType: protocol.PluginType("unknown"),
	}

	_, err := executor.ExecuteInput(context.Background(), plugin.InputOptions{})
	if err == nil {
		t.Error("Expected error for unsupported protocol")
	}
}

// TestExecuteOutputUnsupportedProtocol tests error handling for unsupported protocol.
func TestExecuteOutputUnsupportedProtocol(t *testing.T) {
	executor := &PluginExecutor{
		path:         "/tmp/test",
		protocolType: protocol.PluginType("unknown"),
	}

	_, err := executor.ExecuteOutput(context.Background(), plugin.PaletteData{})
	if err == nil {
		t.Error("Expected error for unsupported protocol")
	}
}

// TestGetWallpaperPathEmpty tests GetWallpaperPath when no wallpaper is set.
func TestGetWallpaperPathEmpty(t *testing.T) {
	executor := &PluginExecutor{
		protocolType: protocol.PluginTypeJSON,
	}

	wallpaperPath := executor.GetWallpaperPath()
	if wallpaperPath != "" {
		t.Errorf("Expected empty wallpaper path, got '%s'", wallpaperPath)
	}
}

// TestConvertInputOpts tests JSON encoding of input options.
func TestConvertInputOpts(t *testing.T) {
	opts := plugin.InputOptions{
		Verbose:         true,
		DryRun:          true,
		ColourOverrides: []string{"background=#ff0000"},
		PluginArgs: map[string]any{
			"key": "value",
		},
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("Failed to marshal opts: %v", err)
	}

	var decoded plugin.InputOptions
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal opts: %v", err)
	}

	if decoded.Verbose != opts.Verbose {
		t.Error("Verbose not preserved")
	}
	if decoded.DryRun != opts.DryRun {
		t.Error("DryRun not preserved")
	}
}

// TestExecuteInputJSONTimeout tests timeout handling using a mock process runner.
func TestExecuteInputJSONTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Create a mock process runner that simulates a timeout by blocking until context is cancelled
	mockRunner := NewTimeoutMockProcessRunner()

	// Create executor with the mock runner
	// Note: We still need a valid plugin path for protocol detection
	pluginPath := copyTestScript(t, "basic-input.sh")

	executor, err := NewWithVerboseAndRunner(pluginPath, false, mockRunner)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	opts := plugin.InputOptions{}

	// Use a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = executor.ExecuteInput(ctx, opts)
	if err == nil {
		t.Error("Expected timeout error")
	}

	// Verify the error is a context timeout error
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
	}
}

// copyTestScript copies a test script from testdata to a temporary directory.
// Returns the path to the copied script with execute permissions set.
func copyTestScript(t *testing.T, scriptName string) string {
	t.Helper()

	// Read the testdata script
	scriptPath := filepath.Join("testdata", "scripts", scriptName)
	scriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("Failed to read testdata script %s: %v", scriptName, err)
	}

	// Create temp directory and copy script
	tmpDir := t.TempDir()
	pluginPath := filepath.Join(tmpDir, scriptName)
	if err := os.WriteFile(pluginPath, scriptContent, 0o755); err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	return pluginPath
}

// --- Configure -------------------------------------------------------------

// Configure returns its error rather than swallowing it, so a failure to
// hand a plugin its args is diagnosable instead of silent. It stays
// non-fatal at the call site, but the executor must not hide it.
func TestConfigureReturnsErrorForUnknownKind(t *testing.T) {
	pluginPath := copyTestScript(t, "basic-input.sh")
	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	// Force the go-plugin branch so the kind switch is reached; a JSON
	// stdio executor returns early before it.
	executor.protocolType = protocol.PluginTypeGoPlugin

	err = executor.Configure(context.Background(), "bogus", plugin.ConfigureRequest{})
	if err == nil {
		t.Fatal("Configure accepted an unknown plugin kind")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q, want it to name the offending kind", err.Error())
	}
}

// JSON-stdio plugins have no Configure RPC; the call is a no-op rather
// than an error, so those plugins keep working untouched.
func TestConfigureIsNoOpForJSONStdio(t *testing.T) {
	pluginPath := copyTestScript(t, "basic-input.sh")
	executor, err := NewWithVerbose(pluginPath, false)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	if executor.protocolType != protocol.PluginTypeJSON {
		t.Skipf("fixture is not a JSON stdio plugin (got %v)", executor.protocolType)
	}

	for _, kind := range []string{"input", "output", "bogus"} {
		if err := executor.Configure(context.Background(), kind, plugin.ConfigureRequest{}); err != nil {
			t.Errorf("Configure(%q) returned %v, want nil for a JSON stdio plugin", kind, err)
		}
	}
}
