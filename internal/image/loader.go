// Package image provides utilities for loading and processing images.
package image

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"strings"

	_ "golang.org/x/image/webp" // Register WebP format

	httputil "github.com/jmylchreest/tinct/internal/util/http"
)

// Loader handles loading images from various sources.
type Loader interface {
	// Load loads an image from the given path.
	Load(path string) (image.Image, error)
}

// FileLoader loads images from the local filesystem.
type FileLoader struct{}

// NewFileLoader creates a new FileLoader instance.
func NewFileLoader() *FileLoader {
	return &FileLoader{}
}

// Load loads an image from a file path.
// Supported formats: JPEG, PNG, GIF, WebP.
func (l *FileLoader) Load(path string) (image.Image, error) {
	// Validate path.
	if path == "" {
		return nil, fmt.Errorf("image path cannot be empty")
	}

	// Check if file exists.
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("image file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to stat image file: %w", err)
	}

	// Check if it's a directory.
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Open the file.
	file, err := os.Open(path) // #nosec G304 - User-specified image path, intended to be read
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Decode the image.
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}

	return img, nil
}

// ValidateImagePath checks if the given path is valid and points to a supported image file or directory.
// Supports both local file paths, directories, and HTTP(S) URLs.
// For local files, it verifies the file exists and can be decoded.
// For directories, it verifies the directory exists (actual scanning happens later).
// For HTTP(S) URLs, it just validates the URL format (actual fetching happens later).
func ValidateImagePath(path string) error {
	// Check if path is empty.
	if path == "" {
		return fmt.Errorf("image path cannot be empty")
	}

	// Check if it's an HTTP(S) URL.
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		// URL validation - just ensure it looks like a valid URL.
		// We don't fetch it here to avoid double-fetching.
		return nil
	}

	// Local file path validation.
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("image file or directory not found: %s", path)
		}
		return fmt.Errorf("failed to access image path: %w", err)
	}

	// If it's a directory, just verify it exists (scanning happens later).
	if info.IsDir() {
		return nil
	}

	// Attempt to decode the image config to verify it's a supported format.
	// This will work with any format that the image library supports.
	file, err := os.Open(path) // #nosec G304 - User-specified image path, intended to be read
	if err != nil {
		return fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	_, format, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("unsupported or invalid image format: %w", err)
	}

	// Successfully decoded - the image library supports this format.
	_ = format // format contains the detected format name (jpeg, png, gif, webp, etc.)

	return nil
}

// SupportedImageExtensions returns a list of supported image file extensions.
func SupportedImageExtensions() []string {
	return []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
}

// isImageFile checks if a file has a supported image extension.
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(SupportedImageExtensions(), ext)
}

// ScanDirectoryForImages scans a directory and returns all valid image files.
// It does not recurse into subdirectories, but follows symlinks.
func ScanDirectoryForImages(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var imageFiles []string
	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		// For symlinks, stat the target to determine if it's a file.
		info, err := os.Stat(fullPath)
		if err != nil {
			// Skip entries we can't stat (broken symlinks, permission issues).
			continue
		}

		// Skip directories (including symlinks to directories).
		if info.IsDir() {
			continue
		}

		// Check if file has a supported image extension.
		if isImageFile(entry.Name()) {
			imageFiles = append(imageFiles, fullPath)
		}
	}

	if len(imageFiles) == 0 {
		return nil, fmt.Errorf("no supported image files found in directory: %s", dirPath)
	}

	return imageFiles, nil
}

// SelectRandomImage selects a random image from a list of image paths.
// Uses crypto/rand for cryptographically secure randomness.
func SelectRandomImage(imagePaths []string) (string, error) {
	if len(imagePaths) == 0 {
		return "", fmt.Errorf("image path list is empty")
	}

	// Use crypto/rand for truly random selection.
	maxIndex := big.NewInt(int64(len(imagePaths)))
	randomIndex, err := rand.Int(rand.Reader, maxIndex)
	if err != nil {
		// Fallback to using binary random bytes if crypto/rand.Int fails.
		var buf [8]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		index := int(binary.LittleEndian.Uint64(buf[:]) % uint64(len(imagePaths)))
		return imagePaths[index], nil
	}

	return imagePaths[randomIndex.Int64()], nil
}

// ResolveImagePath resolves a path that could be a file or directory.
// If the path is a directory, it scans for images and returns a random one.
// If the path is a file, it returns the path as-is.
// For HTTP(S) URLs, it returns the URL as-is.
func ResolveImagePath(path string) (string, error) {
	// URLs are returned as-is.
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}

	// Check if path exists.
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to access path: %w", err)
	}

	// If it's a file, return as-is.
	if !info.IsDir() {
		return path, nil
	}

	// It's a directory - scan for images and select randomly.
	imageFiles, err := ScanDirectoryForImages(path)
	if err != nil {
		return "", err
	}

	selectedImage, err := SelectRandomImage(imageFiles)
	if err != nil {
		return "", err
	}

	return selectedImage, nil
}

// GetImageDimensions returns the width and height of an image without fully loading it.
func GetImageDimensions(path string) (width, height int, err error) {
	file, err := os.Open(path) // #nosec G304 - User-specified image path, intended to be read
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}

	return config.Width, config.Height, nil
}

// SmartLoader loads images from both local files and HTTP(S) URLs.
type SmartLoader struct {
	fileLoader *FileLoader
}

// NewSmartLoader creates a new SmartLoader instance.
func NewSmartLoader() *SmartLoader {
	return &SmartLoader{
		fileLoader: NewFileLoader(),
	}
}

// Load loads an image from either a local file path or HTTP(S) URL.
func (l *SmartLoader) Load(path string) (image.Image, error) {
	// Check if it's an HTTP(S) URL.
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return l.loadFromURL(path)
	}

	// Load from local file.
	return l.fileLoader.Load(path)
}

// loadFromURL fetches and decodes an image from an HTTP(S) URL.
func (l *SmartLoader) loadFromURL(url string) (image.Image, error) {
	// Fetch the image data.
	ctx := context.Background()
	data, err := httputil.Fetch(ctx, url, httputil.FetchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image from URL: %w", err)
	}

	// Decode the image from the fetched data.
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}

	return img, nil
}
