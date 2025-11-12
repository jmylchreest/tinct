package compression

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/security"
)

// extractFromZip extracts a plugin from a zip archive.
func extractFromZip(data []byte, targetFile, archiveName, destDir string, verbose bool) (*ExtractResult, error) {
	// Create zip reader
	reader := bytes.NewReader(data)
	zr, err := zip.NewReader(reader, int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
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
		return nil, fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	} else if len(foundFiles) == 0 {
		return nil, fmt.Errorf("no files found in archive")
	} else if len(foundFiles) > 1 {
		return nil, fmt.Errorf("multiple files in archive but none match expected plugin name '%s' (found: %v)", archiveName, foundFiles)
	} else {
		targetZipFile = zr.File[0]
	}

	// Extract the file
	destPath := filepath.Join(destDir, filepath.Base(targetZipFile.Name))

	rc, err := targetZipFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file in archive: %w", err)
	}
	defer rc.Close()

	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin file: %w", err)
	}
	defer out.Close()

	// Limit decompression size to prevent zip bombs (100MB limit for plugins)
	limitedReader := security.NewLimitedReader(rc, 100*1024*1024)
	if _, err := io.Copy(out, limitedReader); err != nil {
		return nil, fmt.Errorf("failed to extract plugin: %w", err)
	}

	// Make executable
	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return nil, fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Extracted plugin to: %s\n", destPath)
	}

	return &ExtractResult{
		Path:       destPath,
		WasArchive: true,
	}, nil
}
