package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// isValidPath checks if a path is safe to use in commands.
func isValidPath(path string) bool {
	// Reject paths with suspicious characters
	if strings.Contains(path, "..") || strings.ContainsAny(path, "|&;`$()") {
		return false
	}
	// Clean the path and ensure it matches
	cleaned := filepath.Clean(path)
	return cleaned == path
}

// isValidExecutableName checks if a string is a safe executable name (no path separators).
func isValidExecutableName(name string) bool {
	// Reject if it contains path separators or suspicious characters
	if strings.ContainsAny(name, "/\\|&;`$()") {
		return false
	}
	// Should not be empty
	return name != ""
}

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

// RuntimePaths holds all runtime file paths
type RuntimePaths struct {
	Dir        string
	Pipe       string
	Config     string
	PID        string
	ConfigInfo string
}

// ConfigInfo tracks config sources for reload detection
type ConfigInfo struct {
	BaseConfig    string   `json:"base_config"`
	AppendConfigs []string `json:"append_configs"`
}

func main() {
	// Handle --plugin-info flag for protocol detection
	if len(os.Args) > 1 && os.Args[1] == "--plugin-info" {
		p := &WobPlugin{}
		info := p.GetMetadata()
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
		// Tinct plugin protocol: check if plugin can run
		p := &WobPlugin{}
		skip, reason, err := p.PreExecute(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(2)
		}
		if skip {
			fmt.Println(reason)
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
		fmt.Printf("wob-tinct %s\n", Version)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// lookupWobBinary finds the wob binary on PATH
func lookupWobBinary() (string, error) {
	wobBin := os.Getenv("WOB_BIN")
	if wobBin == "" {
		wobBin = "wob"
	}

	if !isValidExecutableName(wobBin) {
		return "", fmt.Errorf("invalid WOB_BIN executable name: contains suspicious characters")
	}

	return exec.LookPath(wobBin)
}

func printUsage() {
	fmt.Printf(`wob-tinct - Wob theme generator and wrapper

PLUGIN MODE (used by Tinct):
  wob-tinct plugin              Run as Tinct external plugin

WRAPPER MODE (direct usage):
  wob-tinct start [OPTIONS]     Start wob with theme
    --base-config FILE          Base wob.ini
    --append-config FILE        Additional config (repeatable)

  wob-tinct send VALUE [OPTIONS]     Send value to wob (0-100)
    --style STYLE               Style name (normal, critical, etc.)

  wob-tinct send CURRENT MAX    Send current/max as percentage

  wob-tinct stop                Stop running wob instance
  wob-tinct status              Check if wob is running
  wob-tinct help                Show this help
  wob-tinct version             Show version

EXAMPLES:
  # Start wob with tinct theme
  wob-tinct start --base-config ~/.config/wob/base.ini \
                  --append-config ~/.config/wob/themes/tinct.ini

  # Send volume level (auto-reloads if theme changed)
  wob-tinct send 45

  # Send brightness (current/max)
  wob-tinct send 4234 9600

  # Send with style
  wob-tinct send 95 --style critical

HYPRLAND INTEGRATION:
  exec-once = wob-tinct start --base-config ~/.config/wob/base.ini \
                               --append-config ~/.config/wob/themes/tinct.ini

  bind = , XF86AudioRaiseVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%%+ && \
         wob-tinct send $(wpctl get-volume @DEFAULT_SINK@ | awk '{print $2 * 100}')

NOTE:
  Config changes are automatically detected on send - wob will restart if needed.
`)
}

// getRuntimePaths returns the runtime directory paths
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
		baseDir = filepath.Join("/tmp", fmt.Sprintf("%s-wob-runtime", username))
	}

	// Create runtime directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create runtime directory: %w", err)
	}

	// Verify directory ownership
	fileInfo, err := os.Stat(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat runtime directory: %w", err)
	}

	stat := fileInfo.Sys().(*syscall.Stat_t)
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

// Apply env defaults
func init() {
	if os.Getenv("WOB_PIPE") == "" {
		_ = os.Setenv("WOB_PIPE", defaultPipeName) // Setenv only fails if key/value are invalid
	}
	if os.Getenv("WOB_MERGED_CONFIG") == "" {
		_ = os.Setenv("WOB_MERGED_CONFIG", defaultConfigName) // Setenv only fails if key/value are invalid
	}
}

// ensureFIFO creates the FIFO pipe if it doesn't exist
func ensureFIFO(paths *RuntimePaths) error {
	// Check if pipe exists
	if info, err := os.Stat(paths.Pipe); err == nil {
		// Exists - check if it's a FIFO
		if info.Mode()&os.ModeNamedPipe == 0 {
			// Not a FIFO, remove it
			if err := os.Remove(paths.Pipe); err != nil {
				return fmt.Errorf("failed to remove non-FIFO at pipe path: %w", err)
			}
		} else {
			// Already a FIFO
			return nil
		}
	}

	// Create FIFO
	if err := syscall.Mkfifo(paths.Pipe, 0600); err != nil {
		return fmt.Errorf("failed to create FIFO: %w", err)
	}

	return nil
}

// isWobRunning checks if wob is currently running
func isWobRunning() (bool, error) {
	paths, err := getRuntimePaths()
	if err != nil {
		return false, err
	}

	// Check PID file
	pidData, err := os.ReadFile(paths.PID)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil {
		return false, nil
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil, nil
}

// writePIDFile writes the PID to the PID file
func writePIDFile(paths *RuntimePaths, pid int) error {
	return os.WriteFile(paths.PID, []byte(fmt.Sprintf("%d\n", pid)), 0600)
}

// saveConfigInfo saves config info to JSON file
func saveConfigInfo(paths *RuntimePaths, info *ConfigInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(paths.ConfigInfo, data, 0600)
}

// loadConfigInfo loads config info from JSON file
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

// needsConfigReload checks if any source configs are newer than merged config
func needsConfigReload(paths *RuntimePaths, configInfo *ConfigInfo) (bool, error) {
	// Check if merged config exists
	mergedInfo, err := os.Stat(paths.Config)
	if err != nil {
		// Merged config doesn't exist - reload needed
		return true, nil
	}
	mergedTime := mergedInfo.ModTime()

	// Check base config mtime
	if configInfo.BaseConfig != "" {
		baseInfo, err := os.Stat(configInfo.BaseConfig)
		if err == nil && baseInfo.ModTime().After(mergedTime) {
			return true, nil
		}
	}

	// Check append configs mtime
	for _, appendPath := range configInfo.AppendConfigs {
		appendInfo, err := os.Stat(appendPath)
		if err == nil && appendInfo.ModTime().After(mergedTime) {
			return true, nil
		}
	}

	return false, nil
}

// runStart starts wob with optional config (runs in background, does not block)
func runStart(args []string) error {
	var baseConfig string
	var appendConfigs []string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] { // #nosec G602 - bounds checked above
		case "--base-config":
			if i+1 >= len(args) {
				return fmt.Errorf("--base-config requires an argument")
			}
			baseConfig = args[i+1] // #nosec G602 - bounds checked
			i++
		case "--append-config":
			if i+1 >= len(args) {
				return fmt.Errorf("--append-config requires an argument")
			}
			appendConfigs = append(appendConfigs, args[i+1])
			i++
		default:
			return fmt.Errorf("unknown option: %s", args[i]) // #nosec G602 - loop condition ensures valid index
		}
	}

	paths, err := getRuntimePaths()
	if err != nil {
		return err
	}

	// Check if wob is already running
	running, err := isWobRunning()
	if err != nil {
		return fmt.Errorf("failed to check wob status: %w", err)
	}
	if running {
		fmt.Println("wob is already running")
		return nil
	}

	// Ensure FIFO exists
	if err := ensureFIFO(paths); err != nil {
		return err
	}

	// Determine wob command
	var wobCmd *exec.Cmd
	wobBin := os.Getenv("WOB_BIN")
	if wobBin == "" {
		wobBin = "wob"
	}

	// Validate executable name to prevent command injection
	if !isValidExecutableName(wobBin) {
		return fmt.Errorf("invalid WOB_BIN executable name: contains suspicious characters")
	}

	if baseConfig != "" {
		// Merge configs
		mergedPath, err := mergeConfigs(paths, baseConfig, appendConfigs)
		if err != nil {
			return fmt.Errorf("failed to merge configs: %w", err)
		}
		// Validate merged config path
		if !isValidPath(mergedPath) {
			return fmt.Errorf("invalid merged config path: contains suspicious characters")
		}

		// #nosec G204 -- wobBin and mergedPath are validated
		wobCmd = exec.Command(wobBin, "-c", mergedPath)
	} else {
		// #nosec G204 -- wobBin is validated as a safe executable name
		wobCmd = exec.Command(wobBin)
	}

	// Validate pipe path to prevent command injection
	if !isValidPath(paths.Pipe) {
		return fmt.Errorf("invalid pipe path: contains suspicious characters")
	}

	// Start tail | wob pipeline
	// #nosec G204 -- paths.Pipe is validated to be a safe file path
	tailCmd := exec.Command("tail", "-f", paths.Pipe)
	wobCmd.Stdin, err = tailCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}

	// Start both processes
	if err := tailCmd.Start(); err != nil {
		return fmt.Errorf("failed to start tail: %w", err)
	}

	if err := wobCmd.Start(); err != nil {
		_ = tailCmd.Process.Kill() // Best effort cleanup
		return fmt.Errorf("failed to start wob: %w", err)
	}

	// Write PID file
	if err := writePIDFile(paths, wobCmd.Process.Pid); err != nil {
		_ = wobCmd.Process.Kill()  // Best effort cleanup
		_ = tailCmd.Process.Kill() // Best effort cleanup
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Save config info for reload detection (only if config was provided)
	if baseConfig != "" {
		configInfo := &ConfigInfo{
			BaseConfig:    baseConfig,
			AppendConfigs: appendConfigs,
		}
		if err := saveConfigInfo(paths, configInfo); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to save config info: %v\n", err)
			// Not fatal, continue
		}
	}

	fmt.Fprintf(os.Stderr, "Started wob (PID: %d)\n", wobCmd.Process.Pid)

	return nil
}

// mergeConfigs merges base and append configs into a single file
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

	if err := os.Chmod(finalPath, 0600); err != nil {
		return "", fmt.Errorf("failed to chmod merged config: %w", err)
	}

	return finalPath, nil
}

// runSend sends a value to wob, restarting it if config changed
func runSend(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("send requires at least one argument")
	}

	var style string
	var values []int

	// Parse arguments
	for i := 0; i < len(args); i++ {
		if args[i] == "--style" {
			if i+1 >= len(args) {
				return fmt.Errorf("--style requires an argument")
			}
			style = args[i+1]
			i++
		} else {
			val, err := strconv.Atoi(args[i])
			if err != nil {
				return fmt.Errorf("invalid numeric value: %s", args[i])
			}
			values = append(values, val)
		}
	}

	if len(values) == 0 {
		return fmt.Errorf("no numeric values provided")
	}

	paths, err := getRuntimePaths()
	if err != nil {
		return err
	}

	// Check if wob is running
	running, err := isWobRunning()
	if err != nil {
		return err
	}

	if !running {
		return fmt.Errorf("wob is not running - start it first with: wob-tinct start")
	}

	// Try to load saved config info and check if reload needed
	configInfo, err := loadConfigInfo(paths)
	if err == nil {
		needsReload, err := needsConfigReload(paths, configInfo)
		if err == nil && needsReload {
			fmt.Fprintf(os.Stderr, "Detected config change, reloading wob\n")

			// Stop current wob
			if err := runStop(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to stop wob: %v\n", err)
			} else {
				// Wait for cleanup
				time.Sleep(200 * time.Millisecond)

				// Restart with saved config
				startArgs := []string{}
				if configInfo.BaseConfig != "" {
					startArgs = append(startArgs, "--base-config", configInfo.BaseConfig)
				}
				for _, appendPath := range configInfo.AppendConfigs {
					startArgs = append(startArgs, "--append-config", appendPath)
				}

				if err := runStart(startArgs); err != nil {
					return fmt.Errorf("failed to restart wob: %w", err)
				}

				// Wait for wob to be ready
				time.Sleep(300 * time.Millisecond)
			}
		}
	}

	// Ensure FIFO exists
	if err := ensureFIFO(paths); err != nil {
		return err
	}

	// Calculate value to send
	var percentage int
	if len(values) == 2 && values[1] > 0 {
		// Two values: treat as current/max
		current := values[0]
		max := values[1]
		percentage = (100*current + max/2) / max // Rounded percentage
	} else {
		// Single value or multiple: send first value
		percentage = values[0]
	}

	// Clamp to 0-100
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	// Open FIFO for writing
	pipe, err := os.OpenFile(paths.Pipe, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open FIFO: %w", err)
	}
	defer pipe.Close()

	// Write value
	var line string
	if style != "" {
		line = fmt.Sprintf("%d %s\n", percentage, style)
	} else {
		line = fmt.Sprintf("%d\n", percentage)
	}

	if _, err := pipe.WriteString(line); err != nil {
		return fmt.Errorf("failed to write to FIFO: %w", err)
	}

	return nil
}

// runStop stops the running wob instance
func runStop() error {
	paths, err := getRuntimePaths()
	if err != nil {
		return err
	}

	// Read PID
	pidData, err := os.ReadFile(paths.PID)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wob is not running")
		}
		return err
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
	if err != nil {
		return fmt.Errorf("invalid PID file")
	}

	// Kill process
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to terminate wob: %w", err)
	}

	// Wait a bit for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	// Force kill if still running
	if err := process.Signal(syscall.Signal(0)); err == nil {
		_ = process.Kill() // Best effort force kill
	}

	// Cleanup
	_ = os.Remove(paths.PID)
	_ = os.Remove(paths.ConfigInfo)

	fmt.Println("Stopped wob")
	return nil
}

// runPlugin runs in Tinct plugin mode (legacy JSON-stdio, deprecated)
func runPlugin() error {
	// Read palette from stdin (JSON)
	var palette map[string]interface{}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&palette); err != nil {
		return fmt.Errorf("failed to decode palette: %w", err)
	}

	// Generate wob theme
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	themesDir := filepath.Join(homeDir, ".config", "wob", "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil { // #nosec G301 - Themes directory needs standard permissions
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	themeFile := filepath.Join(themesDir, "tinct.ini")

	// Generate theme content
	themeContent, err := generateWobThemeFromMap(palette)
	if err != nil {
		return fmt.Errorf("failed to generate theme: %w", err)
	}

	if err := os.WriteFile(themeFile, []byte(themeContent), 0644); err != nil { // #nosec G306 - Theme file needs standard read permissions
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	// Install wrapper (copy self)
	scriptsDir := filepath.Join(homeDir, ".config", "wob", "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil { // #nosec G301 - Scripts directory needs standard permissions
		return fmt.Errorf("failed to create scripts directory: %w", err)
	}

	wrapperPath := filepath.Join(scriptsDir, "wob-tinct")
	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Copy self to wrapper location
	if err := copyFile(selfPath, wrapperPath); err != nil {
		return fmt.Errorf("failed to install wrapper: %w", err)
	}

	if err := os.Chmod(wrapperPath, 0755); err != nil { // #nosec G302 - Wrapper executable needs execute permission
		return fmt.Errorf("failed to chmod wrapper: %w", err)
	}

	// Output success message with instructions
	fmt.Fprintf(os.Stderr, "\nGenerated wob theme: %s\n", themeFile)
	fmt.Fprintf(os.Stderr, "Installed wrapper: %s\n\n", wrapperPath)
	fmt.Fprintf(os.Stderr, "To use with Hyprland, add to your hyprland.conf:\n\n")
	fmt.Fprintf(os.Stderr, "  exec-once = %s start --base-config ~/.config/wob/base.ini \\\n", wrapperPath)
	fmt.Fprintf(os.Stderr, "                       --append-config ~/.config/wob/themes/tinct.ini\n\n")
	fmt.Fprintf(os.Stderr, "Then bind keys to send values:\n\n")
	fmt.Fprintf(os.Stderr, "  bind = , XF86AudioRaiseVolume, exec, wpctl set-volume @DEFAULT_SINK@ 5%%+ && \\\n")
	fmt.Fprintf(os.Stderr, "         %s send $(wpctl get-volume @DEFAULT_SINK@ | awk '{print $2 * 100}')\n\n", wrapperPath)

	return nil
}

// generateWobThemeFromMap creates wob INI content from palette map (JSON-stdio mode)
func generateWobThemeFromMap(palette map[string]interface{}) (string, error) {
	// Load template from embedded filesystem
	tmplContent, err := templatesFS.ReadFile("templates/tinct.ini.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Helper to get hex color from nested palette structure
	getColor := func(key string) string {
		// Access palette.colours[key].hex
		if colours, ok := palette["colours"].(map[string]interface{}); ok {
			if colorObj, ok := colours[key].(map[string]interface{}); ok {
				if hex, ok := colorObj["hex"].(string); ok {
					return strings.TrimPrefix(hex, "#")
				}
			}
		}
		return "000000"
	}

	// Create a simple color map for the template
	colorMap := map[string]string{
		"background": getColor("background"),
		"foreground": getColor("foreground"),
		"accent1":    getColor("accent1"),
		"success":    getColor("success"),
		"warning":    getColor("warning"),
		"danger":     getColor("danger"),
	}

	// Parse and execute template
	tmpl, err := template.New("wob").Funcs(template.FuncMap{
		"get": func(m map[string]string, key string) string {
			if color, ok := m[key]; ok {
				return color
			}
			return "000000"
		},
	}).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, colorMap); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src) // #nosec G304 - User-specified source file, intended to be read
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst) // #nosec G304 - User-specified destination file
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
