package appdetect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsPresentAny_Binaries(t *testing.T) {
	// Should find at least one existing binary
	if !IsPresentAny([]string{"ls", "nonexistent"}, nil) {
		t.Error("IsPresentAny() should find 'ls' in list of binaries")
	}

	// Should return false when all binaries don't exist
	if IsPresentAny([]string{"nonexistent1", "nonexistent2"}, nil) {
		t.Error("IsPresentAny() should return false when no binaries exist")
	}

	// Should handle empty slice
	if IsPresentAny([]string{}, nil) {
		t.Error("IsPresentAny() should return false for empty slices")
	}

	// Should skip empty strings
	if !IsPresentAny([]string{"", "ls", ""}, nil) {
		t.Error("IsPresentAny() should skip empty strings and find 'ls'")
	}
}

func TestIsPresentAny_Directories(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDir2 := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(tmpDir2, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Should find at least one existing directory
	if !IsPresentAny(nil, []string{tmpDir, "/nonexistent"}) {
		t.Errorf("IsPresentAny() should find existing directory: %s", tmpDir)
	}

	// Should return false when all directories don't exist
	if IsPresentAny(nil, []string{"/nonexistent1", "/nonexistent2"}) {
		t.Error("IsPresentAny() should return false when no directories exist")
	}

	// Should skip empty strings
	if !IsPresentAny(nil, []string{"", tmpDir, ""}) {
		t.Errorf("IsPresentAny() should skip empty strings and find directory: %s", tmpDir)
	}
}

func TestIsPresentAny_BinariesAndDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Should succeed if binary exists (even if dir doesn't)
	if !IsPresentAny([]string{"ls"}, []string{"/nonexistent"}) {
		t.Error("IsPresentAny() should return true when binary exists")
	}

	// Should succeed if directory exists (even if binary doesn't)
	if !IsPresentAny([]string{"nonexistent"}, []string{tmpDir}) {
		t.Errorf("IsPresentAny() should return true when directory exists: %s", tmpDir)
	}

	// Should fail if neither exists
	if IsPresentAny([]string{"nonexistent"}, []string{"/nonexistent"}) {
		t.Error("IsPresentAny() should return false when both checks fail")
	}
}

func TestIsPresentAll_Binaries(t *testing.T) {
	// Should require all binaries to exist
	if !IsPresentAll([]string{"ls", "cat"}, nil) {
		t.Error("IsPresentAll() should find both 'ls' and 'cat'")
	}

	// Should fail if any binary doesn't exist
	if IsPresentAll([]string{"ls", "nonexistent"}, nil) {
		t.Error("IsPresentAll() should return false when any binary doesn't exist")
	}

	// Should handle empty slice
	if !IsPresentAll([]string{}, nil) {
		t.Error("IsPresentAll() should return true for empty slices (vacuous truth)")
	}

	// Should skip empty strings
	if !IsPresentAll([]string{"", "ls", ""}, nil) {
		t.Error("IsPresentAll() should skip empty strings and check only 'ls'")
	}
}

func TestIsPresentAll_Directories(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDir2 := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(tmpDir2, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Should require all directories to exist
	if !IsPresentAll(nil, []string{tmpDir, tmpDir2}) {
		t.Errorf("IsPresentAll() should find both directories: %s, %s", tmpDir, tmpDir2)
	}

	// Should fail if any directory doesn't exist
	if IsPresentAll(nil, []string{tmpDir, "/nonexistent"}) {
		t.Error("IsPresentAll() should return false when any directory doesn't exist")
	}

	// Should skip empty strings
	if !IsPresentAll(nil, []string{"", tmpDir, ""}) {
		t.Errorf("IsPresentAll() should skip empty strings and check only: %s", tmpDir)
	}
}

func TestIsPresentAll_BinariesAndDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Should require both binary and directory to exist
	if !IsPresentAll([]string{"ls"}, []string{tmpDir}) {
		t.Errorf("IsPresentAll() should return true when both exist: ls, %s", tmpDir)
	}

	// Should fail if binary doesn't exist (even if dir does)
	if IsPresentAll([]string{"nonexistent"}, []string{tmpDir}) {
		t.Error("IsPresentAll() should return false when binary doesn't exist")
	}

	// Should fail if directory doesn't exist (even if binary does)
	if IsPresentAll([]string{"ls"}, []string{"/nonexistent"}) {
		t.Error("IsPresentAll() should return false when directory doesn't exist")
	}

	// Should fail if neither exists
	if IsPresentAll([]string{"nonexistent"}, []string{"/nonexistent"}) {
		t.Error("IsPresentAll() should return false when both don't exist")
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"~/.config", filepath.Join(home, ".config")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, test := range tests {
		result := expandPath(test.input)
		if result != test.expected {
			t.Errorf("expandPath(%q) = %q, want %q", test.input, result, test.expected)
		}
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		str     string
		pattern string
		want    bool
	}{
		{"app.devsuite.Ptyxis", "app.devsuite.Ptyxis", true},
		{"app.devsuite.Ptyxis", "app.devsuite.*", true},
		{"org.gnome.Terminal", "org.gnome.*", true},
		{"something.else", "org.gnome.*", false},
		{"exact", "exact", true},
		{"exact", "different", false},
	}

	for _, test := range tests {
		got := matchesPattern(test.str, test.pattern)
		if got != test.want {
			t.Errorf("matchesPattern(%q, %q) = %v, want %v",
				test.str, test.pattern, got, test.want)
		}
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file (not a directory)
	tmpFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		path string
		want bool
	}{
		{tmpDir, true},   // Existing directory
		{tmpFile, false}, // File, not directory
		{filepath.Join(tmpDir, "nonexistent"), false}, // Non-existent path
	}

	for _, test := range tests {
		got := dirExists(test.path)
		if got != test.want {
			t.Errorf("dirExists(%q) = %v, want %v", test.path, got, test.want)
		}
	}
}

func TestGetInstallationType(t *testing.T) {
	// Test native binary detection
	installType := GetInstallationType([]string{"ls"}, nil)
	if installType != "native" {
		t.Errorf("GetInstallationType([\"ls\"], nil) = %q, want \"native\"", installType)
	}

	// Test directory detection
	tmpDir := t.TempDir()
	installType = GetInstallationType(nil, []string{tmpDir})
	if installType != "directory" {
		t.Errorf("GetInstallationType(nil, [tmpDir]) = %q, want \"directory\"", installType)
	}

	// Test not found
	installType = GetInstallationType([]string{"nonexistent"}, []string{"/nonexistent"})
	if installType != "" {
		t.Errorf("GetInstallationType([\"nonexistent\"], [\"/nonexistent\"]) = %q, want \"\"", installType)
	}

	// Test priority: binary checks happen before directory
	installType = GetInstallationType([]string{"ls"}, []string{tmpDir})
	if installType != "native" {
		t.Errorf("GetInstallationType([\"ls\"], [tmpDir]) = %q, want \"native\" (binary should take precedence)", installType)
	}
}
