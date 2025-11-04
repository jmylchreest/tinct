// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/spf13/cobra"
)

var (
	// Extract command flags.
	extractInputPlugin string
	extractFormat      string
	extractOutput      string
	extractShowPreview bool
)

// extractCmd represents the extract command.
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract colour palette from various sources",
	Long: `Extract a colour palette from various sources and output in multiple formats.

The extract command generates a colour palette from images, files, or remote sources.
Output formats include a simple text format with role and position hints compatible
with the file input plugin, as well as hex, RGB, JSON, and categorised formats.

The extract command supports all the same input options as generate (ambient extraction,
colour overrides, etc.) but outputs the full palette with all hints instead of theme files.

Examples:
  # Extract from image in palette format (default, includes all role/position hints)
  tinct extract -i image -p wallpaper.jpg

  # Extract from image with ambient edge extraction
  tinct extract -i image -p wallpaper.jpg --image.extractAmbience

  # Extract 32 colours with preview
  tinct extract -i image -p wallpaper.jpg -c 32 --preview

  # Extract from HTTP(S) URL and save to file
  tinct extract -i image -p https://example.com/image.jpg -o my-palette.txt

  # Extract from remote JSON source
  tinct extract -i remote-json --remote-json.url https://example.com/palette.json

  # Extract as JSON with full metadata
  tinct extract -i image -p wallpaper.jpg --format json

  # Extract and show categorised output with preview
  tinct extract -i image -p wallpaper.jpg --format categorised --preview

  # Extract for dark theme
  tinct extract -i image -p wallpaper.jpg --theme dark

Note: All role names use camelCase (e.g., backgroundMuted, accent1)
      Position hints use camelCase (e.g., positionTopLeft, positionBottom)`,
	Args: cobra.NoArgs,
	RunE: runExtract,
}

func init() {
	// Note: Plugin manager is initialised in root.go and flags are registered there.
	// Define extract-specific flags.

	// Input plugin selection (required).
	extractCmd.Flags().StringVarP(&extractInputPlugin, "input", "i", "image", "Input plugin (image, file, remote-css, remote-json)")
	extractCmd.MarkFlagRequired("input")

	extractCmd.Flags().StringVarP(&extractFormat, "format", "f", "palette", "output format (palette, hex, rgb, json, categorised)")
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "", "output file (default: stdout)")
	extractCmd.Flags().BoolVar(&extractShowPreview, "preview", false, "show colour previews in terminal")
}

// runExtract executes the extract command.
func runExtract(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Reload plugin manager config from lock file if available (overrides env).
	lock, _, err := loadPluginLock()
	if err == nil && lock != nil {
		config := manager.Config{
			EnabledPlugins:  lock.EnabledPlugins,
			DisabledPlugins: lock.DisabledPlugins,
		}
		sharedPluginManager.UpdateConfig(config)
	}

	// Get input plugin from shared manager.
	inputPlugin, ok := sharedPluginManager.GetInputPlugin(extractInputPlugin)
	if !ok {
		availablePlugins := make([]string, 0)
		for pluginName := range sharedPluginManager.AllInputPlugins() {
			availablePlugins = append(availablePlugins, pluginName)
		}
		return fmt.Errorf("unknown input plugin: %s (available: %s)", extractInputPlugin, strings.Join(availablePlugins, ", "))
	}

	// Validate input plugin.
	if err := inputPlugin.Validate(); err != nil {
		return fmt.Errorf("input plugin validation failed: %w", err)
	}

	// Generate palette using input plugin.
	if verbose {
		fmt.Fprintf(os.Stderr, " Input plugin: %s\n", inputPlugin.Name())
		fmt.Fprintf(os.Stderr, "   %s\n", inputPlugin.Description())
	}

	// Prepare options for input plugin.
	inputOpts := input.GenerateOptions{
		Verbose:         verbose,
		DryRun:          false,
		ColourOverrides: []string{},
		PluginArgs:      make(map[string]any),
	}

	// Generate raw palette from input plugin.
	palette, err := inputPlugin.Generate(ctx, inputOpts)
	if err != nil {
		return fmt.Errorf("failed to extract colours: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Successfully extracted %d colours\n", len(palette.Colors))
	}

	// Determine theme type from global flag.
	themeType := colour.ThemeAuto
	switch globalTheme {
	case "dark":
		themeType = colour.ThemeDark
	case "light":
		themeType = colour.ThemeLight
	case "auto":
		// Check if plugin provides a theme hint (optional).
		if hinter, ok := inputPlugin.(input.ThemeHinter); ok {
			hint := hinter.ThemeHint()
			if verbose && hint != "" && hint != "auto" {
				fmt.Fprintf(os.Stderr, "Plugin suggests theme: %s\n", hint)
			}
			// Plugin hints are advisory only - we let the categorizer decide.
			// based on weighted color distribution.
		}
		themeType = colour.ThemeAuto
	}

	// Categorize the palette (auto-detection uses weighted color distribution).
	config := colour.DefaultCategorisationConfig()
	config.ThemeType = themeType
	categorised := colour.Categorise(palette, config)

	if verbose {
		fmt.Fprintf(os.Stderr, "Categorized palette with theme: %s\n", categorised.ThemeType.String())
	}

	// Format the output.
	var output string
	switch extractFormat {
	case "palette":
		// File input plugin compatible format (role=hex).
		output = formatPaletteFile(categorised)
	case "hex":
		output = formatHexFromCategorised(categorised, extractShowPreview)
	case "rgb":
		output = formatRGBFromCategorised(categorised, extractShowPreview)
	case "json":
		jsonBytes, err := categorised.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to convert to JSON: %w", err)
		}
		output = string(jsonBytes) + "\n"
	case "categorised":
		output = categorised.StringWithPreview(extractShowPreview)
	default:
		return fmt.Errorf("unsupported format: %s (supported: palette, hex, rgb, json, categorised)", extractFormat)
	}

	// Write output to file or stdout.
	if extractOutput != "" {
		if verbose {
			fmt.Fprintf(os.Stderr, "Writing output to: %s\n", extractOutput)
		}
		if err := os.WriteFile(extractOutput, []byte(output), 0o600); err != nil {
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

// formatPaletteFile formats a categorised palette as a simple text file.
// with all role and position hints, compatible with the file input plugin.
// Colors with role/position hints are listed first, then indexed colors (colourN).
func formatPaletteFile(categorised *colour.CategorisedPalette) string {
	output := "# Tinct colour palette\n"
	output += "# Generated from extraction\n"
	output += "# Format: role=hex or colourN=hex\n"
	output += "# Use with: tinct generate --input file --file.path <this-file>\n\n"

	// First pass: output all colors with role/position hints.
	for _, color := range categorised.AllColours {
		if color.Role != "" {
			output += fmt.Sprintf("%s=%s\n", color.Role, color.Hex)
		}
	}

	// Second pass: output indexed colors without role hints.
	for _, color := range categorised.AllColours {
		if color.Role == "" {
			output += fmt.Sprintf("colour%d=%s\n", color.Index, color.Hex)
		}
	}

	return output
}

// formatHexFromCategorised formats a categorised palette as hex colour codes.
// If showPreview is true, color blocks are displayed before each hex value.
func formatHexFromCategorised(categorised *colour.CategorisedPalette, showPreview bool) string {
	output := ""
	for _, color := range categorised.AllColours {
		if showPreview {
			output += colour.FormatColourWithPreview(color.RGB, 8) + "\n"
		} else {
			output += color.Hex + "\n"
		}
	}
	return output
}

// formatRGBFromCategorised formats a categorised palette as RGB values.
// If showPreview is true, color blocks are displayed before each RGB value.
func formatRGBFromCategorised(categorised *colour.CategorisedPalette, showPreview bool) string {
	output := ""
	for _, color := range categorised.AllColours {
		if showPreview {
			output += colour.FormatColourWithPreview(color.RGB, 8) + "  " + color.RGB.String() + "\n"
		} else {
			output += color.RGB.String() + "\n"
		}
	}
	return output
}
