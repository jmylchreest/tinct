package semver

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		expected *Version
	}{
		{"0.53.0", &Version{Major: 0, Minor: 53, Patch: 0, Original: "0.53.0"}},
		{"1.2.3", &Version{Major: 1, Minor: 2, Patch: 3, Original: "1.2.3"}},
		{"v1.0.0", &Version{Major: 1, Minor: 0, Patch: 0, Original: "v1.0.0"}},
		{"1.2", &Version{Major: 1, Minor: 2, Patch: 0, Original: "1.2"}},
		{"1", &Version{Major: 1, Minor: 0, Patch: 0, Original: "1"}},
		{"1.2.3-beta", &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta", Original: "1.2.3-beta"}},
		{"1.0.0-rc.1", &Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "rc.1", Original: "1.0.0-rc.1"}},
		// GitVersion-style with build metadata
		{"1.2.3+build.123", &Version{Major: 1, Minor: 2, Patch: 3, Build: "build.123", Original: "1.2.3+build.123"}},
		{"1.2.3-beta.1+build.123", &Version{Major: 1, Minor: 2, Patch: 3, Prerelease: "beta.1", Build: "build.123", Original: "1.2.3-beta.1+build.123"}},
		{"1.0.0+20130313144700", &Version{Major: 1, Minor: 0, Patch: 0, Build: "20130313144700", Original: "1.0.0+20130313144700"}},
		{"1.0.0-beta+exp.sha.5114f85", &Version{Major: 1, Minor: 0, Patch: 0, Prerelease: "beta", Build: "exp.sha.5114f85", Original: "1.0.0-beta+exp.sha.5114f85"}},
		{"", nil},
		{"invalid", nil},
		{"a.b.c", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := Parse(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Parse(%q) = %v, want nil", tt.input, result)
				}
				return
			}
			if result == nil {
				t.Errorf("Parse(%q) = nil, want %v", tt.input, tt.expected)
				return
			}
			if result.Major != tt.expected.Major ||
				result.Minor != tt.expected.Minor ||
				result.Patch != tt.expected.Patch ||
				result.Prerelease != tt.expected.Prerelease ||
				result.Build != tt.expected.Build {

				t.Errorf("Parse(%q) = {%d.%d.%d-%s+%s}, want {%d.%d.%d-%s+%s}",
					tt.input,
					result.Major, result.Minor, result.Patch, result.Prerelease, result.Build,
					tt.expected.Major, tt.expected.Minor, tt.expected.Patch, tt.expected.Prerelease, tt.expected.Build)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.1", "1.0.0", 1},
		{"0.53.0", "0.52.0", 1},
		{"0.52.0", "0.53.0", -1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"v1.0.0", "1.0.0", 0},
		// Build metadata should be ignored in comparisons (per SemVer spec)
		{"1.0.0+build.1", "1.0.0+build.2", 0},
		{"1.0.0+build.1", "1.0.0", 0},
		{"1.0.0-alpha+build.1", "1.0.0-alpha+build.2", 0},
		{"1.0.0-alpha+build", "1.0.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			a := Parse(tt.a)
			b := Parse(tt.b)
			result := a.Compare(b)
			if result != tt.expected {
				t.Errorf("Compare(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestFindBestMatch(t *testing.T) {
	tests := []struct {
		target    string
		available []string
		expected  string
	}{
		{
			target:    "0.53.0",
			available: []string{"0.50", "0.52", "0.53", "0.54"},
			expected:  "0.53",
		},
		{
			target:    "0.53.0",
			available: []string{"0.50", "0.52"},
			expected:  "0.52",
		},
		{
			target:    "0.53.0",
			available: []string{"0.54", "0.55"},
			expected:  "",
		},
		{
			target:    "1.0.0",
			available: []string{"0.9", "1.0", "1.1"},
			expected:  "1.0",
		},
		{
			target:    "0.52.5",
			available: []string{"0.52", "0.53"},
			expected:  "0.52",
		},
		{
			target:    "0.53.0",
			available: []string{},
			expected:  "",
		},
		{
			target:    "",
			available: []string{"0.50", "0.52"},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			result := FindBestMatch(tt.target, tt.available)
			if result != tt.expected {
				t.Errorf("FindBestMatch(%q, %v) = %q, want %q",
					tt.target, tt.available, result, tt.expected)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3", "1.2.3"},
		{"v1.0.0", "1.0.0"},
		{"1.2", "1.2.0"},
		{"1", "1.0.0"},
		{"1.0.0-beta", "1.0.0-beta"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v := Parse(tt.input)
			if v == nil {
				t.Fatalf("Parse(%q) returned nil", tt.input)
			}
			result := v.String()
			if result != tt.expected {
				t.Errorf("Version.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestComparisonMethods(t *testing.T) {
	v1 := Parse("1.0.0")
	v2 := Parse("2.0.0")

	if !v1.LessThan(v2) {
		t.Error("1.0.0 should be less than 2.0.0")
	}
	if !v1.LessThanOrEqual(v2) {
		t.Error("1.0.0 should be less than or equal to 2.0.0")
	}
	if v1.GreaterThan(v2) {
		t.Error("1.0.0 should not be greater than 2.0.0")
	}
	if v1.GreaterThanOrEqual(v2) {
		t.Error("1.0.0 should not be greater than or equal to 2.0.0")
	}
	if v1.Equal(v2) {
		t.Error("1.0.0 should not equal 2.0.0")
	}

	v3 := Parse("1.0.0")
	if !v1.Equal(v3) {
		t.Error("1.0.0 should equal 1.0.0")
	}
	if !v1.LessThanOrEqual(v3) {
		t.Error("1.0.0 should be less than or equal to 1.0.0")
	}
	if !v1.GreaterThanOrEqual(v3) {
		t.Error("1.0.0 should be greater than or equal to 1.0.0")
	}
}
