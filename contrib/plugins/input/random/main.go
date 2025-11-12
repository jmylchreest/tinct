// random - Random Colour Palette Generator (Tinct Input Plugin)
//
// Generates random colour palettes with configurable seed and colour count.
// Uses the go-plugin RPC protocol for better performance and process isolation.
//
// Features:
// - Generate 32 random colours by default (configurable via plugin args)
// - Deterministic generation with seed support for reproducibility
// - Process reuse across multiple invocations
// - Dry-run mode support
// - Verbose output option
//
// Build:
//   go build -o tinct-plugin-random
//
// Usage:
//   tinct plugins add ./random --type input
//   tinct plugins enable random
//   tinct generate -i random -o tailwind
//
// Plugin Args:
//   count: Number of colours to generate (default: 32)
//   seed: Random seed for reproducible generation
//
// Author: Tinct Contributors
// License: MIT

package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image/color"
	mathrand "math/rand/v2"
	"os"

	"github.com/hashicorp/go-plugin"

	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

// RandomPlugin implements the tinctplugin.InputPlugin interface.
type RandomPlugin struct{}

// Generate creates a random color palette.
func (p *RandomPlugin) Generate(ctx context.Context, opts tinctplugin.InputOptions) ([]color.Color, error) {
	// Extract configuration from plugin args
	seed := uint64(0)
	if seedArg, ok := opts.PluginArgs["seed"].(float64); ok {
		seed = uint64(seedArg)
	} else {
		// Generate a truly random seed from crypto/rand
		var randomBytes [8]byte
		if _, err := rand.Read(randomBytes[:]); err == nil {
			seed = binary.LittleEndian.Uint64(randomBytes[:])
		}
	}

	// Create a new random source with the seed
	var seedArray [32]byte
	binary.LittleEndian.PutUint64(seedArray[:8], seed)
	// #nosec G404 -- Using math/rand intentionally for deterministic color generation, not cryptography
	rng := mathrand.New(mathrand.NewChaCha8(seedArray))

	// Number of colors to generate (default 32)
	colorCount := 32
	if count, ok := opts.PluginArgs["count"].(float64); ok {
		colorCount = int(count)
	}

	// Handle dry-run mode
	if opts.DryRun {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "DRY-RUN MODE: Would generate %d random colours\n", colorCount)
			fmt.Fprintf(os.Stderr, "Random seed: %d\n", seed)
		}
		return []color.Color{}, nil
	}

	// Generate random colours
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Generating %d random colours (seed: %d)\n", colorCount, seed)
	}

	return generateRandomColors(colorCount, rng), nil
}

// GetMetadata returns plugin metadata.
func (p *RandomPlugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            "random",
		Type:            "input",
		Version:         "0.0.1",
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Generate random colour palettes with configurable seed and colour count",
		PluginProtocol:  "go-plugin",
	}
}

// WallpaperPath returns an empty string as random plugin doesn't provide wallpapers.
// This implements the required InputPlugin interface method.
func (p *RandomPlugin) WallpaperPath() string {
	return ""
}

// GetFlagHelp returns help information for plugin flags.
// This implements the required InputPlugin interface method.
func (p *RandomPlugin) GetFlagHelp() []tinctplugin.FlagHelp {
	return []tinctplugin.FlagHelp{
		{
			Name:        "count",
			Type:        "int",
			Default:     "32",
			Description: "Number of colours to generate",
			Required:    false,
		},
		{
			Name:        "seed",
			Type:        "uint64",
			Default:     "random",
			Description: "Random seed for reproducible generation",
			Required:    false,
		},
	}
}

// generateRandomColors creates n random colours.
func generateRandomColors(n int, rng *mathrand.Rand) []color.Color {
	colors := make([]color.Color, n)

	for i := range n {
		colors[i] = color.RGBA{
			// #nosec G115 -- rng.IntN(256) returns 0-255, safe for uint8
			R: uint8(rng.IntN(256)),
			// #nosec G115 -- rng.IntN(256) returns 0-255, safe for uint8
			G: uint8(rng.IntN(256)),
			// #nosec G115 -- rng.IntN(256) returns 0-255, safe for uint8
			B: uint8(rng.IntN(256)),
			A: 255,
		}
	}

	return colors
}

func main() {
	// Handle --plugin-info flag
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		p := &RandomPlugin{}
		info := p.GetMetadata()

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding plugin info: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Serve the plugin using go-plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: tinctplugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			"input": &tinctplugin.InputPluginRPC{
				Impl: &RandomPlugin{},
			},
		},
	})
}
