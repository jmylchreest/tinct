// Package image provides an input plugin for extracting colour palettes from images.
package image

import (
	"context"
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/image"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/regions"
)

const (
	// RegionWeightFactor determines how much weight region colors receive.
	// relative to main palette colors. Region colors get 10% of the total weight
	// to avoid over-representing edge/corner colors in the final palette.
	RegionWeightFactor = 0.1

	// MainColorWeightRatio is the proportion of total weight allocated to.
	// main palette colors when region extraction is enabled (90%).
	MainColorWeightRatio = 0.9
)

// SeedMode determines how the random seed for k-means clustering is generated.
type SeedMode string

const (
	// SeedModeContent generates seed from image content hash (default, deterministic by content).
	SeedModeContent SeedMode = "content"
	// SeedModeFilepath generates seed from absolute file path hash (deterministic by path).
	SeedModeFilepath SeedMode = "filepath"
	// SeedModeManual uses a user-provided seed value.
	SeedModeManual SeedMode = "manual"
	// SeedModeRandom uses non-deterministic random seed (varies each run).
	SeedModeRandom SeedMode = "random"
)

// SeedConfig holds configuration for k-means clustering seed generation.
type SeedConfig struct {
	Mode  SeedMode `json:"mode"`
	Value *int64   `json:"value,omitempty"` // Only used when Mode is SeedModeManual
}

// Plugin implements the input.Plugin interface for image-based colour extraction.
type Plugin struct {
	path      string
	algorithm string
	colours   int

	// Region extraction (ambient lighting).
	extractAmbience bool   // Whether to extract edge/corner regions (default: false)
	regions         int    // Number of regions to extract (4, 8, 12, 16, 0=disabled)
	samplePercent   int    // Percentage of edge to sample
	sampleMethod    string // "average" or "dominant"

	// Seed configuration for k-means clustering.
	seedMode  string // Seed mode: "content", "filepath", "manual", "random"
	seedValue int64  // Seed value (only used when seedMode is "manual")

	// Wallpaper support.
	loadedImagePath string // Stores the actual path to the loaded image (for wallpaper setting)
}

// New creates a new image input plugin with default settings.
func New() *Plugin {
	return &Plugin{
		algorithm:       "kmeans",
		colours:         16,
		extractAmbience: false,
		regions:         8,
		samplePercent:   10,
		sampleMethod:    "average",
		seedMode:        string(SeedModeContent), // Default to content-based seed
		seedValue:       0,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "image"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Extract colour palette from an image file or HTTP(S) URL (optionally includes edge/corner regions for ambient lighting)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&p.path, "image.path", "p", "", "Path to image file or HTTP(S) URL (required)")
	cmd.Flags().StringVarP(&p.algorithm, "image.algorithm", "a", "kmeans", "Extraction algorithm (kmeans)")
	cmd.Flags().IntVarP(&p.colours, "image.colours", "c", 16, "Number of colours to extract (1-256)")

	// Region extraction flags (for ambient lighting).
	cmd.Flags().BoolVar(&p.extractAmbience, "image.extractAmbience", false, "Extract edge/corner colors for ambient lighting (with reduced weight)")
	cmd.Flags().IntVar(&p.regions, "image.regions", 8, "Number of edge/corner regions to extract (4, 8, 12, 16)")
	cmd.Flags().IntVar(&p.samplePercent, "image.sample-size", 10, "Percentage of edge to sample (1-50)")
	cmd.Flags().StringVar(&p.sampleMethod, "image.sample-method", "average", "Sampling method: 'average' or 'dominant'")

	// Seed configuration flags.
	cmd.Flags().StringVar(&p.seedMode, "image.seed-mode", string(SeedModeContent), "K-means seed mode: content, filepath, manual, random")
	cmd.Flags().Int64Var(&p.seedValue, "image.seed-value", 0, "K-means seed value (only used with --image.seed-mode=manual)")
}

// Validate checks if the plugin has all required inputs configured.
func (p *Plugin) Validate() error {
	if p.path == "" {
		return fmt.Errorf("image path or URL is required (use --image.path or -p)")
	}
	if err := image.ValidateImagePath(p.path); err != nil {
		return fmt.Errorf("invalid image path or URL: %w", err)
	}

	// Validate colours.
	if p.colours < 1 || p.colours > 256 {
		return fmt.Errorf("colours must be between 1 and 256, got %d", p.colours)
	}

	// Validate regions (if ambient extraction is enabled).
	if p.extractAmbience {
		if _, err := regions.ConfigurationFromInt(p.regions); err != nil {
			return fmt.Errorf("invalid regions value: %w (use 4, 8, 12, 16)", err)
		}
		// Validate sample percent.
		if p.samplePercent < 1 || p.samplePercent > 50 {
			return fmt.Errorf("sample size must be between 1 and 50, got %d", p.samplePercent)
		}
		// Validate sample method.
		if p.sampleMethod != "average" && p.sampleMethod != "dominant" {
			return fmt.Errorf("sample method must be 'average' or 'dominant', got %s", p.sampleMethod)
		}
	}

	// Validate seed mode.
	validSeedModes := []string{
		string(SeedModeContent),
		string(SeedModeFilepath),
		string(SeedModeManual),
		string(SeedModeRandom),
	}
	valid := slices.Contains(validSeedModes, p.seedMode)
	if !valid {
		return fmt.Errorf("invalid seed mode '%s' (valid: content, filepath, manual, random)", p.seedMode)
	}

	return nil
}

// WallpaperPath returns the path to the source image for wallpaper setting.
// Implements the input.WallpaperProvider interface.
func (p *Plugin) WallpaperPath() string {
	return p.loadedImagePath
}

// Generate creates a raw colour palette by extracting colours from the image.
// Returns only the extracted colors - categorization happens separately.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	// Load the image using SmartLoader (handles both files and URLs).
	loader := image.NewSmartLoader()
	img, err := loader.Load(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Store the loaded image path for wallpaper setting.
	// For local files, use the original path.
	// For URLs, we'll need to handle downloading/caching separately.
	p.loadedImagePath = p.path

	// Calculate seed based on configured mode.
	seed, err := p.calculateSeed(img, p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate seed: %w", err)
	}

	// Extract palette using k-means with deterministic seed.
	// Create the colour extractor with seed configuration.
	extractorOpts := colour.ExtractorOptions{}
	if SeedMode(p.seedMode) != SeedModeRandom {
		// Only set seed if not in random mode.
		extractorOpts.Seed = &seed
	}

	extractor, err := colour.NewExtractor(colour.Algorithm(p.algorithm), extractorOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create extractor: %w", err)
	}

	if opts.Verbose {
		if extractorOpts.Seed != nil {
			fmt.Printf("→ Using seed mode: %s (seed: %d)\n", p.seedMode, seed)
		} else {
			fmt.Printf("→ Using seed mode: %s (non-deterministic)\n", p.seedMode)
		}
	}

	// Extract and return the raw colour palette.
	palette, err := extractor.Extract(img, p.colours)
	if err != nil {
		return nil, fmt.Errorf("failed to extract colours: %w", err)
	}

	// If ambient extraction is enabled, also extract edge/corner colors with reduced weight.
	if !p.extractAmbience {
		return palette, nil
	}

	// Convert regions count to configuration.
	config, err := regions.ConfigurationFromInt(p.regions)
	if err != nil {
		return nil, fmt.Errorf("invalid regions configuration: %w", err)
	}

	// Create region sampler with custom settings.
	sampler := &regions.Sampler{
		SamplePercent: p.samplePercent,
		Method:        p.sampleMethod,
	}

	if opts.Verbose {
		fmt.Printf("→ Also extracting %d edge/corner regions using %s method\n", p.regions, p.sampleMethod)
	}

	// Extract colors from regions.
	regionPalette, err := sampler.Extract(img, config)
	if err != nil {
		return nil, fmt.Errorf("failed to extract region colors: %w", err)
	}

	// Merge the palettes with adjusted weights.
	// Edge colors get reduced weight to avoid over-representing them.
	numMainColors := len(palette.Colors)
	numRegionColors := len(regionPalette.Colors)

	// Calculate weight for region colors using the configured factor.
	if palette.Weights != nil {
		// Calculate average weight per palette color.
		avgPaletteWeight := 1.0 / float64(numMainColors)
		regionWeightPerColor := avgPaletteWeight * RegionWeightFactor / float64(numRegionColors)

		// Create region weights.
		regionWeights := make([]float64, numRegionColors)
		for i := range regionWeights {
			regionWeights[i] = regionWeightPerColor
		}

		// Normalize main palette weights to leave room for region weights.
		totalRegionWeight := regionWeightPerColor * float64(numRegionColors)
		mainWeightMultiplier := (1.0 - totalRegionWeight)

		adjustedMainWeights := make([]float64, numMainColors)
		for i, w := range palette.Weights {
			adjustedMainWeights[i] = w * mainWeightMultiplier
		}

		// Merge colors and weights.
		palette.Colors = append(palette.Colors, regionPalette.Colors...)
		adjustedMainWeights = append(adjustedMainWeights, regionWeights...)
		palette.Weights = adjustedMainWeights
	} else {
		// No weights in main palette, create them.
		// Main colors get equal weight, region colors get reduced weight.
		totalColors := numMainColors + numRegionColors
		mainWeight := MainColorWeightRatio / float64(numMainColors)
		regionWeight := RegionWeightFactor / float64(numRegionColors)

		weights := make([]float64, totalColors)
		for i := range numMainColors {
			weights[i] = mainWeight
		}
		for i := numMainColors; i < totalColors; i++ {
			weights[i] = regionWeight
		}

		palette.Colors = append(palette.Colors, regionPalette.Colors...)
		palette.Weights = weights
	}

	// Adjust role hints indices to account for the merged colors.
	if regionPalette.RoleHints != nil {
		if palette.RoleHints == nil {
			palette.RoleHints = make(map[colour.Role]int)
		}
		offset := numMainColors
		for role, index := range regionPalette.RoleHints {
			palette.RoleHints[role] = index + offset
		}
	}

	return palette, nil
}
