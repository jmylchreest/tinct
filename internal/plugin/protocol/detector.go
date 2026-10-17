// Package protocol defines the plugin protocol version and compatibility checking.
package protocol

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmylchreest/tinct/internal/plugin/metadata"
	"github.com/jmylchreest/tinct/pkg/plugin"
)

// DetectorResult contains information about a detected plugin protocol.
type DetectorResult struct {
	// Type indicates which protocol the plugin uses.
	Type PluginType

	// SupportsGoPlugin indicates if the plugin binary has go-plugin support.
	SupportsGoPlugin bool

	// PluginInfo contains metadata from --plugin-info.
	PluginInfo PluginInfo
}

// PluginInfo is a type alias to the public plugin.PluginInfo type.
// External plugins should import github.com/jmylchreest/tinct/pkg/plugin directly.
type PluginInfo = plugin.PluginInfo

// DetectProtocol detects which protocol a plugin uses by querying it.
func DetectProtocol(pluginPath string) (*DetectorResult, error) {
	output, err := metadata.Query(context.Background(), pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to query plugin: %w", err)
	}

	var info PluginInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("failed to parse plugin info: %w", err)
	}

	result := &DetectorResult{
		PluginInfo: info,
	}

	// Determine protocol type from plugin_protocol field.
	switch info.PluginProtocol {
	case "go-plugin":
		result.Type = PluginTypeGoPlugin
		result.SupportsGoPlugin = true
	case "json-stdio", "":
		// Empty defaults to json-stdio for backward compatibility.
		result.Type = PluginTypeJSON
		result.SupportsGoPlugin = false
	default:
		return nil, fmt.Errorf("unknown plugin_protocol: %s", info.PluginProtocol)
	}

	return result, nil
}
