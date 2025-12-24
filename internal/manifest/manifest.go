// Package manifest provides file tracking for tinct-generated files.
package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// ManifestVersion is the current manifest format version.
	ManifestVersion = 1

	// DefaultManifestPath is the default location for the manifest file.
	DefaultManifestPath = ".local/share/tinct/manifest.json"
)

// Action represents the type of file operation.
type Action string

const (
	// ActionCreated indicates the file was created by tinct.
	ActionCreated Action = "created"
	// ActionModified indicates the file existed and was modified by tinct.
	ActionModified Action = "modified"
)

// FileEntry represents a tracked file in the manifest.
type FileEntry struct {
	Plugin    string    `json:"plugin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Checksum  string    `json:"checksum"`
	Size      int64     `json:"size"`
	Action    Action    `json:"action"`
}

// Manifest represents the complete manifest data structure.
type Manifest struct {
	Version   int                   `json:"version"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	Files     map[string]*FileEntry `json:"files"`
}

// Manager handles manifest operations.
type Manager struct {
	path     string
	manifest *Manifest
	mu       sync.RWMutex
}

// NewManager creates a new manifest manager.
// If path is empty, uses the default path in user's home directory.
func NewManager(path string) (*Manager, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, DefaultManifestPath)
	}

	m := &Manager{
		path: path,
		manifest: &Manifest{
			Version:   ManifestVersion,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Files:     make(map[string]*FileEntry),
		},
	}

	return m, nil
}

// Load reads the manifest from disk.
// If the file doesn't exist, initializes an empty manifest.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize empty manifest
			m.manifest = &Manifest{
				Version:   ManifestVersion,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Files:     make(map[string]*FileEntry),
			}
			return nil
		}
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Initialize Files map if nil (for backwards compatibility)
	if manifest.Files == nil {
		manifest.Files = make(map[string]*FileEntry)
	}

	m.manifest = &manifest
	return nil
}

// Save writes the manifest to disk.
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.manifest.UpdatedAt = time.Now()

	// Ensure directory exists
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	data, err := json.MarshalIndent(m.manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(m.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// RecordFile adds or updates a file entry in the manifest.
func (m *Manager) RecordFile(plugin, path string, content []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Normalize path to absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Calculate checksum
	hash := sha256.Sum256(content)
	checksum := "sha256:" + hex.EncodeToString(hash[:])

	now := time.Now()

	// Determine action
	action := ActionCreated
	if entry, exists := m.manifest.Files[absPath]; exists {
		action = entry.Action // Preserve original action
	} else {
		// Check if file existed before tinct
		if _, err := os.Stat(absPath); err == nil {
			action = ActionModified
		}
	}

	// Update or create entry
	if entry, exists := m.manifest.Files[absPath]; exists {
		entry.UpdatedAt = now
		entry.Checksum = checksum
		entry.Size = int64(len(content))
		entry.Plugin = plugin
	} else {
		m.manifest.Files[absPath] = &FileEntry{
			Plugin:    plugin,
			CreatedAt: now,
			UpdatedAt: now,
			Checksum:  checksum,
			Size:      int64(len(content)),
			Action:    action,
		}
	}

	return nil
}

// RemoveFile removes a file entry from the manifest.
func (m *Manager) RemoveFile(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	delete(m.manifest.Files, absPath)
	return nil
}

// DeleteFile removes a file from disk and from the manifest.
func (m *Manager) DeleteFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Remove from disk
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Remove from manifest
	return m.RemoveFile(absPath)
}

// GetFile returns the entry for a specific file.
func (m *Manager) GetFile(path string) (*FileEntry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, false
	}

	entry, exists := m.manifest.Files[absPath]
	return entry, exists
}

// GetFilesByPlugin returns all files for a specific plugin.
func (m *Manager) GetFilesByPlugin(plugin string) map[string]*FileEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	files := make(map[string]*FileEntry)
	for path, entry := range m.manifest.Files {
		if entry.Plugin == plugin {
			files[path] = entry
		}
	}
	return files
}

// GetAllFiles returns all tracked files.
func (m *Manager) GetAllFiles() map[string]*FileEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	files := make(map[string]*FileEntry, len(m.manifest.Files))
	for path, entry := range m.manifest.Files {
		entryCopy := *entry
		files[path] = &entryCopy
	}
	return files
}

// IsTracked checks if a file is tracked in the manifest.
func (m *Manager) IsTracked(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	_, exists := m.manifest.Files[absPath]
	return exists
}

// VerifyFile checks if a file's current checksum matches the manifest.
// Returns: (exists on disk, matches checksum, error).
func (m *Manager) VerifyFile(path string) (bool, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, false, fmt.Errorf("failed to get absolute path: %w", err)
	}

	entry, exists := m.manifest.Files[absPath]
	if !exists {
		return false, false, fmt.Errorf("file not in manifest")
	}

	// Read current file
	content, err := os.ReadFile(absPath) // #nosec G304 -- Path from manifest tracking
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate current checksum
	hash := sha256.Sum256(content)
	currentChecksum := "sha256:" + hex.EncodeToString(hash[:])

	return true, currentChecksum == entry.Checksum, nil
}

// GetOrphans returns files that are tracked but whose plugin is no longer available.
func (m *Manager) GetOrphans(availablePlugins []string) map[string]*FileEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Build plugin lookup map
	pluginSet := make(map[string]bool, len(availablePlugins))
	for _, p := range availablePlugins {
		pluginSet[p] = true
	}

	orphans := make(map[string]*FileEntry)
	for path, entry := range m.manifest.Files {
		if !pluginSet[entry.Plugin] {
			orphans[path] = entry
		}
	}
	return orphans
}

// GetMissing returns files that are in the manifest but don't exist on disk.
func (m *Manager) GetMissing() map[string]*FileEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	missing := make(map[string]*FileEntry)
	for path, entry := range m.manifest.Files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missing[path] = entry
		}
	}
	return missing
}

// Path returns the manifest file path.
func (m *Manager) Path() string {
	return m.path
}

// FileCount returns the number of tracked files.
func (m *Manager) FileCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.manifest.Files)
}

// TotalSize returns the total size of all tracked files.
func (m *Manager) TotalSize() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var total int64
	for _, entry := range m.manifest.Files {
		total += entry.Size
	}
	return total
}

// GetPlugins returns a list of unique plugin names in the manifest.
func (m *Manager) GetPlugins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pluginSet := make(map[string]bool)
	for _, entry := range m.manifest.Files {
		pluginSet[entry.Plugin] = true
	}

	plugins := make([]string, 0, len(pluginSet))
	for plugin := range pluginSet {
		plugins = append(plugins, plugin)
	}
	return plugins
}
