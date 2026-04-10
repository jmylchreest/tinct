// Package remotejson provides an input plugin for fetching colour palettes from remote JSON sources with JSONPath queries.
package remotejson

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/input/shared/palettebuilder"
	"github.com/jmylchreest/tinct/internal/security"
	httputil "github.com/jmylchreest/tinct/internal/util/http"
	"github.com/jmylchreest/tinct/internal/version"
)

// Plugin implements the input.Plugin interface for remote JSON palette fetching with JSONPath queries.
type Plugin struct {
	url     string
	query   string // JSONPath query to extract colors (optional)
	timeout time.Duration
	mapping map[string]string // Maps source color names to tinct roles
}

// New creates a new remote-json input plugin.
func New() *Plugin {
	return &Plugin{
		timeout: 10 * time.Second,
		mapping: make(map[string]string),
	}
}

// Name returns the plugin name.
func (p *Plugin) Name() string {
	return "remote-json"
}

// Description returns the plugin description.
func (p *Plugin) Description() string {
	return "Fetch colours from remote JSON with JSONPath queries"
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return version.Version
}

// RegisterFlags registers plugin-specific flags with the cobra command.
func (p *Plugin) RegisterFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&p.url, "remote-json.url", "", "URL to fetch JSON palette from (required)")
	cmd.Flags().StringVar(&p.query, "remote-json.query", "", "JSONPath query to extract colors (e.g. '$.colors', optional)")
	cmd.Flags().DurationVar(&p.timeout, "remote-json.timeout", 10*time.Second, "HTTP timeout")
	cmd.Flags().StringToStringVar(&p.mapping, "remote-json.map", map[string]string{}, "Map colors to roles (e.g. base=background,text=foreground)")
}

// Validate checks if the plugin has all required inputs configured.
func (p *Plugin) Validate() error {
	if p.url == "" {
		return fmt.Errorf("--remote-json.url is required")
	}

	// Validate URL: HTTPS-only by default, blocks localhost/private IPs.
	// Set TINCT_ALLOW_INSECURE_HTTP=1 to permit plain HTTP.
	if err := security.ValidateRuntimeHTTPURL(p.url); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	return nil
}

// GetFlagHelp returns help information for all plugin flags.
func (p *Plugin) GetFlagHelp() []input.FlagHelp {
	return []input.FlagHelp{
		{Name: "remote-json.url", Type: "string", Default: "", Description: "URL to fetch JSON palette from (required)", Required: true},
		{Name: "remote-json.query", Type: "string", Default: "", Description: "JSONPath query to extract colors (e.g. '$.colors')", Required: false},
		{Name: "remote-json.timeout", Type: "duration", Default: "10s", Description: "HTTP timeout", Required: false},
		{Name: "remote-json.map", Type: "stringToString", Default: "map[]", Description: "Map colors to roles (e.g. base=background,text=foreground)", Required: false},
	}
}

// Generate fetches and parses a remote JSON colour palette.
func (p *Plugin) Generate(ctx context.Context, opts input.GenerateOptions) (*colour.Palette, error) {
	if opts.Verbose {
		fmt.Printf("→ Fetching JSON palette from: %s\n", p.url)
	}

	// Fetch content.
	content, err := p.fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch palette: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("   Size: %d bytes\n", len(content))
	}

	// Parse JSON.
	colors, err := p.parseJSON(content, opts.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if opts.Verbose {
		fmt.Printf("   Extracted %d colors\n", len(colors))
	}

	// Convert to palette.
	palette, err := p.buildPalette(colors, opts.Verbose)
	if err != nil {
		return nil, err
	}

	return palette, nil
}

// fetch retrieves content from the remote URL.
func (p *Plugin) fetch(ctx context.Context) ([]byte, error) {
	return httputil.Fetch(ctx, p.url, httputil.FetchOptions{
		Timeout: p.timeout,
	})
}

// parseJSON parses JSON content and extracts colors.
func (p *Plugin) parseJSON(content []byte, verbose bool) (map[string]string, error) {
	var data any
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	// If query is provided, navigate to that path.
	if p.query != "" {
		if verbose {
			fmt.Printf("   Applying query: %s\n", p.query)
		}
		var err error
		data, err = p.applyQuery(data, p.query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
	}

	// Extract colors from the data.
	colors := make(map[string]string)
	p.extractColors(data, "", colors)

	return colors, nil
}

// applyQuery applies a simplified JSONPath-like query to the data.
// Supports: $.path.to.field or just path.to.field.
func (p *Plugin) applyQuery(data any, query string) (any, error) {
	// Remove leading $. if present.
	query = strings.TrimPrefix(query, "$.")
	query = strings.TrimPrefix(query, "$")

	if query == "" {
		return data, nil
	}

	// Split path into segments.
	segments := strings.Split(query, ".")

	current := data
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]any:
			val, ok := v[segment]
			if !ok {
				return nil, fmt.Errorf("path not found: %s", segment)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot navigate into %T at segment '%s'", current, segment)
		}
	}

	return current, nil
}

// extractColors recursively extracts color values from JSON data.
func (p *Plugin) extractColors(data any, prefix string, colors map[string]string) { //nolint:gocyclo,gocognit // recursive JSON traversal with type switching
	switch v := data.(type) {
	case map[string]any:
		// Check if this object looks like a color object (has hex/rgb properties).
		if hexVal, hasHex := v["hex"].(string); hasHex && isColor(hexVal) {
			// This is a color object (e.g., Catppuccin format: {"hex": "#1e1e2e", "rgb": {...}}).
			// Use the prefix as the key.
			if prefix != "" {
				key := prefix
				if idx := strings.LastIndex(prefix, "."); idx >= 0 {
					key = prefix[idx+1:]
				}
				colors[key] = hexVal
			}
			return
		}

		// Not a color object, traverse its properties.
		for key, value := range v {
			fullKey := key
			if prefix != "" {
				fullKey = prefix + "." + key
			}

			// Check if this is a color value (string that looks like a hex color).
			if strVal, ok := value.(string); ok {
				if isColor(strVal) {
					colors[key] = strVal
				} else {
					// Might be nested, recurse.
					p.extractColors(value, fullKey, colors)
				}
			} else {
				// Recurse into nested structures.
				p.extractColors(value, fullKey, colors)
			}
		}
	case []any:
		for i, item := range v {
			p.extractColors(item, fmt.Sprintf("%s[%d]", prefix, i), colors)
		}
	case string:
		if isColor(v) && prefix != "" {
			// Use the last segment of the path as the key.
			key := prefix
			if idx := strings.LastIndex(prefix, "."); idx >= 0 {
				key = prefix[idx+1:]
			}
			colors[key] = v
		}
	}
}

// isColor checks if a string looks like a color value.
func isColor(s string) bool {
	s = strings.TrimSpace(s)

	// Hex color: #RGB or #RRGGBB.
	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		return len(hex) == 3 || len(hex) == 6
	}

	// RGB/RGBA format: rgb(r, g, b) or rgba(r, g, b, a).
	if strings.HasPrefix(s, "rgb(") || strings.HasPrefix(s, "rgba(") {
		return true
	}

	// HSL/HSLA format: hsl(h, s, l) or hsla(h, s, l, a).
	if strings.HasPrefix(s, "hsl(") || strings.HasPrefix(s, "hsla(") {
		return true
	}

	return false
}

// buildPalette converts extracted colors to a Palette.
func (p *Plugin) buildPalette(colors map[string]string, verbose bool) (*colour.Palette, error) {
	return palettebuilder.BuildPalette(colors, p.mapping, verbose)
}
