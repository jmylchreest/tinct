package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// extractFromTarGz extracts a plugin from a tar.gz archive.
func extractFromTarGz(data []byte, targetFile, archiveName, destDir string, verbose bool) (*ExtractResult, error) {
	return extractFromTar(func() (io.ReadCloser, error) {
		gzr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return gzr, nil
	}, targetFile, archiveName, destDir, verbose)
}
