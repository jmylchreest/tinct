// Test image generator for creating sample images for testing color extraction
package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	// Create a simple test image with distinct color blocks
	width := 400
	height := 400
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Define test colors
	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},     // Red
		{R: 0, G: 255, B: 0, A: 255},     // Green
		{R: 0, G: 0, B: 255, A: 255},     // Blue
		{R: 255, G: 255, B: 0, A: 255},   // Yellow
		{R: 255, G: 0, B: 255, A: 255},   // Magenta
		{R: 0, G: 255, B: 255, A: 255},   // Cyan
		{R: 128, G: 128, B: 128, A: 255}, // Gray
		{R: 255, G: 128, B: 0, A: 255},   // Orange
	}

	// Fill image with color blocks (2x4 grid)
	blockWidth := width / 2
	blockHeight := height / 4

	colorIndex := 0
	for row := 0; row < 4; row++ {
		for col := 0; col < 2; col++ {
			c := colors[colorIndex]
			colorIndex++

			// Fill the block
			for y := row * blockHeight; y < (row+1)*blockHeight; y++ {
				for x := col * blockWidth; x < (col+1)*blockWidth; x++ {
					img.Set(x, y, c)
				}
			}
		}
	}

	// Save the image
	file, err := os.Create("testdata/sample.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		panic(err)
	}

	println("Test image created: testdata/sample.png")
}
