// Package plugin provides the public API for tinct plugins.
package plugin

import (
	"context"
	"encoding/json"
	"image/color"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
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
func (p *InputPluginRPC) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
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
func (s *InputPluginRPCServer) GetMetadata(_ any, resp *PluginInfo) error {
	*resp = s.Impl.GetMetadata()
	return nil
}

// WallpaperPath implements the RPC method for fetching wallpaper path.
func (s *InputPluginRPCServer) WallpaperPath(_ any, resp *string) error {
	*resp = s.Impl.WallpaperPath()
	return nil
}

// GetFlagHelp implements the RPC method for fetching flag help.
func (s *InputPluginRPCServer) GetFlagHelp(_ any, resp *[]FlagHelp) error {
	*resp = s.Impl.GetFlagHelp()
	return nil
}

// InputPluginRPCClient is the RPC client implementation for input plugins.
type InputPluginRPCClient struct {
	client *rpc.Client
}

// Generate calls the remote Generate method.
func (c *InputPluginRPCClient) Generate(_ context.Context, opts InputOptions) ([]color.Color, error) {
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
	err := c.client.Call("Plugin.GetMetadata", new(any), &info)
	return info, err
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
func (s *OutputPluginRPCServer) Generate(palette PaletteData, resp *map[string][]byte) error {
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

// OutputPluginRPCClient is the RPC client implementation for output plugins.
type OutputPluginRPCClient struct {
	client *rpc.Client
}

// Generate calls the remote Generate method.
func (c *OutputPluginRPCClient) Generate(_ context.Context, palette PaletteData) (map[string][]byte, error) {
	var result map[string][]byte
	err := c.client.Call("Plugin.Generate", palette, &result)
	return result, err
}

// PreExecute calls the remote PreExecute method.
func (c *OutputPluginRPCClient) PreExecute(_ context.Context) (bool, string, error) {
	var resp struct {
		Skip   bool
		Reason string
		Error  string
	}
	err := c.client.Call("Plugin.PreExecute", new(any), &resp)
	if err != nil {
		return false, "", err
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
	err := c.client.Call("Plugin.GetMetadata", new(any), &info)
	return info, err
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

// RPCError represents an error returned from an RPC call.
type RPCError struct {
	Message string
}

// Error implements the error interface.
func (e *RPCError) Error() string {
	return e.Message
}
