package compression

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/security"
)

// extractFromTarGz extracts a plugin from a tar.gz archive.
func extractFromTarGz(data []byte, targetFile, archiveName, destDir string, verbose bool) (*ExtractResult, error) {
	// Create gzip reader
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
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
			return nil, fmt.Errorf("failed to read tar archive: %w", err)
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
		return nil, fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	} else if len(foundFiles) == 0 {
		return nil, fmt.Errorf("no files found in archive")
	} else if len(foundFiles) > 1 {
		return nil, fmt.Errorf("multiple files in archive but none match expected plugin name '%s' (found: %v)", archiveName, foundFiles)
	} else {
		targetPath = foundFiles[0]
	}

	// Reset readers to extract the target file
	gzr, err = gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr = tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("file not found in archive")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar archive: %w", err)
		}

		if header.Name != targetPath {
			continue
		}

		// Extract the file
		destPath := filepath.Join(destDir, filepath.Base(targetPath))

		out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin file: %w", err)
		}

		// Limit decompression size to prevent zip bombs (100MB limit for plugins)
		limitedReader := security.NewLimitedReader(tr, 100*1024*1024)
		_, copyErr := io.Copy(out, limitedReader)
		closeErr := out.Close() // Close immediately instead of defer

		if copyErr != nil {
			return nil, fmt.Errorf("failed to extract plugin: %w", copyErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close plugin file: %w", closeErr)
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
}
