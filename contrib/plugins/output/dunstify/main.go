// dunstify - Tinct Output Plugin for Desktop Notifications
//
// DEMONSTRATION GO-PLUGIN: This plugin showcases the go-plugin protocol.
// It demonstrates how to create a simple notification plugin using RPC communication
// instead of JSON-stdio, providing better performance through process reuse.
//
// Functionality:
// - Sends desktop notifications via dunstify or notify-send when themes are generated
// - Automatically detects which notification binary is available
// - Falls back gracefully if neither is found
// - Uses go-plugin RPC protocol for performance
//
// Build:
//   go build -o dunstify main.go
//
// Usage:
//   tinct plugins add ./dunstify --type output
//   tinct plugins enable dunstify
//   tinct generate -i image -o dunstify,hyprland
//
// Author: Tinct Contributors
// License: MIT

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"github.com/jmylchreest/tinct/internal/plugin/protocol"
)

// DunstifyPlugin sends desktop notifications when themes are generated.
// This is a demonstration of the go-plugin protocol for Tinct plugins.
type DunstifyPlugin struct{}

// Generate is a no-op for notification plugins.
// Notification plugins typically don't generate files, they only act in PostExecute.
// This demonstrates that not all plugin methods need complex implementations.
func (p *DunstifyPlugin) Generate(ctx context.Context, palette protocol.PaletteData) (map[string][]byte, error) {
	// Notification plugins don't generate files
	return map[string][]byte{}, nil
}

// PreExecute checks if dunstify or notify-send is available.
// This demonstrates the PreExecute hook which can skip plugin execution
// if prerequisites aren't met. Return (true, reason, nil) to skip gracefully.
func (p *DunstifyPlugin) PreExecute(ctx context.Context) (skip bool, reason string, err error) {
	// Try dunstify first
	if _, err := exec.LookPath("dunstify"); err == nil {
		return false, "", nil
	}

	// Fall back to notify-send
	if _, err := exec.LookPath("notify-send"); err == nil {
		return false, "", nil
	}

	// Neither available - skip plugin gracefully
	return true, "Neither dunstify nor notify-send found on $PATH", nil
}

// PostExecute sends a desktop notification with theme generation details.
// This demonstrates the PostExecute hook which runs after theme files are written.
// It receives the list of files that were created, allowing plugins to act on them.
func (p *DunstifyPlugin) PostExecute(ctx context.Context, writtenFiles []string) error {
	fileCount := len(writtenFiles)
	var body string
	if fileCount == 1 {
		body = fmt.Sprintf("Generated %d theme file", fileCount)
	} else {
		body = fmt.Sprintf("Generated %d theme files", fileCount)
	}

	summary := "Theme Generated"
	icon := "preferences-desktop-theme"

	// Try dunstify first (supports more features)
	if dunstifyPath, err := exec.LookPath("dunstify"); err == nil {
		return sendDunstifyNotification(ctx, dunstifyPath, summary, body, icon)
	}

	// Fall back to notify-send
	if notifySendPath, err := exec.LookPath("notify-send"); err == nil {
		return sendNotifySendNotification(ctx, notifySendPath, summary, body, icon)
	}

	return fmt.Errorf("neither dunstify nor notify-send found on $PATH")
}

// GetMetadata returns plugin metadata.
func (p *DunstifyPlugin) GetMetadata() protocol.PluginInfo {
	return protocol.PluginInfo{
		Name:            "dunstify",
		Type:            "output",
		Version:         "0.0.1",
		ProtocolVersion: protocol.ProtocolVersion,
		Description:     "Send desktop notifications via dunstify or notify-send",
		PluginProtocol:  "go-plugin",
	}
}

// sendDunstifyNotification sends notification via dunstify (supports more features).
func sendDunstifyNotification(ctx context.Context, dunstifyPath, summary, body, icon string) error {
	// dunstify supports additional features like urgency, timeout, etc.
	// #nosec G204 -- dunstifyPath is validated via exec.LookPath
	cmd := exec.CommandContext(ctx, dunstifyPath,
		"-a", "tinct",
		"-i", icon,
		"-u", "low", // urgency: low, normal, critical
		"-t", "5000", // timeout in milliseconds
		summary,
		body,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dunstify failed: %w", err)
	}

	return nil
}

// sendNotifySendNotification sends notification via notify-send (basic features).
func sendNotifySendNotification(ctx context.Context, notifySendPath, summary, body, icon string) error {
	// notify-send has limited options compared to dunstify
	// #nosec G204 -- notifySendPath is validated via exec.LookPath
	cmd := exec.CommandContext(ctx, notifySendPath,
		"-a", "tinct",
		"-i", icon,
		"-u", "low", // urgency: low, normal, critical
		"-t", "5000", // timeout in milliseconds
		summary,
		body,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("notify-send failed: %w", err)
	}

	return nil
}

func main() {
	// Handle --plugin-info flag
	// All plugins must respond to --plugin-info for Tinct to discover them.
	// This returns JSON metadata about the plugin including its protocol type.
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		p := &DunstifyPlugin{}
		info := p.GetMetadata()

		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding plugin info: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Serve the plugin using go-plugin
	// This starts the RPC server that Tinct will communicate with.
	// The process stays alive for multiple invocations (6.5x faster than JSON-stdio).
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: protocol.Handshake,
		Plugins: map[string]plugin.Plugin{
			"output": &protocol.OutputPluginRPC{
				Impl: &DunstifyPlugin{},
			},
		},
	})
}
