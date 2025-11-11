// Package protocol defines the plugin protocol version and compatibility checking.
package protocol

import (
	tinctplugin "github.com/jmylchreest/tinct/pkg/plugin"
)

// Handshake is an alias to the public plugin.Handshake configuration.
var Handshake = tinctplugin.Handshake

// PluginType is an alias to the public plugin.PluginType type.
type PluginType = tinctplugin.PluginType

const (
	// PluginTypeGoPlugin is an alias to the public constant.
	PluginTypeGoPlugin = tinctplugin.PluginTypeGoPlugin

	// PluginTypeJSON is an alias to the public constant.
	PluginTypeJSON = tinctplugin.PluginTypeJSON
)
