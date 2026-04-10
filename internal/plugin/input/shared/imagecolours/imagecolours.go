// Package imagecolours provides shared helpers for extracting colour palettes
// from images and saving generation metadata. Used by AI input plugins.
package imagecolours

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/image"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/regions"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/seed"
)

const (
	// MainColorWeightRatio is the proportion of total weight allocated to
	// main palette colors when region extraction is enabled (90%).
	MainColorWeightRatio = 0.9

	// RegionWeightFactor determines how much weight region colors receive
	// relative to main palette colors. Region colors get 10% of the total weight.
	RegionWeightFactor = 0.1
)

// ExtractColors extracts a colour palette from the image at imagePath.
// It honours all commonflags settings (seed mode, colour count, region extraction).
func ExtractColors(imagePath string, verbose bool) (*colour.Palette, error) { //nolint:gocyclo
	// Load image using tinct's SmartLoader.
	loader := image.NewSmartLoader()
	img, err := loader.Load(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Prepare extractor options with seed.
	extractorOpts := colour.ExtractorOptions{}

	// Parse seed mode.
	seedMode, err := seed.ParseMode(commonflags.SeedMode)
	if err != nil {
		return nil, fmt.Errorf("invalid seed mode: %w", err)
	}

	// Calculate seed using shared utility.
	seedConfig := seed.Config{
		Mode:  seedMode,
		Value: nil,
	}
	if seedMode == seed.ModeManual {
		seedValue := commonflags.SeedValue
		seedConfig.Value = &seedValue
	}

	calculatedSeed, err := seed.Calculate(img, imagePath, seedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate seed: %w", err)
	}

	// Set seed in extractor options (except for random mode).
	if seedMode != seed.ModeRandom {
		extractorOpts.Seed = &calculatedSeed
	}

	if verbose {
		if extractorOpts.Seed != nil {
			fmt.Fprintf(os.Stderr, "Using seed mode: %s (seed: %d)\n", commonflags.SeedMode, calculatedSeed)
		} else {
			fmt.Fprintf(os.Stderr, "Using seed mode: %s (non-deterministic)\n", commonflags.SeedMode)
		}
	}

	// Create k-means extractor.
	extractor, err := colour.NewExtractor(colour.AlgorithmKMeans, extractorOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create extractor: %w", err)
	}

	// Extract main palette colors.
	palette, err := extractor.Extract(img, commonflags.Count)
	if err != nil {
		return nil, fmt.Errorf("failed to extract colors: %w", err)
	}

	// If ambience extraction is disabled, return main colors.
	if !commonflags.ExtractAmbience {
		return palette, nil
	}

	// Extract region colors for ambient lighting.
	if verbose {
		fmt.Fprintf(os.Stderr, "Also extracting %d edge/corner regions using %s method\n", commonflags.Regions, commonflags.SampleMethod)
	}

	regionConfig, err := regions.ConfigurationFromInt(commonflags.Regions)
	if err != nil {
		return nil, fmt.Errorf("invalid regions configuration: %w", err)
	}

	sampler := &regions.Sampler{
		SamplePercent: commonflags.SamplePercent,
		Method:        commonflags.SampleMethod,
	}

	regionPalette, err := sampler.Extract(img, regionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to extract region colors: %w", err)
	}

	// Combine main and region colors with weights.
	totalColors := len(palette.Colors) + len(regionPalette.Colors)
	allColors := make([]color.Color, totalColors)
	weights := make([]float64, totalColors)

	mainWeight := MainColorWeightRatio / float64(len(palette.Colors))
	for i, c := range palette.Colors {
		allColors[i] = c
		weights[i] = mainWeight
	}

	regionWeight := RegionWeightFactor / float64(len(regionPalette.Colors))
	for i, c := range regionPalette.Colors {
		allColors[len(palette.Colors)+i] = c
		weights[len(palette.Colors)+i] = regionWeight
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Extracted %d main colors + %d region colors = %d total\n",
			len(palette.Colors), len(regionPalette.Colors), totalColors)
	}

	return colour.NewPaletteWithWeights(allColors, weights), nil
}

// SaveMetadata saves generation metadata as a JSON file alongside the image.
// metadata may be any JSON-serialisable value.
func SaveMetadata(imagePath string, metadata any) error {
	metadataPath := strings.TrimSuffix(imagePath, filepath.Ext(imagePath)) + ".json"

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}
