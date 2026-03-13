package utils

import (
	"fmt"
	"os"
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

// MakeScriptExecutable searches writtenFiles for a file with the given base
// name and makes it executable (0o750). Errors are logged as warnings when
// verbose is true but never returned — a non-executable script is not fatal.
func MakeScriptExecutable(writtenFiles []string, scriptName string, verbose bool) {
	for _, f := range writtenFiles {
		if filepath.Base(f) == scriptName {
			if err := os.Chmod(f, 0o750); err != nil && verbose { // #nosec G302 -- 0o750 is intentional: generated shell scripts must be executable by the owner
				fmt.Fprintf(os.Stderr, "   Warning: failed to make %s executable: %v\n", f, err)
			}
			break
		}
	}
}
