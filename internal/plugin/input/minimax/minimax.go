// Package minimax provides an input plugin for generating images using MiniMax's
// (Hailuo) image generation API.
package minimax

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	_ "image/jpeg" // Required for JPEG image decoding
	_ "image/png"  // Required for PNG image decoding
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiprompt"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/imagecolours"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/imagepost"
	"github.com/jmylchreest/tinct/internal/version"
)

// extJPG is the fallback extension for a generated image whose format the
// provider did not state, and the name under which such files are cached.
const extJPG = ".jpg"

const (
	// defaultAPIURL is the MiniMax text-to-image endpoint. It is not
	// OpenAI-compatible; it has its own request/response shape.
	defaultAPIURL = "https://api.minimax.io/v1/image_generation"

	// defaultModel is the model used when none is specified.
	defaultModel = "image-01"

	// defaultImageSize is the target long-edge resolution. MiniMax defaults to
	// 720p when only an aspect ratio is given, so we request the maximum the
	// model supports for crisp wallpapers.
	defaultImageSize = "2K"

	// minImageDimension and maxImageDimension are MiniMax's accepted width/height
	// bounds (and each must be divisible by 8).
	minImageDimension = 512
	maxImageDimension = 2048
)

// ImageMetadata contains metadata about a generated image.
type ImageMetadata struct {
	// Generation parameters
	Prompt         string `json:"prompt"`
	EnhancedPrompt string `json:"enhanced_prompt,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Model          string `json:"model"`
	AspectRatio    string `json:"aspect_ratio"`

	// Generation details
	CreatedAt  time.Time `json:"created_at"`
	ImagePath  string    `json:"image_path"`
	ImageSize_ int       `json:"image_bytes"`
}

// Plugin implements the input.Plugin interface for MiniMax image generation.
type Plugin struct {
	// Plugin-specific flags (not shared)
	apiURL    string
	imageSize string

	// Wallpaper support
	loadedImagePath string
}

// New creates a new MiniMax input plugin with default settings.
func New() *Plugin {
	return &Plugin{
		apiURL:    defaultAPIURL,
		imageSize: defaultImageSize,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "minimax"
}

// Description returns a human-readable description.
func (p *Plugin) Description() string {
	return "Generate images with MiniMax (Hailuo) and extract colours"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	// Register shared AI flags (prompt, model, list-models, etc.)
	aiflags.RegisterFlags(cmd)

	// Register shared common flags (count, aspect-ratio, cache, regions, seed, etc.)
	commonflags.RegisterFlags(cmd)

	// Register plugin-specific flags with minimax prefix
	if cmd.Flags().Lookup("minimax.base-url") == nil {
		cmd.Flags().StringVar(&p.apiURL, "minimax.base-url", p.apiURL, "MiniMax image generation endpoint URL (override for proxies)")
	}
	if cmd.Flags().Lookup("minimax.image-size") == nil {
		cmd.Flags().StringVar(&p.imageSize, "minimax.image-size", p.imageSize, "Target long-edge resolution: 1K, 2K, or pixels (512-2048)")
	}
}

// Validate checks if required inputs are configured.
func (p *Plugin) Validate() error {
	// Skip validation if just listing models
	if aiflags.ListModels {
		return nil
	}
	if aiflags.Prompt == "" {
		return fmt.Errorf("--ai.prompt is required")
	}
	return nil
}

// Generate creates an image using MiniMax and extracts colors.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	// If list-models flag is set, print the maintained catalogue and exit.
	// Listing does not call the API: MiniMax exposes no image-model listing
	// endpoint and no pricing, so we maintain the list (with prices) here.
	if aiflags.ListModels {
		ListModels()
		os.Exit(0)
	}

	// Determine model to use
	model := aiflags.Model
	if model == "" || model == "auto" {
		model = defaultModel
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "MiniMax Plugin Configuration:\n")
		fmt.Fprintf(os.Stderr, "  Prompt: %s\n", aiflags.Prompt)
		fmt.Fprintf(os.Stderr, "  Model: %s\n", model)
		fmt.Fprintf(os.Stderr, "  Aspect Ratio: %s\n", commonflags.AspectRatio)
		fmt.Fprintf(os.Stderr, "  Cache: %v (dir: %s)\n", commonflags.CacheEnabled, commonflags.CacheDir)
		fmt.Fprintf(os.Stderr, "  Colors: %d\n", commonflags.Count)
	}

	if opts.DryRun {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "DRY-RUN MODE: Would generate image with prompt: %s\n", aiflags.Prompt)
		}
		return colour.NewPalette([]color.Color{}), nil
	}

	// Determine base image path (without extension); the actual extension is
	// determined from the image format in the API response.
	imageBasePath, err := p.getImageBasePath(model)
	if err != nil {
		return nil, fmt.Errorf("failed to determine image path: %w", err)
	}

	// Check for existing cached image (any supported extension)
	var imagePath string
	if !commonflags.CacheOverwrite {
		imagePath = findCachedImage(imageBasePath)
	}

	// Generate image if not cached
	if imagePath == "" {
		enhancedPrompt := aiprompt.EnhanceForWallpaper(aiflags.Prompt)
		additionalPrompt := enhancedPrompt[len(aiflags.Prompt):]
		fmt.Fprintf(os.Stderr, "[minimax] model=%s prompt=\"%s\" additional=\"%s\"\n",
			model, aiflags.Prompt, additionalPrompt)
		fmt.Fprintf(os.Stderr, "Waiting for response...\n")

		imagePath, err = p.generateImage(ctx, model, imageBasePath, opts.Verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to generate image: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Image generated: %s\n", imagePath)
		imagepost.TrimLetterboxIfEnabled(imagePath, commonflags.AspectRatio, commonflags.TrimLetterbox, opts.Verbose)
	} else {
		fmt.Fprintf(os.Stderr, "Using cached image: %s\n", imagePath)
	}

	// Store path for wallpaper support
	p.loadedImagePath = imagePath

	// Extract colors
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Extracting %d colors from image using k-means...\n", commonflags.Count)
	}

	palette, err := p.extractColors(imagePath, opts.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to extract colors: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Successfully extracted %d colors\n", len(palette.Colors))
	}

	return palette, nil
}

// WallpaperPath returns the canonical path to the generated image for wallpaper use.
// For AI-generated images, this is always an absolute path to the cached image.
// Implements the input.WallpaperProvider interface.
func (p *Plugin) WallpaperPath() string {
	return p.loadedImagePath
}

// WallpaperRawPath returns the literal path as provided by the user.
// For AI-generated images, this is the same as WallpaperPath since there's
// no user-provided input path - the path is generated based on the prompt hash.
// Implements the input.WallpaperProvider interface.
func (p *Plugin) WallpaperRawPath() string {
	return p.loadedImagePath
}

// getImageBasePath determines the base path (without extension) for the generated image.
// The actual extension is determined after we know the image format from the API response.
func (p *Plugin) getImageBasePath(model string) (string, error) {
	if !commonflags.CacheEnabled {
		// For temp files, create without extension - we'll rename after
		tmpFile, err := os.CreateTemp("", "tinct-minimax-*")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		_ = tmpFile.Close() // File will be rewritten; close error is not actionable
		return tmpFile.Name(), nil
	}

	// Use plugin-specific subdirectory
	cacheDir := filepath.Join(commonflags.CacheDir, "minimax")
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	filename := commonflags.CacheFilename
	if filename == "" {
		hash := sha256.Sum256([]byte(aiflags.Prompt + model))
		hashStr := hex.EncodeToString(hash[:])[:16]
		// Return without extension - will be added based on actual image format
		filename = fmt.Sprintf("minimax-%s", hashStr)
	} else {
		// If user provided a filename, strip any extension - we'll add the correct one
		if ext := filepath.Ext(filename); ext != "" {
			filename = strings.TrimSuffix(filename, ext)
		}
	}

	return filepath.Join(cacheDir, filename), nil
}

// getExtensionForMIME returns the file extension for an image MIME type.
func getExtensionForMIME(mimeType string) string {
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return extJPG
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		// Default to jpg: MiniMax returns JPEG.
		return extJPG
	}
}

// findCachedImage looks for an existing cached image with any supported extension.
// Returns the full path if found, empty string if not found.
func findCachedImage(basePath string) string {
	extensions := []string{".png", ".jpg", ".jpeg", ".webp", ".gif"}
	for _, ext := range extensions {
		path := basePath + ext
		if fileExists(path) {
			return path
		}
	}
	return ""
}

// getAPIKey retrieves the MiniMax API key from the environment.
func getAPIKey() (string, error) {
	apiKey := os.Getenv("MINIMAX_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("MINIMAX_API_KEY environment variable is required\nGet one at: https://platform.minimax.io/")
	}
	return apiKey, nil
}

// imageRequest is the MiniMax text-to-image request body. Width/Height take
// precedence over AspectRatio when set, so we send explicit dimensions to pin
// the resolution and omit the aspect ratio.
type imageRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	AspectRatio    string `json:"aspect_ratio,omitempty"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
	ResponseFormat string `json:"response_format"`
	N              int    `json:"n"`
}

// resolveDimensions converts an aspect ratio ("W:H") and a size flag ("1K", "2K",
// or a pixel count) into explicit width/height for the request, with the long edge
// at the target and both dimensions clamped to [512, 2048] and rounded to a
// multiple of 8. It returns ok=false when the inputs can't be parsed, in which case
// the caller falls back to sending the aspect ratio and letting MiniMax pick a size.
func resolveDimensions(aspectRatio, size string) (width, height int, ok bool) {
	target := parseImageSize(size)
	if target == 0 {
		return 0, 0, false
	}

	w, h, ok := parseAspectRatio(aspectRatio)
	if !ok {
		return 0, 0, false
	}

	if w >= h {
		width = target
		height = int(float64(target) * float64(h) / float64(w))
	} else {
		height = target
		width = int(float64(target) * float64(w) / float64(h))
	}

	return clampDimension(width), clampDimension(height), true
}

// parseImageSize maps a size flag to a target long-edge pixel count, clamped to
// MiniMax's supported range. It returns 0 for an unrecognised value.
func parseImageSize(size string) int {
	switch strings.ToLower(strings.TrimSpace(size)) {
	case "1k":
		return 1024
	case "2k", "":
		return maxImageDimension
	default:
		px, err := strconv.Atoi(strings.TrimSpace(size))
		if err != nil {
			return 0
		}
		return clampDimension(px)
	}
}

// parseAspectRatio parses a "W:H" ratio into its integer components.
func parseAspectRatio(ratio string) (w, h int, ok bool) {
	parts := strings.SplitN(strings.TrimSpace(ratio), ":", 2)
	if len(parts) != 2 {
		return 0, 0, false
	}
	w, errW := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, errH := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errW != nil || errH != nil || w <= 0 || h <= 0 {
		return 0, 0, false
	}
	return w, h, true
}

// clampDimension clamps a pixel dimension to MiniMax's [512, 2048] range and rounds
// it down to the nearest multiple of 8 (required by the API).
func clampDimension(px int) int {
	if px > maxImageDimension {
		px = maxImageDimension
	}
	px -= px % 8
	if px < minImageDimension {
		px = minImageDimension
	}
	return px
}

// imageResponse is the MiniMax text-to-image response. MiniMax reports errors
// through base_resp.status_code even on an HTTP 200, so it must be checked.
type imageResponse struct {
	Data struct {
		ImageBase64 []string `json:"image_base64"`
		ImageURLs   []string `json:"image_urls"`
	} `json:"data"`
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

// generateImage calls the MiniMax API to create an image.
// outputBasePath is the path without extension - the actual extension is determined
// from the API response and the final path is returned.
func (p *Plugin) generateImage(ctx context.Context, model, outputBasePath string, verbose bool) (actualPath string, err error) { //nolint:gocyclo
	apiKey, err := getAPIKey()
	if err != nil {
		return "", err
	}

	// Build the prompt, appending the negative prompt as an instruction (the basic
	// MiniMax T2I API has no dedicated negative-prompt field).
	enhancedPrompt := aiprompt.EnhanceForWallpaper(aiflags.Prompt)
	negativePrompt := aiprompt.BuildNegative(aiflags.NegativePrompt, aiflags.NoNegativePrompt)
	fullPrompt := enhancedPrompt
	if negativePrompt != "" {
		fullPrompt = fmt.Sprintf("%s\n\nAvoid: %s", enhancedPrompt, negativePrompt)
	}

	request := imageRequest{
		Model:          model,
		Prompt:         fullPrompt,
		ResponseFormat: "base64",
		N:              1,
	}

	// Prefer explicit dimensions (pins resolution at the requested aspect ratio);
	// fall back to letting MiniMax size the image from the aspect ratio alone.
	width, height, ok := resolveDimensions(commonflags.AspectRatio, p.imageSize)
	if ok {
		request.Width = width
		request.Height = height
	} else {
		request.AspectRatio = commonflags.AspectRatio
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Calling MiniMax API with model: %s\n", model)
		fmt.Fprintf(os.Stderr, "  Aspect ratio: %s\n", commonflags.AspectRatio)
		if ok {
			fmt.Fprintf(os.Stderr, "  Resolution: %dx%d\n", width, height)
		}
		fmt.Fprintf(os.Stderr, "  Enhanced prompt: %s\n", enhancedPrompt)
		if negativePrompt != "" {
			fmt.Fprintf(os.Stderr, "  Negative prompt: %s\n", negativePrompt)
		}
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req) // #nosec G704 -- URL is the MiniMax API endpoint
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Response status: %d\n", resp.StatusCode) // #nosec G705 -- status code is an integer
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response imageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// MiniMax signals failures via base_resp even on HTTP 200.
	if response.BaseResp.StatusCode != 0 {
		return "", fmt.Errorf("API error %d: %s", response.BaseResp.StatusCode, response.BaseResp.StatusMsg)
	}

	imageBytes, err := p.extractImageBytes(ctx, response)
	if err != nil {
		return "", err
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Received image data: %d bytes\n", len(imageBytes))
	}

	// Determine format from the bytes themselves (MiniMax does not return a MIME type).
	mimeType := http.DetectContentType(imageBytes)
	ext := getExtensionForMIME(mimeType)
	actualPath = outputBasePath + ext

	if verbose {
		fmt.Fprintf(os.Stderr, "Image format: %s (extension: %s)\n", mimeType, ext)
	}

	if err := os.WriteFile(actualPath, imageBytes, 0o600); err != nil {
		return "", fmt.Errorf("failed to write image to file: %w", err)
	}

	metadata := &ImageMetadata{
		Prompt:         aiflags.Prompt,
		EnhancedPrompt: enhancedPrompt,
		NegativePrompt: negativePrompt,
		Model:          model,
		AspectRatio:    commonflags.AspectRatio,
		CreatedAt:      time.Now(),
		ImagePath:      actualPath,
		ImageSize_:     len(imageBytes),
	}
	if err := p.saveMetadata(actualPath, metadata); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to save metadata: %v\n", err)
	}

	return actualPath, nil
}

// extractImageBytes pulls the raw image bytes from a MiniMax response, preferring
// inline base64 and falling back to a (short-lived) image URL.
func (p *Plugin) extractImageBytes(ctx context.Context, response imageResponse) ([]byte, error) {
	if len(response.Data.ImageBase64) > 0 {
		imageBytes, err := base64.StdEncoding.DecodeString(response.Data.ImageBase64[0])
		if err != nil {
			return nil, fmt.Errorf("failed to decode image data: %w", err)
		}
		return imageBytes, nil
	}

	if len(response.Data.ImageURLs) > 0 {
		return fetchImageBytes(ctx, response.Data.ImageURLs[0])
	}

	return nil, fmt.Errorf("no image returned in response")
}

// fetchImageBytes downloads image bytes from a URL (used when MiniMax returns a URL
// instead of inline base64; these URLs expire 24h after generation).
func fetchImageBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create image request: %w", err)
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	resp, err := client.Do(req) // #nosec G704 -- URL returned by the MiniMax API
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image fetch returned status %d", resp.StatusCode) // #nosec G705
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	return imageBytes, nil
}

// saveMetadata saves generation metadata as a JSON file alongside the image.
func (p *Plugin) saveMetadata(imagePath string, metadata *ImageMetadata) error {
	return imagecolours.SaveMetadata(imagePath, metadata)
}

// extractColors extracts colors from the generated image.
func (p *Plugin) extractColors(imagePath string, verbose bool) (*colour.Palette, error) {
	return imagecolours.ExtractColors(imagePath, verbose)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ListModels prints the maintained MiniMax image-generation model catalogue to stdout.
// MiniMax has no models-list API for images and does not expose pricing, so the list
// (with approximate prices) is maintained here.
func ListModels() {
	models := []struct {
		ID          string
		Name        string
		Description string
		Cost        string
	}{
		{
			ID:          "image-01",
			Name:        "MiniMax Image 01 (Hailuo)",
			Description: "Text-to-image generation with strong prompt adherence (default)",
			Cost:        "~$0.0035 per image",
		},
	}

	fmt.Println("Available Image Generation Models:")
	fmt.Println()
	fmt.Printf("Default Model: %s\n", defaultModel)
	fmt.Println()
	fmt.Println("Note: Pricing shown is for indication only and may not reflect current rates.")
	fmt.Println("      Visit https://platform.minimax.io/ for up-to-date pricing.")
	fmt.Println()

	for _, model := range models {
		fmt.Printf("ID: %s\n", model.ID)
		fmt.Printf("  Name: %s\n", model.Name)
		fmt.Printf("  Description: %s\n", model.Description)
		fmt.Printf("  Cost: %s (approximate)\n", model.Cost)
		fmt.Println()
	}

	fmt.Println("You can use any valid MiniMax image model ID with --ai.model.")
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		// Shared AI flags
		{Name: "ai.prompt", Type: "string", Default: "", Description: "Text description for AI image generation (required)", Required: true},
		{Name: "ai.model", Type: "string", Default: defaultModel, Description: "AI model to use for image generation", Required: false},
		{Name: "ai.list-models", Type: "bool", Default: "false", Description: "List available AI models and exit", Required: false},
		{Name: "ai.no-extended-prompt", Type: "bool", Default: "false", Description: "Disable automatic wallpaper prompt enhancements", Required: false},
		{Name: "ai.no-negative-prompt", Type: "bool", Default: "false", Description: "Disable default negative prompt", Required: false},
		{Name: "ai.negative-prompt", Type: "string", Default: "", Description: "Custom negative prompt to discourage certain elements", Required: false},
		// Shared common flags
		{Name: "count", Type: "int", Default: "32", Description: "Number of colors to extract", Required: false},
		{Name: "aspect-ratio", Type: "string", Default: "16:9", Description: "Image aspect ratio (1:1, 3:4, 4:3, 9:16, 16:9, 21:9)", Required: false},
		{Name: "cache", Type: "bool", Default: "true", Description: "Enable image caching", Required: false},
		{Name: "cache-dir", Type: "string", Default: commonflags.GetDefaultCacheDir(), Description: "Cache directory", Required: false},
		{Name: "cache-filename", Type: "string", Default: "", Description: "Custom cache filename", Required: false},
		{Name: "cache-overwrite", Type: "bool", Default: "false", Description: "Overwrite existing cache", Required: false},
		{Name: "extract-ambience", Type: "bool", Default: "false", Description: "Extract edge/corner colors for ambient lighting", Required: false},
		{Name: "regions", Type: "int", Default: "8", Description: "Number of edge regions (4, 8, 12, 16)", Required: false},
		{Name: "sample-percent", Type: "int", Default: "10", Description: "Percentage of edge to sample (1-50)", Required: false},
		{Name: "sample-method", Type: "string", Default: "average", Description: "Sampling method (average or dominant)", Required: false},
		{Name: "seed-mode", Type: "string", Default: "content", Description: "Seed mode (content, manual, random)", Required: false},
		{Name: "seed-value", Type: "int64", Default: "0", Description: "Manual seed value", Required: false},
		{Name: "trim-letterbox", Type: "bool", Default: "true", Description: "Trim solid letterbox borders from generated images", Required: false},
		// Plugin-specific flags
		{Name: "minimax.base-url", Type: "string", Default: defaultAPIURL, Description: "MiniMax image generation endpoint URL (override for proxies)", Required: false},
		{Name: "minimax.image-size", Type: "string", Default: defaultImageSize, Description: "Target long-edge resolution: 1K, 2K, or pixels (512-2048)", Required: false},
	}
}
