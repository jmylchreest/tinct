// Package image provides an input plugin for extracting colour palettes from images.
package image

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"math/rand"
	"path/filepath"
	"time"
)

// calculateSeed determines the seed value for k-means clustering based on the seed mode.
func (p *Plugin) calculateSeed(img image.Image, imagePath string) (int64, error) {
	switch SeedMode(p.seedMode) {
	case SeedModeContent:
		return calculateContentSeed(img)
	case SeedModeFilepath:
		return calculateFilepathSeed(imagePath)
	case SeedModeManual:
		return p.seedValue, nil
	case SeedModeRandom:
		return generateRandomSeed(), nil
	default:
		return 0, fmt.Errorf("unknown seed mode: %s", p.seedMode)
	}
}

// calculateContentSeed generates a deterministic seed from image content.
// This hashes the pixel data to create a seed that's consistent for the same image content,
// regardless of filename or location.
func calculateContentSeed(img image.Image) (int64, error) {
	bounds := img.Bounds()
	hasher := sha256.New()

	// Hash image dimensions.
	dimBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(dimBytes[0:4], uint32(bounds.Dx()))
	binary.LittleEndian.PutUint32(dimBytes[4:8], uint32(bounds.Dy()))
	hasher.Write(dimBytes)

	// Sample pixels in a grid pattern for efficiency.
	// We don't need to hash every single pixel, just enough to uniquely identify the image.
	step := max(bounds.Dx()/100, bounds.Dy()/100, 1)
	pixelBytes := make([]byte, 4)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, a := img.At(x, y).RGBA()
			// Convert to 8-bit values.
			pixelBytes[0] = byte(r >> 8)
			pixelBytes[1] = byte(g >> 8)
			pixelBytes[2] = byte(b >> 8)
			pixelBytes[3] = byte(a >> 8)
			hasher.Write(pixelBytes)
		}
	}

	// Convert hash to int64 seed.
	hash := hasher.Sum(nil)
	seed := int64(binary.LittleEndian.Uint64(hash[:8]))
	return seed, nil
}

// calculateFilepathSeed generates a deterministic seed from the absolute file path.
// This creates a seed that's consistent for the same file path, allowing different
// images at the same location to produce different results.
func calculateFilepathSeed(imagePath string) (int64, error) {
	// Resolve to absolute path.
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		// If we can't resolve absolute path, use the path as-is.
		absPath = imagePath
	}

	// For URLs, just use the URL as-is.
	if isURL(imagePath) {
		absPath = imagePath
	}

	// Hash the absolute path.
	hasher := sha256.New()
	hasher.Write([]byte(absPath))
	hash := hasher.Sum(nil)
	seed := int64(binary.LittleEndian.Uint64(hash[:8]))
	return seed, nil
}

// generateRandomSeed generates a non-deterministic random seed.
func generateRandomSeed() int64 {
	// #nosec G404 -- Random seed generation is intentionally non-deterministic
	return time.Now().UnixNano() + int64(rand.Intn(1000000))
}

// isURL checks if a path is an HTTP/HTTPS URL.
func isURL(path string) bool {
	return len(path) >= 7 && (path[:7] == "http://" || path[:8] == "https://")
}
