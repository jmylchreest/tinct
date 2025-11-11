// Package plugin provides the public API for tinct plugins.
package plugin

import (
	"github.com/hashicorp/go-plugin"
)

const (
	// ProtocolVersion defines the current plugin API version.
	// Format: MAJOR.MINOR.PATCH.
	// - Increment MAJOR for breaking changes (incompatible API changes).
	// - Increment MINOR for backward-compatible additions.
	// - Increment PATCH for backward-compatible bug fixes.
	ProtocolVersion = "0.0.1"

	// MinCompatibleVersion is the oldest protocol version this tinct version can work with.
	MinCompatibleVersion = "0.0.1"
)

// Handshake is the handshake configuration for go-plugin protocol.
// This ensures that plugins using go-plugin can only connect to compatible hosts.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  0, // Major version from ProtocolVersion
	MagicCookieKey:   "TINCT_PLUGIN",
	MagicCookieValue: "tinct_color_palette",
}

// PluginType defines the type of plugin communication protocol.
type PluginType string

const (
	// PluginTypeGoPlugin indicates the plugin uses HashiCorp go-plugin RPC protocol.
	PluginTypeGoPlugin PluginType = "go-plugin"

	// PluginTypeJSON indicates the plugin uses simple JSON over stdin/stdout.
	PluginTypeJSON PluginType = "json-stdio"
)
