// random.go - Random Color Palette Generator (Tinct Input Plugin Example)
//
// This is an example Tinct input plugin written in Go. It demonstrates:
// - Plugin metadata via --plugin-info flag
// - Reading plugin arguments from JSON stdin
// - Generating random color palettes with semantic role assignments
// - Returning proper CategorisedPalette output
// - Supporting dry-run mode
// - Handling verbose output
// - Proper error handling
//
// This plugin generates a random color palette suitable for terminal
// themes, color schemes, and other applications. The colors are completely
// random, making this a simple example rather than a practical tool.
//
// Build:
//   go build -o random random.go
//
// Plugin Protocol:
//   # Get plugin metadata
//   ./random --plugin-info
//
// Direct Usage:
//   # Generate default 16-color palette
//   echo '{}' | ./random
//
//   # Generate with specific number of colors
//   echo '{"plugin_args":{"count":8}}' | ./random
//
//   # Generate with specific seed (reproducible)
//   echo '{"plugin_args":{"seed":12345}}' | ./random
//
//   # Dry-run mode (returns empty palette)
//   echo '{"dry_run":true}' | ./random
//
//   # Verbose mode (logs to stderr)
//   echo '{"verbose":true}' | ./random 2>&1
//
//   # Multiple arguments
//   echo '{"verbose":true,"plugin_args":{"count":24,"seed":99999}}' | ./random 2>&1
//
// Integration with Tinct:
//   # Add plugin to Tinct
//   tinct plugins add ./contrib/random
//
//   # Enable plugin
//   tinct plugins enable random
//
//   # Generate with default settings
//   tinct generate -i random -o tailwind
//
//   # Generate with custom arguments
//   tinct generate -i random -o hyprland --plugin-args 'random={"count":16,"seed":12345}'
//
//   # Preview the generated palette
//   tinct generate -i random --preview --plugin-args 'random={"seed":99999}'
//
//   # Save generated palette to file
//   tinct generate -i random --save-palette my-palette.json
//
//   # Verbose mode (shows plugin stderr)
//   tinct generate -i random -o tailwind -v
//
//   # Dry-run mode
//   tinct generate -i random -o tailwind --dry-run
//
// Plugin Arguments:
//   count (int): Number of colors to generate (default: 16)
//   seed  (int): Random seed for reproducibility (default: random)
//
// Output Format:
//   The plugin outputs a CategorisedPalette with:
//   - colours: Map of semantic roles (background, foreground, accent1-4, danger, etc.)
//   - all_colours: Array of all generated colors with metadata
//   - theme_type: 0=auto, 1=dark, 2=light
//
// Testing Reproducibility:
//   # Same seed should produce same colors
//   echo '{"plugin_args":{"seed":12345}}' | ./random > output1.json
//   echo '{"plugin_args":{"seed":12345}}' | ./random > output2.json
//   diff output1.json output2.json  # Should be identical
//
// Installation via Repository:
//   # From official plugin repository
//   tinct plugins repo add official https://raw.githubusercontent.com/jmylchreest/tinct-plugins/main/repository.json
//   tinct plugins install random
//
// Note: This is an example plugin demonstrating the protocol. Colors are
// completely random with no color theory or accessibility considerations.
// For production use, consider image-based extraction or color theory algorithms.
//
// Author: Tinct Contributors
// License: MIT

package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	mathrand "math/rand/v2"
	"os"
)

// PluginInfo represents the metadata returned by --plugin-info
type PluginInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
}

// RGB represents an RGB color (simple output format)
type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// InputOptions represents the options passed from Tinct
type InputOptions struct {
	Verbose         bool           `json:"verbose"`
	DryRun          bool           `json:"dry_run"`
	ColourOverrides []string       `json:"colour_overrides,omitempty"`
	PluginArgs      map[string]any `json:"plugin_args,omitempty"`
}

func main() {
	// Handle --plugin-info flag
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		info := PluginInfo{
			Name:        "random",
			Type:        "input",
			Version:     "1.0.0",
			Description: "Generate a random color palette (example Go plugin)",
			Author:      "Tinct Contributors",
		}

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding plugin info: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Read input options from stdin
	var opts InputOptions
	if err := json.NewDecoder(os.Stdin).Decode(&opts); err != nil {
		// If no stdin or invalid JSON, use defaults
		opts = InputOptions{
			PluginArgs: make(map[string]any),
			DryRun:     false,
			Verbose:    false,
		}
	}

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

	// Number of colors to generate
	colorCount := 16
	if count, ok := opts.PluginArgs["count"].(float64); ok {
		colorCount = int(count)
	}

	// Handle dry-run mode
	if opts.DryRun {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "DRY-RUN MODE: Would generate %d random colors\n", colorCount)
			fmt.Fprintf(os.Stderr, "Random seed: %d\n", seed)
		}

		// Output empty array for dry-run
		json.NewEncoder(os.Stdout).Encode([]RGB{})
		os.Exit(0)
	}

	// Generate random colors
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Generating %d random colors (seed: %d)\n", colorCount, seed)
	}

	colors := generateRandomColors(colorCount, rng)

	// Output as simple JSON array of RGB colors
	if err := json.NewEncoder(os.Stdout).Encode(colors); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding output: %v\n", err)
		os.Exit(1)
	}
}

// generateRandomColors creates n random colors as simple RGB values
func generateRandomColors(n int, rng *mathrand.Rand) []RGB {
	colors := make([]RGB, n)

	for i := range n {
		colors[i] = RGB{
			// #nosec G115 -- rng.IntN(256) returns 0-255, safe for uint8
			R: uint8(rng.IntN(256)),
			// #nosec G115 -- rng.IntN(256) returns 0-255, safe for uint8
			G: uint8(rng.IntN(256)),
			// #nosec G115 -- rng.IntN(256) returns 0-255, safe for uint8
			B: uint8(rng.IntN(256)),
		}
	}

	return colors
}
