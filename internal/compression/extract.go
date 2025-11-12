// Package compression provides utilities for extracting and decompressing plugin archives.
package compression

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExtractResult contains the result of an extraction operation.
type ExtractResult struct {
	// Path to the extracted plugin file
	Path string
	// Whether the input was an archive (true) or a direct file (false)
	WasArchive bool
}

// ExtractPlugin detects format and extracts a plugin from downloaded data.
// It handles:
// - Tar archives (.tar.gz, .tar.xz, .tar.bz2)
// - Zip archives (.zip)
// - Standalone compressed files (.gz, .xz, .bz2)
// - Raw uncompressed files (.py, .sh, binaries)
//
// Parameters:
//   - data: The downloaded file data
//   - url: Original download URL (used for format detection)
//   - filename: Base filename (used for naming output)
//   - targetFile: Optional specific file to extract from archive (e.g., "tinct-plugin-wob")
//   - archiveName: Base name of archive for finding plugin binary (e.g., "tinct-plugin-wob")
//   - destDir: Destination directory for extracted files
//   - contentType: HTTP Content-Type header (optional, used for detection)
//   - verbose: Whether to print extraction progress
//
// Returns the path to the extracted plugin file.
func ExtractPlugin(data []byte, url, filename, targetFile, archiveName, destDir, contentType string, verbose bool) (*ExtractResult, error) {
	// First try Content-Type based detection
	if contentType != "" {
		if result, err := extractByContentType(data, url, filename, targetFile, archiveName, destDir, contentType, verbose); result != nil || err != nil {
			return result, err
		}
	}

	// Fall back to filename extension detection
	if result, err := extractByFileExtension(data, url, filename, targetFile, archiveName, destDir, verbose); result != nil || err != nil {
		return result, err
	}

	// Not an archive - treat as direct plugin file
	destPath := filepath.Join(destDir, filename)

	// Write file
	// #nosec G306 -- Plugin executable needs exec permissions
	if err := os.WriteFile(destPath, data, 0o755); err != nil {
		return nil, fmt.Errorf("failed to write plugin file: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Saved plugin to: %s\n", destPath)
	}

	return &ExtractResult{
		Path:       destPath,
		WasArchive: false,
	}, nil
}

// extractByContentType attempts to extract based on HTTP Content-Type header.
func extractByContentType(data []byte, url, filename, targetFile, archiveName, destDir, contentType string, verbose bool) (*ExtractResult, error) {
	switch {
	// Tar+gzip archives
	case strings.Contains(contentType, "application/gzip") || strings.Contains(contentType, "application/x-gzip"):
		// Could be tar.gz or plain .gz - check filename
		if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
			return extractFromTarGz(data, targetFile, archiveName, destDir, verbose)
		}
		// Plain gzip file
		return decompressGz(data, strings.TrimSuffix(filename, ".gz"), destDir, verbose)

	// Tar+xz archives
	case strings.Contains(contentType, "application/x-xz"):
		// Could be tar.xz or plain .xz - check filename
		if strings.HasSuffix(url, ".tar.xz") || strings.HasSuffix(url, ".txz") {
			return extractFromTarXz(data, targetFile, archiveName, destDir, verbose)
		}
		// Plain xz file
		return decompressXz(data, strings.TrimSuffix(filename, ".xz"), destDir, verbose)

	// Tar+bzip2 archives
	case strings.Contains(contentType, "application/x-bzip2"):
		// Could be tar.bz2 or plain .bz2 - check filename
		if strings.HasSuffix(url, ".tar.bz2") || strings.HasSuffix(url, ".tbz") || strings.HasSuffix(url, ".tbz2") {
			return extractFromTarBz2(data, targetFile, archiveName, destDir, verbose)
		}
		// Plain bzip2 file
		return decompressBz2(data, strings.TrimSuffix(filename, ".bz2"), destDir, verbose)

	// Zip archives
	case strings.Contains(contentType, "application/zip") || strings.Contains(contentType, "application/x-zip-compressed"):
		return extractFromZip(data, targetFile, archiveName, destDir, verbose)

	// Tar archives (uncompressed)
	case strings.Contains(contentType, "application/x-tar"):
		// Assume it's compressed based on filename
		if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
			return extractFromTarGz(data, targetFile, archiveName, destDir, verbose)
		} else if strings.HasSuffix(url, ".tar.xz") || strings.HasSuffix(url, ".txz") {
			return extractFromTarXz(data, targetFile, archiveName, destDir, verbose)
		} else if strings.HasSuffix(url, ".tar.bz2") || strings.HasSuffix(url, ".tbz") || strings.HasSuffix(url, ".tbz2") {
			return extractFromTarBz2(data, targetFile, archiveName, destDir, verbose)
		}
		return nil, fmt.Errorf("uncompressed tar archives are not supported")

	// Generic binary/octet-stream - continue to filename detection
	case strings.Contains(contentType, "application/octet-stream"), contentType == "":
		return nil, nil
	}

	// Plain text files (scripts) - treat as direct files
	if strings.HasPrefix(contentType, "text/") {
		return nil, nil
	}

	// Unknown content type - continue to filename detection
	return nil, nil
}

// extractByFileExtension attempts to extract based on filename extension.
func extractByFileExtension(data []byte, url, filename, targetFile, archiveName, destDir string, verbose bool) (*ExtractResult, error) {
	// Check for tar archives
	switch {
	case strings.HasSuffix(url, ".tar.gz"), strings.HasSuffix(url, ".tgz"):
		return extractFromTarGz(data, targetFile, archiveName, destDir, verbose)
	case strings.HasSuffix(url, ".tar.xz"), strings.HasSuffix(url, ".txz"):
		return extractFromTarXz(data, targetFile, archiveName, destDir, verbose)
	case strings.HasSuffix(url, ".tar.bz2"), strings.HasSuffix(url, ".tbz"), strings.HasSuffix(url, ".tbz2"):
		return extractFromTarBz2(data, targetFile, archiveName, destDir, verbose)
	case strings.HasSuffix(url, ".zip"):
		return extractFromZip(data, targetFile, archiveName, destDir, verbose)
	}

	// Check for standalone compressed files
	if before, ok := strings.CutSuffix(filename, ".gz"); ok {
		return decompressGz(data, before, destDir, verbose)
	}
	if before, ok := strings.CutSuffix(filename, ".xz"); ok {
		return decompressXz(data, before, destDir, verbose)
	}
	if before, ok := strings.CutSuffix(filename, ".bz2"); ok {
		return decompressBz2(data, before, destDir, verbose)
	}

	// Not a recognized archive format
	return nil, nil
}

// GetArchiveBaseName extracts the base name from an archive filename.
// For example: "tinct-plugin-wob_0.0.1_Linux_x86_64.tar.gz" -> "tinct-plugin-wob".
func GetArchiveBaseName(filename string) string {
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
