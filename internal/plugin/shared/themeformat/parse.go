package themeformat

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses a markdown theme file.
func ParseFile(path string) (*Theme, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- User-provided theme file path is intentional
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return Parse(data)
}

// Parse parses markdown theme data from bytes.
func Parse(data []byte) (*Theme, error) {
	// Extract YAML front matter
	frontMatter, err := extractFrontMatter(data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract front matter: %w", err)
	}

	var theme Theme
	if err := yaml.Unmarshal(frontMatter, &theme); err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %w", err)
	}

	return &theme, nil
}

// extractFrontMatter extracts YAML front matter from markdown content.
func extractFrontMatter(data []byte) ([]byte, error) {
	reader := bufio.NewReader(bytes.NewReader(data))

	// Read first line - must be "---"
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read first line: %w", err)
	}
	if strings.TrimSpace(line) != "---" {
		return nil, fmt.Errorf("file must start with YAML front matter (---)")
	}

	// Read until closing "---"
	var frontMatter bytes.Buffer
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return nil, fmt.Errorf("unclosed front matter (missing closing ---)")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read front matter: %w", err)
		}
		if strings.TrimSpace(line) == "---" {
			break
		}
		frontMatter.WriteString(line)
	}

	return frontMatter.Bytes(), nil
}

// DecodeWallpaper extracts wallpaper bytes from a theme.
// For embedded wallpapers, decodes the base64 data.
// For external references, reads from path or URL.
func DecodeWallpaper(theme *Theme, basePath string) ([]byte, string, error) {
	if theme.Wallpaper == nil {
		return nil, "", fmt.Errorf("no wallpaper in theme")
	}

	wp := theme.Wallpaper

	if wp.Embedded && wp.Data != "" {
		// Decode base64 embedded data
		data, err := base64.StdEncoding.DecodeString(wp.Data)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode wallpaper data: %w", err)
		}
		return data, wp.Format, nil
	}

	if wp.Path != "" {
		// Read from file path
		path := wp.Path
		if !filepath.IsAbs(path) {
			// Relative to the theme file
			path = filepath.Join(filepath.Dir(basePath), path)
		}
		data, err := os.ReadFile(path) // #nosec G304 -- User-provided wallpaper path is intentional
		if err != nil {
			return nil, "", fmt.Errorf("failed to read wallpaper file: %w", err)
		}
		// Detect format from extension
		format := strings.TrimPrefix(filepath.Ext(path), ".")
		return data, format, nil
	}

	if wp.URL != "" {
		// Fetch from URL
		resp, err := http.Get(wp.URL)
		if err != nil {
			return nil, "", fmt.Errorf("failed to fetch wallpaper URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("wallpaper URL returned status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read wallpaper data: %w", err)
		}

		// Try to detect format from content-type or URL
		format := detectFormatFromContentType(resp.Header.Get("Content-Type"))
		if format == "" {
			format = strings.TrimPrefix(filepath.Ext(wp.URL), ".")
		}
		return data, format, nil
	}

	return nil, "", fmt.Errorf("wallpaper has no data, path, or URL")
}

// detectFormatFromContentType extracts image format from Content-Type header.
func detectFormatFromContentType(contentType string) string {
	switch {
	case strings.Contains(contentType, "image/png"):
		return "png"
	case strings.Contains(contentType, "image/jpeg"):
		return "jpg"
	case strings.Contains(contentType, "image/webp"):
		return "webp"
	case strings.Contains(contentType, "image/gif"):
		return "gif"
	default:
		return ""
	}
}
