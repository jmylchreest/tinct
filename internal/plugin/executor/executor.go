// Package executor provides a unified interface for executing plugins
// regardless of their underlying protocol (go-plugin RPC or JSON-stdio).
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	goplug "github.com/hashicorp/go-plugin"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/input"
	"github.com/jmylchreest/tinct/internal/plugin/protocol"
	"github.com/jmylchreest/tinct/pkg/plugin"
)

// PluginExecutor provides a unified interface for executing plugins.
type PluginExecutor struct {
	path              string
	protocolType      protocol.PluginType
	client            *goplug.Client
	rpcClient         any // Either *plugin.InputPluginRPCClient or *plugin.OutputPluginRPCClient
	verbose           bool
	lastWallpaperPath string        // Stores wallpaper path from JSON stdio plugins
	processRunner     ProcessRunner // Abstraction for running external processes
}

// NewWithVerbose creates a new PluginExecutor with verbose logging control.
func NewWithVerbose(pluginPath string, verbose bool) (*PluginExecutor, error) {
	return NewWithVerboseAndRunner(pluginPath, verbose, NewRealProcessRunner())
}

// NewWithVerboseAndRunner creates a new PluginExecutor with a custom process runner.
// This constructor is primarily used for testing with mock process runners.
func NewWithVerboseAndRunner(pluginPath string, verbose bool, runner ProcessRunner) (*PluginExecutor, error) {
	// Detect protocol.
	result, err := protocol.DetectProtocol(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect plugin protocol: %w", err)
	}

	executor := &PluginExecutor{
		path:          pluginPath,
		protocolType:  result.Type,
		verbose:       verbose,
		processRunner: runner,
	}

	// If it's a go-plugin, initialize the RPC client.
	if result.Type == protocol.PluginTypeGoPlugin {
		// We'll initialize the client lazily on first use to avoid keeping
		// connections open unnecessarily.
	}

	return executor, nil
}

// ExecuteInput runs an input plugin and returns colors.
func (e *PluginExecutor) ExecuteInput(ctx context.Context, opts plugin.InputOptions) ([]color.Color, error) {
	switch e.protocolType {
	case protocol.PluginTypeGoPlugin:
		return e.executeInputGoPlugin(ctx, opts)
	case protocol.PluginTypeJSON:
		return e.executeInputJSON(ctx, opts)
	default:
		return nil, fmt.Errorf("unsupported protocol type: %s", e.protocolType)
	}
}

// ExecuteOutput runs an output plugin and returns generated files.
func (e *PluginExecutor) ExecuteOutput(ctx context.Context, palette plugin.PaletteData) (map[string][]byte, error) {
	switch e.protocolType {
	case protocol.PluginTypeGoPlugin:
		return e.executeOutputGoPlugin(ctx, palette)
	case protocol.PluginTypeJSON:
		return e.executeOutputJSON(ctx, palette)
	default:
		return nil, fmt.Errorf("unsupported protocol type: %s", e.protocolType)
	}
}

// PreExecute runs the output plugin's pre-execution hook.
func (e *PluginExecutor) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	switch e.protocolType {
	case protocol.PluginTypeGoPlugin:
		return e.preExecuteGoPlugin(ctx)
	case protocol.PluginTypeJSON:
		return e.preExecuteJSON(ctx)
	default:
		return false, "", fmt.Errorf("unsupported protocol type: %s", e.protocolType)
	}
}

// PostExecute runs the output plugin's post-execution hook.
func (e *PluginExecutor) PostExecute(ctx context.Context, writtenFiles []string) error {
	switch e.protocolType {
	case protocol.PluginTypeGoPlugin:
		return e.postExecuteGoPlugin(ctx, writtenFiles)
	case protocol.PluginTypeJSON:
		return e.postExecuteJSON(ctx, writtenFiles)
	default:
		return fmt.Errorf("unsupported protocol type: %s", e.protocolType)
	}
}

// Close cleans up any resources held by the executor.
func (e *PluginExecutor) Close() {
	if e.client != nil {
		e.client.Kill()
		e.client = nil
		e.rpcClient = nil
	}
}

// GetFlagHelp retrieves flag help information from a plugin.
// Works for both go-plugin RPC and JSON stdio protocols.
func (e *PluginExecutor) GetFlagHelp(ctx context.Context) ([]input.FlagHelp, error) {
	switch e.protocolType {
	case protocol.PluginTypeGoPlugin:
		return e.getFlagHelpGoPlugin(ctx)
	case protocol.PluginTypeJSON:
		return e.getFlagHelpJSON(ctx)
	default:
		return nil, fmt.Errorf("unsupported protocol type: %s", e.protocolType)
	}
}

// GetWallpaperPath retrieves the wallpaper path from an input plugin if available.
// Works for both go-plugin RPC and JSON stdio protocols.
func (e *PluginExecutor) GetWallpaperPath() string {
	switch e.protocolType {
	case protocol.PluginTypeGoPlugin:
		// For RPC plugins, query via RPC
		if e.rpcClient == nil {
			return ""
		}

		if inputClient, ok := e.rpcClient.(*plugin.InputPluginRPCClient); ok {
			return inputClient.WallpaperPath()
		}

		return ""

	case protocol.PluginTypeJSON:
		// For JSON stdio plugins, return stored value
		return e.lastWallpaperPath

	default:
		return ""
	}
}

// --- Go-Plugin RPC implementations ---

func (e *PluginExecutor) getInputRPCClient(ctx context.Context) (*plugin.InputPluginRPCClient, error) {
	if e.rpcClient != nil {
		if client, ok := e.rpcClient.(*plugin.InputPluginRPCClient); ok {
			return client, nil
		}
	}

	// Configure logger based on verbose flag.
	var logger hclog.Logger
	if e.verbose {
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Output: log.Writer(),
			Level:  hclog.Debug,
		})
	} else {
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Output: io.Discard,
			Level:  hclog.Off,
		})
	}

	// Initialize go-plugin client.
	e.client = goplug.NewClient(&goplug.ClientConfig{
		HandshakeConfig: protocol.Handshake,
		Plugins: map[string]goplug.Plugin{
			"input": &plugin.InputPluginRPC{},
		},
		Cmd:              exec.Command(e.path),
		AllowedProtocols: []goplug.Protocol{goplug.ProtocolNetRPC},
		Logger:           logger,
		SyncStderr:       os.Stderr, // Forward plugin stderr to parent
	})

	// Connect via RPC.
	rpcClient, err := e.client.Client()
	if err != nil {
		e.client.Kill()
		return nil, fmt.Errorf("failed to get RPC client: %w", err)
	}

	// Request the plugin.
	raw, err := rpcClient.Dispense("input")
	if err != nil {
		e.client.Kill()
		return nil, fmt.Errorf("failed to dispense plugin: %w", err)
	}

	client, ok := raw.(*plugin.InputPluginRPCClient)
	if !ok {
		return nil, fmt.Errorf("unexpected plugin type")
	}
	e.rpcClient = client

	return client, nil
}

func (e *PluginExecutor) getOutputRPCClient(ctx context.Context) (*plugin.OutputPluginRPCClient, error) {
	if e.rpcClient != nil {
		if client, ok := e.rpcClient.(*plugin.OutputPluginRPCClient); ok {
			return client, nil
		}
	}

	// Configure logger based on verbose flag.
	var logger hclog.Logger
	if e.verbose {
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Output: log.Writer(),
			Level:  hclog.Debug,
		})
	} else {
		logger = hclog.New(&hclog.LoggerOptions{
			Name:   "plugin",
			Output: io.Discard,
			Level:  hclog.Off,
		})
	}

	// Initialize go-plugin client.
	e.client = goplug.NewClient(&goplug.ClientConfig{
		HandshakeConfig: protocol.Handshake,
		Plugins: map[string]goplug.Plugin{
			"output": &plugin.OutputPluginRPC{},
		},
		Cmd:              exec.Command(e.path),
		AllowedProtocols: []goplug.Protocol{goplug.ProtocolNetRPC},
		Logger:           logger,
		SyncStderr:       os.Stderr, // Forward plugin stderr to parent
	})

	// Connect via RPC.
	rpcClient, err := e.client.Client()
	if err != nil {
		e.client.Kill()
		return nil, fmt.Errorf("failed to get RPC client: %w", err)
	}

	// Request the plugin.
	raw, err := rpcClient.Dispense("output")
	if err != nil {
		e.client.Kill()
		return nil, fmt.Errorf("failed to dispense plugin: %w", err)
	}

	client, ok := raw.(*plugin.OutputPluginRPCClient)
	if !ok {
		return nil, fmt.Errorf("unexpected plugin type")
	}
	e.rpcClient = client

	return client, nil
}

func (e *PluginExecutor) executeInputGoPlugin(ctx context.Context, opts plugin.InputOptions) ([]color.Color, error) {
	client, err := e.getInputRPCClient(ctx)
	if err != nil {
		return nil, err
	}

	return client.Generate(ctx, opts)
}

func (e *PluginExecutor) executeOutputGoPlugin(ctx context.Context, palette plugin.PaletteData) (map[string][]byte, error) {
	client, err := e.getOutputRPCClient(ctx)
	if err != nil {
		return nil, err
	}

	return client.Generate(ctx, palette)
}

func (e *PluginExecutor) preExecuteGoPlugin(ctx context.Context) (bool, string, error) {
	client, err := e.getOutputRPCClient(ctx)
	if err != nil {
		return false, "", err
	}

	return client.PreExecute(ctx)
}

func (e *PluginExecutor) postExecuteGoPlugin(ctx context.Context, writtenFiles []string) error {
	client, err := e.getOutputRPCClient(ctx)
	if err != nil {
		return err
	}

	return client.PostExecute(ctx, writtenFiles)
}

func (e *PluginExecutor) getFlagHelpGoPlugin(ctx context.Context) ([]input.FlagHelp, error) {
	// Try input client first
	if inputClient, err := e.getInputRPCClient(ctx); err == nil {
		return inputClient.GetFlagHelp(), nil
	}

	// Try output client
	if outputClient, err := e.getOutputRPCClient(ctx); err == nil {
		return outputClient.GetFlagHelp(), nil
	}

	return []input.FlagHelp{}, nil
}

// --- JSON-stdio implementations ---

func (e *PluginExecutor) executeInputJSON(ctx context.Context, opts plugin.InputOptions) ([]color.Color, error) {
	// Convert to JSON.
	optsJSON, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	// Execute plugin using the process runner.
	stdoutBytes, stderrBytes, err := e.processRunner.Run(ctx, e.path, nil, bytes.NewReader(optsJSON))
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w\nStderr: %s", err, string(stderrBytes))
	}

	// Parse output - try new format with wallpaper path first
	var response struct {
		Colors []struct {
			R uint8 `json:"r"`
			G uint8 `json:"g"`
			B uint8 `json:"b"`
		} `json:"colors"`
		WallpaperPath string `json:"wallpaper_path,omitempty"`
	}

	if err := json.Unmarshal(stdoutBytes, &response); err == nil && len(response.Colors) > 0 {
		colors := make([]color.Color, len(response.Colors))
		for i, rgb := range response.Colors {
			colors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
		}
		// Store wallpaper path if provided
		e.lastWallpaperPath = response.WallpaperPath
		return colors, nil
	}

	// Try simple color array (backwards compatibility).
	var rawColors []struct {
		R uint8 `json:"r"`
		G uint8 `json:"g"`
		B uint8 `json:"b"`
	}

	if err := json.Unmarshal(stdoutBytes, &rawColors); err == nil {
		colors := make([]color.Color, len(rawColors))
		for i, rgb := range rawColors {
			colors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
		}
		return colors, nil
	}

	// Try categorised palette format (backwards compatibility).
	var categorised colour.CategorisedPalette
	if err := json.Unmarshal(stdoutBytes, &categorised); err == nil {
		colors := make([]color.Color, 0)

		if len(categorised.AllColours) > 0 {
			for _, cc := range categorised.AllColours {
				colors = append(colors, color.RGBA{
					R: cc.RGB.R,
					G: cc.RGB.G,
					B: cc.RGB.B,
					A: 255,
				})
			}
		} else {
			for _, cc := range categorised.Colours {
				colors = append(colors, color.RGBA{
					R: cc.RGB.R,
					G: cc.RGB.G,
					B: cc.RGB.B,
					A: 255,
				})
			}
		}

		return colors, nil
	}

	return nil, fmt.Errorf("failed to parse plugin output\nOutput: %s", string(stdoutBytes))
}

func (e *PluginExecutor) executeOutputJSON(ctx context.Context, palette plugin.PaletteData) (map[string][]byte, error) {
	// Convert to JSON.
	paletteJSON, err := json.Marshal(palette)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal palette: %w", err)
	}

	// Execute plugin using the process runner.
	stdoutBytes, stderrBytes, err := e.processRunner.Run(ctx, e.path, nil, bytes.NewReader(paletteJSON))
	if err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w\nStderr: %s", err, string(stderrBytes))
	}

	// Return stdout as virtual file.
	result := make(map[string][]byte)
	if len(stdoutBytes) > 0 {
		result["output.txt"] = stdoutBytes
	}

	return result, nil
}

func (e *PluginExecutor) preExecuteJSON(ctx context.Context) (bool, string, error) {
	execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stdoutBytes, stderrBytes, err := e.processRunner.Run(execCtx, e.path, []string{"--pre-execute"}, nil)

	// Exit code 0 = continue, 1 = skip, 2+ = error.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exitCode := exitErr.ExitCode()

		if exitCode == 1 {
			reason := strings.TrimSpace(string(stdoutBytes))
			if reason == "" {
				reason = "plugin requested skip"
			}
			return true, reason, nil
		}

		errMsg := strings.TrimSpace(string(stderrBytes))
		if errMsg == "" {
			errMsg = fmt.Sprintf("exit code %d", exitCode)
		}
		return false, "", fmt.Errorf("pre-execute failed: %s", errMsg)
	}

	return false, "", nil
}

func (e *PluginExecutor) postExecuteJSON(ctx context.Context, writtenFiles []string) error {
	execCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	filesJSON, err := json.Marshal(map[string]any{
		"written_files": writtenFiles,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal files: %w", err)
	}

	_, stderrBytes, err := e.processRunner.Run(execCtx, e.path, []string{"--post-execute"}, bytes.NewReader(filesJSON))
	if err != nil {
		errMsg := strings.TrimSpace(string(stderrBytes))
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("post-execute failed: %s", errMsg)
	}

	return nil
}

func (e *PluginExecutor) getFlagHelpJSON(ctx context.Context) ([]input.FlagHelp, error) {
	// For JSON stdio plugins, we don't have a standard way to query flag help
	// This would require the plugin to support a --flag-help or similar command
	// For now, return empty array
	// TODO: Implement JSON stdio flag help protocol if needed
	return []input.FlagHelp{}, nil
}
