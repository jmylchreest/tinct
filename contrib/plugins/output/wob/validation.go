package main

import (
	"path/filepath"
	"strings"
)

// isValidPath checks if a path is safe to use in commands.
func isValidPath(path string) bool {
	// Reject paths with suspicious characters
	if strings.Contains(path, "..") || strings.ContainsAny(path, "|&;`$()") {
		return false
	}
	// Clean the path and ensure it matches
	cleaned := filepath.Clean(path)
	return cleaned == path
}

// isValidExecutableName checks if a string is a safe executable name (no path separators).
func isValidExecutableName(name string) bool {
	// Reject if it contains path separators or suspicious characters
	if strings.ContainsAny(name, "/\\|&;`$()") {
		return false
	}
	// Should not be empty
	return name != ""
}
