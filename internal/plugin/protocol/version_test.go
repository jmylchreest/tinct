package protocol

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		version     string
		expectError bool
		major       int
		minor       int
		patch       int
	}{
		{"0.0.1", false, 0, 0, 1},
		{"1.0.0", false, 1, 0, 0},
		{"2.5.3", false, 2, 5, 3},
		{"10.99.42", false, 10, 99, 42},
		{"invalid", true, 0, 0, 0},
		{"1", true, 0, 0, 0},
		{"1.2", true, 0, 0, 0},
	}

	for _, tt := range tests {
		v, err := Parse(tt.version)
		if tt.expectError {
			if err == nil {
				t.Errorf("Parse(%q) expected error but got none", tt.version)
			}
		} else {
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.version, err)
			}
			if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
				t.Errorf("Parse(%q) = %d.%d.%d, want %d.%d.%d", tt.version, v.Major, v.Minor, v.Patch, tt.major, tt.minor, tt.patch)
			}
		}
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		pluginVersion string
		compatible    bool
		errorContains string
	}{
		// At the general floor - compatible
		{MinCompatibleVersion, true, ""},

		// Same major, higher minor - compatible (forward compatible)
		{"0.1.0", true, ""},
		{"0.5.2", true, ""},

		// Same major.minor, higher patch - compatible
		{"0.0.2", true, ""},
		{"0.0.10", true, ""},

		// Different major version - incompatible
		{"1.0.0", false, "incompatible major version"},
		{"2.0.0", false, "incompatible major version"},

		// Invalid format
		{"invalid", false, "failed to parse"},
		{"1", false, "invalid version format"},
		{"1.2", false, "invalid version format"},
	}

	for _, tt := range tests {
		compatible, err := IsCompatible(tt.pluginVersion)

		if !tt.compatible {
			if compatible {
				t.Errorf("IsCompatible(%q) = true, want false", tt.pluginVersion)
			}
			if err != nil && tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("IsCompatible(%q) error = %q, want error containing %q",
					tt.pluginVersion, err.Error(), tt.errorContains)
			}
			continue
		}

		if !compatible {
			t.Errorf("IsCompatible(%q) = false, want true", tt.pluginVersion)
		}
		if err != nil {
			t.Errorf("IsCompatible(%q) unexpected error: %v", tt.pluginVersion, err)
		}
	}
}

func TestVersionString(t *testing.T) {
	v := Version{Major: 0, Minor: 0, Patch: 1}
	if v.String() != "0.0.1" {
		t.Errorf("Version.String() = %q, want %q", v.String(), "0.0.1")
	}

	v2 := Version{Major: 1, Minor: 5, Patch: 3}
	if v2.String() != "1.5.3" {
		t.Errorf("Version.String() = %q, want %q", v2.String(), "1.5.3")
	}
}

// The input floor is higher than the general one: the host calls
// WallpaperRawPath on every input plugin, and that method postdates
// 0.1.0, so an older input plugin blocks the run rather than erroring.
func TestIsCompatibleForType(t *testing.T) {
	tests := []struct {
		version    string
		pluginType string
		compatible bool
	}{
		// Output plugins keep the general floor — every RPC the host
		// calls on them unconditionally predates 0.1.0, so these work.
		{"0.0.1", "output", true},
		{"0.1.0", "output", true},
		{"0.2.0", "output", true},

		// Input plugins must be at least 0.2.0.
		{"0.0.1", "input", false},
		{"0.1.0", "input", false},
		{"0.2.0", "input", true},
		{"0.3.0", "input", true},

		// Unknown types fall back to the general floor.
		{"0.1.0", "", true},
	}

	for _, tt := range tests {
		compatible, err := IsCompatibleForType(tt.version, tt.pluginType)
		if compatible != tt.compatible {
			t.Errorf("IsCompatibleForType(%q, %q) = %v (err %v), want %v",
				tt.version, tt.pluginType, compatible, err, tt.compatible)
		}
	}
}

func TestMinCompatibleVersionForType(t *testing.T) {
	if got := MinCompatibleVersionForType("input"); got != MinInputCompatibleVersion {
		t.Errorf("input floor = %q, want %q", got, MinInputCompatibleVersion)
	}
	for _, other := range []string{"output", "", "unknown"} {
		if got := MinCompatibleVersionForType(other); got != MinCompatibleVersion {
			t.Errorf("%q floor = %q, want %q", other, got, MinCompatibleVersion)
		}
	}
}
