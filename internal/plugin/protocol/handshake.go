// Package protocol defines the plugin protocol version and compatibility checking.
package protocol

import (
	"github.com/hashicorp/go-plugin"
)

// Handshake is the handshake configuration for go-plugin protocol.
// This ensures that plugins using go-plugin can only connect to compatible hosts.
//
// NOTE: go-plugin's ProtocolVersion is a single uint that must match exactly.
// We use the major version from ProtocolVersion (defined in version.go) for this.
// The full semantic version checking (including MinCompatibleVersion) happens
// separately via the --plugin-info query and IsCompatible() function.
//
// This means:
// - go-plugin handshake: Major version must match exactly (e.g., both 0)
// - Semantic version check: Full MAJOR.MINOR.PATCH validation with MinCompatibleVersion
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  uint(GetCurrentVersion().Major),
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
