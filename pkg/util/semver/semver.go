// Package semver provides semantic version parsing and comparison utilities.
// It supports comparing version strings like "0.53.0", "1.2.3-beta", etc.
package semver

import (
	"regexp"
	"strconv"
	"strings"
)

// Version represents a parsed semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string // Build metadata (ignored in comparisons per SemVer spec)
	Original   string
}

// versionRegex matches semantic versions like "0.53.0", "1.2.3-beta", "v1.0.0", "1.2.3+build"
// Format: [v]MAJOR[.MINOR][.PATCH][-PRERELEASE][+BUILD]
var versionRegex = regexp.MustCompile(`^v?(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z\-\.]+))?(?:\+([0-9A-Za-z\-\.]+))?$`)

// Parse parses a version string into a Version struct.
// It accepts versions like "0.53.0", "1.2", "1", "v1.0.0", "1.2.3-beta", "1.2.3+build".
// Build metadata (after +) is captured but ignored in comparisons per SemVer spec.
// Returns nil if the version string cannot be parsed.
func Parse(version string) *Version {
	version = strings.TrimSpace(version)
	if version == "" {
		return nil
	}

	matches := versionRegex.FindStringSubmatch(version)
	if matches == nil {
		return nil
	}

	v := &Version{
		Original: version,
	}

	// Major is always present if we matched
	v.Major, _ = strconv.Atoi(matches[1])

	// Minor is optional
	if matches[2] != "" {
		v.Minor, _ = strconv.Atoi(matches[2])
	}

	// Patch is optional
	if matches[3] != "" {
		v.Patch, _ = strconv.Atoi(matches[3])
	}

	// Prerelease is optional
	v.Prerelease = matches[4]

	// Build metadata is optional (captured but ignored in comparisons)
	if len(matches) > 5 {
		v.Build = matches[5]
	}

	return v
}

// String returns the version as a normalized string (without "v" prefix).
func (v *Version) String() string {
	s := strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor) + "." + strconv.Itoa(v.Patch)
	if v.Prerelease != "" {
		s += "-" + v.Prerelease
	}
	return s
}

// Compare compares two versions.
// Returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v *Version) Compare(other *Version) int {
	if other == nil {
		return 1
	}

	// Compare major
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	// Compare minor
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	// Compare patch
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// Compare prerelease: a version with prerelease is less than one without
	if v.Prerelease != "" && other.Prerelease == "" {
		return -1
	}
	if v.Prerelease == "" && other.Prerelease != "" {
		return 1
	}
	if v.Prerelease != other.Prerelease {
		if v.Prerelease < other.Prerelease {
			return -1
		}
		return 1
	}

	return 0
}

// LessThan returns true if v < other.
func (v *Version) LessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// LessThanOrEqual returns true if v <= other.
func (v *Version) LessThanOrEqual(other *Version) bool {
	return v.Compare(other) <= 0
}

// GreaterThan returns true if v > other.
func (v *Version) GreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// GreaterThanOrEqual returns true if v >= other.
func (v *Version) GreaterThanOrEqual(other *Version) bool {
	return v.Compare(other) >= 0
}

// Equal returns true if v == other.
func (v *Version) Equal(other *Version) bool {
	return v.Compare(other) == 0
}

// CompareStrings compares two version strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Returns 0 if either version cannot be parsed.
func CompareStrings(a, b string) int {
	va := Parse(a)
	vb := Parse(b)
	if va == nil || vb == nil {
		return 0
	}
	return va.Compare(vb)
}

// FindBestMatch finds the best matching version from a list of available versions.
// It returns the highest version that is <= targetVersion.
// If no suitable version is found, it returns an empty string.
func FindBestMatch(targetVersion string, availableVersions []string) string {
	target := Parse(targetVersion)
	if target == nil {
		return ""
	}

	var bestMatch *Version
	for _, v := range availableVersions {
		parsed := Parse(v)
		if parsed == nil {
			continue
		}

		// Skip versions higher than target
		if parsed.GreaterThan(target) {
			continue
		}

		// Keep track of the highest version that doesn't exceed target
		if bestMatch == nil || parsed.GreaterThan(bestMatch) {
			bestMatch = parsed
		}
	}

	if bestMatch == nil {
		return ""
	}
	return bestMatch.Original
}
