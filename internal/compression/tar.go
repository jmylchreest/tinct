package compression

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/security"
)

// newReaderFunc creates a fresh decompressing reader over the raw archive
// data.  It is called twice: once for the enumeration pass and once for the
// extraction pass.  The returned io.ReadCloser is closed by the caller; if
// the decompressor has no Close method, wrap it with io.NopCloser.
type newReaderFunc func() (io.ReadCloser, error)

// extractFromTar implements the shared two-pass tar extraction logic used by
// all tar-based formats (gzip, xz, bzip2).
//
// Pass 1 — enumerate entries through a 100 MB LimitedReader to find the best
// candidate file.  Pass 2 — re-open the archive and extract just that file
// with a per-file 100 MB decompression limit.
func extractFromTar(newReader newReaderFunc, targetFile, archiveName, destDir string, verbose bool) (*ExtractResult, error) { //nolint:gocyclo,gocognit // archive extraction with two-pass error handling
	// --- Pass 1: enumerate entries to find the best candidate ---

	rc, err := newReader()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	limitedEnum := security.NewLimitedReader(rc, 100*1024*1024)
	tr := tar.NewReader(limitedEnum)

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

		// Only consider regular files — skip directories, symlinks,
		// hardlinks, and device nodes to prevent symlink attacks.
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA {
			continue
		}

		// Validate entry path against directory traversal.
		if security.ValidateFilePath(header.Name, destDir) != nil {
			continue
		}

		foundFiles = append(foundFiles, header.Name)
		priority := selectFile(header.Name, header.FileInfo().Mode())

		if best == nil || priority > best.priority {
			best = &candidate{path: header.Name, priority: priority}
			// If we found explicit target or archive match, stop searching.
			if priority >= 90 {
				break
			}
		}
	}

	// Determine target path or error.
	targetPath := ""
	switch {
	case best != nil:
		targetPath = best.path
	case targetFile != "":
		return nil, fmt.Errorf("file '%s' not found in archive (found: %v)", targetFile, foundFiles)
	case len(foundFiles) == 0:
		return nil, fmt.Errorf("no files found in archive")
	case len(foundFiles) > 1:
		return nil, fmt.Errorf("multiple files in archive but none match expected plugin name '%s' (found: %v)", archiveName, foundFiles)
	default:
		targetPath = foundFiles[0]
	}

	// --- Pass 2: re-open and extract the target file ---

	rc2, err := newReader()
	if err != nil {
		return nil, err
	}
	defer rc2.Close()

	tr = tar.NewReader(rc2)

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

		// Extract the file.
		destPath := filepath.Join(destDir, filepath.Base(targetPath))

		out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin file: %w", err)
		}

		// Limit decompression size to prevent zip bombs (100 MB limit for plugins).
		limitedReader := security.NewLimitedReader(tr, 100*1024*1024)
		_, copyErr := io.Copy(out, limitedReader)
		closeErr := out.Close() // Close immediately instead of defer.

		if copyErr != nil {
			return nil, fmt.Errorf("failed to extract plugin: %w", copyErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close plugin file: %w", closeErr)
		}

		// Make executable.
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
