// Package imagepost provides post-processing for generated images, shared across the
// AI image input plugins. Its main job is removing the solid near-black letterbox bars
// that some providers (e.g. MiniMax image-01) bake into their output, then restoring the
// requested aspect ratio. It is stdlib-only — no image library dependency.
package imagepost

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"

	_ "image/gif" // register the GIF decoder so a provider's format always decodes
)

// Letterbox-detection tuning. A scan line (row/column) counts as a bar when at least
// barFraction of its sampled pixels are near-black (mean RGB < darkThreshold on 0–255),
// sampling every stride'th pixel for speed. maxTrimFraction caps how much of a dimension a
// trim may remove, so a genuinely dark image is never mistaken for one giant bar.
const (
	darkThreshold   = 12
	barFraction     = 0.97
	scanStride      = 4
	maxTrimFraction = 0.40
)

// TrimLetterboxFile removes solid near-black borders from the image at path and re-frames
// the remainder to targetAspect (width/height; <= 0 to skip re-framing), rewriting the file
// in place only when pixels actually change. An image with no bars and the right aspect is
// left byte-for-byte untouched. Unsupported formats (anything but JPEG/PNG) and decode
// errors are a no-op, so callers can apply it unconditionally. Returns whether the file was
// rewritten.
func TrimLetterboxFile(path string, targetAspect float64) (changed bool, err error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- path is a tinct-managed cache file
	if err != nil {
		return false, fmt.Errorf("failed to read image: %w", err)
	}

	trimmed, ok := trimLetterbox(raw, targetAspect)
	if !ok {
		return false, nil
	}

	if err := os.WriteFile(path, trimmed, 0o600); err != nil {
		return false, fmt.Errorf("failed to write trimmed image: %w", err)
	}
	return true, nil
}

// TrimLetterboxIfEnabled applies TrimLetterboxFile to path when enabled, logging to stderr
// in verbose mode. Trimming is a cosmetic post-process, so any error is non-fatal (logged
// when verbose, otherwise ignored). aspectRatio is a "W:H" string; an empty/invalid value
// skips re-framing.
func TrimLetterboxIfEnabled(path, aspectRatio string, enabled, verbose bool) {
	if !enabled {
		return
	}
	changed, err := TrimLetterboxFile(path, ParseAspect(aspectRatio))
	switch {
	case err != nil && verbose:
		fmt.Fprintf(os.Stderr, "Warning: letterbox trim failed: %v\n", err)
	case changed && verbose:
		fmt.Fprintf(os.Stderr, "Trimmed letterbox borders\n")
	}
}

// ParseAspect parses a "W:H" (or decimal) aspect ratio into a width/height float, or 0 when
// absent or invalid (which disables re-framing in TrimLetterboxFile).
func ParseAspect(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if i := strings.IndexByte(s, ':'); i > 0 {
		w, err1 := strconv.ParseFloat(strings.TrimSpace(s[:i]), 64)
		h, err2 := strconv.ParseFloat(strings.TrimSpace(s[i+1:]), 64)
		if err1 == nil && err2 == nil && h > 0 {
			return w / h
		}
		return 0
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 {
		return v
	}
	return 0
}

// trimLetterbox detects solid near-black borders, crops them, and center-crops the remainder
// to targetAspect. It returns the re-encoded image and true only when pixels changed.
func trimLetterbox(raw []byte, targetAspect float64) ([]byte, bool) {
	img, format, err := image.Decode(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png") {
		return nil, false
	}
	sub, ok := img.(subImager)
	if !ok {
		return nil, false
	}
	b := img.Bounds()
	left := barThickness(img, "left")
	right := barThickness(img, "right")
	top := barThickness(img, "top")
	bottom := barThickness(img, "bottom")

	rect := image.Rect(b.Min.X+left, b.Min.Y+top, b.Max.X-right, b.Max.Y-bottom)
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return nil, false
	}
	rect = reframe(rect, targetAspect)

	if rect == b {
		return nil, false // nothing to do — keep the original bytes (lossless)
	}
	cropped := sub.SubImage(rect)
	return encodeImage(cropped, format)
}

// subImager is implemented by the stdlib image types (YCbCr, RGBA, NRGBA, …); SubImage is a
// zero-copy view, so crop and re-frame allocate nothing.
type subImager interface {
	SubImage(image.Rectangle) image.Image
}

// reframe center-crops r's long axis so its aspect matches targetAspect (width/height).
// targetAspect <= 0 leaves r unchanged.
func reframe(r image.Rectangle, targetAspect float64) image.Rectangle {
	if targetAspect <= 0 {
		return r
	}
	w, h := r.Dx(), r.Dy()
	cur := float64(w) / float64(h)
	switch {
	case cur > targetAspect: // too wide → trim width
		nw := int(math.Round(float64(h) * targetAspect))
		off := (w - nw) / 2
		return image.Rect(r.Min.X+off, r.Min.Y, r.Min.X+off+nw, r.Max.Y)
	case cur < targetAspect: // too tall → trim height
		nh := int(math.Round(float64(w) / targetAspect))
		off := (h - nh) / 2
		return image.Rect(r.Min.X, r.Min.Y+off, r.Max.X, r.Min.Y+off+nh)
	default:
		return r
	}
}

// barThickness counts how many consecutive scan lines in from an edge are solid near-black,
// capped at maxTrimFraction of that dimension.
func barThickness(img image.Image, edge string) int {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	switch edge {
	case "top":
		return scanBars(max(1, int(float64(h)*maxTrimFraction)), func(i int) bool { return rowDark(img, b.Min.Y+i) })
	case "bottom":
		return scanBars(max(1, int(float64(h)*maxTrimFraction)), func(i int) bool { return rowDark(img, b.Max.Y-1-i) })
	case "left":
		return scanBars(max(1, int(float64(w)*maxTrimFraction)), func(i int) bool { return colDark(img, b.Min.X+i) })
	default: // right
		return scanBars(max(1, int(float64(w)*maxTrimFraction)), func(i int) bool { return colDark(img, b.Max.X-1-i) })
	}
}

func scanBars(limit int, dark func(int) bool) int {
	n := 0
	for n < limit && dark(n) {
		n++
	}
	return n
}

func rowDark(img image.Image, y int) bool {
	b := img.Bounds()
	total, darkN := 0, 0
	for x := b.Min.X; x < b.Max.X; x += scanStride {
		total++
		if pixelDark(img, x, y) {
			darkN++
		}
	}
	return total > 0 && float64(darkN)/float64(total) >= barFraction
}

func colDark(img image.Image, x int) bool {
	b := img.Bounds()
	total, darkN := 0, 0
	for y := b.Min.Y; y < b.Max.Y; y += scanStride {
		total++
		if pixelDark(img, x, y) {
			darkN++
		}
	}
	return total > 0 && float64(darkN)/float64(total) >= barFraction
}

func pixelDark(img image.Image, x, y int) bool {
	r, g, b, _ := img.At(x, y).RGBA() // 16-bit per channel
	mean := (int(r>>8) + int(g>>8) + int(b>>8)) / 3
	return mean < darkThreshold
}

func encodeImage(img image.Image, format string) ([]byte, bool) {
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, false
		}
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, false
		}
	default:
		return nil, false
	}
	return buf.Bytes(), true
}
