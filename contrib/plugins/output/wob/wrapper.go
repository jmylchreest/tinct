package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

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

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to terminate wob: %w", err)
	}

	// Cleanup PID files immediately - process will die on its own
	_ = os.Remove(paths.PID)
	_ = os.Remove(paths.ConfigInfo)

	fmt.Println("Stopped wob")
	return nil
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
		return fmt.Errorf("wob is not running - start it first with: tinct-plugin-wob start")
	}

	// Try to load saved config info and check if reload needed
	configInfo, err := loadConfigInfo(paths)
	if err == nil {
		needsReload, err := needsConfigReload(paths, configInfo)
		if err == nil && needsReload {
			// Stop current wob
			if err := runStop(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to stop wob: %v\n", err)
			} else {
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
