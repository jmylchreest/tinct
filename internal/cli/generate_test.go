// Package cli_test provides tests for the CLI package.
package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/cli"
)

// setupTests initializes the CLI for testing and returns a cleanup function.
func setupTests(t *testing.T) (string, func()) {
	// Keep track of original os.Args.
	originalArgs := os.Args
	// Create a temporary directory for test data.
	tempDir, err := os.MkdirTemp("", "tinct-tests-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a dummy image file.
	dummyImagePath := filepath.Join(tempDir, "test.png")
	if err := os.WriteFile(dummyImagePath, []byte("dummy image data"), 0600); err != nil {
		t.Fatalf("Failed to create dummy image file: %v", err)
	}

	// Return a cleanup function.
	return dummyImagePath, func() {
		os.Args = originalArgs
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}
}

// TestGenerateCommandWithKmeansBackend tests the generate command with the kmeans backend.
func TestGenerateCommandWithKmeansBackend(t *testing.T) {
	dummyImagePath, cleanup := setupTests(t)
	defer cleanup()

	// Set os.Args to simulate running the command.
	os.Args = []string{"tinct", "generate", "-i", "image", "-p", dummyImagePath, "--backend", "kmeans"}

	// Capture stdout and stderr.
	var outBuf, errBuf bytes.Buffer
	rootCmd := cli.NewRootCmd()
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)

	// Execute the command.
	if err := rootCmd.Execute(); err != nil {
		// We expect an error because the dummy image is invalid, but we can check
		// the error message to see if the backend was processed correctly.
		if !strings.Contains(err.Error(), "failed to load image") {
			t.Fatalf("Expected error loading image, but got: %v", err)
		}
	}

	// Check that there are no errors about the backend.
	if strings.Contains(errBuf.String(), "invalid backend") {
		t.Errorf("Did not expect error about invalid backend, but got: %s", errBuf.String())
	}
}

// TestGenerateCommandWithInvalidBackend tests the generate command with an invalid backend.
func TestGenerateCommandWithInvalidBackend(t *testing.T) {
	dummyImagePath, cleanup := setupTests(t)
	defer cleanup()

	// Set os.Args to simulate running the command.
	os.Args = []string{"tinct", "generate", "-i", "image", "-p", dummyImagePath, "--backend", "invalid-backend"}

	// Capture stdout and stderr.
	var outBuf, errBuf bytes.Buffer
	rootCmd := cli.NewRootCmd()
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)

	// Execute the command.
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("Expected an error for invalid backend, but got none")
	}

	// Check for the specific error message.
	if !strings.Contains(err.Error(), "invalid backend: invalid-backend") {
		t.Errorf("Expected error about invalid backend, but got: %v", err)
	}
}
