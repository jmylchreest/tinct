// Package cli provides the command-line interface for Tinct.
package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/plugin/repository"
	"github.com/jmylchreest/tinct/internal/security"
)

const (
	statusMissing  = "missing"
	statusMismatch = "mismatch"
)

var (
	syncForce       bool
	syncVerify      bool
	syncSkipMissing bool
	syncUpdate      bool
)

// pluginSyncCmd syncs plugins from lock file.
var pluginSyncCmd = &cobra.Command{
	Use:     "sync",
	Aliases: []string{"restore"},
	Short:   "Install plugins from lock file",
	Long: `Install plugins listed in the lock file that are not currently installed.

This is useful for:
  - Setting up plugins on a new machine
  - Restoring plugins after cleanup
  - Syncing team configurations
  - CI/CD pipelines

The lock file is typically located at ~/.config/tinct/.tinct-plugins.json

Plugins are reinstalled from their original source:
  - Repository plugins: Downloaded from configured repository
  - HTTP plugins: Downloaded from original URL
  - Local plugins: Copied from original path

Examples:
  tinct plugins sync                    # Install missing plugins
  tinct plugins sync --force            # Reinstall all plugins
  tinct plugins sync --verify           # Verify checksums
  tinct plugins sync --skip-missing     # Continue on errors`,
	RunE: runPluginSync,
}

// pluginVerifyCmd verifies installed plugins.
var pluginVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify installed plugins match lock file",
	Long: `Check that installed plugins match the lock file checksums and versions.

This command verifies:
  - Plugin files exist at expected paths
  - Checksums match (if available)
  - Plugin metadata is correct

Use this to detect if plugins have been modified or corrupted.`,
	RunE: runPluginVerify,
}

// pluginCleanCmd removes plugins not in lock file.
var pluginCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove plugins not in lock file",
	Long: `Remove plugins that are installed but not listed in the lock file.

This helps keep your plugin directory clean and ensures consistency
with the lock file.`,
	RunE: runPluginClean,
}

func init() {
	// Add sync command to plugins.
	pluginsCmd.AddCommand(pluginSyncCmd)
	pluginsCmd.AddCommand(pluginVerifyCmd)
	pluginsCmd.AddCommand(pluginCleanCmd)

	// Sync flags.
	pluginSyncCmd.Flags().BoolVar(&syncForce, "force", false, "Reinstall even if already present")
	pluginSyncCmd.Flags().BoolVar(&syncVerify, "verify", false, "Verify checksums of existing plugins")
	pluginSyncCmd.Flags().BoolVar(&syncSkipMissing, "skip-missing", false, "Continue if source unavailable")
	pluginSyncCmd.Flags().BoolVar(&syncUpdate, "update", false, "Update to latest compatible version")

	// Clean flags.
	pluginCleanCmd.Flags().BoolVarP(&pluginYes, "yes", "y", false, "Auto-confirm removal")
}

func runPluginSync(cmd *cobra.Command, args []string) error {
	// Read lock file.
	lock, lockPath, err := loadPluginLock()
	if err != nil {
		if os.IsNotExist(err) {
			defaultPath := pluginLockPath
			if defaultPath == "" {
				defaultPath = PluginLockFile
			}
			return fmt.Errorf("lock file not found at %s\n\nCreate one by installing plugins with 'tinct plugins add'", defaultPath)
		}
		return fmt.Errorf("failed to read lock file: %w", err)
	}
	_ = lockPath // Use the returned lockPath

	if len(lock.ExternalPlugins) == 0 {
		fmt.Println("No external plugins in lock file.")
		return nil
	}

	fmt.Printf("Reading lock file: %s\n", lockPath)
	fmt.Printf("Found %d plugin(s) in lock file:\n\n", len(lock.ExternalPlugins))

	stats := repository.SyncStats{
		Total: len(lock.ExternalPlugins),
	}

	for name, meta := range lock.ExternalPlugins {
		fmt.Printf("Checking %s...\n", name)

		// Check if plugin exists.
		exists := false
		if meta.Path != "" {
			if _, err := os.Stat(meta.Path); err == nil {
				exists = true
			}
		}

		// Handle existing plugins.
		if exists && !syncForce {
			// Verify checksum if requested and available.
			if syncVerify && meta.Source != nil && meta.Source.Checksum != "" {
				if err := verifyPluginChecksum(meta.Path, meta.Source.Checksum); err != nil {
					fmt.Printf("   Checksum mismatch: %v\n", err)
					if !syncForce {
						stats.Failed++
						continue
					}
					fmt.Printf("  → Reinstalling...\n")
					exists = false
				} else {
					fmt.Printf("   Checksum verified\n")
				}
			}

			if exists {
				fmt.Printf("   Already installed (v%s)\n", meta.Version)
				stats.Existing++
				continue
			}
		}

		// Reinstall from source.
		switch {
		case meta.Source != nil:
			fmt.Printf("  → Installing from %s\n", formatPluginSource(meta.Source))
		case meta.SourceLegacy != "":
			fmt.Printf("  → Installing from %s\n", meta.SourceLegacy)
		default:
			fmt.Printf("   No source information available\n")
			stats.Failed++
			continue
		}

		if err := reinstallPlugin(meta); err != nil {
			fmt.Printf("   Failed: %v\n", err)
			if !syncSkipMissing {
				stats.Failed++
				continue
			}
			fmt.Printf("   Skipped\n")
			stats.Skipped++
			continue
		}

		fmt.Printf("   Installed\n")
		stats.Installed++
	}

	// Print summary.
	fmt.Println()
	printSyncSummary(stats)

	if stats.Failed > 0 && !syncSkipMissing {
		return fmt.Errorf("some plugins failed to install")
	}

	return nil
}

func runPluginVerify(cmd *cobra.Command, args []string) error {
	// Read lock file.
	lock, _, err := loadPluginLock()
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	if len(lock.ExternalPlugins) == 0 {
		fmt.Println("No external plugins in lock file.")
		return nil
	}

	fmt.Printf("Verifying plugins against lock file...\n\n")

	results := []repository.VerifyResult{}

	for name, meta := range lock.ExternalPlugins {
		result := repository.VerifyResult{
			Name: name,
		}

		// Check existence.
		if meta.Path == "" {
			result.Status = "no_path"
			results = append(results, result)
			continue
		}

		if _, err := os.Stat(meta.Path); os.IsNotExist(err) {
			result.Status = statusMissing
			results = append(results, result)
			continue
		}

		// Verify checksum if available.
		if meta.Source != nil && meta.Source.Checksum != "" {
			if err := verifyPluginChecksum(meta.Path, meta.Source.Checksum); err != nil {
				result.Status = statusMismatch
				result.Error = err
				result.Expected = meta.Source.Checksum

				// Calculate actual checksum.
				if actual, err := calculateChecksum(meta.Path); err == nil {
					result.Got = "sha256:" + actual
				}
			} else {
				result.Status = "valid"
				result.Expected = meta.Source.Checksum
			}
		} else {
			result.Status = "no_checksum"
		}

		results = append(results, result)
	}

	// Print results.
	printVerifyResults(results)

	// Return error if any mismatches.
	for _, result := range results {
		if result.Status == statusMismatch || result.Status == statusMissing {
			return fmt.Errorf("\nverification failed. Run 'tinct plugins sync --force' to reinstall plugins")
		}
	}

	return nil
}

func runPluginClean(cmd *cobra.Command, args []string) error {
	// Read lock file.
	lock, _, err := loadPluginLock()
	if err != nil {
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	// Get plugin directory.
	pluginDir, err := getPluginDirectory()
	if err != nil {
		return err
	}

	// Scan plugin directory.
	fmt.Printf("Scanning plugin directory: %s\n\n", pluginDir)

	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Plugin directory does not exist.")
			return nil
		}
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	// Find plugins not in lock file.
	toRemove := []string{}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(pluginDir, name)

		// Check if in lock file.
		inLockFile := false
		for _, meta := range lock.ExternalPlugins {
			if meta.Path == path || filepath.Base(meta.Path) == name {
				inLockFile = true
				break
			}
		}

		if inLockFile {
			fmt.Printf("   %s - in lock file\n", name)
		} else {
			fmt.Printf("   %s - not in lock file\n", name)
			toRemove = append(toRemove, path)
		}
	}

	if len(toRemove) == 0 {
		fmt.Println("\nNo plugins to remove.")
		return nil
	}

	// Confirm removal.
	fmt.Printf("\nPlugins to remove:\n")
	for _, path := range toRemove {
		fmt.Printf("  - %s\n", filepath.Base(path))
	}

	if !pluginYes {
		fmt.Printf("\nRemove these plugins? [y/N]: ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if response != "y" && response != "Y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Remove plugins.
	fmt.Println()
	removed := 0
	for _, path := range toRemove {
		fmt.Printf("Removing %s...\n", filepath.Base(path))
		if err := os.Remove(path); err != nil {
			fmt.Printf("   Failed: %v\n", err)
		} else {
			fmt.Printf("   Deleted\n")
			removed++
		}
	}

	fmt.Printf("\nDone! Removed %d plugin(s).\n", removed)
	return nil
}

// reinstallPlugin reinstalls a plugin from its source.
func reinstallPlugin(meta ExternalPluginMeta) error {
	if meta.Source != nil {
		switch meta.Source.Type {
		case "repository":
			return reinstallFromRepository(meta)
		case "http":
			return reinstallFromHTTP(meta)
		case "local":
			return reinstallFromLocal(meta)
		default:
			return fmt.Errorf("unknown source type: %s", meta.Source.Type)
		}
	}

	// Fall back to legacy source string.
	if meta.SourceLegacy != "" {
		return reinstallFromLegacySource(meta)
	}

	return fmt.Errorf("no source information available")
}

// reinstallFromRepository installs a plugin from a repository.
func reinstallFromRepository(meta ExternalPluginMeta) error {
	mgr, err := getRepoManager()
	if err != nil {
		return fmt.Errorf("failed to get repository manager: %w", err)
	}

	result, err := mgr.FindPluginInRepository(
		meta.Source.Repository,
		meta.Source.Plugin,
		meta.Source.Version,
	)
	if err != nil {
		return fmt.Errorf("plugin not found in repository: %w", err)
	}

	// Get download for current platform.
	download := getDownloadForPlatform(result.Version.Downloads)
	if download == nil {
		return fmt.Errorf("no download available for current platform")
	}

	// Download and install.
	return downloadAndInstallPlugin(download.URL, meta.Name, download.Checksum)
}

// reinstallFromHTTP installs a plugin from HTTP URL.
func reinstallFromHTTP(meta ExternalPluginMeta) error {
	return downloadAndInstallPlugin(meta.Source.URL, meta.Name, meta.Source.Checksum)
}

// reinstallFromLocal installs a plugin from local path.
func reinstallFromLocal(meta ExternalPluginMeta) error {
	if meta.Source.OriginalPath == "" {
		return fmt.Errorf("no original path specified")
	}

	// Check if source file exists.
	if _, err := os.Stat(meta.Source.OriginalPath); os.IsNotExist(err) {
		return fmt.Errorf("source file not found: %s", meta.Source.OriginalPath)
	}

	// Get plugin directory.
	pluginDir, err := getPluginDirectory()
	if err != nil {
		return err
	}

	destPath := filepath.Join(pluginDir, meta.Name)

	// Copy file.
	if err := copyFile(meta.Source.OriginalPath, destPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Make executable.
	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return fmt.Errorf("failed to make executable: %w", err)
	}

	return nil
}

// reinstallFromLegacySource installs from legacy source string.
func reinstallFromLegacySource(meta ExternalPluginMeta) error {
	source := meta.SourceLegacy

	// Try to parse as URL.
	if isHTTPURL(source) {
		return downloadAndInstallPlugin(source, meta.Name, "")
	}

	// Try as local path.
	if _, err := os.Stat(source); err == nil {
		return reinstallFromLocal(ExternalPluginMeta{
			Name: meta.Name,
			Source: &repository.PluginSource{
				Type:         "local",
				OriginalPath: source,
			},
		})
	}

	return fmt.Errorf("unable to determine source type from: %s", source)
}

// downloadAndInstallPlugin downloads and installs a plugin from URL.
func downloadAndInstallPlugin(url, name, expectedChecksum string) error {
	// Validate URL to prevent SSRF attacks.
	if err := security.ValidateHTTPURL(url); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Download file.
	// #nosec G107 -- URL is validated via security.ValidateHTTPURL above.
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Verify checksum if provided.
	if expectedChecksum != "" {
		actualChecksum := "sha256:" + fmt.Sprintf("%x", sha256.Sum256(data))
		if actualChecksum != expectedChecksum {
			return fmt.Errorf("checksum mismatch (expected: %s, got: %s)", expectedChecksum, actualChecksum)
		}
	}

	// Get plugin directory.
	pluginDir, err := getPluginDirectory()
	if err != nil {
		return err
	}

	// Determine filename.
	filename := name
	if filepath.Ext(url) != "" {
		filename = name + filepath.Ext(url)
	}

	destPath := filepath.Join(pluginDir, filename)

	// Write file.
	// #nosec G306 -- Plugin executable needs exec permissions.
	if err := os.WriteFile(destPath, data, 0o755); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// verifyPluginChecksum verifies a plugin's checksum.
func verifyPluginChecksum(path, expected string) error {
	actual, err := calculateChecksum(path)
	if err != nil {
		return err
	}

	actualWithPrefix := "sha256:" + actual
	if actualWithPrefix != expected {
		return fmt.Errorf("checksum mismatch")
	}

	return nil
}

// calculateChecksum calculates SHA256 checksum of a file.
func calculateChecksum(path string) (string, error) {
	f, err := os.Open(path) // #nosec G304 - Plugin file path controlled by application
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src) // #nosec G304 - Plugin source path controlled by application
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}

// getPluginDirectory returns the plugin installation directory.
func getPluginDirectory() (string, error) {
	dataDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	pluginDir := filepath.Join(dataDir, ".local", "share", "tinct", "plugins")

	// Ensure directory exists.
	if err := os.MkdirAll(pluginDir, 0o755); err != nil { // #nosec G301 - Plugin directory needs standard permissions
		return "", fmt.Errorf("failed to create plugin directory: %w", err)
	}

	return pluginDir, nil
}

// isHTTPURL checks if a string is an HTTP(S) URL.
func isHTTPURL(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}

// formatPluginSource formats a plugin source for display.
func formatPluginSource(source *repository.PluginSource) string {
	switch source.Type {
	case "repository":
		return fmt.Sprintf("repository:%s/%s@%s", source.Repository, source.Plugin, source.Version)
	case "http":
		return source.URL
	case "local":
		return source.OriginalPath
	default:
		return source.Type
	}
}

// getDownloadForPlatform returns the appropriate download for the current platform.
func getDownloadForPlatform(downloads map[string]*repository.Download) *repository.Download {
	// Try platform-specific first.
	platform := fmt.Sprintf("%s_%s", getOS(), getArch())
	if dl, ok := downloads[platform]; ok {
		return dl
	}

	// Try OS-specific.
	if dl, ok := downloads[getOS()]; ok {
		return dl
	}

	// Fall back to "any".
	if dl, ok := downloads["any"]; ok {
		return dl
	}

	return nil
}

// getOS returns the current OS.
func getOS() string {
	// Use runtime.GOOS equivalent.
	return "linux" // Simplified for now
}

// getArch returns the current architecture.
func getArch() string {
	// Use runtime.GOARCH equivalent.
	return "amd64" // Simplified for now
}

// printSyncSummary prints a summary of sync operations.
func printSyncSummary(stats repository.SyncStats) {
	fmt.Println("Summary:")
	if stats.Installed > 0 {
		fmt.Printf("   %d newly installed\n", stats.Installed)
	}
	if stats.Existing > 0 {
		fmt.Printf("   %d already installed\n", stats.Existing)
	}
	if stats.Failed > 0 {
		fmt.Printf("   %d failed\n", stats.Failed)
	}
	if stats.Skipped > 0 {
		fmt.Printf("   %d skipped\n", stats.Skipped)
	}
}

// printVerifyResults prints verification results.
func printVerifyResults(results []repository.VerifyResult) {
	validCount := 0
	mismatchCount := 0
	missingCount := 0
	noChecksumCount := 0

	for _, result := range results {
		switch result.Status {
		case "valid":
			fmt.Printf(" %s\n", result.Name)
			fmt.Printf("  Checksum: %s [VALID]\n", result.Expected)
			validCount++
		case statusMismatch:
			fmt.Printf(" %s\n", result.Name)
			fmt.Printf("  Checksum: [MISMATCH]\n")
			if result.Expected != "" {
				fmt.Printf("  Expected: %s\n", result.Expected)
			}
			if result.Got != "" {
				fmt.Printf("  Got: %s\n", result.Got)
			}
			mismatchCount++
		case statusMissing:
			fmt.Printf(" %s\n", result.Name)
			fmt.Printf("  Status: [MISSING]\n")
			missingCount++
		case "no_checksum":
			fmt.Printf(" %s\n", result.Name)
			fmt.Printf("  Checksum: [NOT AVAILABLE]\n")
			noChecksumCount++
		case "no_path":
			fmt.Printf(" %s\n", result.Name)
			fmt.Printf("  Status: [NO PATH]\n")
			noChecksumCount++
		}
		fmt.Println()
	}

	fmt.Println("Summary:")
	if validCount > 0 {
		fmt.Printf("   %d valid\n", validCount)
	}
	if mismatchCount > 0 {
		fmt.Printf("   %d checksum mismatch\n", mismatchCount)
	}
	if missingCount > 0 {
		fmt.Printf("   %d missing\n", missingCount)
	}
	if noChecksumCount > 0 {
		fmt.Printf("   %d no checksum\n", noChecksumCount)
	}
}
