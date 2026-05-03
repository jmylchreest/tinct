package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// RuntimePaths holds all runtime file paths.
type RuntimePaths struct {
	Dir        string
	Pipe       string
	Config     string
	PID        string
	ConfigInfo string
}

// ConfigInfo tracks config sources for reload detection.
type ConfigInfo struct {
	BaseConfig    string   `json:"base_config"`
	AppendConfigs []string `json:"append_configs"`
}

// getRuntimePaths returns the runtime directory paths.
func getRuntimePaths() (*RuntimePaths, error) {
	var baseDir string

	if runtimeDir := os.Getenv("WOB_RUNTIME_DIR"); runtimeDir != "" {
		baseDir = runtimeDir
	} else if xdgRuntime := os.Getenv("XDG_RUNTIME_DIR"); xdgRuntime != "" {
		baseDir = filepath.Join(xdgRuntime, defaultRuntimeDir)
	} else {
		username := os.Getenv("USER")
		if username == "" {
			username = "unknown"
		}
		baseDir = filepath.Join(os.TempDir(), fmt.Sprintf("%s-wob-runtime", username))
	}

	// Create runtime directory if it doesn't exist.
	if err := os.MkdirAll(baseDir, 0o700); err != nil { //nolint:gosec // baseDir is XDG_RUNTIME_DIR or $TMPDIR/$USER-wob-runtime
		return nil, fmt.Errorf("failed to create runtime directory: %w", err)
	}

	// Verify directory ownership.
	fileInfo, err := os.Stat(baseDir) //nolint:gosec // baseDir is XDG_RUNTIME_DIR or $TMPDIR/$USER-wob-runtime
	if err != nil {
		return nil, fmt.Errorf("failed to stat runtime directory: %w", err)
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("unexpected stat type %T (need syscall.Stat_t)", fileInfo.Sys())
	}
	// Safe conversion: UIDs are always positive
	currentUID := os.Getuid()
	if currentUID < 0 {
		return nil, fmt.Errorf("invalid current UID")
	}
	if stat.Uid != uint32(currentUID) { // #nosec G115 - UID validated to be >= 0
		return nil, fmt.Errorf("runtime directory not owned by current user")
	}

	return &RuntimePaths{
		Dir:        baseDir,
		Pipe:       filepath.Join(baseDir, os.Getenv("WOB_PIPE")),
		Config:     filepath.Join(baseDir, os.Getenv("WOB_MERGED_CONFIG")),
		PID:        filepath.Join(baseDir, defaultPIDFile),
		ConfigInfo: filepath.Join(baseDir, "config.json"),
	}, nil
}

// saveConfigInfo saves config info to JSON file.
func saveConfigInfo(paths *RuntimePaths, info *ConfigInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(paths.ConfigInfo, data, 0o600)
}

// loadConfigInfo loads config info from JSON file.
func loadConfigInfo(paths *RuntimePaths) (*ConfigInfo, error) {
	data, err := os.ReadFile(paths.ConfigInfo)
	if err != nil {
		return nil, err
	}
	var info ConfigInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// needsConfigReload checks if any source configs are newer than merged config.
func needsConfigReload(paths *RuntimePaths, configInfo *ConfigInfo) bool {
	// Check if merged config exists
	mergedInfo, err := os.Stat(paths.Config)
	if err != nil {
		// Merged config doesn't exist - reload needed
		return true
	}
	mergedTime := mergedInfo.ModTime()

	// Check base config mtime
	if configInfo.BaseConfig != "" {
		baseInfo, err := os.Stat(configInfo.BaseConfig)
		if err == nil && baseInfo.ModTime().After(mergedTime) {
			return true
		}
	}

	// Check append configs mtime
	for _, appendPath := range configInfo.AppendConfigs {
		appendInfo, err := os.Stat(appendPath)
		if err == nil && appendInfo.ModTime().After(mergedTime) {
			return true
		}
	}

	return false
}

// mergeConfigs merges base and append configs into a single file.
// The user-supplied config paths are read and merged into the runtime
// dir verbatim — gosec's path-traversal/XSS taint analysis flags every
// path use here, but the inputs are intentionally user-controlled
// config file paths exactly as wob expects.
//
//nolint:gosec // user-supplied wob config paths read intentionally
func mergeConfigs(paths *RuntimePaths, baseConfig string, appendConfigs []string) (string, error) {
	tmpFile, err := os.CreateTemp(paths.Dir, ".wob-merged.*.ini")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Write header
	fmt.Fprintf(tmpFile, "# Auto-generated merged wob config (%s)\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(tmpFile, "# Base: %s\n\n", baseConfig)

	// Copy base config
	baseData, err := os.ReadFile(baseConfig) // #nosec G304 - User-specified config file, intended to be read
	if err != nil {
		return "", fmt.Errorf("failed to read base config: %w", err)
	}
	if _, err := tmpFile.Write(baseData); err != nil {
		return "", fmt.Errorf("failed to write base config: %w", err)
	}

	// Append additional configs
	for _, appendConfig := range appendConfigs {
		fmt.Fprintf(tmpFile, "\n# Append: %s\n", appendConfig)
		appendData, err := os.ReadFile(appendConfig) // #nosec G304 - User-specified config file, intended to be read
		if err != nil {
			return "", fmt.Errorf("failed to read append config %s: %w", appendConfig, err)
		}
		if _, err := tmpFile.Write(appendData); err != nil {
			return "", fmt.Errorf("failed to write append config %s: %w", appendConfig, err)
		}
	}

	tmpPath := tmpFile.Name()
	finalPath := paths.Config

	// Atomic rename
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath) // Ignore cleanup errors
		return "", fmt.Errorf("failed to rename merged config: %w", err)
	}

	if err := os.Chmod(finalPath, 0o600); err != nil {
		return "", fmt.Errorf("failed to chmod merged config: %w", err)
	}

	return finalPath, nil
}
