package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	// Plugin metadata
	pluginName        = "wob"
	pluginDescription = "Generate wob (Wayland Overlay Bar) theme and provide wrapper functionality"
	pluginVersion     = "0.0.1"
	protocolVersion   = "0.0.1" // Tinct plugin protocol version

	// Runtime configuration
	defaultRuntimeDir = "wob" // under $XDG_RUNTIME_DIR
	defaultPipeName   = "wob.fifo"
	defaultConfigName = "wob-merged.ini"
	defaultPIDFile    = "wob.pid"
)

// RuntimePaths holds all runtime file paths
type RuntimePaths struct {
	Dir    string
	Pipe   string
	Config string
	PID    string
}

// WobConfig represents wob configuration
type WobConfig struct {
	BaseConfig   string
	AppendConfig []string
}

func main() {
	// Default to plugin mode when called without arguments (Tinct integration)
	// Otherwise use the specified command
	command := "plugin"
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}

	switch command {
	case "plugin":
		// Plugin mode: run as Tinct external plugin
		if err := runPlugin(); err != nil {
			fmt.Fprintf(os.Stderr, "Plugin error: %v\n", err)
			os.Exit(1)
		}

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

	case "help", "-h", "--help":
		printUsage()

	case "version", "-v", "--version":
		fmt.Printf("wob-tinct %s\n", pluginVersion)

	case "--plugin-info":
		// Plugin info mode: output JSON metadata for Tinct discovery
		printPluginInfo()

	case "--pre-execute":
		// Pre-execute hook: check if wob binary exists
		if err := runPreExecute(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

	case "--post-execute":
		// Post-execute hook: stop running wob instance so it picks up new theme on next use
		if err := runPostExecute(); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printPluginInfo() {
	info := map[string]interface{}{
		"name":             pluginName,
		"type":             "output",
		"description":      pluginDescription,
		"version":          pluginVersion,
		"protocol_version": protocolVersion,
		"author":           "Tinct Contributors",
		"capabilities": map[string]bool{
			"generate": true,
			"preview":  false,
		},
		"options": []map[string]string{},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(info); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode plugin info: %v\n", err)
		os.Exit(1)
	}
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

  # Send volume level
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
	if stat.Uid != uint32(os.Getuid()) {
		return nil, fmt.Errorf("runtime directory not owned by current user")
	}

	return &RuntimePaths{
		Dir:    baseDir,
		Pipe:   filepath.Join(baseDir, os.Getenv("WOB_PIPE")),
		Config: filepath.Join(baseDir, os.Getenv("WOB_MERGED_CONFIG")),
		PID:    filepath.Join(baseDir, defaultPIDFile),
	}, nil
}

// Apply env defaults
func init() {
	if os.Getenv("WOB_PIPE") == "" {
		os.Setenv("WOB_PIPE", defaultPipeName)
	}
	if os.Getenv("WOB_MERGED_CONFIG") == "" {
		os.Setenv("WOB_MERGED_CONFIG", defaultConfigName)
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

// runStart starts wob with optional config
func runStart(args []string) error {
	var baseConfig string
	var appendConfigs []string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--base-config":
			if i+1 >= len(args) {
				return fmt.Errorf("--base-config requires an argument")
			}
			baseConfig = args[i+1]
			i++
		case "--append-config":
			if i+1 >= len(args) {
				return fmt.Errorf("--append-config requires an argument")
			}
			appendConfigs = append(appendConfigs, args[i+1])
			i++
		default:
			return fmt.Errorf("unknown option: %s", args[i])
		}
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

	paths, err := getRuntimePaths()
	if err != nil {
		return err
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

	if baseConfig != "" {
		// Merge configs
		mergedPath, err := mergeConfigs(paths, baseConfig, appendConfigs)
		if err != nil {
			return fmt.Errorf("failed to merge configs: %w", err)
		}
		wobCmd = exec.Command(wobBin, "-c", mergedPath)
	} else {
		wobCmd = exec.Command(wobBin)
	}

	// Start tail | wob pipeline
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
		tailCmd.Process.Kill()
		return fmt.Errorf("failed to start wob: %w", err)
	}

	// Write PID file
	if err := writePIDFile(paths, wobCmd.Process.Pid); err != nil {
		wobCmd.Process.Kill()
		tailCmd.Process.Kill()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Started wob (PID: %d)\n", wobCmd.Process.Pid)

	// Wait for wob to exit
	wobCmd.Wait()
	tailCmd.Process.Kill()

	// Cleanup
	os.Remove(paths.Pipe)
	os.Remove(paths.PID)

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
	baseData, err := os.ReadFile(baseConfig)
	if err != nil {
		return "", fmt.Errorf("failed to read base config: %w", err)
	}
	tmpFile.Write(baseData)

	// Append additional configs
	for _, appendConfig := range appendConfigs {
		fmt.Fprintf(tmpFile, "\n# Append: %s\n", appendConfig)
		appendData, err := os.ReadFile(appendConfig)
		if err != nil {
			return "", fmt.Errorf("failed to read append config %s: %w", appendConfig, err)
		}
		tmpFile.Write(appendData)
	}

	tmpPath := tmpFile.Name()
	finalPath := paths.Config

	// Atomic rename
	if err := os.Rename(tmpPath, finalPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to rename merged config: %w", err)
	}

	if err := os.Chmod(finalPath, 0600); err != nil {
		return "", fmt.Errorf("failed to chmod merged config: %w", err)
	}

	return finalPath, nil
}

// runSend sends a value to wob
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

	// Check if wob is running, start if not
	running, err := isWobRunning()
	if err != nil {
		return err
	}
	if !running {
		// Start wob in background with default config
		go func() {
			runStart([]string{})
		}()
		time.Sleep(100 * time.Millisecond)
	}

	paths, err := getRuntimePaths()
	if err != nil {
		return err
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
		process.Kill()
	}

	// Cleanup
	os.Remove(paths.PID)

	fmt.Println("Stopped wob")
	return nil
}

// runPreExecute implements the pre-execute hook for Tinct
// Exit code 0 = continue, 1 = skip plugin, 2+ = error
func runPreExecute() error {
	// Check if wob binary exists on PATH
	_, err := exec.LookPath("wob")
	if err != nil {
		// Exit code 1 = skip plugin gracefully
		fmt.Println("wob binary not found on $PATH")
		os.Exit(1)
	}

	// Check if config directory exists (create if it doesn't)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Exit code 2+ = actual error
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "wob")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Create it - this is expected on first run
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create wob config directory: %w", err)
		}
	}

	// Exit code 0 = continue
	return nil
}

// runPostExecute implements the post-execute hook for Tinct
// Stops any running wob instance so it picks up the new theme on next use
func runPostExecute() error {
	// Read written files JSON from stdin (we don't actually need this for wob)
	var input struct {
		WrittenFiles []string `json:"written_files"`
	}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&input); err != nil {
		// Don't fail if we can't read stdin - just continue
		fmt.Fprintf(os.Stderr, "Warning: failed to read post-execute input: %v\n", err)
	}

	// Check if wob is running
	running, err := isWobRunning()
	if err != nil {
		// If we can't check, just continue silently
		return nil
	}

	if !running {
		// Not running, nothing to do
		fmt.Fprintf(os.Stderr, "wob is not running, theme will be used on next start\n")
		return nil
	}

	// Stop the running instance
	if err := runStop(); err != nil {
		// Don't fail the whole operation if stop fails
		fmt.Fprintf(os.Stderr, "Warning: failed to stop wob: %v\n", err)
		fmt.Fprintf(os.Stderr, "You may need to manually restart wob to use the new theme\n")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Stopped wob - theme will be applied on next use\n")
	return nil
}

// runPlugin runs in Tinct plugin mode
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
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	themeFile := filepath.Join(themesDir, "tinct.ini")

	// Generate theme content
	themeContent, err := generateWobTheme(palette)
	if err != nil {
		return fmt.Errorf("failed to generate theme: %w", err)
	}

	if err := os.WriteFile(themeFile, []byte(themeContent), 0644); err != nil {
		return fmt.Errorf("failed to write theme file: %w", err)
	}

	// Install wrapper (copy self)
	scriptsDir := filepath.Join(homeDir, ".config", "wob", "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
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

	if err := os.Chmod(wrapperPath, 0755); err != nil {
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

// generateWobTheme creates wob INI content from palette
func generateWobTheme(palette map[string]interface{}) (string, error) {
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

	bg := getColor("background")
	border := getColor("border")
	accent1 := getColor("accent1")
	success := getColor("success")
	warning := getColor("warning")
	danger := getColor("danger")

	theme := fmt.Sprintf(`# Wob theme generated by Tinct
# https://github.com/jmylchreest/tinct

[default]
# Bar dimensions
height = 50
border_offset = 4
border_size = 2
bar_padding = 3

# Colors (ARGB format)
background_color = FF%s
border_color = FF%s
bar_color = FF%s

[normal]
bar_color = FF%s

[critical]
bar_color = FF%s

[warning]
bar_color = FF%s
`, bg, border, accent1, success, danger, warning)

	return theme, nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
