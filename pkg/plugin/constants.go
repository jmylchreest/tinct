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
	//
	// 0.3.0: Input plugins can optionally implement RoleHinter / ThemeHinter.
	// The Generate RPC response now uses an object envelope when hints are
	// present; the legacy bare array is still emitted (and accepted) when no
	// hints are provided, so 0.2.0 plugins continue to work unchanged.
	ProtocolVersion = "0.3.0"

	// MinCompatibleVersion is the oldest protocol version this tinct version can work with.
	// Unchanged at 0.0.1 because the wire-format change is purely additive — the
	// host still accepts the legacy bare-array Generate response.
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
