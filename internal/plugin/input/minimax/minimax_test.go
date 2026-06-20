package minimax

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmylchreest/tinct/internal/plugin/input/shared/aiflags"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/commonflags"
	"github.com/spf13/cobra"
)

// resetFlags resets all shared flags to their defaults before each test.
func resetFlags() {
	aiflags.Reset()
	commonflags.Reset()
}

// pngBytes returns a tiny valid PNG image as bytes.
func pngBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test png: %v", err)
	}
	return buf.Bytes()
}

func TestNew(t *testing.T) {
	plugin := New()
	if plugin == nil {
		t.Fatal("New() returned nil")
	}
	if plugin.Name() != "minimax" {
		t.Errorf("Expected name 'minimax', got '%s'", plugin.Name())
	}
	if plugin.apiURL != defaultAPIURL {
		t.Errorf("Expected default apiURL '%s', got '%s'", defaultAPIURL, plugin.apiURL)
	}
}

func TestDescriptionAndVersion(t *testing.T) {
	plugin := New()
	if plugin.Description() == "" {
		t.Error("Description should not be empty")
	}
	if plugin.Version() == "" {
		t.Error("Version should not be empty")
	}
}

func TestValidate(t *testing.T) {
	resetFlags()
	defer resetFlags()

	plugin := New()

	// No prompt -> error.
	if err := plugin.Validate(); err == nil {
		t.Error("expected error when prompt is empty")
	}

	// With prompt -> ok.
	aiflags.Prompt = "a calm lake at dawn"
	if err := plugin.Validate(); err != nil {
		t.Errorf("unexpected error with prompt set: %v", err)
	}

	// list-models skips validation even without a prompt.
	resetFlags()
	aiflags.ListModels = true
	if err := plugin.Validate(); err != nil {
		t.Errorf("expected no error in list-models mode, got: %v", err)
	}
}

func TestRegisterFlags(t *testing.T) {
	resetFlags()
	defer resetFlags()

	plugin := New()
	cmd := &cobra.Command{Use: "test"}
	plugin.RegisterFlags(cmd)

	for _, name := range []string{"ai.prompt", "ai.model", "aspect-ratio", "minimax.base-url"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to be registered", name)
		}
	}
}

func TestGetFlagHelp(t *testing.T) {
	plugin := New()
	help := plugin.GetFlagHelp()
	if len(help) == 0 {
		t.Fatal("GetFlagHelp returned no entries")
	}
	found := false
	for _, h := range help {
		if h.Name == "minimax.base-url" {
			found = true
		}
	}
	if !found {
		t.Error("expected minimax.base-url in flag help")
	}
}

func TestGetExtensionForMIME(t *testing.T) {
	tests := map[string]string{
		"image/jpeg":    ".jpg",
		"image/jpg":     ".jpg",
		"image/png":     ".png",
		"image/webp":    ".webp",
		"image/gif":     ".gif",
		"application/x": ".jpg", // unknown -> jpg (MiniMax default)
	}
	for mime, want := range tests {
		if got := getExtensionForMIME(mime); got != want {
			t.Errorf("getExtensionForMIME(%q) = %q, want %q", mime, got, want)
		}
	}
}

func TestResolveDimensions(t *testing.T) {
	tests := []struct {
		aspect, size string
		wantW, wantH int
		wantOK       bool
	}{
		{"16:9", "2K", 2048, 1152, true},  // landscape max
		{"16:9", "1K", 1024, 576, true},   // 1K
		{"1:1", "2K", 2048, 2048, true},   // square max
		{"9:16", "2K", 1152, 2048, true},  // portrait
		{"21:9", "2K", 2048, 872, true},   // ultrawide, rounded to /8
		{"16:9", "", 2048, 1152, true},    // empty size -> 2K default
		{"16:9", "1600", 1600, 896, true}, // explicit px (896 = 900 rounded down to /8)
		{"bogus", "2K", 0, 0, false},      // unparseable aspect -> fall back
		{"16:9", "huge", 0, 0, false},     // unparseable size -> fall back
	}
	for _, tt := range tests {
		w, h, ok := resolveDimensions(tt.aspect, tt.size)
		if ok != tt.wantOK || w != tt.wantW || h != tt.wantH {
			t.Errorf("resolveDimensions(%q, %q) = (%d, %d, %v), want (%d, %d, %v)",
				tt.aspect, tt.size, w, h, ok, tt.wantW, tt.wantH, tt.wantOK)
		}
		// Dimensions must stay within MiniMax bounds and divisible by 8.
		if ok {
			for _, d := range []int{w, h} {
				if d < minImageDimension || d > maxImageDimension || d%8 != 0 {
					t.Errorf("resolveDimensions(%q, %q) produced invalid dimension %d", tt.aspect, tt.size, d)
				}
			}
		}
	}
}

func TestFindCachedImage(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "minimax-abc123")

	if got := findCachedImage(base); got != "" {
		t.Errorf("expected empty result for missing image, got %q", got)
	}

	want := base + ".jpg"
	if err := os.WriteFile(want, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := findCachedImage(base); got != want {
		t.Errorf("findCachedImage = %q, want %q", got, want)
	}
}

func TestExtractImageBytes(t *testing.T) {
	plugin := New()
	raw := pngBytes(t)

	// base64 path
	var resp imageResponse
	resp.Data.ImageBase64 = []string{base64.StdEncoding.EncodeToString(raw)}
	got, err := plugin.extractImageBytes(context.Background(), resp)
	if err != nil {
		t.Fatalf("base64 path: %v", err)
	}
	if !bytes.Equal(got, raw) {
		t.Error("base64 path returned mismatched bytes")
	}

	// no data path
	if _, err := plugin.extractImageBytes(context.Background(), imageResponse{}); err == nil {
		t.Error("expected error when response has no image data")
	}
}

func TestGenerateImageSuccess(t *testing.T) {
	resetFlags()
	defer resetFlags()

	raw := pngBytes(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("unexpected Authorization header: %q", got)
		}
		var req imageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Model == "" || req.Prompt == "" {
			t.Errorf("request missing model/prompt: %+v", req)
		}
		var out imageResponse
		out.Data.ImageBase64 = []string{base64.StdEncoding.EncodeToString(raw)}
		_ = json.NewEncoder(w).Encode(out)
	}))
	defer srv.Close()

	t.Setenv("MINIMAX_API_KEY", "test-key")
	aiflags.Prompt = "a quiet mountain road in fog"

	plugin := New()
	plugin.apiURL = srv.URL

	base := filepath.Join(t.TempDir(), "out")
	path, err := plugin.generateImage(context.Background(), defaultModel, base, false)
	if err != nil {
		t.Fatalf("generateImage failed: %v", err)
	}
	if path != base+".png" {
		t.Errorf("expected .png extension, got %q", path)
	}
	if !fileExists(path) {
		t.Errorf("expected image file written at %q", path)
	}
}

func TestGenerateImageBaseRespError(t *testing.T) {
	resetFlags()
	defer resetFlags()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		var out imageResponse
		out.BaseResp.StatusCode = 1008
		out.BaseResp.StatusMsg = "insufficient balance"
		_ = json.NewEncoder(w).Encode(out)
	}))
	defer srv.Close()

	t.Setenv("MINIMAX_API_KEY", "test-key")
	aiflags.Prompt = "x"

	plugin := New()
	plugin.apiURL = srv.URL

	base := filepath.Join(t.TempDir(), "out")
	if _, err := plugin.generateImage(context.Background(), defaultModel, base, false); err == nil {
		t.Error("expected error from non-zero base_resp.status_code")
	}
}

func TestGenerateImageHTTPError(t *testing.T) {
	resetFlags()
	defer resetFlags()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	t.Setenv("MINIMAX_API_KEY", "test-key")
	aiflags.Prompt = "x"

	plugin := New()
	plugin.apiURL = srv.URL

	base := filepath.Join(t.TempDir(), "out")
	if _, err := plugin.generateImage(context.Background(), defaultModel, base, false); err == nil {
		t.Error("expected error from HTTP 401")
	}
}

func TestGenerateImageMissingAPIKey(t *testing.T) {
	resetFlags()
	defer resetFlags()

	t.Setenv("MINIMAX_API_KEY", "")
	aiflags.Prompt = "x"

	plugin := New()
	base := filepath.Join(t.TempDir(), "out")
	if _, err := plugin.generateImage(context.Background(), defaultModel, base, false); err == nil {
		t.Error("expected error when MINIMAX_API_KEY is unset")
	}
}
