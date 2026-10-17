package main

import "strings"

// isValidPath validates a literal process argument.
func isValidPath(path string) bool { return path != "" && !strings.ContainsRune(path, 0) }

// isValidExecutableName checks if a string is a safe executable name (no path separators).
func isValidExecutableName(name string) bool {
	// Reject if it contains path separators or suspicious characters
	if strings.ContainsAny(name, "/\\|&;`$()") {
		return false
	}
	// Should not be empty
	return name != ""
}
