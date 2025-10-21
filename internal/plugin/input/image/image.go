// Package image provides an input plugin for extracting colour palettes from images.
package image

import (
	"context"
	"fmt"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/image"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/spf13/cobra"
)

// Plugin implements the input.Plugin interface for image-based colour extraction.
type Plugin struct {
	path      string
	algorithm string
	colours   int
}

// New creates a new image input plugin with default settings.
func New() *Plugin {
	return &Plugin{
		algorithm: "kmeans",
		colours:   16,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "image"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Extract colour palette from an image file"
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&p.path, "image.path", "p", "", "Path to image file (required)")
	cmd.Flags().StringVarP(&p.algorithm, "image.algorithm", "a", "kmeans", "Extraction algorithm (kmeans)")
	cmd.Flags().IntVarP(&p.colours, "image.colours", "c", 16, "Number of colours to extract (1-256)")
}

// Validate checks if the plugin has all required inputs configured.
func (p *Plugin) Validate() error {
	if p.path == "" {
		return fmt.Errorf("image path is required (use --image.path or -p)")
	}
	if err := image.ValidateImagePath(p.path); err != nil {
		return fmt.Errorf("invalid image path: %w", err)
	}
	if p.colours < 1 || p.colours > 256 {
		return fmt.Errorf("colours must be between 1 and 256, got %d", p.colours)
	}
	return nil
}

// Generate creates a raw colour palette by extracting colours from the image.
// Returns only the extracted colors - categorization happens separately.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	// Load the image
	loader := image.NewFileLoader()
	img, err := loader.Load(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Create the colour extractor
	extractor, err := colour.NewExtractor(colour.Algorithm(p.algorithm))
	if err != nil {
		return nil, fmt.Errorf("failed to create extractor: %w", err)
	}

	// Extract and return the raw colour palette
	palette, err := extractor.Extract(img, p.colours)
	if err != nil {
		return nil, fmt.Errorf("failed to extract colours: %w", err)
	}

	return palette, nil
}
