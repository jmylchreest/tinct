package repocli

import (
	"os"
	"testing"
	"time"

	"github.com/jmylchreest/tinct/internal/repomanager"
)

func TestMetadataHydrationCache_GetMetadata(t *testing.T) {
	cache := NewMetadataHydrationCache()

	// Test getting non-existent metadata
	_, exists := cache.GetMetadata("test-plugin", "1.0.0")
	if exists {
		t.Error("Expected metadata to not exist")
	}

	// Add metadata
	metadata := &repomanager.PluginMetadata{
		Name:            "test-plugin",
		Version:         "1.0.0",
		ProtocolVersion: "0.0.1",
		Type:            "input",
	}
	cache.metadata["test-plugin:1.0.0"] = metadata

	// Test getting existing metadata
	retrieved, exists := cache.GetMetadata("test-plugin", "1.0.0")
	if !exists {
		t.Error("Expected metadata to exist")
	}
	if retrieved.Name != "test-plugin" {
		t.Errorf("Expected name 'test-plugin', got '%s'", retrieved.Name)
	}
	if retrieved.ProtocolVersion != "0.0.1" {
		t.Errorf("Expected protocol '0.0.1', got '%s'", retrieved.ProtocolVersion)
	}
}

func TestMetadataHydrationCache_AddPending(t *testing.T) {
	cache := NewMetadataHydrationCache()

	// Add pending plugin
	pending := &PendingPlugin{
		PluginName:    "test-plugin",
		PluginVersion: "1.0.0",
		Platform:      "linux_x86",
		DownloadURL:   "https://example.com/plugin.tar.gz",
		Checksum:      "sha256:abc123",
		Size:          1024,
		Released:      time.Now(),
	}
	cache.AddPending(pending)

	// Verify it was added
	key := "test-plugin:1.0.0"
	if len(cache.pending[key]) != 1 {
		t.Errorf("Expected 1 pending plugin, got %d", len(cache.pending[key]))
	}
	if cache.pending[key][0].Platform != "linux_x86" {
		t.Errorf("Expected platform 'linux_x86', got '%s'", cache.pending[key][0].Platform)
	}

	// Add another pending plugin for same version
	pending2 := &PendingPlugin{
		PluginName:    "test-plugin",
		PluginVersion: "1.0.0",
		Platform:      "linux_arm64",
		DownloadURL:   "https://example.com/plugin-arm.tar.gz",
		Checksum:      "sha256:def456",
		Size:          2048,
		Released:      time.Now(),
	}
	cache.AddPending(pending2)

	// Verify both are pending
	if len(cache.pending[key]) != 2 {
		t.Errorf("Expected 2 pending plugins, got %d", len(cache.pending[key]))
	}
}

func TestMetadataHydrationCache_SetMetadata_WithoutPending(t *testing.T) {
	cache := NewMetadataHydrationCache()

	// Create a dummy manifest manager (not used in this test)
	mgr := &repomanager.ManifestManager{}

	// Set metadata when no pending plugins exist
	metadata := &repomanager.PluginMetadata{
		Name:            "test-plugin",
		Version:         "1.0.0",
		ProtocolVersion: "0.0.1",
		Type:            "input",
	}

	hydrated := cache.SetMetadata("test-plugin", "1.0.0", metadata, mgr, true, false)

	// Should return 0 hydrated
	if hydrated != 0 {
		t.Errorf("Expected 0 hydrated plugins, got %d", hydrated)
	}

	// Verify metadata was cached
	retrieved, exists := cache.GetMetadata("test-plugin", "1.0.0")
	if !exists {
		t.Error("Expected metadata to be cached")
	}
	if retrieved.ProtocolVersion != "0.0.1" {
		t.Errorf("Expected protocol '0.0.1', got '%s'", retrieved.ProtocolVersion)
	}
}

func TestMetadataHydrationCache_SetMetadata_WithPending(t *testing.T) {
	cache := NewMetadataHydrationCache()

	// Add pending plugins
	pending1 := &PendingPlugin{
		PluginName:    "test-plugin",
		PluginVersion: "1.0.0",
		Platform:      "linux_arm64",
		DownloadURL:   "https://example.com/plugin-arm.tar.gz",
		Checksum:      "sha256:abc123",
		Size:          1024,
		Released:      time.Now(),
	}
	cache.AddPending(pending1)

	pending2 := &PendingPlugin{
		PluginName:    "test-plugin",
		PluginVersion: "1.0.0",
		Platform:      "linux_arm",
		DownloadURL:   "https://example.com/plugin-armv7.tar.gz",
		Checksum:      "sha256:def456",
		Size:          2048,
		Released:      time.Now(),
	}
	cache.AddPending(pending2)

	// Create a temp manifest file for testing
	tmpDir := t.TempDir()
	manifestPath := tmpDir + "/manifest.json"

	// Write a valid empty manifest JSON
	manifestJSON := `{"plugins":{}}`
	err := os.WriteFile(manifestPath, []byte(manifestJSON), 0600)
	if err != nil {
		t.Fatalf("Failed to create temp manifest: %v", err)
	}

	// Load manifest manager
	mgr, err := repomanager.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Set metadata - should hydrate pending plugins
	metadata := &repomanager.PluginMetadata{
		Name:            "test-plugin",
		Version:         "1.0.0",
		ProtocolVersion: "0.0.1",
		Type:            "input",
	}

	hydrated := cache.SetMetadata("test-plugin", "1.0.0", metadata, mgr, true, false)

	// Should return 2 hydrated (dry-run mode)
	if hydrated != 2 {
		t.Errorf("Expected 2 hydrated plugins, got %d", hydrated)
	}

	// Verify pending list was cleared
	key := "test-plugin:1.0.0"
	if _, exists := cache.pending[key]; exists {
		t.Error("Expected pending list to be cleared after hydration")
	}

	// Verify metadata was cached
	_, exists := cache.GetMetadata("test-plugin", "1.0.0")
	if !exists {
		t.Error("Expected metadata to be cached")
	}
}

func TestProtocolVersionTracker_NewTracker(t *testing.T) {
	tracker := NewProtocolVersionTracker()

	if tracker == nil {
		t.Error("Expected non-nil tracker")
	}
	if tracker.failures == nil {
		t.Error("Expected failures map to be initialized")
	}
	if len(tracker.failures) != 0 {
		t.Errorf("Expected empty failures map, got %d entries", len(tracker.failures))
	}
}

func TestProtocolVersionTracker_RecordFailure(t *testing.T) {
	tracker := NewProtocolVersionTracker()

	// Record first failure
	tracker.RecordFailure("test-plugin", "1.0.0")

	if tracker.failures["test-plugin"] != "1.0.0" {
		t.Errorf("Expected failure version '1.0.0', got '%s'", tracker.failures["test-plugin"])
	}

	// Record higher version failure
	tracker.RecordFailure("test-plugin", "2.0.0")

	if tracker.failures["test-plugin"] != "2.0.0" {
		t.Errorf("Expected failure version '2.0.0', got '%s'", tracker.failures["test-plugin"])
	}

	// Record lower version failure (should not update)
	tracker.RecordFailure("test-plugin", "1.5.0")

	if tracker.failures["test-plugin"] != "2.0.0" {
		t.Errorf("Expected failure version to remain '2.0.0', got '%s'", tracker.failures["test-plugin"])
	}
}

func TestProtocolVersionTracker_ShouldSkip(t *testing.T) {
	tracker := NewProtocolVersionTracker()

	tests := []struct {
		name            string
		pluginName      string
		pluginVersion   string
		protocolVersion string
		minProtocol     string
		setupFailure    bool
		failureVersion  string
		expectedSkip    bool
		expectedReason  string
	}{
		{
			name:            "no minimum protocol",
			pluginName:      "test-plugin",
			pluginVersion:   "1.0.0",
			protocolVersion: "0.0.1",
			minProtocol:     "",
			expectedSkip:    false,
			expectedReason:  "",
		},
		{
			name:            "protocol meets minimum",
			pluginName:      "test-plugin",
			pluginVersion:   "1.0.0",
			protocolVersion: "0.0.2",
			minProtocol:     "0.0.1",
			expectedSkip:    false,
			expectedReason:  "",
		},
		{
			name:            "protocol below minimum",
			pluginName:      "test-plugin",
			pluginVersion:   "1.0.0",
			protocolVersion: "0.0.1",
			minProtocol:     "0.0.2",
			expectedSkip:    true,
			expectedReason:  "protocol 0.0.1 < minimum 0.0.2",
		},
		{
			name:            "cascade skip - same version as failure",
			pluginName:      "test-plugin",
			pluginVersion:   "1.0.0",
			protocolVersion: "0.0.2",
			minProtocol:     "",
			setupFailure:    true,
			failureVersion:  "1.0.0",
			expectedSkip:    true,
			expectedReason:  "older than or equal to failed version 1.0.0",
		},
		{
			name:            "cascade skip - older than failure",
			pluginName:      "test-plugin",
			pluginVersion:   "0.9.0",
			protocolVersion: "0.0.2",
			minProtocol:     "",
			setupFailure:    true,
			failureVersion:  "1.0.0",
			expectedSkip:    true,
			expectedReason:  "older than or equal to failed version 1.0.0",
		},
		{
			name:            "no cascade skip - newer than failure",
			pluginName:      "test-plugin",
			pluginVersion:   "2.0.0",
			protocolVersion: "0.0.2",
			minProtocol:     "",
			setupFailure:    true,
			failureVersion:  "1.0.0",
			expectedSkip:    false,
			expectedReason:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset tracker for each test
			tracker = NewProtocolVersionTracker()

			if tt.setupFailure {
				tracker.RecordFailure(tt.pluginName, tt.failureVersion)
			}

			skip, reason := tracker.ShouldSkip(tt.pluginName, tt.pluginVersion, tt.protocolVersion, tt.minProtocol)

			if skip != tt.expectedSkip {
				t.Errorf("Expected skip=%v, got skip=%v", tt.expectedSkip, skip)
			}

			if skip && reason != tt.expectedReason {
				t.Errorf("Expected reason '%s', got '%s'", tt.expectedReason, reason)
			}
		})
	}
}
