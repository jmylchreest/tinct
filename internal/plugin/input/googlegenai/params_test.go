package googlegenai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
)

func TestParamsReachSDKWireBody(t *testing.T) {
	for _, model := range []string{defaultModel, modelImagen4} {
		t.Run(model, func(t *testing.T) {
			aiflags.Reset()
			commonflags.Reset()
			defer aiflags.Reset()
			defer commonflags.Reset()
			var imageData bytes.Buffer
			if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
				t.Fatal(err)
			}
			encoded := base64.StdEncoding.EncodeToString(imageData.Bytes())
			aiflags.Prompt = "landscape"
			aiflags.Params = `{"generationConfig":{"temperature":0.25},"parameters":{"sampleCount":2},"future_option":{"seed":9007199254740993}}`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if !bytes.Contains(body["future_option"], []byte("9007199254740993")) {
					t.Errorf("unknown field lost: %s", body["future_option"])
				}
				if !bytes.Contains(body["generationConfig"], []byte("0.25")) {
					t.Errorf("nested override lost: %s", body["generationConfig"])
				}
				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(r.URL.Path, ":generateContent") {
					if !bytes.Contains(body["generationConfig"], []byte("responseModalities")) {
						t.Error("nested default lost")
					}
					fmt.Fprintf(w, `{"candidates":[{"content":{"parts":[{"inlineData":{"mimeType":"image/png","data":%q}}]}}]}`, encoded)
				} else {
					if !bytes.Contains(body["parameters"], []byte(`"sampleCount":2`)) {
						t.Error("Imagen override lost")
					}
					fmt.Fprintf(w, `{"predictions":[{"bytesBase64Encoded":%q,"mimeType":"image/png"}]}`, encoded)
				}
			}))
			defer server.Close()
			t.Setenv("GOOGLE_GEMINI_BASE_URL", server.URL)
			t.Setenv("GOOGLE_API_KEY", "test")
			p := New()
			if model == modelImagen4 {
				p.backend = "vertex-ai"
				t.Setenv("GOOGLE_VERTEX_BASE_URL", server.URL)
				t.Setenv("GOOGLE_CLOUD_PROJECT", "")
				t.Setenv("GOOGLE_CLOUD_LOCATION", "")
			}
			if _, err := p.generateImage(context.Background(), model, filepath.Join(t.TempDir(), "params"), false); err != nil {
				t.Fatal(err)
			}
		})
	}
}
