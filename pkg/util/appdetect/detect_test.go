package appdetect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsPresent_NativeBinary(t *testing.T) {
	// Test with a binary that should exist on most systems
	if !IsPresent("ls", "") {
		t.Error("IsPresent() should detect 'ls' binary on PATH")
	}

	// Test with a binary that definitely doesn't exist
	if IsPresent("definitely-not-a-real-binary-name-12345", "") {
		t.Error("IsPresent() should not detect non-existent binary")
	}
}

func TestIsPresent_ConfigDir(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Test with existing directory
	if !IsPresent("", tmpDir) {
		t.Errorf("IsPresent() should detect existing directory: %s", tmpDir)
	}

	// Test with non-existent directory
	if IsPresent("", filepath.Join(tmpDir, "nonexistent")) {
		t.Error("IsPresent() should not detect non-existent directory")
	}
}

func TestIsPresent_BothMethods(t *testing.T) {
	tmpDir := t.TempDir()

	// Should succeed if either binary or dir exists
	if !IsPresent("ls", tmpDir) {
		t.Error("IsPresent() should return true when binary exists (even if dir check would fail)")
	}

	// Should fail if neither exists
	if IsPresent("nonexistent-binary", filepath.Join(tmpDir, "nonexistent")) {
		t.Error("IsPresent() should return false when both checks fail")
	}
}

func TestIsPresent_EmptyParams(t *testing.T) {
	// Should return false if both params are empty
	if IsPresent("", "") {
		t.Error("IsPresent() should return false when both parameters are empty")
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
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
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
	installType := GetInstallationType("ls", "")
	if installType != "native" {
		t.Errorf("GetInstallationType(\"ls\", \"\") = %q, want \"native\"", installType)
	}

	// Test config-dir detection
	tmpDir := t.TempDir()
	installType = GetInstallationType("", tmpDir)
	if installType != "config-dir" {
		t.Errorf("GetInstallationType(\"\", tmpDir) = %q, want \"config-dir\"", installType)
	}

	// Test not found
	installType = GetInstallationType("nonexistent", "/nonexistent")
	if installType != "" {
		t.Errorf("GetInstallationType(nonexistent) = %q, want \"\"", installType)
	}
}
