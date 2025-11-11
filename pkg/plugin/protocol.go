// Package plugin provides the public API for tinct plugins.
// External plugins should import this package instead of internal packages.
package plugin

// FlagHelp represents help information for a single plugin flag.
// This type is part of the plugin protocol and is used by both internal and external plugins.
type FlagHelp struct {
	Name        string `json:"name"`        // Flag name (e.g., "prompt", "model")
	Shorthand   string `json:"shorthand"`   // Short flag (e.g., "p")
	Type        string `json:"type"`        // Type (e.g., "string", "int", "bool")
	Default     string `json:"default"`     // Default value as string
	Description string `json:"description"` // Help text
	Required    bool   `json:"required"`    // Is this flag required?
}

// PluginInfo contains metadata about a plugin.
type PluginInfo struct {
	Name            string `json:"name"`
	Type            string `json:"type"` // "input" or "output"
	Version         string `json:"version"`
	ProtocolVersion string `json:"protocol_version"`
	Description     string `json:"description"`
	PluginProtocol  string `json:"plugin_protocol"` // "json-stdio" or "go-plugin"
}
