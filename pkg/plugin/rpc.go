// Package plugin provides the public API for tinct plugins.
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"net/rpc"
	"strings"

	"github.com/hashicorp/go-plugin"

	"github.com/jmylchreest/tinct/pkg/plugin/hooks"
)

// InputPluginRPC implements the go-plugin Plugin interface for input plugins.
type InputPluginRPC struct {
	plugin.Plugin
	Impl InputPlugin
}

// Server returns an RPC server for this plugin.
func (p *InputPluginRPC) Server(*plugin.MuxBroker) (any, error) {
	return &InputPluginRPCServer{Impl: p.Impl}, nil
}

// Client returns an RPC client for this plugin.
func (p *InputPluginRPC) Client(_ *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &InputPluginRPCClient{client: c}, nil
}

// InputPluginRPCServer is the RPC server implementation for input plugins.
type InputPluginRPCServer struct {
	Impl InputPlugin
}

// Generate implements the RPC method for palette generation.
//
// Wire format (protocol >= 0.3.0): if the plugin implements RoleHinter or
// ThemeHinter, the response is a JSON object with shape
// {"colors": [...], "role_hints": {...}, "theme_hint": "..."}. Otherwise the
// legacy bare colour array is emitted, which any older client can parse.
func (s *InputPluginRPCServer) Generate(opts InputOptions, resp *[]byte) error {
	colours, err := s.Impl.Generate(context.Background(), opts)
	if err != nil {
		return err
	}

	rgbColours := make([]map[string]uint8, len(colours))
	for i, c := range colours {
		r, g, b, _ := c.RGBA()
		rgbColours[i] = map[string]uint8{
			"r": uint8(r >> 8), //nolint:gosec // G115: RGBA returns uint32, right-shift by 8 to get 8-bit color value
			"g": uint8(g >> 8), //nolint:gosec // G115: RGBA returns uint32, right-shift by 8 to get 8-bit color value
			"b": uint8(b >> 8), //nolint:gosec // G115: RGBA returns uint32, right-shift by 8 to get 8-bit color value
		}
	}

	var roleHints map[string]int
	if hinter, ok := s.Impl.(RoleHinter); ok {
		if hints := hinter.RoleHints(); len(hints) > 0 {
			roleHints = hints
		}
	}

	var themeHint string
	if hinter, ok := s.Impl.(ThemeHinter); ok {
		themeHint = hinter.ThemeHint()
	}

	// Stay on the legacy array wire format when there are no hints, so 0.2.0
	// hosts can still read responses from a 0.3.0-compiled plugin.
	if roleHints == nil && themeHint == "" {
		data, err := json.Marshal(rgbColours)
		if err != nil {
			return fmt.Errorf("failed to marshal colours: %w", err)
		}
		*resp = data
		return nil
	}

	envelope := map[string]any{"colors": rgbColours}
	if roleHints != nil {
		envelope["role_hints"] = roleHints
	}
	if themeHint != "" {
		envelope["theme_hint"] = themeHint
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal colours: %w", err)
	}
	*resp = data
	return nil
}

// GetMetadata implements the RPC method for fetching plugin metadata.
func (s *InputPluginRPCServer) GetMetadata(_ any, resp *PluginInfo) error {
	*resp = s.Impl.GetMetadata()
	return nil
}

// Validate implements the optional Validator RPC method. Plugins that
// don't implement the Validator interface report success — the host
// falls back to no-op validation when an old plugin is missing this
// method anyway, so this branch keeps the wire shape consistent for
// 0.3.0+ plugins built against the new SDK.
func (s *InputPluginRPCServer) Validate(args map[string]any, resp *string) error {
	if v, ok := s.Impl.(Validator); ok {
		if err := v.Validate(args); err != nil {
			*resp = err.Error()
			return nil
		}
	}
	*resp = ""
	return nil
}

// WallpaperPath implements the RPC method for fetching wallpaper path.
func (s *InputPluginRPCServer) WallpaperPath(_ any, resp *string) error {
	*resp = s.Impl.WallpaperPath()
	return nil
}

// WallpaperRawPath implements the RPC method for fetching raw wallpaper path.
func (s *InputPluginRPCServer) WallpaperRawPath(_ any, resp *string) error {
	*resp = s.Impl.WallpaperRawPath()
	return nil
}

// GetFlagHelp implements the RPC method for fetching flag help.
func (s *InputPluginRPCServer) GetFlagHelp(_ any, resp *[]FlagHelp) error {
	*resp = s.Impl.GetFlagHelp()
	return nil
}

// InputPluginRPCClient is the RPC client implementation for input plugins.
//
// LastRoleHints and LastThemeHint expose any hints from the most recent
// Generate call (protocol >= 0.3.0). They return zero-values when the plugin
// is older or did not provide hints.
type InputPluginRPCClient struct {
	client         *rpc.Client
	lastRoleHints  map[string]int
	lastThemeHint  string
}

// rgbBlock is the per-colour wire shape used by both response formats.
type rgbBlock struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// Generate calls the remote Generate method and parses both the legacy bare-
// array response and the 0.3.0 envelope response. Any hints in the envelope
// are cached and exposed via LastRoleHints / LastThemeHint.
func (c *InputPluginRPCClient) Generate(_ context.Context, opts InputOptions) ([]color.Color, error) {
	var respBytes []byte
	err := c.client.Call("Plugin.Generate", opts, &respBytes)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}

	c.lastRoleHints = nil
	c.lastThemeHint = ""

	rgbs, hints, theme, err := parseInputResponse(respBytes)
	if err != nil {
		return nil, err
	}
	c.lastRoleHints = hints
	c.lastThemeHint = theme

	colours := make([]color.Color, len(rgbs))
	for i, rgb := range rgbs {
		colours[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
	}
	return colours, nil
}

// LastRoleHints returns the role hints captured from the most recent Generate
// call, or nil if the plugin did not supply any.
func (c *InputPluginRPCClient) LastRoleHints() map[string]int {
	return c.lastRoleHints
}

// LastThemeHint returns the theme hint captured from the most recent Generate
// call, or "" if the plugin did not supply one.
func (c *InputPluginRPCClient) LastThemeHint() string {
	return c.lastThemeHint
}

// parseInputResponse accepts either the legacy bare-array response or the
// 0.3.0 envelope `{"colors": [...], "role_hints": {...}, "theme_hint": "..."}`.
// Detection is by first non-whitespace byte to keep the two formats
// unambiguous.
func parseInputResponse(data []byte) ([]rgbBlock, map[string]int, string, error) {
	for _, b := range data {
		switch b {
		case ' ', '\t', '\n', '\r':
			continue
		case '{':
			var env struct {
				Colors    []rgbBlock     `json:"colors"`
				RoleHints map[string]int `json:"role_hints,omitempty"`
				ThemeHint string         `json:"theme_hint,omitempty"`
			}
			if err := json.Unmarshal(data, &env); err != nil {
				return nil, nil, "", fmt.Errorf("failed to unmarshal colours envelope: %w", err)
			}
			return env.Colors, env.RoleHints, env.ThemeHint, nil
		case '[':
			var arr []rgbBlock
			if err := json.Unmarshal(data, &arr); err != nil {
				return nil, nil, "", fmt.Errorf("failed to unmarshal colours: %w", err)
			}
			return arr, nil, "", nil
		default:
			return nil, nil, "", fmt.Errorf("unexpected response shape (first byte %q)", b)
		}
	}
	return nil, nil, "", fmt.Errorf("empty response from plugin")
}

// GetMetadata calls the remote GetMetadata method.
func (c *InputPluginRPCClient) GetMetadata() (PluginInfo, error) {
	var info PluginInfo
	err := c.client.Call("Plugin.GetMetadata", new(any), &info)
	if err != nil {
		return info, fmt.Errorf("RPC call failed: %w", err)
	}
	return info, nil
}

// Validate calls the remote optional Validate method. Plugins built
// against pre-0.3.0 SDKs don't expose the method — net/rpc surfaces
// that as "can't find method", which is mapped to a successful
// validation so older plugins remain compatible.
func (c *InputPluginRPCClient) Validate(args map[string]any) error {
	if args == nil {
		args = map[string]any{}
	}
	var msg string
	err := c.client.Call("Plugin.Validate", args, &msg)
	if err != nil {
		if isMissingMethodErr(err) {
			return nil
		}
		return fmt.Errorf("RPC call failed: %w", err)
	}
	if msg != "" {
		return &RPCError{Message: msg}
	}
	return nil
}

// WallpaperPath calls the remote WallpaperPath method.
func (c *InputPluginRPCClient) WallpaperPath() string {
	var path string
	err := c.client.Call("Plugin.WallpaperPath", new(any), &path)
	if err != nil {
		return ""
	}
	return path
}

// WallpaperRawPath calls the remote WallpaperRawPath method.
func (c *InputPluginRPCClient) WallpaperRawPath() string {
	var path string
	err := c.client.Call("Plugin.WallpaperRawPath", new(any), &path)
	if err != nil {
		return ""
	}
	return path
}

// GetFlagHelp calls the remote GetFlagHelp method.
func (c *InputPluginRPCClient) GetFlagHelp() []FlagHelp {
	var help []FlagHelp
	err := c.client.Call("Plugin.GetFlagHelp", new(any), &help)
	if err != nil {
		return []FlagHelp{}
	}
	return help
}

// OutputPluginRPC implements the go-plugin Plugin interface for output plugins.
type OutputPluginRPC struct {
	plugin.Plugin
	Impl OutputPlugin
}

// Server returns an RPC server for this plugin.
func (p *OutputPluginRPC) Server(*plugin.MuxBroker) (any, error) {
	return &OutputPluginRPCServer{Impl: p.Impl}, nil
}

// Client returns an RPC client for this plugin.
func (p *OutputPluginRPC) Client(_ *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &OutputPluginRPCClient{client: c}, nil
}

// OutputPluginRPCServer is the RPC server implementation for output plugins.
type OutputPluginRPCServer struct {
	Impl OutputPlugin
}

// Generate implements the RPC method for output generation.
func (s *OutputPluginRPCServer) Generate(palette PaletteData, resp *map[string][]byte) error { //nolint:gocritic // net/rpc requires pointer response params
	result, err := s.Impl.Generate(context.Background(), palette)
	if err != nil {
		return err
	}
	*resp = result
	return nil
}

// PreExecute implements the RPC method for pre-execution hooks.
func (s *OutputPluginRPCServer) PreExecute(_ any, resp *struct {
	Skip   bool
	Reason string
	Error  string
}) error {

	skip, reason, err := s.Impl.PreExecute(context.Background())
	resp.Skip = skip
	resp.Reason = reason
	if err != nil {
		resp.Error = err.Error()
	}
	return nil
}

// PostExecute implements the RPC method for post-execution hooks.
func (s *OutputPluginRPCServer) PostExecute(files []string, resp *string) error {
	err := s.Impl.PostExecute(context.Background(), files)
	if err != nil {
		*resp = err.Error()
		return err
	}
	return nil
}

// GetMetadata implements the RPC method for fetching plugin metadata.
func (s *OutputPluginRPCServer) GetMetadata(_ any, resp *PluginInfo) error {
	*resp = s.Impl.GetMetadata()
	return nil
}

// GetFlagHelp implements the RPC method for fetching flag help.
func (s *OutputPluginRPCServer) GetFlagHelp(_ any, resp *[]FlagHelp) error {
	*resp = s.Impl.GetFlagHelp()
	return nil
}

// Validate implements the optional Validator RPC method for output plugins.
func (s *OutputPluginRPCServer) Validate(args map[string]any, resp *string) error {
	if v, ok := s.Impl.(Validator); ok {
		if err := v.Validate(args); err != nil {
			*resp = err.Error()
			return nil
		}
	}
	*resp = ""
	return nil
}

// GetHooks implements the optional HooksProvider RPC method. Returns
// the gob-safe subset of the plugin's hooks.Spec; function fields
// (ReloadFn, InstructionsFn, Wallpaper) are not transmitted because
// gob cannot marshal them — plugins needing dynamic behaviour keep the
// imperative PostExecute path.
//
// First arg is `struct{}` (not the legacy `any`) so the wire encoding
// from the client is a concrete zero-value rather than a nil interface;
// nil-interface encoding is the one shape that wedges the server-side
// missing-method discard path on plugins built before this method
// existed.
func (s *OutputPluginRPCServer) GetHooks(_ struct{}, resp *HookSpecPayload) error {
	if h, ok := s.Impl.(HooksProvider); ok {
		*resp = HookSpecFromSpec(h.Hooks())
	}
	return nil
}

// GetTemplates implements the optional TemplateLister RPC method.
// Returns the plugin's bundled templates so `tinct plugins templates
// list` / `templates dump` can introspect them across the RPC boundary.
//
// Plugins that don't implement TemplateLister return an empty map; the
// host treats that as "no templates" and skips the plugin in template
// commands. As with GetHooks, the first arg is concrete `struct{}` to
// keep the missing-method path safe on older clients.
func (s *OutputPluginRPCServer) GetTemplates(_ struct{}, resp *map[string][]byte) error {
	out := map[string][]byte{}
	if t, ok := s.Impl.(TemplateLister); ok {
		if listed := t.Templates(); len(listed) > 0 {
			for k, v := range listed {
				out[k] = v
			}
		}
	}
	*resp = out
	return nil
}

// OutputPluginRPCClient is the RPC client implementation for output plugins.
type OutputPluginRPCClient struct {
	client *rpc.Client
}

// Generate calls the remote Generate method.
func (c *OutputPluginRPCClient) Generate(_ context.Context, palette PaletteData) (map[string][]byte, error) {
	var result map[string][]byte
	err := c.client.Call("Plugin.Generate", palette, &result)
	if err != nil {
		return result, fmt.Errorf("RPC call failed: %w", err)
	}
	return result, nil
}

// PreExecute calls the remote PreExecute method.
func (c *OutputPluginRPCClient) PreExecute(_ context.Context) (skip bool, reason string, err error) {
	var resp struct {
		Skip   bool
		Reason string
		Error  string
	}
	err = c.client.Call("Plugin.PreExecute", new(any), &resp)
	if err != nil {
		return false, "", fmt.Errorf("RPC call failed: %w", err)
	}
	if resp.Error != "" {
		return resp.Skip, resp.Reason, &RPCError{Message: resp.Error}
	}
	return resp.Skip, resp.Reason, nil
}

// PostExecute calls the remote PostExecute method.
func (c *OutputPluginRPCClient) PostExecute(_ context.Context, files []string) error {
	var errMsg string
	err := c.client.Call("Plugin.PostExecute", files, &errMsg)
	if err != nil {
		return fmt.Errorf("RPC call failed: %w", err)
	}
	if errMsg != "" {
		return &RPCError{Message: errMsg}
	}
	return nil
}

// GetMetadata calls the remote GetMetadata method.
func (c *OutputPluginRPCClient) GetMetadata() (PluginInfo, error) {
	var info PluginInfo
	err := c.client.Call("Plugin.GetMetadata", new(any), &info)
	if err != nil {
		return info, fmt.Errorf("RPC call failed: %w", err)
	}
	return info, nil
}

// Validate calls the remote optional Validate method. Plugins built
// against pre-0.3.0 SDKs don't expose the method; that is treated as a
// successful validation so older plugins remain compatible.
func (c *OutputPluginRPCClient) Validate(args map[string]any) error {
	if args == nil {
		args = map[string]any{}
	}
	var msg string
	err := c.client.Call("Plugin.Validate", args, &msg)
	if err != nil {
		if isMissingMethodErr(err) {
			return nil
		}
		return fmt.Errorf("RPC call failed: %w", err)
	}
	if msg != "" {
		return &RPCError{Message: msg}
	}
	return nil
}

// GetHooks calls the remote optional GetHooks method, returning the
// plugin's static hooks.Spec. Plugins built against pre-0.3.0 SDKs
// don't expose the method — net/rpc surfaces that as "can't find
// method" and we return ok=false so the host falls back to the
// imperative PreExecute / PostExecute methods alone. ok is also false
// on transport errors.
func (c *OutputPluginRPCClient) GetHooks() (hooks.Spec, bool) {
	var payload HookSpecPayload
	err := c.client.Call("Plugin.GetHooks", noArgs, &payload)
	if err != nil {
		return hooks.Spec{}, false
	}
	return HookSpecToSpec(payload), true
}

// GetTemplates calls the remote optional GetTemplates method,
// returning the plugin's bundled templates. Plugins built against
// pre-0.3.0 SDKs (or 0.3.0 plugins that don't implement
// TemplateLister) cause a "can't find method" or empty map; in either
// case ok=false and the caller treats the plugin as having no
// templates exposed.
func (c *OutputPluginRPCClient) GetTemplates() (map[string][]byte, bool) {
	var payload map[string][]byte
	err := c.client.Call("Plugin.GetTemplates", noArgs, &payload)
	if err != nil || len(payload) == 0 {
		return nil, false
	}
	return payload, true
}

// GetFlagHelp calls the remote GetFlagHelp method.
func (c *OutputPluginRPCClient) GetFlagHelp() []FlagHelp {
	var help []FlagHelp
	err := c.client.Call("Plugin.GetFlagHelp", new(any), &help)
	if err != nil {
		return []FlagHelp{}
	}
	return help
}

// noArgs is the canonical zero-value used for RPC methods that take no
// arguments. We do NOT use `new(any)` (a pointer to a nil interface)
// here because gob cannot reliably encode it — the resulting wire
// payload is truncated, and on the server side a missing-method
// response is built by reading + discarding the body via
// gob.Decode(nil), which blocks forever waiting for the rest of the
// malformed value. The end result is that calling any RPC method that
// doesn't exist on the server hangs the client. A concrete zero-value
// (here, an empty struct) round-trips through gob cleanly, so the
// server's missing-method error response reaches the client as
// intended.
//
// Always pass noArgs (or a typed argument) to client.Call — never
// `new(any)` or `nil`.
var noArgs = struct{}{}

// RPCError represents an error returned from an RPC call.
type RPCError struct {
	Message string
}

// Error implements the error interface.
func (e *RPCError) Error() string {
	return e.Message
}

// isMissingMethodErr reports whether an error from net/rpc means "the
// remote plugin does not implement this method". Used by clients of
// optional 0.3.0+ RPC methods (Validate, GetHooks, …) to fall back
// gracefully when talking to a plugin built against an older SDK.
//
// net/rpc returns these as plain errors with messages like
// `rpc: can't find method "Plugin.Validate"` or
// `rpc: service/method request ill-formed`. Match by substring since
// the package exposes no sentinel.
func isMissingMethodErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "can't find method") ||
		strings.Contains(msg, "can't find service")
}
