// Package color provides color extraction and palette generation functionality.
package colour

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
)

// KMeansExtractor implements color extraction using k-means clustering.
type KMeansExtractor struct {
	maxIterations int
	convergence   float64
	maxSamples    int
}

// NewKMeansExtractor creates a new KMeansExtractor with default settings.
func NewKMeansExtractor() *KMeansExtractor {
	return &KMeansExtractor{
		maxIterations: 20,   // Reduced from 50
		convergence:   2.0,  // Increased from 1.0 for faster convergence
		maxSamples:    5000, // Limit total samples for performance
	}
}

// Extract extracts colors from an image using k-means clustering.
// Returns colors with their relative weights (cluster sizes).
func (e *KMeansExtractor) Extract(img image.Image, count int) (*Palette, error) {
	if img == nil {
		return nil, fmt.Errorf("image cannot be nil")
	}
	if count < 1 {
		return nil, fmt.Errorf("color count must be at least 1, got %d", count)
	}
	if count > 256 {
		return nil, fmt.Errorf("color count too large: %d (maximum: 256)", count)
	}

	// Sample pixels from the image
	pixels := samplePixels(img)
	if len(pixels) == 0 {
		return nil, fmt.Errorf("no pixels found in image")
	}

	// Get unique colors first
	uniqueColors := make([]color.Color, 0, len(pixels))
	seen := make(map[RGB]bool)
	for _, p := range pixels {
		rgb := ToRGB(p)
		if !seen[rgb] {
			uniqueColors = append(uniqueColors, p)
			seen[rgb] = true
		}
	}

	// If we want more colors than unique colors exist, return all unique colors
	if count >= len(uniqueColors) {
		return NewPalette(uniqueColors), nil
	}

	// Run k-means clustering and get cluster weights
	centroids, weights := e.kmeans(pixels, count)

	// Convert centroids to colors
	colors := make([]color.Color, len(centroids))
	for i, c := range centroids {
		colors[i] = color.RGBA{
			R: uint8(c.R),
			G: uint8(c.G),
			B: uint8(c.B),
			A: 255,
		}
	}

	return NewPaletteWithWeights(colors, weights), nil
}

// point3D represents a point in 3D RGB color space.
type point3D struct {
	R, G, B float64
}

// distance calculates the Euclidean distance between two points in RGB space.
func (p point3D) distance(other point3D) float64 {
	dr := p.R - other.R
	dg := p.G - other.G
	db := p.B - other.B
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// samplePixels samples pixels from the image.
// For large images, we sample a subset to improve performance.
func samplePixels(img image.Image) []color.Color {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	totalPixels := width * height

	// Aggressive sampling for performance
	const maxSamples = 2000 // Reduced from 10000

	if totalPixels <= maxSamples {
		// Small image, sample all pixels
		pixels := make([]color.Color, 0, totalPixels)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				pixels = append(pixels, img.At(x, y))
			}
		}
		return pixels
	}

	// Large image, use grid sampling
	// Calculate step size to get approximately maxSamples
	step := max(int(math.Sqrt(float64(totalPixels)/float64(maxSamples))), 1)

	pixels := make([]color.Color, 0, maxSamples)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			pixels = append(pixels, img.At(x, y))
			if len(pixels) >= maxSamples {
				return pixels
			}
		}
	}

	return pixels
}

// kmeans performs k-means clustering on the pixel data.
// Returns centroids and their weights (relative cluster sizes).
func (e *KMeansExtractor) kmeans(pixels []color.Color, k int) ([]point3D, []float64) {
	// Convert colors to 3D points
	points := make([]point3D, len(pixels))
	for i, c := range pixels {
		rgb := ToRGB(c)
		points[i] = point3D{
			R: float64(rgb.R),
			G: float64(rgb.G),
			B: float64(rgb.B),
		}
	}

	// Initialize centroids using k-means++ algorithm
	centroids := e.initializeCentroidsKMeansPlusPlus(points, k)

	// Track cluster assignments
	assignments := make([]int, len(points))

	// Iterate until convergence or max iterations
	for iter := 0; iter < e.maxIterations; iter++ {
		// Assign each point to nearest centroid
		changed := 0
		for i, point := range points {
			nearest := e.findNearestCentroid(point, centroids)
			if assignments[i] != nearest {
				assignments[i] = nearest
				changed++
			}
		}

		// If very few assignments changed (< 1%), we've converged
		if float64(changed)/float64(len(points)) < 0.01 {
			break
		}

		// Recalculate centroids
		newCentroids := e.recalculateCentroids(points, assignments, k)

		// Check for convergence based on centroid movement
		totalMovement := 0.0
		for i := range centroids {
			totalMovement += centroids[i].distance(newCentroids[i])
		}
		avgMovement := totalMovement / float64(k)

		centroids = newCentroids

		// If centroids barely moved, we've converged
		if avgMovement < e.convergence {
			break
		}
	}

	// Calculate cluster weights (relative sizes)
	weights := make([]float64, k)
	for _, assignment := range assignments {
		weights[assignment]++
	}

	// Normalize weights to sum to 1.0
	totalPixels := float64(len(assignments))
	for i := range weights {
		weights[i] /= totalPixels
	}

	return centroids, weights
}

// initializeCentroidsKMeansPlusPlus initializes centroids using k-means++ algorithm.
// This provides better initial centroids than random selection.
func (e *KMeansExtractor) initializeCentroidsKMeansPlusPlus(points []point3D, k int) []point3D {
	if len(points) == 0 || k == 0 {
		return []point3D{}
	}

	centroids := make([]point3D, 0, k)

	// Choose first centroid randomly
	firstIdx := rand.Intn(len(points))
	centroids = append(centroids, points[firstIdx])

	// Choose remaining centroids
	for len(centroids) < k {
		// Calculate distances from each point to nearest centroid
		distances := make([]float64, len(points))
		totalDistance := 0.0

		for i, point := range points {
			minDist := math.MaxFloat64
			for _, centroid := range centroids {
				dist := point.distance(centroid)
				if dist < minDist {
					minDist = dist
				}
			}
			// Square the distance for k-means++
			distances[i] = minDist * minDist
			totalDistance += distances[i]
		}

		// Choose next centroid with probability proportional to squared distance
		if totalDistance == 0 {
			// All remaining points are too close or identical to existing centroids
			// Just duplicate an existing centroid slightly perturbed
			if len(centroids) > 0 {
				// Duplicate the last centroid with a tiny perturbation
				lastCentroid := centroids[len(centroids)-1]
				centroids = append(centroids, point3D{
					R: lastCentroid.R + 0.1,
					G: lastCentroid.G + 0.1,
					B: lastCentroid.B + 0.1,
				})
			}
			continue
		}

		target := rand.Float64() * totalDistance
		cumulative := 0.0
		for i, dist := range distances {
			cumulative += dist
			if cumulative >= target {
				centroids = append(centroids, points[i])
				break
			}
		}
	}

	return centroids
}

// findNearestCentroid finds the index of the nearest centroid to a point.
func (e *KMeansExtractor) findNearestCentroid(point point3D, centroids []point3D) int {
	minDist := math.MaxFloat64
	nearest := 0

	for i, centroid := range centroids {
		dist := point.distance(centroid)
		if dist < minDist {
			minDist = dist
			nearest = i
		}
	}

	return nearest
}

// recalculateCentroids recalculates centroid positions based on assigned points.
func (e *KMeansExtractor) recalculateCentroids(points []point3D, assignments []int, k int) []point3D {
	// Sum up all points assigned to each cluster
	sums := make([]point3D, k)
	counts := make([]int, k)

	for i, point := range points {
		cluster := assignments[i]
		sums[cluster].R += point.R
		sums[cluster].G += point.G
		sums[cluster].B += point.B
		counts[cluster]++
	}

	// Calculate averages
	centroids := make([]point3D, k)
	for i := range k {
		if counts[i] > 0 {
			centroids[i] = point3D{
				R: sums[i].R / float64(counts[i]),
				G: sums[i].G / float64(counts[i]),
				B: sums[i].B / float64(counts[i]),
			}
		} else {
			// Empty cluster - reinitialize randomly
			centroids[i] = points[rand.Intn(len(points))]
		}
	}

	return centroids
}
