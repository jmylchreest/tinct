// Package cli provides the command-line interface for Tinct.
package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/security"
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
func installPluginFromSource(source, pluginName, pluginDir string, verbose bool) (string, error) {
	// Parse source to determine type.
	sourceType, sourceInfo := parsePluginSource(source)

	switch sourceType {
	case sourceTypeLocal:
		return installFromLocal(sourceInfo, pluginDir, verbose)
	case sourceTypeHTTP:
		return installFromHTTP(sourceInfo, pluginName, pluginDir, verbose)
	case sourceTypeGit:
		return installFromGit(sourceInfo, pluginName, pluginDir, verbose)
	default:
		return "", fmt.Errorf("unsupported source type: %s", source)
	}
}

// parsePluginSource determines the source type and extracts relevant info.
func parsePluginSource(source string) (string, PluginSourceInfo) {
	info := PluginSourceInfo{}

	// Git repository (https://github.com/user/repo.git or git@github.com:user/repo.git).
	isGit := strings.HasSuffix(source, ".git") ||
		strings.Contains(source, "github.com") ||
		strings.Contains(source, "gitlab.com") ||
		strings.Contains(source, "bitbucket.org")

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

	// HTTP/HTTPS URL.
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Check for file path specification: url.tar.gz:path/to/plugin.
		idx := strings.LastIndex(source, ":")
		if idx <= 0 {
			info.URL = source
			return sourceTypeHTTP, info
		}

		// Check if it's part of the protocol (http:// or https://).
		if idx > 7 && source[idx-2:idx] != "tp" && source[idx-3:idx] != "tps" {
			info.URL = source[:idx]
			info.FilePath = source[idx+1:]
		} else {
			info.URL = source
		}
		return sourceTypeHTTP, info
	}

	// Local file.
	info.FilePath = source
	return sourceTypeLocal, info
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
func installFromHTTP(info PluginSourceInfo, pluginName, pluginDir string, verbose bool) (string, error) {
	if verbose {
		fmt.Fprintf(os.Stderr, "Downloading from %s...\n", info.URL)
	}

	// Download the file.
	resp, err := http.Get(info.URL)
	if err != nil {
		return "", fmt.Errorf("failed to download plugin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download plugin: HTTP %d", resp.StatusCode)
	}

	// Read the entire response into memory for archive detection.
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read download: %w", err)
	}

	// Determine filename from URL.
	filename := filepath.Base(info.URL)
	if filename == "" || filename == "." {
		filename = pluginName
	}

	// Check if it's an archive.
	if strings.HasSuffix(info.URL, ".tar.gz") || strings.HasSuffix(info.URL, ".tgz") {
		// Extract from tar.gz archive.
		return extractFromTarGz(data, info.FilePath, pluginDir, verbose)
	} else if strings.HasSuffix(info.URL, ".zip") {
		// Extract from zip archive.
		return extractFromZip(data, info.FilePath, pluginDir, verbose)
	}

	// Not an archive - treat as direct plugin file.
	destPath := filepath.Join(pluginDir, filename)

	// Write file.
	// #nosec G306 -- Plugin executable needs exec permissions.
	if err := os.WriteFile(destPath, data, 0o755); err != nil {
		return "", fmt.Errorf("failed to write plugin file: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Downloaded plugin to: %s\n", destPath)
	}

	return destPath, nil
}

// extractFromTarGz extracts a plugin from a tar.gz archive.
func extractFromTarGz(data []byte, targetFile, pluginDir string, verbose bool) (string, error) {
	// Create gzip reader.
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader.
	tr := tar.NewReader(gzr)

	// If no specific file requested, find the first executable or use first file.
	var targetPath string
	foundFiles := []string{}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar archive: %w", err)
		}

		// Skip directories.
		if header.Typeflag == tar.TypeDir {
			continue
		}

		foundFiles = append(foundFiles, header.Name)

		// Check if this is the file we want.
		if targetFile != "" {
			if header.Name == targetFile || strings.HasSuffix(header.Name, "/"+targetFile) {
				targetPath = header.Name
				break
			}
		} else {
			// Auto-detect: prefer executable files.
			if header.FileInfo().Mode()&0o111 != 0 {
				targetPath = header.Name
				break
			}
		}
	}

	// If we didn't find the target, reset and look for any match.
	if targetPath == "" && targetFile != "" {
		return "", fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	}

	// If still no target and we have files, use the first one.
	if targetPath == "" && len(foundFiles) > 0 {
		targetPath = foundFiles[0]
	}

	if targetPath == "" {
		return "", fmt.Errorf("no files found in archive")
	}

	// Reset readers to extract the target file.
	gzr, err = gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr = tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return "", fmt.Errorf("file not found in archive")
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Name != targetPath {
			continue
		}

		// Extract the file.
		destPath := filepath.Join(pluginDir, filepath.Base(targetPath))

		out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
		if err != nil {
			return "", fmt.Errorf("failed to create plugin file: %w", err)
		}

		// Limit decompression size to prevent zip bombs (100MB limit for plugins).
		limitedReader := security.NewLimitedReader(tr, 100*1024*1024)
		_, copyErr := io.Copy(out, limitedReader)
		closeErr := out.Close() // Close immediately instead of defer

		if copyErr != nil {
			return "", fmt.Errorf("failed to extract plugin: %w", copyErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("failed to close plugin file: %w", closeErr)
		}

		// Make executable.
		if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
			return "", fmt.Errorf("failed to make plugin executable: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Extracted plugin to: %s\n", destPath)
		}

		return destPath, nil
	}
}

// extractFromZip extracts a plugin from a zip archive.
func extractFromZip(data []byte, targetFile, pluginDir string, verbose bool) (string, error) {
	// Create zip reader.
	reader := bytes.NewReader(data)
	zr, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to create zip reader: %w", err)
	}

	// If no specific file requested, find the first executable or use first file.
	var targetZipFile *zip.File
	foundFiles := []string{}

	for _, f := range zr.File {
		// Skip directories.
		if f.FileInfo().IsDir() {
			continue
		}

		foundFiles = append(foundFiles, f.Name)

		// Check if this is the file we want.
		if targetFile != "" {
			if f.Name == targetFile || strings.HasSuffix(f.Name, "/"+targetFile) {
				targetZipFile = f
				break
			}
		} else {
			// Auto-detect: prefer executable files.
			if f.FileInfo().Mode()&0o111 != 0 {
				targetZipFile = f
				break
			}
		}
	}

	// If we didn't find the target, check if any file matches.
	if targetZipFile == nil && targetFile != "" {
		return "", fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	}

	// If still no target and we have files, use the first one.
	if targetZipFile == nil && len(foundFiles) > 0 {
		targetZipFile = zr.File[0]
	}

	if targetZipFile == nil {
		return "", fmt.Errorf("no files found in archive")
	}

	// Extract the file.
	destPath := filepath.Join(pluginDir, filepath.Base(targetZipFile.Name))

	rc, err := targetZipFile.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file in archive: %w", err)
	}
	defer rc.Close()

	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return "", fmt.Errorf("failed to create plugin file: %w", err)
	}
	defer out.Close()

	// Limit decompression size to prevent zip bombs (100MB limit for plugins).
	limitedReader := security.NewLimitedReader(rc, 100*1024*1024)
	if _, err := io.Copy(out, limitedReader); err != nil {
		return "", fmt.Errorf("failed to extract plugin: %w", err)
	}

	// Make executable.
	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Extracted plugin to: %s\n", destPath)
	}

	return destPath, nil
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
		// Specific file path provided.
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
