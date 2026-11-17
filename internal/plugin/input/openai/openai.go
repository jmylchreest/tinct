// Package openai provides configurable OpenAI-compatible image generation.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoding.
	_ "image/png"  // Register PNG decoding.
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiprompt"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/imagecolours"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/imagepost"
	"github.com/jmylchreest/tinct/internal/version"
	"github.com/jmylchreest/tinct/pkg/ai"
	"github.com/spf13/cobra"
)

const defaultBaseURL = "https://api.openai.com/v1"
const defaultModel = "gpt-image-2.6-flare"

// Plugin implements input.Plugin and input.WallpaperProvider.
type Plugin struct {
	baseURL         string
	protocol        string
	keyEnv          string
	loadedImagePath string
}

// New creates an OpenAI-compatible plugin. Custom endpoints require an explicit model.
func New() *Plugin {
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Plugin{baseURL: baseURL, protocol: "images", keyEnv: "OPENAI_API_KEY"}
}

// Name returns the registry name.
func (p *Plugin) Name() string { return "openai" }

// Description describes the plugin.
func (p *Plugin) Description() string {
	return "Generate wallpapers with OpenAI-compatible image APIs, including local servers"
}

// Version returns the tinct version.
func (p *Plugin) Version() string { return version.Version }

// WallpaperPath returns the generated image path.
func (p *Plugin) WallpaperPath() string { return p.loadedImagePath }

// WallpaperRawPath returns the generated image path.
func (p *Plugin) WallpaperRawPath() string { return p.loadedImagePath }

// RegisterFlags registers shared controls and provider connection settings.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	aiflags.RegisterFlags(cmd)
	commonflags.RegisterFlags(cmd)
	if cmd.Flags().Lookup("openai.base-url") == nil {
		cmd.Flags().StringVar(&p.baseURL, "openai.base-url", p.baseURL, "API base URL including version prefix, e.g. http://localhost:8080/v1")
		cmd.Flags().StringVar(&p.protocol, "openai.protocol", p.protocol, "API protocol: images or chat (image-output extension)")
		cmd.Flags().StringVar(&p.keyEnv, "openai.api-key-env", p.keyEnv, "Environment variable containing the API key; empty disables authentication")
	}
}

func (p *Plugin) client() ai.Options {
	key := ""
	if p.keyEnv != "" {
		key = os.Getenv(p.keyEnv)
	}
	return ai.Options{BaseURL: p.baseURL, APIKey: key}
}

func (p *Plugin) model() (string, error) {
	if aiflags.Model != "" && aiflags.Model != "auto" {
		return aiflags.Model, nil
	}
	if strings.TrimRight(p.baseURL, "/") == defaultBaseURL && p.protocol == "images" {
		return defaultModel, nil
	}
	return "", fmt.Errorf("--ai.model is required for a custom endpoint or chat protocol")
}

// Validate checks request syntax and configuration before any network or cache activity.
func (p *Plugin) Validate() error {
	if _, err := aiflags.ParseParams(); err != nil {
		return err
	}
	u, err := url.Parse(p.baseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("AI base URL must be an http(s) URL without credentials, query or fragment")
	}
	if p.protocol != "images" && p.protocol != "chat" {
		return fmt.Errorf("unsupported AI protocol %q: use images or chat", p.protocol)
	}
	if aiflags.ListModels {
		return nil
	}
	if aiflags.Prompt == "" {
		return fmt.Errorf("--ai.prompt is required")
	}
	_, err = p.model()
	return err
}

func (p *Plugin) request(model string) ([]byte, error) {
	prompt := aiprompt.EnhanceForWallpaper(aiflags.Prompt)
	if commonflags.AspectRatio != "" {
		prompt = fmt.Sprintf("%s\nAspect ratio: %s.", prompt, commonflags.AspectRatio)
	}
	if negative := aiprompt.BuildNegative(aiflags.NegativePrompt, aiflags.NoNegativePrompt); negative != "" {
		prompt += "\nAvoid: " + negative
	}
	body := map[string]any{"model": model, "prompt": prompt, "n": 1}
	if p.protocol == "chat" {
		body = map[string]any{"model": model, "messages": []map[string]string{{"role": "user", "content": prompt}}, "modalities": []string{"image", "text"}, "image_config": map[string]string{"aspect_ratio": commonflags.AspectRatio}}
	}
	return aiflags.MarshalRequest(body)
}

// Generate generates or loads a wallpaper, then extracts its palette.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if aiflags.ListModels {
		fmt.Printf("OpenAI Images default: %s\nCustom endpoints: supply --ai.model explicitly; model availability and image capability depend on the server.\n", defaultModel)
		os.Exit(0)
	}
	model, err := p.model()
	if err != nil {
		return nil, err
	}
	body, err := p.request(model)
	if err != nil {
		return nil, err
	}
	if opts.DryRun {
		return colour.NewPalette(nil), nil
	}
	key, err := aiflags.CacheKey(model, p.baseURL, p.protocol, string(body))
	if err != nil {
		return nil, err
	}
	base, err := imageBasePath(key)
	if err != nil {
		return nil, err
	}
	path := cachedImage(base)
	if path == "" {
		path, err = p.generateImage(ctx, body, base, model, opts.Verbose)
		if err != nil {
			return nil, err
		}
	}

	p.loadedImagePath = path
	return imagecolours.ExtractColors(path, opts.Verbose)
}

func cachedImage(base string) string {
	if !commonflags.CacheEnabled || commonflags.CacheOverwrite {
		return ""
	}
	for _, ext := range []string{".png", ".jpg", ".webp", ".gif"} {
		candidate := base + ext
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func (p *Plugin) generateImage(ctx context.Context, body []byte, base, model string, verbose bool) (string, error) {
	client := p.client()
	if strings.TrimRight(p.baseURL, "/") == defaultBaseURL && client.APIKey == "" {
		return "", fmt.Errorf("%s is required for OpenAI", p.keyEnv)
	}
	var params map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&params); err != nil {
		return "", err
	}
	var driver ai.ImageGenerator = ai.OpenAI{Options: client}
	if p.protocol == "chat" {
		driver = ai.OpenRouter{Options: client}
	}
	result, err := driver.GenerateImage(ctx, ai.ImageRequest{Params: params})
	if err != nil {
		return "", err
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(result.Bytes))
	if err != nil {
		return "", fmt.Errorf("provider returned invalid image data: %w", err)
	}
	ext := "." + format
	if format == "jpeg" {
		ext = ".jpg"
	}
	path := base + ext
	if err := os.WriteFile(path, result.Bytes, 0o600); err != nil {
		return "", fmt.Errorf("save generated image: %w", err)
	}
	imagepost.TrimLetterboxIfEnabled(path, commonflags.AspectRatio, commonflags.TrimLetterbox, verbose)
	metadata := map[string]any{"model": model, "created_at": time.Now(), "image_path": path, "request": json.RawMessage(body), "base_url": p.baseURL, "protocol": p.protocol}
	if err := imagecolours.SaveMetadata(path, metadata); err != nil {
		return "", fmt.Errorf("save image metadata: %w", err)
	}
	return path, nil
}

func imageBasePath(key string) (string, error) {
	dir := filepath.Join(commonflags.CacheDir, "openai")
	if !commonflags.CacheEnabled {
		var err error
		dir, err = os.MkdirTemp("", "tinct-openai-*")
		if err != nil {
			return "", fmt.Errorf("create image directory: %w", err)
		}
	} else if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("create cache directory: %w", err)
	}
	name := "openai-" + key
	if commonflags.CacheFilename != "" {
		name = strings.TrimSuffix(filepath.Base(commonflags.CacheFilename), filepath.Ext(commonflags.CacheFilename))
		if name == "" || name == "." || name == ".." {
			return "", fmt.Errorf("invalid cache filename")
		}
	}
	return filepath.Join(dir, name), nil
}

// GetFlagHelp describes the plugin flags.
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		{Name: "openai.base-url", Type: "string", Default: defaultBaseURL, Description: "API base URL (or OPENAI_BASE_URL)"},
		{Name: "openai.protocol", Type: "string", Default: "images", Description: "images or chat (image-output extension)"},
		{Name: "openai.api-key-env", Type: "string", Default: "OPENAI_API_KEY", Description: "API key environment variable; empty disables authentication"},
		// Shared AI flags
		{Name: "ai.params", Type: "string", Default: "", Description: "JSON request-body overrides (recursive merge); field names depend on provider/model"},
		{Name: "ai.prompt", Type: "string", Default: "", Description: "Text description for AI image generation (required)", Required: true},
		{Name: "ai.model", Type: "string", Default: "auto", Description: "Model ID; required for custom endpoints", Required: false},
		{Name: "ai.list-models", Type: "bool", Default: "false", Description: "Show default model and selection guidance (offline)", Required: false},
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
		{Name: "trim-letterbox", Type: "bool", Default: "true", Description: "Trim solid letterbox borders from generated images", Required: false},
	}
}
