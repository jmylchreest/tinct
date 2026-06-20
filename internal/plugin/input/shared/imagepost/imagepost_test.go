package imagepost

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// makeLetterboxed builds a w×h image with solid black bars of `bar` px on top and bottom
// and a bright (non-dark) content band in the middle.
func makeLetterboxed(w, h, bar int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		c := color.RGBA{R: 0, G: 0, B: 0, A: 255} // black bar
		if y >= bar && y < h-bar {
			c = color.RGBA{R: 200, G: 180, B: 160, A: 255} // content
		}
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

func writePNG(t *testing.T, img image.Image) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "img.png")
	f, err := os.Create(path) // #nosec G304 -- test temp file
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return path
}

func decodeDims(t *testing.T, path string) (width, height int) {
	t.Helper()
	raw, err := os.ReadFile(path) // #nosec G304 -- test temp file
	if err != nil {
		t.Fatal(err)
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	return cfg.Width, cfg.Height
}

func TestTrimLetterboxFileRemovesBars(t *testing.T) {
	// 200×200 with 40px black bars top & bottom -> content is 200×120.
	path := writePNG(t, makeLetterboxed(200, 200, 40))

	// No re-framing: expect the bars gone (height 120), width unchanged.
	changed, err := TrimLetterboxFile(path, 0)
	if err != nil {
		t.Fatalf("TrimLetterboxFile: %v", err)
	}
	if !changed {
		t.Fatal("expected image to be trimmed")
	}
	w, h := decodeDims(t, path)
	if w != 200 || h != 120 {
		t.Errorf("after trim got %dx%d, want 200x120", w, h)
	}
}

func TestTrimLetterboxFileCleanImageUntouched(t *testing.T) {
	// Uniform bright image: no bars, square aspect -> no change.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := range 100 {
		for x := range 100 {
			img.Set(x, y, color.RGBA{R: 180, G: 180, B: 180, A: 255})
		}
	}
	path := writePNG(t, img)
	before, err := os.ReadFile(path) // #nosec G304 -- test temp file
	if err != nil {
		t.Fatal(err)
	}

	changed, err := TrimLetterboxFile(path, 1.0)
	if err != nil {
		t.Fatalf("TrimLetterboxFile: %v", err)
	}
	if changed {
		t.Error("expected clean image to be left unchanged")
	}
	after, _ := os.ReadFile(path) // #nosec G304 -- test temp file
	if !bytes.Equal(before, after) {
		t.Error("clean image bytes were rewritten")
	}
}

func TestParseAspect(t *testing.T) {
	tests := map[string]float64{
		"16:9": 16.0 / 9.0,
		"1:1":  1.0,
		"1.5":  1.5,
		"":     0,
		"bad":  0,
		"1:0":  0,
	}
	for in, want := range tests {
		if got := ParseAspect(in); got != want {
			t.Errorf("ParseAspect(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestTrimLetterboxIfEnabledDisabled(t *testing.T) {
	path := writePNG(t, makeLetterboxed(200, 200, 40))
	before, _ := os.ReadFile(path) // #nosec G304 -- test temp file

	// enabled=false must be a complete no-op even on a letterboxed image.
	TrimLetterboxIfEnabled(path, "16:9", false, false)

	after, _ := os.ReadFile(path) // #nosec G304 -- test temp file
	if !bytes.Equal(before, after) {
		t.Error("disabled trim modified the file")
	}
}
