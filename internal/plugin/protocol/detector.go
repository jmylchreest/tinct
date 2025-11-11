// Package protocol defines the plugin protocol version and compatibility checking.
package protocol

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query plugin info.
	cmd := exec.CommandContext(ctx, pluginPath, "--plugin-info")
	output, err := cmd.Output()
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

// IsGoPlugin checks if a plugin uses the go-plugin protocol.
// This is a quick check that just looks at the plugin_protocol field.
func IsGoPlugin(pluginPath string) bool {
	result, err := DetectProtocol(pluginPath)
	if err != nil {
		return false
	}
	return result.Type == PluginTypeGoPlugin
}

// QueryPluginInfo queries a plugin for its metadata without full protocol detection.
// This is useful when you just need basic info without determining the communication protocol.
func QueryPluginInfo(pluginPath string) (PluginInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, pluginPath, "--plugin-info")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return PluginInfo{}, fmt.Errorf("failed to query plugin: %w\nStderr: %s", err, stderr.String())
	}

	var info PluginInfo
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return PluginInfo{}, fmt.Errorf("failed to parse plugin info: %w", err)
	}

	return info, nil
}
