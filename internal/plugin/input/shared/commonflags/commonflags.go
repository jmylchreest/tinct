// Package commonflags provides shared unprefixed flags for input plugins.
// These flags are common across multiple input plugins and use no prefix.
package commonflags

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Shared common flag variables.
var (
	// Count is the number of colors to extract.
	Count int

	// AspectRatio is the image aspect ratio.
	AspectRatio string

	// Cache flags
	CacheEnabled   bool
	CacheDir       string
	CacheFilename  string
	CacheOverwrite bool

	// Region extraction flags
	ExtractAmbience bool
	Regions         int
	SamplePercent   int
	SampleMethod    string

	// Seed flags
	SeedMode  string
	SeedValue int64

	// defaultCacheDir stores the default cache directory.
	defaultCacheDir string
)

func init() {
	// Set default cache directory
	home, err := os.UserHomeDir()
	if err == nil {
		defaultCacheDir = filepath.Join(home, ".cache", "tinct", "generated")
	} else {
		defaultCacheDir = ".cache/tinct/generated"
	}
}

// RegisterFlags registers the shared common flags on the given command.
// This should be called by plugins that need these flags.
// It checks if flags already exist to avoid duplicate registration.
func RegisterFlags(cmd *cobra.Command) {
	// Color extraction
	if cmd.Flags().Lookup("count") == nil {
		cmd.Flags().IntVar(&Count, "count", 32, "Number of colors to extract")
	}

	// Aspect ratio
	if cmd.Flags().Lookup("aspect-ratio") == nil {
		cmd.Flags().StringVar(&AspectRatio, "aspect-ratio", "16:9", "Image aspect ratio (1:1, 3:4, 4:3, 9:16, 16:9, 21:9)")
	}

	// Cache flags
	if cmd.Flags().Lookup("cache") == nil {
		cmd.Flags().BoolVar(&CacheEnabled, "cache", true, "Enable image caching")
	}
	if cmd.Flags().Lookup("cache-dir") == nil {
		cmd.Flags().StringVar(&CacheDir, "cache-dir", defaultCacheDir, "Cache directory for generated images")
	}
	if cmd.Flags().Lookup("cache-filename") == nil {
		cmd.Flags().StringVar(&CacheFilename, "cache-filename", "", "Custom cache filename (auto-generated if empty)")
	}
	if cmd.Flags().Lookup("cache-overwrite") == nil {
		cmd.Flags().BoolVar(&CacheOverwrite, "cache-overwrite", false, "Overwrite existing cached images")
	}

	// Region extraction flags
	if cmd.Flags().Lookup("extract-ambience") == nil {
		cmd.Flags().BoolVar(&ExtractAmbience, "extract-ambience", false, "Extract edge/corner colors for ambient lighting")
	}
	if cmd.Flags().Lookup("regions") == nil {
		cmd.Flags().IntVar(&Regions, "regions", 8, "Number of edge regions (4, 8, 12, 16)")
	}
	if cmd.Flags().Lookup("sample-percent") == nil {
		cmd.Flags().IntVar(&SamplePercent, "sample-percent", 10, "Percentage of edge to sample (1-50)")
	}
	if cmd.Flags().Lookup("sample-method") == nil {
		cmd.Flags().StringVar(&SampleMethod, "sample-method", "average", "Sampling method (average or dominant)")
	}

	// Seed flags
	if cmd.Flags().Lookup("seed-mode") == nil {
		cmd.Flags().StringVar(&SeedMode, "seed-mode", "content", "Seed mode for deterministic extraction (content, manual, random)")
	}
	if cmd.Flags().Lookup("seed-value") == nil {
		cmd.Flags().Int64Var(&SeedValue, "seed-value", 0, "Manual seed value (used with --seed-mode=manual)")
	}
}

// Reset resets all flag values to defaults. Useful for testing.
func Reset() {
	Count = 32
	AspectRatio = "16:9"
	CacheEnabled = true
	CacheDir = defaultCacheDir
	CacheFilename = ""
	CacheOverwrite = false
	ExtractAmbience = false
	Regions = 8
	SamplePercent = 10
	SampleMethod = "average"
	SeedMode = "content"
	SeedValue = 0
}

// GetDefaultCacheDir returns the default cache directory.
func GetDefaultCacheDir() string {
	return defaultCacheDir
}
