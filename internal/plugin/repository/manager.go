// Package repository provides plugin repository management for Tinct.
package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Manager handles plugin repository operations
type Manager struct {
	configPath string
	config     *RepositoryConfig
	cachePath  string
	client     *http.Client
}

// NewManager creates a new repository manager
func NewManager(configPath, cachePath string) (*Manager, error) {
	m := &Manager{
		configPath: configPath,
		cachePath:  cachePath,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Load or initialise config
	if err := m.loadConfig(); err != nil {
		// If file doesn't exist, create default config
		if os.IsNotExist(err) {
			m.config = &RepositoryConfig{
				Repositories: []*Repository{},
				Cache: &CacheConfig{
					TTL:        3600, // 1 hour
					LastUpdate: make(map[string]int64),
				},
			}
			return m, nil
		}
		return nil, err
	}

	return m, nil
}

// AddRepository adds a new repository
func (m *Manager) AddRepository(name, url string, priority int) error {
	// Check if repository already exists
	for _, repo := range m.config.Repositories {
		if repo.Name == name {
			return fmt.Errorf("repository %q already exists", name)
		}
		if repo.URL == url {
			return fmt.Errorf("repository URL %q already added as %q", url, repo.Name)
		}
	}

	// Fetch manifest to validate
	manifest, err := m.fetchManifest(url)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	repo := &Repository{
		Name:     name,
		URL:      url,
		Enabled:  true,
		Priority: priority,
		Manifest: manifest,
	}

	m.config.Repositories = append(m.config.Repositories, repo)
	m.sortRepositories()

	return m.saveConfig()
}

// RemoveRepository removes a repository by name
func (m *Manager) RemoveRepository(name string) error {
	found := false
	filtered := make([]*Repository, 0)

	for _, repo := range m.config.Repositories {
		if repo.Name != name {
			filtered = append(filtered, repo)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("repository %q not found", name)
	}

	m.config.Repositories = filtered

	// Clean up cache
	delete(m.config.Cache.LastUpdate, name)
	cachePath := m.getManifestCachePath(name)
	_ = os.Remove(cachePath) // Ignore error

	return m.saveConfig()
}

// ListRepositories returns all configured repositories
func (m *Manager) ListRepositories() []*Repository {
	return m.config.Repositories
}

// GetRepository returns a repository by name
func (m *Manager) GetRepository(name string) (*Repository, error) {
	for _, repo := range m.config.Repositories {
		if repo.Name == name {
			return repo, nil
		}
	}
	return nil, fmt.Errorf("repository %q not found", name)
}

// UpdateRepository refreshes a repository's manifest
func (m *Manager) UpdateRepository(name string) error {
	repo, err := m.GetRepository(name)
	if err != nil {
		return err
	}

	manifest, err := m.fetchManifest(repo.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	repo.Manifest = manifest

	// Update cache timestamp
	m.config.Cache.LastUpdate[name] = time.Now().Unix()

	// Save manifest to cache
	if err := m.saveManifestCache(name, manifest); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	return m.saveConfig()
}

// UpdateAllRepositories refreshes all enabled repositories
func (m *Manager) UpdateAllRepositories() error {
	var errors []error

	for _, repo := range m.config.Repositories {
		if !repo.Enabled {
			continue
		}

		if err := m.UpdateRepository(repo.Name); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", repo.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to update some repositories: %v", errors)
	}

	return nil
}

// Search searches for plugins across all enabled repositories
func (m *Manager) Search(filter SearchFilter) ([]*SearchResult, error) {
	if err := m.ensureManifestsLoaded(); err != nil {
		return nil, err
	}

	var results []*SearchResult

	for _, repo := range m.config.Repositories {
		if !repo.Enabled || repo.Manifest == nil {
			continue
		}

		for _, plugin := range repo.Manifest.Plugins {
			if m.matchesFilter(plugin, filter) {
				// Get latest version
				var latest *Version
				if len(plugin.Versions) > 0 {
					latest = &plugin.Versions[0]
				}

				results = append(results, &SearchResult{
					Plugin:     plugin,
					Version:    latest,
					Repository: repo.Name,
				})
			}
		}
	}

	return results, nil
}

// FindPlugin finds a specific plugin by name across all repositories
func (m *Manager) FindPlugin(name, version string) (*SearchResult, error) {
	if err := m.ensureManifestsLoaded(); err != nil {
		return nil, err
	}

	for _, repo := range m.config.Repositories {
		if !repo.Enabled || repo.Manifest == nil {
			continue
		}

		plugin, ok := repo.Manifest.Plugins[name]
		if !ok {
			continue
		}

		// Find matching version
		var targetVersion *Version
		if version == "" || version == "latest" {
			if len(plugin.Versions) > 0 {
				targetVersion = &plugin.Versions[0]
			}
		} else {
			for i := range plugin.Versions {
				if plugin.Versions[i].Version == version {
					targetVersion = &plugin.Versions[i]
					break
				}
			}
		}

		if targetVersion == nil {
			continue
		}

		return &SearchResult{
			Plugin:     plugin,
			Version:    targetVersion,
			Repository: repo.Name,
		}, nil
	}

	return nil, fmt.Errorf("plugin %q not found", name)
}

// FindPluginInRepository finds a plugin in a specific repository
func (m *Manager) FindPluginInRepository(repoName, pluginName, version string) (*SearchResult, error) {
	repo, err := m.GetRepository(repoName)
	if err != nil {
		return nil, err
	}

	if err := m.ensureManifestLoaded(repo); err != nil {
		return nil, err
	}

	plugin, ok := repo.Manifest.Plugins[pluginName]
	if !ok {
		return nil, fmt.Errorf("plugin %q not found in repository %q", pluginName, repoName)
	}

	// Find matching version
	var targetVersion *Version
	if version == "" || version == "latest" {
		if len(plugin.Versions) > 0 {
			targetVersion = &plugin.Versions[0]
		}
	} else {
		for i := range plugin.Versions {
			if plugin.Versions[i].Version == version {
				targetVersion = &plugin.Versions[i]
				break
			}
		}
	}

	if targetVersion == nil {
		return nil, fmt.Errorf("version %q not found for plugin %q", version, pluginName)
	}

	return &SearchResult{
		Plugin:     plugin,
		Version:    targetVersion,
		Repository: repoName,
	}, nil
}

// matchesFilter checks if a plugin matches the search filter
func (m *Manager) matchesFilter(plugin *Plugin, filter SearchFilter) bool {
	// Query match (name or description)
	if filter.Query != "" {
		query := strings.ToLower(filter.Query)
		if !strings.Contains(strings.ToLower(plugin.Name), query) &&
			!strings.Contains(strings.ToLower(plugin.Description), query) {
			return false
		}
	}

	// Type filter
	if filter.Type != "" && plugin.Type != filter.Type {
		return false
	}

	// Author filter
	if filter.Author != "" && !strings.EqualFold(plugin.Author, filter.Author) {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, filterTag := range filter.Tags {
			for _, pluginTag := range plugin.Tags {
				if strings.EqualFold(pluginTag, filterTag) {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	return true
}

// ensureManifestsLoaded ensures all repository manifests are loaded
func (m *Manager) ensureManifestsLoaded() error {
	for _, repo := range m.config.Repositories {
		if !repo.Enabled {
			continue
		}

		if err := m.ensureManifestLoaded(repo); err != nil {
			return err
		}
	}
	return nil
}

// ensureManifestLoaded ensures a single repository manifest is loaded
func (m *Manager) ensureManifestLoaded(repo *Repository) error {
	if repo.Manifest != nil {
		return nil
	}

	// Try to load from cache first
	lastUpdate, ok := m.config.Cache.LastUpdate[repo.Name]
	cacheValid := ok && time.Now().Unix()-lastUpdate < int64(m.config.Cache.TTL)

	if cacheValid {
		manifest, err := m.loadManifestCache(repo.Name)
		if err == nil {
			repo.Manifest = manifest
			return nil
		}
		// Cache load failed, fetch fresh
	}

	// Fetch fresh manifest
	manifest, err := m.fetchManifest(repo.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest for %q: %w", repo.Name, err)
	}

	repo.Manifest = manifest

	// Update cache
	m.config.Cache.LastUpdate[repo.Name] = time.Now().Unix()
	_ = m.saveManifestCache(repo.Name, manifest) // Ignore cache save errors

	return nil
}

// fetchManifest fetches a manifest from a URL
func (m *Manager) fetchManifest(url string) (*Manifest, error) {
	resp, err := m.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// loadConfig loads the repository configuration
func (m *Manager) loadConfig() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var config RepositoryConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialise cache if nil
	if config.Cache == nil {
		config.Cache = &CacheConfig{
			TTL:        3600,
			LastUpdate: make(map[string]int64),
		}
	}
	if config.Cache.LastUpdate == nil {
		config.Cache.LastUpdate = make(map[string]int64)
	}

	m.config = &config
	m.sortRepositories()

	return nil
}

// saveConfig saves the repository configuration
func (m *Manager) saveConfig() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// getManifestCachePath returns the cache path for a repository manifest
func (m *Manager) getManifestCachePath(repoName string) string {
	return filepath.Join(m.cachePath, fmt.Sprintf("%s.json", repoName))
}

// saveManifestCache saves a manifest to cache
func (m *Manager) saveManifestCache(repoName string, manifest *Manifest) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(m.cachePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cachePath := m.getManifestCachePath(repoName)

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// loadManifestCache loads a manifest from cache
func (m *Manager) loadManifestCache(repoName string) (*Manifest, error) {
	cachePath := m.getManifestCachePath(repoName)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse cached manifest: %w", err)
	}

	return &manifest, nil
}

// sortRepositories sorts repositories by priority (lower = higher priority)
func (m *Manager) sortRepositories() {
	sort.Slice(m.config.Repositories, func(i, j int) bool {
		return m.config.Repositories[i].Priority < m.config.Repositories[j].Priority
	})
}
