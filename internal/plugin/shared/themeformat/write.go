package themeformat

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF decoder
	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder
	"os"
	"strings"

	_ "golang.org/x/image/webp" // Register WebP decoder

	"gopkg.in/yaml.v3"
)

// WriteFile writes a theme to a markdown file.
func WriteFile(path string, theme *Theme, includeThumbnail bool, thumbnailData string) error {
	content, err := Generate(theme, includeThumbnail, thumbnailData)
	if err != nil {
		return fmt.Errorf("failed to generate markdown: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Generate creates the complete markdown content for a theme.
func Generate(theme *Theme, includeThumbnail bool, thumbnailData string) (string, error) {
	var buf bytes.Buffer

	// Write YAML front matter
	buf.WriteString("---\n")
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(theme); err != nil {
		return "", fmt.Errorf("failed to encode front matter: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("failed to close encoder: %w", err)
	}
	buf.WriteString("---\n\n")

	// Write markdown body
	buf.WriteString(generateMarkdownBody(theme, includeThumbnail, thumbnailData))

	return buf.String(), nil
}

// generateMarkdownBody creates the human-readable markdown content.
func generateMarkdownBody(theme *Theme, includeThumbnail bool, thumbnailData string) string {
	var buf bytes.Buffer

	// Title
	buf.WriteString(fmt.Sprintf("# %s\n\n", theme.Name))

	// Description/source
	if theme.Source != nil && theme.Source.Prompt != "" {
		buf.WriteString(fmt.Sprintf("> Generated with tinct from \"%s\"\n\n", theme.Source.Prompt))
	}

	// Thumbnail preview
	if includeThumbnail && thumbnailData != "" {
		buf.WriteString("## Preview\n\n")
		buf.WriteString(fmt.Sprintf("![Wallpaper Preview](data:image/jpeg;base64,%s)\n\n", thumbnailData))
	}

	// Color palette table
	buf.WriteString("## Color Palette\n\n")
	buf.WriteString("| Swatch | Hex | Role | Weight |\n")
	buf.WriteString("|--------|-----|------|--------|\n")

	for _, c := range theme.Colors {
		swatch := fmt.Sprintf("<span style=\"background:%s;color:%s\">████</span>", c.Hex, c.Hex)
		role := c.Role
		if role == "" {
			role = "-"
		}
		weight := fmt.Sprintf("%.1f%%", c.Weight*100)
		buf.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n", swatch, c.Hex, role, weight))
	}
	buf.WriteString("\n")

	// Ambient colors table (if present)
	if len(theme.Ambience) > 0 {
		buf.WriteString("## Ambient Colors\n\n")
		buf.WriteString("| Swatch | Hex | Region | Weight |\n")
		buf.WriteString("|--------|-----|--------|--------|\n")

		for _, a := range theme.Ambience {
			swatch := fmt.Sprintf("<span style=\"background:%s;color:%s\">████</span>", a.Hex, a.Hex)
			weight := fmt.Sprintf("%.1f%%", a.Weight*100)
			buf.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n", swatch, a.Hex, a.Region, weight))
		}
		buf.WriteString("\n")
	}

	// Wallpaper info
	buf.WriteString("## Wallpaper\n\n")
	if theme.Wallpaper != nil {
		if theme.Wallpaper.Embedded {
			size := len(theme.Wallpaper.Data)
			// Estimate actual size (base64 is ~1.33x original)
			actualSize := float64(size) * 0.75 / 1024 / 1024
			buf.WriteString(fmt.Sprintf("Embedded as base64 in front matter (%dx%d %s, %.1f MB)\n\n",
				theme.Wallpaper.Width, theme.Wallpaper.Height,
				strings.ToUpper(theme.Wallpaper.Format), actualSize))
		} else if theme.Wallpaper.Path != "" {
			buf.WriteString(fmt.Sprintf("External file: `%s`\n\n", theme.Wallpaper.Path))
		} else if theme.Wallpaper.URL != "" {
			buf.WriteString(fmt.Sprintf("External URL: %s\n\n", theme.Wallpaper.URL))
		}
	} else {
		buf.WriteString("No wallpaper embedded.\n\n")
	}

	// Source info
	if theme.Source != nil {
		buf.WriteString("## Source\n\n")
		buf.WriteString(fmt.Sprintf("- **Plugin**: %s\n", theme.Source.Plugin))
		if theme.Source.Model != "" {
			buf.WriteString(fmt.Sprintf("- **Model**: %s\n", theme.Source.Model))
		}
		if theme.Source.Prompt != "" {
			buf.WriteString(fmt.Sprintf("- **Prompt**: \"%s\"\n", theme.Source.Prompt))
		}
		buf.WriteString(fmt.Sprintf("- **Generated**: %s\n", theme.Generated.Format("2006-01-02T15:04:05Z07:00")))
	}

	return buf.String()
}

// EncodeWallpaper encodes image bytes to base64 and returns a Wallpaper struct.
func EncodeWallpaper(data []byte, format string) (*Wallpaper, error) {
	// Decode to get dimensions
	img, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	return &Wallpaper{
		Embedded: true,
		Format:   format,
		Width:    img.Width,
		Height:   img.Height,
		Data:     base64.StdEncoding.EncodeToString(data),
	}, nil
}

// ExternalWallpaper creates a Wallpaper struct for an external file reference.
func ExternalWallpaper(path string) *Wallpaper {
	return &Wallpaper{
		Embedded: false,
		Path:     path,
	}
}
