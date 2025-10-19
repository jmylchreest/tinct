// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"os"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/image"
	"github.com/spf13/cobra"
)

var (
	// Extract command flags
	extractColours     int
	extractAlgorithm   string
	extractFormat      string
	extractOutput      string
	extractCategorise  bool
	extractTheme       string
	extractShowPreview bool
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract <image>",
	Short: "Extract colour palette from an image",
	Long: `Extract a colour palette from an image using various algorithms.

The extract command analyses an image and generates a colour palette
that can be used for theming applications. You can control the number
of colours, the extraction algorithm, and the output format.

Supported image formats: JPEG, PNG, GIF, WebP

Examples:
  # Extract 16 colours (default) from an image
  tinct extract wallpaper.jpg

  # Extract 8 colours with preview
  tinct extract --preview --colours 8 wallpaper.png

  # Extract colours and output as JSON
  tinct extract --format json wallpaper.jpg

  # Extract colours and save to a file
  tinct extract --output palette.txt wallpaper.jpg

  # Extract 24 colours in RGB format
  tinct extract -c 24 -f rgb wallpaper.jpg

  # Categorise colours for theming (auto-detect theme type)
  tinct extract --categorise wallpaper.jpg

  # Categorise for dark theme (light text on dark background)
  tinct extract --categorise --theme dark wallpaper.jpg

  # Categorise with colour previews in terminal
  tinct extract --categorise --preview wallpaper.jpg`,
	Args: cobra.ExactArgs(1),
	RunE: runExtract,
}

func init() {
	// Define flags for the extract command
	extractCmd.Flags().IntVarP(&extractColours, "colours", "c", 16, "number of colours to extract (1-256)")
	extractCmd.Flags().StringVarP(&extractAlgorithm, "algorithm", "a", "kmeans", "extraction algorithm (kmeans)")
	extractCmd.Flags().StringVarP(&extractFormat, "format", "f", "hex", "output format (hex, rgb, json, categorised)")
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "", "output file (default: stdout)")
	extractCmd.Flags().BoolVar(&extractCategorise, "categorise", false, "categorise colours by role (background, foreground, accent)")
	extractCmd.Flags().StringVar(&extractTheme, "theme", "auto", "theme type for categorisation (auto, dark, light)")
	extractCmd.Flags().BoolVar(&extractShowPreview, "preview", false, "show colour previews in terminal")
}

// runExtract executes the extract command.
func runExtract(cmd *cobra.Command, args []string) error {
	imagePath := args[0]

	// Validate the image path
	if err := image.ValidateImagePath(imagePath); err != nil {
		return fmt.Errorf("invalid image path: %w", err)
	}

	// Validate configuration
	config := colour.ExtractorConfig{
		Algorithm:  colour.Algorithm(extractAlgorithm),
		ColorCount: extractColours,
	}
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Load the image
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Fprintf(os.Stderr, "Loading image: %s\n", imagePath)
	}

	loader := image.NewFileLoader()
	img, err := loader.Load(imagePath)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}

	if verbose {
		bounds := img.Bounds()
		fmt.Fprintf(os.Stderr, "Image loaded: %dx%d\n", bounds.Dx(), bounds.Dy())
	}

	// Create the colour extractor
	if verbose {
		fmt.Fprintf(os.Stderr, "Extracting %d colours using %s algorithm...\n", extractColours, extractAlgorithm)
	}

	extractor, err := colour.NewExtractor(config.Algorithm)
	if err != nil {
		return fmt.Errorf("failed to create extractor: %w", err)
	}

	// Extract the colour palette
	palette, err := extractor.Extract(img, extractColours)
	if err != nil {
		return fmt.Errorf("failed to extract colors: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Successfully extracted %d colours\n", palette.Len())
	}

	// Categorise colours if requested
	var output string
	if extractCategorise || extractFormat == "categorised" {
		if verbose {
			fmt.Fprintf(os.Stderr, "Categorising colours (theme: %s)...\n", extractTheme)
		}

		catConfig := colour.DefaultCategorisationConfig()
		switch extractTheme {
		case "dark":
			catConfig.ThemeType = colour.ThemeDark
		case "light":
			catConfig.ThemeType = colour.ThemeLight
		case "auto":
			catConfig.ThemeType = colour.ThemeAuto
		default:
			return fmt.Errorf("invalid theme type: %s (valid: auto, dark, light)", extractTheme)
		}

		categorised := colour.Categorise(palette, catConfig)

		var err error
		output, err = formatCategorisedPalette(categorised, extractFormat, extractShowPreview)
		if err != nil {
			return fmt.Errorf("failed to format categorised output: %w", err)
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Categorisation complete (theme type: %s)\n", categorised.ThemeType.String())
		}
	} else {
		// Format the output normally
		var err error
		output, err = formatPalette(palette, extractFormat, extractShowPreview)
		if err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	}

	// Write output to file or stdout
	if extractOutput != "" {
		if verbose {
			fmt.Fprintf(os.Stderr, "Writing output to: %s\n", extractOutput)
		}
		if err := os.WriteFile(extractOutput, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "Successfully wrote palette to %s\n", extractOutput)
		}
	} else {
		fmt.Print(output)
	}

	return nil
}

// formatPalette formats the palette according to the specified format.
func formatPalette(palette *colour.Palette, format string, showPreview bool) (string, error) {
	switch format {
	case "hex":
		return formatHex(palette, showPreview), nil
	case "rgb":
		return formatRGB(palette, showPreview), nil
	case "json":
		jsonBytes, err := palette.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to convert to JSON: %w", err)
		}
		return string(jsonBytes) + "\n", nil
	case "categorised":
		// This will be handled by formatCategorisedPalette
		return "", fmt.Errorf("categorised format requires --categorise flag")
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: hex, rgb, json, categorised)", format)
	}
}

// formatCategorisedPalette formats a categorised palette according to the specified format.
func formatCategorisedPalette(palette *colour.CategorisedPalette, format string, showPreview bool) (string, error) {
	switch format {
	case "hex", "rgb", "categorised", "":
		// Default categorised text format
		return palette.StringWithPreview(showPreview), nil
	case "json":
		jsonBytes, err := palette.ToJSON()
		if err != nil {
			return "", fmt.Errorf("failed to convert to JSON: %w", err)
		}
		return string(jsonBytes) + "\n", nil
	default:
		return "", fmt.Errorf("unsupported format for categorised output: %s (supported: json, categorised)", format)
	}
}

// formatHex formats the palette as hex colour codes.
func formatHex(palette *colour.Palette, showPreview bool) string {
	hexColours := palette.ToHex()
	output := ""
	for _, hex := range hexColours {
		if showPreview {
			rgb := colour.RGB{
				R: parseHexByte(hex[1:3]),
				G: parseHexByte(hex[3:5]),
				B: parseHexByte(hex[5:7]),
			}
			output += colour.FormatColourWithPreview(rgb, 8) + "\n"
		} else {
			output += hex + "\n"
		}
	}
	return output
}

// formatRGB formats the palette as RGB values.
func formatRGB(palette *colour.Palette, showPreview bool) string {
	rgbColours := palette.ToRGBSlice()
	output := ""
	for _, rgb := range rgbColours {
		if showPreview {
			output += colour.FormatColourWithPreview(rgb, 8) + "  " + rgb.String() + "\n"
		} else {
			output += rgb.String() + "\n"
		}
	}
	return output
}

// parseHexByte converts a two-character hex string to a byte.
func parseHexByte(s string) uint8 {
	var result uint8
	for i := 0; i < len(s); i++ {
		result *= 16
		c := s[i]
		if c >= '0' && c <= '9' {
			result += c - '0'
		} else if c >= 'a' && c <= 'f' {
			result += c - 'a' + 10
		} else if c >= 'A' && c <= 'F' {
			result += c - 'A' + 10
		}
	}
	return result
}
