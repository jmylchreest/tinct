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
	path                  string
	protocolType          protocol.PluginType
	pluginProtocolVersion string // Protocol version reported by the plugin (e.g. "0.2.0")
	client                *goplug.Client
	rpcClient             any // Either *plugin.InputPluginRPCClient or *plugin.OutputPluginRPCClient
	verbose               bool
	lastWallpaperPath     string        // Stores canonical wallpaper path from JSON stdio plugins
	lastWallpaperRawPath  string        // Stores raw wallpaper path from JSON stdio plugins
	processRunner         ProcessRunner // Abstraction for running external processes
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
		path:                  pluginPath,
		protocolType:          result.Type,
		pluginProtocolVersion: result.PluginInfo.ProtocolVersion,
		verbose:               verbose,
		processRunner:         runner,
	}

	// For go-plugins, we initialize the RPC client lazily on first use
	// to avoid keeping connections open unnecessarily.

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

// wallpaperFromInput returns a wallpaper path from the input plugin, dispatching
// on the active protocol. For RPC plugins it calls rpcGetter on the dispensed
// InputPluginRPCClient (returning "" if no client is connected); for JSON stdio
// plugins it returns the stored value the executor cached on the last call.
func (e *PluginExecutor) wallpaperFromInput(rpcGetter func(*plugin.InputPluginRPCClient) string, jsonValue string) string {
	switch e.protocolType {
	case protocol.PluginTypeGoPlugin:
		if e.rpcClient == nil {
			return ""
		}
		if inputClient, ok := e.rpcClient.(*plugin.InputPluginRPCClient); ok {
			return rpcGetter(inputClient)
		}
		return ""
	case protocol.PluginTypeJSON:
		return jsonValue
	default:
		return ""
	}
}

// GetWallpaperPath retrieves the canonical wallpaper path from an input plugin if available.
// Works for both go-plugin RPC and JSON stdio protocols.
func (e *PluginExecutor) GetWallpaperPath() string {
	return e.wallpaperFromInput(
		(*plugin.InputPluginRPCClient).WallpaperPath,
		e.lastWallpaperPath,
	)
}

// GetWallpaperRawPath retrieves the raw wallpaper path from an input plugin if available.
// This is the literal path as provided by the user before any canonicalization.
// Works for both go-plugin RPC and JSON stdio protocols.
func (e *PluginExecutor) GetWallpaperRawPath() string {
	return e.wallpaperFromInput(
		(*plugin.InputPluginRPCClient).WallpaperRawPath,
		e.lastWallpaperRawPath,
	)
}

// --- Go-Plugin RPC implementations ---

// dispenseRPCPlugin creates a go-plugin client, connects via RPC, and
// dispenses the named plugin.  It is the shared core of getInputRPCClient
// and getOutputRPCClient.
func (e *PluginExecutor) dispenseRPCPlugin(name string, pluginMap map[string]goplug.Plugin) (any, error) {
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
		HandshakeConfig:  protocol.Handshake,
		Plugins:          pluginMap,
		Cmd:              exec.Command(e.path), //nolint:gosec // G204: Plugin path validated during installation and locked in plugin.lock
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
	raw, err := rpcClient.Dispense(name)
	if err != nil {
		e.client.Kill()
		return nil, fmt.Errorf("failed to dispense plugin: %w", err)
	}

	return raw, nil
}

// rpcClientFor returns the cached RPC client of type T if it matches, or
// dispenses a fresh one under kind (with the supplied goplug.Plugin) and
// caches it. Generics let getInputRPCClient and getOutputRPCClient share the
// same fetch/cast/cache flow without duplicating method bodies. Implemented
// as a free function because Go does not allow generic methods.
func rpcClientFor[T any](e *PluginExecutor, kind string, p goplug.Plugin) (T, error) {
	var zero T
	if e.rpcClient != nil {
		if existing, ok := e.rpcClient.(T); ok {
			return existing, nil
		}
	}

	raw, err := e.dispenseRPCPlugin(kind, map[string]goplug.Plugin{kind: p})
	if err != nil {
		return zero, err
	}

	client, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("unexpected plugin type")
	}
	e.rpcClient = client
	return client, nil
}

func (e *PluginExecutor) getInputRPCClient(_ context.Context) (*plugin.InputPluginRPCClient, error) {
	return rpcClientFor[*plugin.InputPluginRPCClient](e, "input", &plugin.InputPluginRPC{})
}

func (e *PluginExecutor) getOutputRPCClient(_ context.Context) (*plugin.OutputPluginRPCClient, error) {
	return rpcClientFor[*plugin.OutputPluginRPCClient](e, "output", &plugin.OutputPluginRPC{})
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

func (e *PluginExecutor) preExecuteGoPlugin(ctx context.Context) (skip bool, reason string, err error) {
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
		WallpaperPath    string `json:"wallpaper_path,omitempty"`
		WallpaperRawPath string `json:"wallpaper_raw_path,omitempty"`
	}

	if err := json.Unmarshal(stdoutBytes, &response); err == nil && len(response.Colors) > 0 {
		colors := make([]color.Color, len(response.Colors))
		for i, rgb := range response.Colors {
			colors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
		}
		// Store wallpaper paths if provided
		e.lastWallpaperPath = response.WallpaperPath
		e.lastWallpaperRawPath = response.WallpaperRawPath
		// If raw path not provided, default to canonical path
		if e.lastWallpaperRawPath == "" && e.lastWallpaperPath != "" {
			e.lastWallpaperRawPath = e.lastWallpaperPath
		}
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

// jsonStdioResponse is the structured response for protocol >= 0.2.0.
// Plugins write their own files and report results via this JSON envelope on stdout.
type jsonStdioResponse struct {
	Success      bool     `json:"success"`
	FilesWritten []string `json:"files_written"`
	Message      string   `json:"message"`
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

	// Version-gated response parsing.
	if e.supportsStructuredResponse() {
		return e.parseStructuredResponse(stdoutBytes)
	}

	// Legacy behavior (< 0.2.0): stdout is freeform text, display to stderr.
	// Return empty map — legacy plugins either write nothing or manage their own files.
	if len(stdoutBytes) > 0 {
		fmt.Fprint(os.Stderr, string(stdoutBytes))
	}
	return make(map[string][]byte), nil
}

// supportsStructuredResponse returns true if the plugin protocol version >= 0.2.0.
func (e *PluginExecutor) supportsStructuredResponse() bool {
	if e.pluginProtocolVersion == "" {
		return false
	}
	v, err := protocol.Parse(e.pluginProtocolVersion)
	if err != nil {
		return false
	}
	return v.AtLeast(protocol.Version{Major: 0, Minor: 2, Patch: 0})
}

// parseStructuredResponse parses the 0.2.0+ JSON response from a plugin.
// Plugins write their own files; the returned map contains the reported paths
// as keys with nil content (signalling tinct to track but NOT re-write them).
func (e *PluginExecutor) parseStructuredResponse(stdout []byte) (map[string][]byte, error) {
	if len(stdout) == 0 {
		return nil, fmt.Errorf("plugin returned empty response (protocol >= 0.2.0 requires JSON on stdout)")
	}

	var resp jsonStdioResponse
	if err := json.Unmarshal(stdout, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse plugin JSON response: %w\nRaw output: %s", err, string(stdout))
	}

	if !resp.Success {
		msg := resp.Message
		if msg == "" {
			msg = "plugin reported failure"
		}
		return nil, fmt.Errorf("plugin error: %s", msg)
	}

	result := make(map[string][]byte, len(resp.FilesWritten))
	for _, path := range resp.FilesWritten {
		// nil content signals that the plugin already wrote the file;
		// tinct should track it in the manifest but not re-write it.
		result[path] = nil
	}

	return result, nil
}

func (e *PluginExecutor) preExecuteJSON(ctx context.Context) (skip bool, reason string, err error) {
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

func (e *PluginExecutor) getFlagHelpJSON(_ context.Context) ([]input.FlagHelp, error) {
	// For JSON stdio plugins, we don't have a standard way to query flag help
	// This would require the plugin to support a --flag-help or similar command
	// For now, return empty array
	// TODO: Implement JSON stdio flag help protocol if needed
	return []input.FlagHelp{}, nil
}
