package ai_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jmylchreest/tinct/pkg/ai"
)

var (
	_ ai.ImageGenerator = ai.OpenAI{}
	_ ai.ImageGenerator = ai.OpenRouter{}
	_ ai.ImageGenerator = ai.MiniMax{}
	_ ai.ImageGenerator = ai.Google{}
	_ ai.TextGenerator  = ai.OpenAIText{}
	_ ai.TextGenerator  = ai.Anthropic{}
)

func TestLocalOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/images/generations" || r.Method != http.MethodPost {
			t.Errorf("wrong endpoint: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("local endpoint should support no key")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["model"] != "local-model" || body["prompt"] != "a forest" || body["size"] != "1024x1024" {
			t.Errorf("wrong payload: %v", body)
		}
		fmt.Fprint(w, `{"data":[{"b64_json":"aW1hZ2U="}]}`)
	}))
	defer server.Close()
	image, err := (ai.OpenAI{Options: ai.Options{BaseURL: server.URL + "/v1/", Client: server.Client()}}).GenerateImage(context.Background(), ai.ImageRequest{Model: "local-model", Prompt: "a forest", Size: "1024x1024"})
	if err != nil || string(image.Bytes) != "image" {
		t.Fatalf("image=%v err=%v", image, err)
	}
}

func TestProtocols(t *testing.T) {
	for _, provider := range []string{"openrouter", "minimax", "anthropic", "openai-text"} {
		t.Run(provider, func(t *testing.T) {
			server := httptest.NewServer(protocolHandler(t, provider))
			defer server.Close()
			opts := ai.Options{BaseURL: server.URL, APIKey: "key"}
			var err error
			if provider == "anthropic" || provider == "openai-text" {
				var driver ai.TextGenerator = ai.Anthropic{Options: opts}
				if provider == "openai-text" {
					driver = ai.OpenAIText{Options: opts}
				}
				result, e := driver.GenerateText(context.Background(), ai.TextRequest{Model: "test-model", Prompt: "hello", MaxTokens: 100})
				err = e
				if result.Content != "hello world" {
					t.Errorf("wrong text: %v", result)
				}
			} else {
				var driver ai.ImageGenerator = ai.OpenRouter{Options: opts}
				if provider == "minimax" {
					driver = ai.MiniMax{Options: opts}
				}
				result, e := driver.GenerateImage(context.Background(), ai.ImageRequest{Model: "test-model", Prompt: "hello", AspectRatio: "16:9", Width: 1024, Height: 1024})
				err = e
				if string(result.Bytes) != "image" {
					t.Errorf("wrong image: %v", result)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResponseFailures(t *testing.T) {
	for _, body := range []string{`{}`, `{"data":[]}`, `{"data":[{"b64_json":"invalid!"}]}`, `{"error":{"message":"bad model"}}`, `not json`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, body) }))
			defer server.Close()
			_, err := (ai.OpenAI{Options: ai.Options{BaseURL: server.URL}}).GenerateImage(context.Background(), ai.ImageRequest{})
			if err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
}

func TestTransportBounds(t *testing.T) {
	t.Run("status without retries", func(t *testing.T) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls++
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprint(w, "secret body")
		}))
		defer server.Close()
		_, err := (ai.OpenAI{Options: ai.Options{BaseURL: server.URL}}).GenerateImage(context.Background(), ai.ImageRequest{})
		var apiErr *ai.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != 429 || calls != 1 {
			t.Fatalf("calls=%d err=%v", calls, err)
		}
		if strings.Contains(err.Error(), "secret") {
			t.Fatal("body leaked")
		}
	})
	t.Run("response limit", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, strings.Repeat("x", 65)) }))
		defer server.Close()
		_, err := (ai.OpenAI{Options: ai.Options{BaseURL: server.URL, MaxResponseBytes: 64}}).GenerateImage(context.Background(), ai.ImageRequest{})
		if err == nil || !strings.Contains(err.Error(), "exceeds") {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("deadline with injected client", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(time.Second):
			}
		}))
		defer server.Close()
		_, err := (ai.OpenAI{Options: ai.Options{BaseURL: server.URL, Client: server.Client(), Timeout: 20 * time.Millisecond}}).GenerateImage(context.Background(), ai.ImageRequest{})
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestDownloadsAndRedirectsDoNotForwardKey(t *testing.T) {
	download := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("key leaked to download")
		}
		fmt.Fprint(w, "image")
	}))
	defer download.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintf(w, `{"data":[{"url":%q}]}`, download.URL) }))
	defer server.Close()
	_, err := (ai.OpenAI{Options: ai.Options{BaseURL: server.URL, APIKey: "secret"}}).GenerateImage(context.Background(), ai.ImageRequest{})
	if err != nil {
		t.Fatal(err)
	}
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, download.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	_, err = (ai.OpenAI{Options: ai.Options{BaseURL: redirect.URL, APIKey: "secret"}}).GenerateImage(context.Background(), ai.ImageRequest{})
	var apiErr *ai.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 307 {
		t.Fatalf("redirect should fail: %v", err)
	}
}

func protocolHandler(t *testing.T, provider string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["model"] != "test-model" {
			t.Errorf("missing model: %v", body)
		}
		if provider == "anthropic" {
			if r.Header.Get("x-api-key") != "key" || r.Header.Get("anthropic-version") == "" || r.URL.Path != "/messages" {
				t.Error("invalid Anthropic protocol")
			}
			fmt.Fprint(w, `{"content":[{"type":"text","text":"hello"},{"type":"tool_use"},{"type":"text","text":" world"}],"stop_reason":"end_turn"}`)
			return
		}
		if r.Header.Get("Authorization") != "Bearer key" {
			t.Error("missing bearer key")
		}
		switch provider {
		case "openrouter":
			if r.URL.Path != "/chat/completions" || body["image_config"].(map[string]any)["aspect_ratio"] != "16:9" {
				t.Error("invalid image extension")
			}
			fmt.Fprint(w, `{"choices":[{"message":{"images":[{"image_url":{"url":"data:image/png;base64,aW1hZ2U="}}]},"finish_reason":"stop"}]}`)
		case "minimax":
			if r.URL.Path != "/image_generation" || body["width"] != float64(1024) || body["aspect_ratio"] != nil {
				t.Error("dimensions must override aspect ratio")
			}
			fmt.Fprint(w, `{"data":{"image_base64":["aW1hZ2U="]},"base_resp":{"status_code":0}}`)
		default:
			if r.URL.Path != "/chat/completions" {
				t.Error("invalid text endpoint")
			}
			fmt.Fprint(w, `{"choices":[{"message":{"content":"hello world"},"finish_reason":"stop"}]}`)
		}
	}
}

// A signed image URL is a credential, even though it is not the API key.
func TestTransportErrorDoesNotExposeSignedURL(t *testing.T) {
	transport := errorTransport{}
	_, err := (ai.OpenAI{Options: ai.Options{BaseURL: "http://localhost/v1", Client: &http.Client{Transport: transport}}}).GenerateImage(context.Background(), ai.ImageRequest{})
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unsafe error: %v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lost wrapped error: %v", err)
	}
}

type errorTransport struct{}

func (errorTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Method == http.MethodPost {
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"data":[{"url":"https://images.example/image?token=secret"}]}`)), Request: r}, nil
	}
	return nil, context.DeadlineExceeded
}
