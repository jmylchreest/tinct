package ai

import (
	"context"
	"fmt"
)

// OpenAI implements the Images API, including compatible local endpoints.
// Configure BaseURL explicitly for a local server that supports /images/generations.
// Chat-only compatibility does not imply support for this capability.
type OpenAI struct{ Options Options }

func (p OpenAI) GenerateImage(ctx context.Context, r ImageRequest) (Image, error) {
	payload := struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
		Size   string `json:"size,omitempty"`
		N      int    `json:"n"`
	}{r.Model, r.Prompt, r.Size, 1}
	var response struct {
		Data []struct {
			Base64 string `json:"b64_json"`
			URL    string `json:"url"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := p.Options.post(ctx, p.Options.endpoint("https://api.openai.com/v1", "/images/generations"), payload, &response, p.Options.bearer()); err != nil {
		return Image{}, err
	}
	if response.Error != nil {
		return Image{}, fmt.Errorf("image generation: %s", response.Error.Message)
	}
	if len(response.Data) == 0 {
		return Image{}, fmt.Errorf("no images in response")
	}
	image := response.Data[0]
	if image.Base64 != "" {
		return decodeImage(image.Base64, "")
	}
	if image.URL != "" {
		return p.Options.download(ctx, image.URL)
	}
	return Image{}, fmt.Errorf("no image data in response")
}

// OpenAIText implements the non-streaming Chat Completions text protocol.
// It can be used independently of the Images API with compatible local servers.
type OpenAIText struct{ Options Options }

func (p OpenAIText) GenerateText(ctx context.Context, r TextRequest) (Text, error) {
	payload := struct {
		Model     string    `json:"model"`
		Messages  []message `json:"messages"`
		MaxTokens int       `json:"max_tokens,omitempty"`
	}{r.Model, []message{{Role: roleUser, Content: r.Prompt}}, r.MaxTokens}
	var response struct {
		Choices []struct {
			Message      message `json:"message"`
			FinishReason string  `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := p.Options.post(ctx, p.Options.endpoint("https://api.openai.com/v1", "/chat/completions"), payload, &response, p.Options.bearer()); err != nil {
		return Text{}, err
	}
	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return Text{}, fmt.Errorf("no text in response")
	}
	return Text{Content: response.Choices[0].Message.Content, FinishReason: response.Choices[0].FinishReason}, nil
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
