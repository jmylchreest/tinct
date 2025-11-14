package main

// This file contains the gRPC plugin interface implementation.
// The wob plugin uses json-stdio protocol (in main.go) for wrapper functionality.
// This gRPC code is not used and exists only for protocol compatibility.

import (
	"context"

	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

// WobPlugin implements the tinctplugin.OutputPlugin interface
type WobPlugin struct{}

// Generate creates wob theme files
func (p *WobPlugin) Generate(ctx context.Context, palette tinctplugin.PaletteData) (map[string][]byte, error) {
	// Not used - json-stdio protocol is used instead (see main.go)
	return nil, nil
}

// PreExecute checks if wob binary exists
func (p *WobPlugin) PreExecute(ctx context.Context) (bool, string, error) {
	// Not used - json-stdio protocol is used instead (see main.go)
	return false, "", nil
}

// PostExecute reports what was done and installs wrapper
func (p *WobPlugin) PostExecute(ctx context.Context, writtenFiles []string) error {
	// Not used - json-stdio protocol is used instead (see main.go)
	return nil
}

// GetMetadata returns plugin metadata
func (p *WobPlugin) GetMetadata() tinctplugin.PluginInfo {
	return tinctplugin.PluginInfo{
		Name:            pluginName,
		Type:            "output",
		Version:         Version,
		ProtocolVersion: tinctplugin.ProtocolVersion,
		Description:     pluginDescription,
		PluginProtocol:  "json-stdio",
	}
}

// GetFlagHelp returns help information for plugin flags
func (p *WobPlugin) GetFlagHelp() []tinctplugin.FlagHelp {
	// Wob plugin doesn't accept any command-line flags
	return []tinctplugin.FlagHelp{}
}
