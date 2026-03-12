package compression

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ulikunitz/xz"
)

// extractFromTarXz extracts a plugin from a tar.xz archive.
func extractFromTarXz(data []byte, targetFile, archiveName, destDir string, verbose bool) (*ExtractResult, error) {
	return extractFromTar(func() (io.ReadCloser, error) {
		xzr, err := xz.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to create xz reader: %w", err)
		}
		return io.NopCloser(xzr), nil
	}, targetFile, archiveName, destDir, verbose)
}
