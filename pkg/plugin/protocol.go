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
	Name            string    `json:"name"`
	Type            string    `json:"type"` // "input" or "output"
	Version         string    `json:"version"`
	ProtocolVersion string    `json:"protocol_version"`
	Description     string    `json:"description"`
	PluginProtocol  string    `json:"plugin_protocol"` // "json-stdio" or "go-plugin"
	Metadata        *Metadata `json:"metadata,omitempty"`
}

// Metadata is an optional, machine-checkable extension to PluginInfo.
// It mirrors the subset of hooks.Spec that matters to end users / docs
// (binary requirements, output paths, reload behaviour) in a form
// serialisable across the json-stdio boundary.
//
// Internal plugins do not need to populate this — the manager uses
// Hooks() directly. External plugins that opt in expose the same data
// via --plugin-info so README frontmatter can be diffed against
// runtime reality by the tinct-check-readmes tool.
type Metadata struct {
	RequiredBinaries []string        `json:"required_binaries,omitempty"`
	OptionalBinaries []string        `json:"optional_binaries,omitempty"`
	DefaultOutputDir string          `json:"default_output_dir,omitempty"` // ~ stays unexpanded
	GeneratedFiles   []string        `json:"generated_files,omitempty"`
	Pattern          string          `json:"pattern,omitempty"` // two-file | single-file | drop-in | flavor-pack
	Reload           *ReloadMetadata `json:"reload,omitempty"`
}

// ReloadMetadata describes the post-execute reload behaviour in a form
// that round-trips through JSON. Method matches the docs vocabulary;
// Command is purely descriptive (the exact thing the plugin runs).
type ReloadMetadata struct {
	Method             string `json:"method,omitempty"` // ipc | signal | watch | none
	Command            string `json:"command,omitempty"`
	UserActionRequired bool   `json:"user_action_required,omitempty"`
}
