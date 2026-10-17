package ai_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmylchreest/tinct/pkg/ai"
	"google.golang.org/genai"
)

func TestGoogleResponses(t *testing.T) {
	tests := []struct {
		name, body        string
		gemini, wantError bool
	}{
		{"imagen", `{"predictions":[{"bytesBase64Encoded":"aW1hZ2U=","mimeType":"image/png"}]}`, false, false},
		{"filtered", `{"predictions":[{"raiFilteredReason":"blocked"}]}`, false, true},
		{"empty imagen", `{}`, false, true},
		{"gemini", `{"candidates":[{"content":{"parts":[{"text":"hello"},{"inlineData":{"mimeType":"image/png","data":"aW1hZ2U="}}]}}]}`, true, false},
		{"empty gemini", `{"candidates":[{"content":{"parts":[{"text":"no image"}]}}]}`, true, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, test.body) }))
			defer server.Close()
			backend := genai.BackendGeminiAPI
			if !test.gemini {
				backend = genai.BackendVertexAI
			}
			client, err := genai.NewClient(context.Background(), &genai.ClientConfig{APIKey: "test", Backend: backend, HTTPOptions: genai.HTTPOptions{BaseURL: server.URL}})
			if err != nil {
				t.Fatal(err)
			}
			image, err := (ai.Google{Client: client, Gemini: test.gemini}).GenerateImage(context.Background(), ai.ImageRequest{Model: "test", Prompt: "forest"})
			if (err != nil) != test.wantError {
				t.Fatalf("unexpected error: %v", err)
			}
			if !test.wantError && (string(image.Bytes) != "image" || image.MIMEType != "image/png") {
				t.Fatalf("unexpected image: %v", image)
			}
		})
	}
}
