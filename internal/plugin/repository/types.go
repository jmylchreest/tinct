// Package repository provides plugin repository management for Tinct.
package repository

import "time"

// Manifest represents a plugin repository manifest.
type Manifest struct {
	Version      string             `json:"version"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	URL          string             `json:"url"`
	MaintainedBy string             `json:"maintained_by,omitempty"`
	LastUpdated  time.Time          `json:"last_updated"`
	Plugins      map[string]*Plugin `json:"plugins"`
}

// Plugin represents a plugin in the repository.
type Plugin struct {
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "input" or "output"
	Description string    `json:"description"`
	Repository  string    `json:"repository"`
	Author      string    `json:"author,omitempty"`
	License     string    `json:"license,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Versions    []Version `json:"versions"`
}

// Version represents a specific version of a plugin.
type Version struct {
	Version       string               `json:"version"`
	Released      time.Time            `json:"released"`
	Compatibility string               `json:"compatibility,omitempty"` // e.g., ">=1.0.0"
	ChangelogURL  string               `json:"changelog_url,omitempty"`
	Downloads     map[string]*Download `json:"downloads"`
}

// Download represents a downloadable artifact for a specific platform.
type Download struct {
	URL          string   `json:"url"`
	Checksum     string   `json:"checksum"` // Format: "sha256:..."
	Size         int64    `json:"size,omitempty"`
	Runtime      string   `json:"runtime,omitempty"`      // e.g., "python3", "bash"
	Dependencies []string `json:"dependencies,omitempty"` // Runtime dependencies
}

// Repository represents a configured plugin repository.
type Repository struct {
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Enabled  bool      `json:"enabled"`
	Priority int       `json:"priority,omitempty"` // Lower numbers = higher priority
	Manifest *Manifest `json:"-"`                  // Cached manifest (not persisted)
}

// Config contains all configured repositories.
type Config struct {
	Repositories []*Repository `json:"repositories"`
	Cache        *CacheConfig  `json:"cache,omitempty"`
}

// CacheConfig contains cache settings.
type CacheConfig struct {
	TTL        int              `json:"ttl"`         // Cache TTL in seconds
	LastUpdate map[string]int64 `json:"last_update"` // repo name -> unix timestamp
}

// PluginSource represents where a plugin was installed from.
type PluginSource struct {
	Type         string `json:"type"` // "repository", "http", "local"
	Repository   string `json:"repository,omitempty"`
	Plugin       string `json:"plugin,omitempty"`
	Version      string `json:"version,omitempty"`
	URL          string `json:"url,omitempty"`
	Checksum     string `json:"checksum,omitempty"`
	OriginalPath string `json:"original_path,omitempty"`
}

// SearchFilter contains criteria for searching plugins.
type SearchFilter struct {
	Query      string
	Type       string // "input" or "output"
	Tags       []string
	Author     string
	Repository string
}

// SearchResult represents a plugin found in search.
type SearchResult struct {
	Plugin     *Plugin
	Version    *Version
	Repository string
}

// SyncStats tracks statistics for sync operations.
type SyncStats struct {
	Total     int
	Installed int
	Existing  int
	Failed    int
	Skipped   int
}

// VerifyResult represents the result of verifying a plugin.
type VerifyResult struct {
	Name     string
	Status   string // "valid", "mismatch", "missing", "no_checksum"
	Error    error
	Expected string
	Got      string
}
