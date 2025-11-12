package repomanager

import (
	"path/filepath"
	"strings"
)

// PruneValidator checks if downloads should be kept during pruning.
type PruneValidator struct{}

// NewPruneValidator creates a new prune validator.
func NewPruneValidator() *PruneValidator {
	return &PruneValidator{}
}

// ShouldKeepDownload checks if a download entry should be kept during pruning.
// It validates that the URL matches expected plugin asset patterns.
func (v *PruneValidator) ShouldKeepDownload(url string) (bool, string) {
	// Extract filename from URL
	filename := filepath.Base(url)

	// Check if it's a valid plugin asset
	if !isValidPluginDownload(filename) {
		return false, "not a valid plugin asset"
	}

	return true, ""
}

// isValidPluginDownload checks if a filename represents a valid plugin download.
// This uses the same logic as isPluginAsset but for pruning validation.
func isValidPluginDownload(filename string) bool {
	// Must be an archive or script
	hasValidExtension := false
	for _, ext := range []string{".tar.gz", ".tar.xz", ".tar.bz2", ".zip", ".py", ".sh", ".rb", ".js"} {
		if strings.HasSuffix(filename, ext) {
			hasValidExtension = true
			break
		}
	}

	if !hasValidExtension {
		return false
	}

	// Exclude known non-plugin patterns
	excludePatterns := []string{
		".sbom.json", // SBOM files
		"sbom",       // Generic SBOM
		"checksums",  // Checksum files
		"checksum",   // Singular checksum
		"provenance", // Supply chain provenance
		"metadata",   // Metadata files
		"_metadata",  // Underscore metadata
		".sig",       // Signatures
		".asc",       // ASCII armor signatures
		".pem",       // Certificates
	}

	lowerFilename := strings.ToLower(filename)

	for _, pattern := range excludePatterns {
		if strings.Contains(lowerFilename, pattern) {
			return false
		}
	}

	return true
}
