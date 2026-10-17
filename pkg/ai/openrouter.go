package ai

import (
	"context"
	"fmt"
	"strings"
)

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
	ID      string           `json:"id"`
	Model   string           `json:"model"`
	Choices []Choice         `json:"choices"`
	Error   *OpenRouterError `json:"error,omitempty"`
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

// OpenRouterError represents an API error response.
type OpenRouterError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    any    `json:"code"` // Can be string or number depending on provider
}

// OpenRouter implements its chat image extension, not the OpenAI Images API.
type OpenRouter struct{ Options Options }

func (p OpenRouter) GenerateImage(ctx context.Context, r ImageRequest) (Image, error) {
	request := ChatCompletionRequest{Model: r.Model, Messages: []Message{{Role: roleUser, Content: r.Prompt}}, Modalities: []string{"image", "text"}}
	if r.AspectRatio != "" {
		request.ImageConfig = &ImageConfig{AspectRatio: r.AspectRatio}
	}
	headers := p.Options.bearer()
	headers["HTTP-Referer"] = "https://github.com/jmylchreest/tinct"
	headers["X-Title"] = "tinct"
	var response ChatCompletionResponse
	if err := p.Options.post(ctx, p.Options.endpoint("https://openrouter.ai/api/v1", "/chat/completions"), request, &response, headers); err != nil {
		return Image{}, err
	}
	if response.Error != nil {
		return Image{}, fmt.Errorf("image generation: %s", response.Error.Message)
	}
	if len(response.Choices) == 0 || len(response.Choices[0].Message.Images) == 0 {
		return Image{}, fmt.Errorf("no images in response")
	}
	choice := response.Choices[0]
	header, data, ok := strings.Cut(choice.Message.Images[0].ImageURL.URL, ",")
	if !ok || !strings.HasPrefix(header, "data:image/") || !strings.HasSuffix(header, ";base64") {
		return Image{}, fmt.Errorf("expected base64 image data URL")
	}
	mime := strings.TrimSuffix(strings.TrimPrefix(header, "data:"), ";base64")
	image, err := decodeImage(data, mime)
	if err != nil {
		return Image{}, err
	}
	image.FinishReason = choice.FinishReason
	return image, nil
}
