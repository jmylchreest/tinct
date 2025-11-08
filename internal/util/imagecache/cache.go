// Package imagecache provides utilities for downloading and caching remote images.
package imagecache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	httputil "github.com/jmylchreest/tinct/internal/util/http"
)

// CacheOptions configures image caching behavior.
type CacheOptions struct {
	// CacheDir is the directory where images will be cached.
	// If empty, defaults to ~/.cache/tinct/images
	CacheDir string

	// Filename is the filename to use for the cached image.
	// If empty, uses a hash of the URL + original extension.
	Filename string

	// AllowOverwrite determines if existing cached files can be overwritten.
	// Default: false (reuse existing cached files).
	AllowOverwrite bool
}

// DefaultCacheDir returns the default cache directory path.
func DefaultCacheDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to home directory if cache dir not available.
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to determine cache directory: %w", err)
		}
		return filepath.Join(home, ".cache", "tinct", "images"), nil
	}
	return filepath.Join(cacheDir, "tinct", "images"), nil
}

// generateFilename creates a deterministic filename from a URL.
// Uses SHA256 hash of URL + original file extension.
func generateFilename(url string) string {
	// Hash the URL.
	hash := sha256.Sum256([]byte(url))
	hashStr := fmt.Sprintf("%x", hash[:16]) // Use first 16 bytes (32 hex chars)

	// Extract extension from URL (if present).
	ext := filepath.Ext(url)
	// Remove query parameters from extension.
	if idx := strings.IndexByte(ext, '?'); idx != -1 {
		ext = ext[:idx]
	}
	// Default to .jpg if no extension found.
	if ext == "" || len(ext) > 5 {
		ext = ".jpg"
	}

	return hashStr + ext
}

// DownloadAndCache downloads a remote image and saves it to the cache directory.
// Returns the local file path where the image was saved.
func DownloadAndCache(ctx context.Context, url string, opts CacheOptions) (string, error) {
	// Validate URL.
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return "", fmt.Errorf("invalid URL: must start with http:// or https://")
	}

	// Determine cache directory.
	cacheDir := opts.CacheDir
	if cacheDir == "" {
		defaultDir, err := DefaultCacheDir()
		if err != nil {
			return "", fmt.Errorf("failed to determine cache directory: %w", err)
		}
		cacheDir = defaultDir
	}

	// Create cache directory if it doesn't exist.
	if err := os.MkdirAll(cacheDir, 0o755); err != nil { // #nosec G301 - Cache directory needs standard permissions
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Determine filename.
	filename := opts.Filename
	if filename == "" {
		filename = generateFilename(url)
	}

	// Full path to cached file.
	cachedPath := filepath.Join(cacheDir, filename)

	// Check if file already exists.
	if !opts.AllowOverwrite {
		if _, err := os.Stat(cachedPath); err == nil {
			// File exists - return cached path.
			return cachedPath, nil
		}
	}

	// Download the image.
	data, err := httputil.Fetch(ctx, url, httputil.FetchOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}

	// Write to cache file.
	if err := os.WriteFile(cachedPath, data, 0o644); err != nil { // #nosec G306 - Cache files need standard read permissions
		return "", fmt.Errorf("failed to write cached image: %w", err)
	}

	return cachedPath, nil
}
