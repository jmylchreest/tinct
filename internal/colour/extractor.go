// Package color provides color extraction and palette generation functionality.
package colour

import (
	"fmt"
	"image"
	"slices"
)

// Extractor defines the interface for color extraction algorithms.
type Extractor interface {
	// Extract extracts a color palette from an image.
	// The count parameter specifies the number of colors to extract.
	Extract(img image.Image, count int) (*Palette, error)
}

// Algorithm represents the color extraction algorithm type.
type Algorithm string

const (
	// AlgorithmKMeans uses k-means clustering for color extraction.
	AlgorithmKMeans Algorithm = "kmeans"

	// AlgorithmMedianCut uses median cut algorithm for color extraction.
	// Not yet implemented - placeholder for future.
	AlgorithmMedianCut Algorithm = "mediancut"

	// AlgorithmDominant extracts the most dominant (frequent) colors.
	// Not yet implemented - placeholder for future.
	AlgorithmDominant Algorithm = "dominant"
)

// ValidAlgorithms returns a list of valid algorithm names.
func ValidAlgorithms() []Algorithm {
	return []Algorithm{
		AlgorithmKMeans,
		// Future algorithms will be added here
	}
}

// IsValidAlgorithm checks if the given algorithm name is valid.
func IsValidAlgorithm(alg Algorithm) bool {
	return slices.Contains(ValidAlgorithms(), alg)
}

// NewExtractor creates a new Extractor based on the specified algorithm.
// Returns an error if the algorithm is not recognized or not yet implemented.
func NewExtractor(alg Algorithm) (Extractor, error) {
	switch alg {
	case AlgorithmKMeans:
		return NewKMeansExtractor(), nil
	case AlgorithmMedianCut:
		return nil, fmt.Errorf("median cut algorithm not yet implemented")
	case AlgorithmDominant:
		return nil, fmt.Errorf("dominant color algorithm not yet implemented")
	default:
		return nil, fmt.Errorf("unknown algorithm: %s (valid algorithms: %v)", alg, ValidAlgorithms())
	}
}
