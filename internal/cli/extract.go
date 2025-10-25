// Package cli provides the command-line interface for Tinct.
package cli

import (
	"fmt"
	"os"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/manager"
	"github.com/spf13/cobra"
)

var (
	// Extract command flags
	extractFormat      string
	extractOutput      string
	extractShowPreview bool
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract colour palette from an image",
	Long: `Extract a colour palette from an image using various algorithms.

The extract command analyses an image and generates a colour palette
that can be used for theming applications. Output formats include a
simple text format compatible with the file input plugin, as well as
hex, RGB, JSON, and categorised formats.

Supported image formats: JPEG, PNG, GIF, WebP

Examples:
  # Extract colours in palette format (default, file input plugin compatible)
  tinct extract -p wallpaper.jpg

  # Extract with preview
  tinct extract -p wallpaper.jpg --preview

  # Extract and save to file for later use with generate
  tinct extract -p wallpaper.jpg -o my-palette.txt

  # Extract as JSON
  tinct extract -p wallpaper.jpg --format json

  # Extract 24 colours in RGB format with preview
  tinct extract -p wallpaper.jpg -c 24 -f rgb --preview

  # Extract and show categorised output
  tinct extract -p wallpaper.jpg --format categorised --preview

  # Extract for dark theme
  tinct extract -p wallpaper.jpg --image.theme dark

Note: All role names use camelCase (e.g., backgroundMuted, accent1)`,
	Args: cobra.NoArgs,
	RunE: runExtract,
}

func init() {
	// Note: Plugin manager is initialised in root.go and flags are registered there
	// Define extract-specific flags
	extractCmd.Flags().StringVarP(&extractFormat, "format", "f", "palette", "output format (palette, hex, rgb, json, categorised)")
	extractCmd.Flags().StringVarP(&extractOutput, "output", "o", "", "output file (default: stdout)")
	extractCmd.Flags().BoolVar(&extractShowPreview, "preview", false, "show colour previews in terminal")
}

// runExtract executes the extract command.
func runExtract(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Reload plugin manager config from lock file if available (overrides env)
	lock, _, err := loadPluginLock()
	if err == nil && lock != nil {
		config := manager.Config{
			EnabledPlugins:  lock.EnabledPlugins,
			DisabledPlugins: lock.DisabledPlugins,
		}
		sharedPluginManager.UpdateConfig(config)
	}

	// Get image plugin from shared manager
	imagePlugin, ok := sharedPluginManager.GetInputPlugin("image")
	if !ok {
		return fmt.Errorf("image plugin not found")
	}

	// Check if plugin is enabled
	if !sharedPluginManager.IsInputEnabled(imagePlugin) {
		return fmt.Errorf("image plugin is currently disabled")
	}

	// Validate input plugin
	if err := imagePlugin.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Generate categorised palette using image plugin
	if verbose {
		fmt.Fprintf(os.Stderr, "Extracting colours from image...\n")
	}

	// Prepare options for input plugin
	inputOpts := input.GenerateOptions{
		Verbose:         verbose,
		DryRun:          false,
		ColourOverrides: []string{},
		PluginArgs:      make(map[string]any),
	}

	// Generate raw palette from input plugin
	palette, err := imagePlugin.Generate(ctx, inputOpts)
	if err != nil {
		return fmt.Errorf("failed to extract colours: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Successfully extracted %d colours\n", len(palette.Colors))
	}

	// Determine theme type from global flag
	themeType := colour.ThemeAuto
	switch globalTheme {
	case "dark":
		themeType = colour.ThemeDark
	case "light":
		themeType = colour.ThemeLight
	case "auto":
		// Check if plugin provides a theme hint (optional)
		if hinter, ok := imagePlugin.(input.ThemeHinter); ok {
			hint := hinter.ThemeHint()
			if verbose && hint != "" && hint != "auto" {
				fmt.Fprintf(os.Stderr, "Plugin suggests theme: %s\n", hint)
			}
			// Plugin hints are advisory only - we let the categorizer decide
			// based on weighted color distribution
		}
		themeType = colour.ThemeAuto
	}

	// Categorize the palette (auto-detection uses weighted color distribution)
	config := colour.DefaultCategorisationConfig()
	config.ThemeType = themeType
	categorised := colour.Categorise(palette, config)

	if verbose {
		fmt.Fprintf(os.Stderr, "Categorized palette with theme: %s\n", categorised.ThemeType.String())
	}

	// Format the output
	var output string
	switch extractFormat {
	case "palette":
		// File input plugin compatible format (role=hex)
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

// formatPaletteFile formats a categorised palette as a simple text file
// compatible with the file input plugin (role=hex format).
func formatPaletteFile(categorised *colour.CategorisedPalette) string {
	output := "# Tinct colour palette\n"
	output += "# Generated from image extraction\n"
	output += "# Format: role=hex\n"
	output += "# Use with: tinct generate --input file --file.path <this-file>\n\n"

	// Map of roles to output in order
	roles := []colour.ColourRole{
		colour.RoleBackground,
		colour.RoleBackgroundMuted,
		colour.RoleForeground,
		colour.RoleForegroundMuted,
		colour.RoleAccent1,
		colour.RoleAccent2,
		colour.RoleAccent3,
		colour.RoleAccent4,
		colour.RoleDanger,
		colour.RoleWarning,
		colour.RoleSuccess,
		colour.RoleInfo,
		colour.RoleNotification,
	}

	for _, role := range roles {
		if color, ok := categorised.Get(role); ok {
			output += fmt.Sprintf("%s=%s\n", role, color.Hex)
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
