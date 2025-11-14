package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

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
