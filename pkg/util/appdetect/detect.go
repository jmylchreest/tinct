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

// IsPresent checks if an application is installed using multiple detection methods.
//
// Parameters:
//   - binaryName: Name of the binary to look for on PATH (e.g., "ptyxis")
//   - configDir: Configuration or data directory that indicates app is installed
//     (e.g., "~/.local/share/org.gnome.Ptyxis/")
//
// Detection methods (in order):
//  1. Native binary on PATH (if binaryName provided)
//  2. Flatpak installation (if binaryName provided)
//  3. AppImage in common locations (if binaryName provided)
//  4. Configuration/data directory exists (if configDir provided)
//
// Returns true if any detection method succeeds.
//
// Examples:
//
//	IsPresent("ptyxis", "")                                    // Check for ptyxis binary only
//	IsPresent("", "~/.local/share/org.gnome.Ptyxis/")         // Check for config dir only
//	IsPresent("ptyxis", "~/.local/share/org.gnome.Ptyxis/")   // Check both
func IsPresent(binaryName, configDir string) bool {
	// Check native PATH
	if binaryName != "" {
		if _, err := exec.LookPath(binaryName); err == nil {
			return true
		}

		// Check Flatpak
		if isFlatpakInstalled(binaryName) {
			return true
		}

		// Check AppImage in common locations
		if isAppImagePresent(binaryName) {
			return true
		}
	}

	// Check config/data directory
	if configDir != "" {
		if dirExists(expandPath(configDir)) {
			return true
		}
	}

	return false
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
		apps := strings.Split(string(output), "\n")
		for _, app := range apps {
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
// Returns "native", "flatpak", "appimage", "config-dir", or "" if not found.
func GetInstallationType(binaryName, configDir string) string {
	if binaryName != "" {
		if _, err := exec.LookPath(binaryName); err == nil {
			return "native"
		}

		if isFlatpakInstalled(binaryName) {
			return "flatpak"
		}

		if isAppImagePresent(binaryName) {
			return "appimage"
		}
	}

	if configDir != "" {
		if dirExists(expandPath(configDir)) {
			return "config-dir"
		}
	}

	return ""
}
