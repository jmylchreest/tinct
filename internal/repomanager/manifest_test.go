package repomanager

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
)

func TestManifestManager_MarkDirty(t *testing.T) {
	mgr := &ManifestManager{
		manifest: &repository.Manifest{},
		path:     "/tmp/test.json",
		dirty:    false,
	}

	if mgr.dirty {
		t.Error("Expected manifest to start clean")
	}

	mgr.MarkDirty()

	if !mgr.dirty {
		t.Error("Expected manifest to be marked dirty")
	}
}

func TestManifestManager_Save_SkipWhenClean(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create initial manifest
	manifest := &repository.Manifest{
		Plugins: make(map[string]*repository.Plugin),
	}

	mgr := &ManifestManager{
		manifest: manifest,
		path:     manifestPath,
		dirty:    false, // Clean
	}

	// Save should skip writing
	err := mgr.Save()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(manifestPath); err == nil {
		t.Error("Expected file to not be created when manifest is clean")
	}
}

func TestManifestManager_Save_WriteWhenDirty(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create initial manifest
	manifest := &repository.Manifest{
		Plugins: make(map[string]*repository.Plugin),
	}

	mgr := &ManifestManager{
		manifest: manifest,
		path:     manifestPath,
		dirty:    true, // Dirty
	}

	// Save should write file
	err := mgr.Save()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// File should exist
	if _, err := os.Stat(manifestPath); err != nil {
		t.Errorf("Expected file to be created: %v", err)
	}

	// Dirty flag should be reset
	if mgr.dirty {
		t.Error("Expected dirty flag to be reset after save")
	}

	// LastUpdated should be set
	if mgr.manifest.LastUpdated.IsZero() {
		t.Error("Expected LastUpdated to be set")
	}
}

func TestManifestManager_AddOrUpdatePluginVersion_NewPlugin(t *testing.T) {
	manifest := &repository.Manifest{
		Plugins: make(map[string]*repository.Plugin),
	}

	mgr := &ManifestManager{
		manifest: manifest,
		path:     "/tmp/test.json",
		dirty:    false,
	}

	// Add new plugin version
	version := &repository.Version{
		Version:       "1.0.0",
		Released:      time.Now(),
		Compatibility: ">=0.0.0",
		Downloads: map[string]*repository.Download{
			"linux_x86": {
				URL:       "https://example.com/plugin.tar.gz",
				Checksum:  "sha256:abc123",
				Size:      1024,
				Available: true,
			},
		},
	}

	err := mgr.AddOrUpdatePluginVersion("test-plugin", version)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should be marked dirty
	if !mgr.dirty {
		t.Error("Expected manifest to be marked dirty after adding new plugin")
	}

	// Verify plugin was added
	plugin, exists := manifest.Plugins["test-plugin"]
	if !exists {
		t.Error("Expected plugin to be added to manifest")
	}

	if len(plugin.Versions) != 1 {
		t.Errorf("Expected 1 version, got %d", len(plugin.Versions))
	}

	if plugin.Versions[0].Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", plugin.Versions[0].Version)
	}
}

func TestManifestManager_AddOrUpdatePluginVersion_DuplicateDownload(t *testing.T) {
	// Setup existing plugin with download
	manifest := &repository.Manifest{
		Plugins: map[string]*repository.Plugin{
			"test-plugin": {
				Name: "test-plugin",
				Type: "input",
				Versions: []repository.Version{
					{
						Version:  "1.0.0",
						Released: time.Now(),
						Downloads: map[string]*repository.Download{
							"linux_x86": {
								URL:       "https://example.com/plugin.tar.gz",
								Checksum:  "sha256:abc123",
								Size:      1024,
								Available: true,
							},
						},
					},
				},
			},
		},
	}

	mgr := &ManifestManager{
		manifest: manifest,
		path:     "/tmp/test.json",
		dirty:    false,
	}

	// Add same download again (should not mark dirty)
	version := &repository.Version{
		Version:  "1.0.0",
		Released: time.Now(),
		Downloads: map[string]*repository.Download{
			"linux_x86": {
				URL:       "https://example.com/plugin.tar.gz",
				Checksum:  "sha256:abc123",
				Size:      1024,
				Available: true,
			},
		},
	}

	err := mgr.AddOrUpdatePluginVersion("test-plugin", version)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should NOT be marked dirty (no actual change)
	if mgr.dirty {
		t.Error("Expected manifest to remain clean when adding duplicate download")
	}
}

func TestManifestManager_AddOrUpdatePluginVersion_NewPlatform(t *testing.T) {
	// Setup existing plugin with one platform
	manifest := &repository.Manifest{
		Plugins: map[string]*repository.Plugin{
			"test-plugin": {
				Name: "test-plugin",
				Type: "input",
				Versions: []repository.Version{
					{
						Version:  "1.0.0",
						Released: time.Now(),
						Downloads: map[string]*repository.Download{
							"linux_x86": {
								URL:       "https://example.com/plugin.tar.gz",
								Checksum:  "sha256:abc123",
								Size:      1024,
								Available: true,
							},
						},
					},
				},
			},
		},
	}

	mgr := &ManifestManager{
		manifest: manifest,
		path:     "/tmp/test.json",
		dirty:    false,
	}

	// Add new platform for same version
	version := &repository.Version{
		Version:  "1.0.0",
		Released: time.Now(),
		Downloads: map[string]*repository.Download{
			"linux_arm64": {
				URL:       "https://example.com/plugin-arm.tar.gz",
				Checksum:  "sha256:def456",
				Size:      2048,
				Available: true,
			},
		},
	}

	err := mgr.AddOrUpdatePluginVersion("test-plugin", version)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Should be marked dirty (new platform)
	if !mgr.dirty {
		t.Error("Expected manifest to be marked dirty when adding new platform")
	}

	// Verify both platforms exist
	plugin := manifest.Plugins["test-plugin"]
	if len(plugin.Versions[0].Downloads) != 2 {
		t.Errorf("Expected 2 platforms, got %d", len(plugin.Versions[0].Downloads))
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"equal versions", "1.0.0", "1.0.0", 0},
		{"a greater major", "2.0.0", "1.0.0", 1},
		{"b greater major", "1.0.0", "2.0.0", -1},
		{"a greater minor", "1.2.0", "1.1.0", 1},
		{"b greater minor", "1.1.0", "1.2.0", -1},
		{"a greater patch", "1.0.2", "1.0.1", 1},
		{"b greater patch", "1.0.1", "1.0.2", -1},
		{"with v prefix", "v1.0.0", "v2.0.0", -1},
		{"mixed prefix", "v1.0.0", "2.0.0", -1},
		{"different lengths", "1.0", "1.0.0", 0},
		{"complex version", "1.2.3", "1.2.10", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareVersions(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("CompareVersions(%s, %s) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
