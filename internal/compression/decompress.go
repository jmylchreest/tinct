package compression

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jmylchreest/tinct/internal/security"
	"github.com/ulikunitz/xz"
)

// decompressGz decompresses a single gzipped file.
func decompressGz(data []byte, filename, destDir string, verbose bool) (*ExtractResult, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	destPath := filepath.Join(destDir, filename)
	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin file: %w", err)
	}

	limitedReader := security.NewLimitedReader(gzr, 100*1024*1024)
	_, copyErr := io.Copy(out, limitedReader)
	closeErr := out.Close()

	if copyErr != nil {
		return nil, fmt.Errorf("failed to decompress plugin: %w", copyErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close plugin file: %w", closeErr)
	}

	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return nil, fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Decompressed plugin to: %s\n", destPath)
	}

	return &ExtractResult{
		Path:       destPath,
		WasArchive: false,
	}, nil
}

// decompressXz decompresses a single xz-compressed file.
func decompressXz(data []byte, filename, destDir string, verbose bool) (*ExtractResult, error) {
	xzr, err := xz.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create xz reader: %w", err)
	}

	destPath := filepath.Join(destDir, filename)
	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin file: %w", err)
	}

	limitedReader := security.NewLimitedReader(xzr, 100*1024*1024)
	_, copyErr := io.Copy(out, limitedReader)
	closeErr := out.Close()

	if copyErr != nil {
		return nil, fmt.Errorf("failed to decompress plugin: %w", copyErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close plugin file: %w", closeErr)
	}

	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return nil, fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Decompressed plugin to: %s\n", destPath)
	}

	return &ExtractResult{
		Path:       destPath,
		WasArchive: false,
	}, nil
}

// decompressBz2 decompresses a single bzip2-compressed file.
func decompressBz2(data []byte, filename, destDir string, verbose bool) (*ExtractResult, error) {
	bzr := bzip2.NewReader(bytes.NewReader(data))

	destPath := filepath.Join(destDir, filename)
	out, err := os.Create(destPath) // #nosec G304 - Plugin destination path controlled by application
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin file: %w", err)
	}

	limitedReader := security.NewLimitedReader(bzr, 100*1024*1024)
	_, copyErr := io.Copy(out, limitedReader)
	closeErr := out.Close()

	if copyErr != nil {
		return nil, fmt.Errorf("failed to decompress plugin: %w", copyErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close plugin file: %w", closeErr)
	}

	if err := os.Chmod(destPath, 0o755); err != nil { // #nosec G302 - Plugin executable needs execute permission
		return nil, fmt.Errorf("failed to make plugin executable: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Decompressed plugin to: %s\n", destPath)
	}

	return &ExtractResult{
		Path:       destPath,
		WasArchive: false,
	}, nil
}
