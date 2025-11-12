// Package testing provides shared test utilities for output plugins.
package testing

import (
	"context"
	"image/color"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jmylchreest/tinct/internal/colour"
	"github.com/jmylchreest/tinct/internal/plugin/output"
)

// TestBasicInterface tests the basic plugin interface methods that all plugins must implement.
func TestBasicInterface(t *testing.T, p output.Plugin, expectedName string, expectedDirSubstring string) {
	t.Run("Name", func(t *testing.T) {
		if p.Name() != expectedName {
			t.Errorf("Name() = %s, want %s", p.Name(), expectedName)
		}
	})

	t.Run("Description", func(t *testing.T) {
		desc := p.Description()
		if desc == "" {
			t.Error("Description() should not be empty")
		}
	})

	t.Run("DefaultOutputDir", func(t *testing.T) {
		dir := p.DefaultOutputDir()
		if dir == "" {
			t.Error("DefaultOutputDir() should not be empty")
		}
		// Use expectedDirSubstring if provided, otherwise fall back to expectedName
		checkString := expectedDirSubstring
		if checkString == "" {
			checkString = expectedName
		}
		if !strings.Contains(dir, checkString) {
			t.Errorf("DefaultOutputDir() = %s, should contain '%s'", dir, checkString)
		}
	})

	t.Run("Validate", func(t *testing.T) {
		if err := p.Validate(); err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})
}

// TestGeneration tests the Generate method with various scenarios.
func TestGeneration(t *testing.T, p output.Plugin, expectedFiles []string) {
	t.Run("Generate", func(t *testing.T) {
		palette := CreateTestPalette(colour.ThemeDark)
		themeData := colour.NewThemeData(palette, "", "")
		files, err := p.Generate(themeData)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(files) != len(expectedFiles) {
			t.Fatalf("Generate() returned %d files, want %d", len(files), len(expectedFiles))
		}

		for _, expectedFile := range expectedFiles {
			if _, ok := files[expectedFile]; !ok {
				t.Errorf("Generate() did not return %s", expectedFile)
			}
		}
	})

	t.Run("GenerateNilPalette", func(t *testing.T) {
		_, err := p.Generate(nil)
		if err == nil {
			t.Error("Generate() with nil palette should return error")
		}
	})

	t.Run("GenerateWithLightTheme", func(t *testing.T) {
		palette := CreateTestPalette(colour.ThemeLight)
		themeData := colour.NewThemeData(palette, "", "")
		files, err := p.Generate(themeData)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if len(files) == 0 {
			t.Error("Generate() returned no files")
		}
	})
}

// TestVerbosePlugin tests verbose functionality if the plugin supports it.
func TestVerbosePlugin(t *testing.T, p any) {
	vp, ok := p.(interface {
		SetVerbose(bool)
	})
	if !ok {
		t.Skip("Plugin does not implement SetVerbose")
	}

	t.Run("SetVerbose", func(_ *testing.T) {
		// Just test that it doesn't panic.
		vp.SetVerbose(true)
		vp.SetVerbose(false)
	})
}

// TestFlags tests plugin-specific flag registration.
func TestFlags(t *testing.T, p output.Plugin, expectedFlagPrefix string) {
	t.Run("RegisterFlags", func(t *testing.T) {
		cmd := &cobra.Command{
			Use: "test",
		}

		p.RegisterFlags(cmd)

		expectedFlag := expectedFlagPrefix + ".output-dir"
		flag := cmd.Flags().Lookup(expectedFlag)
		if flag == nil {
			t.Errorf("RegisterFlags() did not register %s flag", expectedFlag)
		}
	})
}

// TestPreExecuteHook tests the PreExecute hook if the plugin implements it.
func TestPreExecuteHook(t *testing.T, p any, expectedBinaryName string) {
	peh, ok := p.(interface {
		PreExecute(context.Context) (bool, string, error)
	})
	if !ok {
		t.Skip("Plugin does not implement PreExecute")
	}

	t.Run("PreExecute", func(t *testing.T) {
		ctx := context.Background()
		skip, reason, err := peh.PreExecute(ctx)

		if err != nil {
			t.Errorf("PreExecute() unexpected error = %v", err)
		}

		// If skipped, reason should mention the binary name.
		if skip && !strings.Contains(reason, expectedBinaryName) {
			t.Errorf("PreExecute() skip reason should mention %s, got: %s", expectedBinaryName, reason)
		}
	})
}

// TestEmbeddedTemplates tests that embedded templates are accessible.
func TestEmbeddedTemplates(t *testing.T, _ any, _ []string) {
	// This is a bit awkward because embed.FS isn't directly testable.
	// Skip this test for now - it's plugin-specific.
	t.Skip("Template testing should be done per-plugin due to embed.FS limitations")
}

// CreateTestPalette creates a test palette with standard colors for testing.
func CreateTestPalette(themeType colour.ThemeType) *colour.CategorisedPalette {
	colors := []color.Color{
		color.RGBA{R: 26, G: 27, B: 38, A: 255},    // Dark background
		color.RGBA{R: 192, G: 202, B: 245, A: 255}, // Light foreground
		color.RGBA{R: 122, G: 162, B: 247, A: 255}, // Blue accent1
		color.RGBA{R: 187, G: 154, B: 247, A: 255}, // Purple accent2
		color.RGBA{R: 125, G: 207, B: 255, A: 255}, // Cyan accent3
		color.RGBA{R: 158, G: 206, B: 106, A: 255}, // Green accent4/success
		color.RGBA{R: 247, G: 118, B: 142, A: 255}, // Red/danger
		color.RGBA{R: 224, G: 175, B: 104, A: 255}, // Orange/warning
	}

	palette := &colour.Palette{Colors: colors}
	config := colour.DefaultCategorisationConfig()
	config.ThemeType = themeType

	return colour.Categorise(palette, config)
}

// RunAllTests runs all standard tests for a plugin.
func RunAllTests(t *testing.T, p output.Plugin, config TestConfig) {
	TestBasicInterface(t, p, config.ExpectedName, config.ExpectedDirSubstring)
	TestGeneration(t, p, config.ExpectedFiles)
	TestVerbosePlugin(t, p)
	TestFlags(t, p, config.ExpectedName)
	TestPreExecuteHook(t, p, config.ExpectedBinaryName)
}

// TestConfig holds configuration for running plugin tests.
type TestConfig struct {
	ExpectedName         string   // Plugin name
	ExpectedFiles        []string // Files that Generate() should return
	ExpectedBinaryName   string   // Binary name to check in PreExecute (e.g., "fuzzel", "dunst")
	ExpectedDirSubstring string   // Optional: substring to check in DefaultOutputDir (defaults to ExpectedName if empty)
}
