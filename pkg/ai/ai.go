// Package ai provides generation drivers shared by built-in and external plugins.
// Callers own credentials, prompts, caching, model selection and file output.
package ai

import (
	"context"
	"net/http"
	"time"
)

// ImageGenerator is implemented by providers that produce image bytes.
type ImageGenerator interface {
	GenerateImage(context.Context, ImageRequest) (Image, error)
}

// TextGenerator is separate because text providers need not support images.
type TextGenerator interface {
	GenerateText(context.Context, TextRequest) (Text, error)
}

// ImageRequest describes one generation. Size is a provider-specific size string;
// Width and Height are used by MiniMax, AspectRatio by OpenRouter and Google.
type ImageRequest struct {
	Model, Prompt, NegativePrompt, AspectRatio, Size string
	Width, Height                                    int
}

// Image contains the first generated image and optional provider metadata.
type Image struct {
	Bytes                  []byte
	MIMEType, FinishReason string
	SafetyAttributes       map[string]any
}

// TextRequest describes a non-streaming text generation.
type TextRequest struct {
	Model, Prompt string
	MaxTokens     int
}

// Text contains generated text and a provider stop reason.
type Text struct{ Content, FinishReason string }

// Options configures HTTP providers. BaseURL includes the API prefix (e.g. /v1).
// APIKey may be empty for local servers. Zero limits use bounded defaults.
// No environment variables are read and generation POSTs are never retried.
type Options struct {
	BaseURL, APIKey  string
	Client           *http.Client
	Timeout          time.Duration
	MaxResponseBytes int64
}

const roleUser = "user"
