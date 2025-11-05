// Package protocol defines the plugin protocol version and compatibility checking.
package protocol

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	// ProtocolVersion defines the current plugin API version.
	// Format: MAJOR.MINOR.PATCH.
	// - Increment MAJOR for breaking changes (incompatible API changes).
	// - Increment MINOR for backward-compatible additions.
	// - Increment PATCH for backward-compatible bug fixes.
	ProtocolVersion = "0.0.1"

	// MinCompatibleVersion is the oldest protocol version this tinct version can work with.
	MinCompatibleVersion = "0.0.1"
)

// Version represents a parsed protocol version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// Parse parses a version string in "MAJOR.MINOR.PATCH" format.
func Parse(version string) (Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version format: %s (expected MAJOR.MINOR.PATCH)", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// String returns the string representation of the version.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsCompatible checks if a plugin protocol version is compatible with the current tinct version.
// Rules:
// - Major version must match exactly (breaking changes).
// - Minor version can be higher (backward compatible).
// - Patch version can be any value (bug fixes only).
func IsCompatible(pluginVersionStr string) (bool, error) {
	pluginVersion, err := Parse(pluginVersionStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse plugin version: %w", err)
	}

	currentVersion, err := Parse(ProtocolVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse current protocol version: %w", err)
	}

	minVersion, err := Parse(MinCompatibleVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse minimum compatible version: %w", err)
	}

	// Major version must match exactly
	if pluginVersion.Major != currentVersion.Major {
		return false, fmt.Errorf(
			"incompatible major version: plugin is %s, tinct requires %d.x.x",
			pluginVersion.String(),
			currentVersion.Major,
		)
	}

	// Check if version is below minimum compatible version
	if pluginVersion.Major == minVersion.Major {
		if pluginVersion.Minor < minVersion.Minor {
			return false, fmt.Errorf(
				"plugin version %s is too old, minimum required is %s",
				pluginVersion.String(),
				MinCompatibleVersion,
			)
		}
		if pluginVersion.Minor == minVersion.Minor && pluginVersion.Patch < minVersion.Patch {
			return false, fmt.Errorf(
				"plugin version %s is too old, minimum required is %s",
				pluginVersion.String(),
				MinCompatibleVersion,
			)
		}
	}

	// Plugin can have higher minor/patch version (forward compatible)
	return true, nil
}

// GetCurrentVersion returns the current protocol version as a Version struct.
func GetCurrentVersion() Version {
	v, err := Parse(ProtocolVersion)
	if err != nil {
		// This should never happen since ProtocolVersion is a constant with valid format.
		panic(fmt.Sprintf("invalid ProtocolVersion constant: %v", err))
	}
	return v
}
