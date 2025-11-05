// Package executor provides a unified interface for executing plugins
// regardless of their underlying protocol (go-plugin RPC or JSON-stdio).
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/protocol"
)

// PluginExecutor provides a unified interface for executing plugins.
type PluginExecutor struct {
	path         string
	protocolType protocol.PluginType
	client       *plugin.Client
	rpcClient    interface{} // Either *protocol.InputPluginRPCClient or *protocol.OutputPluginRPCClient
	verbose      bool
}

// New creates a new PluginExecutor by detecting the plugin's protocol.
func New(pluginPath string) (*PluginExecutor, error) {
	return NewWithVerbose(pluginPath, false)
}

// NewWithVerbose creates a new PluginExecutor with verbose logging control.
func NewWithVerbose(pluginPath string, verbose bool) (*PluginExecutor, error) {
	// Detect protocol.
	result, err := protocol.DetectProtocol(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("failed to detect plugin protocol: %w", err)
	}

	executor := &PluginExecutor{
		path:         pluginPath,
		protocolType: result.Type,
		verbose:      verbose,
	}

	// If it's a go-plugin, initialize the RPC client.
	if result.Type == protocol.PluginTypeGoPlugin {
		// We'll initialize the client lazily on first use to avoid keeping
		// connections open unnecessarily.
	}

	return executor, nil
}

// ExecuteInput runs an input plugin and returns colors.
func (e *PluginExecutor) ExecuteInput(ctx context.Context, opts protocol.InputOptions) ([]color.Color, error) {
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
func (e *PluginExecutor) ExecuteOutput(ctx context.Context, palette protocol.PaletteData) (map[string][]byte, error) {
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

// --- Go-Plugin RPC implementations ---

func (e *PluginExecutor) getInputRPCClient(ctx context.Context) (*protocol.InputPluginRPCClient, error) {
	if e.rpcClient != nil {
		if client, ok := e.rpcClient.(*protocol.InputPluginRPCClient); ok {
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
	e.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: protocol.Handshake,
		Plugins: map[string]plugin.Plugin{
			"input": &protocol.InputPluginRPC{},
		},
		Cmd:              exec.Command(e.path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
		Logger:           logger,
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

	client := raw.(*protocol.InputPluginRPCClient)
	e.rpcClient = client

	return client, nil
}

func (e *PluginExecutor) getOutputRPCClient(ctx context.Context) (*protocol.OutputPluginRPCClient, error) {
	if e.rpcClient != nil {
		if client, ok := e.rpcClient.(*protocol.OutputPluginRPCClient); ok {
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
	e.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: protocol.Handshake,
		Plugins: map[string]plugin.Plugin{
			"output": &protocol.OutputPluginRPC{},
		},
		Cmd:              exec.Command(e.path),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
		Logger:           logger,
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

	client := raw.(*protocol.OutputPluginRPCClient)
	e.rpcClient = client

	return client, nil
}

func (e *PluginExecutor) executeInputGoPlugin(ctx context.Context, opts protocol.InputOptions) ([]color.Color, error) {
	client, err := e.getInputRPCClient(ctx)
	if err != nil {
		return nil, err
	}

	return client.Generate(ctx, opts)
}

func (e *PluginExecutor) executeOutputGoPlugin(ctx context.Context, palette protocol.PaletteData) (map[string][]byte, error) {
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

// --- JSON-stdio implementations ---

func (e *PluginExecutor) executeInputJSON(ctx context.Context, opts protocol.InputOptions) ([]color.Color, error) {
	// Convert to JSON.
	optsJSON, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	// Execute plugin.
	cmd := exec.CommandContext(ctx, e.path)
	cmd.Stdin = bytes.NewReader(optsJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w\nStderr: %s", err, stderr.String())
	}

	// Parse output - try simple color array first.
	var rawColors []struct {
		R uint8 `json:"r"`
		G uint8 `json:"g"`
		B uint8 `json:"b"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &rawColors); err == nil {
		colors := make([]color.Color, len(rawColors))
		for i, rgb := range rawColors {
			colors[i] = color.RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: 255}
		}
		return colors, nil
	}

	// Try categorised palette format (backwards compatibility).
	var categorised colour.CategorisedPalette
	if err := json.Unmarshal(stdout.Bytes(), &categorised); err == nil {
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

	return nil, fmt.Errorf("failed to parse plugin output\nOutput: %s", stdout.String())
}

func (e *PluginExecutor) executeOutputJSON(ctx context.Context, palette protocol.PaletteData) (map[string][]byte, error) {
	// Convert to JSON.
	paletteJSON, err := json.Marshal(palette)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal palette: %w", err)
	}

	// Execute plugin.
	cmd := exec.CommandContext(ctx, e.path)
	cmd.Stdin = bytes.NewReader(paletteJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w\nStderr: %s", err, stderr.String())
	}

	// Return stdout as virtual file.
	result := make(map[string][]byte)
	if stdout.Len() > 0 {
		result["output.txt"] = stdout.Bytes()
	}

	return result, nil
}

func (e *PluginExecutor) preExecuteJSON(ctx context.Context) (bool, string, error) {
	execCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(execCtx, e.path, "--pre-execute")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Exit code 0 = continue, 1 = skip, 2+ = error.
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()

		if exitCode == 1 {
			reason := strings.TrimSpace(stdout.String())
			if reason == "" {
				reason = "plugin requested skip"
			}
			return true, reason, nil
		}

		errMsg := strings.TrimSpace(stderr.String())
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

	cmd := exec.CommandContext(execCtx, e.path, "--post-execute")
	cmd.Stdin = bytes.NewReader(filesJSON)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("post-execute failed: %s", errMsg)
	}

	return nil
}
