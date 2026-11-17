package ai_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/pkg/ai"
)

func TestImageDriversRejectStreaming(t *testing.T) {
	for name, driver := range map[string]ai.ImageGenerator{
		"openai": ai.OpenAI{}, "openrouter": ai.OpenRouter{}, "minimax": ai.MiniMax{}, "google": ai.Google{},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := driver.GenerateImage(context.Background(), ai.ImageRequest{Params: map[string]any{"stream": true}})
			if err == nil || !strings.Contains(err.Error(), "stream") {
				t.Fatalf("streaming not rejected: %v", err)
			}
		})
	}
}

func TestOpenRouterImageURL(t *testing.T) {
	download := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("API credentials leaked to image download")
		}
		fmt.Fprint(w, "image bytes")
	}))
	defer download.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"choices":[{"finish_reason":"stop","message":{"images":[{"image_url":{"url":%q}}]}}]}`, download.URL)
	}))
	defer server.Close()
	result, err := (ai.OpenRouter{Options: ai.Options{BaseURL: server.URL, APIKey: "secret"}}).GenerateImage(context.Background(), ai.ImageRequest{Model: "test", Prompt: "forest"})
	if err != nil {
		t.Fatal(err)
	}
	if string(result.Bytes) != "image bytes" || result.FinishReason != "stop" {
		t.Fatalf("wrong result: %+v", result)
	}
}
