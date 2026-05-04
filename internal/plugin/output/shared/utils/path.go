package utils

import (
	"path/filepath"
	"strings"
)

// IsValidPath checks if a path is safe to use in exec.Command arguments.
// It rejects paths containing directory traversal (..) or shell metacharacters
// that could be exploited if the path were ever passed through a shell.
// This is defense-in-depth — exec.Command does not invoke a shell, but we
// guard against future misuse.
func IsValidPath(path string) bool {
	if strings.Contains(path, "..") || strings.ContainsAny(path, "|&;`$()") {
		return false
	}
	cleaned := filepath.Clean(path)
	return cleaned == path
}
