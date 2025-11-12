// Package cli provides the command-line interface for Tinct.
package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmylchreest/tinct/internal/security"
	"github.com/ulikunitz/xz"
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
func installPluginFromSource(source, pluginName, pluginDir, forcedSourceType string, verbose bool) (string, error) {
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
		return installFromHTTP(sourceInfo, pluginName, pluginDir, verbose)
	case sourceTypeGit:
		return installFromGit(sourceInfo, pluginName, pluginDir, verbose)
	default:
		return "", fmt.Errorf("unsupported source type: %s", source)
	}
}

// getArchiveBaseName extracts the base name from an archive filename.
// For example: "tinct-plugin-wob_0.0.1_Linux_x86_64.tar.gz" -> "tinct-plugin-wob".
func getArchiveBaseName(filename string) string {
	// Remove extension
	base := filename
	for _, ext := range []string{".tar.gz", ".tgz", ".tar.xz", ".txz", ".tar.bz2", ".tbz", ".tbz2", ".zip"} {
		if before, ok := strings.CutSuffix(base, ext); ok {
			base = before
			break
		}
	}

	// Find the part before the first underscore
	if idx := strings.Index(base, "_"); idx > 0 {
		return base[:idx]
	}

	return base
}

// isDownloadableFile performs a HEAD request to check if the URL points to a downloadable file.
func isDownloadableFile(url string) bool {
	// Perform HEAD request to check Content-Type
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
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

	// Determine archive name (for finding plugin binary)
	archiveName := getArchiveBaseName(filename)

	// Get Content-Type from response header
	contentType := resp.Header.Get("Content-Type")

	// Detect format by Content-Type first, then fall back to filename extension
	switch {
	// Tar+gzip archives
	case strings.Contains(contentType, "application/gzip") || strings.Contains(contentType, "application/x-gzip"):
		// Could be tar.gz or plain .gz - check filename
		if strings.HasSuffix(info.URL, ".tar.gz") || strings.HasSuffix(info.URL, ".tgz") {
			return extractFromTarGz(data, info.FilePath, archiveName, pluginDir, verbose)
		}
		// Plain gzip file
		return decompressGz(data, strings.TrimSuffix(filename, ".gz"), pluginDir, verbose)

	// Tar+xz archives
	case strings.Contains(contentType, "application/x-xz"):
		// Could be tar.xz or plain .xz - check filename
		if strings.HasSuffix(info.URL, ".tar.xz") || strings.HasSuffix(info.URL, ".txz") {
			return extractFromTarXz(data, info.FilePath, archiveName, pluginDir, verbose)
		}
		// Plain xz file
		return decompressXz(data, strings.TrimSuffix(filename, ".xz"), pluginDir, verbose)

	// Tar+bzip2 archives
	case strings.Contains(contentType, "application/x-bzip2"):
		// Could be tar.bz2 or plain .bz2 - check filename
		if strings.HasSuffix(info.URL, ".tar.bz2") || strings.HasSuffix(info.URL, ".tbz") || strings.HasSuffix(info.URL, ".tbz2") {
			return extractFromTarBz2(data, info.FilePath, archiveName, pluginDir, verbose)
		}
		// Plain bzip2 file
		return decompressBz2(data, strings.TrimSuffix(filename, ".bz2"), pluginDir, verbose)

	// Zip archives
	case strings.Contains(contentType, "application/zip") || strings.Contains(contentType, "application/x-zip-compressed"):
		return extractFromZip(data, info.FilePath, archiveName, pluginDir, verbose)

	// Tar archives (uncompressed)
	case strings.Contains(contentType, "application/x-tar"):
		// Assume it's compressed based on filename
		if strings.HasSuffix(info.URL, ".tar.gz") || strings.HasSuffix(info.URL, ".tgz") {
			return extractFromTarGz(data, info.FilePath, archiveName, pluginDir, verbose)
		} else if strings.HasSuffix(info.URL, ".tar.xz") || strings.HasSuffix(info.URL, ".txz") {
			return extractFromTarXz(data, info.FilePath, archiveName, pluginDir, verbose)
		} else if strings.HasSuffix(info.URL, ".tar.bz2") || strings.HasSuffix(info.URL, ".tbz") || strings.HasSuffix(info.URL, ".tbz2") {
			return extractFromTarBz2(data, info.FilePath, archiveName, pluginDir, verbose)
		}
		return "", fmt.Errorf("uncompressed tar archives are not supported")

	// Generic binary/octet-stream - fall back to filename detection
	case strings.Contains(contentType, "application/octet-stream") || contentType == "":
		// Fall back to filename-based detection
		if strings.HasSuffix(info.URL, ".tar.gz") || strings.HasSuffix(info.URL, ".tgz") {
			return extractFromTarGz(data, info.FilePath, archiveName, pluginDir, verbose)
		} else if strings.HasSuffix(info.URL, ".tar.xz") || strings.HasSuffix(info.URL, ".txz") {
			return extractFromTarXz(data, info.FilePath, archiveName, pluginDir, verbose)
		} else if strings.HasSuffix(info.URL, ".tar.bz2") || strings.HasSuffix(info.URL, ".tbz") || strings.HasSuffix(info.URL, ".tbz2") {
			return extractFromTarBz2(data, info.FilePath, archiveName, pluginDir, verbose)
		} else if strings.HasSuffix(info.URL, ".zip") {
			return extractFromZip(data, info.FilePath, archiveName, pluginDir, verbose)
		} else if before, ok := strings.CutSuffix(filename, ".gz"); ok {
			return decompressGz(data, before, pluginDir, verbose)
		} else if before, ok := strings.CutSuffix(filename, ".xz"); ok {
			return decompressXz(data, before, pluginDir, verbose)
		} else if before, ok := strings.CutSuffix(filename, ".bz2"); ok {
			return decompressBz2(data, before, pluginDir, verbose)
		}
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
func extractFromTarGz(data []byte, targetFile, archiveName, pluginDir string, verbose bool) (string, error) {
	// Create gzip reader.
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader.
	tr := tar.NewReader(gzr)

	// Define file selection strategy
	type candidate struct {
		path     string
		priority int
	}

	selectFile := func(name string, mode os.FileMode) int {
		// Priority 1: Explicit target file (highest priority)
		if targetFile != "" && (name == targetFile || strings.HasSuffix(name, "/"+targetFile)) {
			return 100
		}

		// Priority 2: File matching archive name
		if filepath.Base(name) == archiveName {
			return 90
		}

		// Priority 3: Executable file
		if mode&0o111 != 0 {
			return 80
		}

		// Priority 4: Any regular file (fallback)
		return 10
	}

	var best *candidate
	var foundFiles []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		foundFiles = append(foundFiles, header.Name)
		priority := selectFile(header.Name, header.FileInfo().Mode())

		if best == nil || priority > best.priority {
			best = &candidate{path: header.Name, priority: priority}
			// If we found explicit target or archive match, stop searching
			if priority >= 90 {
				break
			}
		}
	}

	// Determine target path or error
	targetPath := ""
	if best != nil {
		targetPath = best.path
	} else if targetFile != "" {
		return "", fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	} else if len(foundFiles) == 0 {
		return "", fmt.Errorf("no files found in archive")
	} else if len(foundFiles) > 1 {
		return "", fmt.Errorf("multiple files in archive but none match expected plugin name '%s' (found: %v). Specify the plugin file with :filename", archiveName, foundFiles)
	} else {
		targetPath = foundFiles[0]
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
func extractFromZip(data []byte, targetFile, archiveName, pluginDir string, verbose bool) (string, error) {
	// Create zip reader.
	reader := bytes.NewReader(data)
	zr, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Define file selection strategy (same as tar)
	type candidate struct {
		file     *zip.File
		priority int
	}

	selectFile := func(name string, mode os.FileMode) int {
		if targetFile != "" && (name == targetFile || strings.HasSuffix(name, "/"+targetFile)) {
			return 100
		}
		if filepath.Base(name) == archiveName {
			return 90
		}
		if mode&0o111 != 0 {
			return 80
		}
		return 10
	}

	var best *candidate
	var foundFiles []string

	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}

		foundFiles = append(foundFiles, f.Name)
		priority := selectFile(f.Name, f.FileInfo().Mode())

		if best == nil || priority > best.priority {
			best = &candidate{file: f, priority: priority}
			if priority >= 90 {
				break
			}
		}
	}

	var targetZipFile *zip.File
	if best != nil {
		targetZipFile = best.file
	} else if targetFile != "" {
		return "", fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	} else if len(foundFiles) == 0 {
		return "", fmt.Errorf("no files found in archive")
	} else if len(foundFiles) > 1 {
		return "", fmt.Errorf("multiple files in archive but none match expected plugin name '%s' (found: %v). Specify the plugin file with :filename", archiveName, foundFiles)
	} else {
		targetZipFile = zr.File[0]
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

// extractFromTarXz extracts a plugin from a tar.xz archive.
func extractFromTarXz(data []byte, targetFile, archiveName, pluginDir string, verbose bool) (string, error) {
	// Create xz reader.
	xzr, err := xz.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create xz reader: %w", err)
	}

	// Create tar reader.
	tr := tar.NewReader(xzr)

	// Define file selection strategy (same as tar.gz)
	type candidate struct {
		path     string
		priority int
	}

	selectFile := func(name string, mode os.FileMode) int {
		if targetFile != "" && (name == targetFile || strings.HasSuffix(name, "/"+targetFile)) {
			return 100
		}
		if filepath.Base(name) == archiveName {
			return 90
		}
		if mode&0o111 != 0 {
			return 80
		}
		return 10
	}

	var best *candidate
	var foundFiles []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		foundFiles = append(foundFiles, header.Name)
		priority := selectFile(header.Name, header.FileInfo().Mode())

		if best == nil || priority > best.priority {
			best = &candidate{path: header.Name, priority: priority}
			if priority >= 90 {
				break
			}
		}
	}

	// Determine target path or error
	targetPath := ""
	if best != nil {
		targetPath = best.path
	} else if targetFile != "" {
		return "", fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	} else if len(foundFiles) == 0 {
		return "", fmt.Errorf("no files found in archive")
	} else if len(foundFiles) > 1 {
		return "", fmt.Errorf("multiple files in archive but none match expected plugin name '%s' (found: %v). Specify the plugin file with :filename", archiveName, foundFiles)
	} else {
		targetPath = foundFiles[0]
	}

	// Reset readers to extract the target file.
	xzr, err = xz.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create xz reader: %w", err)
	}

	tr = tar.NewReader(xzr)

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

		limitedReader := security.NewLimitedReader(tr, 100*1024*1024)
		_, copyErr := io.Copy(out, limitedReader)
		closeErr := out.Close()

		if copyErr != nil {
			return "", fmt.Errorf("failed to extract plugin: %w", copyErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("failed to close plugin file: %w", closeErr)
		}

		if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
			return "", fmt.Errorf("failed to make plugin executable: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Extracted plugin to: %s\n", destPath)
		}

		return destPath, nil
	}
}

// extractFromTarBz2 extracts a plugin from a tar.bz2 archive.
func extractFromTarBz2(data []byte, targetFile, archiveName, pluginDir string, verbose bool) (string, error) {
	// Create bzip2 reader.
	bzr := bzip2.NewReader(bytes.NewReader(data))

	// Create tar reader.
	tr := tar.NewReader(bzr)

	// Define file selection strategy (same as tar.gz)
	type candidate struct {
		path     string
		priority int
	}

	selectFile := func(name string, mode os.FileMode) int {
		if targetFile != "" && (name == targetFile || strings.HasSuffix(name, "/"+targetFile)) {
			return 100
		}
		if filepath.Base(name) == archiveName {
			return 90
		}
		if mode&0o111 != 0 {
			return 80
		}
		return 10
	}

	var best *candidate
	var foundFiles []string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Typeflag == tar.TypeDir {
			continue
		}

		foundFiles = append(foundFiles, header.Name)
		priority := selectFile(header.Name, header.FileInfo().Mode())

		if best == nil || priority > best.priority {
			best = &candidate{path: header.Name, priority: priority}
			if priority >= 90 {
				break
			}
		}
	}

	// Determine target path or error
	targetPath := ""
	if best != nil {
		targetPath = best.path
	} else if targetFile != "" {
		return "", fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	} else if len(foundFiles) == 0 {
		return "", fmt.Errorf("no files found in archive")
	} else if len(foundFiles) > 1 {
		return "", fmt.Errorf("multiple files in archive but none match expected plugin name '%s' (found: %v). Specify the plugin file with :filename", archiveName, foundFiles)
	} else {
		targetPath = foundFiles[0]
	}

	// Reset readers to extract the target file.
	bzr = bzip2.NewReader(bytes.NewReader(data))
	tr = tar.NewReader(bzr)

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

		limitedReader := security.NewLimitedReader(tr, 100*1024*1024)
		_, copyErr := io.Copy(out, limitedReader)
		closeErr := out.Close()

		if copyErr != nil {
			return "", fmt.Errorf("failed to extract plugin: %w", copyErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("failed to close plugin file: %w", closeErr)
		}

		if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
			return "", fmt.Errorf("failed to make plugin executable: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Extracted plugin to: %s\n", destPath)
		}

		return destPath, nil
	}
}

// decompressGz decompresses a single gzipped file.
func decompressGz(data []byte, filename, pluginDir string, verbose bool) (string, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	destPath := filepath.Join(pluginDir, filename)
	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return "", fmt.Errorf("failed to create plugin file: %w", err)
	}

	limitedReader := security.NewLimitedReader(gzr, 100*1024*1024)
	_, copyErr := io.Copy(out, limitedReader)
	closeErr := out.Close()

	if copyErr != nil {
		return "", fmt.Errorf("failed to decompress plugin: %w", copyErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("failed to close plugin file: %w", closeErr)
	}

	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Decompressed plugin to: %s\n", destPath)
	}

	return destPath, nil
}

// decompressXz decompresses a single xz-compressed file.
func decompressXz(data []byte, filename, pluginDir string, verbose bool) (string, error) {
	xzr, err := xz.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create xz reader: %w", err)
	}

	destPath := filepath.Join(pluginDir, filename)
	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return "", fmt.Errorf("failed to create plugin file: %w", err)
	}

	limitedReader := security.NewLimitedReader(xzr, 100*1024*1024)
	_, copyErr := io.Copy(out, limitedReader)
	closeErr := out.Close()

	if copyErr != nil {
		return "", fmt.Errorf("failed to decompress plugin: %w", copyErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("failed to close plugin file: %w", closeErr)
	}

	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Decompressed plugin to: %s\n", destPath)
	}

	return destPath, nil
}

// decompressBz2 decompresses a single bzip2-compressed file.
func decompressBz2(data []byte, filename, pluginDir string, verbose bool) (string, error) {
	bzr := bzip2.NewReader(bytes.NewReader(data))

	destPath := filepath.Join(pluginDir, filename)
	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return "", fmt.Errorf("failed to create plugin file: %w", err)
	}

	limitedReader := security.NewLimitedReader(bzr, 100*1024*1024)
	_, copyErr := io.Copy(out, limitedReader)
	closeErr := out.Close()

	if copyErr != nil {
		return "", fmt.Errorf("failed to decompress plugin: %w", copyErr)
	}
	if closeErr != nil {
		return "", fmt.Errorf("failed to close plugin file: %w", closeErr)
	}

	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return "", fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Decompressed plugin to: %s\n", destPath)
	}

	return destPath, nil
}
