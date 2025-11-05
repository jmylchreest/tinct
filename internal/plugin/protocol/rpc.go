// Package protocol defines the plugin protocol version and compatibility checking.
package protocol

import (
	"context"
	"encoding/json"
	"image/color"
	"net/rpc"

	"github.com/hashicorp/go-plugin"

	"github.com/jmylchreest/tinct/internal/colour"
)

// InputPlugin is the interface that input plugins must implement for go-plugin RPC.
type InputPlugin interface {
	// Generate creates a palette from plugin-specific inputs.
	Generate(ctx context.Context, opts InputOptions) ([]color.Color, error)

	// GetMetadata returns plugin metadata.
	GetMetadata() PluginInfo
}

// OutputPlugin is the interface that output plugins must implement for go-plugin RPC.
type OutputPlugin interface {
	// Generate creates output file(s) from the given theme data.
	Generate(ctx context.Context, palette PaletteData) (map[string][]byte, error)

	// PreExecute runs before Generate() for validation checks.
	PreExecute(ctx context.Context) (skip bool, reason string, err error)

	// PostExecute runs after successful Generate() and file writing.
	PostExecute(ctx context.Context, writtenFiles []string) error

	// GetMetadata returns plugin metadata.
	GetMetadata() PluginInfo
}

// InputOptions holds options for input plugin generation (matches input.GenerateOptions).
type InputOptions struct {
	Verbose         bool           `json:"verbose"`
	DryRun          bool           `json:"dry_run"`
	ColourOverrides []string       `json:"colour_overrides,omitempty"`
	PluginArgs      map[string]any `json:"plugin_args,omitempty"`
}

// PaletteData is the palette data sent to output plugins.
type PaletteData struct {
	Colours    map[string]CategorisedColour `json:"colours"`
	AllColours []CategorisedColour          `json:"all_colours"`
	ThemeType  string                       `json:"theme_type"`
	PluginArgs map[string]any               `json:"plugin_args,omitempty"`
	DryRun     bool                         `json:"dry_run"`
}

// CategorisedColour represents a color with metadata for RPC transfer.
type CategorisedColour struct {
	RGB        RGBColour `json:"rgb"`
	Hex        string    `json:"hex"`
	Role       string    `json:"role,omitempty"`
	Luminance  float64   `json:"luminance,omitempty"`
	IsLight    bool      `json:"is_light,omitempty"`
	Hue        float64   `json:"hue,omitempty"`
	Saturation float64   `json:"saturation,omitempty"`
	Index      int       `json:"index,omitempty"`
}

// RGBColour represents an RGB color.
type RGBColour struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// InputPluginRPC implements the go-plugin Plugin interface for input plugins.
type InputPluginRPC struct {
	plugin.Plugin
	Impl InputPlugin
}

// Server returns an RPC server for this plugin.
func (p *InputPluginRPC) Server(*plugin.MuxBroker) (interface{}, error) {
	return &InputPluginRPCServer{Impl: p.Impl}, nil
}

// Client returns an RPC client for this plugin.
func (p *InputPluginRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &InputPluginRPCClient{client: c}, nil
}

// InputPluginRPCServer is the RPC server implementation for input plugins.
type InputPluginRPCServer struct {
	Impl InputPlugin
}

// Generate implements the RPC method for palette generation.
func (s *InputPluginRPCServer) Generate(opts InputOptions, resp *[]byte) error {
	colors, err := s.Impl.Generate(context.Background(), opts)
	if err != nil {
		return err
	}

	// Convert to JSON-compatible format.
	result := make([]map[string]uint8, len(colors))
	for i, c := range colors {
		r, g, b, _ := c.RGBA()
		result[i] = map[string]uint8{
			"r": uint8(r >> 8),
			"g": uint8(g >> 8),
			"b": uint8(b >> 8),
		}
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	*resp = data
	return nil
}

// GetMetadata implements the RPC method for fetching plugin metadata.
func (s *InputPluginRPCServer) GetMetadata(_ interface{}, resp *PluginInfo) error {
	*resp = s.Impl.GetMetadata()
	return nil
}

// InputPluginRPCClient is the RPC client implementation for input plugins.
type InputPluginRPCClient struct {
	client *rpc.Client
}

// Generate calls the remote Generate method.
func (c *InputPluginRPCClient) Generate(ctx context.Context, opts InputOptions) ([]color.Color, error) {
	var respBytes []byte
	err := c.client.Call("Plugin.Generate", opts, &respBytes)
	if err != nil {
		return nil, err
	}

	var result []struct {
		R uint8 `json:"r"`
		G uint8 `json:"g"`
		B uint8 `json:"b"`
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		return nil, err
	}

	colors := make([]color.Color, len(result))
	for i, rgb := range result {
		colors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
	}

	return colors, nil
}

// GetMetadata calls the remote GetMetadata method.
func (c *InputPluginRPCClient) GetMetadata() (PluginInfo, error) {
	var info PluginInfo
	err := c.client.Call("Plugin.GetMetadata", new(interface{}), &info)
	return info, err
}

// OutputPluginRPC implements the go-plugin Plugin interface for output plugins.
type OutputPluginRPC struct {
	plugin.Plugin
	Impl OutputPlugin
}

// Server returns an RPC server for this plugin.
func (p *OutputPluginRPC) Server(*plugin.MuxBroker) (interface{}, error) {
	return &OutputPluginRPCServer{Impl: p.Impl}, nil
}

// Client returns an RPC client for this plugin.
func (p *OutputPluginRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &OutputPluginRPCClient{client: c}, nil
}

// OutputPluginRPCServer is the RPC server implementation for output plugins.
type OutputPluginRPCServer struct {
	Impl OutputPlugin
}

// Generate implements the RPC method for output generation.
func (s *OutputPluginRPCServer) Generate(palette PaletteData, resp *map[string][]byte) error {
	result, err := s.Impl.Generate(context.Background(), palette)
	if err != nil {
		return err
	}
	*resp = result
	return nil
}

// PreExecute implements the RPC method for pre-execution hooks.
func (s *OutputPluginRPCServer) PreExecute(_ interface{}, resp *struct {
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
func (s *OutputPluginRPCServer) GetMetadata(_ interface{}, resp *PluginInfo) error {
	*resp = s.Impl.GetMetadata()
	return nil
}

// OutputPluginRPCClient is the RPC client implementation for output plugins.
type OutputPluginRPCClient struct {
	client *rpc.Client
}

// Generate calls the remote Generate method.
func (c *OutputPluginRPCClient) Generate(ctx context.Context, palette PaletteData) (map[string][]byte, error) {
	var result map[string][]byte
	err := c.client.Call("Plugin.Generate", palette, &result)
	return result, err
}

// PreExecute calls the remote PreExecute method.
func (c *OutputPluginRPCClient) PreExecute(ctx context.Context) (bool, string, error) {
	var resp struct {
		Skip   bool
		Reason string
		Error  string
	}
	err := c.client.Call("Plugin.PreExecute", new(interface{}), &resp)
	if err != nil {
		return false, "", err
	}
	if resp.Error != "" {
		return resp.Skip, resp.Reason, &RPCError{Message: resp.Error}
	}
	return resp.Skip, resp.Reason, nil
}

// PostExecute calls the remote PostExecute method.
func (c *OutputPluginRPCClient) PostExecute(ctx context.Context, files []string) error {
	var errMsg string
	err := c.client.Call("Plugin.PostExecute", files, &errMsg)
	if err != nil {
		return err
	}
	if errMsg != "" {
		return &RPCError{Message: errMsg}
	}
	return nil
}

// GetMetadata calls the remote GetMetadata method.
func (c *OutputPluginRPCClient) GetMetadata() (PluginInfo, error) {
	var info PluginInfo
	err := c.client.Call("Plugin.GetMetadata", new(interface{}), &info)
	return info, err
}

// RPCError represents an error returned from an RPC call.
type RPCError struct {
	Message string
}

// Error implements the error interface.
func (e *RPCError) Error() string {
	return e.Message
}

// ConvertCategorisedPalette converts a colour.CategorisedPalette to PaletteData for RPC.
func ConvertCategorisedPalette(palette *colour.CategorisedPalette, pluginArgs map[string]any, dryRun bool) PaletteData {
	data := PaletteData{
		Colours:    make(map[string]CategorisedColour),
		AllColours: make([]CategorisedColour, len(palette.AllColours)),
		ThemeType:  palette.ThemeType.String(),
		PluginArgs: pluginArgs,
		DryRun:     dryRun,
	}

	// Convert colours map.
	for role, cc := range palette.Colours {
		data.Colours[string(role)] = CategorisedColour{
			RGB: RGBColour{
				R: cc.RGB.R,
				G: cc.RGB.G,
				B: cc.RGB.B,
			},
			Hex:        cc.Hex,
			Role:       string(cc.Role),
			Luminance:  cc.Luminance,
			IsLight:    cc.IsLight,
			Hue:        cc.Hue,
			Saturation: cc.Saturation,
			Index:      cc.Index,
		}
	}

	// Convert all colours slice.
	for i, cc := range palette.AllColours {
		data.AllColours[i] = CategorisedColour{
			RGB: RGBColour{
				R: cc.RGB.R,
				G: cc.RGB.G,
				B: cc.RGB.B,
			},
			Hex:        cc.Hex,
			Role:       string(cc.Role),
			Luminance:  cc.Luminance,
			IsLight:    cc.IsLight,
			Hue:        cc.Hue,
			Saturation: cc.Saturation,
			Index:      cc.Index,
		}
	}

	return data
}
