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

	// MinCompatibleVersion is the oldest protocol version this tinct
	// version can work with for any plugin type.
	//
	// Every RPC the host calls unconditionally on an *output* plugin
	// (Generate, GetMetadata, PreExecute, PostExecute, GetFlagHelp) has
	// existed since before the 0.1.0 bump, and the optional 0.3.0+
	// methods degrade gracefully via isMissingMethodErr. So old output
	// plugins remain genuinely usable and are not rejected.
	MinCompatibleVersion = "0.0.1"

	// MinInputCompatibleVersion is the floor for input plugins.
	//
	// Higher than the general floor because WallpaperRawPath was added
	// mid-cycle — after the 0.1.0 bump, before 0.2.0 — so a plugin
	// reporting "0.1.0" may or may not implement it. The host calls it
	// unconditionally after Generate, and an absent method does not fail
	// cleanly: net/rpc blocks on the reply forever. 0.2.0 is the oldest
	// version at which its presence is guaranteed.
	MinInputCompatibleVersion = "0.2.0"
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
