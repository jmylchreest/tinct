package neovim

import (
	"strings"
	"testing"

	"github.com/jmylchreest/tinct/internal/colour"
	plugintesting "github.com/jmylchreest/tinct/internal/plugin/output/testing"
)

// TestNeovimPlugin runs all standard plugin tests using shared utilities.
func TestNeovimPlugin(t *testing.T) {
	plugin := New()

	config := plugintesting.TestConfig{
		ExpectedName:         "neovim",
		ExpectedFiles:        []string{"tinct.lua"},
		ExpectedBinaryName:   "nvim",
		ExpectedDirSubstring: "nvim", // Plugin uses .config/nvim/colors, not .config/neovim
	}

	plugintesting.RunAllTests(t, plugin, config)
}

// TestNeovimPlugin_ContentValidation tests neovim-specific content requirements.
func TestNeovimPlugin_ContentValidation(t *testing.T) {
	palette := plugintesting.CreateTestPalette(colour.ThemeDark)
	plugin := New()

	themeData := colour.NewThemeData(palette, "", "")
	files, err := plugin.Generate(themeData)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := string(files["tinct.lua"])

	// Check for required Neovim theme structure.
	requiredStrings := []string{
		"-- Tinct Color Scheme for Neovim",
		"vim.cmd('highlight clear')",
		"vim.o.termguicolors",
		"vim.g.colors_name",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(content, required) {
			t.Errorf("Generated content missing required string: %s", required)
		}
	}
}
