// Package appdetect provides utilities for detecting installed applications
// across different installation methods (native, Flatpak, AppImage, etc.).
//
// This package is part of the public tinct API and can be used by both
// internal components and external plugins to check for application presence
// before generating configuration files.
package appdetect

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsPresentAny checks if ANY of the specified binaries OR directories exist.
// This is useful for checking if at least one of several optional components is present.
//
// Parameters:
//   - binaries: List of binary names to check (can be nil or empty)
//   - directories: List of directories to check (can be nil or empty)
//
// Returns true if at least one binary OR directory is found.
// Empty string entries in slices are ignored.
//
// Examples:
//
//	IsPresentAny([]string{"ptyxis", "kgx"}, nil)                   // Check for either binary
//	IsPresentAny(nil, []string{"~/.config/hypr", "~/.config/i3"}) // Check for either config
//	IsPresentAny([]string{"kitty"}, []string{"~/.config/kitty"})  // Check for binary OR config
func IsPresentAny(binaries []string, directories []string) bool {
	// Check binaries
	for _, binary := range binaries {
		if binary == "" {
			continue
		}

		// Check native PATH
		if _, err := exec.LookPath(binary); err == nil {
			return true
		}

		// Check Flatpak
		if isFlatpakInstalled(binary) {
			return true
		}

		// Check AppImage
		if isAppImagePresent(binary) {
			return true
		}
	}

	// Check directories
	for _, dir := range directories {
		if dir == "" {
			continue
		}
		if dirExists(expandPath(dir)) {
			return true
		}
	}

	return false
}

// IsPresentAll checks if ALL specified binaries AND directories exist.
// This is useful for checking if all required components are present.
//
// Parameters:
//   - binaries: List of binary names to check (can be nil or empty)
//   - directories: List of directories to check (can be nil or empty)
//
// Returns true only if ALL non-empty binaries AND directories are found.
// Empty string entries in slices are ignored.
//
// Examples:
//
//	IsPresentAll([]string{"kitty", "kitten"}, nil)                  // Require both binaries
//	IsPresentAll(nil, []string{"~/.config/hypr"})                   // Require config dir
//	IsPresentAll([]string{"ptyxis"}, []string{"~/.local/share/org.gnome.Ptyxis/"}) // Require both
func IsPresentAll(binaries []string, directories []string) bool {
	// Check all binaries
	for _, binary := range binaries {
		if binary == "" {
			continue
		}

		found := false

		// Check native PATH
		if _, err := exec.LookPath(binary); err == nil {
			found = true
		}

		// Check Flatpak
		if !found && isFlatpakInstalled(binary) {
			found = true
		}

		// Check AppImage
		if !found && isAppImagePresent(binary) {
			found = true
		}

		if !found {
			return false
		}
	}

	// Check all directories
	for _, dir := range directories {
		if dir == "" {
			continue
		}
		if !dirExists(expandPath(dir)) {
			return false
		}
	}

	return true
}

// isFlatpakInstalled checks if an application is installed via Flatpak.
// Common Flatpak app ID patterns:
//   - app.devsuite.<App>
//   - org.gnome.<App>
//   - com.github.<user>.<App>
func isFlatpakInstalled(appName string) bool {
	// Check if flatpak command exists
	if _, err := exec.LookPath("flatpak"); err != nil {
		return false
	}

	// Try common Flatpak app ID patterns
	patterns := []string{
		"app.devsuite." + strings.Title(appName),
		"org.gnome." + strings.Title(appName),
		"com.github.*." + strings.Title(appName),
		// Also try lowercase variants
		"app.devsuite." + appName,
		"org.gnome." + appName,
	}

	for _, pattern := range patterns {
		cmd := exec.Command("flatpak", "list", "--app", "--columns=application")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		// Check if any installed Flatpak matches the pattern
		apps := strings.SplitSeq(string(output), "\n")
		for app := range apps {
			app = strings.TrimSpace(app)
			if matchesPattern(app, pattern) {
				return true
			}
		}
	}

	return false
}

// isAppImagePresent checks for AppImage files in common locations.
func isAppImagePresent(appName string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Common AppImage locations
	locations := []string{
		filepath.Join(home, "Applications"),
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, "Downloads"),
		"/opt",
	}

	// Look for files matching pattern: <AppName>*.AppImage or <appname>*.AppImage
	patterns := []string{
		strings.Title(appName) + "*.AppImage",
		strings.ToLower(appName) + "*.AppImage",
		strings.ToUpper(appName) + "*.AppImage",
	}

	for _, location := range locations {
		if !dirExists(location) {
			continue
		}

		for _, pattern := range patterns {
			matches, err := filepath.Glob(filepath.Join(location, pattern))
			if err == nil && len(matches) > 0 {
				return true
			}
		}
	}

	return false
}

// dirExists checks if a directory exists.
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, path[1:])
}

// matchesPattern performs simple pattern matching (* wildcard support).
func matchesPattern(s, pattern string) bool {
	// If pattern contains *, use filepath.Match
	if strings.Contains(pattern, "*") {
		matched, _ := filepath.Match(pattern, s)
		return matched
	}

	// Exact match
	return s == pattern
}

// GetInstallationType returns a string indicating how the application is installed.
// Returns "native", "flatpak", "appimage", "directory", or "" if not found.
//
// Parameters:
//   - binaries: List of binary names to check
//   - directories: List of directories to check
//
// Examples:
//
//	GetInstallationType([]string{"ptyxis"}, nil)           // Check only for binary
//	GetInstallationType(nil, []string{"/path/to/config"})  // Check only for directory
func GetInstallationType(binaries []string, directories []string) string {
	// Check binaries first
	for _, binary := range binaries {
		if binary == "" {
			continue
		}

		if _, err := exec.LookPath(binary); err == nil {
			return "native"
		}

		if isFlatpakInstalled(binary) {
			return "flatpak"
		}

		if isAppImagePresent(binary) {
			return "appimage"
		}
	}

	// Check directories
	for _, dir := range directories {
		if dir == "" {
			continue
		}
		if dirExists(dir) {
			return "directory"
		}
	}

	return ""
}
