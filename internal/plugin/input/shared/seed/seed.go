// Package seed provides utilities for deterministic seed generation for k-means clustering.
// This is used by input plugins that process images to ensure reproducible color extraction.
package seed

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"image"
	"math/rand"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Mode determines how the random seed for k-means clustering is generated.
type Mode string

const (
	// ModeContent generates seed from image content hash (default, deterministic by content).
	ModeContent Mode = "content"
	// ModeFilepath generates seed from absolute file path hash (deterministic by path).
	ModeFilepath Mode = "filepath"
	// ModeManual uses a user-provided seed value.
	ModeManual Mode = "manual"
	// ModeRandom uses non-deterministic random seed (varies each run).
	ModeRandom Mode = "random"
)

// Config holds configuration for seed generation.
type Config struct {
	Mode  Mode   // Seed mode
	Value *int64 // Seed value (only used when Mode is ModeManual)
}

// Calculate determines the seed value based on the seed mode.
// img: the image to extract seed from (required for ModeContent)
// imagePath: the path to the image file (required for ModeFilepath)
// config: seed configuration
func Calculate(img image.Image, imagePath string, config Config) (int64, error) {
	switch config.Mode {
	case ModeContent:
		if img == nil {
			return 0, fmt.Errorf("image is required for content-based seed mode")
		}
		return CalculateContentSeed(img)
	case ModeFilepath:
		if imagePath == "" {
			return 0, fmt.Errorf("image path is required for filepath-based seed mode")
		}
		return CalculateFilepathSeed(imagePath)
	case ModeManual:
		if config.Value == nil {
			return 0, fmt.Errorf("seed value is required for manual seed mode")
		}
		return *config.Value, nil
	case ModeRandom:
		return GenerateRandomSeed(), nil
	default:
		return 0, fmt.Errorf("unknown seed mode: %s", config.Mode)
	}
}

// CalculateContentSeed generates a deterministic seed from image content.
// This hashes the pixel data to create a seed that's consistent for the same image content,
// regardless of filename or location.
func CalculateContentSeed(img image.Image) (int64, error) {
	if img == nil {
		return 0, fmt.Errorf("image cannot be nil")
	}

	bounds := img.Bounds()
	hasher := sha256.New()

	// Hash image dimensions
	dimBytes := make([]byte, 8)
	binary.LittleEndian.PutUint32(dimBytes[0:4], uint32(bounds.Dx())) // #nosec G115 -- image dimensions are safe to convert
	binary.LittleEndian.PutUint32(dimBytes[4:8], uint32(bounds.Dy())) // #nosec G115 -- image dimensions are safe to convert
	hasher.Write(dimBytes)

	// Sample pixels in a grid pattern for efficiency
	// We don't need to hash every single pixel, just enough to uniquely identify the image
	step := max(bounds.Dx()/100, bounds.Dy()/100, 1)
	pixelBytes := make([]byte, 4)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, a := img.At(x, y).RGBA()
			// Convert to 8-bit values
			pixelBytes[0] = byte(r >> 8)
			pixelBytes[1] = byte(g >> 8)
			pixelBytes[2] = byte(b >> 8)
			pixelBytes[3] = byte(a >> 8)
			hasher.Write(pixelBytes)
		}
	}

	// Convert hash to int64 seed
	hash := hasher.Sum(nil)
	seed := int64(binary.LittleEndian.Uint64(hash[:8])) // #nosec G115 -- hash conversion is safe
	return seed, nil
}

// CalculateFilepathSeed generates a deterministic seed from the absolute file path.
// This creates a seed that's consistent for the same file path, allowing different
// images at the same location to produce different results.
func CalculateFilepathSeed(imagePath string) (int64, error) {
	if imagePath == "" {
		return 0, fmt.Errorf("image path cannot be empty")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(imagePath)
	if err != nil {
		// If we can't resolve absolute path, use the path as-is
		absPath = imagePath
	}

	// For URLs, just use the URL as-is
	if isURL(imagePath) {
		absPath = imagePath
	}

	// Hash the absolute path
	hasher := sha256.New()
	hasher.Write([]byte(absPath))
	hash := hasher.Sum(nil)
	seed := int64(binary.LittleEndian.Uint64(hash[:8])) // #nosec G115 -- hash conversion is safe
	return seed, nil
}

// GenerateRandomSeed generates a non-deterministic random seed.
func GenerateRandomSeed() int64 {
	// #nosec G404 -- Random seed generation is intentionally non-deterministic
	return time.Now().UnixNano() + int64(rand.Intn(1000000))
}

// isURL checks if a path is an HTTP/HTTPS URL.
func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://")
}

// ValidModes returns a list of valid seed modes.
func ValidModes() []Mode {
	return []Mode{ModeContent, ModeFilepath, ModeManual, ModeRandom}
}

// ParseMode converts a string to a Mode.
// Returns an error if the string is not a valid mode.
func ParseMode(s string) (Mode, error) {
	mode := Mode(s)
	if slices.Contains(ValidModes(), mode) {
		return mode, nil
	}
	return "", fmt.Errorf("invalid seed mode: %s (valid: content, filepath, manual, random)", s)
}
