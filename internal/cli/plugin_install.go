// Package cli provides the command-line interface for Tinct.
package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmylchreest/tinct/internal/compression"
	"github.com/jmylchreest/tinct/internal/security"
	"github.com/jmylchreest/tinct/internal/ui/progress"
	httputil "github.com/jmylchreest/tinct/internal/util/http"
)

const (
	// Plugin source types.
	sourceTypeRepository = "repository"
	sourceTypeHTTP       = "http"
	sourceTypeLocal      = "local"
	sourceTypeGit        = "git"
)

// PluginSourceInfo represents the type of plugin source.
type PluginSourceInfo struct {
	URL      string
	FilePath string // For git repos, path to file within repo
	Ref      string // For git repos, branch/tag/commit
}

// installPluginFromSource installs a plugin from various source types.
// If forcedSourceType is non-empty, it overrides auto-detection.
// If quiet is true, progress bar messages are suppressed (useful when a table displays status).
func installPluginFromSource(source, pluginName, pluginDir, forcedSourceType string, verbose, quiet bool) (string, error) {
	// Parse source to determine type.
	sourceType, sourceInfo := parsePluginSource(source)

	// Override with forced source type if provided
	if forcedSourceType != "" {
		if forcedSourceType != sourceTypeLocal && forcedSourceType != sourceTypeHTTP && forcedSourceType != sourceTypeGit {
			return "", fmt.Errorf("invalid source type '%s': must be one of: local, http, git", forcedSourceType)
		}
		sourceType = forcedSourceType
		// Update sourceInfo for forced type
		if forcedSourceType == sourceTypeLocal {
			sourceInfo.FilePath = source
		} else {
			sourceInfo.URL = source
		}
	}

	switch sourceType {
	case sourceTypeLocal:
		return installFromLocal(sourceInfo, pluginDir, verbose)
	case sourceTypeHTTP:
		return installFromHTTP(sourceInfo, pluginName, pluginDir, verbose, quiet)
	case sourceTypeGit:
		return installFromGit(sourceInfo, pluginName, pluginDir, verbose)
	default:
		return "", fmt.Errorf("unsupported source type: %s", source)
	}
}

// isDownloadableFile performs a HEAD request to check if the URL points to a downloadable file.
func isDownloadableFile(url string) bool {
	// Perform HEAD request to check Content-Type
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Head(url)
	if err != nil {
		// If HEAD fails, assume it might be downloadable (don't block on network errors)
		return true
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	// Check for common archive/binary content types
	downloadableTypes := []string{
		"application/gzip",
		"application/x-gzip",
		"application/x-tar",
		"application/x-compressed-tar",
		"application/octet-stream",
		"application/zip",
		"application/x-zip-compressed",
		"binary/octet-stream",
	}

	for _, dt := range downloadableTypes {
		if strings.Contains(contentType, dt) {
			return true
		}
	}

	// Check for HTML content type (likely a git web interface)
	if strings.Contains(contentType, "text/html") {
		return false
	}

	// If Content-Type is empty or unknown, check status code
	// 200 OK with no content type might be a file
	if contentType == "" && resp.StatusCode == 200 {
		return true
	}

	// Default to true to avoid false negatives
	return true
}

// parsePluginSource determines the source type and extracts relevant info.
func parsePluginSource(source string) (string, PluginSourceInfo) {
	info := PluginSourceInfo{}

	// HTTP/HTTPS URL takes precedence (e.g., GitHub releases)
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Check for file path specification: url.tar.gz:path/to/plugin.
		idx := strings.LastIndex(source, ":")
		if idx <= 0 {
			info.URL = source
		} else if idx > 7 && source[idx-2:idx] != "tp" && source[idx-3:idx] != "tps" {
			// Check if it's part of the protocol (http:// or https://).
			info.URL = source[:idx]
			info.FilePath = source[idx+1:]
		} else {
			info.URL = source
		}

		// Verify it's actually a downloadable file by checking Content-Type
		if isDownloadableFile(info.URL) {
			return sourceTypeHTTP, info
		}

		// If Content-Type suggests it's not a file, might be a git repo
		if strings.HasSuffix(info.URL, ".git") {
			return sourceTypeGit, info
		}

		// Default to HTTP anyway
		return sourceTypeHTTP, info
	}

	// Git repository (https://github.com/user/repo.git or git@github.com:user/repo.git).
	isGit := strings.HasSuffix(source, ".git")

	if isGit {
		// Check for file path specification: repo.git:path/to/file.sh.
		idx := strings.LastIndex(source, ":")
		if idx <= 0 || strings.HasPrefix(source, "git@") {
			info.URL = source
			return sourceTypeGit, info
		}

		// Make sure it's not the : in git@github.com.
		if idx > 6 && source[idx-1] != 'm' { // Not ending in ".com:"
			info.URL = source[:idx]
			info.FilePath = source[idx+1:]
		} else {
			info.URL = source
		}
		return sourceTypeGit, info
	}

	// Local file.
	info.FilePath = source
	return sourceTypeLocal, info
}

// copyFile copies a file from src to dst.
// maxPluginFileSize is the maximum size for a plugin file (100 MB).
const maxPluginFileSize int64 = 100 * 1024 * 1024

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

	// Limit copy size to prevent unbounded disk/memory usage.
	limitedReader := &io.LimitedReader{R: sourceFile, N: maxPluginFileSize + 1}
	n, err := io.Copy(destFile, limitedReader)
	if err != nil {
		return err
	}
	if n > maxPluginFileSize {
		return fmt.Errorf("plugin file exceeds maximum size (%d bytes)", maxPluginFileSize)
	}

	return destFile.Sync()
}

// installFromLocal installs a plugin from a local file.
func installFromLocal(info PluginSourceInfo, pluginDir string, verbose bool) (string, error) {
	absSource, err := filepath.Abs(info.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve plugin path: %w", err)
	}

	// Verify plugin exists.
	if _, err := os.Stat(absSource); err != nil {
		return "", fmt.Errorf("plugin file not found: %w", err)
	}

	// Copy plugin to plugin directory.
	destPath := filepath.Join(pluginDir, filepath.Base(absSource))
	if err := copyFile(absSource, destPath); err != nil {
		return "", fmt.Errorf("failed to copy plugin: %w", err)
	}

	// Make it executable.
	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Copied plugin to: %s\n", destPath)
	}

	return destPath, nil
}

// installFromHTTP downloads a plugin from an HTTP/HTTPS URL.
// If quiet is true, progress bar messages are suppressed.
func installFromHTTP(info PluginSourceInfo, pluginName, pluginDir string, verbose, quiet bool) (string, error) {
	// Validate URL before fetching (HTTPS-only, no localhost/private IPs).
	if err := security.ValidateHTTPURL(info.URL); err != nil {
		return "", fmt.Errorf("unsafe download URL: %w", err)
	}

	// Download the file using our HTTP utility with progress tracking
	ctx := context.Background()

	var data []byte
	var downloadErr error

	// Show progress bar for downloads unless in quiet mode
	var bar *progress.ProgressBar

	data, downloadErr = httputil.Fetch(ctx, info.URL, httputil.FetchOptions{
		ProgressCallback: func(current, total int64) {
			if quiet {
				return // Skip progress bar in quiet mode
			}
			if bar == nil && total > 0 {
				bar = progress.NewProgressBar(total, fmt.Sprintf("Downloading %s", pluginName))
				bar.Set(current)
			} else if bar != nil {
				bar.Set(current)
			}
		},
	})

	if bar != nil {
		// In quiet mode, don't print final message (table will show status)
		if quiet {
			bar.Finish("")
		} else if downloadErr == nil {
			bar.Finish(fmt.Sprintf("Downloaded %s", pluginName))
		} else {
			bar.Finish("Download failed")
		}
	}

	if downloadErr != nil {
		return "", fmt.Errorf("failed to download plugin: %w", downloadErr)
	}

	// Determine filename from URL.
	filename := filepath.Base(info.URL)
	if filename == "" || filename == "." {
		filename = pluginName
	}

	// Determine archive name (for finding plugin binary)
	archiveName := compression.GetArchiveBaseName(filename)

	// Content-Type is not available after fetch completes, but compression
	// utilities can detect format from content and filename
	contentType := ""

	// Use shared compression utilities for extraction
	result, err := compression.ExtractPlugin(
		data,
		info.URL,      // url for format detection
		filename,      // filename for fallback detection
		info.FilePath, // targetFile (specific file to extract from archive)
		archiveName,   // archiveName for matching plugin binary
		pluginDir,     // destination directory
		contentType,   // HTTP Content-Type for detection
		verbose,       // verbose output
	)
	if err != nil {
		return "", err
	}

	if verbose {
		if result.WasArchive {
			fmt.Fprintf(os.Stderr, "Extracted plugin to: %s\n", result.Path)
		} else {
			fmt.Fprintf(os.Stderr, "Downloaded plugin to: %s\n", result.Path)
		}
	}

	return result.Path, nil
}

// installFromGit clones a git repository and extracts the plugin.
func installFromGit(info PluginSourceInfo, pluginName, pluginDir string, verbose bool) (string, error) {
	// Check if git is available.
	if _, err := exec.LookPath("git"); err != nil {
		return "", fmt.Errorf("git is not installed or not in PATH")
	}

	// Validate the git URL for security.
	if err := security.ValidateGitURL(info.URL); err != nil {
		return "", fmt.Errorf("invalid git URL: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Cloning from %s...\n", info.URL)
	}

	// Create temporary directory for cloning.
	tmpDir, err := os.MkdirTemp("", "tinct-plugin-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repository using -- to separate options from arguments (prevents command injection).
	// #nosec G204 -- URL is validated via security.ValidateGitURL above.
	cloneCmd := exec.Command("git", "clone", "--depth", "1", "--", info.URL, tmpDir)
	if output, err := cloneCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, string(output))
	}

	// Determine which file to copy.
	var sourceFile string
	if info.FilePath != "" {
		// Specific file path provided — validate against directory traversal.
		if err := security.ValidateFilePath(info.FilePath, tmpDir); err != nil {
			return "", fmt.Errorf("unsafe plugin file path: %w", err)
		}
		sourceFile = filepath.Join(tmpDir, info.FilePath)
		if _, err := os.Stat(sourceFile); err != nil {
			return "", fmt.Errorf("plugin file not found in repository: %s", info.FilePath)
		}
	} else {
		// Try to find a plugin file automatically.
		// Look for common patterns: plugin.sh, *.py, *.sh in root or bin/.
		candidates := []string{
			filepath.Join(tmpDir, pluginName),
			filepath.Join(tmpDir, pluginName+".sh"),
			filepath.Join(tmpDir, pluginName+".py"),
			filepath.Join(tmpDir, "plugin.sh"),
			filepath.Join(tmpDir, "plugin.py"),
			filepath.Join(tmpDir, "bin", pluginName),
			filepath.Join(tmpDir, "bin", pluginName+".sh"),
			filepath.Join(tmpDir, "bin", pluginName+".py"),
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				sourceFile = candidate
				break
			}
		}

		if sourceFile == "" {
			return "", fmt.Errorf("could not find plugin file in repository, please specify path: repo.git:path/to/plugin.sh")
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Found plugin file: %s\n", filepath.Base(sourceFile))
	}

	// Copy plugin to plugin directory.
	destPath := filepath.Join(pluginDir, filepath.Base(sourceFile))
	if err := copyFile(sourceFile, destPath); err != nil {
		return "", fmt.Errorf("failed to copy plugin: %w", err)
	}

	// Make it executable.
	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Installed plugin to: %s\n", destPath)
	}

	return destPath, nil
}
