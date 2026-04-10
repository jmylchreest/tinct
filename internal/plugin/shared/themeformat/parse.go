package themeformat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/jmylchreest/tinct/internal/security"
	httputil "github.com/jmylchreest/tinct/internal/util/http"
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
func DecodeWallpaper(theme *Theme, basePath string) (data []byte, format string, err error) {
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
		// Validate URL against SSRF (theme files are shareable content).
		if err := security.ValidateRuntimeHTTPURL(wp.URL); err != nil {
			return nil, "", fmt.Errorf("unsafe wallpaper URL: %w", err)
		}

		// Fetch with timeout and size limit.
		data, err := httputil.Fetch(context.Background(), wp.URL, httputil.FetchOptions{})
		if err != nil {
			return nil, "", fmt.Errorf("failed to fetch wallpaper URL: %w", err)
		}

		// Detect format from URL extension, falling back to content sniffing.
		format := strings.TrimPrefix(filepath.Ext(wp.URL), ".")
		if format == "" {
			format = detectFormatFromBytes(data)
		}
		return data, format, nil
	}

	return nil, "", fmt.Errorf("wallpaper has no data, path, or URL")
}

// detectFormatFromBytes sniffs the image format from the raw data using
// Go's http.DetectContentType (magic-byte detection).
func detectFormatFromBytes(data []byte) string {
	ct := http.DetectContentType(data)
	switch {
	case strings.Contains(ct, "image/png"):
		return "png"
	case strings.Contains(ct, "image/jpeg"):
		return "jpg"
	case strings.Contains(ct, "image/webp"):
		return "webp"
	case strings.Contains(ct, "image/gif"):
		return "gif"
	default:
		return ""
	}
}
