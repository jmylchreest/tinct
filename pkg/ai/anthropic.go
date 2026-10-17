package ai

import (
	"context"
	"fmt"
	"strings"
)

// Anthropic implements text generation via Messages, not image generation.
type Anthropic struct{ Options Options }

func (p Anthropic) GenerateText(ctx context.Context, r TextRequest) (Text, error) {
	if r.MaxTokens <= 0 {
		return Text{}, fmt.Errorf("Anthropic MaxTokens must be positive")
	}
	payload := struct {
		Model     string    `json:"model"`
		Messages  []message `json:"messages"`
		MaxTokens int       `json:"max_tokens"`
	}{r.Model, []message{{Role: roleUser, Content: r.Prompt}}, r.MaxTokens}
	headers := map[string]string{"anthropic-version": "2023-06-01"}
	if p.Options.APIKey != "" {
		headers["x-api-key"] = p.Options.APIKey
	}
	var response struct {
		Content    []struct{ Type, Text string } `json:"content"`
		StopReason string                        `json:"stop_reason"`
	}
	if err := p.Options.post(ctx, p.Options.endpoint("https://api.anthropic.com/v1", "/messages"), payload, &response, headers); err != nil {
		return Text{}, err
	}
	var text strings.Builder
	for _, block := range response.Content {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
	}
	if text.Len() == 0 {
		return Text{}, fmt.Errorf("no text in response")
	}
	return Text{Content: text.String(), FinishReason: response.StopReason}, nil
}
