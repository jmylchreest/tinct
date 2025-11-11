// Package image provides an input plugin for extracting colour palettes from images.
package image

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/image"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/regions"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/seed"
	"github.com/jmylchreest/tinct/internal/util/imagecache"
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

// Note: SeedMode, SeedConfig, and seed calculation functions have been moved to
// internal/plugin/input/shared/seed package for reuse by other image-processing plugins.

// Plugin implements the input.Plugin interface for image-based colour extraction.
type Plugin struct {
	path    string
	colours int

	// Region extraction (ambient lighting).
	extractAmbience bool   // Whether to extract edge/corner regions (default: false)
	regions         int    // Number of regions to extract (4, 8, 12, 16, 0=disabled)
	samplePercent   int    // Percentage of edge to sample
	sampleMethod    string // "average" or "dominant"

	// Seed configuration for k-means clustering.
	seedMode  string // Seed mode: "content", "filepath", "manual", "random"
	seedValue int64  // Seed value (only used when seedMode is "manual")

	// Remote image caching (for wallpaper support).
	cacheEnabled   bool   // Enable caching of remote images (default: false)
	cacheDir       string // Directory to cache downloaded images
	cacheFilename  string // Filename for cached image (empty = auto-generate from URL hash)
	cacheOverwrite bool   // Allow overwriting existing cached images

	// Wallpaper support.
	loadedImagePath string // Stores the actual path to the loaded image (for wallpaper setting)
}

// New creates a new image input plugin with default settings.
func New() *Plugin {
	// Check environment variables for cache settings.
	cacheEnabled := false
	if val := os.Getenv("TINCT_IMAGE_CACHE"); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			cacheEnabled = parsed
		}
	}

	cacheDir := os.Getenv("TINCT_IMAGE_CACHE_DIR")
	cacheFilename := os.Getenv("TINCT_IMAGE_CACHE_FILENAME")

	cacheOverwrite := false
	if val := os.Getenv("TINCT_IMAGE_CACHE_OVERWRITE"); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			cacheOverwrite = parsed
		}
	}

	return &Plugin{
		colours:         16,
		extractAmbience: false,
		regions:         8,
		samplePercent:   10,
		sampleMethod:    "average",
		seedMode:        string(seed.ModeContent), // Default to content-based seed
		seedValue:       0,
		cacheEnabled:    cacheEnabled,
		cacheDir:        cacheDir,
		cacheFilename:   cacheFilename,
		cacheOverwrite:  cacheOverwrite,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "image"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Extract colour palette from an image file, directory (random selection), or HTTP(S) URL (optionally includes edge/corner regions for ambient lighting)"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&p.path, "image.path", "p", "", "Path to image file, directory, or HTTP(S) URL (required, directories will select a random image)")
	cmd.Flags().IntVarP(&p.colours, "image.colours", "c", 16, "Number of colours to extract (1-256)")

	// Region extraction flags (for ambient lighting).
	cmd.Flags().BoolVar(&p.extractAmbience, "image.extractAmbience", false, "Extract edge/corner colors for ambient lighting (with reduced weight)")
	cmd.Flags().IntVar(&p.regions, "image.regions", 8, "Number of edge/corner regions to extract (4, 8, 12, 16)")
	cmd.Flags().IntVar(&p.samplePercent, "image.sample-size", 10, "Percentage of edge to sample (1-50)")
	cmd.Flags().StringVar(&p.sampleMethod, "image.sample-method", "average", "Sampling method: 'average' or 'dominant'")

	// Seed configuration flags.
	cmd.Flags().StringVar(&p.seedMode, "image.seed-mode", string(seed.ModeContent), "K-means seed mode: content, filepath, manual, random")
	cmd.Flags().Int64Var(&p.seedValue, "image.seed-value", 0, "K-means seed value (only used with --image.seed-mode=manual)")

	// Remote image caching flags (use struct values as defaults, which may come from env vars).
	cmd.Flags().BoolVar(&p.cacheEnabled, "image.cache", p.cacheEnabled, "Enable caching of remote images for wallpaper support")
	cmd.Flags().StringVar(&p.cacheDir, "image.cache-dir", p.cacheDir, "Directory to cache downloaded images (default: ~/.cache/tinct/images)")
	cmd.Flags().StringVar(&p.cacheFilename, "image.cache-filename", p.cacheFilename, "Filename for cached image (default: auto-generated from URL hash)")
	cmd.Flags().BoolVar(&p.cacheOverwrite, "image.cache-overwrite", p.cacheOverwrite, "Allow overwriting existing cached images")
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
		string(seed.ModeContent),
		string(seed.ModeFilepath),
		string(seed.ModeManual),
		string(seed.ModeRandom),
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

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		{Name: "image.path", Shorthand: "p", Type: "string", Default: "", Description: "Path to image file, directory, or HTTP(S) URL (required)", Required: true},
		{Name: "image.colours", Shorthand: "c", Type: "int", Default: "16", Description: "Number of colours to extract (1-256)", Required: false},
		{Name: "image.extractAmbience", Type: "bool", Default: "false", Description: "Extract edge/corner colors for ambient lighting", Required: false},
		{Name: "image.regions", Type: "int", Default: "8", Description: "Number of edge/corner regions (4, 8, 12, 16)", Required: false},
		{Name: "image.sample-size", Type: "int", Default: "10", Description: "Percentage of edge to sample (1-50)", Required: false},
		{Name: "image.sample-method", Type: "string", Default: "average", Description: "Sampling method: 'average' or 'dominant'", Required: false},
		{Name: "image.seed-mode", Type: "string", Default: "content", Description: "K-means seed mode: content, filepath, manual, random", Required: false},
		{Name: "image.seed-value", Type: "int64", Default: "0", Description: "K-means seed value (only used with --image.seed-mode=manual)", Required: false},
		{Name: "image.cache", Type: "bool", Default: fmt.Sprintf("%v", p.cacheEnabled), Description: "Enable caching of remote images", Required: false},
		{Name: "image.cache-dir", Type: "string", Default: p.cacheDir, Description: "Directory to cache downloaded images", Required: false},
		{Name: "image.cache-filename", Type: "string", Default: p.cacheFilename, Description: "Filename for cached image (auto-generated if empty)", Required: false},
		{Name: "image.cache-overwrite", Type: "bool", Default: fmt.Sprintf("%v", p.cacheOverwrite), Description: "Allow overwriting existing cached images", Required: false},
	}
}

// Generate creates a raw colour palette by extracting colours from the image.
// Returns only the extracted colors - categorization happens separately.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	// Validate the backend first before doing any expensive operations.
	if opts.Backend != "kmeans" {
		return nil, fmt.Errorf("invalid backend: %s (only kmeans is currently supported)", opts.Backend)
	}

	// Resolve the path - if it's a directory, select a random image.
	resolvedPath, err := image.ResolveImagePath(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve image path: %w", err)
	}

	// If a random image was selected from a directory, log it.
	if opts.Verbose && resolvedPath != p.path {
		fmt.Printf("→ Selected random image from directory: %s\n", resolvedPath)
	}

	// For remote images (HTTP/HTTPS), optionally download and cache them for wallpaper support.
	wallpaperPath := resolvedPath
	isRemoteImage := strings.HasPrefix(resolvedPath, "http://") || strings.HasPrefix(resolvedPath, "https://")

	if isRemoteImage && p.cacheEnabled {
		cacheOpts := imagecache.CacheOptions{
			CacheDir:       p.cacheDir,
			Filename:       p.cacheFilename,
			AllowOverwrite: p.cacheOverwrite,
		}

		cachedPath, err := imagecache.DownloadAndCache(ctx, resolvedPath, cacheOpts)
		if err != nil {
			if opts.Verbose {
				fmt.Printf("   Warning: Failed to cache remote image for wallpaper: %v\n", err)
				fmt.Printf("   Color extraction will continue, but wallpaper path will not be available\n")
			}
			// Continue with color extraction even if caching fails.
			// The wallpaperPath will remain as the URL.
		} else {
			// Successfully cached - use local path for wallpaper.
			wallpaperPath = cachedPath
			if opts.Verbose {
				fmt.Printf("→ Cached remote image to: %s\n", cachedPath)
			}
		}
	}

	// Load the image using SmartLoader (handles both files and URLs).
	loader := image.NewSmartLoader()
	img, err := loader.Load(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Store the wallpaper path (local file for remote images, original path otherwise).
	p.loadedImagePath = wallpaperPath

	// Calculate seed based on configured mode using shared utility.
	// Use the resolved path for filepath-based seeds.
	seedMode, err := seed.ParseMode(p.seedMode)
	if err != nil {
		return nil, fmt.Errorf("invalid seed mode: %w", err)
	}

	seedConfig := seed.Config{
		Mode:  seedMode,
		Value: nil,
	}
	if seedMode == seed.ModeManual {
		seedConfig.Value = &p.seedValue
	}

	calculatedSeed, err := seed.Calculate(img, resolvedPath, seedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate seed: %w", err)
	}

	// Extract palette using k-means with deterministic seed.
	// Create the colour extractor with seed configuration.
	extractorOpts := colour.ExtractorOptions{}
	if seedMode != seed.ModeRandom {
		// Only set seed if not in random mode.
		extractorOpts.Seed = &calculatedSeed
	}

	extractor, err := colour.NewExtractor(colour.Algorithm(opts.Backend), extractorOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create extractor: %w", err)
	}

	if opts.Verbose {
		if extractorOpts.Seed != nil {
			fmt.Printf("→ Using seed mode: %s (seed: %d)\n", p.seedMode, calculatedSeed)
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
