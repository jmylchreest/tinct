// Package protocol defines the plugin protocol version and compatibility checking.
package protocol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jmylchreest/tinct/pkg/plugin"
)

// ProtocolVersion is an alias to the public plugin protocol version.
const ProtocolVersion = plugin.ProtocolVersion

// MinCompatibleVersion is an alias to the public minimum compatible version.
const MinCompatibleVersion = plugin.MinCompatibleVersion

// MinInputCompatibleVersion is an alias to the public input-plugin floor.
const MinInputCompatibleVersion = plugin.MinInputCompatibleVersion

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

// AtLeast returns true if v is greater than or equal to other.
func (v Version) AtLeast(other Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch >= other.Patch
}

// IsCompatible checks if a plugin protocol version is compatible with the current tinct version.
// Rules:
// - Major version must match exactly (breaking changes).
// - Minor version can be higher (backward compatible).
// - Patch version can be any value (bug fixes only).
func IsCompatible(pluginVersionStr string) (bool, error) {
	return isCompatibleAgainst(pluginVersionStr, MinCompatibleVersion)
}

// IsCompatibleForType is IsCompatible with the floor that applies to the
// given plugin type ("input" or "output").
//
// Input plugins carry a higher floor: the host calls WallpaperRawPath on
// every input plugin, and that method postdates 0.1.0, so an older
// plugin blocks the run instead of erroring. Output plugins have no such
// method and keep the general floor.
func IsCompatibleForType(pluginVersionStr, pluginType string) (bool, error) {
	return isCompatibleAgainst(pluginVersionStr, MinCompatibleVersionForType(pluginType))
}

// MinCompatibleVersionForType returns the protocol floor for a plugin
// type. Unknown types get the general floor.
func MinCompatibleVersionForType(pluginType string) string {
	if pluginType == "input" {
		return MinInputCompatibleVersion
	}
	return MinCompatibleVersion
}

func isCompatibleAgainst(pluginVersionStr, minCompatible string) (bool, error) {
	pluginVersion, err := Parse(pluginVersionStr)
	if err != nil {
		return false, fmt.Errorf("failed to parse plugin version: %w", err)
	}

	currentVersion, err := Parse(ProtocolVersion)
	if err != nil {
		return false, fmt.Errorf("failed to parse current protocol version: %w", err)
	}

	minVersion, err := Parse(minCompatible)
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
				minCompatible,
			)
		}
		if pluginVersion.Minor == minVersion.Minor && pluginVersion.Patch < minVersion.Patch {
			return false, fmt.Errorf(
				"plugin version %s is too old, minimum required is %s",
				pluginVersion.String(),
				minCompatible,
			)
		}
	}

	// Plugin can have higher minor/patch version (forward compatible)
	return true, nil
}

// FlagHelp is a type alias to the public plugin.FlagHelp type.
// External plugins should import github.com/jmylchreest/tinct/pkg/plugin directly.
type FlagHelp = plugin.FlagHelp
