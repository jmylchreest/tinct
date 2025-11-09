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
	if err := os.WriteFile(dummyImagePath, []byte("dummy image data"), 0o600); err != nil {
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

// TestGenerateCommand tests the generate command with various backends.
func TestGenerateCommand(t *testing.T) {
	dummyImagePath, cleanup := setupTests(t)
	defer cleanup()

	// Create root command once for all subtests.
	// Set os.Args to simulate running the command.
	os.Args = []string{"tinct"}

	// Capture stdout and stderr.
	var outBuf, errBuf bytes.Buffer
	rootCmd := cli.NewRootCmd()
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)

	t.Run("WithKmeansBackend", func(t *testing.T) {
		// Reset buffers.
		outBuf.Reset()
		errBuf.Reset()

		// Set command args.
		rootCmd.SetArgs([]string{"generate", "-i", "image", "-p", dummyImagePath, "--backend", "kmeans"})

		// Execute the command.
		if err := rootCmd.Execute(); err != nil {
			// We expect an error because the dummy image is invalid, but we can check
			// the error message to see if the backend was processed correctly.
			if !strings.Contains(err.Error(), "unsupported or invalid image format") && !strings.Contains(err.Error(), "failed to load image") {
				t.Fatalf("Expected error loading image, but got: %v", err)
			}
		}

		// Check that there are no errors about the backend.
		if strings.Contains(errBuf.String(), "invalid backend") {
			t.Errorf("Did not expect error about invalid backend, but got: %s", errBuf.String())
		}
	})

	t.Run("WithInvalidBackend", func(t *testing.T) {
		// Reset buffers.
		outBuf.Reset()
		errBuf.Reset()

		// Create a valid dummy PNG file (1x1 black pixel).
		validImagePath := filepath.Join(filepath.Dir(dummyImagePath), "valid.png")
		// Minimal valid PNG: 1x1 black pixel.
		pngData := []byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, // PNG signature
			0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1
			0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde,
			0x00, 0x00, 0x00, 0x0c, 0x49, 0x44, 0x41, 0x54, // IDAT chunk
			0x08, 0xd7, 0x63, 0x60, 0x60, 0x60, 0x00, 0x00, 0x00, 0x04, 0x00, 0x01,
			0x27, 0x9b, 0x71, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, // IEND chunk
			0xae, 0x42, 0x60, 0x82,
		}
		if err := os.WriteFile(validImagePath, pngData, 0o600); err != nil {
			t.Fatalf("Failed to create valid PNG: %v", err)
		}

		// Set command args with valid image but invalid backend.
		rootCmd.SetArgs([]string{"generate", "-i", "image", "-p", validImagePath, "--backend", "invalid-backend"})

		// Execute the command.
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("Expected an error for invalid backend, but got none")
		}

		// Check for the specific error message.
		if !strings.Contains(err.Error(), "invalid backend: invalid-backend") {
			t.Errorf("Expected error about invalid backend, but got: %v", err)
		}
	})
}
