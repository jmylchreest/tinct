package repomanager

import (
	"testing"
)

func TestFilter_Match(t *testing.T) {
	tests := []struct {
		name         string
		include      []string
		exclude      []string
		assetName    string
		expectMatch  bool
		expectPlugin string
	}{
		{
			name:         "simple match",
			include:      []string{"tinct-plugin-*"},
			exclude:      []string{},
			assetName:    "tinct-plugin-random_1.0.0_Linux_x86_64.tar.gz",
			expectMatch:  true,
			expectPlugin: "random",
		},
		{
			name:         "no match",
			include:      []string{"tinct-plugin-*"},
			exclude:      []string{},
			assetName:    "tinct_1.0.0_Linux_x86_64.tar.gz",
			expectMatch:  false,
			expectPlugin: "",
		},
		{
			name:         "excluded pattern",
			include:      []string{"tinct-plugin-*"},
			exclude:      []string{"*test*"},
			assetName:    "tinct-plugin-test_1.0.0_Linux_x86_64.tar.gz",
			expectMatch:  false,
			expectPlugin: "",
		},
		{
			name:         "not excluded",
			include:      []string{"tinct-plugin-*"},
			exclude:      []string{"*test*"},
			assetName:    "tinct-plugin-random_1.0.0_Linux_x86_64.tar.gz",
			expectMatch:  true,
			expectPlugin: "random",
		},
		{
			name:         "sbom files excluded",
			include:      []string{"tinct-plugin-*"},
			exclude:      []string{"*.sbom.json"},
			assetName:    "tinct-plugin-random_1.0.0_Linux_x86_64.tar.gz.sbom.json",
			expectMatch:  false,
			expectPlugin: "",
		},
		{
			name:         "checksums excluded",
			include:      []string{"tinct-plugin-*"},
			exclude:      []string{"checksums.txt"},
			assetName:    "checksums.txt",
			expectMatch:  false,
			expectPlugin: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewFilter(tt.include, tt.exclude)
			match, plugin := filter.Match(tt.assetName)

			if match != tt.expectMatch {
				t.Errorf("Expected match=%v, got match=%v", tt.expectMatch, match)
			}

			if plugin != tt.expectPlugin {
				t.Errorf("Expected plugin='%s', got plugin='%s'", tt.expectPlugin, plugin)
			}
		})
	}
}

func TestNewFilter(t *testing.T) {
	include := []string{"tinct-plugin-*"}
	exclude := []string{"*test*"}

	filter := NewFilter(include, exclude)

	if filter == nil {
		t.Error("Expected non-nil filter")
	}

	if len(filter.Include) != 1 {
		t.Errorf("Expected 1 include pattern, got %d", len(filter.Include))
	}

	if len(filter.Exclude) != 1 {
		t.Errorf("Expected 1 exclude pattern, got %d", len(filter.Exclude))
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected bool
	}{
		{"exact match", "test.txt", "test.txt", true},
		{"wildcard match", "test.txt", "*.txt", true},
		{"prefix wildcard", "test.txt", "test*", true},
		{"no match", "test.txt", "*.md", false},
		{"complex wildcard", "tinct-plugin-random.tar.gz", "tinct-plugin-*", true},
		{"invalid pattern", "test.txt", "[invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPattern(tt.text, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchPattern(%s, %s) = %v, expected %v", tt.text, tt.pattern, result, tt.expected)
			}
		})
	}
}
