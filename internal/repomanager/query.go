package repomanager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jmylchreest/tinct/internal/compression"
)

// PluginMetadata contains queried plugin information.
type PluginMetadata struct {
	Name            string
	Type            string
	Version         string
	ProtocolVersion string
	Description     string
	Author          string
	Repository      string
	License         string
	Tags            []string
	Enabled         bool
}

// QueryPlugin downloads, extracts, and queries a plugin binary.
// Uses the shared compression package for format detection and extraction.
func QueryPlugin(downloadURL string) (*PluginMetadata, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "tinct-repo-query-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download file
	resp, err := http.Get(downloadURL) // #nosec G107 - URL is from user input, validated upstream
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response data into memory
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read download: %w", err)
	}

	// Get filename and content type for format detection
	filename := filepath.Base(downloadURL)
	contentType := resp.Header.Get("Content-Type")

	// Extract archive base name for plugin detection
	archiveName := compression.GetArchiveBaseName(filename)

	// Extract plugin using shared compression utilities
	result, err := compression.ExtractPlugin(
		data,
		downloadURL, // url for format detection
		filename,    // filename for fallback detection
		"",          // targetFile (empty = auto-detect)
		archiveName, // archiveName for matching plugin binary
		tempDir,     // destination directory
		contentType, // HTTP Content-Type for detection
		false,       // verbose
	)
	if err != nil {
		return nil, fmt.Errorf("extract failed: %w", err)
	}

	// Make executable (if not already)
	if err := os.Chmod(result.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to make plugin executable: %w", err)
	}

	// Query plugin info
	return queryPluginInfo(result.Path)
}

// QueryScriptPlugin queries a script-based plugin (no download needed).
func QueryScriptPlugin(scriptPath string) (*PluginMetadata, error) {
	// Make executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to make script executable: %w", err)
	}

	// Query plugin info
	return queryPluginInfo(scriptPath)
}

// queryPluginInfo executes a plugin and queries its metadata.
func queryPluginInfo(pluginPath string) (*PluginMetadata, error) {
	cmd := exec.Command(pluginPath, "--plugin-info")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("plugin-info query failed: %w", err)
	}

	// Parse JSON (use map to handle variations)
	var info map[string]any
	if err := json.Unmarshal(output, &info); err != nil {
		return nil, fmt.Errorf("invalid plugin-info JSON: %w", err)
	}

	// Extract fields with type assertions
	metadata := &PluginMetadata{
		Name:            getStringField(info, "name"),
		Type:            getStringField(info, "type"),
		Version:         getStringField(info, "version"),
		ProtocolVersion: getStringField(info, "protocol_version"),
		Description:     getStringField(info, "description"),
		Author:          getStringField(info, "author"),
		Repository:      getStringField(info, "repository"),
		License:         getStringField(info, "license"),
		Tags:            getStringSliceField(info, "tags"),
		Enabled:         getBoolField(info, "enabled"),
	}

	// Validate required fields
	if metadata.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	if metadata.Type == "" {
		return nil, fmt.Errorf("plugin type is required")
	}

	return metadata, nil
}

// CalculateCompatibility generates compatibility string from protocol version.
func CalculateCompatibility(protocolVersion string) string {
	if protocolVersion == "" {
		return ""
	}

	// Parse protocol version (e.g., "0.0.1")
	parts := strings.Split(protocolVersion, ".")
	if len(parts) < 2 {
		return ""
	}

	major := parts[0]
	minor := parts[1]

	// Generate compatibility string
	// Protocol 0.0.x requires tinct >= 0.0.1
	// Protocol 0.1.x requires tinct >= 0.1.0
	return fmt.Sprintf(">=%s.%s.0", major, minor)
}

// Helper functions

func getStringField(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolField(m map[string]any, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getStringSliceField(m map[string]any, key string) []string {
	if val, ok := m[key]; ok {
		if arr, ok := val.([]any); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return nil
}
