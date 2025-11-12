package compression

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/security"
)

// extractFromTarBz2 extracts a plugin from a tar.bz2 archive.
func extractFromTarBz2(data []byte, targetFile, archiveName, destDir string, verbose bool) (*ExtractResult, error) {
	// Create bzip2 reader
	bzr := bzip2.NewReader(bytes.NewReader(data))

	// Create tar reader
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
			return nil, fmt.Errorf("failed to read tar archive: %w", err)
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
		return nil, fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	} else if len(foundFiles) == 0 {
		return nil, fmt.Errorf("no files found in archive")
	} else if len(foundFiles) > 1 {
		return nil, fmt.Errorf("multiple files in archive but none match expected plugin name '%s' (found: %v)", archiveName, foundFiles)
	} else {
		targetPath = foundFiles[0]
	}

	// Reset readers to extract the target file
	bzr = bzip2.NewReader(bytes.NewReader(data))
	tr = tar.NewReader(bzr)

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

		limitedReader := security.NewLimitedReader(tr, 100*1024*1024)
		_, copyErr := io.Copy(out, limitedReader)
		closeErr := out.Close()

		if copyErr != nil {
			return nil, fmt.Errorf("failed to extract plugin: %w", copyErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close plugin file: %w", closeErr)
		}

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
