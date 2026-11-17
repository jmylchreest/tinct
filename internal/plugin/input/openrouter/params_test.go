package openrouter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
)

type paramsTransport func(*http.Request) (*http.Response, error)

func (f paramsTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestParamsReachSharedTransport(t *testing.T) {
	aiflags.Reset()
	commonflags.Reset()
	defer aiflags.Reset()
	defer commonflags.Reset()
	aiflags.Prompt = "forest"
	aiflags.Params = `{"image_config":{"quality":"high"},"seed":9007199254740993}`
	t.Setenv("OPENROUTER_API_KEY", "test")
	var imageData bytes.Buffer
	if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	previous := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = previous })
	http.DefaultTransport = paramsTransport(func(r *http.Request) (*http.Response, error) {
		var body map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if string(body["seed"]) != "9007199254740993" {
			t.Error("seed lost")
		}
		if !bytes.Contains(body["image_config"], []byte("quality")) || !bytes.Contains(body["image_config"], []byte("aspect_ratio")) {
			t.Error("nested merge failed")
		}
		if r.URL.Path != "/api/v1/chat/completions" {
			t.Error(r.URL.Path)
		}
		raw := fmt.Sprintf(`{"choices":[{"message":{"images":[{"image_url":{"url":"data:image/png;base64,%s"}}]}}]}`, base64.StdEncoding.EncodeToString(imageData.Bytes()))
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(raw)), Header: make(http.Header)}, nil
	})
	if _, err := New().generateImage(context.Background(), "test", filepath.Join(t.TempDir(), "params"), false); err != nil {
		t.Fatal(err)
	}
}
