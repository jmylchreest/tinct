// Package hyprpaper provides an output plugin for Hyprpaper wallpaper manager configuration.
// This file contains version-specific logic for hyprpaper v0.8.1+.
// Version 0.8.1 reintroduced support for source= includes.
package hyprpaper

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// restoreSourceConfig checks if we can restore the original hyprpaper.conf that uses source= includes.
// This is used for hyprpaper v0.8.1+ which reintroduced source= support.
// Conditions for restoration:
//   - hyprpaper.conf is a symlink to tinct-hyprpaper.conf
//   - hyprpaper.conf.bak exists
//   - hyprpaper.conf.bak contains a source= line that includes tinct-hyprpaper.conf
//
// If all conditions are met, removes the symlink and restores the backup.
func (p *Plugin) restoreSourceConfig(writtenFiles []string) error {
	paths := resolveConfigPaths(writtenFiles)
	if paths == nil {
		return nil // No config file written, nothing to restore.
	}

	// Check all preconditions for restoration.
	if !p.shouldRestoreConfig(paths) {
		return nil
	}

	// All conditions met - restore the backup.
	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Restoring original config with source= includes (hyprpaper 0.8.1+ detected)\n")
	}

	// Remove the symlink.
	if err := os.Remove(paths.hyprpaperConf); err != nil {
		return fmt.Errorf("failed to remove symlink: %w", err)
	}

	// Copy backup back to hyprpaper.conf (use copy to preserve backup).
	backupContent, err := os.ReadFile(paths.backup) // #nosec G304 -- paths.backup is constructed from configDir
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	if err := os.WriteFile(paths.hyprpaperConf, backupContent, 0o600); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	if p.verbose {
		fmt.Fprintf(os.Stderr, "   Restored config: %s (from backup)\n", paths.hyprpaperConf)
	}

	return nil
}

// shouldRestoreConfig checks if all conditions are met to restore the backup config.
func (p *Plugin) shouldRestoreConfig(paths *configPaths) bool {
	// Check if hyprpaper.conf is a symlink to our config.
	linkTarget, err := os.Readlink(paths.hyprpaperConf)
	if err != nil {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   hyprpaper.conf is not a symlink, no restoration needed\n")
		}
		return false
	}

	// Check if the symlink points to our tinct config.
	if linkTarget != tinctConfigFilename && linkTarget != paths.tinctConfig {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   hyprpaper.conf symlink points to %s, not tinct config, skipping restoration\n", linkTarget)
		}
		return false
	}

	// Check if backup exists.
	if _, err := os.Stat(paths.backup); os.IsNotExist(err) {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   No backup file found at %s, skipping restoration\n", paths.backup)
		}
		return false
	}

	// Check if backup contains source= line for tinct-hyprpaper.conf.
	if !backupContainsTinctSource(paths.backup) {
		if p.verbose {
			fmt.Fprintf(os.Stderr, "   Backup does not contain source= for tinct-hyprpaper.conf, skipping restoration\n")
		}
		return false
	}

	return true
}

// backupContainsTinctSource checks if the backup file contains a source= line
// that includes tinct-hyprpaper.conf.
func backupContainsTinctSource(backupPath string) bool {
	file, err := os.Open(backupPath) // #nosec G304 -- backupPath is constructed from configDir
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments.
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Look for source= or source = patterns that include tinct-hyprpaper.conf.
		if rest, found := strings.CutPrefix(line, "source"); found {
			rest = strings.TrimSpace(rest)
			if rest, found = strings.CutPrefix(rest, "="); found {
				rest = strings.TrimSpace(rest)
				if strings.Contains(rest, tinctConfigFilename) {
					return true
				}
			}
		}
	}

	return false
}
