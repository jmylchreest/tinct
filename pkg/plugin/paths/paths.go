// Package paths provides cross-platform user-config and user-data path
// helpers for tinct plugins. Both internal plugins and external Go plugins
// should use these helpers instead of computing platform-specific paths
// themselves; this keeps `DefaultOutputDir` correct on Linux, macOS, and
// Windows without per-plugin per-OS code.
package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// XDGConfigDir returns the directory under which XDG-style config files
// live for the current user.
//
// Resolution order:
//
//   - $XDG_CONFIG_HOME if set (any platform).
//   - %AppData% on Windows (Roaming).
//   - $HOME/.config everywhere else.
//
// Linux follows the XDG spec. macOS uses ~/.config when $XDG_CONFIG_HOME
// is unset because most modern CLI tools (alacritty, kitty, neovim,
// ghostty) treat that as the canonical config root on macOS too.
func XDGConfigDir() string {
	if v := os.Getenv("XDG_CONFIG_HOME"); v != "" {
		return v
	}
	if runtime.GOOS == "windows" {
		if v := os.Getenv("AppData"); v != "" {
			return v
		}
	}
	home, _ := os.UserHomeDir() //nolint:errcheck // empty home is acceptable; caller gets a relative path
	return filepath.Join(home, ".config")
}

// XDGDataDir returns the directory under which XDG-style data files live.
//
// Resolution order:
//
//   - $XDG_DATA_HOME if set (any platform).
//   - %LocalAppData% on Windows.
//   - $HOME/.local/share everywhere else.
func XDGDataDir() string {
	if v := os.Getenv("XDG_DATA_HOME"); v != "" {
		return v
	}
	if runtime.GOOS == "windows" {
		if v := os.Getenv("LocalAppData"); v != "" {
			return v
		}
	}
	home, _ := os.UserHomeDir() //nolint:errcheck // empty home is acceptable; caller gets a relative path
	return filepath.Join(home, ".local", "share")
}

// MacOSAppSupport returns ~/Library/Application Support/<bundle> for plugins
// targeting macOS-bundled apps (e.g. Ghostty, Zellij).
//
// The caller is responsible for only invoking this on darwin (check
// runtime.GOOS first); on other platforms the path is meaningless.
func MacOSAppSupport(bundle string) string {
	home, _ := os.UserHomeDir() //nolint:errcheck // empty home is acceptable; caller gets a relative path
	return filepath.Join(home, "Library", "Application Support", bundle)
}
