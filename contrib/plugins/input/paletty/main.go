// paletty - paletty.dev palette fetcher (Tinct external input plugin).
//
// Fetches a curated terminal palette from https://paletty.dev's public JSON
// API, applies a default role-mapping (background, foreground, danger,
// success, warning, info, notification) and surfaces the source's is_dark
// metadata as a theme hint via the 0.3.0 plugin protocol.
//
// Build:
//
//	go build -o tinct-plugin-paletty
//
// Plugin args (PluginArgs map keys):
//
//	palette  – paletty palette ID or full paletty.dev URL (required)
//	timeout  – HTTP timeout, e.g. "10s" (optional; default 10s)
//	base_url – override paletty.dev base URL (optional; testing only)
//	map      – extra role-name → paletty-key mapping merged on top of defaults,
//	           supplied as a JSON object (optional)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"

	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

var (
	// Version is the semantic version of the plugin.
	// Injected at build time via: -ldflags "-X main.Version=x.y.z"
	Version = "0.0.0"

	// Commit is the git commit hash of the build.
	Commit = "unknown"

	// Date is the build date in RFC3339 format.
	Date = "unknown"
)

const (
	defaultBaseURL = "https://paletty.dev"
	defaultTimeout = 10 * time.Second
	pluginName     = "paletty"
)

// idPattern matches a paletty palette ID — the opaque short token used in URLs.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9]{6,32}$`)

// urlIDPattern extracts the ID from a paletty.dev URL like /p/<id>/<slug> or /api/palettes/<id>.
var urlIDPattern = regexp.MustCompile(`paletty\.dev/(?:p|api/palettes)/([A-Za-z0-9]+)`)

// defaultRoleMap maps tinct role names → paletty colour keys.
// Inverted (role-as-key) so user --plugin-arg map entries naturally read
// "accent1=ansi.normal.6".
var defaultRoleMap = map[string]string{
	"background":   "core.background",
	"foreground":   "core.foreground",
	"danger":       "ansi.normal.1",
	"success":      "ansi.normal.2",
	"warning":      "ansi.normal.3",
	"info":         "ansi.normal.4",
	"notification": "ansi.normal.5",
}

// PalettyPlugin implements tinctplugin.InputPlugin plus the optional
// Validator, RoleHinter, and ThemeHinter interfaces (protocol 0.3.0+).
type PalettyPlugin struct {
	roleHints map[string]int
	themeHint string
}

// Validate fail-fast checks the persistent palette argument supplied at
// `tinct plugins config` time, so misconfigured plugins error out
// before the host attempts a fetch. Per-run --plugin-args overrides
// land at Generate time and are validated there. Implements
// tinctplugin.Validator.
func (p *PalettyPlugin) Validate(args map[string]any) error {
	raw, _ := args["palette"].(string)
	if raw == "" {
		return nil
	}
	if _, err := resolveID(raw); err != nil {
		return err
	}
	return nil
}

// apiResponse is the minimal subset of the paletty.dev /api/palettes/{id} response we care about.
type apiResponse struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Slug   string            `json:"slug"`
	Colors map[string]string `json:"colors"`
	IsDark bool              `json:"is_dark"`
}

// Generate fetches a paletty palette and returns its colours plus role hints.
func (p *PalettyPlugin) Generate(ctx context.Context, opts tinctplugin.InputOptions) ([]color.Color, error) {
	p.roleHints = nil
	p.themeHint = ""

	paletteRef, _ := opts.PluginArgs["palette"].(string)
	if paletteRef == "" {
		return nil, fmt.Errorf("plugin-arg 'palette' is required (paletty.dev URL or ID)")
	}
	id, err := resolveID(paletteRef)
	if err != nil {
		return nil, err
	}

	timeout := defaultTimeout
	if raw, ok := opts.PluginArgs["timeout"].(string); ok && raw != "" {
		if d, perr := time.ParseDuration(raw); perr == nil {
			timeout = d
		}
	}

	base := defaultBaseURL
	if raw, ok := opts.PluginArgs["base_url"].(string); ok && raw != "" {
		base = raw
	}

	url := fmt.Sprintf("%s/api/palettes/%s", strings.TrimRight(base, "/"), id)

	if opts.DryRun {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "DRY-RUN: would fetch %s\n", url)
		}
		return []color.Color{}, nil
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "→ Fetching paletty palette: %s\n", url)
	}

	body, err := fetch(ctx, url, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch palette: %w", err)
	}

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse paletty response: %w", err)
	}
	if len(resp.Colors) == 0 {
		return nil, fmt.Errorf("paletty palette %q has no colours", id)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "   %s (slug=%s) — %d colours, dark=%v\n", resp.Name, resp.Slug, len(resp.Colors), resp.IsDark)
	}

	// Build a deterministic colour ordering: sorted by paletty key for stable
	// indices and predictable role-hint mappings.
	keys := make([]string, 0, len(resp.Colors))
	for k := range resp.Colors {
		keys = append(keys, k)
	}
	// Sort for reproducibility (json map iteration is random).
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}

	colours := make([]color.Color, 0, len(keys))
	indexByKey := make(map[string]int, len(keys))
	for _, k := range keys {
		rgba, perr := parseHex(resp.Colors[k])
		if perr != nil {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "   Skipping invalid colour %s=%q: %v\n", k, resp.Colors[k], perr)
			}
			continue
		}
		indexByKey[k] = len(colours)
		colours = append(colours, rgba)
	}

	// Build role mapping: defaults overlaid with --plugin-arg map=… entries.
	mapping := make(map[string]string, len(defaultRoleMap))
	for role, key := range defaultRoleMap {
		mapping[role] = key
	}
	if extra, ok := opts.PluginArgs["map"].(map[string]any); ok {
		for role, v := range extra {
			if s, isStr := v.(string); isStr && s != "" {
				mapping[role] = s
			}
		}
	}

	hints := make(map[string]int, len(mapping))
	for role, key := range mapping {
		idx, ok := indexByKey[key]
		if !ok {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "   Skipping role %s: paletty key %q not found in palette\n", role, key)
			}
			continue
		}
		hints[role] = idx
	}
	p.roleHints = hints

	if resp.IsDark {
		p.themeHint = "dark"
	} else {
		p.themeHint = "light"
	}

	return colours, nil
}

// RoleHints returns the role → colour-index map populated by Generate.
// Implements tinctplugin.RoleHinter (protocol 0.3.0+).
func (p *PalettyPlugin) RoleHints() map[string]int {
	return p.roleHints
}

// ThemeHint returns "dark" or "light" based on paletty's is_dark field.
// Implements tinctplugin.ThemeHinter (protocol 0.3.0+).
func (p *PalettyPlugin) ThemeHint() string {
	return p.themeHint
}

// GetMetadata returns plugin metadata.
func (p *PalettyPlugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            pluginName,
		Type:            "input",
		Version:         Version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     "Fetch terminal palettes from paletty.dev",
		PluginProtocol:  "go-plugin",
	}
}

// WallpaperPath returns "" — paletty does not provide wallpapers.
func (p *PalettyPlugin) WallpaperPath() string { return "" }

// WallpaperRawPath returns "" — paletty does not provide wallpapers.
func (p *PalettyPlugin) WallpaperRawPath() string { return "" }

// GetFlagHelp returns the plugin's argument help.
func (p *PalettyPlugin) GetFlagHelp() []tinctplugin.FlagHelp {
	return []tinctplugin.FlagHelp{
		{Name: "palette", Type: "string", Default: "", Description: "paletty.dev palette ID or full URL (e.g. MDRHC0lqRj or https://paletty.dev/p/MDRHC0lqRj/midnight-a)", Required: true},
		{Name: "timeout", Type: "duration", Default: defaultTimeout.String(), Description: "HTTP timeout"},
		{Name: "base_url", Type: "string", Default: defaultBaseURL, Description: "Override paletty.dev base URL (advanced)"},
		{Name: "map", Type: "map", Default: "", Description: "Extra role→paletty-key mappings merged with defaults (e.g. {\"accent1\":\"ansi.normal.6\"})"},
	}
}

func resolveID(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("empty palette reference")
	}
	if m := urlIDPattern.FindStringSubmatch(s); m != nil {
		return m[1], nil
	}
	if idPattern.MatchString(s) {
		return s, nil
	}
	return "", fmt.Errorf("invalid paletty palette reference %q (expected an ID like 'MDRHC0lqRj' or a paletty.dev URL)", s)
}

func fetch(ctx context.Context, url string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("tinct-plugin-paletty/%s", Version))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

func parseHex(s string) (color.Color, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return nil, fmt.Errorf("expected 6 hex digits, got %d", len(s))
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return nil, err
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		p := &PalettyPlugin{}
		info := p.GetMetadata()
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding plugin info: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: tinctplugin.Handshake,
		Plugins: map[string]plugin.Plugin{
			"input": &tinctplugin.InputPluginRPC{Impl: &PalettyPlugin{}},
		},
	})
}
