// Package hyprpaper provides an output plugin for Hyprpaper wallpaper manager configuration.
// This file contains version-specific logic for hyprpaper v0.8.0.
package hyprpaper

import (
	"fmt"
	"os"
)

// setupConfigSymlink creates a symlink from hyprpaper.conf to tinct-hyprpaper.conf.
// This is needed for hyprpaper v0.8.0 which doesn't support source= includes.
// If hyprpaper.conf exists and is not already our symlink, it's backed up first.
func (p *Plugin) setupConfigSymlink(writtenFiles []string) error {
	paths := resolveConfigPaths(writtenFiles)
	if paths == nil {
		return nil // No config file written, nothing to symlink.
	}

	// Check current state of hyprpaper.conf.
	linkTarget, err := os.Readlink(paths.hyprpaperConf)
	switch {
	case err == nil:
		// It's a symlink - check if it already points to our file.
		if linkTarget == tinctConfigFilename || linkTarget == paths.tinctConfig {
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Config symlink already exists: %s -> %s\n", paths.hyprpaperConf, linkTarget)
			}
			return nil
		}
		// Symlink points elsewhere - remove it so we can recreate.
		if err := os.Remove(paths.hyprpaperConf); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	case os.IsNotExist(err):
		// Doesn't exist - we can create the symlink.
	default:
		// It's a regular file (or other error) - check if it exists.
		if _, statErr := os.Stat(paths.hyprpaperConf); statErr == nil {
			// Regular file exists - back it up.
			if p.verbose {
				fmt.Fprintf(os.Stderr, "   Backing up existing config: %s -> %s\n", paths.hyprpaperConf, paths.backup)
			}
			if err := os.Rename(paths.hyprpaperConf, paths.backup); err != nil {
				return fmt.Errorf("failed to backup existing config: %w", err)
			}
		}
	}

	// Create symlink using relative path.
	if err := os.Symlink(tinctConfigFilename, paths.hyprpaperConf); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Created config symlink: %s -> %s\n", paths.hyprpaperConf, tinctConfigFilename)
	}

	return nil
}
