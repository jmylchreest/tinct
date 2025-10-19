// Package image provides utilities for loading and processing images.
package image

import (
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"os"
	"path/filepath"

	_ "golang.org/x/image/webp" // Register WebP format
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
// Supported formats: JPEG, PNG, GIF, WebP
func (l *FileLoader) Load(path string) (image.Image, error) {
	// Validate path
	if path == "" {
		return nil, fmt.Errorf("image path cannot be empty")
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("image file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to stat image file: %w", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Decode the image
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image (format: %s): %w", format, err)
	}

	return img, nil
}

// ValidateImagePath checks if the given path is valid and points to a supported image file.
func ValidateImagePath(path string) error {
	// Check if path is empty
	if path == "" {
		return fmt.Errorf("image path cannot be empty")
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("image file not found: %s", path)
		}
		return fmt.Errorf("failed to access image file: %w", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	// Check file extension (basic validation)
	ext := filepath.Ext(path)
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !supportedExts[ext] {
		return fmt.Errorf("unsupported image format: %s (supported: jpg, jpeg, png, gif, webp)", ext)
	}

	return nil
}

// GetImageDimensions returns the width and height of an image without fully loading it.
func GetImageDimensions(path string) (width, height int, err error) {
	file, err := os.Open(path)
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
