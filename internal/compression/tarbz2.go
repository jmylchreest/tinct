package compression

import (
	"bytes"
	"compress/bzip2"
	"io"
)

// extractFromTarBz2 extracts a plugin from a tar.bz2 archive.
func extractFromTarBz2(data []byte, targetFile, archiveName, destDir string, verbose bool) (*ExtractResult, error) {
	return extractFromTar(func() (io.ReadCloser, error) {
		return io.NopCloser(bzip2.NewReader(bytes.NewReader(data))), nil
	}, targetFile, archiveName, destDir, verbose)
}
