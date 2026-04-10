// Package googlegenai provides an input plugin for generating images using Google's Imagen models.
package googlegenai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/color"
	_ "image/jpeg" // Required for JPEG image decoding
	_ "image/png"  // Required for PNG image decoding
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/genai"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/imagecolours"
	"github.com/jmylchreest/tinct/internal/version"
)

const (
	// wallpaperEnhancement contains the suffix added to prompts to optimize
	// generated images for use as desktop wallpapers.
	wallpaperEnhancement = ", high quality desktop wallpaper suitable for widescreen and ultrawidescreen computer monitors, edge-to-edge composition, full bleed, seamless edges, vibrant colors, no borders, no frames, no padding"

	// defaultNegativePrompt contains the default negative prompt used to prevent
	// unwanted borders, frames, and visual artifacts in generated images.
	defaultNegativePrompt = "white borders, white edges, black borders, black edges, gray borders, padding, margins, letterbox, pillarbox, widescreen bars, black bars, frames, picture frames, border around image, vignette edges, faded edges, cropped edges, incomplete edges, cut off edges, canvas texture, matting, mounting"

	// modelPrefix is the prefix that Google API returns for model names.
	modelPrefix = "models/"

	// defaultModel is the default model used when none is specified.
	defaultModel = "gemini-2.5-flash-image"

	// defaultBackend is the default backend used when none is specified.
	defaultBackend = "gemini-api"
)

// ImageMetadata contains metadata about a generated image.
type ImageMetadata struct {
	// Generation parameters
	Prompt         string `json:"prompt"`
	EnhancedPrompt string `json:"enhanced_prompt,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Model          string `json:"model"`
	Backend        string `json:"backend"`
	AspectRatio    string `json:"aspect_ratio"`
	ImageSize      string `json:"image_size,omitempty"`

	// Generation details
	CreatedAt  time.Time `json:"created_at"`
	ImagePath  string    `json:"image_path"`
	ImageSize_ int       `json:"image_bytes"`

	// API response data
	RAIFilteredReason string         `json:"rai_filtered_reason,omitempty"`
	SafetyAttributes  map[string]any `json:"safety_attributes,omitempty"`
}

// Plugin implements the input.Plugin interface for Google Imagen image generation.
type Plugin struct {
	// Plugin-specific flags (not shared)
	backend   string
	imageSize string

	// Wallpaper support
	loadedImagePath string
}

// New creates a new Google Gen AI input plugin with default settings.
func New() *Plugin {
	return &Plugin{
		backend:   defaultBackend,
		imageSize: "2K",
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "google-genai"
}

// Description returns a human-readable description.
func (p *Plugin) Description() string {
	return "Generate images with Google Imagen and extract colours"
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

	// Register plugin-specific flags with googlegenai prefix
	if cmd.Flags().Lookup("googlegenai.backend") == nil {
		cmd.Flags().StringVar(&p.backend, "googlegenai.backend", p.backend, "Google Gen AI backend (gemini-api or vertex-ai)")
	}
	if cmd.Flags().Lookup("googlegenai.image-size") == nil {
		cmd.Flags().StringVar(&p.imageSize, "googlegenai.image-size", p.imageSize, "Image size (1K or 2K, only for Imagen Standard/Ultra models)")
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

// Generate creates an image using Google Gen AI and extracts colors.
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
		model = defaultModel
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Google Gen AI Plugin Configuration:\n")
		fmt.Fprintf(os.Stderr, "  Prompt: %s\n", aiflags.Prompt)
		fmt.Fprintf(os.Stderr, "  Model: %s\n", model)
		fmt.Fprintf(os.Stderr, "  Backend: %s\n", p.backend)
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

	// Determine image path
	imagePath, err := p.getImagePath(model)
	if err != nil {
		return nil, fmt.Errorf("failed to determine image path: %w", err)
	}

	// Generate image if needed
	if commonflags.CacheOverwrite || !fileExists(imagePath) {
		enhancedPrompt := p.enhancePromptForWallpaper(aiflags.Prompt)
		additionalPrompt := enhancedPrompt[len(aiflags.Prompt):]
		fmt.Fprintf(os.Stderr, "[google-genai] backend=%s model=%s prompt=\"%s\" additional=\"%s\"\n",
			p.backend, model, aiflags.Prompt, additionalPrompt)
		fmt.Fprintf(os.Stderr, "Waiting for response...\n")

		if err := p.generateImage(ctx, model, imagePath, opts.Verbose); err != nil {
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

// getImagePath determines where to save/load the generated image.
func (p *Plugin) getImagePath(model string) (string, error) {
	if !commonflags.CacheEnabled {
		tmpFile, err := os.CreateTemp("", "tinct-genai-*.png")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		_ = tmpFile.Close() // File will be rewritten; close error is not actionable
		return tmpFile.Name(), nil
	}

	// Use plugin-specific subdirectory
	cacheDir := filepath.Join(commonflags.CacheDir, "google-genai")
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	filename := commonflags.CacheFilename
	if filename == "" {
		hash := sha256.Sum256([]byte(aiflags.Prompt + model))
		hashStr := hex.EncodeToString(hash[:])[:16]
		filename = fmt.Sprintf("genai-%s.png", hashStr)
	}

	return filepath.Join(cacheDir, filename), nil
}

// clientSetup encapsulates client configuration, creation, and logging.
// Returns the configured client or an error.
func (p *Plugin) clientSetup(ctx context.Context, verbose bool) (*genai.Client, error) {
	clientConfig := &genai.ClientConfig{}

	if p.backend == "vertex-ai" {
		clientConfig.Backend = genai.BackendVertexAI
	} else {
		clientConfig.Backend = genai.BackendGeminiAPI
	}

	// Check for API key (required for Gemini API backend)
	if clientConfig.Backend == genai.BackendGeminiAPI {
		apiKey := os.Getenv("GOOGLE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GOOGLE_API_KEY environment variable is required\nGet one at: https://aistudio.google.com/api-keys")
		}
		clientConfig.APIKey = apiKey
	}

	// Create client
	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gen AI client: %w", err)
	}

	// Log backend information if verbose
	if verbose {
		backendName := "Gemini API"
		if client.ClientConfig().Backend == genai.BackendVertexAI {
			backendName = "Vertex AI"
		}
		fmt.Fprintf(os.Stderr, "Using %s backend\n", backendName)
	}

	return client, nil
}

// enhancePromptForWallpaper adds wallpaper-specific enhancements to a user prompt.
// Returns the original prompt unchanged if noExtendedPrompt is enabled.
func (p *Plugin) enhancePromptForWallpaper(basePrompt string) string {
	if aiflags.NoExtendedPrompt {
		return basePrompt
	}
	return basePrompt + wallpaperEnhancement
}

// buildNegativePrompt constructs the final negative prompt by combining
// user-provided and default negative prompts.
func buildNegativePrompt(userPrompt string, noNegative bool) string {
	if noNegative {
		return ""
	}
	if userPrompt == "" {
		return defaultNegativePrompt
	}
	return fmt.Sprintf("%s, %s", userPrompt, defaultNegativePrompt)
}

// isGeminiModel checks if a model uses the Gemini API (GenerateContent) vs Imagen API (GenerateImages).
func isGeminiModel(model string) bool {
	return model == "gemini-2.5-flash-image"
}

// generateImage calls Google Gen AI SDK to create an image.
// Routes to appropriate generation method based on model type.
func (p *Plugin) generateImage(ctx context.Context, model, outputPath string, verbose bool) error {
	if isGeminiModel(model) {
		return p.generateImageWithGemini(ctx, model, outputPath, verbose)
	}
	return p.generateImageWithImagen(ctx, model, outputPath, verbose)
}

// generateImageWithImagen generates an image using the Imagen API (GenerateImages).
func (p *Plugin) generateImageWithImagen(ctx context.Context, model, outputPath string, verbose bool) error { //nolint:gocyclo
	client, err := p.clientSetup(ctx, verbose)
	if err != nil {
		return err
	}

	// Enhance prompt for wallpaper suitability
	enhancedPrompt := p.enhancePromptForWallpaper(aiflags.Prompt)

	// Build generation config
	genConfig := &genai.GenerateImagesConfig{
		NumberOfImages: 1,
		AspectRatio:    commonflags.AspectRatio,
		OutputMIMEType: "image/png",
	}

	// Set image size if supported (only for Standard and Ultra models)
	if p.imageSize != "" && (model == "imagen-4.0-generate-001" || model == "imagen-4.0-ultra-generate-001") {
		genConfig.ImageSize = p.imageSize
	}

	// Build negative prompt to avoid borders, frames, and unwanted elements
	// Note: Negative prompts are only supported in Vertex AI backend, not Gemini API
	if client.ClientConfig().Backend == genai.BackendVertexAI && !aiflags.NoNegativePrompt {
		genConfig.NegativePrompt = buildNegativePrompt(aiflags.NegativePrompt, aiflags.NoNegativePrompt)
	} else if aiflags.NegativePrompt != "" && !aiflags.NoNegativePrompt {
		// Gemini API doesn't support negative prompts, warn user if they provided one
		fmt.Fprintf(os.Stderr, "Warning: Negative prompts are not supported with Gemini API backend (only Vertex AI)\n")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Calling GenerateImages with model: %s\n", model)
		fmt.Fprintf(os.Stderr, "  Aspect ratio: %s\n", commonflags.AspectRatio)
		if genConfig.ImageSize != "" {
			fmt.Fprintf(os.Stderr, "  Image size: %s\n", genConfig.ImageSize)
		}
		fmt.Fprintf(os.Stderr, "  Enhanced prompt: %s\n", enhancedPrompt)
		if genConfig.NegativePrompt != "" {
			fmt.Fprintf(os.Stderr, "  Negative prompt: %s\n", genConfig.NegativePrompt)
		} else {
			fmt.Fprintf(os.Stderr, "  Negative prompt: (not supported with Gemini API backend)\n")
		}
	}

	// Generate images
	response, err := client.Models.GenerateImages(ctx, model, enhancedPrompt, genConfig)
	if err != nil {
		return fmt.Errorf("image generation failed: %w", err)
	}

	// Check if we got any images
	if len(response.GeneratedImages) == 0 {
		return fmt.Errorf("no images generated in response")
	}

	// Get the first generated image
	generatedImage := response.GeneratedImages[0]

	// Check if image was filtered
	if generatedImage.RAIFilteredReason != "" {
		return fmt.Errorf("image was filtered by safety system: %s", generatedImage.RAIFilteredReason)
	}

	// Get image data
	if generatedImage.Image == nil {
		return fmt.Errorf("generated image has no image data")
	}

	imageBytes := generatedImage.Image.ImageBytes
	if len(imageBytes) == 0 {
		return fmt.Errorf("generated image has empty image data")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Received image data: %d bytes\n", len(imageBytes))
	}

	// Write image to file
	if err := os.WriteFile(outputPath, imageBytes, 0o600); err != nil {
		return fmt.Errorf("failed to write image to file: %w", err)
	}

	// Save metadata alongside the image
	metadata := &ImageMetadata{
		Prompt:            aiflags.Prompt,
		EnhancedPrompt:    enhancedPrompt,
		NegativePrompt:    genConfig.NegativePrompt,
		Model:             model,
		Backend:           p.backend,
		AspectRatio:       commonflags.AspectRatio,
		ImageSize:         p.imageSize,
		CreatedAt:         time.Now(),
		ImagePath:         outputPath,
		ImageSize_:        len(imageBytes),
		RAIFilteredReason: generatedImage.RAIFilteredReason,
	}

	// Extract safety attributes if available
	if generatedImage.SafetyAttributes != nil {
		metadata.SafetyAttributes = make(map[string]any)
		if len(generatedImage.SafetyAttributes.Categories) > 0 {
			metadata.SafetyAttributes["categories"] = generatedImage.SafetyAttributes.Categories
		}
		if len(generatedImage.SafetyAttributes.Scores) > 0 {
			metadata.SafetyAttributes["scores"] = generatedImage.SafetyAttributes.Scores
		}
	}

	if err := p.saveMetadata(outputPath, metadata); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to save metadata: %v\n", err)
	}

	return nil
}

// generateImageWithGemini generates an image using the Gemini API (GenerateContent).
func (p *Plugin) generateImageWithGemini(ctx context.Context, model, outputPath string, verbose bool) error { //nolint:gocyclo
	client, err := p.clientSetup(ctx, verbose)
	if err != nil {
		return err
	}

	// Enhance prompt for wallpaper suitability
	enhancedPrompt := p.enhancePromptForWallpaper(aiflags.Prompt)

	if verbose {
		fmt.Fprintf(os.Stderr, "Calling GenerateContent with model: %s\n", model)
		fmt.Fprintf(os.Stderr, "  Aspect ratio: %s\n", commonflags.AspectRatio)
		fmt.Fprintf(os.Stderr, "  Enhanced prompt: %s\n", enhancedPrompt)
	}

	// Build generation config for Gemini
	// Note: ResponseMIMEType is only for text outputs, not for image generation
	genConfig := &genai.GenerateContentConfig{
		ResponseModalities: []string{"Image"},
	}

	// Create the prompt with system instructions for aspect ratio
	promptText := fmt.Sprintf("Generate an image with aspect ratio %s: %s", commonflags.AspectRatio, enhancedPrompt)

	// Build content array with text prompt
	// genai.Text() returns []*Content, which we can use directly
	contents := genai.Text(promptText)

	// Generate content
	response, err := client.Models.GenerateContent(ctx, model, contents, genConfig)
	if err != nil {
		return fmt.Errorf("image generation failed: %w", err)
	}

	// Check if response is valid
	if response == nil {
		return fmt.Errorf("received nil response from API")
	}

	// Check if we got any parts in the response
	if len(response.Candidates) == 0 {
		return fmt.Errorf("no candidates in response")
	}

	if response.Candidates[0].Content == nil || len(response.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("no image data in response")
	}

	// Extract image data from the first part's inline data
	var imageBytes []byte
	for _, part := range response.Candidates[0].Content.Parts {
		if part.InlineData != nil && part.InlineData.Data != nil {
			imageBytes = part.InlineData.Data
			break
		}
	}

	if len(imageBytes) == 0 {
		return fmt.Errorf("no inline image data found in response")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Received image data: %d bytes\n", len(imageBytes))
	}

	// Write image to file
	if err := os.WriteFile(outputPath, imageBytes, 0o600); err != nil {
		return fmt.Errorf("failed to write image to file: %w", err)
	}

	// Save metadata alongside the image
	metadata := &ImageMetadata{
		Prompt:         aiflags.Prompt,
		EnhancedPrompt: enhancedPrompt,
		Model:          model,
		Backend:        p.backend,
		AspectRatio:    commonflags.AspectRatio,
		CreatedAt:      time.Now(),
		ImagePath:      outputPath,
		ImageSize_:     len(imageBytes),
	}

	if err := p.saveMetadata(outputPath, metadata); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to save metadata: %v\n", err)
	}

	return nil
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
		// Plugin-specific flags
		{Name: "googlegenai.backend", Type: "string", Default: defaultBackend, Description: "Google Gen AI backend (gemini-api or vertex-ai)", Required: false},
		{Name: "googlegenai.image-size", Type: "string", Default: "2K", Description: "Image size (1K or 2K, only for Imagen Standard/Ultra models)", Required: false},
	}
}

// listAvailableModels lists available Imagen models from the API.
func (p *Plugin) listAvailableModels(ctx context.Context, verbose bool) error {
	client, err := p.clientSetup(ctx, verbose)
	if err != nil {
		// For list-models, fall back to hardcoded list on API key error
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		fmt.Fprintf(os.Stderr, "Showing known models instead:\n\n")
		ListModels()
		return nil
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Fetching available models from API...\n\n")
	}

	// List all models
	fmt.Println("Available Image Generation Models:")
	fmt.Println()
	fmt.Printf("Default Model: %s\n", defaultModel)
	fmt.Printf("Default Backend: %s\n", defaultBackend)
	fmt.Println()
	fmt.Println("Note: Pricing shown is for indication only and may not reflect current rates.")
	fmt.Println("      Visit https://ai.google.dev/gemini-api/docs/pricing for up-to-date pricing.")
	fmt.Println()

	modelCount := 0
	for model, err := range client.Models.All(ctx) {
		if err != nil {
			// If we encounter an error, show what we got so far and fall back
			if modelCount > 0 {
				fmt.Fprintf(os.Stderr, "\nWarning: Error during model listing: %v\n", err)
				fmt.Fprintf(os.Stderr, "Showing %d models retrieved before error\n\n", modelCount)
				return nil
			}
			// If no models retrieved, fall back to hardcoded list
			fmt.Fprintf(os.Stderr, "Warning: Could not fetch models from API: %v\n", err)
			fmt.Fprintf(os.Stderr, "Showing known models instead:\n\n")
			ListModels()
			return nil
		}

		// Filter for image generation models (Imagen and Gemini)
		if model.Name != "" && isImageGenerationModel(model.Name) {
			modelCount++

			// Remove "models/" prefix from model name for cleaner output
			modelName := strings.TrimPrefix(model.Name, modelPrefix)

			fmt.Printf("Model: %s\n", modelName)
			if model.DisplayName != "" {
				fmt.Printf("  Display Name: %s\n", model.DisplayName)
			}
			if model.Description != "" {
				fmt.Printf("  Description: %s\n", model.Description)
			}

			// Show backend availability
			if isGeminiModel(modelName) {
				fmt.Printf("  Backend: gemini-api only\n")
				fmt.Printf("  Pricing: Free tier (500 RPD)\n")
			} else {
				fmt.Printf("  Backend: gemini-api, vertex-ai\n")
				fmt.Printf("  Pricing: gemini-api - Free tier (15 RPM); vertex-ai - Pay per use\n")
			}

			fmt.Println()
		}
	}

	if modelCount == 0 {
		fmt.Println("No image generation models found via API.")
		fmt.Println("Showing known models instead:")
		fmt.Println()
		ListModels()
	} else if verbose {
		fmt.Fprintf(os.Stderr, "\nTotal image generation models found: %d\n", modelCount)
	}

	// Show pricing and free tier information
	fmt.Println()
	fmt.Println("Pricing Information:")
	fmt.Println("  For current pricing and free tier details, visit:")
	fmt.Println("  https://ai.google.dev/gemini-api/docs/pricing")
	fmt.Println()
	fmt.Println("  Free tier available via Gemini API:")
	fmt.Println("  - 15 requests per minute")
	fmt.Println("  - Rate limits vary by model")

	return nil
}

// isImageGenerationModel checks if a model is an image generation model (Imagen or Gemini image models).
func isImageGenerationModel(name string) bool {
	// Convert to lowercase for case-insensitive matching
	nameLower := ""
	for _, r := range name {
		if r >= 'A' && r <= 'Z' {
			nameLower += string(r + 32)
		} else {
			nameLower += string(r)
		}
	}

	// Check for Imagen models
	if containsSubstring(nameLower, "imagen") {
		return true
	}

	// Check for Gemini image generation models
	// These contain "flash" (or similar) AND either "image" or support image generation
	if containsSubstring(nameLower, "flash") && containsSubstring(nameLower, "image") {
		return true
	}

	// Check for other Gemini image models
	if containsSubstring(nameLower, "gemini") && containsSubstring(nameLower, "image") {
		return true
	}

	return false
}

// containsSubstring checks if haystack contains needle (case-sensitive).
func containsSubstring(haystack, needle string) bool {
	if len(haystack) < len(needle) {
		return false
	}

	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// ListModels prints available Imagen and Gemini image generation models to stdout.
func ListModels() {
	models := []struct {
		ID          string
		Name        string
		Description string
		Cost        string
		Generation  string
	}{
		{
			ID:          "imagen-3.0-generate-002",
			Name:        "Imagen 3",
			Description: "Previous generation model (stable)",
			Cost:        "$0.03",
			Generation:  "3",
		},
		{
			ID:          "imagen-4.0-fast-generate-001",
			Name:        "Imagen 4 Fast",
			Description: "Fastest generation, ideal for high-volume tasks",
			Cost:        "$0.02",
			Generation:  "4",
		},
		{
			ID:          "imagen-4.0-generate-001",
			Name:        "Imagen 4",
			Description: "Flagship model with balanced quality and speed",
			Cost:        "$0.04",
			Generation:  "4",
		},
		{
			ID:          "imagen-4.0-ultra-generate-001",
			Name:        "Imagen 4 Ultra",
			Description: "Highest quality, best prompt alignment",
			Cost:        "$0.06",
			Generation:  "4",
		},
		{
			ID:          "gemini-2.5-flash-image",
			Name:        "Gemini 2.5 Flash Image (Nano Banana)",
			Description: "Fast multimodal image generation and editing, conversational workflow support (default)",
			Cost:        "$0.039 per image",
			Generation:  "Gemini 2.5",
		},
	}

	fmt.Println("Available Image Generation Models:")
	fmt.Println()
	fmt.Println("Default Model: gemini-2.5-flash-image")
	fmt.Println("Default Backend: gemini-api")
	fmt.Println()
	fmt.Println("Note: Pricing shown is for indication only and may not reflect current rates.")
	fmt.Println("      Visit https://ai.google.dev/gemini-api/docs/pricing for up-to-date pricing.")
	fmt.Println()

	for _, model := range models {
		fmt.Printf("ID: %s\n", model.ID)
		fmt.Printf("  Name: %s\n", model.Name)
		fmt.Printf("  Description: %s\n", model.Description)
		fmt.Printf("  Cost: %s (approximate)\n", model.Cost)
		fmt.Printf("  Generation: %s\n", model.Generation)

		// Show backend availability
		if isGeminiModel(model.ID) {
			fmt.Printf("  Backend: gemini-api only\n")
			fmt.Printf("  Pricing: Free tier (500 RPD)\n")
		} else {
			fmt.Printf("  Backend: gemini-api, vertex-ai\n")
			fmt.Printf("  Pricing: gemini-api - Free tier (15 RPM); vertex-ai - Pay per use\n")
		}

		fmt.Println()
	}

	fmt.Println("Pricing Information:")
	fmt.Println("  For current pricing and free tier details, visit:")
	fmt.Println("  https://ai.google.dev/gemini-api/docs/pricing")
	fmt.Println()
	fmt.Println("  Free tier available via Gemini API:")
	fmt.Println("  - 15 requests per minute")
	fmt.Println("  - Rate limits vary by model")
}
