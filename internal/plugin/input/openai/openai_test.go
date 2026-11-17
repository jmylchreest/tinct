package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
	"github.com/spf13/cobra"
)

func setup(t *testing.T) {
	t.Helper()
	aiflags.Reset()
	commonflags.Reset()
	t.Cleanup(func() { aiflags.Reset(); commonflags.Reset() })
	t.Setenv("OPENAI_BASE_URL", "")
	t.Setenv("OPENAI_API_KEY", "")
	commonflags.CacheDir = t.TempDir()
	commonflags.Count = 2
	commonflags.TrimLetterbox = false
	aiflags.Prompt = "forest"
	aiflags.Model = "local-image"
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 16), G: uint8(y * 16), B: 100, A: 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestGenerateCacheAndOverrides(t *testing.T) {
	setup(t)
	encoded := base64.StdEncoding.EncodeToString(testPNG(t))
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "" {
			t.Error("local server received credentials")
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request["quality"] != "high" || request["size"] != "1024x1024" {
			t.Errorf("missing overrides: %v", request)
		}
		fmt.Fprintf(w, `{"data":[{"b64_json":%q}]}`, encoded)
	}))
	defer server.Close()
	p := New()
	p.baseURL = server.URL + "/v1"
	p.keyEnv = ""
	aiflags.Params = `{"quality":"high","size":"1024x1024","seed":9007199254740993}`
	palette, err := p.Generate(context.Background(), input.GenerateOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(palette.Colors) == 0 || p.WallpaperPath() == "" {
		t.Fatal("missing image or palette")
	}
	first := p.WallpaperPath()
	if p.WallpaperRawPath() != first {
		t.Fatal("raw path mismatch")
	}
	if _, err := p.Generate(context.Background(), input.GenerateOptions{}); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("cache miss: %d calls", calls)
	}
	aiflags.Params = `{"quality":"high","size":"1024x1024","seed":9007199254740994}`
	if _, err := p.Generate(context.Background(), input.GenerateOptions{}); err != nil {
		t.Fatal(err)
	}
	if calls != 2 || first == p.WallpaperPath() {
		t.Fatal("params did not invalidate cache")
	}
	metadata, err := os.ReadFile(strings.TrimSuffix(p.WallpaperPath(), filepath.Ext(p.WallpaperPath())) + ".json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(metadata, []byte("9007199254740994")) {
		t.Fatalf("metadata lost precision: %s", metadata)
	}
}

func TestValidationFlagsAndDryRun(t *testing.T) {
	setup(t)
	p := New()
	cmd := &cobra.Command{}
	p.RegisterFlags(cmd)
	aiflags.Prompt = "forest"
	commonflags.CacheDir = t.TempDir()
	for _, flag := range []string{"ai.params", "openai.base-url", "openai.protocol", "openai.api-key-env"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Error(flag)
		}
	}
	if len(p.GetFlagHelp()) == 0 || p.Name() != "openai" || p.Description() == "" {
		t.Fatal("missing plugin metadata")
	}
	p.baseURL = "http://localhost:1/v1"
	aiflags.Model = "auto"
	if err := p.Validate(); err == nil {
		t.Fatal("custom endpoint accepted auto model")
	}
	aiflags.Model = "test"
	aiflags.Params = "[]"
	if _, err := p.Generate(context.Background(), input.GenerateOptions{DryRun: true}); err == nil {
		t.Fatal("invalid params accepted")
	}
	aiflags.Params = "{}"
	if _, err := p.Generate(context.Background(), input.GenerateOptions{DryRun: true}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(commonflags.CacheDir)
	if err != nil || len(entries) != 0 {
		t.Fatal("dry run wrote cache")
	}
	p.protocol = "anthropic"
	if err := p.Validate(); err == nil {
		t.Fatal("unsupported protocol accepted")
	}
	p.protocol = "chat"
	raw, err := p.request("model")
	if err != nil || !bytes.Contains(raw, []byte(`"messages"`)) || bytes.Contains(raw, []byte(`"prompt":`)) {
		t.Fatalf("bad chat request %s %v", raw, err)
	}
}

func TestInvalidImageAndNoCache(t *testing.T) {
	setup(t)
	encoded := "bm90LWFuLWltYWdl"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, `{"data":[{"b64_json":%q}]}`, encoded) }))
	defer server.Close()
	p := New()
	p.baseURL = server.URL
	if _, err := p.Generate(context.Background(), input.GenerateOptions{}); err == nil {
		t.Fatal("invalid image accepted")
	}
	encoded = base64.StdEncoding.EncodeToString(testPNG(t))
	commonflags.CacheEnabled = false
	if _, err := p.Generate(context.Background(), input.GenerateOptions{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(p.WallpaperPath())) })
	if _, err := os.Stat(p.WallpaperPath()); err != nil {
		t.Fatal(err)
	}
}
