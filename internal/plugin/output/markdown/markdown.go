// Package markdown provides an output plugin for exporting themes to markdown format.
package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
	"github.com/jmylchreest/tinct/internal/plugin/shared/themeformat"
)

const version = "0.0.1"

// Plugin implements the output.Plugin interface for markdown theme export.
type Plugin struct {
	name             string
	outputDir        string
	noEmbed          bool
	includeThumbnail bool
	verbose          bool
}

// New creates a new markdown output plugin with default settings.
func New() *Plugin {
	return &Plugin{
		name:             "",
		outputDir:        "",
		noEmbed:          false,
		includeThumbnail: false,
		verbose:          false,
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "markdown"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Markdown theme file with embedded colors and wallpaper"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.name, "markdown.name", "", "Theme name (required, used in filename and metadata)")
	cmd.Flags().StringVar(&p.outputDir, "markdown.dir", "", "Output directory (default: ~/.config/tinct/themes)")
	cmd.Flags().BoolVar(&p.noEmbed, "markdown.no-embed", false, "Reference wallpaper externally instead of embedding")
	cmd.Flags().BoolVar(&p.includeThumbnail, "markdown.include-thumbnail", false, "Include a thumbnail preview in the markdown body")
}

// SetVerbose enables or disables verbose logging for the plugin.
func (p *Plugin) SetVerbose(verbose bool) {
	p.verbose = verbose
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []output.FlagHelp {
	return []output.FlagHelp{
		{Name: "markdown.name", Type: "string", Default: "", Description: "Theme name (required, used in filename and metadata)", Required: true},
		{Name: "markdown.dir", Type: "string", Default: "~/.config/tinct/themes", Description: "Output directory", Required: false},
		{Name: "markdown.no-embed", Type: "bool", Default: "false", Description: "Reference wallpaper externally instead of embedding", Required: false},
		{Name: "markdown.include-thumbnail", Type: "bool", Default: "false", Description: "Include a thumbnail preview in the markdown body", Required: false},
	}
}

// Validate checks if the plugin configuration is valid.
func (p *Plugin) Validate() error {
	if p.name == "" {
		return fmt.Errorf("--markdown.name is required")
	}
	return nil
}

// DefaultOutputDir returns the default output directory for this plugin.
func (p *Plugin) DefaultOutputDir() string {
	if p.outputDir != "" {
		return p.outputDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/tinct/themes"
	}
	return filepath.Join(home, ".config", "tinct", "themes")
}

// Generate creates a markdown theme file from the given theme data.
func (p *Plugin) Generate(themeData *colour.ThemeData) (map[string][]byte, error) {
	// Build theme structure
	theme := &themeformat.Theme{
		Name:      p.name,
		Version:   "1.0",
		Generator: fmt.Sprintf("tinct v%s", version),
		Generated: time.Now(),
		Source:    nil, // Could be populated from input plugin metadata in future
		Colors:    make([]themeformat.Color, 0),
		Ambience:  make([]themeformat.Ambience, 0),
	}

	// Extract colors from palette
	palette := themeData.Palette()
	if palette != nil {
		// Add categorised colors with roles
		// Separate position roles (ambient/region) from regular colors
		for role, cc := range palette.Colours {
			roleStr := string(role)
			if region, found := strings.CutPrefix(roleStr, "position"); found {
				// This is an ambient/region color (e.g. "positionTopLeft")
				// Convert camelCase to readable format (e.g. "TopLeft" -> "top_left")
				region = camelToSnake(region)
				theme.Ambience = append(theme.Ambience, themeformat.Ambience{
					Hex:    cc.Hex,
					Region: region,
					Weight: cc.Weight,
				})
			} else {
				// Regular color
				theme.Colors = append(theme.Colors, themeformat.Color{
					Hex:    cc.Hex,
					Weight: cc.Weight,
					Role:   roleStr,
				})
			}
		}
	}

	// Handle wallpaper
	var thumbnailData string
	if themeData.WallpaperPath != "" {
		if p.noEmbed {
			// Reference externally
			theme.Wallpaper = themeformat.ExternalWallpaper(themeData.WallpaperPath)
		} else {
			// Embed wallpaper
			data, err := os.ReadFile(themeData.WallpaperPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read wallpaper: %w", err)
			}

			// Detect format from file extension
			ext := strings.TrimPrefix(filepath.Ext(themeData.WallpaperPath), ".")
			if ext == "jpeg" {
				ext = "jpg"
			}

			wp, err := themeformat.EncodeWallpaper(data, ext)
			if err != nil {
				return nil, fmt.Errorf("failed to encode wallpaper: %w", err)
			}
			theme.Wallpaper = wp

			// Generate thumbnail if requested
			if p.includeThumbnail {
				thumb, err := themeformat.GenerateThumbnail(data, 800, 75)
				if err != nil {
					if p.verbose {
						fmt.Fprintf(os.Stderr, "Warning: failed to generate thumbnail: %v\n", err)
					}
				} else {
					thumbnailData = thumb
				}
			}
		}
	}

	// Generate markdown content
	content, err := themeformat.Generate(theme, p.includeThumbnail, thumbnailData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate markdown: %w", err)
	}

	// Generate filename from theme name
	filename := slugify(p.name) + ".md"

	return map[string][]byte{
		filename: []byte(content),
	}, nil
}

// camelToSnake converts camelCase to snake_case.
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// slugify converts a string to a URL-friendly slug.
func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)
	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	s = reg.ReplaceAllString(s, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	s = reg.ReplaceAllString(s, "-")
	// Trim hyphens from ends
	s = strings.Trim(s, "-")
	return s
}
