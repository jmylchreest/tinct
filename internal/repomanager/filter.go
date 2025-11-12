package repomanager

import (
	"path/filepath"
	"strings"
)

// Filter handles artifact name filtering.
type Filter struct {
	Include []string // Patterns to include (e.g., "tinct-plugin-*")
	Exclude []string // Patterns to exclude (e.g., "tinct_*")
}

// NewFilter creates a new artifact filter.
func NewFilter(include, exclude []string) *Filter {
	return &Filter{
		Include: include,
		Exclude: exclude,
	}
}

// Match checks if an asset name matches the filter and returns the plugin name.
func (f *Filter) Match(assetName string) (bool, string) {
	// Skip non-plugin assets (SBOM, checksums, etc.)
	if !isPluginAsset(assetName) {
		return false, ""
	}

	// Check exclusions first
	for _, pattern := range f.Exclude {
		if matchPattern(assetName, pattern) {
			return false, ""
		}
	}

	// Check inclusions
	for _, pattern := range f.Include {
		if matchPattern(assetName, pattern) {
			pluginName := extractPluginName(assetName)
			return true, pluginName
		}
	}

	return false, ""
}

// isPluginAsset checks if an asset is a plugin binary/archive.
func isPluginAsset(name string) bool {
	// Must end with supported archive format
	if !strings.HasSuffix(name, ".tar.gz") &&
		!strings.HasSuffix(name, ".tar.bz2") &&
		!strings.HasSuffix(name, ".zip") {
		return false
	}

	// Exclude known non-plugin files
	excludePatterns := []string{
		"sbom",       // Software Bill of Materials
		"checksums",  // Checksum files
		"provenance", // Supply chain provenance
		"metadata",   // Metadata files
		".json",      // JSON files (but .tar.gz with .json in name is ok)
		".txt",       // Text files
	}

	// Check if pattern appears before the extension
	nameWithoutExt := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(name, ".tar.gz"), ".tar.bz2"), ".zip")
	lowerNameWithoutExt := strings.ToLower(nameWithoutExt)

	for _, pattern := range excludePatterns {
		if strings.Contains(lowerNameWithoutExt, pattern) {
			return false
		}
	}

	return true
}

// matchPattern performs glob-style pattern matching.
func matchPattern(name, pattern string) bool {
	matched, _ := filepath.Match(pattern, name)
	return matched
}

// extractPluginName extracts the plugin name from an artifact name.
// "tinct-plugin-random_1.2.3_Linux_x86_64.tar.gz" -> "random"
func extractPluginName(assetName string) string {
	// Remove prefix (tinct-plugin-)
	name := strings.TrimPrefix(assetName, "tinct-plugin-")

	// Remove extension
	name = strings.TrimSuffix(name, ".tar.gz")
	name = strings.TrimSuffix(name, ".zip")

	// Split on underscore and take first part (plugin name)
	// "random_1.2.3_Linux_x86_64" -> "random"
	parts := strings.Split(name, "_")
	if len(parts) > 0 {
		return parts[0]
	}

	return name
}

// ParseAssetName extracts version and platform from asset name.
// "tinct-plugin-random_1.2.3_Linux_x86_64.tar.gz" -> ("1.2.3", "linux_amd64")
func ParseAssetName(assetName string) (version, platform string) {
	// Remove extension
	name := strings.TrimSuffix(assetName, ".tar.gz")
	name = strings.TrimSuffix(name, ".zip")

	// Remove prefix
	name = strings.TrimPrefix(name, "tinct-plugin-")

	// Split by underscore
	// Expected: NAME_VERSION_OS_ARCH
	parts := strings.Split(name, "_")

	if len(parts) >= 4 {
		version = parts[1]              // 1.2.3
		os := strings.ToLower(parts[2]) // Linux -> linux
		arch := parts[3]                // x86_64

		// Normalize platform
		platform = normalizePlatform(os, arch)
	}

	return
}

// normalizePlatform converts OS/arch to standard platform identifiers.
// "Linux", "x86_64" -> "linux_amd64"
func normalizePlatform(os, arch string) string {
	// Normalize arch
	switch arch {
	case "x86_64":
		arch = "amd64"
	case "i386", "i686":
		arch = "386"
	case "armv7", "armv7l":
		arch = "arm"
	}

	return strings.ToLower(os) + "_" + arch
}
