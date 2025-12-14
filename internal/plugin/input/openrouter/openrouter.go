// Package openrouter provides an input plugin for generating images using OpenRouter.ai models.
package openrouter

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
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/image"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/regions"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/seed"
)

const (
	// RegionWeightFactor determines how much weight region colors receive
	// relative to main palette colors. Region colors get 10% of the total weight.
	RegionWeightFactor = 0.1

	// MainColorWeightRatio is the proportion of total weight allocated to
	// main palette colors when region extraction is enabled (90%).
	MainColorWeightRatio = 0.9

	// wallpaperEnhancement contains the suffix added to prompts to optimize
	// generated images for use as desktop wallpapers.
	wallpaperEnhancement = ", high quality desktop wallpaper suitable for widescreen and ultrawidescreen computer monitors, edge-to-edge composition, full bleed, seamless edges, vibrant colors, no borders, no frames, no padding"

	// defaultNegativePrompt contains the default negative prompt used to prevent
	// unwanted borders, frames, and visual artifacts in generated images.
	defaultNegativePrompt = "white borders, white edges, black borders, black edges, gray borders, padding, margins, letterbox, pillarbox, widescreen bars, black bars, frames, picture frames, border around image, vignette edges, faded edges, cropped edges, incomplete edges, cut off edges, canvas texture, matting, mounting"

	// apiBaseURL is the base URL for OpenRouter API.
	apiBaseURL = "https://openrouter.ai/api/v1"
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

	// API response data
	FinishReason string `json:"finish_reason,omitempty"`
}

// Plugin implements the input.Plugin interface for OpenRouter image generation.
type Plugin struct {
	// Plugin-specific flags (not shared)
	preferFree bool

	// Wallpaper support
	loadedImagePath string
}

// New creates a new OpenRouter input plugin with default settings.
func New() *Plugin {
	return &Plugin{
		preferFree: true,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "openrouter"
}

// Description returns a human-readable description.
func (p *Plugin) Description() string {
	return "Generate images with OpenRouter.ai models and extract colours"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return "0.0.1"
}

// RegisterFlags registers plugin-specific flags.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	// Register shared AI flags (prompt, model, list-models, etc.)
	aiflags.RegisterFlags(cmd)

	// Register shared common flags (count, aspect-ratio, cache, regions, seed, etc.)
	commonflags.RegisterFlags(cmd)

	// Register plugin-specific flags with openrouter prefix
	if cmd.Flags().Lookup("openrouter.prefer-free") == nil {
		cmd.Flags().BoolVar(&p.preferFree, "openrouter.prefer-free", p.preferFree, "Prefer free models when using auto selection")
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

// Generate creates an image using OpenRouter and extracts colors.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	// If list-models flag is set, list models and exit
	if aiflags.ListModels {
		if err := p.listAvailableModels(ctx, opts.Verbose); err != nil {
			return nil, fmt.Errorf("failed to list models: %w", err)
		}
		os.Exit(0)
	}

	// Determine model to use
	model := aiflags.Model
	if model == "" || model == "auto" {
		selectedModel, err := p.selectCheapestModel(ctx, opts.Verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-select model: %w", err)
		}
		model = selectedModel
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "OpenRouter Plugin Configuration:\n")
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

	// Determine base image path (without extension)
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
		enhancedPrompt := p.enhancePromptForWallpaper(aiflags.Prompt)
		additionalPrompt := enhancedPrompt[len(aiflags.Prompt):]
		fmt.Fprintf(os.Stderr, "[openrouter] model=%s prompt=\"%s\" additional=\"%s\"\n",
			model, aiflags.Prompt, additionalPrompt)
		fmt.Fprintf(os.Stderr, "Waiting for response...\n")

		imagePath, err = p.generateImage(ctx, model, imageBasePath, opts.Verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to generate image: %w", err)
		}

		fmt.Fprintf(os.Stderr, "Image generated: %s\n", imagePath)
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

// WallpaperPath returns the path to the generated image for wallpaper use.
func (p *Plugin) WallpaperPath() string {
	return p.loadedImagePath
}

// getImageBasePath determines the base path (without extension) for the generated image.
// The actual extension is determined after we know the image format from the API response.
func (p *Plugin) getImageBasePath(model string) (string, error) {
	if !commonflags.CacheEnabled {
		// For temp files, create without extension - we'll rename after
		tmpFile, err := os.CreateTemp("", "tinct-openrouter-*")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpFile.Close()
		return tmpFile.Name(), nil
	}

	// Use plugin-specific subdirectory
	cacheDir := filepath.Join(commonflags.CacheDir, "openrouter")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	filename := commonflags.CacheFilename
	if filename == "" {
		hash := sha256.Sum256([]byte(aiflags.Prompt + model))
		hashStr := hex.EncodeToString(hash[:])[:16]
		// Return without extension - will be added based on actual image format
		filename = fmt.Sprintf("openrouter-%s", hashStr)
	} else {
		// If user provided a filename, strip any extension - we'll add the correct one
		ext := filepath.Ext(filename)
		if ext != "" {
			filename = strings.TrimSuffix(filename, ext)
		}
	}

	return filepath.Join(cacheDir, filename), nil
}

// getExtensionForMIME returns the file extension for an image MIME type.
func getExtensionForMIME(mimeType string) string {
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		// Default to png for unknown types
		return ".png"
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

// getAPIKey retrieves the OpenRouter API key from the environment.
func getAPIKey() (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY environment variable is required\nGet one at: https://openrouter.ai/keys")
	}
	return apiKey, nil
}

// enhancePromptForWallpaper adds wallpaper-specific enhancements to a user prompt.
// Returns the original prompt unchanged if noExtendedPrompt is enabled.
func (p *Plugin) enhancePromptForWallpaper(basePrompt string) string {
	if aiflags.NoExtendedPrompt {
		return basePrompt
	}
	return basePrompt + wallpaperEnhancement
}

// buildNegativePrompt constructs the final negative prompt.
func buildNegativePrompt(noNegativePrompt bool) string {
	if noNegativePrompt {
		return ""
	}
	if aiflags.NegativePrompt != "" {
		return fmt.Sprintf("%s, %s", aiflags.NegativePrompt, defaultNegativePrompt)
	}
	return defaultNegativePrompt
}

// ChatCompletionRequest represents the request body for OpenRouter chat completions.
type ChatCompletionRequest struct {
	Model       string       `json:"model"`
	Messages    []Message    `json:"messages"`
	Modalities  []string     `json:"modalities,omitempty"`
	ImageConfig *ImageConfig `json:"image_config,omitempty"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ImageConfig contains image generation configuration.
type ImageConfig struct {
	AspectRatio string `json:"aspect_ratio,omitempty"`
}

// ChatCompletionResponse represents the response from OpenRouter.
type ChatCompletionResponse struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int             `json:"index"`
	Message      ResponseMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// ResponseMessage represents the assistant's response message.
type ResponseMessage struct {
	Role    string      `json:"role"`
	Content string      `json:"content"`
	Images  []ImageData `json:"images,omitempty"`
}

// ImageData represents an image in the response.
type ImageData struct {
	Type     string   `json:"type"`
	ImageURL ImageURL `json:"image_url"`
}

// ImageURL contains the image URL data.
type ImageURL struct {
	URL string `json:"url"`
}

// APIError represents an API error response.
type APIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Code    interface{} `json:"code"` // Can be string or number depending on provider
}

// CodeString returns the error code as a string.
func (e *APIError) CodeString() string {
	if e.Code == nil {
		return ""
	}
	switch v := e.Code.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", e.Code)
	}
}

// generateImage calls OpenRouter API to create an image.
// outputBasePath is the path without extension - the actual extension is determined
// from the API response and the final path is returned.
func (p *Plugin) generateImage(ctx context.Context, model, outputBasePath string, verbose bool) (actualPath string, err error) {
	apiKey, err := getAPIKey()
	if err != nil {
		return "", err
	}

	// Build the prompt with negative prompt if applicable
	enhancedPrompt := p.enhancePromptForWallpaper(aiflags.Prompt)
	negativePrompt := buildNegativePrompt(aiflags.NoNegativePrompt)

	// Combine prompt with negative prompt instruction
	fullPrompt := enhancedPrompt
	if negativePrompt != "" {
		fullPrompt = fmt.Sprintf("%s\n\nAvoid: %s", enhancedPrompt, negativePrompt)
	}

	// Build request
	request := ChatCompletionRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "user",
				Content: fmt.Sprintf("%s", fullPrompt),
			},
		},
		Modalities: []string{"image", "text"},
	}

	// Add aspect ratio configuration for supported models
	if commonflags.AspectRatio != "" {
		request.ImageConfig = &ImageConfig{
			AspectRatio: commonflags.AspectRatio,
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Calling OpenRouter API with model: %s\n", model)
		fmt.Fprintf(os.Stderr, "  Aspect ratio: %s\n", commonflags.AspectRatio)
		fmt.Fprintf(os.Stderr, "  Enhanced prompt: %s\n", enhancedPrompt)
		if negativePrompt != "" {
			fmt.Fprintf(os.Stderr, "  Negative prompt: %s\n", negativePrompt)
		}
	}

	// Marshal request
	reqBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiBaseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/jmylchreest/tinct")
	req.Header.Set("X-Title", "tinct")

	// Execute request
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Response status: %d\n", resp.StatusCode)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response ChatCompletionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	// Check for choices
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	// Get image from response
	choice := response.Choices[0]
	if len(choice.Message.Images) == 0 {
		return "", fmt.Errorf("no images in response (model may not support image generation)")
	}

	// Extract base64 image data
	imageURL := choice.Message.Images[0].ImageURL.URL
	if !strings.HasPrefix(imageURL, "data:image/") {
		return "", fmt.Errorf("unexpected image URL format: %s", imageURL[:min(50, len(imageURL))])
	}

	// Parse data URL: data:image/png;base64,<data>
	// Extract MIME type to determine file extension
	parts := strings.SplitN(imageURL, ",", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid data URL format")
	}

	// Parse MIME type from "data:image/jpeg;base64"
	mimeType := ""
	header := parts[0] // e.g., "data:image/jpeg;base64"
	if strings.HasPrefix(header, "data:") {
		header = strings.TrimPrefix(header, "data:")
		if idx := strings.Index(header, ";"); idx != -1 {
			mimeType = header[:idx]
		}
	}

	// Determine file extension based on MIME type
	ext := getExtensionForMIME(mimeType)
	actualPath = outputBasePath + ext

	if verbose {
		fmt.Fprintf(os.Stderr, "Image format: %s (extension: %s)\n", mimeType, ext)
	}

	imageBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode image data: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Received image data: %d bytes\n", len(imageBytes))
	}

	// Write image to file with correct extension
	if err := os.WriteFile(actualPath, imageBytes, 0o600); err != nil {
		return "", fmt.Errorf("failed to write image to file: %w", err)
	}

	// Save metadata alongside the image
	metadata := &ImageMetadata{
		Prompt:         aiflags.Prompt,
		EnhancedPrompt: enhancedPrompt,
		NegativePrompt: buildNegativePrompt(aiflags.NoNegativePrompt),
		Model:          model,
		AspectRatio:    commonflags.AspectRatio,
		CreatedAt:      time.Now(),
		ImagePath:      actualPath,
		ImageSize_:     len(imageBytes),
		FinishReason:   choice.FinishReason,
	}

	if err := p.saveMetadata(actualPath, metadata); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to save metadata: %v\n", err)
	}

	return actualPath, nil
}

// saveMetadata saves generation metadata as a JSON file alongside the image.
func (p *Plugin) saveMetadata(imagePath string, metadata *ImageMetadata) error {
	// Create metadata file path by replacing extension with .json
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

// extractColors extracts colors from the generated image.
func (p *Plugin) extractColors(imagePath string, verbose bool) (*colour.Palette, error) {
	// Load image using tinct's SmartLoader
	loader := image.NewSmartLoader()
	img, err := loader.Load(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Prepare extractor options with seed
	extractorOpts := colour.ExtractorOptions{}

	// Parse seed mode
	seedMode, err := seed.ParseMode(commonflags.SeedMode)
	if err != nil {
		return nil, fmt.Errorf("invalid seed mode: %w", err)
	}

	// Calculate seed using shared utility
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

	// Set seed in extractor options (except for random mode)
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

	// Create k-means extractor
	extractor, err := colour.NewExtractor(colour.AlgorithmKMeans, extractorOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create extractor: %w", err)
	}

	// Extract main palette colors
	palette, err := extractor.Extract(img, commonflags.Count)
	if err != nil {
		return nil, fmt.Errorf("failed to extract colors: %w", err)
	}

	// If ambience extraction is disabled, return main colors
	if !commonflags.ExtractAmbience {
		return palette, nil
	}

	// Extract region colors for ambient lighting
	if verbose {
		fmt.Fprintf(os.Stderr, "Also extracting %d edge/corner regions using %s method\n", commonflags.Regions, commonflags.SampleMethod)
	}

	// Convert regions count to configuration
	regionConfig, err := regions.ConfigurationFromInt(commonflags.Regions)
	if err != nil {
		return nil, fmt.Errorf("invalid regions configuration: %w", err)
	}

	// Create region sampler
	sampler := &regions.Sampler{
		SamplePercent: commonflags.SamplePercent,
		Method:        commonflags.SampleMethod,
	}

	// Extract colors from regions
	regionPalette, err := sampler.Extract(img, regionConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to extract region colors: %w", err)
	}

	// Combine main and region colors with weights
	totalColors := len(palette.Colors) + len(regionPalette.Colors)
	allColors := make([]color.Color, totalColors)
	weights := make([]float64, totalColors)

	// Main colors get 90% weight (distributed evenly)
	mainWeight := MainColorWeightRatio / float64(len(palette.Colors))
	for i, c := range palette.Colors {
		allColors[i] = c
		weights[i] = mainWeight
	}

	// Region colors get 10% weight (distributed evenly)
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// stripMarkdownLinks converts markdown links [text](url) to just text.
func stripMarkdownLinks(s string) string {
	// Match [text](url) pattern and replace with just text
	result := s
	for {
		start := strings.Index(result, "[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "](")
		if end == -1 {
			break
		}
		end += start
		urlEnd := strings.Index(result[end:], ")")
		if urlEnd == -1 {
			break
		}
		urlEnd += end
		// Extract the link text
		linkText := result[start+1 : end]
		// Replace the entire markdown link with just the text
		result = result[:start] + linkText + result[urlEnd+1:]
	}
	return result
}

// wrapText wraps text to a specified width, using indent for continuation lines.
// The first line is returned without indent; subsequent lines are prefixed with indent.
func wrapText(text string, width int, indent string) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	lineLen := 0
	firstLine := true

	for i, word := range words {
		wordLen := len(word)

		// Check if we need to start a new line
		if lineLen > 0 && lineLen+1+wordLen > width {
			result.WriteString("\n")
			result.WriteString(indent)
			lineLen = len(indent)
			firstLine = false
		}

		// Add space before word (except at start of line)
		if lineLen > 0 && (firstLine || lineLen > len(indent)) {
			result.WriteString(" ")
			lineLen++
		}

		// Handle words longer than width by breaking them
		if wordLen > width-len(indent) && !firstLine {
			// Write what fits on current line
			remaining := word
			for len(remaining) > 0 {
				available := width - lineLen
				if available <= 0 {
					result.WriteString("\n")
					result.WriteString(indent)
					lineLen = len(indent)
					available = width - lineLen
				}
				if len(remaining) <= available {
					result.WriteString(remaining)
					lineLen += len(remaining)
					remaining = ""
				} else {
					result.WriteString(remaining[:available])
					remaining = remaining[available:]
					lineLen = width
				}
			}
		} else {
			result.WriteString(word)
			lineLen += wordLen
		}

		_ = i // suppress unused variable warning
	}

	return result.String()
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		// Shared AI flags
		{Name: "ai.prompt", Type: "string", Default: "", Description: "Text description for AI image generation (required)", Required: true},
		{Name: "ai.model", Type: "string", Default: "auto", Description: "AI model to use (auto selects cheapest image model)", Required: false},
		{Name: "ai.list-models", Type: "bool", Default: "false", Description: "List available AI models and exit", Required: false},
		{Name: "ai.no-extended-prompt", Type: "bool", Default: "false", Description: "Disable automatic wallpaper prompt enhancements", Required: false},
		{Name: "ai.no-negative-prompt", Type: "bool", Default: "false", Description: "Disable default negative prompt", Required: false},
		{Name: "ai.negative-prompt", Type: "string", Default: "", Description: "Custom negative prompt to discourage certain elements", Required: false},
		// Shared common flags
		{Name: "count", Type: "int", Default: "32", Description: "Number of colors to extract", Required: false},
		{Name: "aspect-ratio", Type: "string", Default: "16:9", Description: "Image aspect ratio (1:1, 3:4, 4:3, 9:16, 16:9)", Required: false},
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
		// Plugin-specific flags
		{Name: "openrouter.prefer-free", Type: "bool", Default: "true", Description: "Prefer free models when using auto selection", Required: false},
	}
}

// ModelsResponse represents the response from the models API.
type ModelsResponse struct {
	Data []Model `json:"data"`
}

// Model represents a model in the API response.
type Model struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Created      int64        `json:"created"`
	Architecture Architecture `json:"architecture"`
	Pricing      Pricing      `json:"pricing"`
	TopProvider  *TopProvider `json:"top_provider,omitempty"`
}

// Architecture contains model capability information.
type Architecture struct {
	Modality         string   `json:"modality"`
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Tokenizer        string   `json:"tokenizer"`
}

// Pricing contains model pricing information.
type Pricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Request    string `json:"request"`
	Image      string `json:"image"`
}

// TopProvider contains provider-specific information.
type TopProvider struct {
	ContextLength       int  `json:"context_length"`
	MaxCompletionTokens int  `json:"max_completion_tokens"`
	IsModerated         bool `json:"is_moderated"`
}

// fetchModels retrieves the list of models from OpenRouter API.
func (p *Plugin) fetchModels(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", apiBaseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// API key is optional for models endpoint
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("API returned status %d (failed to read body: %w)", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Data, nil
}

// filterImageModels returns only models that support image output.
func filterImageModels(models []Model) []Model {
	var imageModels []Model
	for _, model := range models {
		if slices.Contains(model.Architecture.OutputModalities, "image") {
			imageModels = append(imageModels, model)
		}
	}
	return imageModels
}

// parsePrice converts a price string to a float64.
func parsePrice(priceStr string) float64 {
	if priceStr == "" || priceStr == "0" {
		return 0
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0
	}
	return price
}

// PricingType indicates the pricing model used by a model.
type PricingType int

const (
	PricingFree PricingType = iota
	PricingPerImage
	PricingPerRequest
	PricingPerToken
)

// PricingInfo contains parsed pricing information for a model.
type PricingInfo struct {
	Type           PricingType
	ImageCost      float64 // Cost per image
	RequestCost    float64 // Cost per request
	PromptCost     float64 // Cost per input token
	CompletionCost float64 // Cost per output token
}

// Minimum cost threshold - values below this are considered essentially free/zero
// This helps filter out negligible per-image costs when token pricing is the real cost
const minSignificantCost = 0.0001 // $0.0001

// getTypicalTokenUsage returns estimated input/output tokens based on provider.
// Different providers have very different token usage patterns for image generation:
//   - OpenAI: Image returned separately, low token counts
//   - Gemini: Image encoded as output tokens, high output counts
//
// Observed actual costs (same prompt across models):
//   - openai/gpt-5-image: $0.022 -> ~1500 input, ~700 output
//   - openai/gpt-5-image-mini: $0.008 -> ~1500 input, ~700 output
//   - google/gemini-2.5-flash-image: $0.0387 -> ~3000 input, ~14500 output
//   - google/gemini-3-pro-image-preview: $0.139 -> ~3000 input, ~5500 output
func getTypicalTokenUsage(modelID string) (inputTokens, outputTokens int) {
	switch {
	case strings.HasPrefix(modelID, "google/gemini-2.5"):
		// Gemini 2.5 Flash encodes image as ~14500 output tokens
		return 3000, 14500
	case strings.HasPrefix(modelID, "google/gemini-3"):
		// Gemini 3 Pro has lower output but higher per-token cost
		return 3000, 5500
	case strings.HasPrefix(modelID, "google/"):
		// Other Google models - use Gemini 3 estimates as default
		return 3000, 5500
	case strings.HasPrefix(modelID, "openai/"):
		// OpenAI returns image separately, low token usage
		return 1500, 700
	default:
		// Conservative default for unknown providers
		return 2000, 2000
	}
}

// getPricingInfo extracts detailed pricing information from a model.
func getPricingInfo(model Model) PricingInfo {
	info := PricingInfo{
		ImageCost:      parsePrice(model.Pricing.Image),
		RequestCost:    parsePrice(model.Pricing.Request),
		PromptCost:     parsePrice(model.Pricing.Prompt),
		CompletionCost: parsePrice(model.Pricing.Completion),
	}

	// Determine pricing type
	// Priority: per-image > per-request > per-token > free
	// But only if the cost is significant (some APIs return tiny non-zero values)

	// If there's significant per-image cost, use that
	if info.ImageCost >= minSignificantCost {
		info.Type = PricingPerImage
		return info
	}

	// If there's significant per-request cost, use that
	if info.RequestCost >= minSignificantCost {
		info.Type = PricingPerRequest
		return info
	}

	// If there's token pricing, use that
	if info.PromptCost > 0 || info.CompletionCost > 0 {
		info.Type = PricingPerToken
		return info
	}

	// Check for tiny image/request costs that might still be non-free
	if info.ImageCost > 0 {
		info.Type = PricingPerImage
		return info
	}
	if info.RequestCost > 0 {
		info.Type = PricingPerRequest
		return info
	}

	info.Type = PricingFree
	return info
}

// isModelFree returns true if the model has no associated costs.
func isModelFree(model Model) bool {
	return getPricingInfo(model).Type == PricingFree
}

// getModelCost estimates the total cost for a typical image generation request.
// This includes per-image/request costs PLUS estimated token costs.
// Token usage varies significantly by provider (see getTypicalTokenUsage).
func getModelCost(model Model) float64 {
	info := getPricingInfo(model)

	// Get provider-specific token estimates
	inputTokens, outputTokens := getTypicalTokenUsage(model.ID)

	// Calculate token costs
	tokenCost := (float64(inputTokens) * info.PromptCost) +
		(float64(outputTokens) * info.CompletionCost)

	switch info.Type {
	case PricingFree:
		return 0
	case PricingPerImage:
		// Image cost PLUS token costs (hybrid pricing like Gemini)
		return info.ImageCost + tokenCost
	case PricingPerRequest:
		return info.RequestCost + tokenCost
	case PricingPerToken:
		return tokenCost
	default:
		return 0
	}
}

// formatPricing returns a human-readable pricing string for display.
func formatPricing(model Model) string {
	info := getPricingInfo(model)
	estimated := getModelCost(model)

	// Check if model has token costs in addition to base pricing
	hasTokenCosts := info.PromptCost > 0 || info.CompletionCost > 0

	switch info.Type {
	case PricingFree:
		return "free"
	case PricingPerImage:
		if hasTokenCosts {
			// Hybrid pricing (image + tokens) like Gemini
			return fmt.Sprintf("$%.4f/image + tokens (~$%.4f total)", info.ImageCost, estimated)
		}
		return fmt.Sprintf("$%.4f/image", info.ImageCost)
	case PricingPerRequest:
		if hasTokenCosts {
			return fmt.Sprintf("$%.4f/request + tokens (~$%.4f total)", info.RequestCost, estimated)
		}
		return fmt.Sprintf("$%.4f/request", info.RequestCost)
	case PricingPerToken:
		// Display per-million tokens for readability
		promptPerM := info.PromptCost * 1_000_000
		completionPerM := info.CompletionCost * 1_000_000
		return fmt.Sprintf("$%.2f/M input, $%.2f/M output (~$%.4f/image)", promptPerM, completionPerM, estimated)
	default:
		return "unknown"
	}
}

// selectCheapestModel selects the cheapest model that supports image output.
func (p *Plugin) selectCheapestModel(ctx context.Context, verbose bool) (string, error) {
	if verbose {
		fmt.Fprintf(os.Stderr, "Auto-selecting cheapest image generation model...\n")
	}

	models, err := p.fetchModels(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch models: %w", err)
	}

	imageModels := filterImageModels(models)
	if len(imageModels) == 0 {
		return "", fmt.Errorf("no image generation models available")
	}

	// Sort by cost
	sort.Slice(imageModels, func(i, j int) bool {
		costI := getModelCost(imageModels[i])
		costJ := getModelCost(imageModels[j])
		return costI < costJ
	})

	// If preferFree is enabled, try to find a free model first
	if p.preferFree {
		for _, model := range imageModels {
			if isModelFree(model) {
				if verbose {
					fmt.Fprintf(os.Stderr, "Selected free model: %s\n", model.ID)
				}
				return model.ID, nil
			}
		}
	}

	// Return the cheapest model
	selected := imageModels[0]
	if verbose {
		if isModelFree(selected) {
			fmt.Fprintf(os.Stderr, "Selected model: %s (free)\n", selected.ID)
		} else {
			fmt.Fprintf(os.Stderr, "Selected model: %s (%s)\n", selected.ID, formatPricing(selected))
		}
	}

	return selected.ID, nil
}

// listAvailableModels lists available image generation models from the API.
func (p *Plugin) listAvailableModels(ctx context.Context, verbose bool) error {
	if verbose {
		fmt.Fprintf(os.Stderr, "Fetching available models from OpenRouter API...\n\n")
	}

	models, err := p.fetchModels(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		fmt.Fprintf(os.Stderr, "Unable to fetch models from API.\n")
		return nil
	}

	imageModels := filterImageModels(models)

	fmt.Println("Available Image Generation Models:")
	fmt.Println()
	fmt.Printf("Default: auto (auto-selects cheapest model)\n")
	fmt.Println()
	fmt.Println("Note: Pricing shown is retrieved from the API. Token-based pricing (~$X/image)")
	fmt.Println("      is estimated and actual costs may vary. Free models are listed first.")
	fmt.Println()

	if len(imageModels) == 0 {
		fmt.Println("No image generation models found.")
		return nil
	}

	// Sort by cost (free first, then by price)
	sort.Slice(imageModels, func(i, j int) bool {
		costI := getModelCost(imageModels[i])
		costJ := getModelCost(imageModels[j])
		return costI < costJ
	})

	for _, model := range imageModels {
		fmt.Printf("Model: %s\n", model.ID)
		if model.Name != "" && model.Name != model.ID {
			fmt.Printf("  Name: %s\n", model.Name)
		}
		if model.Description != "" {
			// Word-wrap long descriptions with proper indentation
			// Strip markdown links for cleaner display
			desc := stripMarkdownLinks(model.Description)
			fmt.Printf("  Description: %s\n", wrapText(desc, 80, "               "))
		}
		fmt.Printf("  Pricing: %s\n", formatPricing(model))
		fmt.Printf("  Input: %v\n", model.Architecture.InputModalities)
		fmt.Printf("  Output: %v\n", model.Architecture.OutputModalities)
		fmt.Println()
	}

	fmt.Printf("Total image generation models: %d\n", len(imageModels))
	fmt.Println()
	fmt.Println("Note: Not all models appear in the API response. For the complete list, visit:")
	fmt.Println("  https://openrouter.ai/models?fmt=cards&input_modalities=image%2Ctext&output_modalities=image")
	fmt.Println()
	fmt.Println("You can use any model ID with --ai.model, even if not listed above.")

	return nil
}
