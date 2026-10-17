package ai

import (
	"context"
	"fmt"
	"google.golang.org/genai"
	"strings"
)

// Google wraps a caller-configured SDK client (Gemini API or Vertex AI).
// Gemini selects GenerateContent; otherwise GenerateImages is used.
// ImageSize and NegativePrompt support are model/backend dependent. The caller
// must only supply options its selected model supports.
type Google struct {
	Client *genai.Client
	Gemini bool
}

func (p Google) GenerateImage(ctx context.Context, r ImageRequest) (Image, error) {
	if p.Client == nil {
		return Image{}, fmt.Errorf("Google client is required")
	}
	ctx, cancel := (Options{}).deadline(ctx)
	defer cancel()
	if p.Gemini {
		return p.gemini(ctx, r)
	}
	response, err := p.Client.Models.GenerateImages(ctx, r.Model, r.Prompt, &genai.GenerateImagesConfig{NumberOfImages: 1, AspectRatio: r.AspectRatio, ImageSize: r.Size, NegativePrompt: r.NegativePrompt, OutputMIMEType: "image/png"})
	if err != nil {
		return Image{}, fmt.Errorf("Google image generation: %w", err)
	}
	if response == nil || len(response.GeneratedImages) == 0 || response.GeneratedImages[0] == nil {
		return Image{}, fmt.Errorf("no images generated in response")
	}
	generated := response.GeneratedImages[0]
	if generated.RAIFilteredReason != "" {
		return Image{}, fmt.Errorf("image filtered: %s", generated.RAIFilteredReason)
	}
	if generated.Image == nil || len(generated.Image.ImageBytes) == 0 {
		return Image{}, fmt.Errorf("no image data in response")
	}
	image := Image{Bytes: generated.Image.ImageBytes, MIMEType: generated.Image.MIMEType}
	if generated.SafetyAttributes != nil {
		image.SafetyAttributes = map[string]any{"categories": generated.SafetyAttributes.Categories, "scores": generated.SafetyAttributes.Scores}
	}
	return image, nil
}
func (p Google) gemini(ctx context.Context, r ImageRequest) (Image, error) {
	prompt := r.Prompt
	if r.AspectRatio != "" {
		prompt = fmt.Sprintf("Generate an image with aspect ratio %s: %s", r.AspectRatio, prompt)
	}
	response, err := p.Client.Models.GenerateContent(ctx, r.Model, genai.Text(prompt), &genai.GenerateContentConfig{ResponseModalities: []string{"Image"}})
	if err != nil {
		return Image{}, fmt.Errorf("Google image generation: %w", err)
	}
	if response == nil || len(response.Candidates) == 0 || response.Candidates[0] == nil || response.Candidates[0].Content == nil {
		return Image{}, fmt.Errorf("no image data in response")
	}
	for _, part := range response.Candidates[0].Content.Parts {
		if part != nil && part.InlineData != nil && len(part.InlineData.Data) > 0 && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
			return Image{Bytes: part.InlineData.Data, MIMEType: part.InlineData.MIMEType}, nil
		}
	}
	return Image{}, fmt.Errorf("no inline image data in response")
}
