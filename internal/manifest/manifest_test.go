package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if m.path != manifestPath {
		t.Errorf("Expected path %s, got %s", manifestPath, m.path)
	}
}

func TestLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	// Create manager and add a file
	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	content := []byte("test content")
	if err := m.RecordFile("test-plugin", "/tmp/test.txt", content); err != nil {
		t.Fatalf("RecordFile() error = %v", err)
	}

	// Save
	if err := m.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load in new manager
	m2, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	if err := m2.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify file was loaded
	files := m2.GetAllFiles()
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}
}

func TestRecordFile(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Record a file
	content := []byte("test content")
	testPath := filepath.Join(tmpDir, "test.txt")

	if err := m.RecordFile("hyprland", testPath, content); err != nil {
		t.Fatalf("RecordFile() error = %v", err)
	}

	// Verify entry
	entry, exists := m.GetFile(testPath)
	if !exists {
		t.Fatal("File not found in manifest")
	}

	if entry.Plugin != "hyprland" {
		t.Errorf("Expected plugin 'hyprland', got '%s'", entry.Plugin)
	}

	if entry.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), entry.Size)
	}

	if entry.Action != ActionCreated {
		t.Errorf("Expected action 'created', got '%s'", entry.Action)
	}

	if entry.Checksum == "" {
		t.Error("Expected checksum to be set")
	}
}

func TestIsTracked(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	testPath := filepath.Join(tmpDir, "test.txt")

	// Not tracked initially
	if m.IsTracked(testPath) {
		t.Error("File should not be tracked initially")
	}

	// Record file
	if err := m.RecordFile("test", testPath, []byte("content")); err != nil {
		t.Fatalf("RecordFile() error = %v", err)
	}

	// Now tracked
	if !m.IsTracked(testPath) {
		t.Error("File should be tracked after recording")
	}
}

func TestGetFilesByPlugin(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Add files from different plugins
	m.RecordFile("hyprland", filepath.Join(tmpDir, "hypr1.txt"), []byte("1"))
	m.RecordFile("hyprland", filepath.Join(tmpDir, "hypr2.txt"), []byte("2"))
	m.RecordFile("kitty", filepath.Join(tmpDir, "kitty1.txt"), []byte("3"))

	// Get hyprland files
	hyprFiles := m.GetFilesByPlugin("hyprland")
	if len(hyprFiles) != 2 {
		t.Errorf("Expected 2 hyprland files, got %d", len(hyprFiles))
	}

	// Get kitty files
	kittyFiles := m.GetFilesByPlugin("kitty")
	if len(kittyFiles) != 1 {
		t.Errorf("Expected 1 kitty file, got %d", len(kittyFiles))
	}

	// Get non-existent plugin
	noneFiles := m.GetFilesByPlugin("nonexistent")
	if len(noneFiles) != 0 {
		t.Errorf("Expected 0 files for nonexistent plugin, got %d", len(noneFiles))
	}
}

func TestDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create actual file
	testPath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testPath, content, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Record in manifest
	if err := m.RecordFile("test", testPath, content); err != nil {
		t.Fatalf("RecordFile() error = %v", err)
	}

	// Delete
	if err := m.DeleteFile(testPath); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}

	// Verify file is gone from disk
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Error("File should be deleted from disk")
	}

	// Verify removed from manifest
	if m.IsTracked(testPath) {
		t.Error("File should be removed from manifest")
	}
}

func TestVerifyFile(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create actual file
	testPath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testPath, content, 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Record in manifest
	if err := m.RecordFile("test", testPath, content); err != nil {
		t.Fatalf("RecordFile() error = %v", err)
	}

	// Verify - should match
	exists, matches, err := m.VerifyFile(testPath)
	if err != nil {
		t.Fatalf("VerifyFile() error = %v", err)
	}
	if !exists {
		t.Error("File should exist")
	}
	if !matches {
		t.Error("Checksum should match")
	}

	// Modify file
	if err := os.WriteFile(testPath, []byte("modified"), 0o644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Verify - should not match
	exists, matches, err = m.VerifyFile(testPath)
	if err != nil {
		t.Fatalf("VerifyFile() error = %v", err)
	}
	if !exists {
		t.Error("File should still exist")
	}
	if matches {
		t.Error("Checksum should not match after modification")
	}
}

func TestGetOrphans(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Add files from different plugins
	m.RecordFile("hyprland", filepath.Join(tmpDir, "hypr.txt"), []byte("1"))
	m.RecordFile("removed-plugin", filepath.Join(tmpDir, "orphan.txt"), []byte("2"))

	// Get orphans - hyprland is available, removed-plugin is not
	orphans := m.GetOrphans([]string{"hyprland", "kitty"})
	if len(orphans) != 1 {
		t.Errorf("Expected 1 orphan, got %d", len(orphans))
	}

	for _, entry := range orphans {
		if entry.Plugin != "removed-plugin" {
			t.Errorf("Expected orphan from 'removed-plugin', got '%s'", entry.Plugin)
		}
	}
}

func TestGetMissing(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Create actual file
	existingPath := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(existingPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Record both existing and non-existing files
	m.RecordFile("test", existingPath, []byte("content"))
	m.RecordFile("test", filepath.Join(tmpDir, "missing.txt"), []byte("content"))

	// Get missing
	missing := m.GetMissing()
	if len(missing) != 1 {
		t.Errorf("Expected 1 missing file, got %d", len(missing))
	}
}

func TestGetPlugins(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Add files from different plugins
	m.RecordFile("hyprland", filepath.Join(tmpDir, "hypr.txt"), []byte("1"))
	m.RecordFile("kitty", filepath.Join(tmpDir, "kitty.txt"), []byte("2"))
	m.RecordFile("hyprland", filepath.Join(tmpDir, "hypr2.txt"), []byte("3"))

	plugins := m.GetPlugins()
	if len(plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(plugins))
	}

	// Check both plugins are present
	pluginSet := make(map[string]bool)
	for _, p := range plugins {
		pluginSet[p] = true
	}

	if !pluginSet["hyprland"] || !pluginSet["kitty"] {
		t.Errorf("Expected hyprland and kitty, got %v", plugins)
	}
}

func TestTotalSize(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "manifest.json")

	m, err := NewManager(manifestPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Add files
	m.RecordFile("test", filepath.Join(tmpDir, "1.txt"), []byte("12345"))      // 5 bytes
	m.RecordFile("test", filepath.Join(tmpDir, "2.txt"), []byte("1234567890")) // 10 bytes

	total := m.TotalSize()
	if total != 15 {
		t.Errorf("Expected total size 15, got %d", total)
	}
}
