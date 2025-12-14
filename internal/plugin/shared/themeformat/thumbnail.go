package themeformat

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif" // Register GIF decoder
	"image/jpeg"
	_ "image/png" // Register PNG decoder

	_ "golang.org/x/image/webp" // Register WebP decoder

	"golang.org/x/image/draw"
)

const (
	// DefaultThumbnailMaxWidth is the maximum width for thumbnails.
	DefaultThumbnailMaxWidth = 800
	// DefaultThumbnailQuality is the JPEG quality for thumbnails.
	DefaultThumbnailQuality = 75
)

// GenerateThumbnail creates a smaller preview image from the original.
// Returns base64-encoded JPEG data.
func GenerateThumbnail(imageData []byte, maxWidth int, quality int) (string, error) {
	if maxWidth <= 0 {
		maxWidth = DefaultThumbnailMaxWidth
	}
	if quality <= 0 {
		quality = DefaultThumbnailQuality
	}

	// Decode the original image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Calculate new dimensions maintaining aspect ratio
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	// Only resize if larger than max width
	newWidth := origWidth
	newHeight := origHeight
	if origWidth > maxWidth {
		ratio := float64(maxWidth) / float64(origWidth)
		newWidth = maxWidth
		newHeight = int(float64(origHeight) * ratio)
	}

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(resized, resized.Bounds(), img, bounds, draw.Over, nil)

	// Encode as JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: quality}); err != nil {
		return "", fmt.Errorf("failed to encode thumbnail: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
