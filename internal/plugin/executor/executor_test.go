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

// TestExecuteOutputJSONSuccess tests executing a JSON stdio output plugin.
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

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if _, ok := files["output.txt"]; !ok {
		t.Error("Expected 'output.txt' in output")
	}

	if content := string(files["output.txt"]); !strings.Contains(content, "theme configuration") {
		t.Errorf("Expected content to contain 'theme configuration', got '%s'", content)
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
	if err := os.WriteFile(pluginPath, scriptContent, 0755); err != nil {
		t.Fatalf("Failed to write test script: %v", err)
	}

	return pluginPath
}
