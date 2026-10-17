package ai

import (
	"context"
	"fmt"
)

// MiniMaxRequest is the provider's image request wire format.
type MiniMaxRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	AspectRatio    string `json:"aspect_ratio,omitempty"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
	ResponseFormat string `json:"response_format"`
	N              int    `json:"n"`
}

// MiniMaxResponse is the provider response, including errors returned with HTTP 200.
type MiniMaxResponse struct {
	Data struct {
		ImageBase64 []string `json:"image_base64"`
		ImageURLs   []string `json:"image_urls"`
	} `json:"data"`
	BaseResp struct {
		StatusCode int    `json:"status_code"`
		StatusMsg  string `json:"status_msg"`
	} `json:"base_resp"`
}

// MiniMax implements the MiniMax image generation protocol.
// Endpoint, when set, is the full URL; otherwise Options.BaseURL is used.
type MiniMax struct {
	Options  Options
	Endpoint string
}

func (p MiniMax) GenerateImage(ctx context.Context, r ImageRequest) (Image, error) {
	request := MiniMaxRequest{Model: r.Model, Prompt: r.Prompt, AspectRatio: r.AspectRatio, Width: r.Width, Height: r.Height, ResponseFormat: "base64", N: 1}
	if r.Width > 0 && r.Height > 0 {
		request.AspectRatio = ""
	}
	endpoint := p.Endpoint
	if endpoint == "" {
		endpoint = p.Options.endpoint("https://api.minimax.io/v1", "/image_generation")
	}
	var response MiniMaxResponse
	if err := p.Options.post(ctx, endpoint, request, &response, p.Options.bearer()); err != nil {
		return Image{}, err
	}
	return p.DecodeImage(ctx, response)
}

// DecodeImage validates provider status and obtains image bytes. Downloads never
// receive the API authorization header.
func (p MiniMax) DecodeImage(ctx context.Context, response MiniMaxResponse) (Image, error) {
	if response.BaseResp.StatusCode != 0 {
		return Image{}, fmt.Errorf("MiniMax error %d: %s", response.BaseResp.StatusCode, response.BaseResp.StatusMsg)
	}
	if len(response.Data.ImageBase64) > 0 {
		return decodeImage(response.Data.ImageBase64[0], "")
	}
	if len(response.Data.ImageURLs) > 0 {
		return p.Options.download(ctx, response.Data.ImageURLs[0])
	}
	return Image{}, fmt.Errorf("no image returned in response")
}
