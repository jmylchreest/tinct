package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	// Version is the semantic version of the plugin.
	// Injected at build time via: -ldflags "-X main.Version=x.y.z"
	Version = "0.0.0"

	// Commit is the git commit hash of the build.
	// Injected at build time via: -ldflags "-X main.Commit=$(git rev-parse HEAD)"
	Commit = "unknown"

	// Date is the build date in RFC3339 format.
	// Injected at build time via: -ldflags "-X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	Date = "unknown"
)

const (
	// Plugin metadata
	pluginName        = "wob"
	pluginDescription = "Generate wob (Wayland Overlay Bar) theme and provide wrapper functionality"

	// Runtime configuration
	defaultRuntimeDir = "wob" // under $XDG_RUNTIME_DIR
	defaultPipeName   = "wob.fifo"
	defaultConfigName = "wob-merged.ini"
	defaultPIDFile    = "wob.pid"
)

// Apply env defaults
func init() {
	if os.Getenv("WOB_PIPE") == "" {
		_ = os.Setenv("WOB_PIPE", defaultPipeName) // Setenv only fails if key/value are invalid
	}
	if os.Getenv("WOB_MERGED_CONFIG") == "" {
		_ = os.Setenv("WOB_MERGED_CONFIG", defaultConfigName) // Setenv only fails if key/value are invalid
	}
}

func main() {
	// Handle --plugin-info flag for protocol detection
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		info := map[string]string{
			"name":             pluginName,
			"type":             "output",
			"version":          Version,
			"protocol_version": "0.0.1",
			"description":      pluginDescription,
			"plugin_protocol":  "json-stdio",
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding plugin info: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// If no arguments or "plugin" command, run as JSON stdio plugin (Tinct integration)
	if len(os.Args) == 1 || (len(os.Args) > 1 && os.Args[1] == "plugin") {
		if err := runPlugin(); err != nil {
			fmt.Fprintf(os.Stderr, "Plugin error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Otherwise use the specified command for wrapper mode
	command := os.Args[1]

	switch command {
	case "start":
		// Wrapper mode: start wob
		if err := runStart(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Start error: %v\n", err)
			os.Exit(1)
		}

	case "send":
		// Wrapper mode: send value to wob
		if err := runSend(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Send error: %v\n", err)
			os.Exit(1)
		}

	case "stop":
		// Wrapper mode: stop wob
		if err := runStop(); err != nil {
			fmt.Fprintf(os.Stderr, "Stop error: %v\n", err)
			os.Exit(1)
		}

	case "status":
		// Wrapper mode: check if wob is running
		running, err := isWobRunning()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Status error: %v\n", err)
			os.Exit(1)
		}
		if running {
			fmt.Println("wob is running")
		} else {
			fmt.Println("wob is not running")
			os.Exit(1)
		}

	case "--pre-execute":
		// Tinct plugin protocol: check if wob binary exists
		if _, err := lookupWobBinary(); err != nil {
			fmt.Println("wob binary not found on $PATH")
			os.Exit(1)
		}
		os.Exit(0)

	case "--post-execute":
		// Tinct plugin protocol: run post-execution hook
		// This is handled by the plugin mode, not wrapper mode
		fmt.Fprintf(os.Stderr, "post-execute not supported in wrapper mode\n")
		os.Exit(1)

	case "help", "-h", "--help":
		printUsage()

	case "version", "-v", "--version":
		fmt.Printf("tinct-plugin-wob %s\n", Version)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Printf(`tinct-plugin-wob - Wob theme generator and wrapper

PLUGIN MODE (used by Tinct):
  tinct-plugin-wob plugin              Run as Tinct external plugin

WRAPPER MODE (direct usage):
  tinct-plugin-wob start [OPTIONS]     Start wob with theme
    --base-config FILE          Base wob.ini
    --append-config FILE        Additional config (repeatable)

  tinct-plugin-wob send VALUE [OPTIONS]     Send value to wob (0-100)
    --style STYLE               Style name (normal, critical, etc.)

  tinct-plugin-wob send CURRENT MAX    Send current/max as percentage

  tinct-plugin-wob stop                Stop running wob instance
  tinct-plugin-wob status              Check if wob is running
  tinct-plugin-wob help                Show this help
  tinct-plugin-wob version             Show version

EXAMPLES:
  # Start wob with tinct theme
  tinct-plugin-wob start --base-config ~/.config/wob/base.ini \
                  --append-config ~/.config/wob/themes/tinct.ini

  # Send volume level (auto-reloads if theme changed)
  tinct-plugin-wob send 45

  # Send brightness (current/max)
  tinct-plugin-wob send 4234 9600

  # Send with style
  tinct-plugin-wob send 95 --style critical

HYPRLAND INTEGRATION:
  exec-once = tinct-plugin-wob start --base-config ~/.config/wob/base.ini \
                               --append-config ~/.config/wob/themes/tinct.ini

  bind = , XF86AudioRaiseVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%%+ && \
         tinct-plugin-wob send $(wpctl get-volume @DEFAULT_SINK@ | awk '{print $2 * 100}')

NOTE:
  Config changes are automatically detected on send - wob will restart if needed.
`)
}
