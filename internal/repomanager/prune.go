package repomanager

import (
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/plugin/protocol"
)

// PruneValidator checks if downloads should be kept during pruning.
type PruneValidator struct{}

// NewPruneValidator creates a new prune validator.
func NewPruneValidator() *PruneValidator {
	return &PruneValidator{}
}

// IsProtocolCompatible checks if a plugin protocol version is compatible with the current tinct version.
// It wraps the protocol.IsCompatible function for use in the repomanager package.
func IsProtocolCompatible(pluginVersion string) (bool, error) {
	// Strip common comparison operators from version string (e.g., ">=0.0.1" -> "0.0.1")
	pluginVersion = strings.TrimPrefix(pluginVersion, ">=")
	pluginVersion = strings.TrimPrefix(pluginVersion, "<=")
	pluginVersion = strings.TrimPrefix(pluginVersion, ">")
	pluginVersion = strings.TrimPrefix(pluginVersion, "<")
	pluginVersion = strings.TrimPrefix(pluginVersion, "=")
	pluginVersion = strings.TrimSpace(pluginVersion)

	compatible, err := protocol.IsCompatible(pluginVersion)
	if err != nil {
		return false, err
	}
	return compatible, nil
}

// IsProtocolCompatibleWithMin checks if a plugin protocol version meets a minimum version requirement.
// If minVersion is empty, it uses the current protocol version check (same as IsProtocolCompatible).
// Otherwise, it checks if the plugin version is >= minVersion.
func IsProtocolCompatibleWithMin(pluginVersion, minVersion string) (bool, error) {
	if minVersion == "" {
		return IsProtocolCompatible(pluginVersion)
	}

	// Strip common comparison operators from version strings (e.g., ">=0.0.1" -> "0.0.1")
	pluginVersion = strings.TrimPrefix(pluginVersion, ">=")
	pluginVersion = strings.TrimPrefix(pluginVersion, "<=")
	pluginVersion = strings.TrimPrefix(pluginVersion, ">")
	pluginVersion = strings.TrimPrefix(pluginVersion, "<")
	pluginVersion = strings.TrimPrefix(pluginVersion, "=")
	pluginVersion = strings.TrimSpace(pluginVersion)

	minVersion = strings.TrimPrefix(minVersion, ">=")
	minVersion = strings.TrimPrefix(minVersion, "<=")
	minVersion = strings.TrimPrefix(minVersion, ">")
	minVersion = strings.TrimPrefix(minVersion, "<")
	minVersion = strings.TrimPrefix(minVersion, "=")
	minVersion = strings.TrimSpace(minVersion)

	// Parse both versions
	pluginVer, err := protocol.Parse(pluginVersion)
	if err != nil {
		return false, err
	}

	minVer, err := protocol.Parse(minVersion)
	if err != nil {
		return false, err
	}

	// Compare versions: plugin must be >= minVersion
	if pluginVer.Major < minVer.Major {
		return false, nil
	}
	if pluginVer.Major > minVer.Major {
		return true, nil
	}

	// Same major version, check minor
	if pluginVer.Minor < minVer.Minor {
		return false, nil
	}
	if pluginVer.Minor > minVer.Minor {
		return true, nil
	}

	// Same major and minor, check patch
	if pluginVer.Patch < minVer.Patch {
		return false, nil
	}

	return true, nil
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
