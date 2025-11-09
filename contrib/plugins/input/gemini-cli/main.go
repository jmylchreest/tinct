// gemini-cli - AI Image Generation Input Plugin (Tinct)
//
// Generates images using Google's Gemini CLI with the nanobanana extension,
// then extracts color palettes from the generated images using tinct's k-means algorithm.
//
// Features:
// - Generate images from text prompts using Gemini 2.5 Flash Image model
// - Automatic caching of generated images
// - K-means color extraction (reuses tinct's internal color processing)
// - Deterministic palette generation with seed support
//
// Prerequisites:
//   - gemini-cli (Arch Linux: pacman -S gemini-cli)
//   - nanobanana extension for gemini-cli
//   - Gemini API key configured
//
// Build:
//   go build -o gemini-cli
//
// Usage:
//   tinct plugins add ./gemini-cli --type input
//   tinct plugins enable gemini-cli
//   tinct generate -i gemini-cli -o kitty \
//     --plugin-args 'gemini-cli={"prompt":"sunset over mountains"}'
//
// Plugin Args:
//   prompt: Text description for image generation (required)
//   count: Number of colors to extract (default: 32)
//   extract_ambience: Extract edge/corner colors for ambient lighting (default: false)
//   regions: Number of edge regions to extract (4, 8, 12, 16, default: 8)
//   sample_percent: Percentage of edge to sample (1-50, default: 10)
//   sample_method: Sampling method - "average" or "dominant" (default: "average")
//   seed_mode: K-means seed mode (content, manual, random, default: content)
//   seed_value: Manual seed value (only used when seed_mode=manual)
//   cache: Enable caching (default: true)
//   cache_dir: Cache directory (default: ~/.cache/tinct/gemini-cli)
//   cache_filename: Custom filename for cached image
//   cache_overwrite: Allow overwriting existing cache (default: false)
//
// Author: Tinct Contributors
// License: MIT

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	stdimage "image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/image"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/regions"
	"github.com/jmylchreest/tinct/internal/plugin/protocol"
)

// GeminiCLIPlugin implements the protocol.InputPlugin interface.
type GeminiCLIPlugin struct {
	lastImagePath string // Path to the most recently generated/used image
}

// Config holds the plugin configuration.
type Config struct {
	// Image generation
	Prompt string

	// Color extraction
	ColorCount      int
	ExtractAmbience bool
	Regions         int
	SamplePercent   int
	SampleMethod    string
	SeedMode        string
	SeedValue       int64

	// Caching
	CacheEnabled   bool
	CacheDir       string
	CacheFilename  string
	CacheOverwrite bool
}

// Generate creates an image using Gemini CLI and extracts colors.
func (p *GeminiCLIPlugin) Generate(ctx context.Context, opts protocol.InputOptions) ([]color.Color, error) {
	// Parse configuration from plugin args
	config, err := parseConfig(opts)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Gemini CLI Plugin Configuration:\n")
		fmt.Fprintf(os.Stderr, "  Prompt: %s\n", config.Prompt)
		fmt.Fprintf(os.Stderr, "  Cache: %v (dir: %s)\n", config.CacheEnabled, config.CacheDir)
		fmt.Fprintf(os.Stderr, "  Colors: %d\n", config.ColorCount)
	}

	// Handle dry-run mode
	if opts.DryRun {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "DRY-RUN MODE: Would generate image with prompt: %s\n", config.Prompt)
		}
		return []color.Color{}, nil
	}

	// Generate image path (cached or temporary)
	imagePath, err := getImagePath(config, opts.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to determine image path: %w", err)
	}

	// Generate image if it doesn't exist or cache_overwrite is true
	if config.CacheOverwrite || !fileExists(imagePath) {
		// Check prerequisites only when we need to generate a new image
		if err := checkPrerequisites(opts.Verbose); err != nil {
			return nil, err
		}
		fmt.Fprintf(os.Stderr, "Sending prompt to Gemini: %s\n", config.Prompt)
		fmt.Fprintf(os.Stderr, "Waiting for response...\n")

		if err := generateImage(ctx, config, imagePath, opts.Verbose); err != nil {
			return nil, fmt.Errorf("failed to generate image: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Image returned: %s\n", imagePath)
	} else {
		fmt.Fprintf(os.Stderr, "Using cached image: %s\n", imagePath)
	}

	// Store the image path for wallpaper use
	p.lastImagePath = imagePath

	// Extract colors from the generated image using tinct's k-means
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Extracting %d colors from image using k-means...\n", config.ColorCount)
	}

	colors, err := extractColors(imagePath, config, opts.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to extract colors: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Successfully extracted %d colors\n", len(colors))
	}

	return colors, nil
}

// checkPrerequisites verifies that gemini-cli and nanobanana are available.
func checkPrerequisites(verbose bool) error {
	// Check if gemini binary exists
	geminiPath, err := exec.LookPath("gemini")
	if err != nil {
		return fmt.Errorf("gemini-cli not found in PATH\nInstall with: pacman -S gemini-cli")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Found gemini at: %s\n", geminiPath)
	}

	// Check if nanobanana extension is installed
	cmd := exec.Command("gemini", "extensions", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check gemini extensions: %w\nRun: gemini extensions list", err)
	}

	if !strings.Contains(string(output), "nanobanana") {
		return fmt.Errorf("nanobanana extension not found\nInstall with: gemini extensions install https://github.com/gemini-cli-extensions/nanobanana")
	}

	// Check if nanobanana is enabled
	if !strings.Contains(string(output), "Enabled") || strings.Contains(string(output), "Enabled (User): false") {
		return fmt.Errorf("nanobanana extension is disabled\nEnable with: gemini extensions enable nanobanana")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "nanobanana extension is installed and enabled\n")
	}

	// Check if API key is configured
	apiKeys := []string{"NANOBANANA_GEMINI_API_KEY", "NANOBANANA_GOOGLE_API_KEY", "GEMINI_API_KEY", "GOOGLE_API_KEY"}
	hasAPIKey := false
	for _, key := range apiKeys {
		if os.Getenv(key) != "" {
			hasAPIKey = true
			if verbose {
				fmt.Fprintf(os.Stderr, "Found API key in %s\n", key)
			}
			break
		}
	}

	if !hasAPIKey {
		return fmt.Errorf("No Gemini API key found\nSet one of: NANOBANANA_GEMINI_API_KEY, GEMINI_API_KEY, or GOOGLE_API_KEY\nGet an API key at: https://aistudio.google.com/apikey")
	}

	return nil
}

// parseConfig extracts configuration from plugin arguments.
func parseConfig(opts protocol.InputOptions) (*Config, error) {
	config := &Config{
		ColorCount:      32,
		ExtractAmbience: false,
		Regions:         8,
		SamplePercent:   10,
		SampleMethod:    "average",
		SeedMode:        "content",
		SeedValue:       0,
		CacheEnabled:    true,
		CacheDir:        getDefaultCacheDir(),
		CacheOverwrite:  false,
	}

	// Required: prompt
	if prompt, ok := opts.PluginArgs["prompt"].(string); ok && prompt != "" {
		config.Prompt = prompt
	} else {
		return nil, fmt.Errorf("prompt is required")
	}

	// Optional: count
	if count, ok := opts.PluginArgs["count"].(float64); ok {
		config.ColorCount = int(count)
	}

	// Optional: extract_ambience
	if extractAmbience, ok := opts.PluginArgs["extract_ambience"].(bool); ok {
		config.ExtractAmbience = extractAmbience
	}

	// Optional: regions
	if regionCount, ok := opts.PluginArgs["regions"].(float64); ok {
		config.Regions = int(regionCount)
	}

	// Optional: sample_percent
	if samplePercent, ok := opts.PluginArgs["sample_percent"].(float64); ok {
		config.SamplePercent = int(samplePercent)
	}

	// Optional: sample_method
	if sampleMethod, ok := opts.PluginArgs["sample_method"].(string); ok && sampleMethod != "" {
		config.SampleMethod = sampleMethod
	}

	// Optional: seed_mode
	if seedMode, ok := opts.PluginArgs["seed_mode"].(string); ok && seedMode != "" {
		config.SeedMode = seedMode
	}

	// Optional: seed_value
	if seedValue, ok := opts.PluginArgs["seed_value"].(float64); ok {
		config.SeedValue = int64(seedValue)
	}

	// Optional: cache
	if cache, ok := opts.PluginArgs["cache"].(bool); ok {
		config.CacheEnabled = cache
	}

	// Optional: cache_dir
	if cacheDir, ok := opts.PluginArgs["cache_dir"].(string); ok && cacheDir != "" {
		config.CacheDir = expandPath(cacheDir)
	}

	// Optional: cache_filename
	if cacheFilename, ok := opts.PluginArgs["cache_filename"].(string); ok && cacheFilename != "" {
		config.CacheFilename = cacheFilename
	}

	// Optional: cache_overwrite
	if cacheOverwrite, ok := opts.PluginArgs["cache_overwrite"].(bool); ok {
		config.CacheOverwrite = cacheOverwrite
	}

	return config, nil
}

// getImagePath determines where to save/load the generated image.
func getImagePath(config *Config, _ bool) (string, error) {
	if !config.CacheEnabled {
		// Use temporary file
		tmpFile, err := os.CreateTemp("", "tinct-gemini-*.png")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpFile.Close()
		return tmpFile.Name(), nil
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(config.CacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Determine filename
	filename := config.CacheFilename
	if filename == "" {
		// Generate filename from prompt hash
		hash := sha256.Sum256([]byte(config.Prompt))
		hashStr := hex.EncodeToString(hash[:])[:16]
		filename = fmt.Sprintf("gemini-%s.png", hashStr)
	}

	return filepath.Join(config.CacheDir, filename), nil
}

// generateImage calls Gemini CLI with nanobanana to create an image.
func generateImage(ctx context.Context, config *Config, outputPath string, verbose bool) error {
	// Build the gemini-cli command using /generate from nanobanana
	// Augment the prompt to request 4K landscape wallpaper format
	enhancedPrompt := fmt.Sprintf("%s, 4K resolution, landscape orientation, suitable for desktop wallpaper, wide aspect ratio", config.Prompt)

	// Build command arguments
	args := []string{}
	if verbose {
		args = append(args, "--debug")
	}
	// Use --yolo to automatically accept all actions (non-interactive)
	args = append(args, "--yolo")
	// Add the /generate command with prompt and count as separate arguments
	args = append(args, fmt.Sprintf("/generate %q --count 1", enhancedPrompt))

	if verbose {
		fmt.Fprintf(os.Stderr, "Running: gemini %s\n", strings.Join(args, " "))
	}

	// Use positional prompt with stdin from /dev/null to ensure non-interactive
	cmd := exec.CommandContext(ctx, "gemini", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Open /dev/null for stdin to ensure non-interactive execution
	devNull, err := os.Open("/dev/null")
	if err != nil {
		return fmt.Errorf("failed to open /dev/null: %w", err)
	}
	defer devNull.Close()
	cmd.Stdin = devNull

	// Run the command
	if err := cmd.Run(); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "\nGemini CLI Error Details:\n")
			fmt.Fprintf(os.Stderr, "Command: gemini %s\n", strings.Join(args, " "))
			fmt.Fprintf(os.Stderr, "Exit code: %v\n", err)
			if stderr.Len() > 0 {
				fmt.Fprintf(os.Stderr, "Stderr:\n%s\n", stderr.String())
			}
			if stdout.Len() > 0 {
				fmt.Fprintf(os.Stderr, "Stdout:\n%s\n", stdout.String())
			}
		}
		return fmt.Errorf("gemini-cli failed: %w\nStderr: %s\nStdout: %s", err, stderr.String(), stdout.String())
	}

	// Parse output to find the generated image path
	output := stdout.String() + stderr.String() // nanobanana may output to either

	if verbose {
		fmt.Fprintf(os.Stderr, "\nGemini CLI Output:\n")
		if stdout.Len() > 0 {
			fmt.Fprintf(os.Stderr, "Stdout:\n%s\n", stdout.String())
		}
		if stderr.Len() > 0 {
			fmt.Fprintf(os.Stderr, "Stderr:\n%s\n", stderr.String())
		}
	}

	imagePath := extractImagePathFromOutput(output)

	if imagePath == "" {
		return fmt.Errorf("failed to extract image path from gemini-cli output:\n%s", output)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Gemini generated image at: %s\n", imagePath)
	}

	// Move/copy the generated image to our cache location
	if imagePath != outputPath {
		data, err := os.ReadFile(imagePath)
		if err != nil {
			return fmt.Errorf("failed to read generated image: %w", err)
		}
		if err := os.WriteFile(outputPath, data, 0o600); err != nil {
			return fmt.Errorf("failed to write image to cache: %w", err)
		}
	}

	return nil
}

// extractImagePathFromOutput parses gemini-cli output to find the image path.
func extractImagePathFromOutput(output string) string {
	// First, check the nanobanana-output directory for the most recent image
	// This is more reliable than parsing the text output
	nanobananaDir := "nanobanana-output"
	if dirInfo, err := os.Stat(nanobananaDir); err == nil && dirInfo.IsDir() {
		entries, err := os.ReadDir(nanobananaDir)
		if err == nil {
			var newestPath string
			var newestTime time.Time

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				name := entry.Name()
				// Check if it's an image file
				if !strings.HasSuffix(name, ".png") && !strings.HasSuffix(name, ".jpg") && !strings.HasSuffix(name, ".jpeg") {
					continue
				}

				fullPath := filepath.Join(nanobananaDir, name)
				info, err := entry.Info()
				if err != nil {
					continue
				}

				if newestPath == "" || info.ModTime().After(newestTime) {
					newestPath = fullPath
					newestTime = info.ModTime()
				}
			}

			if newestPath != "" {
				return newestPath
			}
		}
	}

	// Fallback: Look for file paths in the output text
	imagePattern := regexp.MustCompile(`(/\S+\.(?:png|jpg|jpeg))`)
	matches := imagePattern.FindAllString(output, -1)

	// Check each match to see if the file exists
	for _, match := range matches {
		if fileExists(match) {
			return match
		}
	}

	return ""
}

// extractColors uses tinct's internal k-means extractor to extract a color palette.
func extractColors(imagePath string, config *Config, verbose bool) ([]color.Color, error) {
	// Load the image using SmartLoader
	loader := image.NewSmartLoader()
	img, err := loader.Load(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Prepare extractor options
	extractorOpts := colour.ExtractorOptions{}

	// Set seed based on mode
	var seed int64
	switch config.SeedMode {
	case "content":
		// Generate seed from image content
		seed = generateImageContentSeed(img)
		extractorOpts.Seed = &seed
	case "manual":
		// Use user-provided seed
		seed = config.SeedValue
		extractorOpts.Seed = &seed
	case "random":
		// No seed - random each time
		extractorOpts.Seed = nil
	default:
		return nil, fmt.Errorf("invalid seed_mode: %s (valid: content, manual, random)", config.SeedMode)
	}

	if verbose {
		if extractorOpts.Seed != nil {
			fmt.Fprintf(os.Stderr, "Using seed mode: %s (seed: %d)\n", config.SeedMode, seed)
		} else {
			fmt.Fprintf(os.Stderr, "Using seed mode: %s (non-deterministic)\n", config.SeedMode)
		}
	}

	// Create k-means extractor
	extractor, err := colour.NewExtractor(colour.AlgorithmKMeans, extractorOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create extractor: %w", err)
	}

	// Extract main palette colors
	palette, err := extractor.Extract(img, config.ColorCount)
	if err != nil {
		return nil, fmt.Errorf("failed to extract colors: %w", err)
	}

	// If ambience extraction is disabled, return main colors
	if !config.ExtractAmbience {
		return palette.Colors, nil
	}

	// Extract region colors for ambient lighting
	if verbose {
		fmt.Fprintf(os.Stderr, "Also extracting %d edge/corner regions using %s method\n", config.Regions, config.SampleMethod)
	}

	// Convert regions count to configuration
	regionConfig, err := regions.ConfigurationFromInt(config.Regions)
	if err != nil {
		return nil, fmt.Errorf("invalid regions configuration: %w", err)
	}

	// Create region sampler
	sampler := &regions.Sampler{
		SamplePercent: config.SamplePercent,
		Method:        config.SampleMethod,
	}

	// Extract colors from regions
	regionPalette, err := sampler.Extract(img, regionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to extract region colors: %w", err)
	}

	// Combine main and region colors
	// Region colors get reduced weight (10% of total weight)
	allColors := make([]color.Color, 0, len(palette.Colors)+len(regionPalette.Colors))
	allColors = append(allColors, palette.Colors...)
	allColors = append(allColors, regionPalette.Colors...)

	if verbose {
		fmt.Fprintf(os.Stderr, "Extracted %d main colors + %d region colors = %d total\n",
			len(palette.Colors), len(regionPalette.Colors), len(allColors))
	}

	return allColors, nil
}

// generateImageContentSeed generates a seed from image content for deterministic extraction.
func generateImageContentSeed(img stdimage.Image) int64 {
	bounds := img.Bounds()
	hash := sha256.New()

	// Sample pixels from the image to generate a content hash
	step := 10 // Sample every 10th pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, a := img.At(x, y).RGBA()
			hash.Write([]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8), byte(a >> 8)})
		}
	}

	sum := hash.Sum(nil)
	var seed int64
	for i := 0; i < 8 && i < len(sum); i++ {
		seed = (seed << 8) | int64(sum[i])
	}
	return seed
}

// GetMetadata returns plugin metadata.
func (p *GeminiCLIPlugin) GetMetadata() protocol.PluginInfo {
	return protocol.PluginInfo{
		Name:            "gemini-cli",
		Type:            "input",
		Version:         "0.0.1",
		ProtocolVersion: protocol.ProtocolVersion,
		Description:     "Generate images using Google Gemini CLI and extract color palettes",
		PluginProtocol:  "go-plugin",
	}
}

// WallpaperPath returns the path to the generated image for wallpaper use.
// This implements the WallpaperProvider interface.
func (p *GeminiCLIPlugin) WallpaperPath() string {
	return p.lastImagePath
}

// Helper functions

func getDefaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cache/tinct/gemini-cli"
	}
	return filepath.Join(home, ".cache", "tinct", "gemini-cli")
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func main() {
	// Handle --plugin-info flag
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		p := &GeminiCLIPlugin{}
		info := p.GetMetadata()

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding plugin info: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Serve the plugin using go-plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: protocol.Handshake,
		Plugins: map[string]plugin.Plugin{
			"input": &protocol.InputPluginRPC{
				Impl: &GeminiCLIPlugin{},
			},
		},
	})
}
